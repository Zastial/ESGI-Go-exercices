package store

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mira/internal/core"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, input core.CreateNoteInput) (core.Note, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return core.Note{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	noteID := uuid.NewString()
	now := time.Now().UTC()
	if _, err := tx.Exec(ctx, `
		INSERT INTO notes (id, title, content, enrichment_status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $4)
	`, noteID, input.Title, input.Content, now); err != nil {
		return core.Note{}, err
	}

	if err := saveTags(ctx, tx, noteID, input.Tags); err != nil {
		return core.Note{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return core.Note{}, err
	}

	return s.loadOne(ctx, s.pool.QueryRow(ctx, noteByIDQuery, noteID))
}

func (s *PostgresStore) Get(ctx context.Context, id string) (core.Note, error) {
	return s.loadOne(ctx, s.pool.QueryRow(ctx, noteByIDQuery, id))
}

func (s *PostgresStore) List(ctx context.Context, params core.ListParams) ([]core.Note, int, error) {
	notes, err := s.loadNotes(ctx, listNotesQuery, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notes`).Scan(&total); err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

func (s *PostgresStore) Update(ctx context.Context, id string, input core.UpdateNoteInput) (core.Note, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return core.Note{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := s.loadOne(ctx, tx.QueryRow(ctx, noteByIDQuery, id))
	if err != nil {
		return core.Note{}, err
	}

	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Content != nil {
		current.Content = *input.Content
	}
	if input.Tags != nil {
		current.Tags = append([]string(nil), *input.Tags...)
	}

	current.EnrichmentStatus = "pending"
	current.Summary = nil
	current.Score = nil
	current.UpdatedAt = time.Now().UTC()

	if _, err := tx.Exec(ctx, `
		UPDATE notes
		SET title = $2, content = $3, summary = NULL, score = NULL, enrichment_status = 'pending', updated_at = $4
		WHERE id = $1
	`, id, current.Title, current.Content, current.UpdatedAt); err != nil {
		return core.Note{}, err
	}

	if input.Tags != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM note_tags WHERE note_id = $1`, id); err != nil {
			return core.Note{}, err
		}
		if err := saveTags(ctx, tx, id, *input.Tags); err != nil {
			return core.Note{}, err
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM note_embeddings WHERE note_id = $1`, id); err != nil {
		return core.Note{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return core.Note{}, err
	}

	return s.loadOne(ctx, s.pool.QueryRow(ctx, noteByIDQuery, id))
}

func (s *PostgresStore) Delete(ctx context.Context, id string) error {
	cmdTag, err := s.pool.Exec(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) Search(ctx context.Context, query string, params core.ListParams) ([]core.Note, int, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.List(ctx, params)
	}

	embedding := queryEmbedding(query)
	notes, err := s.loadNotes(ctx, searchNotesQuery, query, embedding, params.Limit, params.Offset)
	if err != nil {
		return nil, 0, err
	}

	var total int
	if err := s.pool.QueryRow(ctx, searchCountQuery, query, embedding).Scan(&total); err != nil {
		return nil, 0, err
	}

	return notes, total, nil
}

func (s *PostgresStore) ApplyEnrichment(ctx context.Context, id string, result core.EnrichmentResult) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := s.loadOne(ctx, tx.QueryRow(ctx, noteByIDQuery, id))
	if err != nil {
		return err
	}

	mergedTags := mergeTags(current.Tags, result.Tags)
	if _, err := tx.Exec(ctx, `
		UPDATE notes
		SET summary = $2, score = $3, enrichment_status = 'done', updated_at = $4
		WHERE id = $1
	`, id, nullString(result.Summary), result.Score, time.Now().UTC()); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM note_tags WHERE note_id = $1`, id); err != nil {
		return err
	}
	if err := saveTags(ctx, tx, id, mergedTags); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO note_embeddings (note_id, embedding, created_at, updated_at)
		VALUES ($1, $2::vector, NOW(), NOW())
		ON CONFLICT (note_id)
		DO UPDATE SET embedding = EXCLUDED.embedding, updated_at = NOW()
	`, id, result.Embedding); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresStore) MarkEnrichmentFailed(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE notes
		SET enrichment_status = 'failed', updated_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (s *PostgresStore) loadNotes(ctx context.Context, query string, args ...any) ([]core.Note, error) {
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := make([]core.Note, 0)
	for rows.Next() {
		note, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *PostgresStore) loadOne(ctx context.Context, row interface{ Scan(...any) error }) (core.Note, error) {
	note, err := scanNote(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return core.Note{}, core.ErrNotFound
		}
		return core.Note{}, err
	}
	return note, nil
}

func saveTags(ctx context.Context, tx pgx.Tx, noteID string, tags []string) error {
	unique := make(map[string]struct{})
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" {
			unique[tag] = struct{}{}
		}
	}

	ordered := make([]string, 0, len(unique))
	for tag := range unique {
		ordered = append(ordered, tag)
	}
	sort.Strings(ordered)

	for _, tag := range ordered {
		if _, err := tx.Exec(ctx, `
			INSERT INTO note_tags (note_id, tag)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, noteID, tag); err != nil {
			return err
		}
	}
	return nil
}

