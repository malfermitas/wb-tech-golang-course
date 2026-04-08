package value_objects

import (
	"errors"
	"time"
)

func NewDate(year int, month time.Month, day int) (*time.Time, error) {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || date.Month() != month || date.Day() != day {
		return nil, errors.New("invalid date")
	}
	return &date, nil
}

func NewDateFromString(dateStr string) (*time.Time, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	return &date, nil
}
