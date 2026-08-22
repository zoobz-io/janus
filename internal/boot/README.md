# internal/boot

The shared runtime init every janus binary calls before it serves anything. One function
loads config, connects Postgres and Redis, builds the AES encryptor and the store
aggregate, and registers model boundaries — so [`cmd/api`](../../cmd/api/),
[`cmd/mesh`](../../cmd/mesh/), [`cmd/admin`](../../cmd/admin/), and the tooling binaries
don't each reinvent boot.

## `Init` — the common boot

[`runtime.go`](runtime.go). `Init(ctx)` does, in order:

1. `sum.New()` + `sum.Start()` — the service and the registry key.
2. Load the four shared configs: `config.Database`, `config.Redis`, `config.Encryption`,
   `config.OTEL`.
3. Decode the **hex** encryption key into a `cereal.AES` encryptor and register it via
   `svc.WithEncryptor(cereal.EncryptAES, ...)` — this is what makes model field
   encryption work downstream.
4. Connect Postgres, **then** Redis — and if Redis fails, close the DB before returning,
   so a half-open boot leaks nothing.
5. Build the store aggregate: `stores.New(db, astqlpg.New())`.
6. Register model boundaries for `Tenant`, `User`, `Membership`, `Account`, `Session`,
   `Application`, `License`, `Grant`.

It returns `Runtime{Svc, K, DB, Redis, Stores}`.

Two deliberate hand-offs:

- **The registry is left unfrozen.** Boot registers the model boundaries every binary
  shares; each binary then registers its *own* contracts (public, admin, or mesh) and
  calls `Freeze`. Boot can't freeze — it doesn't know the surface.
- **The caller owns the connections.** `DB` and `Redis` are yours to `Close` — defer both.

## Per-concern connectors

[`boot.go`](boot.go) holds the individual connectors `Init` composes, each reading its
config via `sum.MustUse` and handing back a client whose lifecycle the caller owns:

| Connector | Builds |
|-----------|--------|
| `Database` | `sqlx.Connect("postgres", cfg.DSN())` |
| `Redis` | a go-redis client, verified with a `Ping` (closed on failure) |
| `OTEL(serviceName)` | the [`internal/otel`](../otel/) providers |
| `Aperture(providers)` | bridges `capitan.Default()` events to the OTEL log/metric/trace providers |
