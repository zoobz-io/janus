# Janus

Cross-application identity and access backbone for the zoobz-io mesh.

## Overview

Janus is the centralized identity service that other apps in the zoobz-io ecosystem depend on for authentication and authorization. It owns users, tenants, sessions, and per-application entitlements, and exposes them over two surfaces:

- **HTTP API** — user-facing only. End users log in via OIDC and manage their own profile, sessions, linked identities, and tenants.
- **gRPC mesh API** — service-to-service. Other apps on the mesh resolve identities, validate sessions, read the directory, and manage entitlements. Every RPC is authenticated with certificate-based [sctx](https://github.com/zoobz-io/sctx) guards.

The guiding rule: **Janus validates users, not services.** Service-to-service trust comes from the mesh CA and certificates; user trust comes from sessions. Service configuration happens exclusively over gRPC — the HTTP surface stays small and user-scoped.

Built on the zoobzio stack:

- [sum](https://github.com/zoobz-io/sum) — service registry, boundaries, field encryption
- [rocco](https://github.com/zoobz-io/rocco) — HTTP server, OpenAPI, OAuth/session
- [aegis](https://github.com/zoobz-io/aegis) + [sctx](https://github.com/zoobz-io/sctx) — mesh node and cert-based auth
- [astql](https://github.com/zoobz-io/astql) + sqlx/Postgres — data access
- [fig](https://github.com/zoobz-io/fig) — configuration
- [capitan](https://github.com/zoobz-io/capitan) — typed events, bridged to OpenTelemetry via [aperture](https://github.com/zoobz-io/aperture)
- [cereal](https://github.com/zoobz-io/cereal) — AES encryption

## Domain Model

| Entity | Purpose |
|--------|---------|
| **User** | An authenticated person. Belongs to one or more tenants via memberships. |
| **Tenant** | A customer/company. Users join via memberships. |
| **Membership** | A user's role within a tenant (`viewer`, `editor`, `admin`, `owner`), with owner protection. |
| **LinkedIdentity** | Maps an external IdP subject (zitadel, auth0, github, google) to an internal user, enabling federated login across the mesh. |
| **Session** | A mesh-wide token valid across all apps. Only a SHA-256 hash is stored — never the raw token. |
| **Application** | A product/service on the mesh. Its `Slug` matches the cert CN of that service's nodes. |
| **TenantApplication** | Records that a tenant is authorized to use an application. |
| **UserApplication** | Grants a user access to an application within a tenant, carrying app-defined roles/scopes that Janus stores and returns but does not interpret. |

### Session validation flow

When an app calls `ValidateSession`, Janus identifies the **calling application from its certificate CN**, then walks user → memberships → `TenantApplication` → `UserApplication` to return the set of **authorized tenants** — each with the user's tenant role plus the app-defined roles/scopes. If the user has no entitlement for the calling app, the session is invalid *for that caller*. This lets a downstream app authenticate and authorize a request in a single round-trip.

## API Surface

### HTTP (user-facing)

| Method | Path | Purpose |
|--------|------|---------|
| `GET`/`PUT` | `/me` | Get / update own profile |
| `POST` | `/me/tenants` | Create a tenant |
| `GET` | `/me/sessions` | List own sessions |
| `DELETE` | `/me/sessions/{id}` | Revoke a session |
| `DELETE` | `/me/sessions` | Revoke all own sessions |
| `GET` | `/me/identities` | List linked identities |
| `DELETE` | `/me/identities/{id}` | Unlink an identity |
| `GET` | `/applications` | List the application catalog |
| `GET` | `/me/tenants/{tenant_id}/applications` | List apps the user can access in a tenant |
| — | `/login`, `/callback`, `/logout` | OIDC login flow |

### gRPC (service-to-service, sctx-guarded)

- **IdentityService** — resolve/register users by IdP subject, list providers
- **SessionService** — create / validate / revoke sessions, list a user's sessions (event streaming stubbed)
- **DirectoryService** — read users/tenants, create/update tenants
- **EntitlementService** — authorize/revoke apps for tenants, grant/revoke/update per-user access

## Project Structure

```
janus/
├── cmd/app/          # Application entrypoint and wiring
├── config/           # Typed configuration (fig)
├── api/
│   ├── contracts/    # Interface definitions
│   ├── handlers/     # HTTP handlers (user-facing)
│   ├── wire/         # Request/response DTOs
│   └── transformers/ # Model ↔ wire mapping
├── database/
│   ├── models/       # Domain models
│   ├── stores/       # Postgres data access
│   └── migrations/   # SQL migrations
├── internal/
│   ├── auth/         # Cookie + bearer authenticators
│   ├── authz/        # Portable authorization helpers
│   ├── mesh/         # gRPC service implementations
│   ├── boot/         # Infrastructure connection helpers
│   ├── observe/      # Aperture schema sync
│   └── otel/         # OpenTelemetry setup
├── events/           # Typed capitan events
├── testing/          # Integration tests and infra
└── .github/workflows # CI
```

Each directory contains a README explaining its purpose and patterns.

## Getting Started

```bash
# Install dependencies
go mod tidy

# Run the application
make run

# Run tests
make test

# Run linter
make lint

# Full CI check
make check
```

### Prerequisites

- Go 1.24+
- golangci-lint v2.7.2
- PostgreSQL and Redis (see `docker-compose.yml`)

### Install Tools

On a fresh machine, bootstrap the whole toolchain in one idempotent step. This
installs the pinned Go toolchain, `golangci-lint`, `gosec`, `air`, and `pnpm`:

```bash
bash tools/setup.sh   # run directly the first time — a fresh box has no make
make setup            # equivalent, once make is installed
```

Then open a new shell (or source `/etc/profile.d/janus-dev.sh`) so `go` and the
Go-installed tools are on your `PATH`.

Docker is not installed by the script — it's only needed for the compose-based
dev database and integration tests. Install it separately if you need that stack.

If you already have the Go toolchain and only want the lint/hook helpers:

```bash
make install-tools
make install-hooks
```

### Make Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application binary |
| `make run` | Run the application |
| `make test` | Run all tests with race detector |
| `make test-unit` | Run unit tests only |
| `make test-integration` | Run integration tests |
| `make test-bench` | Run benchmarks |
| `make lint` | Run linters |
| `make coverage` | Generate coverage report |
| `make check` | Run tests + lint |
| `make ci` | Full CI simulation |

## License

MIT
