package models

import "time"

// Template represents a signature template
type Template struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	HTMLContent string    `json:"html_content"`
	RTFContent  string    `json:"rtf_content"`
	TextContent string    `json:"text_content"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AgentTemplateResponse represents the response for agent template requests
type AgentTemplateResponse struct {
	Templates []RenderedTemplate `json:"templates"`
	UserEmail string             `json:"user_email"`
	UserID    string             `json:"user_id"`
}

// RenderedTemplate represents a template rendered with user-specific data
type RenderedTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	HTMLContent string `json:"html_content"`
	RTFContent  string `json:"rtf_content"`
	TextContent string `json:"text_content"`
}
