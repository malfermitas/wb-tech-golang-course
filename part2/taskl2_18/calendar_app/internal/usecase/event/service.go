package event

import (
	"calendar_app/internal/entity"
	"calendar_app/internal/entity/value_objects"
	"calendar_app/internal/repository/interfaces"
)

type Service struct {
	repo interfaces.EventRepository
}

func NewService(repo interfaces.EventRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateEvent(userID, title, description, dateStr string) (*entity.Event, error) {
	date, err := value_objects.NewDateFromString(dateStr)
	if err != nil {
		return nil, err
	}

	event, err := entity.NewEvent(userID, *date, title, description)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Service) UpdateEvent(eventID, userID, title, description, dateStr string) (*entity.Event, error) {
	event, err := s.repo.FindByID(eventID)
	if err != nil {
		return nil, err
	}

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

	if err := s.repo.Update(event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *Service) DeleteEvent(eventID string) error {
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
