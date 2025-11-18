package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/opensig/server/internal/models"
)

// TemplateStore manages template data
type TemplateStore struct {
	mu        sync.RWMutex
	templates map[string]*models.Template
}

// NewTemplateStore creates a new template store
func NewTemplateStore() *TemplateStore {
	store := &TemplateStore{
		templates: make(map[string]*models.Template),
	}
	
	// Add a default template for testing
	defaultTemplate := &models.Template{
		ID:       uuid.New().String(),
		TenantID: "default",
		Name:     "OpenSig Default",
		HTMLContent: `<div style="font-family:Segoe UI,Arial,sans-serif">
<strong>{{name}}</strong><br>
{{title}}{{#if department}}, {{department}}{{/if}}<br>
{{#if phone}}Phone: {{phone}}<br>{{/if}}
{{#if email}}Email: {{email}}<br>{{/if}}
{{#if company}}{{company}}{{/if}}
</div>`,
		RTFContent: `{\rtf1\ansi {{name}}\line {{title}}{{#if department}}, {{department}}{{/if}}\line {{#if phone}}Phone: {{phone}}\line{{/if}}{{#if email}}Email: {{email}}\line{{/if}}{{#if company}}{{company}}{{/if}}}`,
		TextContent: `{{name}}
{{title}}{{#if department}}, {{department}}{{/if}}
{{#if phone}}Phone: {{phone}}{{/if}}
{{#if email}}Email: {{email}}{{/if}}
{{#if company}}{{company}}{{/if}}`,
		Active:    true,
		Status:    models.ApprovalStatusApproved, // Default template is pre-approved
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	store.templates[defaultTemplate.ID] = defaultTemplate
	
	return store
}

// GetTemplatesByTenantID returns all active templates for a tenant
func (s *TemplateStore) GetTemplatesByTenantID(tenantID string) ([]*models.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var templates []*models.Template
	for _, template := range s.templates {
		if (template.TenantID == tenantID || template.TenantID == "default") && template.Active {
			templates = append(templates, template)
		}
	}
	
	if len(templates) == 0 {
		return nil, fmt.Errorf("no templates found for tenant %s", tenantID)
	}
	
	return templates, nil
}

// GetTemplate retrieves a template by ID
func (s *TemplateStore) GetTemplate(id string) (*models.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	template, exists := s.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}
	
	return template, nil
}

// CreateTemplate creates a new template
func (s *TemplateStore) CreateTemplate(template *models.Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if template.ID == "" {
		template.ID = uuid.New().String()
	}
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	
	// Set default status to draft if not specified
	if template.Status == "" {
		template.Status = models.ApprovalStatusDraft
	}
	
	// Initialize version to 1 if not set
	if template.Version == 0 {
		template.Version = 1
	}
	
	s.templates[template.ID] = template
	return nil
}

// UpdateTemplate updates an existing template
func (s *TemplateStore) UpdateTemplate(template *models.Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.templates[template.ID]; !exists {
		return fmt.Errorf("template not found: %s", template.ID)
	}
	
	template.UpdatedAt = time.Now()
	s.templates[template.ID] = template
	return nil
}

// DeleteTemplate deletes a template by ID
func (s *TemplateStore) DeleteTemplate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.templates[id]; !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	
	delete(s.templates, id)
	return nil
}

// SubmitForReview submits a template for review
func (s *TemplateStore) SubmitForReview(id string, submittedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	template, exists := s.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	
	if template.Status != models.ApprovalStatusDraft {
		return fmt.Errorf("only draft templates can be submitted for review")
	}
	
	now := time.Now()
	template.Status = models.ApprovalStatusPendingReview
	template.SubmittedBy = submittedBy
	template.SubmittedAt = &now
	template.UpdatedAt = now
	
	return nil
}

// ApproveTemplate approves a template
func (s *TemplateStore) ApproveTemplate(id string, reviewedBy string, comments string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	template, exists := s.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	
	if template.Status != models.ApprovalStatusPendingReview {
		return fmt.Errorf("only templates pending review can be approved")
	}
	
	now := time.Now()
	template.Status = models.ApprovalStatusApproved
	template.ReviewedBy = reviewedBy
	template.ReviewedAt = &now
	template.ReviewComments = comments
	template.UpdatedAt = now
	
	return nil
}

// RejectTemplate rejects a template
func (s *TemplateStore) RejectTemplate(id string, reviewedBy string, comments string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	template, exists := s.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	
	if template.Status != models.ApprovalStatusPendingReview {
		return fmt.Errorf("only templates pending review can be rejected")
	}
	
	now := time.Now()
	template.Status = models.ApprovalStatusRejected
	template.ReviewedBy = reviewedBy
	template.ReviewedAt = &now
	template.ReviewComments = comments
	template.UpdatedAt = now
	
	return nil
}

// PublishTemplate publishes an approved template (sets active=true)
func (s *TemplateStore) PublishTemplate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	template, exists := s.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	
	if template.Status != models.ApprovalStatusApproved {
		return fmt.Errorf("only approved templates can be published")
	}
	
	template.Active = true
	template.UpdatedAt = time.Now()
	
	return nil
}

// UnpublishTemplate unpublishes a template (sets active=false)
func (s *TemplateStore) UnpublishTemplate(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	template, exists := s.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}
	
	template.Active = false
	template.UpdatedAt = time.Now()
	
	return nil
}

// ListTemplates returns all templates with optional status filter
func (s *TemplateStore) ListTemplates(tenantID string, status models.ApprovalStatus) ([]*models.Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var templates []*models.Template
	for _, template := range s.templates {
		if template.TenantID == tenantID || template.TenantID == "default" {
			if status == "" || template.Status == status {
				templates = append(templates, template)
			}
		}
	}
	
	return templates, nil
}
