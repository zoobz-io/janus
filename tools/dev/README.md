# tools/dev

Development environment configuration.

## Files

| File | Purpose |
|------|---------|
| `Dockerfile.dev` | Development image with hot reload (air) |
| `air.api.toml` | Hot-reload config for the public API (`cmd/api`) |
| `air.admin.toml` | Hot-reload config for the admin API (`cmd/admin`) |
| `otel-collector.yaml` | OTEL Collector pipeline configuration |
| `prometheus.yaml` | Prometheus scrape configuration |
| `grafana/provisioning/` | Grafana datasource auto-configuration |

## Compose profiles

The dev environment is split by Docker Compose profile so you start only what you
need. Each app profile brings up its own dependencies (postgres, redis, migrate);
the telemetry stack is optional and separate.

```bash
make dev-api            # public API (:8080) + postgres/redis/migrate
make dev-admin          # admin API (:8081) + postgres/redis/migrate
make dev-observability  # optional: grafana/jaeger/prometheus/loki/otel-collector
make dev-logs           # tail running containers
make dev-down           # stop everything
make dev-reset          # stop + remove volumes
```

Combine profiles directly when you want telemetry alongside an app:

```bash
docker compose --profile api --profile observability up -d
```

The apps export OTLP to the collector but do not depend on it — telemetry is
best-effort, so an app runs fine without the observability profile. (`cmd/mesh`
is not yet wired into compose — it needs a cert keychain first.)

## Observability Stack

The docker-compose.yml sets up a complete observability stack:

```
┌─────────────────────────────────────────────────────────────────┐
│                         Application                              │
│                              │                                   │
│                    OTLP HTTP (port 4318)                        │
│                              ▼                                   │
│                     ┌────────────────┐                          │
│                     │ OTEL Collector │                          │
│                     └────────────────┘                          │
│                       │     │     │                             │
│          ┌────────────┘     │     └────────────┐               │
│          ▼                  ▼                  ▼               │
│    ┌──────────┐      ┌──────────┐       ┌──────────┐          │
│    │  Jaeger  │      │   Loki   │       │Prometheus│          │
│    │ (traces) │      │  (logs)  │       │(metrics) │          │
│    └──────────┘      └──────────┘       └──────────┘          │
│          │                │                  │                 │
│          └────────────────┼──────────────────┘                 │
│                           ▼                                     │
│                     ┌──────────┐                                │
│                     │ Grafana  │                                │
│                     │  (UI)    │                                │
│                     └──────────┘                                │
└─────────────────────────────────────────────────────────────────┘
```

## Ports

| Service | Port | Purpose |
|---------|------|---------|
| API | 8080 | Public API HTTP |
| Admin | 8081 | Admin API HTTP |
| PostgreSQL | 5432 | Database |
| Redis | 6379 | Cache |
| OTEL Collector | 4318 | OTLP HTTP receiver |
| Jaeger | 16686 | Trace UI |
| Loki | 3100 | Log aggregation |
| Prometheus | 9091 | Metrics UI |
| Grafana | 3000 | Unified dashboard |

## Usage

```bash
# Start the public API + its dependencies
make dev-api

# Add the optional telemetry stack
make dev-observability

# View logs / stop / reset
make dev-logs
make dev-down
make dev-reset
```

## Viewing Telemetry

- **Traces**: http://localhost:16686 (Jaeger)
- **Metrics**: http://localhost:9090 (Prometheus)
- **Logs**: Query via Grafana or Loki API
- **Unified**: http://localhost:3000 (Grafana)
