package server

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielkov/gin-helmet/ginhelmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/giorgiodots/dots-beacon/api/internal/auth"
	"github.com/giorgiodots/dots-beacon/api/internal/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Feature interface {
	RegisterRoutes(r gin.IRouter, authMW gin.HandlerFunc)
}

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

	ngin.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authMW := newAuthMiddleware(auth)

	for _, f := range features {
		f.RegisterRoutes(ngin, authMW)
	}

	ngin.Use(requestLogger(logger))

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

func newAuthMiddleware(v *auth.Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")

		result, err := v.Verify(c.Request.Context(), raw)

		if err != nil {
			unauthorized(c, "not authorized")
			return
		}
		SetUserId(c, result.Sub)

		c.Next()
	}
}

func unauthorized(c *gin.Context, msg string) {
	c.Header("WWW-Authenticate", `Bearer error="invalid_token"`)
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
}
