package telemetry

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// InstrumentGin registers tracing (otelgin) and HTTP metrics middleware on the
// router. Every service gets identical instrumentation by calling this once.
func InstrumentGin(r gin.IRouter) {
	r.Use(otelgin.Middleware(serviceName))
	r.Use(metricsMiddleware())
}

// metricsMiddleware records inbound HTTP request duration (an OTel histogram,
// whose count doubles as the request total) labelled by method, route template
// and status. Using c.FullPath() keeps the route label low-cardinality.
func metricsMiddleware() gin.HandlerFunc {
	meter := otel.Meter(instrumentationScope)
	duration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of inbound HTTP requests."),
	)
	if err != nil {
		// A broken instrument shouldn't take the service down; skip metrics.
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		duration.Record(c.Request.Context(), time.Since(start).Seconds(),
			metric.WithAttributes(
				attribute.String("http.request.method", c.Request.Method),
				attribute.String("http.route", route),
				attribute.Int("http.response.status_code", c.Writer.Status()),
			),
		)
	}
}
