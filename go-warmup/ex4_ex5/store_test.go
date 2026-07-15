package main

import (
	. "go-warmup/note"
	"testing"
)

func TestSave_valid(t *testing.T) {
	ms := &MemoryStore{}
	n := NewNote("t1", "content")

	if err := ms.Save(n); err != nil {
		t.Fatalf("unexpected Save error: %v", err)
	}

	got, err := ms.Get("t1")
	if err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}

	if got.Title != "t1" {
		t.Fatalf("got title %q, want %q", got.Title, "t1")
	}

	if got.Content != "content" {
		t.Fatalf("got content %q, want %q", got.Content, "content")
	}
}

func TestSave_emptyTitle(t *testing.T) {
	ms := &MemoryStore{}
	n := NewNote("", "content")

	if err := ms.Save(n); err != ErrValidation {
		t.Fatalf("unexpected Save error: %v", err)
	}
}

func TestSave_duplicate(t *testing.T) {
	ms := &MemoryStore{}
	n := NewNote("t1", "content")

	if err := ms.Save(n); err != nil {
		t.Fatalf("unexpected Save error: %v", err)
	}

	if err := ms.Save(n); err != ErrDuplicate {
		t.Fatalf("unexpected Save error: %v", err)
	}
}

func TestGet_notFound(t *testing.T) {
	ms := &MemoryStore{}

	if _, err := ms.Get("t1"); err != ErrNotFound {
		t.Fatalf("unexpected Save error: %v", err)
	}
}
