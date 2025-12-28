package entity

import (
	"calendar_app/internal/entity/value_objects"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID          `json:"id"`
	UserID      string             `json:"user_id"`
	Date        value_objects.Date `json:"date"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

func NewEvent(userID string, date value_objects.Date, title, description string) (*Event, error) {
	if userID == "" {
		return nil, errors.New("user_id is required")
	}
	if title == "" {
		return nil, errors.New("title is required")
	}

	return &Event{
		ID:          uuid.New(),
		UserID:      userID,
		Date:        date,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}
