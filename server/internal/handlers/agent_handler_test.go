package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/your-org/opensig/server/internal/models"
)

type mockTemplateStore struct {
	templates []*models.Template
	err       error
}

func (m *mockTemplateStore) GetTemplatesByTenantID(tenantID string) ([]*models.Template, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.templates, nil
}

func TestAgentHandler_GetTemplates(t *testing.T) {
	// Create mock template
	mockTemplate := &models.Template{
		ID:          "test-123",
		TenantID:    "default",
		Name:        "Test Template",
		HTMLContent: "<div>{{name}}</div>",
		RTFContent:  "{\\rtf1\\ansi {{name}}}",
		TextContent: "{{name}}",
		Active:      true,
	}

	tests := []struct {
		name           string
		store          *mockTemplateStore
		headers        map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *models.AgentTemplateResponse)
	}{
		{
			name: "successful template fetch",
			store: &mockTemplateStore{
				templates: []*models.Template{mockTemplate},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *models.AgentTemplateResponse) {
				if len(resp.Templates) != 1 {
					t.Errorf("Expected 1 template, got %d", len(resp.Templates))
				}
				if resp.Templates[0].Name != "Test Template" {
					t.Errorf("Expected template name 'Test Template', got %s", resp.Templates[0].Name)
				}
				if resp.Templates[0].HTMLContent == "" {
					t.Error("Expected HTML content to be rendered")
				}
			},
		},
		{
			name: "with custom headers",
			store: &mockTemplateStore{
				templates: []*models.Template{mockTemplate},
			},
			headers: map[string]string{
				"X-User-Email": "custom@example.com",
				"X-User-ID":    "custom-123",
				"X-Tenant-ID":  "custom-tenant",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, resp *models.AgentTemplateResponse) {
				if resp.UserEmail != "custom@example.com" {
					t.Errorf("Expected user email 'custom@example.com', got %s", resp.UserEmail)
				}
				if resp.UserID != "custom-123" {
					t.Errorf("Expected user ID 'custom-123', got %s", resp.UserID)
				}
				// Verify SetDefaultSignatures field is present (default false when env var not set)
				if resp.SetDefaultSignatures != false {
					t.Errorf("Expected SetDefaultSignatures to be false (default), got %v", resp.SetDefaultSignatures)
				}
			},
		},
		{
			name: "template store error",
			store: &mockTemplateStore{
				err: &mockError{msg: "database error"},
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAgentHandler(tt.store)

			req := httptest.NewRequest(http.MethodGet, "/v1/agent/templates", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			rr := httptest.NewRecorder()
			handler.GetTemplates(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if tt.checkResponse != nil && rr.Code == http.StatusOK {
				var response models.AgentTemplateResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				tt.checkResponse(t, &response)
			}
		})
	}
}

func TestAgentHandler_GetTemplates_MethodNotAllowed(t *testing.T) {
	handler := NewAgentHandler(&mockTemplateStore{})

	req := httptest.NewRequest(http.MethodPost, "/v1/agent/templates", nil)
	rr := httptest.NewRecorder()

	handler.GetTemplates(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestAgentHandler_GetTemplates_SetDefaultSignaturesEnabled(t *testing.T) {
	// Set the environment variable
	os.Setenv("OPENSIG_SET_DEFAULT_SIGNATURES", "true")
	defer os.Unsetenv("OPENSIG_SET_DEFAULT_SIGNATURES")

	mockTemplate := &models.Template{
		ID:          "test-123",
		TenantID:    "default",
		Name:        "Test Template",
		HTMLContent: "<div>{{name}}</div>",
		RTFContent:  "{\\rtf1\\ansi {{name}}}",
		TextContent: "{{name}}",
		Active:      true,
	}

	store := &mockTemplateStore{
		templates: []*models.Template{mockTemplate},
	}
	handler := NewAgentHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/v1/agent/templates", nil)
	rr := httptest.NewRecorder()

	handler.GetTemplates(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response models.AgentTemplateResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.SetDefaultSignatures {
		t.Errorf("Expected SetDefaultSignatures to be true when env var is set, got false")
	}
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}
