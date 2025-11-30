package main

import (
	"context"
	"log"

	"imgnation-backend/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadServerCfg()
	if err != nil {
		log.Fatalf("Failed to load config: %s", err.Error())
	}

	serv := NewServer(cfg)

	if err := serv.Init(ctx); err != nil {
		log.Fatal("Error trying to init the server", err)
	}
	if err := serv.Run(ctx); err != nil {
		log.Fatal("Error while running the server", err)
	}
}
