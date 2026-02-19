package main

import (
	"log"

	"github.com/sydlexius/media-reaper/internal/config"
	"github.com/sydlexius/media-reaper/internal/server"
)

func main() {
	cfg := config.Load()

	srv := server.New(cfg)
	log.Printf("Starting media-reaper on port %d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
