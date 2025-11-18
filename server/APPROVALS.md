# M4: Approvals Workflow & Immutable Audit Log

## Overview

This document describes the approvals workflow and immutable audit log implementation for OpenSig template management, completing milestone M4.

## Features

### 1. Approval Workflow

Templates now go through a three-stage approval process: **Draft → Review → Publish**

#### Approval Status

- **`draft`**: Initial state when a template is created
- **`pending_review`**: Template has been submitted for review
- **`approved`**: Template has been approved by an approver
- **`rejected`**: Template has been rejected by an approver

#### Workflow States

```
draft ──submit──> pending_review ──approve──> approved ──publish──> active
                        │
                        └──reject──> rejected
```

### 2. Immutable Audit Log

All template operations are logged in an append-only audit log with:
- **Who**: User ID, email, and role
- **What**: Action performed (create, update, delete, submit_review, approve, reject, publish, unpublish)
- **When**: Timestamp of the action
- **Diffs**: Before/after state changes for update operations

## API Endpoints

### Template Management

#### Create Template (Draft)
```bash
POST /v1/templates
X-User-Role: signature_admin

{
  "name": "Holiday Template",
  "html_content": "<p>Happy Holidays!</p>",
  "rtf_content": "{\\rtf Happy Holidays!}",
  "text_content": "Happy Holidays!"
}

Response: Template with status="draft", version=1
```

#### List Templates
```bash
GET /v1/templates?status=draft
X-User-Role: signature_admin

Response: {templates: [...], count: N}
```

#### Get Template
```bash
GET /v1/templates/{id}
X-User-Role: signature_admin

Response: Template object
```

#### Update Template
```bash
PUT /v1/templates/{id}
X-User-Role: signature_admin

{template fields}

Response: Updated template
```

#### Delete Template
```bash
DELETE /v1/templates/{id}
X-User-Role: signature_admin

Response: 204 No Content
```

### Approval Workflow Actions

#### Submit for Review
```bash
POST /v1/templates/{id}/submit
X-User-Role: signature_admin

Response: Template with status="pending_review"
```

#### Approve Template
```bash
POST /v1/templates/{id}/approve
X-User-Role: approver

{
  "comments": "Looks great!"
}

Response: Template with status="approved"
```

#### Reject Template
```bash
POST /v1/templates/{id}/reject
X-User-Role: approver

{
  "comments": "Needs improvement"
}

Response: Template with status="rejected"
```

#### Publish Template (Make Active)
```bash
POST /v1/templates/{id}/publish
X-User-Role: signature_admin

Response: Template with active=true
```

#### Unpublish Template (Make Inactive)
```bash
POST /v1/templates/{id}/unpublish
X-User-Role: signature_admin

Response: Template with active=false
```

### Audit Log Endpoints

#### List All Audit Entries
```bash
GET /v1/audit?resource_type=template&action=approve&start_time=2025-11-18T00:00:00Z
X-User-Role: auditor

Response: {entries: [...], count: N}
```

Available filters:
- `resource_type`: template, tenant, rule, schedule
- `resource_id`: Specific resource ID
- `action`: create, update, delete, submit_review, approve, reject, publish, unpublish
- `user_id`: Filter by user
- `start_time`: ISO 8601 timestamp
- `end_time`: ISO 8601 timestamp

#### Get Audit Entries for Resource
```bash
GET /v1/audit/resource/template/{template_id}
X-User-Role: auditor

Response: {resource_type, resource_id, entries: [...], count: N}
```

#### Get Audit Statistics
```bash
GET /v1/audit/stats
X-User-Role: auditor

Response: {
  total_entries: N,
  by_action: {...},
  by_resource: {...},
  by_user: {...}
}
```

## RBAC Permissions

| Action | Required Role |
|--------|---------------|
| Create/Update/Delete Template | `signature_admin` or `org_admin` |
| Submit for Review | `signature_admin` or `org_admin` |
| Approve/Reject Template | `approver` or `org_admin` |
| Publish/Unpublish Template | `signature_admin` or `org_admin` |
| View Audit Logs | `auditor` or `org_admin` |

## Data Model

### Template Model

