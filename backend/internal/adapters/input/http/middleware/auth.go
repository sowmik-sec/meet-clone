package middleware

import (
	"context"
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
		authHeader := r.Header.Get("Authorization")

		// DEBUG: Log the auth header
		if authHeader == "" {
			println("DEBUG: No Authorization header found")
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		println("DEBUG: Authorization header present:", authHeader[:min(20, len(authHeader))]+"...")

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			println("DEBUG: Invalid header format. Parts:", len(parts), "First part:", parts[0])
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		println("DEBUG: Extracted token, length:", len(token))

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			println("DEBUG: Token validation failed:", err.Error())
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		println("DEBUG: Token validated successfully. UserID:", claims.UserID)

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
