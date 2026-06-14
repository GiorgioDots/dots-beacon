package telemetry

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/bridges/otelzerolog"
	logglobal "go.opentelemetry.io/otel/log/global"
)

// initLogger configures the global zerolog logger to write to stdout (pretty in
// dev, JSON otherwise) AND forward every record to the OTel logger provider via
// the otelzerolog hook, so logs reach Loki and carry trace/span IDs.
func initLogger() {
	var w io.Writer = os.Stdout
	if os.Getenv("APP_ENV") == "dev" {
		w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	hook := otelzerolog.NewHook(
		instrumentationScope,
		otelzerolog.WithLoggerProvider(logglobal.GetLoggerProvider()),
	)

	log.Logger = zerolog.New(w).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger().
		Hook(hook)
}

// Log returns the global zerolog logger. For trace-correlated logs inside a
// request, attach the request context so the OTel bridge can stamp trace_id /
// span_id, e.g. telemetry.Log().Info().Ctx(c.Request.Context()).Msg("...").
func Log() *zerolog.Logger {
	return &log.Logger
}

// LogCtx returns a logger that emits records bound to ctx, so trace/span IDs are
// attached automatically without calling .Ctx() on every event.
func LogCtx(ctx context.Context) *zerolog.Logger {
	l := log.Logger.With().Ctx(ctx).Logger()
	return &l
}
