package http

import (
	"calendar_app/internal/delivery/http/handlers"
	"calendar_app/internal/service/interfaces"

	"github.com/go-chi/chi/v5"
)

func NewRouter(eventService interfaces.EventUsecase) *chi.Mux {
	r := chi.NewRouter()

	eventHandler := handlers.NewEventHandler(eventService)

	r.Post("/create_event", eventHandler.CreateEvent)
	r.Post("/update_event", eventHandler.UpdateEvent)
	r.Post("/delete_event", eventHandler.DeleteEvent)
	r.Get("/events_for_day", eventHandler.GetEventsForDay)
	r.Get("/events_for_week", eventHandler.GetEventsForWeek)
	r.Get("/events_for_month", eventHandler.GetEventsForMonth)

	return r
}
