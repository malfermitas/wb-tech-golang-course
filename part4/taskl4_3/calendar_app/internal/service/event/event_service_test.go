package event_test

import (
	"calendar_app/internal/entity"
	"calendar_app/internal/entity/value_objects"
	"calendar_app/internal/repository/in_memory"
	"calendar_app/internal/service/event"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock репозитория для тестирования
type MockEventRepository struct {
	mock.Mock
	events map[string]*entity.Event
}

func NewMockEventRepository() *MockEventRepository {
	return &MockEventRepository{
		events: make(map[string]*entity.Event),
	}
}

func (m *MockEventRepository) Create(event *entity.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepository) Update(event *entity.Event) error {
	args := m.Called(event)
	return args.Error(0)
}

func (m *MockEventRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockEventRepository) FindByID(id string) (*entity.Event, error) {
	args := m.Called(id)
	eventObj := args.Get(0)
	if eventObj == nil {
		return nil, args.Error(1)
	}
	return eventObj.(*entity.Event), args.Error(1)
}

func (m *MockEventRepository) FindByUserAndDate(userID string, date value_objects.Date) ([]*entity.Event, error) {
	args := m.Called(userID, date)
	events := args.Get(0)
	if events == nil {
		return nil, args.Error(1)
	}
	return events.([]*entity.Event), args.Error(1)
}

func (m *MockEventRepository) FindByUserAndDateRange(userID string, start, end value_objects.Date) ([]*entity.Event, error) {
	args := m.Called(userID, start, end)
	events := args.Get(0)
	if events == nil {
		return nil, args.Error(1)
	}
	return events.([]*entity.Event), args.Error(1)
}

func (m *MockEventRepository) FindByUserAndPeriod(userID string, periodType string, date value_objects.Date) ([]*entity.Event, error) {
	args := m.Called(userID, periodType, date)
	events := args.Get(0)
	if events == nil {
		return nil, args.Error(1)
	}
	return events.([]*entity.Event), args.Error(1)
}

func TestEventService_CreateEvent(t *testing.T) {
	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	userID := "user123"
	title := "Meeting"
	description := "Team meeting"
	dateStr := "2024-01-15"

	// Настраиваем mock
	mockRepo.On("Create", mock.AnythingOfType("*entity.Event")).Return(nil)

	eventObj, err := service.CreateEvent(userID, title, description, dateStr)

	assert.NoError(t, err)
	assert.NotNil(t, eventObj)
	assert.Equal(t, userID, eventObj.UserID)
	assert.Equal(t, title, eventObj.Title)
	assert.Equal(t, description, eventObj.Description)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_CreateEvent_InvalidDate(t *testing.T) {
	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	userID := "user123"
	title := "Meeting"
	description := "Team meeting"
	dateStr := "invalid-date"

	eventObj, err := service.CreateEvent(userID, title, nil, description, dateStr)

	assert.Error(t, err)
	assert.Nil(t, eventObj)
}

func TestEventService_CreateEvent_MissingUserID(t *testing.T) {
	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	userID := ""
	title := "Meeting"
	description := "Team meeting"
	dateStr := "2024-01-15"

	eventObj, err := service.CreateEvent(userID, title, nil, description, dateStr)

	assert.Error(t, err)
	assert.Nil(t, eventObj)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestEventService_UpdateEvent(t *testing.T) {
	// Создаем тестовое событие
	date, _ := value_objects.NewDateFromString("2024-01-15")
	eventObj, _ := entity.NewEvent("user123", *date, "Old Title", "Old Description")

	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	// Настраиваем mock
	mockRepo.On("FindByID", eventObj.ID.String()).Return(eventObj, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entity.Event")).Return(nil)

	// Обновляем событие
	updatedEvent, err := service.UpdateEvent(eventObj.ID.String(), "user123", "New Title", "New Description", "2024-01-20")

	assert.NoError(t, err)
	assert.NotNil(t, updatedEvent)
	assert.Equal(t, "New Title", updatedEvent.Title)
	assert.Equal(t, "New Description", updatedEvent.Description)

	// Проверяем, что дата обновилась
	newDate, _ := value_objects.NewDateFromString("2024-01-20")
	assert.True(t, updatedEvent.Date.Equal(*newDate))

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_UpdateEvent_NotFound(t *testing.T) {
	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	eventID := uuid.New().String()

	// Настраиваем mock для возврата ошибки
	mockRepo.On("FindByID", eventID).Return(nil, inmemory.ErrEventNotFound)

	updatedEvent, err := service.UpdateEvent(eventID, "user123", "New Title", "New Description", "2024-01-20")

	assert.Error(t, err)
	assert.Nil(t, updatedEvent)
	assert.Equal(t, inmemory.ErrEventNotFound, err)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_DeleteEvent(t *testing.T) {
	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	eventID := uuid.New().String()

	// Настраиваем mock
	mockRepo.On("Delete", eventID).Return(nil)

	err := service.DeleteEvent(eventID)

	assert.NoError(t, err)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_DeleteEvent_NotFound(t *testing.T) {
	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	eventID := uuid.New().String()

	// Настраиваем mock для возврата ошибки
	mockRepo.On("Delete", eventID).Return(inmemory.ErrEventNotFound)

	err := service.DeleteEvent(eventID)

	assert.Error(t, err)
	assert.Equal(t, inmemory.ErrEventNotFound, err)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_GetEventsForDay(t *testing.T) {
	// Создаем тестовые события
	date, _ := value_objects.NewDateFromString("2024-01-15")
	event1, _ := entity.NewEvent("user123", *date, "Event 1", "Description 1")
	event2, _ := entity.NewEvent("user123", *date, "Event 2", "Description 2")
	events := []*entity.Event{event1, event2}

	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	// Настраиваем mock
	mockRepo.On("FindByUserAndPeriod", "user123", "day", *date).Return(events, nil)

	eventsResult, err := service.GetEventsForDay("user123", "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, eventsResult, 2)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_GetEventsForWeek(t *testing.T) {
	// Создаем тестовые события
	date, _ := value_objects.NewDateFromString("2024-01-15")
	event1, _ := entity.NewEvent("user123", *date, "Event 1", "Description 1")
	event2, _ := entity.NewEvent("user123", *date, "Event 2", "Description 2")
	events := []*entity.Event{event1, event2}

	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	// Настраиваем mock
	mockRepo.On("FindByUserAndPeriod", "user123", "week", *date).Return(events, nil)

	eventsResult, err := service.GetEventsForWeek("user123", "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, eventsResult, 2)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}

func TestEventService_GetEventsForMonth(t *testing.T) {
	// Создаем тестовые события
	date, _ := value_objects.NewDateFromString("2024-01-15")
	event1, _ := entity.NewEvent("user123", *date, "Event 1", "Description 1")
	event2, _ := entity.NewEvent("user123", *date, "Event 2", "Description 2")
	events := []*entity.Event{event1, event2}

	mockRepo := NewMockEventRepository()
	service := event.NewService(mockRepo)

	// Настраиваем mock
	mockRepo.On("FindByUserAndPeriod", "user123", "month", *date).Return(events, nil)

	eventsResult, err := service.GetEventsForMonth("user123", "2024-01-15")

	assert.NoError(t, err)
	assert.Len(t, eventsResult, 2)

	// Проверяем, что mock был вызван
	mockRepo.AssertExpectations(t)
}
