package event

import (
	"calendar_app/internal/entity"
	"calendar_app/internal/entity/value_objects"
	"calendar_app/internal/repository/interfaces"
	"calendar_app/internal/service/notification"
	"time"
)

type Service struct {
	repo                     interfaces.EventRepository
	eventNotificationService *notification.EventNotificationService
}

func NewService(repo interfaces.EventRepository, service *notification.EventNotificationService) *Service {
	return &Service{repo: repo, eventNotificationService: service}
}

func (s *Service) CreateEvent(userID, title, description, dateStr, reminderTimeStr string) (*entity.Event, error) {
	date, err := value_objects.NewDateFromString(dateStr)
	if err != nil {
		return nil, err
	}

	var reminderTime *time.Time
	if reminderTimeStr != "" {
		rt, err := value_objects.NewDateFromString(reminderTimeStr)
		if err != nil {
			return nil, err
		}
		reminderTime = rt
	}

	event, err := entity.NewEvent(userID, *date, reminderTime, title, description)
	if err != nil {
		return nil, err
	}
	event.Status = entity.EventStatusCreated

	if err := s.repo.Create(event); err != nil {
		return nil, err
	}

	if event.ReminderTime != nil {
		s.eventNotificationService.RegisterForNotification(event.ID.String(), *event.ReminderTime)
	}

	return event, nil
}

func (s *Service) UpdateEvent(eventID, userID, title, description, dateStr, reminderTimeStr string) (*entity.Event, error) {
	event, err := s.repo.FindByID(eventID)
	if err != nil {
		return nil, err
	}

	oldReminderTime := event.ReminderTime

	// Обновляем поля
	if userID != "" {
		event.UserID = userID
	}
	if title != "" {
		event.Title = title
	}
	if description != "" {
		event.Description = description
	}
	if dateStr != "" {
		date, err := value_objects.NewDateFromString(dateStr)
		if err != nil {
			return nil, err
		}
		event.Date = *date
	}
	if reminderTimeStr != "" {
		reminderTime, err := value_objects.NewDateFromString(reminderTimeStr)
		if err != nil {
			return nil, err
		}
		event.ReminderTime = reminderTime
	}

	if err := s.repo.Update(event); err != nil {
		return nil, err
	}

	if reminderTimeStr != "" && oldReminderTime != nil {
		s.eventNotificationService.UpdateReminder(eventID, *event.ReminderTime)
	}

	return event, nil
}

func (s *Service) DeleteEvent(eventID string) error {
	s.eventNotificationService.RemoveReminder(eventID)
	return s.repo.Delete(eventID)
}

func (s *Service) GetEventsForDay(userID, dateStr string) ([]*entity.Event, error) {
	date, err := value_objects.NewDateFromString(dateStr)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByUserAndPeriod(userID, "day", *date)
}

func (s *Service) GetEventsForWeek(userID, dateStr string) ([]*entity.Event, error) {
	date, err := value_objects.NewDateFromString(dateStr)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByUserAndPeriod(userID, "week", *date)
}

func (s *Service) GetEventsForMonth(userID, dateStr string) ([]*entity.Event, error) {
	date, err := value_objects.NewDateFromString(dateStr)
	if err != nil {
		return nil, err
	}

	return s.repo.FindByUserAndPeriod(userID, "month", *date)
}
