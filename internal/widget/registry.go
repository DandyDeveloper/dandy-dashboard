package widget

import "github.com/labstack/echo/v4"

// Registry holds all registered widgets and mounts their routes.
type Registry struct {
	widgets []Widget
}

// Register adds a widget to the registry.
func (r *Registry) Register(w Widget) {
	r.widgets = append(r.widgets, w)
}

// Mount attaches all widget routes to the Echo instance under /api/widgets/:slug/.
func (r *Registry) Mount(e *echo.Echo) {
	for _, w := range r.widgets {
		g := e.Group("/api/widgets/" + w.Slug())
		w.RegisterRoutes(g)
	}
}

// Slugs returns the slug of every registered widget, used by the
// /api/widgets endpoint to advertise available modules to the frontend.
func (r *Registry) Slugs() []string {
	slugs := make([]string, len(r.widgets))
	for i, w := range r.widgets {
		slugs[i] = w.Slug()
	}
	return slugs
}
