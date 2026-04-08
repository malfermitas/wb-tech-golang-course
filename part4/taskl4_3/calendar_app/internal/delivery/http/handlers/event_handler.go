package handlers

import (
	"calendar_app/internal/delivery/dto"
	"calendar_app/internal/delivery/http/responses"
	"calendar_app/internal/service/interfaces"
	"encoding/json"
	"net/http"
)

type EventHandler struct {
	eventUsecase interfaces.EventUsecase
}

func NewEventHandler(eventUsecase interfaces.EventUsecase) *EventHandler {
	return &EventHandler{eventUsecase: eventUsecase}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEventRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	event, err := h.eventUsecase.CreateEvent(req.UserID, req.Event, req.Description, req.Date, req.ReminderDate)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, map[string]string{"result": "event created", "event_id": event.ID.String()})
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var req dto.UpdateEventRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	event, err := h.eventUsecase.UpdateEvent(req.EventID, req.UserID, req.Event, req.Description, req.Date, req.ReminderDate)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, map[string]string{"result": "event updated", "event_id": event.ID.String()})
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	var req dto.DeleteEventRequest
	if err := decodeRequest(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request format")
		return
	}

	if err := h.eventUsecase.DeleteEvent(req.EventID); err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, map[string]string{"result": "event deleted"})
}

func (h *EventHandler) GetEventsForDay(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	date := r.URL.Query().Get("date")

	events, err := h.eventUsecase.GetEventsForDay(userID, date)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, events)
}

func (h *EventHandler) GetEventsForWeek(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	date := r.URL.Query().Get("date")

	events, err := h.eventUsecase.GetEventsForWeek(userID, date)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, events)
}

func (h *EventHandler) GetEventsForMonth(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	date := r.URL.Query().Get("date")

	events, err := h.eventUsecase.GetEventsForMonth(userID, date)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, events)
}

func decodeRequest(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		return json.NewDecoder(r.Body).Decode(v)
	}

	// Для form-data
	if err := r.ParseForm(); err != nil {
		return err
	}

	// Простая реализация для form-data
	if req, ok := v.(*dto.CreateEventRequest); ok {
		req.UserID = r.FormValue("user_id")
		req.Date = r.FormValue("date")
		req.ReminderDate = r.FormValue("reminder_date")
		req.Event = r.FormValue("event")
		req.Description = r.FormValue("description")
	} else if req, ok := v.(*dto.UpdateEventRequest); ok {
		req.EventID = r.FormValue("event_id")
		req.UserID = r.FormValue("user_id")
		req.Date = r.FormValue("date")
		req.ReminderDate = r.FormValue("reminder_date")
		req.Event = r.FormValue("event")
		req.Description = r.FormValue("description")
	} else if req, ok := v.(*dto.DeleteEventRequest); ok {
		req.EventID = r.FormValue("event_id")
	}

	return nil
}

func writeResponse(w http.ResponseWriter, status int, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(responses.Response{Result: result})
}

func writeError(w http.ResponseWriter, status int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(responses.Response{Error: errorMsg})
}
