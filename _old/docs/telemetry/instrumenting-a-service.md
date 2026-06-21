# Instrumenting a service

All OpenTelemetry + zerolog wiring lives in one shared package,
[`internal/telemetry`](../../internal/telemetry/). A service gets traces,
metrics, and trace-correlated logs by calling `telemetry.Init` once and setting
two env vars. This is the DRY contract — the collector and backends never change
when you add a service.

## 1. Minimal service (HTTP API)

```go
package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Sets up global tracer/meter/logger providers (OTLP → collector) + zerolog.
	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("telemetry init failed")
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = shutdown(c) // flush all signals on exit
	}()

	r := gin.New()
	r.Use(gin.Recovery())
	telemetry.InstrumentGin(r) // otelgin traces + HTTP metrics middleware

	r.GET("/healthz", func(c *gin.Context) {
		telemetry.Log().Info().Ctx(c.Request.Context()).Msg("health check")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	_ = r.Run(":8080")
}
```

See [`apps/api/cmd/main.go`](../../apps/api/cmd/main.go) for the full version
(graceful HTTP shutdown, config parsing).

## 2. Env vars (the only per-service config)

```dotenv
OTEL_SERVICE_NAME=my-worker
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317   # or http://otel-collector:4317 in compose
APP_ENV=dev
```

These are read automatically by the OTel SDK — `Init` does not take an endpoint
argument. `OTEL_SERVICE_NAME` becomes the `service.name` resource attribute and
the value you search on in Tempo/Loki/Mimir.

## 3. The package API

| Function | Use |
|----------|-----|
| `telemetry.Init(ctx) (shutdown, error)` | Bootstrap providers + zerolog. Call once at startup; defer `shutdown`. |
| `telemetry.Log() *zerolog.Logger` | Global logger. Goes to stdout **and** Loki. |
| `telemetry.LogCtx(ctx) *zerolog.Logger` | Logger bound to a context so `trace_id`/`span_id` are attached. |
| `telemetry.InstrumentGin(r)` | Registers tracing (`otelgin`) + HTTP duration metrics on a gin router. |

### Trace-correlated logs

For a log line to carry trace/span IDs, the active span's context must reach the
logger. Either:

```go
telemetry.Log().Info().Ctx(c.Request.Context()).Msg("...")   // attach per-event
telemetry.LogCtx(ctx).Info().Msg("...")                       // logger pre-bound to ctx
```

In dev (`APP_ENV=dev`) logs are pretty-printed to the console; otherwise JSON.
Either way they are also shipped to Loki via the `otelzerolog` bridge.

## 4. Database instrumentation

Build the pool through [`internal/database`](../../internal/database/postgres.go)
instead of `pgxpool` directly. Every query then emits a span (nested under the
request span) and pool stats are exported as metrics — automatically.

```go
import (
	"github.com/giorgio-dots/dots-beacon-internal/database"
	"github.com/giorgio-dots/dots-beacon-internal/database/db" // sqlc-generated
)

pool, err := database.NewPool(ctx, cfg.DatabaseUrl) // adds otelpgx tracer + RecordStats
if err != nil { /* ... */ }
defer pool.Close()

queries := db.New(pool)
sites, err := queries.GetPlants(c.Request.Context()) // pass the request ctx → DB span nests under it
```

`NewPool` relies on the global providers, so call `telemetry.Init` first. Query
spans are named by SQL verb (e.g. `query SELECT`) to stay low-cardinality.

## 5. Worker (no HTTP)

A background worker skips `InstrumentGin` but is otherwise identical: call
`telemetry.Init`, use `telemetry.Log()`/`LogCtx`, and create spans/metrics
manually via the global providers:

```go
tracer := otel.Tracer("my-worker")
ctx, span := tracer.Start(ctx, "process-batch")
defer span.End()
```

Custom metrics use `otel.Meter("my-worker")`. Both inherit the OTLP export
configured by `Init`, so they land in Mimir/Tempo with no extra setup.

## Checklist for a new service

- [ ] `telemetry.Init(ctx)` at startup, `defer shutdown`
- [ ] `OTEL_SERVICE_NAME` + `OTEL_EXPORTER_OTLP_ENDPOINT` in its env
- [ ] (HTTP) `telemetry.InstrumentGin(r)`
- [ ] (DB) build the pool via `database.NewPool`
- [ ] Pass request/job `context.Context` through so spans nest and logs correlate
- [ ] If running in compose, set `OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317`

Nothing else — no collector, Mimir, Tempo, Loki, or Grafana edits required.
