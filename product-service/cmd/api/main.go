package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/ishimweBonheur/order-management/product-service/internal/database"
	"github.com/ishimweBonheur/order-management/product-service/internal/handler"
	"github.com/ishimweBonheur/order-management/product-service/internal/repository"
	"github.com/ishimweBonheur/order-management/product-service/internal/service"
	"github.com/joho/godotenv"

)

func main() {
	err := godotenv.Load("product-service/.env")
	if err != nil {
		log.Println("warning: .env file not found")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set in the environment variables")
	}

	ctx := context.Background()

	dbPool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPool.Close()

	productRepository := repository.NewPostgresProductRepository(dbPool)
	productService := service.NewProductService(productRepository)

	productHandler := handler.NewProductHandler(productService)

	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	router.Route("/products", func(r chi.Router) {
		r.Post("/", productHandler.CreateProduct)
		r.Get("/", productHandler.GetProducts)
		r.Get("/{id}", productHandler.GetProduct)
		r.Put("/{id}", productHandler.UpdateProduct)
		r.Delete("/{id}", productHandler.DeleteProduct)
	})

	fmt.Println("Product service running on :8080")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("server error:", err)
	}
}