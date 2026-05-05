-- +goose Up
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS role;
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'viewer';
ALTER TABLE users ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
