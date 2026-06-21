# Telemetry storage (MinIO / S3)

All three signals persist to **MinIO**, an S3-compatible object store, so
telemetry survives container restarts. Each backend (Tempo, Loki, Mimir) writes
its own native block/chunk format into its own bucket.

## Buckets

| Bucket | Owner | Contents |
|--------|-------|----------|
| `tempo` | Tempo | Trace blocks: `data.parquet`, `bloom-*`, `index`, `meta.json` |
| `loki` | Loki | Log chunks + TSDB index (`index/…tsdb.gz`) |
| `mimir-blocks` | Mimir | Metric TSDB blocks: `chunks/*`, `index`, `meta.json` |
| `mimir-ruler` | Mimir | Recording/alerting rules (empty until rules are added) |

Buckets are created automatically at startup by the one-shot **`minio-init`**
service (using `mc`) before Tempo/Loki/Mimir start. MinIO data itself lives in
the `minio-data` Docker volume.

Browse them in the MinIO console at **http://localhost:9001** (credentials =
`MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD`), or with `mc`:

```bash
docker run --rm --network dots-beacon_default --entrypoint sh minio/mc:latest -c '
  mc alias set local http://minio:9000 "$USER" "$PASS" >/dev/null
  mc ls --recursive local/tempo
  mc du local/tempo local/loki local/mimir-blocks
'
```

## Local working state vs object storage

Each backend keeps a small local working area (in its own named volume) and
flushes durable data to MinIO:

| Backend | Local (volume) | Flushed to MinIO |
|---------|----------------|------------------|
| Tempo | WAL (`tempo-data`) | Completed trace **blocks** |
| Loki | index/wal working dir (`loki-data`) | Log **chunks** + index |
| Mimir | TSDB head + sync dirs (`mimir-data`) | Compacted metric **blocks** |

This is why a freshly-ingested trace/metric is queryable immediately (it's in the
ingester/head) but only appears as an **object** in MinIO after a flush.

## Flush & retention timing

Defaults are tuned so recent data is queryable quickly while blocks accumulate
on a sane cadence:

- **Tempo** — [`tempo.yml`](../../observability/tempo/tempo.yml) sets
  `ingester.max_block_duration: 30s` for **dev**, so blocks land in MinIO within
  ~30–60s. **Raise this (e.g. `5m`) in prod** to avoid many tiny blocks.
  Block retention is `48h` (`compactor.compaction.block_retention`).
- **Loki** — chunks flush on the normal idle/age cadence; a chunk also flushes on
  graceful shutdown. Force one with `curl -XPOST http://localhost:3100/flush`.
- **Mimir** — the ingester head compacts to a block on its own schedule
  (~2h by default), buffering in `mimir-data` until then. Force one with
  `curl -XPOST 'http://localhost:9009/ingester/flush?wait=true'`.

Forcing a flush is handy when you want to **verify** persistence without waiting.

## Wiping data

Named volumes survive `docker compose down`. To reset:

```bash
docker compose down -v          # removes ALL volumes (db, minio, mimir, tempo, loki, grafana)
# or target specific stores:
docker volume rm dots-beacon_tempo-data dots-beacon_loki-data dots-beacon_mimir-data
```

> Switching a backend's storage mode (e.g. filesystem → S3) requires wiping that
> backend's local working volume so it doesn't try to read stale state.

## Production notes

- Local dev uses MinIO over **plain HTTP** (`insecure: true`) on the compose
  network. Point the same S3 config at a real bucket + TLS endpoint for prod.
- Each backend needs distinct buckets (Mimir in particular requires separate
  blocks/ruler buckets).
- Credentials are the MinIO root user in dev; use scoped access keys in prod.
- Mimir runs in **monolithic** mode (`-target=all`, single replica,
  `replication_factor: 1`). For real workloads, run the components separately and
  raise replication.
