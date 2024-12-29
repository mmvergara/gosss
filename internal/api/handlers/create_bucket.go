package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

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
