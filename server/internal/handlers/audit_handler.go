package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/your-org/opensig/server/internal/middleware"
	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

// AuditHandler handles audit log HTTP requests
type AuditHandler struct {
	auditStore *store.AuditStore
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditStore *store.AuditStore) *AuditHandler {
	return &AuditHandler{
		auditStore: auditStore,
	}
}

// ListAuditEntries lists audit entries with optional filters
func (h *AuditHandler) ListAuditEntries(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Parse query parameters for filtering
	filters := store.AuditFilters{
		TenantID: user.TenantID, // Always filter by user's tenant
	}

	if resourceType := r.URL.Query().Get("resource_type"); resourceType != "" {
		filters.ResourceType = models.AuditResourceType(resourceType)
	}

	if resourceID := r.URL.Query().Get("resource_id"); resourceID != "" {
		filters.ResourceID = resourceID
	}

	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = models.AuditAction(action)
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filters.UserID = userID
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filters.StartTime = t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filters.EndTime = t
		}
	}

	entries, err := h.auditStore.GetEntries(filters)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

// GetAuditEntriesForResource gets all audit entries for a specific resource
func (h *AuditHandler) GetAuditEntriesForResource(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	// Parse path: /v1/audit/resource/{type}/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/audit/resource/"), "/")
	if len(parts) != 2 {
		http.Error(w, "invalid resource path", http.StatusBadRequest)
		return
	}

	resourceType := models.AuditResourceType(parts[0])
	resourceID := parts[1]

	entries, err := h.auditStore.GetEntriesForResource(resourceType, resourceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter by tenant for security
	var filteredEntries []*models.AuditEntry
	for _, entry := range entries {
		if entry.TenantID == user.TenantID {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"entries":       filteredEntries,
		"count":         len(filteredEntries),
	})
}

// GetAuditStats returns statistics about audit entries
func (h *AuditHandler) GetAuditStats(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(middleware.UserContextKey).(*models.User)

	entries, err := h.auditStore.GetEntriesByTenant(user.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	stats := map[string]interface{}{
		"total_entries": len(entries),
		"by_action":     make(map[models.AuditAction]int),
		"by_resource":   make(map[models.AuditResourceType]int),
		"by_user":       make(map[string]int),
	}

	byAction := stats["by_action"].(map[models.AuditAction]int)
	byResource := stats["by_resource"].(map[models.AuditResourceType]int)
	byUser := stats["by_user"].(map[string]int)

	for _, entry := range entries {
		byAction[entry.Action]++
		byResource[entry.ResourceType]++
		byUser[entry.UserEmail]++
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}
