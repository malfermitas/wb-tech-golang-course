package notification

import (
	"calendar_app/internal/entity"
	inmemory "calendar_app/internal/repository/in_memory"
	"calendar_app/internal/repository/interfaces"
	"container/heap"
	"context"
	"log"
	"sync"
	"time"
)

type EventNotificationService struct {
	eventRepo    interfaces.EventRepository
	reminderHeap inmemory.ReminderHeap
	heapMutex    *sync.Mutex
	ctx          context.Context
}

func NewEventNotifier(eventRepo interfaces.EventRepository, ctx context.Context) *EventNotificationService {
	h := make(inmemory.ReminderHeap, 0, 10)
	heap.Init(&h)

	return &EventNotificationService{
		eventRepo:    eventRepo,
		reminderHeap: h,
		heapMutex:    &sync.Mutex{},
		ctx:          ctx,
	}
}

func (e *EventNotificationService) RegisterForNotification(eventID string, reminderTime time.Time) {
	e.heapMutex.Lock()
	heap.Push(&e.reminderHeap, inmemory.ReminderItem{
		EventID:  eventID,
		RemindAt: reminderTime,
	})
	e.heapMutex.Unlock()
}

func (e *EventNotificationService) RemoveReminder(eventID string) {
	e.heapMutex.Lock()
	defer e.heapMutex.Unlock()

	for i := 0; i < e.reminderHeap.Len(); i++ {
		if e.reminderHeap[i].EventID == eventID {
			e.reminderHeap.Swap(i, e.reminderHeap.Len()-1)
			heap.Pop(&e.reminderHeap)
			break
		}
	}
}

func (e *EventNotificationService) UpdateReminder(eventID string, newReminderTime time.Time) {
	e.heapMutex.Lock()
	defer e.heapMutex.Unlock()

	for i := 0; i < e.reminderHeap.Len(); i++ {
		if e.reminderHeap[i].EventID == eventID {
			e.reminderHeap.Swap(i, e.reminderHeap.Len()-1)
			heap.Pop(&e.reminderHeap)
			break
		}
	}

	heap.Push(&e.reminderHeap, inmemory.ReminderItem{
		EventID:  eventID,
		RemindAt: newReminderTime,
	})
}

func (e *EventNotificationService) Start() {
	t := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-t.C:
				e.heapMutex.Lock()
				reminderItem := e.reminderHeap.Peek()
				e.heapMutex.Unlock()

				if reminderItem == nil {
					continue
				}

				event, err := e.eventRepo.FindByID(reminderItem.EventID)
				if err != nil {
					log.Printf("Error getting event: %v", err)
					continue
				}
				if event == nil {
					e.heapMutex.Lock()
					heap.Pop(&e.reminderHeap)
					e.heapMutex.Unlock()
					continue
				}
				if event.Status == entity.EventStatusCancelled || event.Status == entity.EventStatusCompleted {
					e.heapMutex.Lock()
					heap.Pop(&e.reminderHeap)
					e.heapMutex.Unlock()
					continue
				}
				if time.Until(*event.ReminderTime) <= 0 {
					log.Printf("🔔 Напоминание: %s", event.Title)
					event.Status = entity.EventStatusCompleted
					err := e.eventRepo.Update(event)
					if err != nil {
						log.Printf("Error updating event: %v", err)
					}
					e.heapMutex.Lock()
					heap.Pop(&e.reminderHeap)
					e.heapMutex.Unlock()
				}
			case <-e.ctx.Done():
				t.Stop()
				return
			}
		}
	}()
}
