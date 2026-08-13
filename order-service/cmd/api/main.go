package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/ishimweBonheur/order-management/order-service/internal/config"
	"github.com/ishimweBonheur/order-management/order-service/internal/handler"
	"github.com/ishimweBonheur/order-management/order-service/internal/messaging"
	sharedauth "github.com/ishimweBonheur/order-management/order-service/internal/platform/auth"
	"github.com/ishimweBonheur/order-management/order-service/internal/platform/database"
	"github.com/ishimweBonheur/order-management/order-service/internal/platform/httpserver"
	"github.com/ishimweBonheur/order-management/order-service/internal/repository"
	"github.com/ishimweBonheur/order-management/order-service/internal/service"
	"log"
	"log/slog"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, err := database.Open(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	producer := messaging.New(cfg.KafkaBrokers, cfg.KafkaTopic)
	h := handler.New(service.New(repository.NewPostgres(db), producer))
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:8080", "http://localhost:3000", "http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))
	r.Get("/health", handler.Health)
	r.Group(func(r chi.Router) {
		r.Use(sharedauth.Middleware(cfg.JWTSecret))
		r.Post("/orders", h.Create)
		r.Get("/orders", h.List)
		r.Get("/orders/{id}", h.Get)
		r.Route("/admin", func(r chi.Router) {
			r.Use(sharedauth.RequireRole("admin"))
			r.Get("/orders", h.List)
			r.Put("/orders/{id}/status", h.Status)
		})
	})
	if err = httpserver.Run(cfg.Port, "order-service", r, logger, func() { _ = producer.Close(); db.Close() }); err != nil {
		log.Fatal(err)
	}
}
