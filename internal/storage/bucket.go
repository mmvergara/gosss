package storage

import (
	"context"
	"fmt"
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
		return fmt.Errorf("failed to create bucket: %w", err)
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
		return fmt.Errorf("failed to read bucket: %w", err)
	}
	if len(entries) > 0 {
		return fmt.Errorf("bucket not empty")
	}

	if err := os.Remove(bucketPath); err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
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
		return false, fmt.Errorf("failed to check bucket: %w", err)
	}
	return true, nil
}
