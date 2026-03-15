package widget

import "net/http"

// Widget is the interface every dashboard module must implement.
// Adding a new widget requires only implementing this interface and
// registering it in main.go — no other files need to change.
type Widget interface {
	// Slug returns the URL-safe identifier used to mount routes,
	// e.g. "claude" → /api/widgets/claude/...
	Slug() string

	// RegisterRoutes attaches the widget's handlers to the provided mux.
	// The mux is scoped to /api/widgets/:slug/ via StripPrefix, so routes
	// should be registered without that prefix (e.g. "GET /events").
	RegisterRoutes(mux *http.ServeMux)
}
