package value_objects_test

import (
	"calendar_app/internal/entity/value_objects"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	year, month, day := 2024, time.January, 15

	date, err := value_objects.NewDate(year, month, day)

	assert.NoError(t, err)
	assert.Equal(t, year, date.Year())
	assert.Equal(t, month, date.Month())
	assert.Equal(t, day, date.Day())
}

func TestNewDateFromString(t *testing.T) {
	dateStr := "2024-01-15"

	date, err := value_objects.NewDateFromString(dateStr)

	assert.NoError(t, err)
	assert.Equal(t, 2024, date.Year())
	assert.Equal(t, time.January, date.Month())
	assert.Equal(t, 15, date.Day())
}

func TestNewDateFromString_InvalidFormat(t *testing.T) {
	dateStr := "invalid-date"

	date, err := value_objects.NewDateFromString(dateStr)

	assert.Error(t, err)
	assert.Nil(t, date)
}

func TestDate_Equal(t *testing.T) {
	date1, _ := value_objects.NewDateFromString("2024-01-15")
	date2, _ := value_objects.NewDateFromString("2024-01-15")
	date3, _ := value_objects.NewDateFromString("2024-01-16")

	assert.True(t, date1.Equal(*date2))
	assert.False(t, date1.Equal(*date3))
}

func TestDate_After(t *testing.T) {
	date1, _ := value_objects.NewDateFromString("2024-01-16")
	date2, _ := value_objects.NewDateFromString("2024-01-15")

	assert.True(t, date1.After(*date2))
	assert.False(t, date2.After(*date1))
}

func TestDate_Before(t *testing.T) {
	date1, _ := value_objects.NewDateFromString("2024-01-15")
	date2, _ := value_objects.NewDateFromString("2024-01-16")

	assert.True(t, date1.Before(*date2))
	assert.False(t, date2.Before(*date1))
}
