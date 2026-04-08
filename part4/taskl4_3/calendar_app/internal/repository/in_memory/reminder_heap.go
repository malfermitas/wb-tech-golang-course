package inmemory

import (
	"time"
)

type ReminderItem struct {
	EventID  string
	RemindAt time.Time
}

type ReminderHeap []ReminderItem

func (h ReminderHeap) Len() int           { return len(h) }
func (h ReminderHeap) Less(i, j int) bool { return h[i].RemindAt.Before(h[j].RemindAt) }
func (h ReminderHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *ReminderHeap) Push(x any) {
	*h = append(*h, x.(ReminderItem))
}

func (h *ReminderHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *ReminderHeap) Peek() *ReminderItem {
	if len(*h) == 0 {
		return nil
	}
	return &(*h)[0]
}
