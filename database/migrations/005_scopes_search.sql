-- +goose Up
-- Cross-application scope search. updated_at mirrors the applications template so
-- scopes carry the same sort/date-filter surface. Existing rows backfill to the
-- migration time (they have no prior update history) — acceptable for a
-- freshly-added column. Trigram GIN indexes accelerate infix ILIKE over the
-- searchable text fields; pg_trgm was installed in 003.
ALTER TABLE scopes ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE INDEX idx_scopes_name_trgm ON scopes USING gin (name gin_trgm_ops);
CREATE INDEX idx_scopes_description_trgm ON scopes USING gin (description gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_scopes_description_trgm;
DROP INDEX IF EXISTS idx_scopes_name_trgm;
ALTER TABLE scopes DROP COLUMN updated_at;
