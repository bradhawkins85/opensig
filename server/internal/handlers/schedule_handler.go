package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

// ScheduleHandler handles schedule-related HTTP requests
type ScheduleHandler struct {
	scheduleStore *store.ScheduleStore
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(scheduleStore *store.ScheduleStore) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleStore: scheduleStore,
	}
}

// CreateSchedule handles POST /v1/schedules
func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var schedule models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.scheduleStore.CreateSchedule(&schedule); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(schedule)
}

// GetSchedule handles GET /v1/schedules/{id}
func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/schedules/")
	if id == "" {
		http.Error(w, "Schedule ID required", http.StatusBadRequest)
		return
	}

	schedule, err := h.scheduleStore.GetSchedule(id)
	if err != nil {
		if err == store.ErrScheduleNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// ListSchedules handles GET /v1/schedules?tenant_id={tenantID}
func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id query parameter required", http.StatusBadRequest)
		return
	}

	schedules, err := h.scheduleStore.GetSchedulesByTenantID(tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedules)
}

// UpdateSchedule handles PUT /v1/schedules/{id}
func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/schedules/")
	if id == "" {
		http.Error(w, "Schedule ID required", http.StatusBadRequest)
		return
	}

	var schedule models.Schedule
	if err := json.NewDecoder(r.Body).Decode(&schedule); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedule.ID = id
	if err := h.scheduleStore.UpdateSchedule(&schedule); err != nil {
		if err == store.ErrScheduleNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schedule)
}

// DeleteSchedule handles DELETE /v1/schedules/{id}
func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/v1/schedules/")
	if id == "" {
		http.Error(w, "Schedule ID required", http.StatusBadRequest)
		return
	}

	if err := h.scheduleStore.DeleteSchedule(id); err != nil {
		if err == store.ErrScheduleNotFound {
			http.Error(w, "Schedule not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
