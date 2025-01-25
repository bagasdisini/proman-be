package util

import "time"

// StartOfDay calculates the start of the day with an offset in days
func StartOfDay(offset int) time.Time {
	now := time.Now().AddDate(0, 0, offset)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// EndOfDay calculates the end of the day with an offset in days
func EndOfDay(offset int) time.Time {
	return StartOfDay(offset + 1).Add(-time.Second)
}

// StartOfWeek calculates the start of the week with an offset in weeks (starting from Monday)
func StartOfWeek(offset int) time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 { // Handle Sunday as the start of a new week
		weekday = 7
	}
	// Move to Monday of the current week and apply week offset
	startOfWeek := now.AddDate(0, 0, -weekday+1+offset*7)
	return time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())
}

// EndOfWeek calculates the end of the week with an offset in weeks
func EndOfWeek(offset int) time.Time {
	return StartOfWeek(offset + 1).Add(-time.Second)
}

// StartOfMonth calculates the start of the month with an offset in months
func StartOfMonth(offset int) time.Time {
	now := time.Now().AddDate(0, offset, 0)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// EndOfMonth calculates the end of the month with an offset in months
func EndOfMonth(offset int) time.Time {
	return StartOfMonth(offset + 1).Add(-time.Second)
}

// StartOfYear calculates the start of the year with an offset in years
func StartOfYear(offset int) time.Time {
	now := time.Now().AddDate(offset, 0, 0)
	return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
}

// EndOfYear calculates the end of the year with an offset in years
func EndOfYear(offset int) time.Time {
	return StartOfYear(offset + 1).Add(-time.Second)
}
