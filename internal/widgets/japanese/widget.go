package japanese

import (
	"fmt"
	"net/http"

	"github.com/dandydeveloper/dandy-dashboard/internal/store"
)

// Widget implements widget.Widget for the Japanese word of the day module.
type Widget struct {
	handler *Handler
}

func New(s store.Store, wkToken string) (*Widget, error) {
	svc, err := NewService(s, wkToken)
	if err != nil {
		return nil, fmt.Errorf("japanese widget: %w", err)
	}
	return &Widget{handler: NewHandler(svc)}, nil
}

func (w *Widget) Slug() string { return "japanese" }

func (w *Widget) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /word-of-day", w.handler.WordOfDay)
}
