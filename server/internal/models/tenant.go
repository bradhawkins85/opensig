package models

import "time"

// Tenant represents a multi-tenant organization
type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Role represents the RBAC role types
type Role string

const (
	RoleOrgAdmin       Role = "org_admin"
	RoleSignatureAdmin Role = "signature_admin"
	RoleApprover       Role = "approver"
	RoleAuditor        Role = "auditor"
)

// User represents a user with tenant and role assignments
type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id"`
	Role     Role   `json:"role"`
}
