package model

import (
	"time"
)

type ListBucketResult struct {
	Name     string           `json:"name"`
	Prefix   string           `json:"prefix"`
	Contents []ObjectMetadata `json:"contents"`
}

type ObjectMetadata struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
}