func mergeTags(a, b []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(a)+len(b))
	for _, tag := range append(a, b...) {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	sort.Strings(result)
	return result
}

func nullString(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func scanNote(scanner interface{ Scan(...any) error }) (core.Note, error) {
	var note core.Note
	var summary sql.NullString
	var score sql.NullInt32
	var tags []string
	if err := scanner.Scan(&note.ID, &note.Title, &note.Content, &tags, &summary, &score, &note.EnrichmentStatus, &note.CreatedAt, &note.UpdatedAt); err != nil {
		return core.Note{}, err
	}
	note.Tags = tags
	if summary.Valid {
		value := summary.String
		note.Summary = &value
	}
	if score.Valid {
		value := int(score.Int32)
		note.Score = &value
	}
	return note, nil
}

const noteSelect = `
SELECT
	n.id::text,
	n.title,
	n.content,
	COALESCE(array_agg(DISTINCT nt.tag) FILTER (WHERE nt.tag IS NOT NULL), '{}'::text[]) AS tags,
	n.summary,
	n.score,
	n.enrichment_status,
	n.created_at,
	n.updated_at
FROM notes n
LEFT JOIN note_tags nt ON nt.note_id = n.id
`

const noteByIDQuery = noteSelect + `
WHERE n.id = $1
GROUP BY n.id
`

const listNotesQuery = noteSelect + `
GROUP BY n.id
ORDER BY n.created_at DESC
LIMIT $1 OFFSET $2
`

const searchNotesQuery = `
WITH q AS (
	SELECT websearch_to_tsquery('simple', $1) AS tsq, $2::vector AS embedding
)
` + noteSelect + `
LEFT JOIN note_embeddings ne ON ne.note_id = n.id
CROSS JOIN q
WHERE q.tsq @@ to_tsvector('simple', coalesce(n.title, '') || ' ' || coalesce(n.content, ''))
   OR COALESCE(1 - (ne.embedding <=> q.embedding), 0) > 0.3
GROUP BY n.id, n.title, n.content, n.summary, n.score, n.enrichment_status, n.created_at, n.updated_at, q.tsq, q.embedding, ne.embedding
ORDER BY
	(ts_rank_cd(to_tsvector('simple', coalesce(n.title, '') || ' ' || coalesce(n.content, '')), q.tsq) + COALESCE(1 - (ne.embedding <=> q.embedding), 0)) DESC,
	n.created_at DESC
LIMIT $3 OFFSET $4
`

const searchCountQuery = `
WITH q AS (
	SELECT websearch_to_tsquery('simple', $1) AS tsq, $2::vector AS embedding
)
SELECT COUNT(DISTINCT n.id)
FROM notes n
LEFT JOIN note_embeddings ne ON ne.note_id = n.id
CROSS JOIN q
WHERE q.tsq @@ to_tsvector('simple', coalesce(n.title, '') || ' ' || coalesce(n.content, ''))
   OR COALESCE(1 - (ne.embedding <=> q.embedding), 0) > 0.3
`

func queryEmbedding(text string) string {
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

	parts := make([]string, 0, len(features))
	for _, value := range features {
		parts = append(parts, strconv.FormatFloat(value, 'f', 4, 64))
	}
	return "[" + strings.Join(parts, ",") + "]"
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
