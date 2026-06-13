package main

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/giorgio-dots/dots-beacon-api/config"
	"github.com/giorgio-dots/dots-beacon-internal/database/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := env.ParseAs[config.AppConfig]()
	if err != nil {
		fmt.Printf("An error occured while parsing environment variables %v\n", err)
		return
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseUrl)

	q := db.New(pool)

	plants, err := q.GetPlants(context.Background())
	if err != nil {
		fmt.Printf("An error occured while retrieving the plants %v\n", err)
		return
	}

	fmt.Printf("Plants: %v\n", plants)
}
