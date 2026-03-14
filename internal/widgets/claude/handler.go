package claude

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Chat handles POST /api/widgets/claude/chat
// It streams the assistant response as Server-Sent Events.
func (h *Handler) Chat(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Message == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}

	w := c.Response().Writer
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)

	textCh, errCh := h.svc.Stream(c.Request().Context(), req)

	flusher, canFlush := w.(http.Flusher)

	for {
		select {
		case text, ok := <-textCh:
			if !ok {
				// Stream finished — send done event.
				fmt.Fprintf(w, "event: done\ndata: {}\n\n")
				if canFlush {
					flusher.Flush()
				}
				return nil
			}
			payload, _ := json.Marshal(map[string]string{"text": text})
			fmt.Fprintf(w, "event: delta\ndata: %s\n\n", payload)
			if canFlush {
				flusher.Flush()
			}
		case err, ok := <-errCh:
			if ok && err != nil {
				payload, _ := json.Marshal(map[string]string{"error": err.Error()})
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", payload)
				if canFlush {
					flusher.Flush()
				}
				return nil
			}
		case <-c.Request().Context().Done():
			return nil
		}
	}
}
