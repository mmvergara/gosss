package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ObjectMetadata represents the metadata of a stored object
type ObjectMetadata struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
}

func (ls *LocalStorage) PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Create full path for object and metadata
	objectPath := filepath.Join(ls.basePath, bucket, key)
	metadataPath := objectPath + ".metadata"

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(objectPath), 0755); err != nil {
		log.Printf("Failed to create directories: %v", err)
		return fmt.Errorf("failed to create directories")
	}

	// Create temporary file for object data
	tempFile, err := os.CreateTemp(filepath.Dir(objectPath), "tmp-")
	if err != nil {
		log.Printf("Failed to create temporary file: %v", err)
		return fmt.Errorf("failed to create temporary file")
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up temp file in case of error

	// Calculate ETag (MD5) while copying data
	hash := md5.New()
	writer := io.MultiWriter(tempFile, hash)

	written, err := io.Copy(writer, data)
	if err != nil {
		tempFile.Close()
		log.Printf("Failed to write data: %v", err)
		return fmt.Errorf("failed to write data")
	}
	tempFile.Close()

	// Create metadata
	metadata := ObjectMetadata{
		Key:          key,
		Size:         written,
		LastModified: time.Now().UTC(),
		ETag:         `"` + hex.EncodeToString(hash.Sum(nil)) + `"`,
		ContentType:  contentType,
	}

	// Write metadata to temporary file
	metadataTempFile, err := os.CreateTemp(filepath.Dir(metadataPath), "tmp-metadata-")
	if err != nil {
		log.Printf("Failed to create temporary metadata file: %v", err)
		return fmt.Errorf("failed to create temporary metadata file")
	}
	metadataTempPath := metadataTempFile.Name()
	defer os.Remove(metadataTempPath) // Clean up temp metadata file in case of error

	if err := json.NewEncoder(metadataTempFile).Encode(metadata); err != nil {
		metadataTempFile.Close()
		log.Printf("Failed to write metadata: %v", err)
		return fmt.Errorf("failed to write metadata")
	}
	metadataTempFile.Close()

	// Atomically move files into place
	if err := os.Rename(tempPath, objectPath); err != nil {
		log.Printf("Failed to move object file: %v", err)
		return fmt.Errorf("failed to move object file")
	}
	if err := os.Rename(metadataTempPath, metadataPath); err != nil {
		// Try to clean up object file if metadata move fails
		os.Remove(objectPath)
		log.Printf("Failed to move metadata file: %v", err)
		return fmt.Errorf("failed to move metadata file")
	}

	return nil
}

func (ls *LocalStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *Object, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	metadataPath := objectPath + ".metadata"

	// Read metadata first
	metadata, err := ls.readMetadata(metadataPath)
	if err != nil {
		log.Printf("Failed to read metadata: %v", err)
		return nil, nil, fmt.Errorf("failed to read metadata")
	}

	// Open the object file
	file, err := os.Open(objectPath)
	if err != nil {
		log.Printf("Failed to open file: %v", err)
		return nil, nil, fmt.Errorf("failed to open file")
	}

	obj := &Object{
		Key:          metadata.Key,
		Size:         metadata.Size,
		LastModified: metadata.LastModified,
		ETag:         metadata.ETag,
		ContentType:  metadata.ContentType,
	}

	return file, obj, nil
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

		// Skip directories and metadata files
		if info.IsDir() || strings.HasSuffix(path, ".metadata") {
			return nil
		}

		// Get relative path and check prefix
		relPath, _ := filepath.Rel(bucketPath, path)
		if prefix == "" || strings.HasPrefix(relPath, prefix) {
			// Read metadata for this object
			metadata, err := ls.readMetadata(path + ".metadata")
			if err != nil {
				// Log error but continue processing other files
				fmt.Printf("Warning: failed to read metadata for %s: %v\n", relPath, err)
				return nil
			}

			objects = append(objects, Object{
				Key:          metadata.Key,
				Size:         metadata.Size,
				LastModified: metadata.LastModified,
				ETag:         metadata.ETag,
				ContentType:  metadata.ContentType,
			})
		}
		return nil
	})

	if err != nil {
		log.Printf("Failed to list objects: %v", err)
		return nil, fmt.Errorf("failed to list objects")
	}

	return objects, nil
}

func (ls *LocalStorage) HeadObject(ctx context.Context, bucket, key string) (*Object, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	metadataPath := objectPath + ".metadata"

	metadata, err := ls.readMetadata(metadataPath)
	if err != nil {
		log.Printf("Failed to read metadata: %v", err)
		return nil, fmt.Errorf("failed to read metadata")
	}

	return &Object{
		Key:          metadata.Key,
		Size:         metadata.Size,
		LastModified: metadata.LastModified,
		ETag:         metadata.ETag,
		ContentType:  metadata.ContentType,
	}, nil
}

// Helper function to read metadata from file
func (ls *LocalStorage) readMetadata(path string) (*ObjectMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata ObjectMetadata
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}

// DeleteObject should also delete the metadata file
func (ls *LocalStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	objectPath := filepath.Join(ls.basePath, bucket, key)
	metadataPath := objectPath + ".metadata"

	// Delete both object and metadata files
	if err := os.Remove(objectPath); err != nil {
		log.Printf("Failed to delete object: %v", err)
		return fmt.Errorf("failed to delete object")
	}

	// Try to delete metadata file, but don't error if it doesn't exist
	_ = os.Remove(metadataPath)

	return nil
}
