package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mmvergara/gosss/internal/api"
	"github.com/mmvergara/gosss/internal/config"
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

	// Setup API handlers
	router := api.NewRouter(store, cfg)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.PORT)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
