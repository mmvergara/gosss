package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/mmvergara/gosss/internal/config"
	"github.com/mmvergara/gosss/internal/middleware"
	"github.com/mmvergara/gosss/internal/storage"
)

func NewRouter(store storage.Storage, cfg *config.Config) *chi.Mux {
	h := NewHandler(store, cfg)
	r := chi.NewRouter()
	r.Use(middleware.CorsMiddleware)
	r.Use(middleware.LoggerMiddleware)

	r.Route("/presign", func(r chi.Router) {
		r.Get("/{bucket}/*", h.GetObject)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.CreateAuthMiddleware(cfg))
		r.Put("/{bucket}", h.CreateBucket)
		r.Delete("/{bucket}", h.DeleteBucket)
		r.Head("/{bucket}", h.HeadBucket)
		r.Get("/{bucket}/*", h.GetObject)
		r.Put("/{bucket}/*", h.PutObject)
		r.Delete("/{bucket}/*", h.DeleteObject)
		r.Get("/{bucket}", h.ListObjects)
		r.Head("/{bucket}/*", h.HeadObject)
	})
	return r
}
