package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	gosssError "github.com/mmvergara/gosss/internal/error"
	"github.com/mmvergara/gosss/internal/model"
	"github.com/mmvergara/gosss/internal/storage"
)

type Handler struct {
	store storage.Storage
}

func NewHandler(store storage.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) CreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	if err := h.store.CreateBucket(r.Context(), bucket); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", "")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) PutObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	err := h.store.PutObject(r.Context(), bucket, key, r.Body, r.ContentLength, r.Header.Get("Content-Type"))
	if err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	obj, metadata, err := h.store.GetObject(r.Context(), bucket, key)
	if err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusNotFound, "Object not found", bucket+"/"+key)
		return
	}
	defer obj.Close()

	// Read the first 512 bytes to detect Content-Type
	buffer := make([]byte, 512)
	n, err := obj.Read(buffer)

	if err != nil && err != io.EOF {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Error reading file", bucket+"/"+key)
		return
	}
	contentType := http.DetectContentType(buffer[:n])

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("ETag", metadata.ETag)
	w.Header().Set("Last-Modified", metadata.LastModified.Format(http.TimeFormat))

	w.Write(buffer[:n])
	if _, err := io.Copy(w, obj); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}
}

func (h *Handler) DeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")

	if err := h.store.DeleteBucket(r.Context(), bucket); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

	if err := h.store.DeleteObject(r.Context(), bucket, key); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListObjects(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
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
		result.Contents = append(result.Contents, model.Object{
			Key:          obj.Key,
			LastModified: obj.LastModified.Format(time.RFC3339),
			ETag:         obj.ETag,
			Size:         obj.Size,
		})
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)

	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(result); err != nil {
		log.Println(err)
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket)
		return
	}
}

func (h *Handler) HeadObject(w http.ResponseWriter, r *http.Request) {
	bucket := r.PathValue("bucket")
	key := r.PathValue("key")

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
