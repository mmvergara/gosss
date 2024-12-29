package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

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
