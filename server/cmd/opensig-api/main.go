package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/your-org/opensig/server/internal/handlers"
	"github.com/your-org/opensig/server/internal/middleware"
	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

type Health struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(Health{Status: "ok", Time: time.Now().UTC().Format(time.RFC3339)})
}

func preview(w http.ResponseWriter, r *http.Request) {
	// Minimal stub: returns a rendered placeholder with fake user data
	type Resp struct {
		HTML string `json:"html"`
		Text string `json:"text"`
	}
	resp := Resp{
		HTML: `<div style="font-family:Segoe UI,Arial,sans-serif"><strong>Jane Doe</strong><br>Senior Engineer<br><img src="https://via.placeholder.com/120x40" alt="Logo"></div>`,
		Text: "Jane Doe\nSenior Engineer",
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// adminOnlyEndpoint demonstrates RBAC enforcement - only Org Admins can access
func adminOnlyEndpoint(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Admin access granted",
		"user":    user.Email,
		"role":    user.Role,
	})
}

func main() {
	// Initialize stores
	tenantStore := store.NewTenantStore()
	tenantHandler := handlers.NewTenantHandler(tenantStore)

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/v1/preview", preview)

	// Tenant CRUD endpoints - require Org Admin role
	mux.Handle("/v1/tenants", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleOrgAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					tenantHandler.CreateTenant(w, r)
				} else if r.Method == http.MethodGet {
					tenantHandler.ListTenants(w, r)
				} else {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	// Tenant GET/PUT/DELETE by ID - require Org Admin role
	mux.Handle("/v1/tenants/", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleOrgAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/v1/tenants/") {
					http.NotFound(w, r)
					return
				}
				switch r.Method {
				case http.MethodGet:
					tenantHandler.GetTenant(w, r)
				case http.MethodPut:
					tenantHandler.UpdateTenant(w, r)
				case http.MethodDelete:
					tenantHandler.DeleteTenant(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	// Sample RBAC-protected endpoint - only Org Admin can access
	mux.Handle("/v1/admin/config", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleOrgAdmin)(
			http.HandlerFunc(adminOnlyEndpoint),
		),
	))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)
	log.Printf("OpenSig API listening on %s", addr)
	log.Printf("RBAC roles: org_admin, signature_admin, approver, auditor")
	log.Printf("Use X-User-ID and X-User-Role headers for testing RBAC")
	log.Fatal(http.ListenAndServe(addr, mux))
}
