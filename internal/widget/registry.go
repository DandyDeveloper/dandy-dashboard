package widget

import "net/http"

// Registry holds all registered widgets and mounts their routes.
type Registry struct {
	widgets []Widget
}

// Register adds a widget to the registry.
func (r *Registry) Register(w Widget) {
	r.widgets = append(r.widgets, w)
}

// Mount attaches all widget routes under /api/widgets/:slug/.
// Each widget receives its own sub-mux via StripPrefix so it registers
// routes relative to its own prefix (e.g. "GET /events" not "GET /api/widgets/calendar/events").
func (r *Registry) Mount(mux *http.ServeMux) {
	for _, w := range r.widgets {
		prefix := "/api/widgets/" + w.Slug()
		sub := http.NewServeMux()
		w.RegisterRoutes(sub)
		mux.Handle(prefix+"/", http.StripPrefix(prefix, sub))
	}
}

// Slugs returns the slug of every registered widget.
func (r *Registry) Slugs() []string {
	slugs := make([]string, len(r.widgets))
	for i, w := range r.widgets {
		slugs[i] = w.Slug()
	}
	return slugs
}
