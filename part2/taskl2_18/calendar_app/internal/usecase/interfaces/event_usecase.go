package interfaces

import (
	"calendar_app/internal/entity"
)

type EventUsecase interface {
	CreateEvent(userID, title, description, dateStr string) (*entity.Event, error)
	UpdateEvent(eventID, userID, title, description, dateStr string) (*entity.Event, error)
	DeleteEvent(eventID string) error
	GetEventsForDay(userID, dateStr string) ([]*entity.Event, error)
	GetEventsForWeek(userID, dateStr string) ([]*entity.Event, error)
	GetEventsForMonth(userID, dateStr string) ([]*entity.Event, error)
}
