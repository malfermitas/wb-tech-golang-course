package event_test

import (
	"calendar_app/internal/entity"
	inmemory "calendar_app/internal/repository/in_memory"
	"calendar_app/internal/repository/interfaces"
	"calendar_app/internal/service/event"
	"calendar_app/internal/service/notification"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockEventRepository struct {
	events map[string]*entity.Event
}

func newMockEventRepository() *mockEventRepository {
	return &mockEventRepository{
		events: make(map[string]*entity.Event),
	}
}

func (m *mockEventRepository) Create(event *entity.Event) error {
	if _, exists := m.events[event.ID.String()]; exists {
		return inmemory.ErrEventConflict
	}
	m.events[event.ID.String()] = event
	return nil
}

func (m *mockEventRepository) Update(event *entity.Event) error {
	if _, exists := m.events[event.ID.String()]; !exists {
		return inmemory.ErrEventNotFound
	}
	m.events[event.ID.String()] = event
	return nil
}

func (m *mockEventRepository) Delete(id string) error {
	if _, exists := m.events[id]; !exists {
		return inmemory.ErrEventNotFound
	}
	delete(m.events, id)
	return nil
}

func (m *mockEventRepository) FindByID(id string) (*entity.Event, error) {
	event, exists := m.events[id]
	if !exists {
		return nil, inmemory.ErrEventNotFound
	}
	return event, nil
}

func (m *mockEventRepository) FindByUserAndDate(userID string, date time.Time) ([]*entity.Event, error) {
	var events []*entity.Event
	for _, e := range m.events {
		if e.UserID == userID && e.Date.Equal(date) {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *mockEventRepository) FindByUserAndDateRange(userID string, start, end time.Time) ([]*entity.Event, error) {
	var events []*entity.Event
	for _, e := range m.events {
		if e.UserID == userID && (e.Date.Equal(start) || e.Date.After(start)) && e.Date.Before(end) {
			events = append(events, e)
		}
	}
	return events, nil
}

func (m *mockEventRepository) FindByUserAndPeriod(userID string, periodType string, date time.Time) ([]*entity.Event, error) {
	switch periodType {
	case "day":
		return m.FindByUserAndDate(userID, date)
	default:
		return nil, nil
	}
}

func (m *mockEventRepository) FindOldEvents(before time.Time) ([]*entity.Event, error) {
	return nil, nil
}

func (m *mockEventRepository) DeleteOldEvents(before time.Time) (int, error) {
	return 0, nil
}

func TestEventService_CreateEvent(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	userID := "user123"
	title := "Meeting"
	description := "Team meeting"
	dateStr := "2024-01-15"
	reminderStr := "2024-01-14"

	eventObj, err := service.CreateEvent(userID, title, description, dateStr, reminderStr)

	assert.NoError(t, err)
	assert.NotNil(t, eventObj)
	assert.Equal(t, userID, eventObj.UserID)
	assert.Equal(t, title, eventObj.Title)
}

func TestEventService_CreateEvent_InvalidDate(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	userID := "user123"
	title := "Meeting"
	description := "Team meeting"
	dateStr := "invalid-date"
	reminderStr := "2024-01-14"

	eventObj, err := service.CreateEvent(userID, title, description, dateStr, reminderStr)

	assert.Error(t, err)
	assert.Nil(t, eventObj)
}

func TestEventService_CreateEvent_MissingUserID(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	userID := ""
	title := "Meeting"
	description := "Team meeting"
	dateStr := "2024-01-15"
	reminderStr := "2024-01-14"

	eventObj, err := service.CreateEvent(userID, title, description, dateStr, reminderStr)

	assert.Error(t, err)
	assert.Nil(t, eventObj)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestEventService_UpdateEvent(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	eventObj, _ := service.CreateEvent("user123", "Old Title", "Old Description", "2024-01-15", "2024-01-14")

	updatedEvent, err := service.UpdateEvent(eventObj.ID.String(), "user123", "New Title", "New Description", "2024-01-20", "2024-01-19")

	assert.NoError(t, err)
	assert.NotNil(t, updatedEvent)
	assert.Equal(t, "New Title", updatedEvent.Title)
	assert.Equal(t, "New Description", updatedEvent.Description)
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	eventID := "nonexistent-id"

	updatedEvent, err := service.UpdateEvent(eventID, "user123", "New Title", "New Description", "2024-01-20", "2024-01-19")

	assert.Error(t, err)
	assert.Nil(t, updatedEvent)
}

func TestEventService_DeleteEvent(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	eventObj, _ := service.CreateEvent("user123", "Test", "Description", "2024-01-15", "2024-01-14")

	err := service.DeleteEvent(eventObj.ID.String())

	assert.NoError(t, err)

	_, err = mockRepo.FindByID(eventObj.ID.String())
	assert.Error(t, err)
}

func TestEventService_GetEventsForDay(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	service.CreateEvent("user123", "Event 1", "Description 1", "2024-01-15", "2024-01-14")
	service.CreateEvent("user123", "Event 2", "Description 2", "2024-01-15", "2024-01-14")

	eventsResult, err := service.GetEventsForDay("user123", "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, eventsResult, 2)
}

func TestEventService_GetEventsForDay_NoEvents(t *testing.T) {
	mockRepo := newMockEventRepository()
	notifier := notification.NewEventNotifier(mockRepo, context.Background())
	service := event.NewService(mockRepo, notifier)

	eventsResult, err := service.GetEventsForDay("user123", "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, eventsResult, 0)
}

func TestEventRepository_Interface(t *testing.T) {
	var _ interfaces.EventRepository = (*inmemory.EventRepository)(nil)
}
