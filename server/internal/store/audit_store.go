package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/opensig/server/internal/models"
)

// AuditStore manages immutable audit log entries
type AuditStore struct {
	mu      sync.RWMutex
	entries []*models.AuditEntry // Append-only slice
}

// NewAuditStore creates a new audit store
func NewAuditStore() *AuditStore {
	return &AuditStore{
		entries: make([]*models.AuditEntry, 0),
	}
}

// LogEntry appends a new audit entry (immutable, append-only)
func (s *AuditStore) LogEntry(entry *models.AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Append-only: create a copy to ensure immutability
	entryCopy := *entry
	s.entries = append(s.entries, &entryCopy)

	return nil
}

// GetEntries retrieves audit entries with optional filters
func (s *AuditStore) GetEntries(filters AuditFilters) ([]*models.AuditEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*models.AuditEntry

	for _, entry := range s.entries {
		if s.matchesFilters(entry, filters) {
			// Return a copy to prevent external modification
			entryCopy := *entry
			filtered = append(filtered, &entryCopy)
		}
	}

	return filtered, nil
}

// GetEntriesForResource retrieves all audit entries for a specific resource
func (s *AuditStore) GetEntriesForResource(resourceType models.AuditResourceType, resourceID string) ([]*models.AuditEntry, error) {
	return s.GetEntries(AuditFilters{
		ResourceType: resourceType,
		ResourceID:   resourceID,
	})
}

// GetEntriesByTenant retrieves all audit entries for a tenant
func (s *AuditStore) GetEntriesByTenant(tenantID string) ([]*models.AuditEntry, error) {
	return s.GetEntries(AuditFilters{
		TenantID: tenantID,
	})
}

// GetEntriesByUser retrieves all audit entries for a user
func (s *AuditStore) GetEntriesByUser(userID string) ([]*models.AuditEntry, error) {
	return s.GetEntries(AuditFilters{
		UserID: userID,
	})
}

// GetEntriesInTimeRange retrieves audit entries within a time range
func (s *AuditStore) GetEntriesInTimeRange(start, end time.Time) ([]*models.AuditEntry, error) {
	return s.GetEntries(AuditFilters{
		StartTime: start,
		EndTime:   end,
	})
}

// matchesFilters checks if an entry matches the provided filters
func (s *AuditStore) matchesFilters(entry *models.AuditEntry, filters AuditFilters) bool {
	if filters.TenantID != "" && entry.TenantID != filters.TenantID {
		return false
	}
	if filters.ResourceType != "" && entry.ResourceType != filters.ResourceType {
		return false
	}
	if filters.ResourceID != "" && entry.ResourceID != filters.ResourceID {
		return false
	}
	if filters.Action != "" && entry.Action != filters.Action {
		return false
	}
	if filters.UserID != "" && entry.UserID != filters.UserID {
		return false
	}
	if !filters.StartTime.IsZero() && entry.Timestamp.Before(filters.StartTime) {
		return false
	}
	if !filters.EndTime.IsZero() && entry.Timestamp.After(filters.EndTime) {
		return false
	}
	return true
}

// Count returns the total number of audit entries
func (s *AuditStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// AuditFilters defines filters for querying audit entries
type AuditFilters struct {
	TenantID     string
	ResourceType models.AuditResourceType
	ResourceID   string
	Action       models.AuditAction
	UserID       string
	StartTime    time.Time
	EndTime      time.Time
}

// CreateAuditEntry is a helper function to create an audit entry with diff
func CreateAuditEntry(tenantID string, resourceType models.AuditResourceType, resourceID string, action models.AuditAction, user *models.User, before, after interface{}) *models.AuditEntry {
	entry := &models.AuditEntry{
		TenantID:     tenantID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Action:       action,
		UserID:       user.ID,
		UserEmail:    user.Email,
		UserRole:     user.Role,
		Timestamp:    time.Now(),
	}

	// Add changes if both before and after are provided
	if before != nil || after != nil {
		entry.Changes = &models.AuditChanges{}
		if before != nil {
			entry.Changes.Before = toMap(before)
		}
		if after != nil {
			entry.Changes.After = toMap(after)
		}
	}

	return entry
}

// toMap converts an interface to a map for diff comparison
func toMap(v interface{}) map[string]interface{} {
	// Simple type conversion - in production, use reflection or JSON marshal/unmarshal
	switch val := v.(type) {
	case map[string]interface{}:
		return val
	case *models.Template:
		return map[string]interface{}{
			"id":              val.ID,
			"tenant_id":       val.TenantID,
			"name":            val.Name,
			"html_content":    val.HTMLContent,
			"rtf_content":     val.RTFContent,
			"text_content":    val.TextContent,
			"active":          val.Active,
			"status":          val.Status,
			"submitted_by":    val.SubmittedBy,
			"submitted_at":    val.SubmittedAt,
			"reviewed_by":     val.ReviewedBy,
			"reviewed_at":     val.ReviewedAt,
			"review_comments": val.ReviewComments,
			"version":         val.Version,
		}
	default:
		return map[string]interface{}{
			"value": fmt.Sprintf("%v", v),
		}
	}
}
