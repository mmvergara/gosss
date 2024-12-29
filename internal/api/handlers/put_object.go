package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

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
