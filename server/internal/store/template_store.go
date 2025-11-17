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
