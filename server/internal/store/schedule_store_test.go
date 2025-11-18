package store

import (
	"testing"
	"time"

	"github.com/your-org/opensig/server/internal/models"
)

func TestScheduleStore_CreateAndGet(t *testing.T) {
	store := NewScheduleStore()

	schedule := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "Test Schedule",
		Description: "A test schedule",
		Active:      true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
			},
		},
	}

	err := store.CreateSchedule(schedule)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	if schedule.ID == "" {
		t.Fatal("Schedule ID should be generated")
	}

	retrieved, err := store.GetSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to get schedule: %v", err)
	}

	if retrieved.Name != schedule.Name {
		t.Errorf("Expected name %s, got %s", schedule.Name, retrieved.Name)
	}
}

func TestScheduleStore_IsTimeInSchedule_DateRange(t *testing.T) {
	store := NewScheduleStore()

	schedule := &models.Schedule{
		ID:       "schedule1",
		TenantID: "tenant1",
		Name:     "Simple Date Range",
		Active:   true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
			},
		},
	}

	// Test time within range
	checkTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Time should be within schedule")
	}

	// Test time before range
	checkTime = time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Time should not be within schedule (before start)")
	}

	// Test time after range
	checkTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Time should not be within schedule (after end)")
	}
}

func TestScheduleStore_IsTimeInSchedule_TimeOfDay(t *testing.T) {
	store := NewScheduleStore()

	schedule := &models.Schedule{
		ID:       "schedule1",
		TenantID: "tenant1",
		Name:     "Business Hours",
		Active:   true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				StartTime: "09:00",
				EndTime:   "17:00",
				Timezone:  "UTC",
			},
		},
	}

	// Test time within business hours
	checkTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Time should be within schedule (business hours)")
	}

	// Test time outside business hours (early)
	checkTime = time.Date(2024, 6, 15, 8, 0, 0, 0, time.UTC)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Time should not be within schedule (before business hours)")
	}

	// Test time outside business hours (late)
	checkTime = time.Date(2024, 6, 15, 18, 0, 0, 0, time.UTC)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Time should not be within schedule (after business hours)")
	}
}

func TestScheduleStore_IsTimeInSchedule_DailyRecurrence(t *testing.T) {
	store := NewScheduleStore()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	schedule := &models.Schedule{
		ID:       "schedule1",
		TenantID: "tenant1",
		Name:     "Every 2 Days",
		Active:   true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: startDate,
				EndDate:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
				Recurrence: &models.Recurrence{
					Frequency: models.RecurrenceDaily,
					Interval:  2,
				},
			},
		},
	}

	// Day 0 (start date) should match
	checkTime := startDate
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Day 0 should be within schedule")
	}

	// Day 2 should match
	checkTime = startDate.Add(48 * time.Hour)
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Day 2 should be within schedule")
	}

	// Day 1 should not match (not a multiple of 2)
	checkTime = startDate.Add(24 * time.Hour)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Day 1 should not be within schedule")
	}
}

func TestScheduleStore_IsTimeInSchedule_WeeklyRecurrence(t *testing.T) {
	store := NewScheduleStore()

	// Start on a Monday
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // Monday

	schedule := &models.Schedule{
		ID:       "schedule1",
		TenantID: "tenant1",
		Name:     "Weekly on Monday and Friday",
		Active:   true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: startDate,
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
				Recurrence: &models.Recurrence{
					Frequency:  models.RecurrenceWeekly,
					Interval:   1,
					DaysOfWeek: []time.Weekday{time.Monday, time.Friday},
				},
			},
		},
	}

	// Monday should match
	checkTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC) // Monday
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Monday should be within schedule")
	}

	// Friday should match
	checkTime = time.Date(2024, 1, 5, 12, 0, 0, 0, time.UTC) // Friday
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Friday should be within schedule")
	}

	// Tuesday should not match
	checkTime = time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC) // Tuesday
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Tuesday should not be within schedule")
	}
}

func TestScheduleStore_IsTimeInSchedule_MonthlyRecurrence(t *testing.T) {
	store := NewScheduleStore()

	startDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC) // 15th of month

	schedule := &models.Schedule{
		ID:       "schedule1",
		TenantID: "tenant1",
		Name:     "Monthly on 15th",
		Active:   true,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: startDate,
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
				Recurrence: &models.Recurrence{
					Frequency:  models.RecurrenceMonthly,
					Interval:   1,
					DayOfMonth: 15,
				},
			},
		},
	}

	// January 15 should match
	checkTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("January 15 should be within schedule")
	}

	// February 15 should match
	checkTime = time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)
	if !store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("February 15 should be within schedule")
	}

	// January 14 should not match
	checkTime = time.Date(2024, 1, 14, 12, 0, 0, 0, time.UTC)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("January 14 should not be within schedule")
	}
}

func TestScheduleStore_UpdateSchedule(t *testing.T) {
	store := NewScheduleStore()

	schedule := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "Original Name",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}

	err := store.CreateSchedule(schedule)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	// Update the schedule
	schedule.Name = "Updated Name"
	err = store.UpdateSchedule(schedule)
	if err != nil {
		t.Fatalf("Failed to update schedule: %v", err)
	}

	// Verify update
	retrieved, err := store.GetSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to get schedule: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}
}

func TestScheduleStore_DeleteSchedule(t *testing.T) {
	store := NewScheduleStore()

	schedule := &models.Schedule{
		TenantID:    "tenant1",
		Name:        "To Delete",
		Active:      true,
		TimeWindows: []models.TimeWindow{},
	}

	err := store.CreateSchedule(schedule)
	if err != nil {
		t.Fatalf("Failed to create schedule: %v", err)
	}

	err = store.DeleteSchedule(schedule.ID)
	if err != nil {
		t.Fatalf("Failed to delete schedule: %v", err)
	}

	_, err = store.GetSchedule(schedule.ID)
	if err == nil {
		t.Error("Expected error when getting deleted schedule")
	}
}

func TestScheduleStore_InactiveSchedule(t *testing.T) {
	store := NewScheduleStore()

	schedule := &models.Schedule{
		ID:       "schedule1",
		TenantID: "tenant1",
		Name:     "Inactive Schedule",
		Active:   false,
		TimeWindows: []models.TimeWindow{
			{
				StartDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
				Timezone:  "UTC",
			},
		},
	}

	checkTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)
	if store.IsTimeInSchedule(schedule, checkTime) {
		t.Error("Inactive schedule should not match any time")
	}
}
