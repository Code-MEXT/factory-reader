package main

import (
	"context"
	"log"

	"github.com/Code-MEXT/factory-reader/config"
	"github.com/Code-MEXT/factory-reader/db"
	"github.com/Code-MEXT/factory-reader/server"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(ctx); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	srv := server.New(cfg.ServerAddr, database)
	log.Fatal(srv.Start())
}
