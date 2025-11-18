package models

import "time"

// Schedule represents a time window with recurrence patterns
type Schedule struct {
	ID          string        `json:"id"`
	TenantID    string        `json:"tenant_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	TimeWindows []TimeWindow  `json:"time_windows"`
	Active      bool          `json:"active"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// TimeWindow represents a specific time range with optional recurrence
type TimeWindow struct {
	StartDate   time.Time   `json:"start_date"`            // Start date (required)
	EndDate     time.Time   `json:"end_date"`              // End date (required)
	StartTime   string      `json:"start_time,omitempty"`  // Start time in HH:MM format (optional, e.g., "09:00")
	EndTime     string      `json:"end_time,omitempty"`    // End time in HH:MM format (optional, e.g., "17:00")
	Recurrence  *Recurrence `json:"recurrence,omitempty"`  // Optional recurrence pattern
	Timezone    string      `json:"timezone"`              // IANA timezone (e.g., "America/New_York")
}

// Recurrence represents a recurrence pattern for a time window
type Recurrence struct {
	Frequency RecurrenceFrequency `json:"frequency"`           // daily, weekly, monthly, yearly
	Interval  int                 `json:"interval"`            // Repeat every N frequency units (e.g., every 2 weeks)
	DaysOfWeek []time.Weekday     `json:"days_of_week,omitempty"` // For weekly recurrence (0=Sunday, 6=Saturday)
	DayOfMonth int                `json:"day_of_month,omitempty"` // For monthly recurrence (1-31)
	EndDate    *time.Time         `json:"end_date,omitempty"`  // Optional end date for recurrence
}

// RecurrenceFrequency represents how often a schedule recurs
type RecurrenceFrequency string

const (
	RecurrenceDaily   RecurrenceFrequency = "daily"
	RecurrenceWeekly  RecurrenceFrequency = "weekly"
	RecurrenceMonthly RecurrenceFrequency = "monthly"
	RecurrenceYearly  RecurrenceFrequency = "yearly"
)
