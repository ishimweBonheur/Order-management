package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ishimweBonheur/order-management/product-service/internal/config"
	"github.com/ishimweBonheur/order-management/product-service/internal/database"
	"github.com/ishimweBonheur/order-management/product-service/internal/handler"
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

	ctx := context.Background()

	dbPool, err := database.NewPool(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	productRepository := repository.NewPostgresProductRepository(dbPool)

	productService := service.NewProductService(
		productRepository,
	)

	productHandler := handler.NewProductHandler(
		productService,
	)

	router := chi.NewRouter()
	healthHandler := handler.NewHealthHandler(dbPool)

	router.Get("/health", healthHandler.Health)

	router.Route("/products", func(r chi.Router) {
		r.Post("/", productHandler.CreateProduct)
		r.Get("/", productHandler.GetProducts)
		r.Get("/{id}", productHandler.GetProduct)
		r.Put("/{id}", productHandler.UpdateProduct)
		r.Delete("/{id}", productHandler.DeleteProduct)
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
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	go func() {
		log.Printf("Product service running on :%s", cfg.Port)
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
			log.Fatal(err)
		}

	case <-shutdown:
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
