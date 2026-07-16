package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"mira-mcp/internal/mira"
)

type SearchNotesArgs struct {
	Query string `json:"query" jsonschema:"Search query (required)"`
	Limit int    `json:"limit,omitempty" jsonschema:"Maximum results, defaults to 10"`
}

func handleSearchNotes(logger *slog.Logger, client *mira.Client) func(ctx context.Context, req *mcp.CallToolRequest, args SearchNotesArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args SearchNotesArgs) (*mcp.CallToolResult, any, error) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if strings.TrimSpace(args.Query) == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "query is required"},
				},
			}, nil, nil
		}

		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		notes, err := client.Search(ctx, args.Query, limit)
		if err != nil {
			logger.Warn("search failed", "error", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("search failed: %v", err)},
				},
			}, nil, nil
		}

		var text strings.Builder
		if len(notes) == 0 {
			text.WriteString("No notes found")
		} else {
			fmt.Fprintf(&text, "Found %d note(s):\n\n", len(notes))
			for i, note := range notes {
				fmt.Fprintf(&text, "[%d] %s (ID: %s)\n", i+1, note.Title, note.ID)
				if note.Summary != nil {
					fmt.Fprintf(&text, "    Summary: %s\n", *note.Summary)
				}
				if len(note.Tags) > 0 {
					fmt.Fprintf(&text, "    Tags: %s\n", strings.Join(note.Tags, ", "))
				}
				fmt.Fprintf(&text, "    Status: %s\n", note.EnrichmentStatus)
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text.String()},
			},
		}, nil, nil
	}
}

type GetNoteArgs struct {
	ID string `json:"id" jsonschema:"Note ID (required)"`
}

func handleGetNote(logger *slog.Logger, client *mira.Client) func(ctx context.Context, req *mcp.CallToolRequest, args GetNoteArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args GetNoteArgs) (*mcp.CallToolResult, any, error) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if strings.TrimSpace(args.ID) == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "id is required"},
				},
			}, nil, nil
		}

		note, err := client.Get(ctx, args.ID)
		if err != nil {
			logger.Warn("get note failed", "id", args.ID, "error", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("get note failed: %v", err)},
				},
			}, nil, nil
		}

		var text strings.Builder
		fmt.Fprintf(&text, "Title: %s\n", note.Title)
		fmt.Fprintf(&text, "ID: %s\n", note.ID)
		fmt.Fprintf(&text, "\nContent:\n%s\n", note.Content)

		if len(note.Tags) > 0 {
			fmt.Fprintf(&text, "\nTags: %s\n", strings.Join(note.Tags, ", "))
		}
		if note.Summary != nil {
			fmt.Fprintf(&text, "\nSummary: %s\n", *note.Summary)
		}
		if note.Score != nil {
			fmt.Fprintf(&text, "\nScore: %d\n", *note.Score)
		}
		fmt.Fprintf(&text, "\nStatus: %s\n", note.EnrichmentStatus)
		fmt.Fprintf(&text, "Created: %s\n", note.CreatedAt.Format(time.RFC3339))
		fmt.Fprintf(&text, "Updated: %s\n", note.UpdatedAt.Format(time.RFC3339))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text.String()},
			},
		}, nil, nil
	}
}

type AddNoteArgs struct {
	Title   string   `json:"title" jsonschema:"Note title (required)"`
	Content string   `json:"content" jsonschema:"Note content (required)"`
	Tags    []string `json:"tags,omitempty" jsonschema:"Optional tags"`
}

func handleAddNote(logger *slog.Logger, client *mira.Client) func(ctx context.Context, req *mcp.CallToolRequest, args AddNoteArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args AddNoteArgs) (*mcp.CallToolResult, any, error) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if strings.TrimSpace(args.Title) == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "title is required"},
				},
			}, nil, nil
		}
		if strings.TrimSpace(args.Content) == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: "content is required"},
				},
			}, nil, nil
		}

		input := mira.CreateNoteInput{
			Title:   args.Title,
			Content: args.Content,
			Tags:    args.Tags,
		}

		note, err := client.Create(ctx, input)
		if err != nil {
			logger.Warn("add note failed", "error", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("add note failed: %v", err)},
				},
			}, nil, nil
		}

		text := fmt.Sprintf("Note created: %s (ID: %s)\nEnrichment status: %s", note.Title, note.ID, note.EnrichmentStatus)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}, nil, nil
	}
}

type ListRecentNotesArgs struct {
	Limit int `json:"limit,omitempty" jsonschema:"Maximum results, defaults to 10"`
}

func handleListRecentNotes(logger *slog.Logger, client *mira.Client) func(ctx context.Context, req *mcp.CallToolRequest, args ListRecentNotesArgs) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, args ListRecentNotesArgs) (*mcp.CallToolResult, any, error) {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		notes, err := client.List(ctx, limit)
		if err != nil {
			logger.Warn("list notes failed", "error", err)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("list notes failed: %v", err)},
				},
			}, nil, nil
		}

		var text strings.Builder
		if len(notes) == 0 {
			text.WriteString("No notes found")
		} else {
			fmt.Fprintf(&text, "Recent notes:\n\n")
			for i, note := range notes {
				fmt.Fprintf(&text, "[%d] %s (ID: %s)\n", i+1, note.Title, note.ID)
				fmt.Fprintf(&text, "    Created: %s\n", note.CreatedAt.Format(time.RFC3339))
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text.String()},
			},
		}, nil, nil
	}
}
