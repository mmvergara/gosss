package handlers

import (
	"sync"
	"time"

	"github.com/mmvergara/gosss/internal/config"
	"github.com/mmvergara/gosss/internal/storage"
)

const (
	MaxFileSize = 10 * 1024 * 1024 * 1024 // 10GB

	MaxConcurrent  = 100
	RequestTimeout = 30 * time.Second
)

var (
	semaphore = make(chan struct{}, MaxConcurrent)
)

type Handler struct {
	store  storage.Storage
	mutex  sync.RWMutex
	config *config.Config
}

func NewHandler(store storage.Storage, config *config.Config) *Handler {
	return &Handler{
		store:  store,
		config: config,
		mutex:  sync.RWMutex{},
	}
}
