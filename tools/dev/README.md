# tools/dev

The local dev stack behind `docker-compose.yml`. These files are the config the
compose services mount; you rarely touch them directly — you drive them through the
`make dev-*` targets.

## Files

| File | Purpose |
|------|---------|
| [`Dockerfile.dev`](Dockerfile.dev) | Hot-reload dev image (runs `air`). |
| [`air.api.toml`](air.api.toml) | `air` hot-reload config for the public API (`cmd/api`). |
| [`air.admin.toml`](air.admin.toml) | `air` hot-reload config for the admin API (`cmd/admin`). |
| [`otel-collector.yaml`](otel-collector.yaml) | OTLP collector pipeline — receives on `:4318`, fans out to Jaeger, Loki, Prometheus. |
| [`prometheus.yaml`](prometheus.yaml) | Prometheus scrape config. |
| [`grafana/provisioning/datasources/datasources.yaml`](grafana/provisioning/datasources/datasources.yaml) | Auto-provisions Grafana's datasources: Prometheus, Loki (`:3100`), Jaeger (`:16686`). |

## Compose profiles

The stack is split by Docker Compose profile so you start only what you need. The
profiles are **`api`**, **`admin`**, **`seed`**, and **`observability`**. Each app
profile brings up its own dependencies (postgres, redis, migrate); the telemetry
stack is a separate profile. (`cmd/mesh` is **not** wired into compose — it needs a
cert keychain first.)

```bash
make dev-api            # public API (:8080) + postgres/redis/migrate   (profile: api)
make dev-admin          # admin API (:8081) + postgres/redis/migrate    (profile: admin)
make dev-observability  # grafana / jaeger / prometheus / loki / otel-collector (profile: observability)
make seed               # fill the dev database with fake data          (profile: seed)
make dev-logs           # tail running containers
make dev-down           # stop everything
make dev-reset          # stop + remove volumes
```

Combine profiles directly when you want telemetry alongside an app:

```bash
docker compose --profile api --profile observability up -d
```

The apps export OTLP to the collector but don't depend on it — telemetry is
best-effort, so an app runs fine without the observability profile.

## Ports

Host ports, as published by `docker-compose.yml`:

| Service | Host port | Notes |
|---------|-----------|-------|
| API | 8080 | Public API HTTP |
| Admin | 8081 | Admin API HTTP |
| PostgreSQL | 5432 | |
| Redis | 6379 | |
| OTEL Collector | 4318 | OTLP HTTP receiver |
| Jaeger | 16686 | Trace UI |
| Loki | 3100 | Log aggregation |
| Prometheus | **9091** | Compose maps `9091:9090` — the container listens on 9090, but you reach it on **9091** |
| Grafana | 3000 | Unified UI |

## Viewing telemetry

- **Traces** — Jaeger at http://localhost:16686
- **Metrics** — Prometheus at http://localhost:9091
- **Logs** — via Grafana, or query Loki at `:3100`
- **Everything** — Grafana at http://localhost:3000 (Prometheus, Loki, and Jaeger
  are provisioned as datasources)

`make dev-observability` prints the Grafana, Jaeger, and Prometheus URLs — including
the `9091` Prometheus port — when it finishes.
