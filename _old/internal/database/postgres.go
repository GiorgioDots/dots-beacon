// Package database provides a shared, observability-instrumented Postgres pool.
// Every service builds its pool through NewPool so DB tracing and pool metrics
// are automatic — no per-service wiring.
package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// sqlVerb extracts the SQL verb (SELECT/INSERT/...) from a statement, skipping
// the leading "-- name: X" comment sqlc prepends. Used as the span name to keep
// query span names low-cardinality and readable.
func sqlVerb(stmt string) string {
	for _, line := range strings.Split(stmt, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		if i := strings.IndexAny(line, " \t("); i > 0 {
			return strings.ToUpper(line[:i])
		}
		return strings.ToUpper(line)
	}
	return "query"
}

// NewPool builds a pgxpool whose queries each emit an OTel span (a child of the
// span in the caller's context, so DB ops nest under the originating request)
// and whose pool statistics are exported as OTel metrics. It relies on the
// global tracer/meter providers, so call telemetry.Init before this.
func NewPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// Per-query spans named by SQL verb (e.g. "query SELECT") to keep span names
	// low-cardinality; the full statement still rides along as a span attribute.
	// WithTrimSQLInSpanName must be set for the custom name func to be used.
	cfg.ConnConfig.Tracer = otelpgx.NewTracer(
		otelpgx.WithTrimSQLInSpanName(),
		otelpgx.WithSpanNameFunc(sqlVerb),
	)

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	// pgxpool stats (acquired/idle/total conns, acquire durations, …) as metrics.
	if err := otelpgx.RecordStats(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("record db stats: %w", err)
	}

	return pool, nil
}