```go
type Template struct {
    ID             string         `json:"id"`
    TenantID       string         `json:"tenant_id"`
    Name           string         `json:"name"`
    HTMLContent    string         `json:"html_content"`
    RTFContent     string         `json:"rtf_content"`
    TextContent    string         `json:"text_content"`
    Active         bool           `json:"active"`
    Status         ApprovalStatus `json:"status"`          // draft, pending_review, approved, rejected
    SubmittedBy    string         `json:"submitted_by"`    // User ID who submitted
    SubmittedAt    *time.Time     `json:"submitted_at"`    // When submitted
    ReviewedBy     string         `json:"reviewed_by"`     // User ID who approved/rejected
    ReviewedAt     *time.Time     `json:"reviewed_at"`     // When reviewed
    ReviewComments string         `json:"review_comments"` // Reviewer comments
    Version        int            `json:"version"`         // Version number
    CreatedAt      time.Time      `json:"created_at"`
    UpdatedAt      time.Time      `json:"updated_at"`
}
```

### Audit Entry Model

```go
type AuditEntry struct {
    ID           string            `json:"id"`
    TenantID     string            `json:"tenant_id"`
    ResourceType AuditResourceType `json:"resource_type"` // template, tenant, rule, schedule
    ResourceID   string            `json:"resource_id"`
    Action       AuditAction       `json:"action"`        // create, update, delete, etc.
    UserID       string            `json:"user_id"`
    UserEmail    string            `json:"user_email"`
    UserRole     Role              `json:"user_role"`
    Timestamp    time.Time         `json:"timestamp"`
    Changes      *AuditChanges     `json:"changes,omitempty"` // Before/after diff
    Metadata     map[string]string `json:"metadata,omitempty"` // Additional context
}

type AuditChanges struct {
    Before map[string]interface{} `json:"before,omitempty"`
    After  map[string]interface{} `json:"after,omitempty"`
}
```

## Example Workflow

### Creating and Publishing a Template

```bash
# 1. Create a draft template (signature_admin)
curl -X POST http://localhost:8080/v1/templates \
  -H "X-User-ID: admin1" \
  -H "X-User-Role: signature_admin" \
  -d '{"name": "Holiday 2025", "html_content": "<p>Happy Holidays!</p>", ...}'

# Response: {id: "abc123", status: "draft", version: 1}

# 2. Submit for review (signature_admin)
curl -X POST http://localhost:8080/v1/templates/abc123/submit \
  -H "X-User-ID: admin1" \
  -H "X-User-Role: signature_admin"

# Response: {id: "abc123", status: "pending_review", submitted_by: "admin1"}

# 3. Approve template (approver)
curl -X POST http://localhost:8080/v1/templates/abc123/approve \
  -H "X-User-ID: approver1" \
  -H "X-User-Role: approver" \
  -d '{"comments": "Looks great!"}'

# Response: {id: "abc123", status: "approved", reviewed_by: "approver1"}

# 4. Publish template (signature_admin)
curl -X POST http://localhost:8080/v1/templates/abc123/publish \
  -H "X-User-ID: admin1" \
  -H "X-User-Role: signature_admin"

# Response: {id: "abc123", status: "approved", active: true}

# 5. View audit trail (auditor)
curl http://localhost:8080/v1/audit/resource/template/abc123 \
  -H "X-User-ID: auditor1" \
  -H "X-User-Role: auditor"

# Response: {
#   entries: [
#     {action: "create", user_email: "admin1@...", timestamp: "..."},
#     {action: "submit_review", user_email: "admin1@...", timestamp: "..."},
#     {action: "approve", user_email: "approver1@...", timestamp: "...", metadata: {comments: "Looks great!"}},
#     {action: "publish", user_email: "admin1@...", timestamp: "..."}
#   ]
# }
```

## Testing

Run the test suite:

```bash
cd server
go test ./...
```

Test coverage includes:
- Template approval workflow (draft → review → approve → publish)
- Template rejection workflow
- Audit log creation and querying
- RBAC permissions enforcement
- Immutability of audit entries
- Before/after state tracking

## Security Considerations

1. **Immutable Audit Log**: Audit entries are append-only and cannot be modified or deleted
2. **RBAC Enforcement**: All endpoints enforce role-based access control
3. **Tenant Isolation**: Users can only access resources within their tenant
4. **State Validation**: Workflow state transitions are validated (e.g., can't approve a draft template)

## Future Enhancements

1. **Database Persistence**: Replace in-memory stores with PostgreSQL
2. **Audit Log Retention**: Implement retention policies and archival
3. **Email Notifications**: Notify approvers when templates are submitted for review
4. **Bulk Operations**: Support bulk approval/rejection of templates
5. **Audit Log Export**: Export audit logs to CSV/JSON for compliance
6. **Template Versioning**: Full version history with rollback capability
