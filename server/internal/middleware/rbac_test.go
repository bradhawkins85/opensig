package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/opensig/server/internal/models"
)

func TestRequireRole_Authorized(t *testing.T) {
	user := &models.User{
		ID:       "user-1",
		Email:    "admin@test.com",
		TenantID: "tenant-1",
		Role:     models.RoleOrgAdmin,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := RequireRole(models.RoleOrgAdmin)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRequireRole_Forbidden(t *testing.T) {
	user := &models.User{
		ID:       "user-1",
		Email:    "auditor@test.com",
		TenantID: "tenant-1",
		Role:     models.RoleAuditor,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := RequireRole(models.RoleOrgAdmin)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, user)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireRole_Unauthorized(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	middleware := RequireRole(models.RoleOrgAdmin)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	tests := []struct {
		name           string
		userRole       models.Role
		allowedRoles   []models.Role
		expectedStatus int
	}{
		{
			name:           "org_admin allowed",
			userRole:       models.RoleOrgAdmin,
			allowedRoles:   []models.Role{models.RoleOrgAdmin, models.RoleSignatureAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "signature_admin allowed",
			userRole:       models.RoleSignatureAdmin,
			allowedRoles:   []models.Role{models.RoleOrgAdmin, models.RoleSignatureAdmin},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "auditor not allowed",
			userRole:       models.RoleAuditor,
			allowedRoles:   []models.Role{models.RoleOrgAdmin, models.RoleSignatureAdmin},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				ID:       "user-1",
				Email:    "test@test.com",
				TenantID: "tenant-1",
				Role:     tt.userRole,
			}

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequireRole(tt.allowedRoles...)(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			ctx := context.WithValue(req.Context(), UserContextKey, user)
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestMockAuthMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserContextKey).(*models.User)
		if user == nil {
			t.Error("expected user to be set in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := MockAuthMiddleware(nil)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-ID", "test-user")
	req.Header.Set("X-User-Role", "org_admin")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestMockAuthMiddleware_DefaultUser(t *testing.T) {
	defaultUser := &models.User{
		ID:       "default",
		Email:    "default@test.com",
		TenantID: "tenant-1",
		Role:     models.RoleAuditor,
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserContextKey).(*models.User)
		if user.ID != defaultUser.ID {
			t.Errorf("expected user ID %s, got %s", defaultUser.ID, user.ID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := MockAuthMiddleware(defaultUser)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
