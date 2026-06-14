package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gin-gonic/gin"
	"github.com/giorgio-dots/dots-beacon-api/config"
	"github.com/giorgio-dots/dots-beacon-internal/telemetry"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := env.ParseAs[config.AppConfig]()
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("failed to parse environment config")
	}

	shutdown, err := telemetry.Init(ctx)
	if err != nil {
		telemetry.Log().Fatal().Err(err).Msg("failed to initialise telemetry")
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(shutdownCtx); err != nil {
			telemetry.Log().Error().Err(err).Msg("telemetry shutdown error")
		}
	}()

	if cfg.AppEnv != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	telemetry.InstrumentGin(r)

	r.GET("/healthz", func(c *gin.Context) {
		telemetry.Log().Info().Ctx(c.Request.Context()).Msg("health check")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:    ":" + cfg.HttpPort,
		Handler: r,
	}

	go func() {
		telemetry.Log().Info().Str("addr", srv.Addr).Msg("api listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			telemetry.Log().Fatal().Err(err).Msg("http server failed")
		}
	}()

	<-ctx.Done()
	stop()
	telemetry.Log().Info().Msg("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		telemetry.Log().Error().Err(err).Msg("http shutdown error")
	}
}
