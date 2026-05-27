// Package utilisé par l'exercice 3, 4 et 5
package note

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

type Note struct {
	Title   string
	Content string
	Tags    []string
}

func NewNote(title, content string) *Note {
	return &Note{Title: title, Content: content, Tags: []string{}}
}

func (note Note) Preview() string {
	if len(note.Content) < 80 {
		return note.Content
	}
	return note.Content[:80]
}

func (note *Note) AddTag(tag string) {
	if slices.Contains(note.Tags, tag) {
		return
	}
	note.Tags = append(note.Tags, tag)
}

func LoadFromFile(path string) ([]*Note, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file not found")
	}
	defer file.Close()

	newNotes := []*Note{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		noteLine := strings.Split(scanner.Text(), `,`)

		newNote := NewNote(noteLine[0], noteLine[1])
		for i := 2; i < len(noteLine); i++ {
			newNote.AddTag(noteLine[i])
		}

		newNotes = append(newNotes, newNote)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return newNotes, nil
}
