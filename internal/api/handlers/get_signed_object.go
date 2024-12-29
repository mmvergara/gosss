package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

// generateSignature creates an HMAC-SHA256 signature for the given parameters
func (h *Handler) generateSignature(expiration, bucket, key string) (string, error) {
	// Create string to sign in same format as client
	stringToSign := strings.Join([]string{expiration, bucket, key}, ":")

	mac := hmac.New(sha256.New, []byte(h.config.SecretKey))
	mac.Write([]byte(stringToSign))
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature, nil
}

func (h *Handler) GetSignedObject(w http.ResponseWriter, r *http.Request) {
	// Validate query parameters
	expiration := r.URL.Query().Get("expiration")
	signature := r.URL.Query().Get("signature")
	bucket := chi.URLParam(r, "bucket")
	key := chi.URLParam(r, "*")

	if expiration == "" || signature == "" {
		gosssError.SendGossError(w, http.StatusBadRequest, "Missing required query parameters", "expiration and signature required")
		return
	}

	// Parse and validate expiration
	exp, err := strconv.ParseInt(expiration, 10, 64)
	if err != nil {
		gosssError.SendGossError(w, http.StatusBadRequest, "Invalid expiration format", "")
		return
	}

	// Check if URL has expired
	if time.Now().Unix() > exp {
		gosssError.SendGossError(w, http.StatusForbidden, "URL has expired", "")
		return
	}

	// Verify signature using bucket and key in the signature generation
	expectedSignature, err := h.generateSignature(expiration, bucket, key)
	if err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Error verifying signature", "")
		return
	}

	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		gosssError.SendGossError(w, http.StatusForbidden, "Invalid signature", "")
		return
	}

	// If signature is valid, proceed with getting the object
	obj, metadata, err := h.store.GetObject(r.Context(), bucket, key)
	if err != nil {
		gosssError.SendGossError(w, http.StatusNotFound, "Object not found", bucket+"/"+key)
		return
	}
	defer obj.Close()

	// Set response headers
	w.Header().Set("Content-Type", metadata.ContentType)
	w.Header().Set("ETag", metadata.ETag)
	w.Header().Set("Last-Modified", metadata.LastModified.Format(http.TimeFormat))

	// Stream the object to the response
	if _, err := io.Copy(w, obj); err != nil {
		gosssError.SendGossError(w, http.StatusInternalServerError, "Internal server error", bucket+"/"+key)
		return
	}
}
