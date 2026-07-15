// Package http assembles the routes and middleware chain for the API.
package http

import (
	"log/slog"
	"net/http"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"

	"mira/internal/core"
	"mira/internal/http/handlers"
	"mira/internal/http/middleware"
)

// NewRouter builds the full HTTP handler: routes wrapped in the middleware chain.
func NewRouter(store core.NoteStore, logger *slog.Logger) http.Handler {
	h := handlers.NewNotesHandler(store, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/notes", h.Create)
	mux.HandleFunc("GET /api/v1/notes", h.List)
	mux.HandleFunc("GET /api/v1/notes/{id}", h.Get)
	mux.HandleFunc("PATCH /api/v1/notes/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/notes/{id}", h.Delete)
	mux.HandleFunc("GET /api/v1/search", h.Search)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	var handler http.Handler = mux
	handler = middleware.Timeout(5 * time.Second)(handler)
	handler = middleware.Recovery(logger)(handler)
	handler = middleware.Logging(logger)(handler)
	handler = middleware.RequestID(handler)
	return handler
}
