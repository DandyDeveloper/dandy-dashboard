package widget

import "github.com/labstack/echo/v4"

// Widget is the interface every dashboard module must implement.
// Adding a new widget requires only implementing this interface and
// registering it in main.go — no other files need to change.
type Widget interface {
	// Slug returns the URL-safe identifier used to mount routes,
	// e.g. "claude" → /api/widgets/claude/...
	Slug() string

	// RegisterRoutes attaches the widget's handlers to the provided Echo group.
	// The group is already scoped to /api/widgets/:slug/.
	RegisterRoutes(g *echo.Group)
}
