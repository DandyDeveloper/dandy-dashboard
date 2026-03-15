package calendar

import (
	"net/http"
	"strconv"

	"github.com/dandydeveloper/dandy-dashboard/internal/httputil"
)

// Handler handles HTTP requests for the Calendar widget.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Events handles GET /api/widgets/calendar/events?days=7
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if n, err := strconv.Atoi(d); err == nil && n >= 1 && n <= 90 {
			days = n
		}
	}

	events, err := h.svc.GetUpcomingEvents(days)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "calendar service unavailable")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{"events": events})
}
