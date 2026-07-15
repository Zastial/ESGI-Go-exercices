package main

import (
	"esgi/mira/internal/notes"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		handleAdd(args)
	case "list":
		handleList()
	case "search":
		handleSearch(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Commande inconnue: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleAdd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: mira add <argument>")
		os.Exit(1)
	}

	if err := notes.Save(args[0], args[1]); err != nil {
		fmt.Println(err)
	}
}

func handleList() {
	notes, err := notes.All()
	if err != nil {
		fmt.Println(err)
	} else {
		for _, note := range notes {
			fmt.Println("ID : " + note.ID)
			fmt.Println("Title : " + note.Title)
			fmt.Println("Content : " + note.Content)
			fmt.Println("------------------------------")
		}
	}
}

func handleSearch(query []string) {
	notes, err := notes.Search(query)
	if err != nil {
		fmt.Println(err)
	} else {
		for _, note := range notes {
			fmt.Println("ID : " + note.ID)
			fmt.Println("Title : " + note.Title)
			fmt.Println("Content : " + note.Content)
			fmt.Println("------------------------------")
		}
	}
}

func printUsage() {
	fmt.Println(`mira - CLI tool
Usage:
  mira add <args>      Add something
  mira list			   List items
  mira search <query>   Search items based on query
  mira help            Show this help`)
}
