package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"mira/internal/core"
	"mira/internal/http/response"
)

// NotesHandler wires the /api/v1/notes and /api/v1/search HTTP routes to a core.NoteStore.
type NotesHandler struct {
	store  core.NoteStore
	logger *slog.Logger
}

func NewNotesHandler(store core.NoteStore, logger *slog.Logger) *NotesHandler {
	return &NotesHandler{store: store, logger: logger}
}

// Create handles POST /api/v1/notes.
//
// @Summary      Create a note
// @Description  Creates a new note
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        note  body      core.CreateNoteInput  true  "Note to create"
// @Success      201   {object}  core.Note
// @Failure      400   {object}  response.ErrorBody
// @Failure      500   {object}  response.ErrorBody
// @Router       /notes [post]
func (h *NotesHandler) Create(w http.ResponseWriter, r *http.Request) {
	noteToCreate := core.CreateNoteInput{}
	if err := decodeJSON(r, &noteToCreate); err != nil {
		writeError(w, err)
		return
	}

	if err := noteToCreate.Validate(); err != nil {
		writeError(w, err)
		return
	}

	note, err := h.store.Create(r.Context(), noteToCreate)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, note)
}

// List handles GET /api/v1/notes.
//
// @Summary      List notes
// @Description  Lists notes, paginated
// @Tags         notes
// @Produce      json
// @Param        limit   query     int  false  "Max number of notes to return (default 20, max 100)"
// @Param        offset  query     int  false  "Number of notes to skip (default 0)"
// @Success      200     {array}   core.Note
// @Failure      500     {object}  response.ErrorBody
// @Router       /notes [get]
func (h *NotesHandler) List(w http.ResponseWriter, r *http.Request) {
	params := parseListParams(r)

	notes, total, err := h.store.List(r.Context(), params)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSONWithMeta(w, http.StatusOK, notes, response.Meta{Limit: params.Limit, Offset: params.Offset, Total: total})
}

// Get handles GET /api/v1/notes/{id}.
//
// @Summary      Get a note
// @Description  Retrieves a single note by id
// @Tags         notes
// @Produce      json
// @Param        id   path      string  true  "Note ID"
// @Success      200  {object}  core.Note
// @Failure      404  {object}  response.ErrorBody
// @Router       /notes/{id} [get]
func (h *NotesHandler) Get(w http.ResponseWriter, r *http.Request) {
	note, err := h.store.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, note)
}

// Update handles PATCH /api/v1/notes/{id}.
//
// @Summary      Update a note
// @Description  Partially updates a note (only provided fields are applied)
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        id    path      string                true  "Note ID"
// @Param        note  body      core.UpdateNoteInput  true  "Fields to update"
// @Success      200   {object}  core.Note
// @Failure      400   {object}  response.ErrorBody
// @Failure      404   {object}  response.ErrorBody
// @Router       /notes/{id} [patch]
func (h *NotesHandler) Update(w http.ResponseWriter, r *http.Request) {
	noteToUpdate := core.UpdateNoteInput{}
	if err := decodeJSON(r, &noteToUpdate); err != nil {
		writeError(w, err)
		return
	}

	if err := noteToUpdate.Validate(); err != nil {
		writeError(w, err)
		return
	}

	note, err := h.store.Update(r.Context(), r.PathValue("id"), noteToUpdate)
	if err != nil {
		writeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, note)
}

// Delete handles DELETE /api/v1/notes/{id}.
//
// @Summary      Delete a note
// @Description  Deletes a note by id
// @Tags         notes
// @Param        id  path  string  true  "Note ID"
// @Success      204
// @Failure      404  {object}  response.ErrorBody
// @Router       /notes/{id} [delete]
func (h *NotesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.store.Delete(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Search handles GET /api/v1/search?q=...
//
// @Summary      Search notes
// @Description  Case-insensitive text search over the title and content of notes, paginated
// @Tags         notes
// @Produce      json
// @Param        q       query     string  true   "Search query"
// @Param        limit   query     int     false  "Max number of notes to return (default 20, max 100)"
// @Param        offset  query     int     false  "Number of notes to skip (default 0)"
// @Success      200     {array}   core.Note
// @Failure      500     {object}  response.ErrorBody
// @Router       /search [get]
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

// --- helpers (shared plumbing, already implemented) ---

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
