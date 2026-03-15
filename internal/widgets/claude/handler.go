package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/dandydeveloper/dandy-dashboard/internal/httputil"
)

const (
	maxRequestBodyBytes = 512 * 1024       // 512 KB
	maxMessageLength    = 32 * 1024        // 32 KB of text per turn
	maxStreamDuration   = 10 * time.Minute // absolute SSE timeout
)

var uuidRE = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) writeEvent(w http.ResponseWriter, flusher http.Flusher, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if flusher != nil {
		flusher.Flush()
	}
}

// Chat handles POST /api/widgets/claude/chat
// Streams the assistant response as Server-Sent Events.
func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Message == "" {
		httputil.WriteError(w, http.StatusBadRequest, "message is required")
		return
	}
	if len(req.Message) > maxMessageLength {
		httputil.WriteError(w, http.StatusRequestEntityTooLarge, "message too long")
		return
	}
	if !uuidRE.MatchString(req.SessionID) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}

	reqID := r.Header.Get("X-Request-Id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	// Enforce an absolute timeout so streams cannot be held open indefinitely.
	ctx, cancel := context.WithTimeout(r.Context(), maxStreamDuration)
	defer cancel()

	textCh, errCh := h.svc.Stream(ctx, req)
	flusher, _ := w.(http.Flusher)

	for {
		select {
		case text, ok := <-textCh:
			if !ok {
				h.writeEvent(w, flusher, "done", "{}")
				return
			}
			payload, _ := json.Marshal(map[string]string{"text": text})
			h.writeEvent(w, flusher, "delta", string(payload))

		case err, ok := <-errCh:
			if ok && err != nil {
				h.log.Error("stream error sent to client",
					"session_id_hash", hashID(req.SessionID),
					"request_id", reqID,
					"error", err,
				)
				h.writeEvent(w, flusher, "error", `{"error":"service unavailable"}`)
				return
			}

		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				h.log.Warn("stream timeout",
					"session_id_hash", hashID(req.SessionID),
					"request_id", reqID,
				)
			} else {
				h.log.Warn("client disconnected",
					"session_id_hash", hashID(req.SessionID),
					"request_id", reqID,
				)
			}
			return
		}
	}
}

// Clear handles DELETE /api/widgets/claude/chat/{session_id}
// Wipes the server-side conversation history for the given session.
func (h *Handler) Clear(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("session_id")
	if !uuidRE.MatchString(sessionID) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid session_id")
		return
	}
	h.svc.ClearSession(sessionID)
	w.WriteHeader(http.StatusNoContent)
}
