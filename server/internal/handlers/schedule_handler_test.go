package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/models"
	"github.com/your-org/opensig/server/internal/store"
)

func TestScheduleHandler_CreateSchedule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	handler := NewScheduleHandler(scheduleStore)

	schedule := models.Schedule{
		TenantID:    "tenant1",
		Name:        "Test Schedule",
		Description: "Test description",
		Active:      true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
			},
		},
	}

	body, _ := json.Marshal(schedule)
	req := httptest.NewRequest(http.MethodPost, "/v1/schedules", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.CreateSchedule(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var created models.Schedule
	json.NewDecoder(w.Body).Decode(&created)

	if created.ID == "" {
		t.Error("Expected schedule ID to be generated")
	}

	if created.Name != schedule.Name {
		t.Errorf("Expected name %s, got %s", schedule.Name, created.Name)
	}
}

func TestScheduleHandler_GetSchedule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	handler := NewScheduleHandler(scheduleStore)

	schedule := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "Test Schedule",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}
	scheduleStore.CreateSchedule(schedule)

	req := httptest.NewRequest(http.MethodGet, "/v1/schedules/"+schedule.ID, nil)
	w := httptest.NewRecorder()

	handler.GetSchedule(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var retrieved models.Schedule
	json.NewDecoder(w.Body).Decode(&retrieved)

	if retrieved.ID != schedule.ID {
		t.Errorf("Expected ID %s, got %s", schedule.ID, retrieved.ID)
	}
}

func TestScheduleHandler_ListSchedules(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	handler := NewScheduleHandler(scheduleStore)

	// Create two schedules for the same tenant
	schedule1 := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "Schedule 1",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}
	schedule2 := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "Schedule 2",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}
	scheduleStore.CreateSchedule(schedule1)
	scheduleStore.CreateSchedule(schedule2)

	req := httptest.NewRequest(http.MethodGet, "/v1/schedules?tenant_id=tenant1", nil)
	w := httptest.NewRecorder()

	handler.ListSchedules(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var schedules []*models.Schedule
	json.NewDecoder(w.Body).Decode(&schedules)

	if len(schedules) != 2 {
		t.Errorf("Expected 2 schedules, got %d", len(schedules))
	}
}

func TestScheduleHandler_UpdateSchedule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	handler := NewScheduleHandler(scheduleStore)

	schedule := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "Original Name",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}
	scheduleStore.CreateSchedule(schedule)

	// Update the schedule
	schedule.Name = "Updated Name"
	body, _ := json.Marshal(schedule)
	req := httptest.NewRequest(http.MethodPut, "/v1/schedules/"+schedule.ID, bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateSchedule(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var updated models.Schedule
	json.NewDecoder(w.Body).Decode(&updated)

	if updated.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updated.Name)
	}
}

func TestScheduleHandler_DeleteSchedule(t *testing.T) {
	scheduleStore := store.NewScheduleStore()
	handler := NewScheduleHandler(scheduleStore)

	schedule := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "To Delete",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}
	scheduleStore.CreateSchedule(schedule)

	req := httptest.NewRequest(http.MethodDelete, "/v1/schedules/"+schedule.ID, nil)
	w := httptest.NewRecorder()

	handler.DeleteSchedule(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify deletion
	_, err := scheduleStore.GetSchedule(schedule.ID)
	if err == nil {
		t.Error("Expected error when getting deleted schedule")
	}
}
