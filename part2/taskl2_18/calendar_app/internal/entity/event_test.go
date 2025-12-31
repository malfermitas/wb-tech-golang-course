package entity_test

import (
	"calendar_app/internal/entity"
	"calendar_app/internal/entity/value_objects"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	date, _ := value_objects.NewDateFromString("2024-01-15")
	userID := "user123"
	title := "Meeting"
	description := "Team meeting"

	event, err := entity.NewEvent(userID, *date, title, description)

	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.NotEqual(t, uuid.Nil, event.ID)
	assert.Equal(t, userID, event.UserID)
	assert.Equal(t, title, event.Title)
	assert.Equal(t, description, event.Description)
	assert.WithinDuration(t, time.Now(), event.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), event.UpdatedAt, time.Second)
}

func TestNewEvent_MissingUserID(t *testing.T) {
	date, _ := value_objects.NewDateFromString("2024-01-15")
	userID := ""
	title := "Meeting"
	description := "Team meeting"

	event, err := entity.NewEvent(userID, *date, title, description)

	assert.Error(t, err)
	assert.Nil(t, event)
	assert.Contains(t, err.Error(), "user_id is required")
}

func TestNewEvent_MissingTitle(t *testing.T) {
	date, _ := value_objects.NewDateFromString("2024-01-15")
	userID := "user123"
	title := ""
	description := "Team meeting"

	event, err := entity.NewEvent(userID, *date, title, description)

	assert.Error(t, err)
	assert.Nil(t, event)
	assert.Contains(t, err.Error(), "title is required")
}
