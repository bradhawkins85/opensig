package store

import (
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/models"
)

func TestAuditStore_LogEntry(t *testing.T) {
	store := NewAuditStore()
	user := &models.User{
		ID:       "user1",
		Email:    "test@example.com",
		TenantID: "tenant1",
		Role:     models.RoleSignatureAdmin,
	}

	entry := CreateAuditEntry(
		"tenant1",
		models.AuditResourceTypeTemplate,
		"template1",
		models.AuditActionCreate,
		user,
		nil,
		map[string]interface{}{"name": "Test Template"},
	)

	err := store.LogEntry(entry)
	if err != nil {
		t.Fatalf("Failed to log entry: %v", err)
	}

	if store.Count() != 1 {
		t.Errorf("Expected 1 entry, got %d", store.Count())
	}
}

func TestAuditStore_GetEntriesByTenant(t *testing.T) {
	store := NewAuditStore()
	user1 := &models.User{ID: "user1", Email: "user1@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}
	user2 := &models.User{ID: "user2", Email: "user2@example.com", TenantID: "tenant2", Role: models.RoleSignatureAdmin}

	entry1 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t1", models.AuditActionCreate, user1, nil, nil)
	entry2 := CreateAuditEntry("tenant2", models.AuditResourceTypeTemplate, "t2", models.AuditActionCreate, user2, nil, nil)
	entry3 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t3", models.AuditActionUpdate, user1, nil, nil)

	_ = store.LogEntry(entry1)
	_ = store.LogEntry(entry2)
	_ = store.LogEntry(entry3)

	entries, err := store.GetEntriesByTenant("tenant1")
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries for tenant1, got %d", len(entries))
	}
}

func TestAuditStore_GetEntriesForResource(t *testing.T) {
	store := NewAuditStore()
	user := &models.User{ID: "user1", Email: "test@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}

	entry1 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "template1", models.AuditActionCreate, user, nil, nil)
	entry2 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "template1", models.AuditActionUpdate, user, nil, nil)
	entry3 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "template2", models.AuditActionCreate, user, nil, nil)

	_ = store.LogEntry(entry1)
	_ = store.LogEntry(entry2)
	_ = store.LogEntry(entry3)

	entries, err := store.GetEntriesForResource(models.AuditResourceTypeTemplate, "template1")
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries for template1, got %d", len(entries))
	}
}

func TestAuditStore_GetEntriesByUser(t *testing.T) {
	store := NewAuditStore()
	user1 := &models.User{ID: "user1", Email: "user1@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}
	user2 := &models.User{ID: "user2", Email: "user2@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}

	entry1 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t1", models.AuditActionCreate, user1, nil, nil)
	entry2 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t2", models.AuditActionCreate, user2, nil, nil)
	entry3 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t3", models.AuditActionUpdate, user1, nil, nil)

	_ = store.LogEntry(entry1)
	_ = store.LogEntry(entry2)
	_ = store.LogEntry(entry3)

	entries, err := store.GetEntriesByUser("user1")
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries for user1, got %d", len(entries))
	}
}

func TestAuditStore_GetEntriesInTimeRange(t *testing.T) {
	store := NewAuditStore()
	user := &models.User{ID: "user1", Email: "test@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	entry1 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t1", models.AuditActionCreate, user, nil, nil)
	entry1.Timestamp = past

	entry2 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t2", models.AuditActionCreate, user, nil, nil)
	entry2.Timestamp = now

	entry3 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t3", models.AuditActionCreate, user, nil, nil)
	entry3.Timestamp = future

	_ = store.LogEntry(entry1)
	_ = store.LogEntry(entry2)
	_ = store.LogEntry(entry3)

	// Get entries in range: past to now
	entries, err := store.GetEntriesInTimeRange(past.Add(-1*time.Minute), now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in time range, got %d", len(entries))
	}
}

func TestAuditStore_GetEntriesWithFilters(t *testing.T) {
	store := NewAuditStore()
	user := &models.User{ID: "user1", Email: "test@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}

	entry1 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t1", models.AuditActionCreate, user, nil, nil)
	entry2 := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t2", models.AuditActionUpdate, user, nil, nil)
	entry3 := CreateAuditEntry("tenant1", models.AuditResourceTypeRule, "r1", models.AuditActionCreate, user, nil, nil)

	_ = store.LogEntry(entry1)
	_ = store.LogEntry(entry2)
	_ = store.LogEntry(entry3)

	// Filter by resource type and action
	filters := AuditFilters{
		TenantID:     "tenant1",
		ResourceType: models.AuditResourceTypeTemplate,
		Action:       models.AuditActionCreate,
	}

	entries, err := store.GetEntries(filters)
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry matching filters, got %d", len(entries))
	}

	if entries[0].ResourceID != "t1" {
		t.Errorf("Expected entry with resource_id t1, got %s", entries[0].ResourceID)
	}
}

func TestAuditStore_ImmutableEntries(t *testing.T) {
	store := NewAuditStore()
	user := &models.User{ID: "user1", Email: "test@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}

	entry := CreateAuditEntry("tenant1", models.AuditResourceTypeTemplate, "t1", models.AuditActionCreate, user, nil, nil)
	originalAction := entry.Action

	_ = store.LogEntry(entry)

	// Attempt to modify the original entry
	entry.Action = models.AuditActionDelete

	// Retrieve the entry from store
	entries, _ := store.GetEntriesForResource(models.AuditResourceTypeTemplate, "t1")

	// Verify the stored entry is unchanged
	if entries[0].Action != originalAction {
		t.Errorf("Entry was modified! Expected %s, got %s", originalAction, entries[0].Action)
	}
}

func TestAuditStore_WithChanges(t *testing.T) {
	store := NewAuditStore()
	user := &models.User{ID: "user1", Email: "test@example.com", TenantID: "tenant1", Role: models.RoleSignatureAdmin}

	before := map[string]interface{}{"name": "Old Name", "active": false}
	after := map[string]interface{}{"name": "New Name", "active": true}

	entry := CreateAuditEntry(
		"tenant1",
		models.AuditResourceTypeTemplate,
		"t1",
		models.AuditActionUpdate,
		user,
		before,
		after,
	)

	_ = store.LogEntry(entry)

	entries, _ := store.GetEntriesForResource(models.AuditResourceTypeTemplate, "t1")

	if entries[0].Changes == nil {
		t.Fatal("Expected changes to be recorded")
	}

	if entries[0].Changes.Before["name"] != "Old Name" {
		t.Errorf("Expected before name 'Old Name', got %v", entries[0].Changes.Before["name"])
	}

	if entries[0].Changes.After["name"] != "New Name" {
		t.Errorf("Expected after name 'New Name', got %v", entries[0].Changes.After["name"])
	}
}
