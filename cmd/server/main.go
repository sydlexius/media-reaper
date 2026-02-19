package main

import (
	"context"
	"log"

	"github.com/sydlexius/media-reaper/internal/auth"
	"github.com/sydlexius/media-reaper/internal/config"
	"github.com/sydlexius/media-reaper/internal/db"
	sqliterepo "github.com/sydlexius/media-reaper/internal/repository/sqlite"
	"github.com/sydlexius/media-reaper/internal/server"
)

// @title Media Reaper API
// @version 0.1.0
// @description API for media-reaper, a tool that integrates Sonarr/Radarr with Emby to identify watched media eligible for deletion.
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey SessionCookie
// @in cookie
// @name media-reaper-session
func main() {
	cfg := config.Load()

	database, err := db.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	userRepo := sqliterepo.NewUserRepository(database)
	authService := auth.NewService(userRepo, cfg)

	if err := authService.Bootstrap(context.Background()); err != nil {
		_ = database.Close()
		log.Fatalf("Failed to bootstrap admin user: %v", err)
	}

	srv := server.New(cfg, authService)
	log.Printf("Starting media-reaper on port %d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
