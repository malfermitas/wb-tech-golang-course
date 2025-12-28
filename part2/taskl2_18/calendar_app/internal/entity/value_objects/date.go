package value_objects

import (
	"errors"
	"time"
)

type Date struct {
	value time.Time
}

func NewDate(year int, month time.Month, day int) (*Date, error) {
	date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || date.Month() != month || date.Day() != day {
		return nil, errors.New("invalid date")
	}
	return &Date{value: date}, nil
}

func NewDateFromString(dateStr string) (*Date, error) {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, errors.New("invalid date format, expected YYYY-MM-DD")
	}
	return &Date{value: date}, nil
}

func (d Date) Value() time.Time {
	return d.value
}

func (d Date) String() string {
	return d.value.Format("2006-01-02")
}

func (d Date) Year() int {
	return d.value.Year()
}

func (d Date) Month() time.Month {
	return d.value.Month()
}

func (d Date) Day() int {
	return d.value.Day()
}

func (d Date) Equal(other Date) bool {
	return d.value.Equal(other.value)
}

func (d Date) Before(other Date) bool {
	return d.value.Before(other.value)
}

func (d Date) After(other Date) bool {
	return d.value.After(other.value)
}
