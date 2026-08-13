package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishimweBonheur/order-management/product-service/internal/cache"
	"github.com/ishimweBonheur/order-management/product-service/internal/config"
	"github.com/ishimweBonheur/order-management/product-service/internal/database"
	"github.com/ishimweBonheur/order-management/product-service/internal/handler"
	"github.com/ishimweBonheur/order-management/product-service/internal/messaging"
	"github.com/ishimweBonheur/order-management/product-service/internal/middleware"
	"github.com/ishimweBonheur/order-management/product-service/internal/repository"
	"github.com/ishimweBonheur/order-management/product-service/internal/service"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	_ = godotenv.Load("product-service/.env")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx := context.Background()

	dbPool, err := database.NewPool(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		log.Fatal(err)
	}
	defer dbPool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	defer redisClient.Close()

	productRepository := repository.NewPostgresProductRepository(dbPool)

	productCache := cache.NewProductCache(
		redisClient,
		cfg.RedisCacheTTL,
	)

	producer := messaging.NewProducer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
	)
	defer producer.Close()

	productService := service.NewProductService(
		productRepository,
		productCache,
		producer,
	)

	productHandler := handler.NewProductHandler(
		productService,
	)

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTSecret)

	router := chi.NewRouter()

	healthHandler := handler.NewHealthHandler(dbPool)

	middleware.Setup(router)

	router.Use(middleware.RequestLogger(logger))

	rateLimiter := middleware.NewRedisRateLimiter(
		redisClient,
		100,
		time.Minute,
	)

	router.Get("/health", healthHandler.Health)

	router.Route("/products", func(r chi.Router) {
		r.Use(rateLimiter.Middleware)

		// Public endpoints
		r.Group(func(r chi.Router) {
			r.Get("/", productHandler.GetProducts)
			r.Get("/{id}", productHandler.GetProduct)
		})

		// Admin-only endpoints
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(middleware.RequireRole("admin"))

			r.Post("/", productHandler.CreateProduct)
			r.Put("/{id}", productHandler.UpdateProduct)
			r.Delete("/{id}", productHandler.DeleteProduct)
		})
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("product service running", "port", cfg.Port, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)

	signal.Notify(
		shutdown,
		os.Interrupt,
		syscall.SIGTERM,
	)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			log.Fatal(err)
		}

	case <-shutdown:
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
}
