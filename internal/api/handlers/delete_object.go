package handlers

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

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
