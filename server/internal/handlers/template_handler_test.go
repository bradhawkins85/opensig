package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/opensig/server/internal/middleware"
	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

func TestTemplateHandler_CreateTemplate(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	template := models.Template{
		Name:        "Test Template",
		HTMLContent: "<p>Test</p>",
		RTFContent:  "{\\rtf Test}",
		TextContent: "Test",
	}

	body, _ := json.Marshal(template)
	req := httptest.NewRequest(http.MethodPost, "/v1/templates", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "user1",
		Email:    "test@example.com",
		TenantID: "tenant1",
		Role:     models.RoleSignatureAdmin,
	}))

	w := httptest.NewRecorder()
	handler.CreateTemplate(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result models.Template
	_ = json.NewDecoder(w.Body).Decode(&result)

	if result.Status != models.ApprovalStatusDraft {
		t.Errorf("Expected status draft, got %s", result.Status)
	}

	if result.Version != 1 {
		t.Errorf("Expected version 1, got %d", result.Version)
	}

	// Verify audit log
	if auditStore.Count() != 1 {
		t.Errorf("Expected 1 audit entry, got %d", auditStore.Count())
	}
}

func TestTemplateHandler_SubmitForReview(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	// Create a draft template
	template := &models.Template{
		ID:          "template1",
		TenantID:    "tenant1",
		Name:        "Test Template",
		HTMLContent: "<p>Test</p>",
		Status:      models.ApprovalStatusDraft,
		Version:     1,
	}
	_ = templateStore.CreateTemplate(template)

	req := httptest.NewRequest(http.MethodPost, "/v1/templates/template1/submit", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "user1",
		Email:    "test@example.com",
		TenantID: "tenant1",
		Role:     models.RoleSignatureAdmin,
	}))

	w := httptest.NewRecorder()
	handler.SubmitForReview(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.Template
	_ = json.NewDecoder(w.Body).Decode(&result)

	if result.Status != models.ApprovalStatusPendingReview {
		t.Errorf("Expected status pending_review, got %s", result.Status)
	}

	if result.SubmittedBy != "user1" {
		t.Errorf("Expected submitted_by user1, got %s", result.SubmittedBy)
	}

	// Verify audit log
	entries, _ := auditStore.GetEntriesForResource(models.AuditResourceTypeTemplate, "template1")
	if len(entries) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Action != models.AuditActionSubmitReview {
		t.Errorf("Expected action submit_review, got %s", entries[0].Action)
	}
}

func TestTemplateHandler_ApproveTemplate(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	// Create a template pending review
	template := &models.Template{
		ID:          "template1",
		TenantID:    "tenant1",
		Name:        "Test Template",
		HTMLContent: "<p>Test</p>",
		Status:      models.ApprovalStatusPendingReview,
		Version:     1,
	}
	_ = templateStore.CreateTemplate(template)

	reqBody := map[string]string{"comments": "Looks good!"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/templates/template1/approve", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "approver1",
		Email:    "approver@example.com",
		TenantID: "tenant1",
		Role:     models.RoleApprover,
	}))

	w := httptest.NewRecorder()
	handler.ApproveTemplate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.Template
	_ = json.NewDecoder(w.Body).Decode(&result)

	if result.Status != models.ApprovalStatusApproved {
		t.Errorf("Expected status approved, got %s", result.Status)
	}

	if result.ReviewedBy != "approver1" {
		t.Errorf("Expected reviewed_by approver1, got %s", result.ReviewedBy)
	}

	if result.ReviewComments != "Looks good!" {
		t.Errorf("Expected review_comments 'Looks good!', got %s", result.ReviewComments)
	}

	// Verify audit log
	entries, _ := auditStore.GetEntriesForResource(models.AuditResourceTypeTemplate, "template1")
	if len(entries) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Action != models.AuditActionApprove {
		t.Errorf("Expected action approve, got %s", entries[0].Action)
	}
}

func TestTemplateHandler_RejectTemplate(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	// Create a template pending review
	template := &models.Template{
		ID:          "template1",
		TenantID:    "tenant1",
		Name:        "Test Template",
		HTMLContent: "<p>Test</p>",
		Status:      models.ApprovalStatusPendingReview,
		Version:     1,
	}
	_ = templateStore.CreateTemplate(template)

	reqBody := map[string]string{"comments": "Needs improvement"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/templates/template1/reject", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "approver1",
		Email:    "approver@example.com",
		TenantID: "tenant1",
		Role:     models.RoleApprover,
	}))

	w := httptest.NewRecorder()
	handler.RejectTemplate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.Template
	_ = json.NewDecoder(w.Body).Decode(&result)

	if result.Status != models.ApprovalStatusRejected {
		t.Errorf("Expected status rejected, got %s", result.Status)
	}

	if result.ReviewComments != "Needs improvement" {
		t.Errorf("Expected review_comments 'Needs improvement', got %s", result.ReviewComments)
	}

	// Verify audit log
	entries, _ := auditStore.GetEntriesForResource(models.AuditResourceTypeTemplate, "template1")
	if len(entries) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Action != models.AuditActionReject {
		t.Errorf("Expected action reject, got %s", entries[0].Action)
	}
}

