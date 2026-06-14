// Package server builds the HTTP engine, wires shared middleware and base
// routes, mounts the feature handlers, and owns the server lifecycle.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/giorgio-dots/dots-beacon-api/internal/config"
	"github.com/giorgio-dots/dots-beacon-internal/auth"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
)

// Feature is anything that can mount its own routes — i.e. a domain handler
// such as *site.Handler. The server stays decoupled from individual features.
type Feature interface {
	RegisterRoutes(r gin.IRouter)
}

// Server owns the gin engine and the underlying HTTP server.
type Server struct {
	httpServer *http.Server
}

// New assembles the engine: recovery + telemetry middleware, a public health
// check, and an authenticated group under which every feature is mounted. Pass
// authenticator == nil to leave the group unauthenticated (dev without Keycloak).
func New(cfg config.Config, authenticator *auth.Authenticator, features ...Feature) *Server {
	if !cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	telemetry.InstrumentGin(engine) // traces + HTTP metrics for every route

	// Public routes.
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Authenticated routes: features and /me live here.
	api := engine.Group("/")
	if authenticator != nil {
		api.Use(authenticator.Middleware())
	}
	api.GET("/me", currentUser)
	for _, f := range features {
		f.RegisterRoutes(api)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.HttpPort,
			Handler: engine,
		},
	}
}

// Run starts serving and blocks until ctx is cancelled, then shuts down
// gracefully (draining in-flight requests).
func (s *Server) Run(ctx context.Context) error {
	go func() {
		telemetry.Log().Info().Str("addr", s.httpServer.Addr).Msg("api listening")
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			telemetry.Log().Fatal().Err(err).Msg("http server failed")
		}
	}()

	<-ctx.Done()
	telemetry.Log().Info().Msg("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(shutdownCtx)
}

// currentUser returns the authenticated caller's identity.
func currentUser(c *gin.Context) {
	user, ok := auth.UserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}
