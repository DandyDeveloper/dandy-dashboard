package calendar

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for the Calendar widget.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Events handles GET /api/widgets/calendar/events?days=7
func (h *Handler) Events(c echo.Context) error {
	days := 7
	if d := c.QueryParam("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil {
			days = n
		}
	}

	events, err := h.svc.GetUpcomingEvents(days)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "failed to fetch events: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
	})
}
