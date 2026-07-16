package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"mira-mcp/internal/mira"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	apiURL := os.Getenv("MIRA_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080/api/v1"
	}

	client := mira.NewClient(apiURL)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mira",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_notes",
		Description: "Search notes by query using hybrid full-text and vector search",
	}, handleSearchNotes(logger, client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_note",
		Description: "Retrieve a complete note by ID including content, tags, summary, and enrichment status",
	}, handleGetNote(logger, client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_note",
		Description: "Create a new note with title and content, optionally with tags",
	}, handleAddNote(logger, client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_recent_notes",
		Description: "List the most recently created notes",
	}, handleListRecentNotes(logger, client))

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
