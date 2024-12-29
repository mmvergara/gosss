package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
	"github.com/mmvergara/gosss/internal/model"
)

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
