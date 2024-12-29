package main

import (
	"log"
	"net/http"

	"github.com/mmvergara/gosss/internal/api"
	"github.com/mmvergara/gosss/internal/config"
	"github.com/mmvergara/gosss/internal/middleware"
	"github.com/mmvergara/gosss/internal/storage"
)

func main() {
	// Initialize configuration
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Initialize storage backend
	store := storage.New(cfg.StoragePath)

	// Setup authentication middleware
	// authMiddleware := middleware.NewAuthMiddleware(cfg)

	// Create the final handler with middleware
	// handler := authMiddleware.Authenticate(router)

	// listenserver(handler)

	// Setup API handlers
	router := api.NewRouter(store)
	handler := middleware.CorsMiddleware(router)
	handler = middleware.LoggerMiddleware(handler)
	handler = middleware.Auth(handler, cfg)

	// Start server
	log.Printf("Starting server on %s", cfg.ListenAddr)
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
