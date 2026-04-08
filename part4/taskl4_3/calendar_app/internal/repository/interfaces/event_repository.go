package interfaces

import (
	"calendar_app/internal/entity"
	"time"
)

type EventRepository interface {
	Create(event *entity.Event) error
	Update(event *entity.Event) error
	Delete(id string) error
	FindByID(id string) (*entity.Event, error)
	FindByUserAndDate(userID string, date time.Time) ([]*entity.Event, error)
	FindByUserAndDateRange(userID string, start, end time.Time) ([]*entity.Event, error)
	FindByUserAndPeriod(userID string, periodType string, date time.Time) ([]*entity.Event, error)
	FindOldEvents(before time.Time) ([]*entity.Event, error)
	DeleteOldEvents(before time.Time) (int, error)
}
