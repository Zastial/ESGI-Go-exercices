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

type Note struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	Tags             []string  `json:"tags,omitempty"`
	Summary          *string   `json:"summary,omitempty"`
	Score            *int      `json:"score,omitempty"`
	EnrichmentStatus string    `json:"enrichment_status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

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

type ListParams struct {
	Limit  int
	Offset int
}

type EnrichmentResult struct {
	Tags      []string
	Summary   string
	Score     int
	Embedding string
}

type NoteRepository interface {
	Create(ctx context.Context, input CreateNoteInput) (Note, error)
	Get(ctx context.Context, id string) (Note, error)
	List(ctx context.Context, params ListParams) (notes []Note, total int, err error)
	Update(ctx context.Context, id string, input UpdateNoteInput) (Note, error)
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, params ListParams) (notes []Note, total int, err error)
	ApplyEnrichment(ctx context.Context, id string, result EnrichmentResult) error
	MarkEnrichmentFailed(ctx context.Context, id string) error
}
