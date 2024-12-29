package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	gosssError "github.com/mmvergara/gosss/internal/error"
	"github.com/mmvergara/gosss/internal/model"
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
	store storage.Storage
	mutex sync.RWMutex
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{
		store: store,
		mutex: sync.RWMutex{},
	}
}

func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	// Check if bucket already exists
	exists, err := h.store.BucketExists(r.Context(), bucket)
	if err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to check bucket", bucket)
		return
	}
	if exists {
		gosssError.SendGossError(w, http.StatusConflict, "Bucket already exists", bucket)
		return
	}

	// Validate bucket name
	isValidBuckName, msg := isValidBucketName(bucket)
	if !isValidBuckName {
		log.Printf("Invalid bucket name: %s (%s)", bucket, msg)
		gosssError.SendGossError(w, http.StatusBadRequest, msg, bucket)
		return
	}

	if err := h.store.CreateBucket(r.Context(), bucket); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, err.Error(), "")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	// Check if bucket exists
	exists, err := h.store.BucketExists(r.Context(), bucket)
	if err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to check bucket", bucket)
		return
	}
	if !exists {
		gosssError.SendGossError(w, http.StatusNotFound, "Bucket not found", bucket)
		return
	}

	// List objects to ensure bucket is empty
	hasObject, err := h.store.HasObject(r.Context(), bucket)
	if err != nil {
		log.Printf("Failed to check if bucket is empty: %v", err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to list objects", bucket)
		return
	}
	if hasObject {
		log.Printf("Bucket not empty: %s", bucket)
		gosssError.SendGossError(w, http.StatusConflict, "Bucket not empty", bucket)
	}

	// Delete bucket with retry
	for i := 0; i < 3; i++ {
		err = h.store.DeleteBucket(r.Context(), bucket)
		if err == nil {
			break
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
	if err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to delete bucket", bucket)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HeadBucket(w http.ResponseWriter, r *http.Request) {
	log.Println("HEAD /{bucket}")
	bucket := r.PathValue("bucket")

	exists, err := h.store.BucketExists(r.Context(), bucket)
	if err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to check bucket", bucket)
		return
	}
	if !exists {
		gosssError.SendGossError(w, http.StatusNotFound, "Bucket not found", bucket)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PutObject(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), RequestTimeout)
	defer cancel()

	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	// Validate bucket name
	isValidBuckName, msg := isValidBucketName(bucket)
	if !isValidBuckName {
		log.Printf("Invalid bucket name: %s (%s)", bucket, msg)
		gosssError.SendGossError(w, http.StatusBadRequest, msg, bucket)
		return
	}

	// Validate object key
	isValidObjKey, msg := isValidObjectKey(key)
	if !isValidObjKey {
		log.Printf("Invalid object key: %s (%s)", key, msg)
		gosssError.SendGossError(w, http.StatusBadRequest, msg, bucket+"/"+key)
		return
	}

	// Check file size (optional warning log)
	if r.ContentLength > MaxFileSize {
		log.Printf("Warning: File size is %d bytes, exceeding the maximum allowed size of %d bytes.\n", r.ContentLength, MaxFileSize)
	}

	select {
	case semaphore <- struct{}{}:
		defer func() { <-semaphore }()
	default:
		gosssError.SendGossError(w, http.StatusTooManyRequests, "Too many concurrent requests", "")
		return
	}

	// Directly stream the data from the request body to the storage backend
	metadata, err := h.store.PutObject(ctx, bucket, key, r.Body, r.ContentLength, r.Header.Get("Content-Type"))
	if err != nil {
		log.Printf("Failed to store object: %v", err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to store object", bucket+"/"+key)
		return
	}
	err = json.NewEncoder(w).Encode(metadata)
	if err != nil {
		log.Printf("Failed to encode metadata: %v", err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Failed to encode metadata", bucket+"/"+key)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	obj, metadata, err := h.store.GetObject(r.Context(), bucket, key)
	if err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusNotFound, "Object not found", bucket+"/"+key)
		return
	}
	defer obj.Close()

	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("ETag", metadata.ETag)
	w.Header().Set("Last-Modified", metadata.LastModified.Format(http.TimeFormat))

	if _, err := io.Copy(w, obj); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}
}

func (h *Handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	if err := h.store.DeleteObject(r.Context(), bucket, key); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	log.Println("ListObjects")
	bucket := r.PathValue("bucket")
	prefix := r.URL.Query().Get("prefix")

	objects, err := h.store.ListObjects(r.Context(), bucket, prefix)
	if err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Something went wrong or the bucket does not exist", bucket)
		return
	}

	result := model.ListBucketResult{
		Name:   bucket,
		Prefix: prefix,
	}

	for _, obj := range objects {
		result.Contents = append(result.Contents, model.ObjectMetadata{
			Key:          obj.Key,
			LastModified: obj.LastModified,
			ETag:         obj.ETag,
			Size:         obj.Size,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (h *Handler) HeadObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	metadata, err := h.store.HeadObject(r.Context(), bucket, key)
	if err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusNotFound, "Object not found", bucket+"/"+key)
		return
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", metadata.Size))
	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("ETag", metadata.ETag)
	w.Header().Set("Last-Modified", metadata.LastModified.Format(http.TimeFormat))

	w.WriteHeader(http.StatusOK)
}
