package claude

import (
	"log/slog"
	"net/http"
)

// Widget implements widget.Widget for the Claude AI chat module.
type Widget struct {
	handler *Handler
}

func New(apiKey string, log *slog.Logger) *Widget {
	l := log.With("widget", "claude")
	svc := NewService(apiKey, l)
	return &Widget{handler: NewHandler(svc, l)}
}

func (w *Widget) Slug() string { return "claude" }

func (w *Widget) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /chat", w.handler.Chat)
	mux.HandleFunc("DELETE /chat/{session_id}", w.handler.Clear)
}
