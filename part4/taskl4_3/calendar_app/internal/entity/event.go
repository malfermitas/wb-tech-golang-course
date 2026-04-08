package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type EventStatus string

var (
	EventStatusCreated   EventStatus = "EventStatusCreated"
	EventStatusCompleted EventStatus = "EventStatusCompleted"
	EventStatusCancelled EventStatus = "EventStatusCancelled"
)

type Event struct {
	ID           uuid.UUID   `json:"id"`
	UserID       string      `json:"user_id"`
	Date         time.Time   `json:"date"`
	ReminderTime *time.Time  `json:"reminder_time"`
	Title        string      `json:"title"`
	Description  string      `json:"description,omitempty"`
	Status       EventStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

func NewEvent(userID string, date time.Time, reminderTime *time.Time, title, description string) (*Event, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if title == "" {
		return nil, errors.New("title is required")
	}

	return &Event{
		ID:           uuid.New(),
		UserID:       userID,
		Date:         date,
		ReminderTime: reminderTime,
		Title:        title,
		Description:  description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}
