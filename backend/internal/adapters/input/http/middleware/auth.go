package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/meet-clone/backend/internal/pkg/jwt"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	jwtService *jwt.JWTService
}

func NewAuthMiddleware(jwtService *jwt.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for cookie first
		cookie, err := r.Cookie("access_token")
		var token string
		if err == nil {
			token = cookie.Value
		} else {
			// Fallback to Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			log.Println("AuthMiddleware: Missing authorization token") // DEBUG
			http.Error(w, "missing authorization token", http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			log.Printf("AuthMiddleware: Token validation failed: %v", err) // DEBUG
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetUserFromContext(ctx context.Context) (*jwt.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*jwt.Claims)
	return claims, ok
}
