# Janus

Identity and access backbone for the zoobz-io mesh.

Every app on the mesh asks Janus the same two questions on every request: **who is
this user, and what are they allowed to do here?** Janus owns the answers — users,
tenants, sessions, and per-application entitlements — and hands them back in a single
round-trip so a downstream service can authenticate *and* authorize a request without
a second hop.

The guiding rule: **Janus validates users, not services.** Service-to-service trust
comes from the mesh CA and client certificates; user trust comes from sessions.
User-facing configuration stays small; everything an operator or a peer service needs
lives behind certificate- or session-guarded surfaces.

## Surfaces

Janus is one domain served four ways. Each surface is a separate binary under
[`cmd/`](cmd/) over the same shared stores; each has its own README with the full
endpoint list and auth model.

| Surface | Binary | Audience | Auth | What it does |
|---------|--------|----------|------|--------------|
| **[Public API](api/)** | [`cmd/api`](cmd/api/) | End users | Session cookie / bearer | Self-service: profile, sessions, linked accounts, tenant creation, own entitlements. ~11 endpoints. |
| **[Mesh API](internal/mesh/)** | [`cmd/mesh`](cmd/mesh/) | Peer services | [sctx](https://github.com/zoobz-io/sctx) client cert | gRPC: resolve/register identities, create & validate sessions, read the directory, manage entitlements. 4 services. |
| **[Admin API](admin/)** | [`cmd/admin`](cmd/admin/) | Operators | Session cookie / bearer | Operator CRUD over every application, tenant, user, and access grant. 33 endpoints, OpenAPI-described. |
| **[Web console](web/)** | — | Operators | Proxied session | Nuxt admin UI, typed against the admin API through a generated SDK. |

Two more binaries support the build rather than serving traffic:
[`cmd/adminspec`](cmd/adminspec/) dumps the admin OpenAPI spec that the web SDK is
generated from, and [`cmd/seed`](cmd/seed/) fills a dev database with representative
data.

## Domain model

IDs are application-generated UUIDs stored as `TEXT`. The schema is defined in
[`database/migrations/001_initial_schema.sql`](database/migrations/001_initial_schema.sql);
field-level detail lives in [`database/models/`](database/models/).

**Identity & directory**

| Entity | Purpose |
|--------|---------|
| **User** | An authenticated person. Belongs to one or more tenants via memberships. |
| **Tenant** | A customer organization. Users join via memberships. |
| **Membership** | A user's role within a tenant (`viewer`, `editor`, `admin`, `owner`), with last-owner protection. |
| **Account** | Maps an external IdP subject (`zitadel`, `auth0`, `github`, `google`) to an internal user — the federation link. |
| **Session** | A mesh-wide session valid across all apps. Only a SHA-256 token hash is stored, never the raw token. |
| **Application** | A product/service on the mesh. Its `slug` matches the cert CN of that service's nodes. |

**Entitlements** — how an application says who may use it, and at what level.

| Entity | Purpose |
|--------|---------|
| **Scope** | A permission an application defines. Janus stores and returns scopes but does not interpret them. |
| **Tier** | A subscription level an application defines (ranked, e.g. `free` → `pro`). |
| **Feature** | Bundles one scope into one tier. A tier's features are the scopes it grants. |
| **License** | Authorizes a **tenant** to use an application. Prerequisite for any grant. |
| **Grant** | Gives a **user** access to an application *within a tenant* — optionally on a tier, plus explicit `roles[]` and `scopes[]`. |

An application owns its own catalog of scopes and tiers. A tenant is licensed for an
application; a user within that tenant is granted access. A user's **effective scopes**
for an app are the union of their grant's explicit scopes and the scopes bundled into
their tier's features — resolved live, so editing a tier takes effect immediately.

### Authorization in one round-trip

When a peer service calls `ValidateSession`, Janus identifies the **calling application
from its client-certificate CN**, then walks user → memberships → **license → grant**
to return the set of **authorized tenants** — each carrying the user's tenant role, the
grant's app-defined roles, and the effective scopes. If the user has no license-and-grant
for the calling app, the session is invalid *for that caller* even though the token is
otherwise valid. One call authenticates and authorizes; the resolver is
[`internal/authz`](internal/authz/), served over gRPC by [`internal/mesh`](internal/mesh/).

## Architecture

Every binary boots through one shared runtime, [`internal/boot`](internal/boot/)
(`boot.Init`): it loads config, connects Postgres and Redis, builds the AES encryptor
and the store aggregate, and registers model boundaries — then hands each binary an
unfrozen registry so it can register its own contracts and freeze.

The two HTTP surfaces ([`api/`](api/) and [`admin/`](admin/)) share one four-package
pattern — **contracts** (interface boundaries over the stores) → **handlers**
([rocco](https://github.com/zoobz-io/rocco) endpoints) → **transformers** (model ↔ wire
mapping) → **wire** (request/response DTOs with validation and boundary masking). The
pattern is documented once in [`api/README.md`](api/); admin reuses it with a broader
capability boundary over the *same* store instances.

Domain mutations emit typed [capitan](https://github.com/zoobz-io/capitan) events
([`events/`](events/)), bridged to OpenTelemetry via
[aperture](https://github.com/zoobz-io/aperture) ([`internal/observe`](internal/observe/),
[`internal/otel`](internal/otel/)).

Built on the zoobz-io stack:
[sum](https://github.com/zoobz-io/sum) (registry, boundaries, field encryption),
[rocco](https://github.com/zoobz-io/rocco) (HTTP, OpenAPI, OAuth/session),
[aegis](https://github.com/zoobz-io/aegis) + [sctx](https://github.com/zoobz-io/sctx)
(mesh node, cert auth),
[astql](https://github.com/zoobz-io/astql) + sqlx/Postgres (data access),
[capitan](https://github.com/zoobz-io/capitan) + [aperture](https://github.com/zoobz-io/aperture)
(events → OpenTelemetry), and
[cereal](https://github.com/zoobz-io/cereal) (AES encryption).

## Repository map

Every directory below carries its own README with the local detail.

```
janus/
├── cmd/               # Entrypoints — one per binary
│   ├── api/           #   public user-facing HTTP API
│   ├── mesh/          #   aegis gRPC mesh node
│   ├── admin/         #   operator HTTP API
│   ├── adminspec/     #   dumps the admin OpenAPI spec (no server)
│   └── seed/          #   dev-only database seeder
├── config/            # Typed configuration structs (APP_* env, via sum)
├── api/               # Public HTTP surface
│   ├── contracts/     #   interface boundaries over the stores
│   ├── handlers/      #   rocco endpoints
│   ├── transformers/  #   model ↔ wire mapping
│   └── wire/          #   request/response DTOs
├── admin/             # Admin HTTP surface (same four-package pattern)
│   ├── contracts/  handlers/  transformers/  wire/
├── database/
│   ├── models/        #   domain models
│   ├── stores/        #   Postgres data access (astql + sqlx)
│   └── migrations/    #   goose SQL migrations
├── internal/
│   ├── auth/          #   cookie + bearer authenticators, OIDC discovery
│   ├── authz/         #   portable role checks + entitlement resolver
│   ├── boot/          #   shared runtime init (boot.Init)
│   ├── labels/        #   application id↔name mapping in Redis
│   ├── mesh/          #   gRPC service implementations
│   ├── observe/       #   aperture schema sync
│   └── otel/          #   OpenTelemetry setup
├── events/            # Typed capitan events
├── web/               # TypeScript monorepo — Nuxt console + generated SDK
├── testing/           # Integration tests (testcontainers) and benchmarks
└── .github/workflows/ # CI + label sync
```

## Getting started

On a fresh machine, bootstrap the whole toolchain in one idempotent step — this
installs the pinned Go toolchain, `golangci-lint`, `gosec`, `air`, and `pnpm`:

```bash
bash tools/setup.sh   # run directly the first time — a fresh box has no make
make setup            # equivalent, once make is installed
```

Then open a new shell (or `source /etc/profile.d/janus-dev.sh`) so `go` and the
Go-installed tools are on your `PATH`. Copy `.env.example` to `.env` and fill in the
OIDC and secret values.

Bring up a surface with its dependencies (Postgres, Redis, migrations) via Docker
Compose, then seed it:

```bash
make dev-api     # public API + postgres + redis + migrate  (profile: api)
make dev-admin   # admin API + the same dependencies         (profile: admin)
make seed        # fill the dev database with fake data
```

Or run a binary directly against an already-running database:

```bash
make run         # public API   (./cmd/api)
make run-admin   # admin API    (./cmd/admin)
make run-mesh    # mesh gRPC node (./cmd/mesh — not wired into compose)
```

### Prerequisites

- Go 1.25 (module target; CI also exercises 1.24)
- golangci-lint v2.7.2
- Node ≥ 22 and pnpm 10 (for the [`web/`](web/) workspace)
- Docker (only for the compose dev stack and integration tests)
- PostgreSQL and Redis (provided by `docker-compose.yml`)

### Make reference

```
Build & run    build            build the api, admin, and mesh binaries
               run / run-admin / run-mesh   run one binary locally

Dev stack      dev-api          public API + deps in compose
               dev-admin        admin API + deps in compose
               dev-observability   grafana, jaeger, prometheus, loki, otel-collector
               dev-down / dev-logs / dev-reset
               seed             seed the dev database (wipes domain tables first)

Test           test             all tests, race detector
               test-unit        unit tests only (-short)
               test-integration integration tests (testcontainers)
               test-bench       benchmarks

Quality        lint / lint-fix  golangci-lint
               coverage         unit coverage report
               check            Go + web tests, lint, typecheck (quick gate)
               ci               full CI simulation (needs Docker)

Web            web-install / web-check / web-lint / web-test / web-build
               openapi-admin    regenerate the admin OpenAPI spec for the SDK

Setup          setup            bootstrap toolchain
               install-tools / install-hooks
```

Run `make help` for the authoritative list.

## License

MIT
