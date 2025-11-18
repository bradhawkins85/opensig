package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/your-org/opensig/server/internal/middleware"
	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

// TemplateHandler handles template-related HTTP requests
type TemplateHandler struct {
	templateStore *store.TemplateStore
	auditStore    *store.AuditStore
}

// NewTemplateHandler creates a new template handler
func NewTemplateHandler(templateStore *store.TemplateStore, auditStore *store.AuditStore) *TemplateHandler {
	return &TemplateHandler{
		templateStore: templateStore,
		auditStore:    auditStore,
	}
}

// CreateTemplate creates a new template (draft status by default)
func (h *TemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var template models.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	template.TenantID = user.TenantID
	template.SubmittedBy = user.ID

	if err := h.templateStore.CreateTemplate(&template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionCreate,
		user,
		nil,
		&template,
	)
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(template)
}

// GetTemplate retrieves a template by ID
func (h *TemplateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	
	template, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}

// ListTemplates lists all templates with optional status filter
func (h *TemplateHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)
	
	statusParam := r.URL.Query().Get("status")
	var status models.ApprovalStatus
	if statusParam != "" {
		status = models.ApprovalStatus(statusParam)
	}
	
	templates, err := h.templateStore.ListTemplates(user.TenantID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// UpdateTemplate updates an existing template
func (h *TemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	
	// Get the existing template for audit log
	oldTemplate, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var template models.Template
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	template.ID = id
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	if err := h.templateStore.UpdateTemplate(&template); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log audit entry with diff
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionUpdate,
		user,
		oldTemplate,
		&template,
	)
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}

// DeleteTemplate deletes a template
func (h *TemplateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	
	// Get the template for audit log before deletion
	template, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	if err := h.templateStore.DeleteTemplate(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionDelete,
		user,
		template,
		nil,
	)
	_ = h.auditStore.LogEntry(auditEntry)

	w.WriteHeader(http.StatusNoContent)
}

// SubmitForReview submits a template for review
func (h *TemplateHandler) SubmitForReview(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	id = strings.TrimSuffix(id, "/submit")
	
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Get the template before submission for audit log
	oldTemplate, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.templateStore.SubmitForReview(id, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the updated template
	template, _ := h.templateStore.GetTemplate(id)

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionSubmitReview,
		user,
		oldTemplate,
		template,
	)
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}

// ApproveTemplate approves a template
func (h *TemplateHandler) ApproveTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	id = strings.TrimSuffix(id, "/approve")
	
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Get the template before approval for audit log
	oldTemplate, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Parse optional review comments
	var req struct {
		Comments string `json:"comments"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.templateStore.ApproveTemplate(id, user.ID, req.Comments); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the updated template
	template, _ := h.templateStore.GetTemplate(id)

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionApprove,
		user,
		oldTemplate,
		template,
	)
	auditEntry.Metadata = map[string]string{"comments": req.Comments}
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}

// RejectTemplate rejects a template
func (h *TemplateHandler) RejectTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	id = strings.TrimSuffix(id, "/reject")
	
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Get the template before rejection for audit log
	oldTemplate, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Parse optional review comments
	var req struct {
		Comments string `json:"comments"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	if err := h.templateStore.RejectTemplate(id, user.ID, req.Comments); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the updated template
	template, _ := h.templateStore.GetTemplate(id)

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionReject,
		user,
		oldTemplate,
		template,
	)
	auditEntry.Metadata = map[string]string{"comments": req.Comments}
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}

// PublishTemplate publishes an approved template
func (h *TemplateHandler) PublishTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	id = strings.TrimSuffix(id, "/publish")
	
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Get the template before publishing for audit log
	oldTemplate, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.templateStore.PublishTemplate(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the updated template
	template, _ := h.templateStore.GetTemplate(id)

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionPublish,
		user,
		oldTemplate,
		template,
	)
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}

// UnpublishTemplate unpublishes a template
func (h *TemplateHandler) UnpublishTemplate(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/templates/")
	id = strings.TrimSuffix(id, "/unpublish")
	
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Get the template before unpublishing for audit log
	oldTemplate, err := h.templateStore.GetTemplate(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := h.templateStore.UnpublishTemplate(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the updated template
	template, _ := h.templateStore.GetTemplate(id)

	// Log audit entry
	auditEntry := store.CreateAuditEntry(
		template.TenantID,
		models.AuditResourceTypeTemplate,
		template.ID,
		models.AuditActionUnpublish,
		user,
		oldTemplate,
		template,
	)
	_ = h.auditStore.LogEntry(auditEntry)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(template)
}
