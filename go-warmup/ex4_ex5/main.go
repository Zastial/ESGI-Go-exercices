package main

import (
	"errors"
	"sync"

	. "go-warmup/note"
)

var ErrDuplicate = errors.New("note already exists")

var ErrNotFound = errors.New("note not found")

var ErrValidation = errors.New("note validation error")

type NoteStore interface {
	Save(n *Note) error
	Get(id string) (*Note, error)
	All() []*Note
}

type MemoryStore struct {
	notes map[string]*Note
	mu    sync.Mutex
}

func (ms *MemoryStore) Save(n *Note) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.notes == nil {
		ms.notes = make(map[string]*Note)
	}

	if n.Title == "" {
		return ErrValidation
	}

	if exists := ms.notes[n.Title]; exists != nil {
		return ErrDuplicate
	}

	ms.notes[n.Title] = n
	return nil
}

func (ms *MemoryStore) Get(id string) (*Note, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if note := ms.notes[id]; note != nil {
		return note, nil
	}

	return nil, ErrNotFound
}
