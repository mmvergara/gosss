package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func (ls *LocalStorage) PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	if err := os.MkdirAll(filepath.Dir(objectPath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	file, err := os.Create(objectPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	metadataPath := objectPath + ".metadata"
	metadataFile, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metadataFile.Close()

	if _, err := io.Copy(metadataFile, strings.NewReader(fmt.Sprintf("%d\n%s\n%s\n", size, contentType, calculateETag(objectPath)))); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func (ls *LocalStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *Object, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	file, err := os.Open(objectPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open file: %w", err)
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	obj := &Object{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		ETag:         calculateETag(objectPath),
	}

	return file, obj, nil
}

func (ls *LocalStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	if err := os.Remove(objectPath); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

func (ls *LocalStorage) ListObjects(ctx context.Context, bucket, prefix string) ([]Object, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	var objects []Object
	bucketPath := filepath.Join(ls.basePath, bucket)

	err := filepath.Walk(bucketPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(bucketPath, path)
			if prefix == "" || strings.HasPrefix(relPath, prefix) {
				objects = append(objects, Object{
					Key:          relPath,
					Size:         info.Size(),
					LastModified: info.ModTime(),
					ETag:         calculateETag(path),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return objects, nil
}

func (ls *LocalStorage) HeadObject(ctx context.Context, bucket, key string) (*Object, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	info, err := os.Stat(objectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get object info: %w", err)
	}

	return &Object{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		ETag:         calculateETag(objectPath),
	}, nil
}

func calculateETag(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}
