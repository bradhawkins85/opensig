package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/renderer"
)

// TemplateStore defines the interface for template operations
type TemplateStore interface {
	GetTemplatesByTenantID(tenantID string) ([]*models.Template, error)
}

// AgentHandler handles agent-specific endpoints
type AgentHandler struct {
	templateStore TemplateStore
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(templateStore TemplateStore) *AgentHandler {
	return &AgentHandler{
		templateStore: templateStore,
	}
}

// GetTemplates returns templates for the authenticated user
// GET /v1/agent/templates
func (h *AgentHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For now, we'll use a mock user. In production, this would come from authentication
	userEmail := r.Header.Get("X-User-Email")
	userID := r.Header.Get("X-User-ID")
	tenantID := r.Header.Get("X-Tenant-ID")
	
	if userEmail == "" {
		userEmail = "user@example.com"
	}
	if userID == "" {
		userID = "user-123"
	}
	if tenantID == "" {
		tenantID = "default"
	}
	
	// Get templates for the tenant
	templates, err := h.templateStore.GetTemplatesByTenantID(tenantID)
	if err != nil {
		log.Printf("Error fetching templates: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch templates",
		})
		return
	}

	// Mock user data for rendering (in production, this would come from Microsoft Graph)
	userData := map[string]string{
		"name":       "John Doe",
		"title":      "Software Engineer",
		"department": "Engineering",
		"phone":      "+1 (555) 123-4567",
		"email":      userEmail,
		"company":    "Example Corp",
	}

	// Render templates with user data
	var renderedTemplates []models.RenderedTemplate
	for _, template := range templates {
		renderedTemplates = append(renderedTemplates, models.RenderedTemplate{
			ID:          template.ID,
			Name:        template.Name,
			HTMLContent: renderer.Render(template.HTMLContent, userData),
			RTFContent:  renderer.Render(template.RTFContent, userData),
			TextContent: renderer.Render(template.TextContent, userData),
			Assets:      template.Assets,
		})
	}

	// Check if the feature flag to set default signatures is enabled
	setDefaultSignatures := false
	if envValue := os.Getenv("OPENSIG_SET_DEFAULT_SIGNATURES"); envValue != "" {
		setDefaultSignatures = strings.ToLower(envValue) == "true" || envValue == "1"
	}

	response := models.AgentTemplateResponse{
		Templates:            renderedTemplates,
		UserEmail:            userEmail,
		UserID:               userID,
		SetDefaultSignatures: setDefaultSignatures,
	}

	log.Printf("Serving %d templates to agent for user %s (set_default_signatures=%v)", len(renderedTemplates), userEmail, setDefaultSignatures)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
