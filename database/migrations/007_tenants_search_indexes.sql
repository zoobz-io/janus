-- +goose Up
-- Admin tenant search. Trigram GIN indexes accelerate infix ILIKE over the
-- searchable text fields; the UNIQUE btree on slug serves equality, not infix
-- matching, so the trigram index is still needed. pg_trgm was installed in 003.
-- No column changes: tenants already has created_at/updated_at.
CREATE INDEX idx_tenants_name_trgm ON tenants USING gin (name gin_trgm_ops);
CREATE INDEX idx_tenants_slug_trgm ON tenants USING gin (slug gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_tenants_slug_trgm;
DROP INDEX IF EXISTS idx_tenants_name_trgm;
