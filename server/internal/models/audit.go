package models

import "time"

// AuditAction represents the type of action performed
type AuditAction string

const (
	AuditActionCreate         AuditAction = "create"
	AuditActionUpdate         AuditAction = "update"
	AuditActionDelete         AuditAction = "delete"
	AuditActionSubmitReview   AuditAction = "submit_review"
	AuditActionApprove        AuditAction = "approve"
	AuditActionReject         AuditAction = "reject"
	AuditActionPublish        AuditAction = "publish"
	AuditActionUnpublish      AuditAction = "unpublish"
)

// AuditResourceType represents the type of resource being audited
type AuditResourceType string

const (
	AuditResourceTypeTemplate AuditResourceType = "template"
	AuditResourceTypeTenant   AuditResourceType = "tenant"
	AuditResourceTypeRule     AuditResourceType = "rule"
	AuditResourceTypeSchedule AuditResourceType = "schedule"
)

// AuditEntry represents an immutable audit log entry
type AuditEntry struct {
	ID           string            `json:"id"`
	TenantID     string            `json:"tenant_id"`
	ResourceType AuditResourceType `json:"resource_type"`
	ResourceID   string            `json:"resource_id"`
	Action       AuditAction       `json:"action"`
	UserID       string            `json:"user_id"`
	UserEmail    string            `json:"user_email"`
	UserRole     Role              `json:"user_role"`
	Timestamp    time.Time         `json:"timestamp"`
	Changes      *AuditChanges     `json:"changes,omitempty"` // Diff of changes (optional)
	Metadata     map[string]string `json:"metadata,omitempty"` // Additional context
}

// AuditChanges represents the before/after state of a resource
type AuditChanges struct {
	Before map[string]interface{} `json:"before,omitempty"`
	After  map[string]interface{} `json:"after,omitempty"`
}
