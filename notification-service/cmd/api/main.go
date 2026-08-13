package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ishimweBonheur/order-management/notification-service/internal/config"
	"github.com/ishimweBonheur/order-management/notification-service/internal/messaging"
	"github.com/ishimweBonheur/order-management/notification-service/internal/service"
	"github.com/redis/go-redis/v9"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	svc := service.New(redisClient, logger)
	consumer := messaging.New(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroup, svc, logger)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}), ReadHeaderTimeout: 5 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	errs := make(chan error, 2)
	go func() {
		for ctx.Err() == nil {
			if err := consumer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("Kafka consumer interrupted; retrying", "error", err)
				select {
				case <-time.After(2 * time.Second):
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	go func() { errs <- server.ListenAndServe() }()
	select {
	case err := <-errs:
		if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			logger.Error("service failed", "error", err)
		}
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	_ = consumer.Close()
	_ = redisClient.Close()
	log.Print("notification service stopped")
}
