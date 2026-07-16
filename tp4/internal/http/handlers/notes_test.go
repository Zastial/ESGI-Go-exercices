package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mira/internal/core"
)

type MockNoteRepository struct {
	CreateFunc func(ctx context.Context, input core.CreateNoteInput) (core.Note, error)
	GetFunc    func(ctx context.Context, id string) (core.Note, error)
	ListFunc   func(ctx context.Context, params core.ListParams) ([]core.Note, int, error)
	UpdateFunc func(ctx context.Context, id string, input core.UpdateNoteInput) (core.Note, error)
	DeleteFunc func(ctx context.Context, id string) error
	SearchFunc func(ctx context.Context, query string, params core.ListParams) ([]core.Note, int, error)
}

func (m *MockNoteRepository) Create(ctx context.Context, input core.CreateNoteInput) (core.Note, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return core.Note{}, nil
}

func (m *MockNoteRepository) Get(ctx context.Context, id string) (core.Note, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, id)
	}
	return core.Note{}, nil
}

func (m *MockNoteRepository) List(ctx context.Context, params core.ListParams) ([]core.Note, int, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, params)
	}
	return []core.Note{}, 0, nil
}

func (m *MockNoteRepository) Update(ctx context.Context, id string, input core.UpdateNoteInput) (core.Note, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, input)
	}
	return core.Note{}, nil
}

func (m *MockNoteRepository) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockNoteRepository) Search(ctx context.Context, query string, params core.ListParams) ([]core.Note, int, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, params)
	}
	return []core.Note{}, 0, nil
}

func (m *MockNoteRepository) ApplyEnrichment(ctx context.Context, id string, result core.EnrichmentResult) error {
	return nil
}

func (m *MockNoteRepository) MarkEnrichmentFailed(ctx context.Context, id string) error {
	return nil
}

type MockJobQueue struct {
	EnqueueFunc func(ctx context.Context, job interface{}) error
}

func (m *MockJobQueue) Enqueue(ctx context.Context, job interface{}) error {
	if m.EnqueueFunc != nil {
		return m.EnqueueFunc(ctx, job)
	}
	return nil
}

func TestCreateNote(t *testing.T) {
	now := time.Now().UTC()
	mockNote := core.Note{
		ID:               "123",
		Title:            "Test Note",
		Content:          "Test content",
		EnrichmentStatus: "pending",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	store := &MockNoteRepository{
		CreateFunc: func(ctx context.Context, input core.CreateNoteInput) (core.Note, error) {
			if input.Title != "Test Note" {
				t.Errorf("expected title 'Test Note', got %q", input.Title)
			}
			return mockNote, nil
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	body := `{"title":"Test Note","content":"Test content"}`
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestCreateNoteInvalidTitle(t *testing.T) {
	store := &MockNoteRepository{}
	handler := NewNotesHandler(store, nil, nil)

	body := `{"title":"","content":"Test content"}`
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader([]byte(body)))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetNote(t *testing.T) {
	now := time.Now().UTC()
	mockNote := core.Note{
		ID:               "123",
		Title:            "Test Note",
		Content:          "Test content",
		EnrichmentStatus: "completed",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	store := &MockNoteRepository{
		GetFunc: func(ctx context.Context, id string) (core.Note, error) {
			if id != "123" {
				return core.Note{}, core.ErrNotFound
			}
			return mockNote, nil
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	req := httptest.NewRequest("GET", "/notes/123", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGetNoteNotFound(t *testing.T) {
	store := &MockNoteRepository{
		GetFunc: func(ctx context.Context, id string) (core.Note, error) {
			return core.Note{}, core.ErrNotFound
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	req := httptest.NewRequest("GET", "/notes/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestListNotes(t *testing.T) {
	now := time.Now().UTC()
	mockNotes := []core.Note{
		{
			ID:               "1",
			Title:            "Note 1",
			Content:          "Content 1",
			EnrichmentStatus: "pending",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		{
			ID:               "2",
			Title:            "Note 2",
			Content:          "Content 2",
			EnrichmentStatus: "completed",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	store := &MockNoteRepository{
		ListFunc: func(ctx context.Context, params core.ListParams) ([]core.Note, int, error) {
			return mockNotes, 2, nil
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	req := httptest.NewRequest("GET", "/notes?limit=20&offset=0", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateNote(t *testing.T) {
	now := time.Now().UTC()
	updatedNote := core.Note{
		ID:               "123",
		Title:            "Updated Title",
		Content:          "Updated content",
		EnrichmentStatus: "pending",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	store := &MockNoteRepository{
		UpdateFunc: func(ctx context.Context, id string, input core.UpdateNoteInput) (core.Note, error) {
			return updatedNote, nil
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	body := `{"title":"Updated Title","content":"Updated content"}`
	req := httptest.NewRequest("PUT", "/notes/123", bytes.NewReader([]byte(body)))
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDeleteNote(t *testing.T) {
	store := &MockNoteRepository{
		DeleteFunc: func(ctx context.Context, id string) error {
			if id == "123" {
				return nil
			}
			return core.ErrNotFound
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	req := httptest.NewRequest("DELETE", "/notes/123", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestSearchNotes(t *testing.T) {
	now := time.Now().UTC()
	mockNotes := []core.Note{
		{
			ID:               "1",
			Title:            "Go Programming",
			Content:          "Go is great",
			EnrichmentStatus: "completed",
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}

	store := &MockNoteRepository{
		SearchFunc: func(ctx context.Context, query string, params core.ListParams) ([]core.Note, int, error) {
			if query != "go" {
				return []core.Note{}, 0, nil
			}
			return mockNotes, 1, nil
		},
	}

	handler := NewNotesHandler(store, nil, nil)

	req := httptest.NewRequest("GET", "/notes/search?q=go&limit=20&offset=0", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestParseListParams(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "default params",
			query:          "",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:           "custom limit",
			query:          "limit=50&offset=10",
			expectedLimit:  50,
			expectedOffset: 10,
		},
		{
			name:           "limit exceeds max",
			query:          "limit=200",
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:           "negative offset",
			query:          "offset=-5",
			expectedLimit:  20,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/notes?"+tt.query, nil)
			params := parseListParams(req)

			if params.Limit != tt.expectedLimit {
				t.Errorf("expected limit %d, got %d", tt.expectedLimit, params.Limit)
			}
			if params.Offset != tt.expectedOffset {
				t.Errorf("expected offset %d, got %d", tt.expectedOffset, params.Offset)
			}
		})
	}
}
