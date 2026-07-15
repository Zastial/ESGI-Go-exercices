package search

import (
	. "esgi/mira/internal/types"
	"slices"
	"strings"
)

func SearchByQuery(notes []*Note, queries []string) ([]*Note, error) {
	var filteredNotes []*Note
	for _, note := range notes {
		for _, query := range queries {
			if strings.Contains(note.Title, query) || strings.Contains(note.Content, query) {
				if slices.Contains(filteredNotes, note) {
					continue
				}
				filteredNotes = append(filteredNotes, note)
			}
		}
	}

	return filteredNotes, nil
}
