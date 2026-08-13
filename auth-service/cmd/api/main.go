package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/ishimweBonheur/order-management/auth-service/internal/config"
	"github.com/ishimweBonheur/order-management/auth-service/internal/handler"
	sharedauth "github.com/ishimweBonheur/order-management/auth-service/internal/platform/auth"
	"github.com/ishimweBonheur/order-management/auth-service/internal/platform/database"
	"github.com/ishimweBonheur/order-management/auth-service/internal/platform/httpserver"
	"github.com/ishimweBonheur/order-management/auth-service/internal/repository"
	"github.com/ishimweBonheur/order-management/auth-service/internal/service"
	"log"
	"log/slog"
	"net/http"
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
	h := handler.New(service.New(repository.NewPostgres(db), cfg.JWTSecret, cfg.TokenTTL))
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID, chimiddleware.RealIP, chimiddleware.Recoverer)
	r.Use(httpserver.RequestLogger("auth-service", logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:8080", "http://localhost:3000", "http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))
	r.Get("/health", handler.Health)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.With(sharedauth.Middleware(cfg.JWTSecret)).Get("/me", h.Me)
	})
	r.With(sharedauth.Middleware(cfg.JWTSecret), sharedauth.RequireRole("admin")).Put("/admin/users/{id}/role", h.AssignRole)
	if err = httpserver.Run(cfg.Port, "auth-service", r, logger, func() { db.Close() }); err != nil {
		log.Fatal(err)
	}
	_ = http.MethodGet
}
