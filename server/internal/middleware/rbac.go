package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/your-org/opensig/server/internal/models"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// RequireRole creates middleware that enforces role-based access control
func RequireRole(allowedRoles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user from context (set by auth middleware)
			user, ok := r.Context().Value(UserContextKey).(*models.User)
			if !ok || user == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Check if user's role is in the allowed roles
			hasRole := false
			for _, role := range allowedRoles {
				if user.Role == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "forbidden: insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// MockAuthMiddleware simulates authentication for demo purposes
// In production, this would validate JWT tokens or session cookies
func MockAuthMiddleware(defaultUser *models.User) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check for a mock user header for testing
			userID := r.Header.Get("X-User-ID")
			role := r.Header.Get("X-User-Role")

			var user *models.User
			if userID != "" && role != "" {
				user = &models.User{
					ID:       userID,
					Email:    userID + "@example.com",
					TenantID: "test-tenant",
					Role:     models.Role(role),
				}
			} else if defaultUser != nil {
				user = defaultUser
			}

			if user != nil {
				ctx := context.WithValue(r.Context(), UserContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
