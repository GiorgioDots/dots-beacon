// Command api is the dots-beacon HTTP API. main is the composition root: it
// loads config, initialises cross-cutting concerns (telemetry, db, auth), wires
// each feature's repository -> service -> handler, and starts the server.
package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/giorgio-dots/dots-beacon-api/internal/config"
	"github.com/giorgio-dots/dots-beacon-api/internal/server"
	"github.com/giorgio-dots/dots-beacon-api/internal/site"
	"github.com/giorgio-dots/dots-beacon-internal/auth"
	"github.com/giorgio-dots/dots-beacon-internal/database"
	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("failed to load config")
	}

	// Telemetry: traces, metrics, logs.
	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("failed to init telemetry")
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(c); err != nil {
			telemetry.Log().Error().Err(err).Msg("telemetry shutdown error")
		}
	}()

	// Database.
	pool, err := database.NewPool(ctx, cfg.DatabaseUrl)
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()
	queries := db.New(pool)

	// Authentication (optional: enabled only when KEYCLOAK_ISSUER_URL is set).
	authenticator := mustAuth(ctx, cfg)

	// Features. Each is wired repository -> service -> handler, then mounted.
	sites := site.NewHandler(site.NewService(site.NewRepository(queries)))

	// HTTP server.
	srv := server.New(cfg, authenticator, sites)
	if err := srv.Run(ctx); err != nil {
		telemetry.Log().Fatal().Err(err).Msg("server error")
	}
}

// mustAuth builds the Keycloak authenticator, or returns nil when auth is
// disabled. It fatals if auth is configured but cannot be initialised.
func mustAuth(ctx context.Context, cfg config.Config) *auth.Authenticator {
	authCfg := auth.Config{IssuerURL: cfg.KeycloakIssuerURL, ClientID: cfg.KeycloakClientID}
	if !authCfg.Enabled() {
		telemetry.Log().Warn().Msg("KEYCLOAK_ISSUER_URL not set — authentication disabled")
		return nil
	}
	authenticator, err := auth.New(ctx, authCfg)
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("failed to init Keycloak authenticator")
	}
	telemetry.Log().Info().Str("issuer", authCfg.IssuerURL).Msg("authentication enabled")
	return authenticator
}
