package japanese

import (
	"net/http"

	"github.com/dandydeveloper/dandy-dashboard/internal/httputil"
)

// Handler handles HTTP requests for the Japanese widget.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// WordOfDay handles GET /api/widgets/japanese/word-of-day
func (h *Handler) WordOfDay(w http.ResponseWriter, r *http.Request) {
	entry, err := h.svc.GetWordOfDay()
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "word service unavailable")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, entry)
}
