package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/danielkov/gin-helmet/ginhelmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/giorgiodots/dots-beacon/api/internal/auth"
	"github.com/giorgiodots/dots-beacon/api/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Feature registers its operations against the shared huma API, which derives
// the OpenAPI spec from the handlers' typed inputs and outputs.
type Feature interface {
	RegisterRoutes(api huma.API)
}

// SecurityScheme is the name of the bearer-token security scheme. Operations
// that opt into authentication reference it in their Security definition, e.g.
//
//	Security: []map[string][]string{{server.SecurityScheme: {}}}
const SecurityScheme = "bearer"

type Server struct {
	httpServer *http.Server
}

func New(cfg config.Config, logger zerolog.Logger, auth *auth.Authenticator, features ...Feature) *Server {
	if !cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	ngin := gin.New()
	ngin.Use(gin.Recovery())
	ngin.Use(ginhelmet.Default())
	ngin.Use(corsMiddleware(cfg))
	// Registered before any routes so it wraps every feature operation; the
	// huma handlers read the per-request logger off the context.
	ngin.Use(requestLogger(logger))

	ngin.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := humagin.New(ngin, openAPIConfig())
	api.UseMiddleware(authMiddleware(api, auth))

	for _, f := range features {
		f.RegisterRoutes(api)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.HttpPort,
			Handler: ngin,
		},
	}
}

func (s *Server) Run(ctx context.Context) error {
	return s.httpServer.ListenAndServe()
}

// corsMiddleware builds the CORS handler from config. We use bearer tokens (not
// cookies), so credentials are off and the Authorization header is allowed for
// preflight. AllowedOrigins == ["*"] allows any origin; otherwise it's an
// explicit allow-list.
func corsMiddleware(cfg config.Config) gin.HandlerFunc {
	c := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
		c.AllowAllOrigins = true
	} else {
		c.AllowOrigins = cfg.AllowedOrigins
	}
	return cors.New(c)
}

func requestLogger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/healthz" {
			c.Next()
			return
		}

		start := time.Now()
		requestID := uuid.New().String()

		reqLogger := logger.With().Str("request_id", requestID).Logger()
		c.Request = c.Request.WithContext(reqLogger.WithContext(c.Request.Context()))
		c.Header("X-Request-ID", requestID)

		c.Next()

		status := c.Writer.Status()
		ev := reqLogger.Info()
		if status > 500 {
			ev = reqLogger.Error()
		} else if status >= 400 {
			ev = reqLogger.Warn()
		}
		ev.Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Dur("latency_ms", time.Since(start)).
			Str("ip", c.ClientIP()).
			Msg("request")
	}
}

// openAPIConfig describes the API and registers the bearer security scheme so
// operations can reference it and the docs render an "Authorize" prompt. huma
// serves the spec at /openapi.json (+ .yaml) and interactive docs at /docs.
func openAPIConfig() huma.Config {
	cfg := huma.DefaultConfig("dots-beacon API", "1.0.0")
	cfg.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		SecurityScheme: {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}
	return cfg
}

// authMiddleware verifies the bearer token for operations that declare the
// bearer security scheme, then stashes the user id on the request context.
// Operations without a security requirement pass through untouched.
func authMiddleware(api huma.API, v *auth.Authenticator) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !requiresAuth(ctx.Operation()) {
			next(ctx)
			return
		}

		raw := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		result, err := v.Verify(ctx.Context(), raw)
		if err != nil {
			ctx.SetHeader("WWW-Authenticate", `Bearer error="invalid_token"`)
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized, "not authorized")
			return
		}

		next(huma.WithValue(ctx, userIDKey, result.Sub))
	}
}

func requiresAuth(op *huma.Operation) bool {
	for _, scheme := range op.Security {
		if _, ok := scheme[SecurityScheme]; ok {
			return true
		}
	}
	return false
}
