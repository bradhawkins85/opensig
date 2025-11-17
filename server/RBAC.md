# Multi-Tenant & RBAC Implementation

This document describes the multi-tenant data model and role-based access control (RBAC) implementation for OpenSig.

## Overview

OpenSig now supports multi-tenancy with role-based access control. Each tenant represents an organization using the system, and users are assigned roles that determine their access permissions.

## RBAC Roles

The following roles are supported:

- **`org_admin`**: Organization administrator with full access to tenant management and configuration
- **`signature_admin`**: Signature administrator who can manage signatures and templates
- **`approver`**: User who can approve signature changes and deployments
- **`auditor`**: Read-only access for audit and compliance purposes

## Tenant Model

A tenant represents an organization and includes:

```json
{
  "id": "tenant-1",
  "name": "Acme Corporation",
  "domain": "acme.com",
  "active": true,
  "created_at": "2025-11-17T07:14:01.733Z",
  "updated_at": "2025-11-17T07:14:01.733Z"
}
```

## API Endpoints

### Tenant Management (Requires `org_admin` role)

#### Create Tenant
```bash
POST /v1/tenants
Content-Type: application/json
X-User-ID: admin
X-User-Role: org_admin

{
  "id": "tenant-1",
  "name": "Test Tenant",
  "domain": "test.com",
  "active": true
}
```

#### List Tenants
```bash
GET /v1/tenants
X-User-ID: admin
X-User-Role: org_admin
```

#### Get Tenant by ID
```bash
GET /v1/tenants/{id}
X-User-ID: admin
X-User-Role: org_admin
```

#### Update Tenant
```bash
PUT /v1/tenants/{id}
Content-Type: application/json
X-User-ID: admin
X-User-Role: org_admin

{
  "name": "Updated Tenant Name",
  "domain": "newdomain.com",
  "active": true
}
```

#### Delete Tenant
```bash
DELETE /v1/tenants/{id}
X-User-ID: admin
X-User-Role: org_admin
```

### Sample RBAC-Protected Endpoint

#### Admin Configuration (Requires `org_admin` role)
```bash
GET /v1/admin/config
X-User-ID: admin
X-User-Role: org_admin
```

## Testing RBAC

The system uses HTTP headers for authentication in development/testing:

- `X-User-ID`: User identifier
- `X-User-Role`: One of: `org_admin`, `signature_admin`, `approver`, `auditor`

### Examples

**Unauthorized Access (no headers):**
```bash
$ curl http://localhost:8080/v1/tenants
{"error":"unauthorized"}
```

**Forbidden Access (insufficient role):**
```bash
$ curl -H "X-User-ID: user1" -H "X-User-Role: auditor" \
  http://localhost:8080/v1/tenants
{"error":"forbidden: insufficient permissions"}
```

**Authorized Access:**
```bash
$ curl -H "X-User-ID: admin" -H "X-User-Role: org_admin" \
  http://localhost:8080/v1/tenants
{"tenants":[...],"count":2}
```

## Architecture

### Components

1. **Models** (`internal/models/tenant.go`)
   - Defines `Tenant`, `User`, and `Role` data structures

2. **Store** (`internal/store/tenant_store.go`)
   - In-memory storage for tenants (foundation for future database integration)
   - Thread-safe CRUD operations

3. **Middleware** (`internal/middleware/rbac.go`)
   - `RequireRole`: Enforces role-based access control
   - `MockAuthMiddleware`: Simulates authentication for testing

4. **Handlers** (`internal/handlers/tenant_handler.go`)
   - HTTP handlers for tenant CRUD operations

### Data Flow

```
Request → MockAuthMiddleware → RequireRole → Handler → Store
           (sets user context)  (checks role)  (processes)
```

## Testing

Run the test suite:

```bash
cd server
go test ./...
```

Test coverage includes:
- Tenant store operations (Create, Read, Update, Delete, List)
- RBAC middleware (authorized, unauthorized, forbidden cases)
- HTTP handlers (all CRUD endpoints with various scenarios)

## Future Enhancements

1. **Database Integration**: Replace in-memory store with PostgreSQL
2. **JWT Authentication**: Replace mock auth with real JWT token validation
3. **Enhanced RBAC**: Add permission-based access control beyond roles
4. **Audit Logging**: Track all tenant operations for compliance
5. **Multi-role Support**: Allow users to have multiple roles
6. **Tenant Isolation**: Ensure data isolation between tenants at the database level

## Production Considerations

- The current implementation uses mock authentication for development
- In production, implement proper JWT/OAuth2 authentication
- Store tenant data in PostgreSQL with proper indexes
- Add rate limiting and request validation
- Implement audit logging for compliance
- Use HTTPS and secure headers
- Consider implementing tenant-specific database schemas or row-level security
