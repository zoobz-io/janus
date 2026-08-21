-- +goose Up
-- Admin user search. Trigram GIN indexes accelerate infix ILIKE over the
-- searchable text fields; the UNIQUE btree on email serves equality, not infix
-- matching, so the trigram index is still needed. pg_trgm was installed in 003.
-- No column changes: users already has created_at/updated_at.
CREATE INDEX idx_users_email_trgm ON users USING gin (email gin_trgm_ops);
CREATE INDEX idx_users_display_name_trgm ON users USING gin (display_name gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_users_display_name_trgm;
DROP INDEX IF EXISTS idx_users_email_trgm;
