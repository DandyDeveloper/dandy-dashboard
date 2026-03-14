package japanese

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for the Japanese widget.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// WordOfDay handles GET /api/widgets/japanese/word-of-day
func (h *Handler) WordOfDay(c echo.Context) error {
	entry, err := h.svc.GetWordOfDay()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "failed to fetch word: "+err.Error())
	}
	return c.JSON(http.StatusOK, entry)
}
