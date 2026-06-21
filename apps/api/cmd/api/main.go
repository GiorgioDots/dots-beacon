package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/giorgiodots/dots-beacon/api/internal/config"
	"github.com/giorgiodots/dots-beacon/api/internal/server"
	"github.com/giorgiodots/dots-beacon/api/internal/sites"
	"github.com/giorgiodots/dots-beacon/package/database"
	"github.com/giorgiodots/dots-beacon/package/database/db"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("An error occured while loading configuration: %v\n", err)
		return
	}

	dbpool, err := database.NewPool(ctx, cfg.DatabaseUrl)
	if err != nil {
		fmt.Printf("An error occured connecting to the database: %v\n", err)
		return
	}
	defer dbpool.Close()

	queries := db.New(dbpool)

	site := sites.NewHandler(sites.NewService(queries))
	srv := server.New(cfg, site)

	if err := srv.Run(ctx); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
