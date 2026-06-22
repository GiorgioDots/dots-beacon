package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/giorgiodots/dots-beacon/api/internal/auth"
	"github.com/giorgiodots/dots-beacon/api/internal/config"
	"github.com/giorgiodots/dots-beacon/api/internal/server"
	"github.com/giorgiodots/dots-beacon/api/internal/sites"
	"github.com/giorgiodots/dots-beacon/package/database"
	"github.com/giorgiodots/dots-beacon/package/database/db"
	"github.com/rs/zerolog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("An error occured while loading configuration: %v\n", err)
		return
	}

	// Log setup
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	var logger zerolog.Logger

	if cfg.IsDev() {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		logger = zerolog.New(os.Stderr).With().Timestamp().Str("service", "api").Str("env", cfg.AppEnv).Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	dbpool, err := database.NewPool(ctx, cfg.DatabaseUrl)
	if err != nil {
		fmt.Printf("An error occured while connecting to the database: %v\n", err)
		return
	}
	defer dbpool.Close()

	queries := db.New(dbpool)

	auth, err := auth.NewAuthVerifier(ctx, cfg)
	if err != nil {
		fmt.Printf("An error occured during auth initialization: %v\n", err)
		return
	}

	site := sites.NewHandler(sites.NewService(queries))
	srv := server.New(cfg, logger, auth, site)

	if err := srv.Run(ctx); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
