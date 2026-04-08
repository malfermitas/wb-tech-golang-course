package inmemory

import (
	"calendar_app/internal/entity"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEventRepository_Create(t *testing.T) {
	repo := NewEventRepository()

	event := &entity.Event{
		ID:        uuid.New(),
		UserID:    "user1",
		Date:      time.Now(),
		Title:     "Test Event",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(event)
	assert.NoError(t, err)

	found, err := repo.FindByID(event.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, event.ID, found.ID)
}

func TestEventRepository_Create_Duplicate(t *testing.T) {
	repo := NewEventRepository()

	event := &entity.Event{
		ID:        uuid.New(),
		UserID:    "user1",
		Date:      time.Now(),
		Title:     "Test Event",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(event)
	assert.NoError(t, err)

	err = repo.Create(event)
	assert.Error(t, err)
	assert.Equal(t, ErrEventConflict, err)
}

func TestEventRepository_Update(t *testing.T) {
	repo := NewEventRepository()

	event := &entity.Event{
		ID:        uuid.New(),
		UserID:    "user1",
		Date:      time.Now(),
		Title:     "Test Event",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(event)

	event.Title = "Updated Event"
	err := repo.Update(event)
	assert.NoError(t, err)

	found, _ := repo.FindByID(event.ID.String())
	assert.Equal(t, "Updated Event", found.Title)
}

func TestEventRepository_Delete(t *testing.T) {
	repo := NewEventRepository()

	event := &entity.Event{
		ID:        uuid.New(),
		UserID:    "user1",
		Date:      time.Now(),
		Title:     "Test Event",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.Create(event)

	err := repo.Delete(event.ID.String())
	assert.NoError(t, err)

	_, err = repo.FindByID(event.ID.String())
	assert.Error(t, err)
}

func TestEventRepository_DeleteNotFound(t *testing.T) {
	repo := NewEventRepository()

	err := repo.Delete("nonexistent-id")
	assert.Error(t, err)
	assert.Equal(t, ErrEventNotFound, err)
}

func TestEventRepository_FindByUserAndDate(t *testing.T) {
	repo := NewEventRepository()

	date := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)

	event1 := &entity.Event{
		ID:     uuid.New(),
		UserID: "user1",
		Date:   date,
		Title:  "Event 1",
	}
	event2 := &entity.Event{
		ID:     uuid.New(),
		UserID: "user1",
		Date:   date,
		Title:  "Event 2",
	}
	event3 := &entity.Event{
		ID:     uuid.New(),
		UserID: "user2",
		Date:   date,
		Title:  "Event 3",
	}

	repo.Create(event1)
	repo.Create(event2)
	repo.Create(event3)

	events, err := repo.FindByUserAndDate("user1", date)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestEventRepository_DeleteOldEvents(t *testing.T) {
	repo := NewEventRepository()

	oldDate := time.Now().Add(-60 * 24 * time.Hour) // 60 days ago
	newDate := time.Now()

	event1 := &entity.Event{
		ID:     uuid.New(),
		UserID: "user1",
		Date:   oldDate,
		Title:  "Old Event",
	}
	event2 := &entity.Event{
		ID:     uuid.New(),
		UserID: "user1",
		Date:   newDate,
		Title:  "New Event",
	}

	repo.Create(event1)
	repo.Create(event2)

	cutoff := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	deleted, err := repo.DeleteOldEvents(cutoff)
	assert.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Old event should be deleted
	_, err = repo.FindByID(event1.ID.String())
	assert.Error(t, err)

	// New event should still exist
	found, err := repo.FindByID(event2.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "New Event", found.Title)
}
