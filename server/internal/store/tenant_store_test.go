package store

import (
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/models"
)

func TestTenantStore_Create(t *testing.T) {
	store := NewTenantStore()
	tenant := &models.Tenant{
		ID:     "test-tenant-1",
		Name:   "Test Tenant",
		Domain: "test.com",
		Active: true,
	}

	err := store.Create(tenant)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify timestamps were set
	if tenant.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if tenant.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}

	// Try to create duplicate
	err = store.Create(tenant)
	if err != ErrTenantAlreadyExists {
		t.Errorf("expected ErrTenantAlreadyExists, got %v", err)
	}
}

func TestTenantStore_Get(t *testing.T) {
	store := NewTenantStore()
	tenant := &models.Tenant{
		ID:     "test-tenant-1",
		Name:   "Test Tenant",
		Domain: "test.com",
		Active: true,
	}

	// Get non-existent tenant
	_, err := store.Get(tenant.ID)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}

	// Create and get
	_ = store.Create(tenant)
	retrieved, err := store.Get(tenant.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if retrieved.ID != tenant.ID || retrieved.Name != tenant.Name {
		t.Error("retrieved tenant doesn't match created tenant")
	}
}

func TestTenantStore_List(t *testing.T) {
	store := NewTenantStore()

	// Empty list
	tenants := store.List()
	if len(tenants) != 0 {
		t.Errorf("expected empty list, got %d items", len(tenants))
	}

	// Add tenants
	tenant1 := &models.Tenant{ID: "tenant-1", Name: "Tenant 1", Domain: "t1.com", Active: true}
	tenant2 := &models.Tenant{ID: "tenant-2", Name: "Tenant 2", Domain: "t2.com", Active: true}
	_ = store.Create(tenant1)
	_ = store.Create(tenant2)

	tenants = store.List()
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(tenants))
	}
}

func TestTenantStore_Update(t *testing.T) {
	store := NewTenantStore()
	tenant := &models.Tenant{
		ID:     "test-tenant-1",
		Name:   "Test Tenant",
		Domain: "test.com",
		Active: true,
	}

	// Update non-existent tenant
	err := store.Update(tenant)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}

	// Create and update
	_ = store.Create(tenant)
	originalUpdatedAt := tenant.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure time difference

	tenant.Name = "Updated Tenant"
	err = store.Update(tenant)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	retrieved, _ := store.Get(tenant.ID)
	if retrieved.Name != "Updated Tenant" {
		t.Error("tenant name was not updated")
	}
	if !retrieved.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt was not refreshed")
	}
}

func TestTenantStore_Delete(t *testing.T) {
	store := NewTenantStore()
	tenant := &models.Tenant{
		ID:     "test-tenant-1",
		Name:   "Test Tenant",
		Domain: "test.com",
		Active: true,
	}

	// Delete non-existent tenant
	err := store.Delete(tenant.ID)
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}

	// Create and delete
	_ = store.Create(tenant)
	err = store.Delete(tenant.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify deletion
	_, err = store.Get(tenant.ID)
	if err != ErrTenantNotFound {
		t.Error("expected tenant to be deleted")
	}
}
