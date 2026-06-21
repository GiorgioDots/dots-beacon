// Package telemetry centralises all OpenTelemetry + zerolog wiring so every
// service/worker gets traces, metrics and logs by calling Init once. All
// connection details are read from the standard OTEL_* environment variables
// (OTEL_SERVICE_NAME, OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_RESOURCE_ATTRIBUTES),
// so adding a new service means setting those vars — not editing this code.
package telemetry

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// instrumentationScope identifies meters/loggers created by this package.
const instrumentationScope = "github.com/giorgio-dots/dots-beacon-internal/telemetry"

// serviceName is captured at Init time for use by the gin middleware/logger.
var serviceName = "unknown_service"

// Init bootstraps the global tracer, meter and logger providers (all exporting
// via OTLP/gRPC to the collector) and configures zerolog with an OTel bridge.
// It returns a shutdown func that flushes and closes every provider; call it
// (with a bounded context) on graceful exit.
func Init(ctx context.Context) (func(context.Context) error, error) {
	if n := os.Getenv("OTEL_SERVICE_NAME"); n != "" {
		serviceName = n
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),       // OTEL_SERVICE_NAME, OTEL_RESOURCE_ATTRIBUTES
		resource.WithTelemetrySDK(),  // telemetry.sdk.*
		resource.WithProcess(),       // process.*
		resource.WithHost(),          // host.*
	)
	if err != nil {
		return nil, err
	}

	// W3C trace context + baggage so spans propagate across services.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var shutdowns []func(context.Context) error

	// Traces.
	traceExp, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	shutdowns = append(shutdowns, tp.Shutdown)

	// Metrics.
	metricExp, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return nil, errors.Join(err, shutdownAll(ctx, shutdowns))
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	shutdowns = append(shutdowns, mp.Shutdown)

	// Logs.
	logExp, err := otlploggrpc.New(ctx)
	if err != nil {
		return nil, errors.Join(err, shutdownAll(ctx, shutdowns))
	}
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)
	logglobal.SetLoggerProvider(lp)
	shutdowns = append(shutdowns, lp.Shutdown)

	// zerolog now that the logger provider exists.
	initLogger()

	return func(ctx context.Context) error {
		return shutdownAll(ctx, shutdowns)
	}, nil
}

// shutdownAll runs shutdown funcs in reverse (LIFO) order, joining errors.
func shutdownAll(ctx context.Context, fns []func(context.Context) error) error {
	var err error
	for i := len(fns) - 1; i >= 0; i-- {
		err = errors.Join(err, fns[i](ctx))
	}
	return err
}
