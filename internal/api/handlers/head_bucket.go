package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

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
