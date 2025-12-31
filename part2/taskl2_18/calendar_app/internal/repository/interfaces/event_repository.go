package interfaces

import (
	"calendar_app/internal/entity"
	"calendar_app/internal/entity/value_objects"
)

type EventRepository interface {
	Create(event *entity.Event) error
	Update(event *entity.Event) error
	Delete(id string) error
	FindByID(id string) (*entity.Event, error)
	FindByUserAndDate(userID string, date value_objects.Date) ([]*entity.Event, error)
	FindByUserAndDateRange(userID string, start, end value_objects.Date) ([]*entity.Event, error)
	FindByUserAndPeriod(userID string, periodType string, date value_objects.Date) ([]*entity.Event, error)
}
