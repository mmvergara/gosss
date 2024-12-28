package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type LocalStorage struct {
	basePath string
	mu       sync.RWMutex
}

func New(basePath string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
	}
}

func (ls *LocalStorage) CreateBucket(ctx context.Context, name string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	bucketPath := filepath.Join(ls.basePath, name)
	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		log.Printf("Failed to create bucket: %v", err)
		return fmt.Errorf("failed to create bucket")
	}
	return nil
}

func (ls *LocalStorage) DeleteBucket(ctx context.Context, name string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	bucketPath := filepath.Join(ls.basePath, name)
	// Check if bucket is empty
	entries, err := os.ReadDir(bucketPath)
	if err != nil {
		log.Printf("Failed to read bucket: %v", err)
		return fmt.Errorf("failed to read bucket")
	}
	if len(entries) > 0 {
		log.Printf("Bucket not empty: %s", name)
		return fmt.Errorf("bucket not empty")
	}

	if err := os.Remove(bucketPath); err != nil {
		log.Printf("Failed to delete bucket: %v", err)
		return fmt.Errorf("failed to delete bucket")
	}
	return nil
}

func (ls *LocalStorage) BucketExists(ctx context.Context, name string) (bool, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	bucketPath := filepath.Join(ls.basePath, name)
	_, err := os.Stat(bucketPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		log.Printf("Failed to check bucket: %v", err)
		return false, fmt.Errorf("failed to check bucket")
	}
	return true, nil
}
