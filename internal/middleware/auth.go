package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/mmvergara/gosss/internal/config"
	gosssError "github.com/mmvergara/gosss/internal/error"
)

// createAuthMiddleware takes a config and returns a middleware function.
func CreateAuthMiddleware(config *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for browser preflight requests
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("Authorization header required")
				gosssError.SendGossError(w, http.StatusUnauthorized, "Authorization header required", "")
				return
			}

			// Extract credentials from authorization header
			parts := strings.Split(authHeader, "=")
			if len(parts) < 2 {
				log.Println("Invalid authorization header format")
				gosssError.SendGossError(w, http.StatusUnauthorized, "Invalid authorization header format", "")
				return
			}

			accessKeyID := parts[0]
			if accessKeyID != config.AccessKeyID {
				log.Println("Invalid access key ID")
				gosssError.SendGossError(w, http.StatusUnauthorized, "Invalid access key ID", "")
				return
			}

			secretAccessKey := parts[1]
			if secretAccessKey != config.SecretKey {
				log.Println("Invalid secret access key")
				gosssError.SendGossError(w, http.StatusUnauthorized, "Invalid secret access key", "")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