func TestTemplateHandler_PublishTemplate(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	// Create an approved template
	template := &models.Template{
		ID:          "template1",
		TenantID:    "tenant1",
		Name:        "Test Template",
		HTMLContent: "<p>Test</p>",
		Status:      models.ApprovalStatusApproved,
		Active:      false,
		Version:     1,
	}
	_ = templateStore.CreateTemplate(template)

	req := httptest.NewRequest(http.MethodPost, "/v1/templates/template1/publish", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "admin1",
		Email:    "admin@example.com",
		TenantID: "tenant1",
		Role:     models.RoleOrgAdmin,
	}))

	w := httptest.NewRecorder()
	handler.PublishTemplate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.Template
	_ = json.NewDecoder(w.Body).Decode(&result)

	if !result.Active {
		t.Error("Expected template to be active after publishing")
	}

	// Verify audit log
	entries, _ := auditStore.GetEntriesForResource(models.AuditResourceTypeTemplate, "template1")
	if len(entries) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Action != models.AuditActionPublish {
		t.Errorf("Expected action publish, got %s", entries[0].Action)
	}
}

func TestTemplateHandler_WorkflowValidation(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	// Create a draft template
	template := &models.Template{
		ID:          "template1",
		TenantID:    "tenant1",
		Name:        "Test Template",
		HTMLContent: "<p>Test</p>",
		Status:      models.ApprovalStatusDraft,
		Version:     1,
	}
	_ = templateStore.CreateTemplate(template)

	user := &models.User{
		ID:       "approver1",
		Email:    "approver@example.com",
		TenantID: "tenant1",
		Role:     models.RoleApprover,
	}

	// Try to approve a draft template (should fail)
	req := httptest.NewRequest(http.MethodPost, "/v1/templates/template1/approve", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, user))

	w := httptest.NewRecorder()
	handler.ApproveTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 when approving draft template, got %d", w.Code)
	}

	// Try to publish a draft template (should fail)
	req = httptest.NewRequest(http.MethodPost, "/v1/templates/template1/publish", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, user))

	w = httptest.NewRecorder()
	handler.PublishTemplate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 when publishing draft template, got %d", w.Code)
	}
}

func TestTemplateHandler_ListTemplates(t *testing.T) {
	templateStore := store.NewTemplateStore()
	auditStore := store.NewAuditStore()
	handler := NewTemplateHandler(templateStore, auditStore)

	// Create templates with different statuses
	templates := []*models.Template{
		{ID: "t1", TenantID: "tenant1", Name: "T1", Status: models.ApprovalStatusDraft, Version: 1},
		{ID: "t2", TenantID: "tenant1", Name: "T2", Status: models.ApprovalStatusPendingReview, Version: 1},
		{ID: "t3", TenantID: "tenant1", Name: "T3", Status: models.ApprovalStatusApproved, Version: 1},
	}
	for _, tmpl := range templates {
		_ = templateStore.CreateTemplate(tmpl)
	}

	// List all templates
	req := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "user1",
		Email:    "test@example.com",
		TenantID: "tenant1",
		Role:     models.RoleSignatureAdmin,
	}))

	w := httptest.NewRecorder()
	handler.ListTemplates(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)

	count := int(result["count"].(float64))
	if count < 3 { // At least 3 (may include default template)
		t.Errorf("Expected at least 3 templates, got %d", count)
	}

	// List only draft templates
	req = httptest.NewRequest(http.MethodGet, "/v1/templates?status=draft", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, &models.User{
		ID:       "user1",
		Email:    "test@example.com",
		TenantID: "tenant1",
		Role:     models.RoleSignatureAdmin,
	}))

	w = httptest.NewRecorder()
	handler.ListTemplates(w, req)

	_ = json.NewDecoder(w.Body).Decode(&result)
	count = int(result["count"].(float64))
	if count != 1 {
		t.Errorf("Expected 1 draft template, got %d", count)
	}
}
