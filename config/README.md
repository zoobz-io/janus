# Config

Typed configuration structs. One struct per concern, each with a `Validate()` that fails
fast on missing or malformed values. Loaded and registered through
[sum](https://github.com/zoobz-io/sum) — `sum.Config[T]` reads and validates, `sum.MustUse[T]`
resolves a loaded struct from context (see [`internal/boot`](../internal/boot/)). fig is
an indirect dependency reached through sum, not used here directly.

Env prefix is `APP_*` for everything **except** OpenTelemetry, which follows the OTEL
convention and reads `OTEL_*`. Struct tags carry the wiring: `env` (variable name),
`default` (fallback), `secret` (secret-store path for sensitive values).

## The structs

| Struct | Env vars | Notes |
|--------|----------|-------|
| **App** ([`app.go`](app.go)) | `APP_PORT` (8080) | Public API listen port. |
| **Admin** ([`admin.go`](admin.go)) | `APP_ADMIN_PORT` (8081) | Admin API listen port. |
| **Auth** ([`auth.go`](auth.go)) | `APP_AUTH_ISSUER`, `APP_AUTH_PROVIDER`, `APP_AUTH_CLIENT_ID`, `APP_AUTH_CLIENT_SECRET`, `APP_AUTH_REDIRECT_URI`, `APP_AUTH_COOKIE_KEY`, `APP_AUTH_POST_LOGIN_REDIRECT` | Full OIDC + session config. `CookieKey()` decodes the cookie key from hex. `Provider` is the name linked accounts are recorded under. |
| **Cookie** ([`cookie.go`](cookie.go)) | `APP_AUTH_COOKIE_KEY` | A session-only subset of Auth: the cookie signing key alone, loaded by surfaces that validate sessions but never initiate OIDC (the admin API). `Key()` decodes it from hex. |
| **Database** ([`database.go`](database.go)) | `APP_DB_HOST`, `APP_DB_NAME`, `APP_DB_USER`, `APP_DB_PASSWORD`, `APP_DB_SSLMODE`, `APP_DB_PORT` | `DSN()` builds the Postgres connection string. |
| **Encryption** ([`encryption.go`](encryption.go)) | `APP_ENCRYPTION_KEY` | Must be 64 hex characters (32 bytes) or `Validate()` rejects it. |
| **Mesh** ([`mesh.go`](mesh.go)) | `APP_MESH_ID`, `APP_MESH_NAME`, `APP_MESH_HOST`, `APP_MESH_CERT_DIR`, `APP_MESH_PORT` | aegis mesh node. `Addr()` returns `host:port`. |
| **OTEL** ([`otel.go`](otel.go)) | `OTEL_EXPORTER_OTLP_ENDPOINT` (4318) | The only `OTEL_*` struct. |
| **Redis** ([`redis.go`](redis.go)) | `APP_REDIS_ADDR` | Label cache / session backing. |

Secrets — `APP_AUTH_CLIENT_SECRET`, `APP_AUTH_COOKIE_KEY`, `APP_DB_PASSWORD`,
`APP_ENCRYPTION_KEY` — carry a `secret:` tag naming their secret-store path, so they can
be resolved from a secret manager rather than a plaintext env var.
