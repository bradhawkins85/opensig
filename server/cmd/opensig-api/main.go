package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/your-org/opensig/server/internal/auth"
	"github.com/your-org/opensig/server/internal/handlers"
	"github.com/your-org/opensig/server/internal/middleware"
	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/renderer"
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
	// Sample templates with placeholders and conditionals
	htmlTemplate := `<div style="font-family:Segoe UI,Arial,sans-serif">
<strong>{{name}}</strong><br>
{{title}}{{#if department}}, {{department}}{{/if}}<br>
{{#if logo}}<img src="{{logo}}" alt="Company Logo">{{/if}}
</div>`
	
	textTemplate := `{{name}}
{{title}}{{#if department}}, {{department}}{{/if}}`
	
	// Sample user data
	data := map[string]string{
		"name":       "Jane Doe",
		"title":      "Senior Engineer",
		"department": "Engineering",
		"logo":       "https://via.placeholder.com/120x40",
	}
	
	// Render templates with safe HTML escaping
	htmlRendered := renderer.RenderSafe(htmlTemplate, data)
	textRendered := renderer.Render(textTemplate, data)
	
	type Resp struct {
		HTML string `json:"html"`
		Text string `json:"text"`
	}
	resp := Resp{
		HTML: htmlRendered,
		Text: textRendered,
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
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	templateHandler := handlers.NewTemplateHandler(templateStore, auditStore)
	auditHandler := handlers.NewAuditHandler(auditStore)
	agentHandler := handlers.NewAgentHandler(templateStore)
	scheduleStore := store.NewScheduleStore()
	scheduleHandler := handlers.NewScheduleHandler(scheduleStore)
	ruleStore := store.NewRuleStore(scheduleStore)
	ruleHandler := handlers.NewRuleHandler(ruleStore)

	// Initialize auth service
	authConfig := auth.NewConfigFromEnv()
	tokenStore := auth.NewInMemoryTokenStore()
	authService, err := auth.NewService(authConfig, tokenStore)
	if err != nil {
		log.Printf("Warning: Auth service initialization failed: %v", err)
		log.Printf("Authentication endpoints will not be available")
		log.Printf("Set AZURE_CLIENT_ID and optionally AZURE_CLIENT_SECRET to enable auth")
	}
	var authHandler *handlers.AuthHandler
	if authService != nil {
		authHandler = handlers.NewAuthHandler(authService)
	}

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/v1/preview", preview)
	
	// Agent endpoints
	mux.HandleFunc("/v1/agent/templates", agentHandler.GetTemplates)

	// Auth endpoints (if auth service is available)
	if authHandler != nil {
		mux.HandleFunc("/auth/login", authHandler.Login)
		mux.HandleFunc("/auth/callback", authHandler.Callback)
		mux.HandleFunc("/auth/status", authHandler.Status)
		mux.HandleFunc("/auth/logout", authHandler.Logout)
		log.Printf("Microsoft Graph auth endpoints enabled")
	}

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

	// Rule endpoints - require Signature Admin role
	mux.Handle("/v1/rules/evaluate", middleware.MockAuthMiddleware(nil)(
		http.HandlerFunc(ruleHandler.EvaluateRules),
	))

	mux.Handle("/v1/rules", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleSignatureAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					ruleHandler.CreateRule(w, r)
				} else if r.Method == http.MethodGet {
					ruleHandler.ListRules(w, r)
				} else {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	mux.Handle("/v1/rules/", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleSignatureAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/rules/evaluate" {
					http.NotFound(w, r)
					return
				}
				if !strings.HasPrefix(r.URL.Path, "/v1/rules/") {
					http.NotFound(w, r)
					return
				}
				switch r.Method {
				case http.MethodGet:
					ruleHandler.GetRule(w, r)
				case http.MethodPut:
					ruleHandler.UpdateRule(w, r)
				case http.MethodDelete:
					ruleHandler.DeleteRule(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	// Schedule endpoints - require Signature Admin role
	mux.Handle("/v1/schedules", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleSignatureAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					scheduleHandler.CreateSchedule(w, r)
				} else if r.Method == http.MethodGet {
					scheduleHandler.ListSchedules(w, r)
				} else {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	mux.Handle("/v1/schedules/", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleSignatureAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasPrefix(r.URL.Path, "/v1/schedules/") {
					http.NotFound(w, r)
					return
				}
				switch r.Method {
				case http.MethodGet:
					scheduleHandler.GetSchedule(w, r)
				case http.MethodPut:
					scheduleHandler.UpdateSchedule(w, r)
				case http.MethodDelete:
					scheduleHandler.DeleteSchedule(w, r)
				default:
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	// Template CRUD endpoints - require Signature Admin role for CRUD, Approver for approve/reject
	mux.Handle("/v1/templates", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleSignatureAdmin)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					templateHandler.CreateTemplate(w, r)
				} else if r.Method == http.MethodGet {
					templateHandler.ListTemplates(w, r)
				} else {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	mux.Handle("/v1/templates/", middleware.MockAuthMiddleware(nil)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/v1/templates/") {
				http.NotFound(w, r)
				return
			}
			
			user := r.Context().Value(middleware.UserContextKey).(*models.User)
			
			// Approval actions require Approver role
			if strings.HasSuffix(r.URL.Path, "/approve") && r.Method == http.MethodPost {
				if user.Role != models.RoleApprover && user.Role != models.RoleOrgAdmin {
					http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
					return
				}
				templateHandler.ApproveTemplate(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/reject") && r.Method == http.MethodPost {
				if user.Role != models.RoleApprover && user.Role != models.RoleOrgAdmin {
					http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
					return
				}
				templateHandler.RejectTemplate(w, r)
				return
			}
			
			// All other template actions require Signature Admin role
			if user.Role != models.RoleSignatureAdmin && user.Role != models.RoleOrgAdmin {
				http.Error(w, "forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			
			// Handle workflow actions
			if strings.HasSuffix(r.URL.Path, "/submit") && r.Method == http.MethodPost {
				templateHandler.SubmitForReview(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/publish") && r.Method == http.MethodPost {
				templateHandler.PublishTemplate(w, r)
				return
			}
			if strings.HasSuffix(r.URL.Path, "/unpublish") && r.Method == http.MethodPost {
				templateHandler.UnpublishTemplate(w, r)
				return
			}
			
			switch r.Method {
			case http.MethodGet:
				templateHandler.GetTemplate(w, r)
			case http.MethodPut:
				templateHandler.UpdateTemplate(w, r)
			case http.MethodDelete:
				templateHandler.DeleteTemplate(w, r)
			default:
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			}
		}),
	))

	// Audit log endpoints - require Auditor role (read-only)
	mux.Handle("/v1/audit", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleAuditor)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					auditHandler.ListAuditEntries(w, r)
				} else {
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				}
			}),
		),
	))

	mux.Handle("/v1/audit/stats", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleAuditor)(
			http.HandlerFunc(auditHandler.GetAuditStats),
		),
	))

	mux.Handle("/v1/audit/resource/", middleware.MockAuthMiddleware(nil)(
		middleware.RequireRole(models.RoleAuditor)(
			http.HandlerFunc(auditHandler.GetAuditEntriesForResource),
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
	log.Printf("Endpoints: /v1/rules, /v1/schedules, /v1/templates (with approval workflow)")
	log.Printf("Audit endpoints: /v1/audit, /v1/audit/stats, /v1/audit/resource/{type}/{id}")
	log.Fatal(http.ListenAndServe(addr, mux))
}
