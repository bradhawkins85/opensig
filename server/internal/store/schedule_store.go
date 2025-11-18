package store

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/your-org/opensig/server/internal/models"
)

var (
	ErrScheduleNotFound = errors.New("schedule not found")
)

// ScheduleStore manages schedule data
type ScheduleStore struct {
	mu        sync.RWMutex
	schedules map[string]*models.Schedule
}

// NewScheduleStore creates a new schedule store
func NewScheduleStore() *ScheduleStore {
	return &ScheduleStore{
		schedules: make(map[string]*models.Schedule),
	}
}

// CreateSchedule creates a new schedule
func (s *ScheduleStore) CreateSchedule(schedule *models.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if schedule.ID == "" {
		schedule.ID = uuid.New().String()
	}
	schedule.CreatedAt = time.Now().UTC()
	schedule.UpdatedAt = time.Now().UTC()

	s.schedules[schedule.ID] = schedule
	return nil
}

// GetSchedule retrieves a schedule by ID
func (s *ScheduleStore) GetSchedule(id string) (*models.Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, exists := s.schedules[id]
	if !exists {
		return nil, ErrScheduleNotFound
	}

	return schedule, nil
}

// GetSchedulesByTenantID returns all schedules for a tenant
func (s *ScheduleStore) GetSchedulesByTenantID(tenantID string) ([]*models.Schedule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Schedule
	for _, schedule := range s.schedules {
		if schedule.TenantID == tenantID {
			result = append(result, schedule)
		}
	}

	return result, nil
}

// UpdateSchedule updates an existing schedule
func (s *ScheduleStore) UpdateSchedule(schedule *models.Schedule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.schedules[schedule.ID]; !exists {
		return ErrScheduleNotFound
	}

	schedule.UpdatedAt = time.Now().UTC()
	s.schedules[schedule.ID] = schedule
	return nil
}

// DeleteSchedule deletes a schedule by ID
func (s *ScheduleStore) DeleteSchedule(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.schedules[id]; !exists {
		return ErrScheduleNotFound
	}

	delete(s.schedules, id)
	return nil
}

// IsTimeInSchedule checks if a given time falls within any of the schedule's time windows
func (s *ScheduleStore) IsTimeInSchedule(schedule *models.Schedule, checkTime time.Time) bool {
	if !schedule.Active {
		return false
	}

	for _, window := range schedule.TimeWindows {
		if s.isTimeInWindow(&window, checkTime) {
			return true
		}
	}

	return false
}

// isTimeInWindow checks if a time falls within a specific time window (including recurrence)
func (s *ScheduleStore) isTimeInWindow(window *models.TimeWindow, checkTime time.Time) bool {
	// Parse timezone
	loc, err := time.LoadLocation(window.Timezone)
	if err != nil {
		loc = time.UTC
	}

	// Convert checkTime to window's timezone
	checkTimeInTZ := checkTime.In(loc)
	startDate := window.StartDate.In(loc)
	endDate := window.EndDate.In(loc)

	// Check if checkTime is within the overall date range
	if checkTimeInTZ.Before(startDate) || checkTimeInTZ.After(endDate) {
		return false
	}

	// If there's recurrence, check if it matches
	if window.Recurrence != nil {
		if !s.matchesRecurrence(window.Recurrence, startDate, checkTimeInTZ) {
			return false
		}
	}

	// Check time of day if specified
	if window.StartTime != "" && window.EndTime != "" {
		if !s.isTimeOfDayInRange(window.StartTime, window.EndTime, checkTimeInTZ) {
			return false
		}
	}

	return true
}

// matchesRecurrence checks if a time matches the recurrence pattern
func (s *ScheduleStore) matchesRecurrence(recurrence *models.Recurrence, startDate, checkTime time.Time) bool {
	// Check if we're past the recurrence end date
	if recurrence.EndDate != nil && checkTime.After(*recurrence.EndDate) {
		return false
	}

	switch recurrence.Frequency {
	case models.RecurrenceDaily:
		// Check if the number of days since start is a multiple of interval
		daysSinceStart := int(checkTime.Sub(startDate).Hours() / 24)
		return daysSinceStart%recurrence.Interval == 0

	case models.RecurrenceWeekly:
		// Check if day of week matches
		if len(recurrence.DaysOfWeek) > 0 {
			matched := false
			for _, day := range recurrence.DaysOfWeek {
				if checkTime.Weekday() == day {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
		// Check if the number of weeks since start is a multiple of interval
		weeksSinceStart := int(checkTime.Sub(startDate).Hours() / (24 * 7))
		return weeksSinceStart%recurrence.Interval == 0

	case models.RecurrenceMonthly:
		// Check if day of month matches
		if recurrence.DayOfMonth > 0 && checkTime.Day() != recurrence.DayOfMonth {
			return false
		}
		// Check if the number of months since start is a multiple of interval
		monthsSinceStart := (checkTime.Year()-startDate.Year())*12 + int(checkTime.Month()-startDate.Month())
		return monthsSinceStart%recurrence.Interval == 0

	case models.RecurrenceYearly:
		// Check if month and day match
		if checkTime.Month() != startDate.Month() || checkTime.Day() != startDate.Day() {
			return false
		}
		// Check if the number of years since start is a multiple of interval
		yearsSinceStart := checkTime.Year() - startDate.Year()
		return yearsSinceStart%recurrence.Interval == 0
	}

	return false
}

// isTimeOfDayInRange checks if the time of day falls within the specified range
func (s *ScheduleStore) isTimeOfDayInRange(startTime, endTime string, checkTime time.Time) bool {
	// Parse start and end times (HH:MM format)
	var startHour, startMin, endHour, endMin int
	_, err := time.Parse("15:04", startTime)
	if err == nil {
		t, _ := time.Parse("15:04", startTime)
		startHour, startMin = t.Hour(), t.Minute()
	}
	_, err = time.Parse("15:04", endTime)
	if err == nil {
		t, _ := time.Parse("15:04", endTime)
		endHour, endMin = t.Hour(), t.Minute()
	}

	checkHour, checkMin := checkTime.Hour(), checkTime.Minute()
	checkMinutes := checkHour*60 + checkMin
	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	return checkMinutes >= startMinutes && checkMinutes <= endMinutes
}
