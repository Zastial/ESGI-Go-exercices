CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS notes (
    id uuid PRIMARY KEY,
    title text NOT NULL,
    content text NOT NULL,
    summary text,
    score integer,
    enrichment_status text NOT NULL DEFAULT 'pending',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    search_document tsvector GENERATED ALWAYS AS (
        to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(content, ''))
    ) STORED,
    CHECK (enrichment_status IN ('pending', 'done', 'failed'))
);

CREATE TABLE IF NOT EXISTS note_tags (
    note_id uuid NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    tag text NOT NULL,
    PRIMARY KEY (note_id, tag)
);

CREATE TABLE IF NOT EXISTS note_embeddings (
    note_id uuid PRIMARY KEY REFERENCES notes(id) ON DELETE CASCADE,
    embedding vector(8) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notes_search_document_idx ON notes USING GIN (search_document);
CREATE INDEX IF NOT EXISTS note_embeddings_hnsw_idx ON note_embeddings USING hnsw (embedding vector_cosine_ops);
