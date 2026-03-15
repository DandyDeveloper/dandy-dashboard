package calendar

import (
	"fmt"
	"net/http"
)

// Widget implements widget.Widget for the Google Calendar module.
type Widget struct {
	handler *Handler
}

func New(credentialsJSON, calendarID string) (*Widget, error) {
	svc, err := NewService(credentialsJSON, calendarID)
	if err != nil {
		return nil, fmt.Errorf("calendar widget: %w", err)
	}
	return &Widget{handler: NewHandler(svc)}, nil
}

func (w *Widget) Slug() string { return "calendar" }

func (w *Widget) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /events", w.handler.Events)
}
