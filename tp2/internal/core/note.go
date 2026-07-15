package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotFound   = errors.New("note not found")
	ErrValidation = errors.New("validation failed")
)

// Note is the domain entity managed by the API.
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateNoteInput is the payload accepted by POST /api/v1/notes.
type CreateNoteInput struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags,omitempty"`
}

const maxTitleLen = 200

func (in CreateNoteInput) Validate() error {
	if strings.TrimSpace(in.Title) == "" {
		return fmt.Errorf("%w: title is required", ErrValidation)
	}
	if len(in.Title) > maxTitleLen {
		return fmt.Errorf("%w: title must be at most %d characters", ErrValidation, maxTitleLen)
	}
	return nil
}

// UpdateNoteInput is the payload accepted by PATCH /api/v1/notes/{id}.
// Pointer fields distinguish "not provided" from "provided empty".
type UpdateNoteInput struct {
	Title   *string   `json:"title,omitempty"`
	Content *string   `json:"content,omitempty"`
	Tags    *[]string `json:"tags,omitempty"`
}

func (in UpdateNoteInput) Validate() error {
	if in.Title == nil && in.Content == nil && in.Tags == nil {
		return fmt.Errorf("%w: at least one field must be provided", ErrValidation)
	}
	if in.Title != nil {
		if strings.TrimSpace(*in.Title) == "" {
			return fmt.Errorf("%w: title cannot be blank", ErrValidation)
		}
		if len(*in.Title) > maxTitleLen {
			return fmt.Errorf("%w: title must be at most %d characters", ErrValidation, maxTitleLen)
		}
	}
	return nil
}

// ListParams carries pagination parameters for List/Search.
type ListParams struct {
	Limit  int
	Offset int
}

// NoteStore is the persistence port used by the HTTP handlers.
type NoteStore interface {
	Create(ctx context.Context, input CreateNoteInput) (Note, error)
	Get(ctx context.Context, id string) (Note, error)
	List(ctx context.Context, params ListParams) (notes []Note, total int, err error)
	Update(ctx context.Context, id string, input UpdateNoteInput) (Note, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, params ListParams) (notes []Note, total int, err error)
}
