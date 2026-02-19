package main

import (
	"log"

	"github.com/sydlexius/media-reaper/internal/config"
	"github.com/sydlexius/media-reaper/internal/db"
	"github.com/sydlexius/media-reaper/internal/server"
)

func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	srv := server.New(cfg)
	log.Printf("Starting media-reaper on port %d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
