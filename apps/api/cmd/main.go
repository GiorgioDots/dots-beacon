package main

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/giorgio-dots/dots-beacon-api/config"
)

func main() {
	fmt.Println("hello worldds")

	cfg, err := env.ParseAs[config.AppConfig]()
	if err != nil {
		fmt.Printf("An error occured while parsing environment variables %v\n", err)
		return
	}

	fmt.Printf("%s\n", cfg.DatabaseUrl)
}
