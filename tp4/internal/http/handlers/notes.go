package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"mira/internal/core"
	"mira/internal/http/response"
	"mira/internal/service"
)

type JobQueue interface {
	Enqueue(ctx context.Context, job service.Job) error
}

type NotesHandler struct {
	store  core.NoteRepository
	queue  JobQueue
	logger *slog.Logger
}

func NewNotesHandler(store core.NoteRepository, queue JobQueue, logger *slog.Logger) *NotesHandler {
	return &NotesHandler{store: store, queue: queue, logger: logger}
}

func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	input := core.CreateNoteInput{}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, err)
		return
	}

	note, err := h.store.Create(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}

	if h.queue != nil {
		if err := h.queue.Enqueue(r.Context(), service.Job{NoteID: note.ID}); err != nil {
			h.logger.Warn("job enqueue failed", "note_id", note.ID, "error", err)
		}
	}

	response.JSON(w, http.StatusCreated, note)
}

func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	params := parseListParams(r)
	notes, total, err := h.store.List(r.Context(), params)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSONWithMeta(w, http.StatusOK, notes, response.Meta{Limit: params.Limit, Offset: params.Offset, Total: total})
}

func (h *NotesHandler) Get(w http.ResponseWriter, r *http.Request) {
	note, err := h.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, note)
}

func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	input := core.UpdateNoteInput{}
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, err)
		return
	}
	if err := input.Validate(); err != nil {
		writeError(w, err)
		return
	}

	note, err := h.store.Update(r.Context(), r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}

	if h.queue != nil {
		if err := h.queue.Enqueue(r.Context(), service.Job{NoteID: note.ID}); err != nil {
			h.logger.Warn("job enqueue failed", "note_id", note.ID, "error", err)
		}
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Delete(r.Context(), r.PathValue("id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *NotesHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	params := parseListParams(r)
	notes, total, err := h.store.Search(r.Context(), query, params)
	if err != nil {
		writeError(w, err)
		return
	}
	response.JSONWithMeta(w, http.StatusOK, notes, response.Meta{Limit: params.Limit, Offset: params.Offset, Total: total})
}

func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func parseListParams(r *http.Request) core.ListParams {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	return core.ListParams{Limit: limit, Offset: offset}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, core.ErrNotFound):
		response.Error(w, http.StatusNotFound, "NOT_FOUND", "note not found")
	case errors.Is(err, core.ErrValidation):
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
