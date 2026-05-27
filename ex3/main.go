package main

import (
	"fmt"
	"log"

	"go-warmup/note"
)

func main() {
	notes, err := note.LoadFromFile("newNotes.txt")
	if err != nil {
		log.Fatal(err)
	}

	for _, n := range notes {
		fmt.Printf("Title: %s\nContent: %s\nTags: %v\n\n", n.Title, n.Preview(), n.Tags)
	}
}
