package utils

import (
	"fmt"
	"time"
)

const DateLayout = "2006-01-02"

// ParseFutureDate parses a YYYY-MM-DD string and enforces that the resulting
// date is strictly after the current moment.
func ParseFutureDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, fmt.Errorf("due_date is required")
	}
	t, err := time.Parse(DateLayout, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("due_date must be in YYYY-MM-DD format")
	}

	// Compare at the calendar-day level so results are consistent regardless of
	// the local timezone.  A task whose due date is today or earlier is rejected.
	nowYear, nowMonth, nowDay := time.Now().Date()
	today := time.Date(nowYear, nowMonth, nowDay, 0, 0, 0, 0, time.UTC)
	dueDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	if !dueDay.After(today) {
		return time.Time{}, fmt.Errorf("due_date must be in the future")
	}
	return t, nil
}
