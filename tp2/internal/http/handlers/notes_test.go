package handlers_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mira/internal/http/handlers"
	"mira/internal/http/response"
	"mira/internal/store"
)

func newTestHandler() *handlers.NotesHandler {
	memStore := store.NewMemoryStore()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return handlers.NewNotesHandler(memStore, logger)
}

func TestNotesHandler_Create_Success(t *testing.T) {
	h := newTestHandler()
	body := strings.NewReader(`{"title":"Hello","content":"World"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var env response.Envelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	data, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data to be an object, got %T", env.Data)
	}
	if id, _ := data["id"].(string); id == "" {
		t.Fatalf("expected a non-empty note id, got %q", data["id"])
	}
}

func TestNotesHandler_Create_InvalidPayload(t *testing.T) {
	h := newTestHandler()
	body := strings.NewReader(`{"title":""}`) // missing/invalid required fields
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notes", body)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestNotesHandler_Get_NotFound(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notes/does-not-exist", nil)
	req.SetPathValue("id", "does-not-exist")
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}
