# Models

The domain types. Each is a plain Go struct with `db` tags; the store layer maps it to
one Postgres table. IDs are application-generated UUIDs stored as `TEXT` — Go `string`
fields, never `int64`, never a database default. For how these entities relate — licenses,
grants, effective scopes, the one-round-trip resolver — see the [root README](../../README.md).

## Persisted entities

Eleven types, each 1:1 with a table in
[`migrations/001_initial_schema.sql`](../migrations/001_initial_schema.sql).

### Identity & directory

| Type | Table | Key fields | Uniqueness |
|------|-------|-----------|------------|
| **User** ([`user.go`](user.go)) | `users` | `email`, `display_name`, `status` (`active`/`inactive`), `last_seen_at *time.Time` | `email` unique |
| **Tenant** ([`tenant.go`](tenant.go)) | `tenants` | `name`, `slug`, `status` (`active`/`suspended`) | `slug` unique |
| **Membership** ([`membership.go`](membership.go)) | `memberships` | `user_id`, `tenant_id`, `role` (`viewer`/`editor`/`admin`/`owner`) | `unique(user_id, tenant_id)` |
| **Account** ([`account.go`](account.go)) | `accounts` | `user_id`, `provider` (`zitadel`/`auth0`/`github`/`google`), `external_subject` | `unique(provider, external_subject)` |
| **Session** ([`session.go`](session.go)) | `sessions` | `token_hash`, `issued_by`, `user_agent`, `ip_address`, `expires_at` | `token_hash` unique |
| **Application** ([`application.go`](application.go)) | `applications` | `name`, `slug`, `status` | `slug` unique (and `name`, added by [`004`](../migrations/004_applications_name_unique.sql)) |

**Account** carries a former `LinkedIdentity` heritage: its method receivers are still
named `l`, and the identity events it drives are `IdentityLinked`/`IdentityUnlinked`.

**Session** stores only a SHA-256 hash of the token. `TokenHash` is tagged `json:"-"` —
the raw token never leaves the process and never lands in the database; the hash is
already irreversible, so no boundary encryption is applied. `Expired()` reports whether
`ExpiresAt` has passed.

### Entitlements

| Type | Table | Key fields | Uniqueness |
|------|-------|-----------|------------|
| **Scope** ([`scope.go`](scope.go)) | `scopes` | `application_id`, `name`, `description` | `unique(application_id, name)` |
| **Tier** ([`tier.go`](tier.go)) | `tiers` | `application_id`, `slug`, `name`, `rank int` | `unique(application_id, slug)` |
| **Feature** ([`feature.go`](feature.go)) | `features` | `tier_id`, `scope_id` | `unique(tier_id, scope_id)` |
| **License** ([`license.go`](license.go)) | `licenses` | `tenant_id`, `application_id` | `unique(tenant_id, application_id)` |
| **Grant** ([`grant.go`](grant.go)) | `grants` | `user_id`, `tenant_id`, `application_id`, `tier_id *string` (nullable), `roles`, `scopes` | `unique(user_id, tenant_id, application_id)` |

**Feature** bundles one scope into one tier. Both foreign keys are `ON DELETE CASCADE` —
delete a tier or a scope and its feature rows go with it.

**Grant**'s `Roles` and `Scopes` are `pq.StringArray` mapped to Postgres `text[]`
(`type:"text[]"`). janus stores and returns them; it does not interpret them. `TierID` is
a nullable `*string` — a grant may sit on a tier or on none.

## Non-persisted helpers

These live in the package but back no table.

- **AuthorizedTenant** ([`authorization.go`](authorization.go)) — the resolved view of a
  user's access to one application through one tenant: `tenant_id`, `tenant_name`,
  `Role`, `AppRoles`, `AppScopes`. Computed from License + Grant (+ tier features) at
  resolution time by [`internal/authz`](../../internal/authz/), never stored.
- **Config** ([`config.go`](config.go)) — a `key`/`value` runtime-config row. Its table
  is added by [`migrations/002`](../migrations/002_aperture_config.sql), not 001; it backs
  the aperture schema polled by [`internal/observe`](../../internal/observe/).
- **OffsetPage** / **OffsetResult[T]** ([`pagination.go`](pagination.go)) — offset
  pagination. `PageSize()` clamps to `DefaultPageSize` (20) when unset and `MaxPageSize`
  (100) at the ceiling.
