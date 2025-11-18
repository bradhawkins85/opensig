package models

import "time"

// Rule represents a signature rule with conditions and template assignment
type Rule struct {
	ID          string       `json:"id"`
	TenantID    string       `json:"tenant_id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	TemplateID  string       `json:"template_id"`
	Conditions  Conditions   `json:"conditions"`
	ScheduleID  string       `json:"schedule_id,omitempty"` // Optional schedule reference
	Priority    int          `json:"priority"`              // Higher number = higher priority
	Exclusive   bool         `json:"exclusive"`             // If true, stop processing other rules if this matches
	Active      bool         `json:"active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Conditions represents the matching conditions for a rule
type Conditions struct {
	SenderEmails    []string    `json:"sender_emails,omitempty"`    // Match if sender is in this list
	SenderDomains   []string    `json:"sender_domains,omitempty"`   // Match if sender domain is in this list
	RecipientEmails []string    `json:"recipient_emails,omitempty"` // Match if any recipient is in this list
	RecipientDomains []string   `json:"recipient_domains,omitempty"` // Match if any recipient domain is in this list
	MessageTypes    []MessageType `json:"message_types,omitempty"`  // Match if message type is in this list
}

// MessageType represents the type of email message
type MessageType string

const (
	MessageTypeNew     MessageType = "new"       // New message
	MessageTypeReply   MessageType = "reply"     // Reply to existing message
	MessageTypeForward MessageType = "forward"   // Forwarded message
)

// TestMessage represents a message to be evaluated against rules
type TestMessage struct {
	SenderEmail     string   `json:"sender_email"`
	RecipientEmails []string `json:"recipient_emails"`
	MessageType     MessageType `json:"message_type"`
	Timestamp       time.Time `json:"timestamp"` // For schedule evaluation
}

// RuleEvaluationResult represents the result of evaluating rules against a message
type RuleEvaluationResult struct {
	MatchedRules []Rule `json:"matched_rules"`
	SelectedRule *Rule  `json:"selected_rule,omitempty"` // The rule selected after priority/exclusivity
}
