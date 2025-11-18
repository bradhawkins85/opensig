package store

import (
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/models"
)

func TestRuleStore_CreateAndGet(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	rule := &models.Rule{
		TenantID:    "tenant1",
		Name:        "Test Rule",
		Description: "A test rule",
		TemplateID:  "template1",
		Priority:    10,
		Active:      true,
		Conditions: models.Conditions{
			SenderDomains: []string{"example.com"},
		},
	}

	err := store.CreateRule(rule)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	if rule.ID == "" {
		t.Fatal("Rule ID should be generated")
	}

	retrieved, err := store.GetRule(rule.ID)
	if err != nil {
		t.Fatalf("Failed to get rule: %v", err)
	}

	if retrieved.Name != rule.Name {
		t.Errorf("Expected name %s, got %s", rule.Name, retrieved.Name)
	}
}

func TestRuleStore_EvaluateConditions_SenderEmail(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	rule := &models.Rule{
		ID:         "rule1",
		TenantID:   "tenant1",
		Name:       "Sender Email Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
		Conditions: models.Conditions{
			SenderEmails: []string{"john@example.com", "jane@example.com"},
		},
	}

	store.CreateRule(rule)

	// Test matching sender
	message := &models.TestMessage{
		SenderEmail:     "john@example.com",
		RecipientEmails: []string{"recipient@test.com"},
		MessageType:     models.MessageTypeNew,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}

	// Test non-matching sender
	message.SenderEmail = "other@example.com"
	result, err = store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 0 {
		t.Errorf("Expected 0 matched rules, got %d", len(result.MatchedRules))
	}
}

func TestRuleStore_EvaluateConditions_SenderDomain(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	rule := &models.Rule{
		ID:         "rule1",
		TenantID:   "tenant1",
		Name:       "Sender Domain Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
		Conditions: models.Conditions{
			SenderDomains: []string{"example.com"},
		},
	}

	store.CreateRule(rule)

	message := &models.TestMessage{
		SenderEmail:     "anyone@example.com",
		RecipientEmails: []string{"recipient@test.com"},
		MessageType:     models.MessageTypeNew,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}

	// Test non-matching domain
	message.SenderEmail = "user@other.com"
	result, err = store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 0 {
		t.Errorf("Expected 0 matched rules, got %d", len(result.MatchedRules))
	}
}

func TestRuleStore_EvaluateConditions_RecipientEmail(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	rule := &models.Rule{
		ID:         "rule1",
		TenantID:   "tenant1",
		Name:       "Recipient Email Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
		Conditions: models.Conditions{
			RecipientEmails: []string{"client@example.com"},
		},
	}

	store.CreateRule(rule)

	message := &models.TestMessage{
		SenderEmail:     "sender@test.com",
		RecipientEmails: []string{"client@example.com", "other@test.com"},
		MessageType:     models.MessageTypeNew,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}
}

func TestRuleStore_EvaluateConditions_MessageType(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	rule := &models.Rule{
		ID:         "rule1",
		TenantID:   "tenant1",
		Name:       "Reply Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     true,
		Conditions: models.Conditions{
			MessageTypes: []models.MessageType{models.MessageTypeReply},
		},
	}

	store.CreateRule(rule)

	// Test matching message type
	message := &models.TestMessage{
		SenderEmail:     "sender@test.com",
		RecipientEmails: []string{"recipient@test.com"},
		MessageType:     models.MessageTypeReply,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule, got %d", len(result.MatchedRules))
	}

	// Test non-matching message type
	message.MessageType = models.MessageTypeNew
	result, err = store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 0 {
		t.Errorf("Expected 0 matched rules, got %d", len(result.MatchedRules))
	}
}

func TestRuleStore_Priority(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	// Create two rules with different priorities
	lowPriorityRule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Low Priority",
		TemplateID: "template1",
		Priority:   5,
		Active:     true,
		Conditions: models.Conditions{},
	}

	highPriorityRule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "High Priority",
		TemplateID: "template2",
		Priority:   20,
		Active:     true,
		Conditions: models.Conditions{},
	}

	store.CreateRule(lowPriorityRule)
	store.CreateRule(highPriorityRule)

	message := &models.TestMessage{
		SenderEmail:     "sender@test.com",
		RecipientEmails: []string{"recipient@test.com"},
		MessageType:     models.MessageTypeNew,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 2 {
		t.Errorf("Expected 2 matched rules, got %d", len(result.MatchedRules))
	}

	// The selected rule should be the high priority one
	if result.SelectedRule.Name != "High Priority" {
		t.Errorf("Expected selected rule to be 'High Priority', got '%s'", result.SelectedRule.Name)
	}
}

func TestRuleStore_Exclusivity(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	// Create an exclusive rule
	exclusiveRule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Exclusive Rule",
		TemplateID: "template1",
		Priority:   10,
		Exclusive:  true,
		Active:     true,
		Conditions: models.Conditions{
			SenderDomains: []string{"example.com"},
		},
	}

	// Create another rule that would also match
	otherRule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Other Rule",
		TemplateID: "template2",
		Priority:   5,
		Active:     true,
		Conditions: models.Conditions{},
	}

	store.CreateRule(exclusiveRule)
	store.CreateRule(otherRule)

	message := &models.TestMessage{
		SenderEmail:     "user@example.com",
		RecipientEmails: []string{"recipient@test.com"},
		MessageType:     models.MessageTypeNew,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	// Only the exclusive rule should match (other rules not evaluated after exclusive match)
	if len(result.MatchedRules) != 1 {
		t.Errorf("Expected 1 matched rule (exclusive), got %d", len(result.MatchedRules))
	}

	if result.SelectedRule.Name != "Exclusive Rule" {
		t.Errorf("Expected selected rule to be 'Exclusive Rule', got '%s'", result.SelectedRule.Name)
	}
}

func TestRuleStore_InactiveRules(t *testing.T) {
	scheduleStore := NewScheduleStore()
	store := NewRuleStore(scheduleStore)

	rule := &models.Rule{
		TenantID:   "tenant1",
		Name:       "Inactive Rule",
		TemplateID: "template1",
		Priority:   10,
		Active:     false,
		Conditions: models.Conditions{},
	}

	store.CreateRule(rule)

	message := &models.TestMessage{
		SenderEmail:     "sender@test.com",
		RecipientEmails: []string{"recipient@test.com"},
		MessageType:     models.MessageTypeNew,
		Timestamp:       time.Now(),
	}

	result, err := store.EvaluateRules("tenant1", message)
	if err != nil {
		t.Fatalf("Failed to evaluate rules: %v", err)
	}

	if len(result.MatchedRules) != 0 {
		t.Errorf("Expected 0 matched rules (inactive), got %d", len(result.MatchedRules))
	}
}
