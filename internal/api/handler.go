package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mmvergara/gosss/internal/config"
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

func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

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
	bucket := chi.URLParam(r, "bucket")

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
	bucket := chi.URLParam(r, "bucket")

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

	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

	// Validate bucket name
	isValidBuckName, msg := isValidBucketName(bucket)
	if !isValidBuckName {
		log.Printf("Invalid bucket name: %s (%s)", bucket, msg)
		gosssError.SendGossError(w, http.StatusBadRequest, msg, bucket)
		return
	}

	log.Println(bucket, key)

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

func (h *Handler) GetSignedObject(w http.ResponseWriter, r *http.Request, bucket string, key string) {
	// Validate query parameters
	expiration := r.URL.Query().Get("expiration")
	signature := r.URL.Query().Get("signature")
	if expiration == "" || signature == "" {
		gosssError.SendGossError(w, http.StatusBadRequest, "Missing required query parameters", "expiration and signature required")
		return
	}

	// Parse and validate expiration
	exp, err := strconv.ParseInt(expiration, 10, 64)
	if err != nil {
		gosssError.SendGossError(w, http.StatusBadRequest, "Invalid expiration format", "")
		return
	}

	// Check if URL has expired
	if time.Now().Unix() > exp {
		gosssError.SendGossError(w, http.StatusForbidden, "URL has expired", "")
		return
	}

	// Verify signature
	expectedSignature, err := h.generateSignature(expiration)
	if err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Error verifying signature", "")
		return
	}

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		gosssError.SendGossError(w, http.StatusForbidden, "Invalid signature", "")
		return
	}

	// If signature is valid, proceed with getting the object
	obj, metadata, err := h.store.GetObject(r.Context(), bucket, key)
	if err != nil {
		gosssError.SendGossError(w, http.StatusNotFound, "Object not found", bucket+"/"+key)
		return
	}
	defer obj.Close()

	// Set response headers
	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("ETag", metadata.ETag)
	w.Header().Set("Last-Modified", metadata.LastModified.Format(http.TimeFormat))

	// Stream the object to the response
	if _, err := io.Copy(w, obj); err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}
}

// generateSignature creates an HMAC-SHA256 signature for the given expiration timestamp
func (h *Handler) generateSignature(expiration string) (string, error) {
	mac := hmac.New(sha256.New, []byte(h.config.SecretKey))
	mac.Write([]byte(expiration))
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature, nil
}

func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

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
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

	if err := h.store.DeleteObject(r.Context(), bucket, key); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, err.Error(), bucket+"/"+key)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	log.Println("ListObjects")
	bucket := chi.URLParam(r, "bucket")
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
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

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
