package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"time"

	gosssError "github.com/mmvergara/gosss/internal/error"
)

// generateSignature creates an HMAC-SHA256 signature for the given expiration timestamp
func (h *Handler) generateSignature(expiration string) (string, error) {
	mac := hmac.New(sha256.New, []byte(h.config.SecretKey))
	mac.Write([]byte(expiration))
	signature := hex.EncodeToString(mac.Sum(nil))
	return signature, nil
}

func (h *Handler) GetSignedObject(w http.ResponseWriter, r *http.Request, bucket string, key string) {
	// Validate query parameters
	expiration := r.URL.Query().Get("expiration")
	signature := r.URL.Query().Get("signature")
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

	// Verify signature
	expectedSignature, err := h.generateSignature(expiration)
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
