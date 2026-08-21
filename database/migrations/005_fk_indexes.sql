-- +goose Up
-- Secondary indexes on FK columns that real queries filter on. 001 created
-- only PKs and unique constraints; a unique constraint's leading column serves
-- as an index, so FKs already covered that way are deliberately absent here:
--   memberships(user_id), scopes(application_id), tiers(application_id),
--   features(tier_id), licenses(tenant_id), grants(user_id).

-- Memberships.ListByTenant / CountOtherOwners filter tenant_id.
CREATE INDEX idx_memberships_tenant_id ON memberships (tenant_id);

-- Accounts.ListByUser filters user_id.
CREATE INDEX idx_accounts_user_id ON accounts (user_id);

-- Sessions.ListByUser / RevokeUserSessions filter user_id (plus expires_at).
CREATE INDEX idx_sessions_user_id ON sessions (user_id);

-- Features cascade-delete when a scope is deleted; the FK check scans scope_id.
CREATE INDEX idx_features_scope_id ON features (scope_id);

-- Licenses.ListByApplication filters application_id.
CREATE INDEX idx_licenses_application_id ON licenses (application_id);

-- Grants.ListByTenantAndApp leads with tenant_id.
CREATE INDEX idx_grants_tenant_id ON grants (tenant_id);

-- Deleting a tier requires the FK check to scan grants.tier_id.
CREATE INDEX idx_grants_tier_id ON grants (tier_id);

-- +goose Down
DROP INDEX IF EXISTS idx_grants_tier_id;
DROP INDEX IF EXISTS idx_grants_tenant_id;
DROP INDEX IF EXISTS idx_licenses_application_id;
DROP INDEX IF EXISTS idx_features_scope_id;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_accounts_user_id;
DROP INDEX IF EXISTS idx_memberships_tenant_id;
