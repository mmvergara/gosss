package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

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
