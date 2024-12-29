package storage

import (
	"context"
	"io"

	"github.com/mmvergara/gosss/internal/model"
)

// Storage defines the interface for storage operations
type Storage interface {
	// Bucket operations
	CreateBucket(ctx context.Context, name string) error
	DeleteBucket(ctx context.Context, name string) error
	BucketExists(ctx context.Context, name string) (bool, error)

	// Object operations
	PutObject(ctx context.Context, bucket, key string, data io.Reader, size int64, contentType string) (*model.ObjectMetadata, error)
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, *model.ObjectMetadata, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	ListObjects(ctx context.Context, bucket, prefix string) ([]model.ObjectMetadata, error)
	HasObject(ctx context.Context, bucket string) (bool, error)
	HeadObject(ctx context.Context, bucket, key string) (*model.ObjectMetadata, error)
}
