package notes

import (
	"errors"
	. "esgi/mira/internal/types"
)

var ErrDuplicate = errors.New("Note already exists")

var ErrNotFound = errors.New("Note not found")

type NoteStore interface {
	Save(n *Note) error
	Get(id string) (*Note, error)
	All() []*Note
}
