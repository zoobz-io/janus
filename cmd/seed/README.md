# cmd/seed

Fills a development database with a small but complete mesh of representative data.
**Dev only** — it truncates domain tables on every run.

It reuses [`boot.Init`](../../internal/boot/) — the same config load, encryption,
infra connections, store construction, and model boundaries every janus binary
uses — freezes the registry, then drives the concrete stores directly (it registers
no contracts of its own).

## What it builds

`reset` runs one raw statement — a `TRUNCATE ... RESTART IDENTITY CASCADE` over the
domain tables (config left intact) — so the seeder is **safe to re-run**. Then it
seeds, in dependency order:

- **Applications** with their own scope catalogs, ranked tiers, and features
  (tiers bundling scopes) — including `janus-admin`, janus's own admin portal
  registered as a first-class application, and an intentionally inactive `vault`.
- **Users** with linked identity accounts and a handful of sessions.
- **Tenants** with members (roles), application **licenses**, and per-user
  **grants** (roles, explicit scopes, optional tier).

The seed set is defined as data at the top of [`main.go`](main.go) — edit the `apps`,
`users`, and `orgs` tables to change it.

## Run it

```bash
make seed   # docker compose --profile seed, in the compose network
```

Or from the host against the published dev ports, given the dev environment (e.g.
`APP_DB_HOST=localhost` via your local `.env`).
