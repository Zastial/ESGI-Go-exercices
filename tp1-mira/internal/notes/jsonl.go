package notes

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "esgi/mira/internal/search"
	. "esgi/mira/internal/types"

	"github.com/google/uuid"
)

func getNotesPath() (string, error) {
	ensureNotesDir()

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".mira", "notes.jsonl"), nil
}

func ensureNotesDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".mira")
	return os.MkdirAll(dir, 0755)
}

func Save(title, contenu string) error {
	path, err := getNotesPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	n := Note{ID: uuid.NewString(), Title: title, Content: contenu}

	note, getError := Get(n.ID)
	if note != nil && getError == nil {
		return ErrDuplicate
	}

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(n); err != nil {
		return err
	}

	return nil
}

func All() ([]*Note, error) {
	path, err := getNotesPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var notes []*Note

	for decoder.More() {
		var note *Note
		if err := decoder.Decode(&note); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	return notes, nil
}

func Get(id string) (*Note, error) {
	notes, err := All()
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		if note.ID == id {
			return note, nil
		}
	}

	return nil, ErrNotFound
}

func Search(query []string) ([]*Note, error) {
	notes, err := All()
	if err != nil {
		return nil, err
	}

	notesByQuery, err := SearchByQuery(notes, query)
	if err != nil {
		return nil, err
	}

	return notesByQuery, nil
}
