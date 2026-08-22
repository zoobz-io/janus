# testing/integration

The real integration suite. Nine `_test.go` files that drive the actual Janus
stores against real Postgres and Redis — no mocks, no hand-run database.

This is a **separate Go module**, `github.com/zoobz-io/janus/testing/integration`,
with its own [`go.mod`](go.mod) that `replace`s the root module back to `../..`. It
lives apart because it depends on [testcontainers-go](https://github.com/testcontainers/testcontainers-go)
and the Docker client tree; keeping that out of the root module is the whole reason
for the boundary.

## What TestMain does

Everything is stood up once, in [`testenv_test.go`](testenv_test.go)'s `TestMain`,
before any test runs:

- Starts an ephemeral **`postgres:16`** container and an ephemeral **`redis:7`**
  container via testcontainers, each gated on its own log-based wait strategy.
- Runs the schema by globbing `../../database/migrations/*.sql`, sorting
  lexicographically (matching goose's numeric-prefix order), and executing only the
  `-- +goose Up` half of each file.
- Builds the real store aggregate with `stores.New(db, astqlpg.New())` — the same
  Postgres renderer the binaries use — plus the Redis-backed application labels.

Because the containers are ephemeral and TestMain owns their whole lifecycle, you
provide **Docker and nothing else**. There is no external database to create, seed,
or point a DSN at; the suite terminates both containers on the way out.

## Coverage

The suites span the domain end to end:

| Area | File |
|------|------|
| Entitlement resolution & authz | [`entitlement_test.go`](entitlement_test.go) |
| Application label mapping (Redis) | [`labels_test.go`](labels_test.go) |
| Membership & owner-protection management | [`management_test.go`](management_test.go) |
| Mesh gRPC services (identity, directory, session, entitlement) | [`mesh_test.go`](mesh_test.go) |
| Tenant provisioning | [`provisioning_test.go`](provisioning_test.go) |
| Search (applications, tenants, users) | [`search_test.go`](search_test.go), [`tenants_search_test.go`](tenants_search_test.go), [`users_search_test.go`](users_search_test.go) |
| Store, tenant & user behavior | [`stores_test.go`](stores_test.go) |

## Running

```bash
make test-integration      # go test over this module (needs Docker)
make coverage-integration  # same, plus -coverpkg over the root module → coverage-integration.out
```

`coverage-integration` is what CI runs: it measures coverage of
`github.com/zoobz-io/janus/...` from inside this module and writes the profile to the
repo root for Codecov's `integration` flag. Both targets need Docker.
