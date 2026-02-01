package middleware

import (
	"net/http"

	"github.com/meet-clone/backend/internal/config"
)

// SecurityHeaders adds security headers to all responses
func SecurityHeaders(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")

			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// Enable XSS protection (for older browsers)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy (adjust based on your needs)
			// This is a basic policy - customize for your application
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

			// Strict Transport Security (only in production with HTTPS)
			if cfg.Environment == "production" {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequestValidator validates common request properties
func RequestValidator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body size to 10MB
		r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)

		next.ServeHTTP(w, r)
	})
}
