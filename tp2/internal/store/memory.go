package store

import (
	"context"
	"strings"
	"sync"
	"time"

	"mira/internal/core"

	"github.com/google/uuid"
)

// MemoryStore is an in-memory implementation of core.NoteStore, backed by a
// map protected by a mutex. Nothing is persisted across restarts.
type MemoryStore struct {
	mu    sync.RWMutex
	notes map[string]core.Note
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		notes: make(map[string]core.Note),
	}
}

func (s *MemoryStore) Create(ctx context.Context, input core.CreateNoteInput) (core.Note, error) {
	now := time.Now()
	note := core.Note{
		ID:        uuid.NewString(),
		Title:     input.Title,
		Content:   input.Content,
		Tags:      input.Tags,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.notes[note.ID] = note
	return note, nil
}

func (s *MemoryStore) Get(ctx context.Context, id string) (core.Note, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	note, ok := s.notes[id]
	if !ok {
		return core.Note{}, core.ErrNotFound
	}

	return note, nil
}

func (s *MemoryStore) List(ctx context.Context, params core.ListParams) ([]core.Note, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notes := make([]core.Note, 0, len(s.notes))
	for _, note := range s.notes {
		notes = append(notes, note)
	}

	return paginate(notes, params), len(notes), nil
}

func (s *MemoryStore) Update(ctx context.Context, id string, input core.UpdateNoteInput) (core.Note, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	note, ok := s.notes[id]
	if !ok {
		return core.Note{}, core.ErrNotFound
	}

	if input.Title != nil {
		note.Title = *input.Title
	}
	if input.Content != nil {
		note.Content = *input.Content
	}
	if input.Tags != nil {
		note.Tags = *input.Tags
	}
	note.UpdatedAt = time.Now()

	s.notes[id] = note
	return note, nil
}

func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.notes[id]; !ok {
		return core.ErrNotFound
	}

	delete(s.notes, id)
	return nil
}

func (s *MemoryStore) Search(ctx context.Context, query string, params core.ListParams) ([]core.Note, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []core.Note
	for _, note := range s.notes {
		if containsIgnoreCase(note.Title, query) || containsIgnoreCase(note.Content, query) {
			filtered = append(filtered, note)
		}
	}

	return paginate(filtered, params), len(filtered), nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// paginate slices notes according to params, clamping to valid bounds.
func paginate(notes []core.Note, params core.ListParams) []core.Note {
	total := len(notes)

	start := params.Offset
	if start > total {
		start = total
	}
	end := min(start+params.Limit, total)

	return notes[start:end]
}
