package mira

import (
	"time"
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
