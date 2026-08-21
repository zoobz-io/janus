-- +goose Up
-- Trigram indexes backing the admin search endpoints. The search contract does
-- infix ILIKE over entity text fields (applications: name, slug); a GIN index
-- with gin_trgm_ops accelerates those %substring% matches at scale. pg_trgm
-- ships in postgres contrib. This is part of the per-entity search template —
-- every searchable text field gets a trigram GIN index in a later migration.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_applications_name_trgm ON applications USING gin (name gin_trgm_ops);
CREATE INDEX idx_applications_slug_trgm ON applications USING gin (slug gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_applications_slug_trgm;
DROP INDEX IF EXISTS idx_applications_name_trgm;
-- pg_trgm is left installed: other entities' indexes will depend on it, and the
-- extension is harmless if unused.
