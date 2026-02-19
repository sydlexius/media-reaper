package main

import (
	"context"
	"fmt"
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
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.Load()

	database, err := db.New(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer func() { _ = database.Close() }()

	userRepo := sqliterepo.NewUserRepository(database)
	authService := auth.NewService(userRepo, cfg)

	if err := authService.Bootstrap(context.Background()); err != nil {
		return fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	srv := server.New(cfg, authService)
	log.Printf("Starting media-reaper on port %d", cfg.Port)
	if err := srv.Start(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
