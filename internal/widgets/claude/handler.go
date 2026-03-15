package claude

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handler) writeEvent(w http.ResponseWriter, flusher http.Flusher, event, data string) error {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if flusher != nil {
		flusher.Flush()
	}
	return nil
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Chat handles POST /api/widgets/claude/chat
// Streams the assistant response as Server-Sent Events.
func (h *Handler) Chat(c echo.Context) error {
	var req ChatRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if req.Message == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "message is required")
	}
	if req.SessionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "session_id is required")
	}

	w := c.Response().Writer
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(http.StatusOK)

	textCh, errCh := h.svc.Stream(c.Request().Context(), req)
	flusher, _ := w.(http.Flusher)

	for {
		select {
		case text, ok := <-textCh:
			if !ok {
				return h.writeEvent(w, flusher, "done", "{}")
			}
			payload, _ := json.Marshal(map[string]string{"text": text})
			h.writeEvent(w, flusher, "delta", string(payload))
		case err, ok := <-errCh:
			if ok && err != nil {
				payload, _ := json.Marshal(map[string]string{"error": err.Error()})
				return h.writeEvent(w, flusher, "error", string(payload))
			}
		case <-c.Request().Context().Done():
			return nil
		}
	}
}

// Clear handles DELETE /api/widgets/claude/chat/:session_id
// Wipes the server-side conversation history for the given session.
func (h *Handler) Clear(c echo.Context) error {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "session_id is required")
	}
	h.svc.ClearSession(sessionID)
	return c.NoContent(http.StatusNoContent)
}
