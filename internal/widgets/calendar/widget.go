package calendar

import (
	"fmt"

	"github.com/labstack/echo/v4"
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

func (w *Widget) RegisterRoutes(g *echo.Group) {
	g.GET("/events", w.handler.Events)
}
