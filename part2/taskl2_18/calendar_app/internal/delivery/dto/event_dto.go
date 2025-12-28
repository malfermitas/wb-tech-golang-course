package dto

type CreateEventRequest struct {
	UserID      string `json:"user_id" form:"user_id"`
	Date        string `json:"date" form:"date"`
	Event       string `json:"event" form:"event"`
	Description string `json:"description" form:"description"`
}

type UpdateEventRequest struct {
	EventID     string `json:"event_id" form:"event_id"`
	UserID      string `json:"user_id" form:"user_id"`
	Date        string `json:"date" form:"date"`
	Event       string `json:"event" form:"event"`
	Description string `json:"description" form:"description"`
}

type DeleteEventRequest struct {
	EventID string `json:"event_id" form:"event_id"`
}
