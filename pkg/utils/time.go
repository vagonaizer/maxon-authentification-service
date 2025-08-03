package utils

import (
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

func IsExpired(expiresAt time.Time) bool {
	return time.Now().UTC().After(expiresAt)
}

func TimeUntilExpiry(expiresAt time.Time) time.Duration {
	return time.Until(expiresAt)
}

func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

func AddHours(t time.Time, hours int) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

func FormatISO8601(t time.Time) string {
	return t.Format(time.RFC3339)
}

func ParseISO8601(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
