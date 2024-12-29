package api

import (
	"net/http"

	gosssError "github.com/mmvergara/gosss/internal/error"
	"github.com/mmvergara/gosss/internal/storage"
)

func NewRouter(store storage.Storage) *http.ServeMux {
	r := http.NewServeMux()
	h := NewHandler(store)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", "/")
	})

	// Bucket operations
	r.HandleFunc("PUT /{bucket}", h.CreateBucket)
	r.HandleFunc("DELETE /{bucket}", h.DeleteBucket)
	r.HandleFunc("HEAD /{bucket}", h.HeadBucket)

	// Object operations
	r.HandleFunc("PUT /{bucket}/{key...}", h.PutObject)
	r.HandleFunc("GET /{bucket}/{key...}", h.GetObject)
	r.HandleFunc("DELETE /{bucket}/{key...}", h.DeleteObject)
	r.HandleFunc("GET /{bucket}", h.ListObjects)
	r.HandleFunc("HEAD /{bucket}/{key...}", h.HeadObject)

	return r
}
