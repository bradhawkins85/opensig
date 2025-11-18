package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

func TestRuleHandler_CreateRule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	ruleStore := store.NewRuleStore(scheduleStore)
	handler := NewRuleHandler(ruleStore)

	rule := models.Rule{
		TenantID:    "tenant1",
		Name:        "Test Rule",
		Description: "Test description",
		TemplateID:  "template1",
		Priority:    10,
		Active:      true,
		Conditions: models.Conditions{
			SenderDomains: []string{"example.com"},
		},
	}

	body, _ := json.Marshal(rule)
	req := httptest.NewRequest(http.MethodPost, "/v1/rules", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateRule(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var created models.Rule
	json.NewDecoder(w.Body).Decode(&created)

	if created.ID == "" {
		t.Error("Expected rule ID to be generated")
	}

	if created.Name != rule.Name {
		t.Errorf("Expected name %s, got %s", rule.Name, created.Name)
	}
}

func TestRuleHandler_EvaluateRules(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	ruleStore := store.NewRuleStore(scheduleStore)
	handler := NewRuleHandler(ruleStore)

	// Create a rule
	rule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Test Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
		Conditions: models.Conditions{
			SenderDomains: []string{"example.com"},
		},
	}
	ruleStore.CreateRule(rule)

	// Evaluate with a matching message
	evalReq := struct {
		TenantID string             `json:"tenant_id"`
		Message  models.TestMessage `json:"message"`
	}{
		TenantID: "tenant1",
		Message: models.TestMessage{
			SenderEmail:     "user@example.com",
			RecipientEmails: []string{"recipient@test.com"},
			MessageType:     models.MessageTypeNew,
			Timestamp:       time.Now(),
		},
	}

	body, _ := json.Marshal(evalReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/rules/evaluate", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.EvaluateRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result models.RuleEvaluationResult
	json.NewDecoder(w.Body).Decode(&result)

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}

	if result.SelectedRule == nil {
		t.Error("Expected a selected rule")
	}
}

func TestRuleHandler_GetRule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	ruleStore := store.NewRuleStore(scheduleStore)
	handler := NewRuleHandler(ruleStore)

	rule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Test Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
	}
	ruleStore.CreateRule(rule)

	req := httptest.NewRequest(http.MethodGet, "/v1/rules/"+rule.ID, nil)
	w := httptest.NewRecorder()

	handler.GetRule(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var retrieved models.Rule
	json.NewDecoder(w.Body).Decode(&retrieved)

	if retrieved.ID != rule.ID {
		t.Errorf("Expected ID %s, got %s", rule.ID, retrieved.ID)
	}
}

func TestRuleHandler_ListRules(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	ruleStore := store.NewRuleStore(scheduleStore)
	handler := NewRuleHandler(ruleStore)

	// Create two rules for the same tenant
	rule1 := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Rule 1",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
	}
	rule2 := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Rule 2",
		TemplateID: "template2",
		Priority:   5,
		Active:     true,
	}
	ruleStore.CreateRule(rule1)
	ruleStore.CreateRule(rule2)

	req := httptest.NewRequest(http.MethodGet, "/v1/rules?tenant_id=tenant1", nil)
	w := httptest.NewRecorder()

	handler.ListRules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var rules []*models.Rule
	json.NewDecoder(w.Body).Decode(&rules)

	if len(rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(rules))
	}
}

func TestRuleHandler_UpdateRule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	ruleStore := store.NewRuleStore(scheduleStore)
	handler := NewRuleHandler(ruleStore)

	rule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Original Name",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
	}
	ruleStore.CreateRule(rule)

	// Update the rule
	rule.Name = "Updated Name"
	body, _ := json.Marshal(rule)
	req := httptest.NewRequest(http.MethodPut, "/v1/rules/"+rule.ID, bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateRule(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var updated models.Rule
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestRuleHandler_DeleteRule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	ruleStore := store.NewRuleStore(scheduleStore)
	handler := NewRuleHandler(ruleStore)

	rule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "To Delete",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
	}
	ruleStore.CreateRule(rule)

	req := httptest.NewRequest(http.MethodDelete, "/v1/rules/"+rule.ID, nil)
	w := httptest.NewRecorder()

	handler.DeleteRule(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify deletion
	_, err := ruleStore.GetRule(rule.ID)
	if err == nil {
		t.Error("Expected error when getting deleted rule")
	}
}
