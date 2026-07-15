package service

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"mira/internal/core"
)

type Processor struct {
	repo core.NoteRepository
}

func NewProcessor(repo core.NoteRepository) *Processor {
	return &Processor{repo: repo}
}

func (p *Processor) Process(ctx context.Context, noteID string) error {
	note, err := p.repo.Get(ctx, noteID)
	if err != nil {
		return err
	}

	combined := note.Title + " " + note.Content
	result := core.EnrichmentResult{
		Tags:      enrichTags(note.Tags, combined),
		Summary:   buildSummary(note.Content),
		Score:     buildScore(combined),
		Embedding: buildEmbedding(combined),
	}

	if err := p.repo.ApplyEnrichment(ctx, noteID, result); err != nil {
		return err
	}
	return nil
}

func (p *Processor) MarkFailed(ctx context.Context, noteID string) error {
	return p.repo.MarkEnrichmentFailed(ctx, noteID)
}

func enrichTags(existing []string, text string) []string {
	tags := make(map[string]struct{})
	for _, tag := range existing {
		normalized := normalizeTag(tag)
		if normalized != "" {
			tags[normalized] = struct{}{}
		}
	}

	words := splitWords(text)
	for _, word := range words {
		if len(word) < 4 {
			continue
		}
		if isStopWord(word) {
			continue
		}
		tags[word] = struct{}{}
	}

	result := make([]string, 0, len(tags))
	for tag := range tags {
		result = append(result, tag)
	}
	sort.Strings(result)
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

func normalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

func splitWords(text string) []string {
	parts := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	words := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			words = append(words, part)
		}
	}
	return words
}

func isStopWord(word string) bool {
	switch word {
	case "the", "and", "for", "avec", "dans", "des", "les", "une", "un", "que", "qui", "sur", "par", "theory", "note":
		return true
	default:
		return false
	}
}

func buildSummary(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	for i, r := range content {
		if r == '.' || r == '!' || r == '?' {
			if i > 0 {
				return trimSummary(content[:i+1])
			}
		}
		if i >= 160 {
			break
		}
	}
	return trimSummary(content)
}

func trimSummary(summary string) string {
	summary = strings.TrimSpace(summary)
	if len(summary) <= 160 {
		return summary
	}
	return strings.TrimSpace(summary[:160])
}

func buildScore(text string) int {
	words := splitWords(text)
	unique := make(map[string]struct{})
	for _, word := range words {
		unique[word] = struct{}{}
	}
	score := len(words)*3 + len(unique)*5
	if score > 100 {
		score = 100
	}
	return score
}

func buildEmbedding(text string) string {
	features := []float64{0, 0, 0, 0, 0, 0, 0, 0}
	words := splitWords(text)
	features[0] = float64(len(words))
	features[1] = float64(len(text))
	features[2] = float64(countVowels(text))
	features[3] = float64(countDigits(text))
	features[4] = float64(len(uniqueWords(words)))
	features[5] = float64(countUppercase(text))
	features[6] = float64(strings.Count(text, " "))
	features[7] = float64(strings.Count(strings.ToLower(text), "go"))

	for i, feature := range features {
		features[i] = feature / 100.0
	}
	return vectorLiteral(features)
}

func uniqueWords(words []string) map[string]struct{} {
	result := make(map[string]struct{}, len(words))
	for _, word := range words {
		result[word] = struct{}{}
	}
	return result
}

func countVowels(text string) int {
	count := 0
	for _, r := range strings.ToLower(text) {
		switch r {
		case 'a', 'e', 'i', 'o', 'u', 'y':
			count++
		}
	}
	return count
}

func countDigits(text string) int {
	count := 0
	for _, r := range text {
		if unicode.IsDigit(r) {
			count++
		}
	}
	return count
}

func countUppercase(text string) int {
	count := 0
	for _, r := range text {
		if unicode.IsUpper(r) {
			count++
		}
	}
	return count
}

func vectorLiteral(values []float64) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(value, 'f', 4, 64))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
