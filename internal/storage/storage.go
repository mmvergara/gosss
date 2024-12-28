package storage

import (
	"context"
	"io"
	"time"
)

// Object represents a stored object's metadata
type Object struct {
	Key          string
	Size         int64
	LastModified time.Time
	ETag         string
	ContentType  string
}

// Storage defines the interface for storage operations
type Storage interface {
	// Bucket operations
	CreateBucket(ctx context.Context, name string) error
	DeleteBucket(ctx context.Context, name string) error
	BucketExists(ctx context.Context, name string) (bool, error)

	// Object operations
	PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *Object, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket, prefix string) ([]Object, error)
	HeadObject(ctx context.Context, bucket, key string) (*Object, error)
}
