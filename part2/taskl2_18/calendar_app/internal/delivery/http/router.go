package http

import (
	"calendar_app/internal/delivery/http/handlers"
	"calendar_app/internal/repository/in_memory"
	"calendar_app/internal/usecase/event"

	"github.com/go-chi/chi/v5"
)

func NewRouter() *chi.Mux {
	r := chi.NewRouter()

	// Инициализация зависимостей
	eventRepo := inmemory.NewEventRepository()
	eventUsecase := event.NewService(eventRepo)
	eventHandler := handlers.NewEventHandler(eventUsecase)

	// Routes
	r.Post("/create_event", eventHandler.CreateEvent)
	r.Post("/update_event", eventHandler.UpdateEvent)
	r.Post("/delete_event", eventHandler.DeleteEvent)
	r.Get("/events_for_day", eventHandler.GetEventsForDay)
	r.Get("/events_for_week", eventHandler.GetEventsForWeek)
	r.Get("/events_for_month", eventHandler.GetEventsForMonth)

	return r
}
