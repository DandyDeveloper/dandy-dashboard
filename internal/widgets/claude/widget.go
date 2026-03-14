package claude

import "github.com/labstack/echo/v4"

// Widget implements widget.Widget for the Claude AI chat module.
type Widget struct {
	handler *Handler
}

func New(apiKey string) *Widget {
	svc := NewService(apiKey)
	return &Widget{handler: NewHandler(svc)}
}

func (w *Widget) Slug() string { return "claude" }

func (w *Widget) RegisterRoutes(g *echo.Group) {
	g.POST("/chat", w.handler.Chat)
}
