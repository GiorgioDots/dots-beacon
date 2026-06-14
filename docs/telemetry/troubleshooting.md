# Telemetry troubleshooting

Common issues, the symptom you'll see, and the fix. Several of these are pitfalls
that were hit while building the stack — documented so you don't re-discover them.

## "variable is not set" warnings / broken credentials

**Symptom:** running `docker compose …` prints
`The "MINIO_ROOT_USER" variable is not set. Defaulting to a blank string.` and
MinIO/Tempo/Loki/Mimir fail to authenticate.

**Cause:** the compose file interpolates `${...}` from an env file, and bare
`docker compose` only reads `.env` by default — not `.env.local`.

**Fix:** always pass `--env-file .env.local` (or use `task`, which loads it):

```bash
docker compose --env-file .env.local up -d --wait
```

## Trace/metric storm; wrong/duplicate spans from infra

**Symptom:** the collector logs huge span batches (many `resource spans`), and
infra containers (Tempo/Loki) try to export traces to `localhost:4317`
("malformed HTTP response" errors).

**Cause:** giving an infra service the **app's** `OTEL_*` env vars (e.g. via
`env_file: .env.local`) makes its embedded OTel SDK self-instrument toward
`OTEL_EXPORTER_OTLP_ENDPOINT` — a feedback loop.

**Fix:** infra services get **only** the vars they need, via `environment:` with
explicit interpolation — never the app's `OTEL_*`. See the `tempo`/`loki`/
`mimir`/`otel-collector` services in
[`docker-compose.yml`](../../docker-compose.yml).

## Grafana won't start: "data source with the same uid already exists"

**Symptom:** `grafana` container exits (1) after a datasource rename.

**Cause:** datasources are provisioned from
[`datasources.yml`](../../observability/grafana/provisioning/datasources/datasources.yml),
but the `grafana-data` volume also persists previously-provisioned ones. Renaming
a datasource while reusing its `uid` collides with the stored copy.

**Fix:** reset Grafana's state (it's fully provisioned-from-code, so nothing is
lost):

```bash
docker volume rm dots-beacon_grafana-data
docker compose --env-file .env.local up -d grafana
```

## Orphaned container after removing a service

**Symptom:** a removed service (e.g. the old `prometheus`) keeps running and
holds its port.

**Cause:** `docker compose down` only manages services still in the compose file.

**Fix:**

```bash
docker compose down --remove-orphans
```

## No data in MinIO yet (but queries work)

**Symptom:** Grafana shows traces/metrics/logs, but the MinIO buckets look empty.

**Cause:** recent data is still in the ingester/head; it becomes an **object**
only after a flush. This is expected — see [storage.md](storage.md).

**Fix (to verify now):** force a flush:

```bash
curl -XPOST 'http://localhost:9009/ingester/flush?wait=true'   # Mimir
curl -XGET   http://localhost:3200/flush                       # Tempo
curl -XPOST  http://localhost:3100/flush                       # Loki
```

## A signal is missing — check it hop by hop

1. **App is exporting?** App logs show requests; `OTEL_EXPORTER_OTLP_ENDPOINT`
   reachable (`http://localhost:4317` from host).
2. **Collector received & forwarded?** It logs each batch via the `debug`
   exporter. Check counters on its self-metrics:
   ```bash
   docker run --rm --network dots-beacon_default curlimages/curl -s \
     http://otel-collector:8888/metrics | grep -E 'otelcol_(receiver_accepted|exporter_sent)_'
   ```
3. **Backend received?**
   - Tempo: `curl 'http://localhost:3200/api/search?q=%7B%7D'`
   - Loki: `curl -G http://localhost:3100/loki/api/v1/query_range --data-urlencode 'query={service_name="api"}'`
   - Mimir: `curl -G http://localhost:9009/prometheus/api/v1/query --data-urlencode 'query=up'`
4. **Grafana → backend?** Datasource health:
   ```bash
   curl http://localhost:3000/api/datasources/uid/prometheus/health
   ```

> When searching Tempo, pass a real time range (`start`/`end`); a query with
> `range_seconds=0` returns nothing.

## Metrics naming

App/DB metrics are prefixed `dots_beacon_` (the collector's
`prometheusremotewrite` `namespace`). Collector self-metrics scraped into the
pipeline are likewise prefixed, e.g. `dots_beacon_otelcol_*`. OTel histograms
become `_bucket`/`_sum`/`_count` series in Mimir.

## Useful endpoints

| URL | What |
|-----|------|
| http://localhost:3000 | Grafana (anonymous admin) |
| http://localhost:9001 | MinIO console |
| http://localhost:3200 | Tempo API |
| http://localhost:3100 | Loki API |
| http://localhost:9009 | Mimir (remote-write `/api/v1/push`, query `/prometheus`) |
| http://localhost:13133 | Collector health check |
