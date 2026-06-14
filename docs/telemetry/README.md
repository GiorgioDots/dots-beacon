# Telemetry & Observability

dots-beacon ships with full observability from day one: **traces, metrics, and
logs** for every service, collected through a single OpenTelemetry Collector and
stored in S3-compatible object storage (MinIO). Adding a new service costs a
couple of env vars and a few lines of bootstrap — never a rewrite of the
collector or backend configuration.

- New service? → [instrumenting-a-service.md](instrumenting-a-service.md)
- Where does data live / retention? → [storage.md](storage.md)
- Something broken? → [troubleshooting.md](troubleshooting.md)

## What you get

| Signal | Produced by | Stored in | Queried via |
|--------|-------------|-----------|-------------|
| Traces | OTel SDK + `otelgin` + `otelpgx` | Tempo → MinIO (`tempo` bucket) | Grafana → Tempo |
| Metrics | OTel SDK + HTTP/DB middleware + `postgresql` receiver | Mimir → MinIO (`mimir-blocks`) | Grafana → Mimir |
| Logs | zerolog + `otelzerolog` bridge | Loki → MinIO (`loki` bucket) | Grafana → Loki |

Logs carry `trace_id`/`span_id`, so you can jump log ↔ trace in Grafana.

## Architecture

Everything from the app side leaves over **one** OTLP connection to the
collector. The collector fans out to the three backends, each of which persists
to MinIO. Grafana reads all three.

```
 Service / worker                         OTel Collector              Backends            Storage
 ─────────────────                        ──────────────              ────────            ───────
 telemetry.Init(ctx)                      receivers:
   TracerProvider ─┐                        otlp (4317/4318)          ┌─ Tempo  ──────────┐
   MeterProvider  ─┼─ OTLP/gRPC :4317 ───►   postgresql (DB metrics)  │   (traces)        │
   LoggerProvider ─┘                         prometheus (self)        │                   │
 zerolog → stdout + otelzerolog                                       ├─ Mimir  ──────────┼──► MinIO (S3)
 gin: otelgin + metrics mw               processors:                  │   (metrics, RW)   │     :9000 / :9001
 pgx: otelpgx (per-query spans)            memory_limiter             │                   │
                                           resource_detection         └─ Loki  ───────────┘
                                           batch
                                         exporters:                   Grafana :3000
                                           otlp_grpc  → Tempo            ← Tempo
                                           prometheusremotewrite→Mimir   ← Mimir (/prometheus)
                                           otlp_http  → Loki             ← Loki
```

Key design choices:

- **Single collector, shared config.** Services point
  `OTEL_EXPORTER_OTLP_ENDPOINT` at the collector; nothing in the collector,
  Tempo, Loki, Mimir, or Grafana changes when you add a service.
- **Metrics are pushed, not scraped.** Mimir is Prometheus-compatible but does
  not scrape — the collector `prometheusremotewrite`s to it. This means new
  services' metrics flow through automatically with no new scrape target.
- **Object storage for all three signals.** Tempo, Loki, and Mimir each persist
  to MinIO, so telemetry survives container restarts. See
  [storage.md](storage.md).

## Components & ports

| Service | Image | Host port | Purpose |
|---------|-------|-----------|---------|
| otel-collector | `otel/opentelemetry-collector-contrib` | 4317 (gRPC), 4318 (HTTP) | Single ingest point for all signals |
| tempo | `grafana/tempo` | 3200 | Trace store (S3-backed) |
| loki | `grafana/loki` | 3100 | Log store (S3-backed) |
| mimir | `grafana/mimir` | 9009 | Metric store (S3-backed), Prometheus-compatible |
| minio | `minio/minio` | 9000 (S3), 9001 (console) | Object storage backing the three stores |
| grafana | `grafana/grafana` | 3000 | Dashboards / Explore for all signals |
| dots-beacon-dev-db | `postgres` | 5433 → 5432 | App database (also scraped for metrics) |

Internal-only: collector self-metrics on `:8888`, health check on `:13133`.

Config files live in [`observability/`](../../observability/); service
definitions in [`docker-compose.yml`](../../docker-compose.yml).

## Running locally

The stack is wired into the Taskfile; `task dev-api` brings up the whole stack
(it depends on `dev-compose`, which waits for everything healthy) and then runs
the API:

```bash
task dev-api
```

To run just the stack:

```bash
docker compose --env-file .env.local up -d --wait
```

> **Always pass `--env-file .env.local`** when using `docker compose` directly —
> the compose file interpolates `${MINIO_ROOT_USER}` etc. from it. `task` loads
> it automatically. Without it you'll get "variable is not set" warnings and
> broken credentials. See [troubleshooting.md](troubleshooting.md).

Tear down (data survives in named volumes):

```bash
docker compose down            # add -v to also wipe MinIO/Mimir/Tempo/Loki data
```

## Accessing & querying

Open **Grafana at http://localhost:3000** (anonymous admin in dev, no login).
Use **Explore** and pick a datasource:

- **Tempo** (traces): search `{ resource.service.name="api" }`, open a trace, and
  use "Logs for this span" to jump to the correlated Loki lines.
- **Loki** (logs): `{service_name="api"}` — each line carries `trace_id`.
- **Mimir** (metrics): try
  `rate(dots_beacon_http_server_request_duration_seconds_count[1m])` or
  `dots_beacon_postgresql_commits_total`.

All app/DB metric names are prefixed `dots_beacon_`.

The **MinIO console** at http://localhost:9001 lets you browse the raw objects in
the `tempo`, `loki`, `mimir-blocks`, and `mimir-ruler` buckets.

## Configuration (env vars)

Set in [`.env.local`](../../.env.example) (see `.env.example` for the template):

| Variable | Used by | Notes |
|----------|---------|-------|
| `OTEL_SERVICE_NAME` | each service | The only per-service difference; becomes `service.name` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | each service | `http://localhost:4317` on host, `http://otel-collector:4317` in compose. `http://` ⇒ insecure gRPC |
| `APP_ENV` | app | `dev` ⇒ pretty console logs; otherwise JSON |
| `HTTP_PORT` | app | API listen port |
| `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` | minio, tempo, loki, mimir | S3 credentials |
| `TEMPO_S3_BUCKET` / `LOKI_S3_BUCKET` | tempo, loki | Bucket names (Mimir buckets are fixed: `mimir-blocks`, `mimir-ruler`) |
| `POSTGRES_USER` / `POSTGRES_PASSWORD` / `POSTGRES_DB` | db, collector | Also used by the collector's `postgresql` receiver |

> Infra services receive **only** the vars they need (not the app's `OTEL_*`),
> on purpose — see the self-tracing pitfall in
> [troubleshooting.md](troubleshooting.md).
