# internal/observe

Aperture schema synchronization from Postgres, via [flux](https://github.com/zoobz-io/flux).
The observability schema — what capitan events map to which OTEL signals — lives in a
config row, and this package keeps the running [aperture](https://github.com/zoobz-io/aperture)
instance in step with it, no redeploy required to change what's observed.

## StartSchemaSync

[`schema.go`](schema.go). `StartSchemaSync(ctx, ap)` builds a
`flux.Capacitor[aperture.Schema]` over a `ConfigWatcher` for the `"aperture_schema"`
config key, YAML-decodes each new value, and applies it to the running aperture instance
(`ap.Apply`). Its `Start` **blocks until the initial schema loads** — the process comes up
already configured, never observing blind.

## ConfigWatcher

[`watcher.go`](watcher.go). Implements `flux.Watcher` by **polling** the config service
every `PollInterval` (**30s**). It emits the initial value immediately, then re-emits only
when the value actually changes (plain string comparison). Read errors mid-poll are
**skipped**, not surfaced — a transient config-store hiccup won't tear down the watcher or
clobber the live schema. Reads go through
`sum.MustUse[contracts.Config].GetByKey`.

The `aperture_schema` row itself is seeded by
[migration 002](../../database/migrations/002_aperture_config.sql).
