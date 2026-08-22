# Migrations

[goose](https://github.com/pressly/goose) SQL migrations. Each file is
`NNN_name.sql` with a `-- +goose Up` block and a `-- +goose Down` block; goose applies
them in numeric order and records what's been run. IDs are `TEXT` throughout — no
`pgcrypto`, no `gen_random_uuid()`, because the application generates UUIDs before insert.

## The seven

| # | File | What it does |
|---|------|--------------|
| **001** | [`001_initial_schema.sql`](001_initial_schema.sql) | The eleven core tables: `users`, `tenants`, `memberships`, `accounts`, `sessions`, `applications`, `scopes`, `tiers`, `features`, `licenses`, `grants`. IDs `TEXT`; `features` FKs to `tiers`/`scopes` are `ON DELETE CASCADE`. |
| **002** | [`002_aperture_config.sql`](002_aperture_config.sql) | The `config` key/value table, and seeds the `aperture_schema` row consumed by [`internal/observe`](../../internal/observe/) — the capitan → OTEL metric definitions and log whitelist. |
| **003** | [`003_search_indexes.sql`](003_search_indexes.sql) | Enables the `pg_trgm` extension and adds trigram GIN indexes on `applications` `name` and `slug`, backing infix ILIKE search. |
| **004** | [`004_applications_name_unique.sql`](004_applications_name_unique.sql) | `UNIQUE` on `applications.name`, so the name can serve as the label identity resolved through Redis by [`internal/labels`](../../internal/labels/). |
| **005** | [`005_fk_indexes.sql`](005_fk_indexes.sql) | btree indexes on the FK columns queries actually filter on: `memberships.tenant_id`, `accounts.user_id`, `sessions.user_id`, `features.scope_id`, `licenses.application_id`, `grants.tenant_id`, `grants.tier_id`. FKs already covered by a unique constraint's leading column are deliberately skipped. |
| **006** | [`006_users_search_indexes.sql`](006_users_search_indexes.sql) | Trigram GIN on `users` `email` and `display_name`. |
| **007** | [`007_tenants_search_indexes.sql`](007_tenants_search_indexes.sql) | Trigram GIN on `tenants` `name` and `slug`. |

## Adding one

Number it next in sequence, write both `Up` and `Down`, and keep the `Down` a clean
inverse — 005 drops in reverse order, 003 leaves `pg_trgm` installed because later
migrations depend on it. Run them through the compose dev stack (`make dev-*` migrates)
or goose directly.
