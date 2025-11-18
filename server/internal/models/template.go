package models

import "time"

// ApprovalStatus represents the approval state of a template
type ApprovalStatus string

const (
	ApprovalStatusDraft         ApprovalStatus = "draft"
	ApprovalStatusPendingReview ApprovalStatus = "pending_review"
	ApprovalStatusApproved      ApprovalStatus = "approved"
	ApprovalStatusRejected      ApprovalStatus = "rejected"
)

// Template represents a signature template
type Template struct {
	ID             string         `json:"id"`
	TenantID       string         `json:"tenant_id"`
	Name           string         `json:"name"`
	HTMLContent    string         `json:"html_content"`
	RTFContent     string         `json:"rtf_content"`
	TextContent    string         `json:"text_content"`
	Active         bool           `json:"active"`
	Status         ApprovalStatus `json:"status"`          // Approval workflow status
	SubmittedBy    string         `json:"submitted_by"`    // User ID who submitted for review
	SubmittedAt    *time.Time     `json:"submitted_at"`    // When submitted for review
	ReviewedBy     string         `json:"reviewed_by"`     // User ID who approved/rejected
	ReviewedAt     *time.Time     `json:"reviewed_at"`     // When approved/rejected
	ReviewComments string         `json:"review_comments"` // Comments from reviewer
	Version        int            `json:"version"`         // Version number for tracking changes
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// AgentTemplateResponse represents the response for agent template requests
type AgentTemplateResponse struct {
	Templates             []RenderedTemplate `json:"templates"`
	UserEmail             string             `json:"user_email"`
	UserID                string             `json:"user_id"`
	SetDefaultSignatures  bool               `json:"set_default_signatures"`
}

// RenderedTemplate represents a template rendered with user-specific data
type RenderedTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HTMLContent string `json:"html_content"`
	RTFContent  string `json:"rtf_content"`
	TextContent string `json:"text_content"`
}
