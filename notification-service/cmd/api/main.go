package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"

	"github.com/ishimweBonheur/order-management/notification-service/internal/config"
	"github.com/ishimweBonheur/order-management/notification-service/internal/email"
	"github.com/ishimweBonheur/order-management/notification-service/internal/messaging"
	"github.com/ishimweBonheur/order-management/notification-service/internal/notification"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "notification-service")
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Redis unavailable", "error", err)
		os.Exit(1)
	}
	sender, err := email.NewGoMailSender(cfg)
	if err != nil {
		logger.Error("SMTP configuration invalid", "error", err)
		os.Exit(1)
	}
	service := notification.New(redisClient, sender, logger)
	reader := kafka.NewReader(kafka.ReaderConfig{Brokers: cfg.KafkaBrokers, Topic: cfg.KafkaTopic, GroupID: cfg.KafkaGroup, MinBytes: 1, MaxBytes: 10e6, MaxWait: time.Second})
	defer reader.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := redisClient.Ping(r.Context()).Err(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"degraded","redis":"unavailable"}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"ok","redis":"ok"}`))
	})
	server := &http.Server{Addr: ":" + cfg.HTTPPort, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		logger.Info("HTTP server started", "port", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server failed", "error", err)
			stop()
		}
	}()
	go func() {
		if err := messaging.Consume(ctx, reader, service, logger); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Kafka consumer stopped", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	logger.Info("notification service stopped")
}
