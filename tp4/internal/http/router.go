package http

import (
	"log/slog"
	"net/http"
	"time"

	"mira/internal/core"
	"mira/internal/http/handlers"
	"mira/internal/http/middleware"
)

func NewRouter(store core.NoteRepository, queue handlers.JobQueue, logger *slog.Logger) http.Handler {
	h := handlers.NewNotesHandler(store, queue, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/notes", h.Create)
	mux.HandleFunc("GET /api/v1/notes", h.List)
	mux.HandleFunc("GET /api/v1/notes/{id}", h.Get)
	mux.HandleFunc("PATCH /api/v1/notes/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/notes/{id}", h.Delete)
	mux.HandleFunc("GET /api/v1/search", h.Search)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	var handler http.Handler = mux
	handler = middleware.Timeout(5 * time.Second)(handler)
	handler = middleware.Recovery(logger)(handler)
	handler = middleware.Logging(logger)(handler)
	handler = middleware.RequestID(handler)
	return handler
}
