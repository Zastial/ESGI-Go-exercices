package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"mira/internal/client"
	"mira/internal/config"
	"mira/internal/core"
)

func main() {
	config.LoadEnv()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	apiURL := os.Getenv("MIRA_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080/api/v1"
	}
	cli := client.New(apiURL)

	switch os.Args[1] {
	case "add":
		handleAdd(cli, os.Args[2:])
	case "list":
		handleList(cli)
	case "search":
		handleSearch(cli, os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Commande inconnue: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func handleAdd(cli *client.Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mira add <title> <content> [tags]")
		os.Exit(1)
	}

	input := core.CreateNoteInput{Title: args[0], Content: args[1]}
	if len(args) >= 3 && strings.TrimSpace(args[2]) != "" {
		input.Tags = strings.Split(args[2], ",")
	}

	note, err := cli.Create(context.Background(), input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printNote(note)
}

func handleList(cli *client.Client) {
	notes, err := cli.List(context.Background(), core.ListParams{Limit: 20, Offset: 0})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, note := range notes {
		printNote(note)
	}
}

func handleSearch(cli *client.Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: mira search <query>")
		os.Exit(1)
	}

	notes, err := cli.Search(context.Background(), strings.Join(args, " "), core.ListParams{Limit: 20, Offset: 0})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, note := range notes {
		printNote(note)
	}
}

func printNote(note core.Note) {
	fmt.Println("ID:", note.ID)
	fmt.Println("Title:", note.Title)
	fmt.Println("Content:", note.Content)
	fmt.Println("Tags:", strings.Join(note.Tags, ", "))
	fmt.Println("Status:", note.EnrichmentStatus)
	if note.Summary != nil {
		fmt.Println("Summary:", *note.Summary)
	}
	if note.Score != nil {
		fmt.Println("Score:", *note.Score)
	}
	fmt.Println("---")
}

func printUsage() {
	fmt.Println(`mira - CLI
Usage:
  mira add <title> <content> [tags]
  mira list
  mira search <query>
  mira help`)
}
