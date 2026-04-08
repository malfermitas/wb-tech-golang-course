package inmemory

import (
	"calendar_app/internal/entity"
	"calendar_app/internal/entity/value_objects"
	"calendar_app/internal/repository/interfaces"
	"errors"
	"sync"
	"time"
)

var (
	ErrEventNotFound = errors.New("event not found")
	ErrEventConflict = errors.New("event already exists")
)

type EventRepository struct {
	events map[string]*entity.Event
	mu     sync.RWMutex
}

func NewEventRepository() interfaces.EventRepository {
	return &EventRepository{
		events: make(map[string]*entity.Event),
	}
}

func (r *EventRepository) Create(event *entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем уникальность по ID
	if _, exists := r.events[event.ID.String()]; exists {
		return ErrEventConflict
	}

	r.events[event.ID.String()] = event
	return nil
}

func (r *EventRepository) Update(event *entity.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.events[event.ID.String()]; !exists {
		return ErrEventNotFound
	}

	event.UpdatedAt = time.Now()
	r.events[event.ID.String()] = event
	return nil
}

func (r *EventRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.events[id]; !exists {
		return ErrEventNotFound
	}

	delete(r.events, id)
	return nil
}

func (r *EventRepository) FindByID(id string) (*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	event, exists := r.events[id]
	if !exists {
		return nil, ErrEventNotFound
	}

	return event, nil
}

func (r *EventRepository) FindByUserAndDate(userID string, date time.Time) ([]*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var events []*entity.Event
	for _, event := range r.events {
		if event.UserID == userID && event.Date.Equal(date) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (r *EventRepository) FindByUserAndDateRange(userID string, start, end time.Time) ([]*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var events []*entity.Event
	for _, event := range r.events {
		if event.UserID == userID &&
			(event.Date.Equal(start) || event.Date.After(start)) &&
			event.Date.Before(end) {
			events = append(events, event)
		}
	}

	return events, nil
}

func (r *EventRepository) FindByUserAndPeriod(userID string, periodType string, date time.Time) ([]*entity.Event, error) {
	switch periodType {
	case "day":
		return r.FindByUserAndDate(userID, date)
	case "week":
		start := date
		end := start.AddDate(0, 0, 7)
		startVO, _ := value_objects.NewDate(start.Year(), start.Month(), start.Day())
		endVO, _ := value_objects.NewDate(end.Year(), end.Month(), end.Day())
		return r.FindByUserAndDateRange(userID, *startVO, *endVO)
	case "month":
		start := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)
		startVO, _ := value_objects.NewDate(start.Year(), start.Month(), start.Day())
		endVO, _ := value_objects.NewDate(end.Year(), end.Month(), end.Day())
		return r.FindByUserAndDateRange(userID, *startVO, *endVO)
	default:
		return nil, errors.New("invalid period type")
	}
}

func (r *EventRepository) FindOldEvents(before time.Time) ([]*entity.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var events []*entity.Event
	for _, event := range r.events {
		if event.Date.Before(before) {
			events = append(events, event)
		}
	}
	return events, nil
}

func (r *EventRepository) DeleteOldEvents(before time.Time) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deleted := 0
	for id, event := range r.events {
		if event.Date.Before(before) {
			delete(r.events, id)
			deleted++
		}
	}
	return deleted, nil
}
