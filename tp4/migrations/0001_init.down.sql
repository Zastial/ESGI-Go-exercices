DROP INDEX IF EXISTS note_embeddings_hnsw_idx;
DROP INDEX IF EXISTS notes_search_document_idx;
DROP TABLE IF EXISTS note_embeddings;
DROP TABLE IF EXISTS note_tags;
DROP TABLE IF EXISTS notes;
DROP EXTENSION IF EXISTS vector;
