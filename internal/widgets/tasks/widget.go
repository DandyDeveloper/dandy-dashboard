package tasks

import (
	"net/http"

	"github.com/dandydeveloper/dandy-dashboard/internal/store"
)

type Widget struct {
	handler *Handler
}

func New(s store.Store) *Widget {
	return &Widget{handler: NewHandler(NewService(s))}
}

func (w *Widget) Slug() string { return "tasks" }

func (w *Widget) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /items", w.handler.List)
	mux.HandleFunc("POST /items", w.handler.Create)
	mux.HandleFunc("PATCH /items/{id}", w.handler.Update)
	mux.HandleFunc("DELETE /items/{id}", w.handler.Delete)
}
