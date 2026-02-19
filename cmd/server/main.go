package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sydlexius/media-reaper/internal/auth"
	"github.com/sydlexius/media-reaper/internal/config"
	"github.com/sydlexius/media-reaper/internal/connection"
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

	// Encryption
	encryptor, err := connection.NewEncryptor(cfg.MasterKey)
	if err != nil {
		return fmt.Errorf("failed to initialize encryption: %w", err)
	}

	// Repositories
	userRepo := sqliterepo.NewUserRepository(database)
	connRepo := sqliterepo.NewConnectionRepository(database)

	// Services
	authService := auth.NewService(userRepo, cfg)
	connService := connection.NewService(connRepo, encryptor)

	if err := authService.Bootstrap(context.Background()); err != nil {
		return fmt.Errorf("failed to bootstrap admin user: %w", err)
	}

	// Health checker
	healthChecker := connection.NewHealthChecker(connRepo, encryptor, cfg.HealthCheckInterval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go healthChecker.Start(ctx)

	srv := server.New(cfg, authService, connService)
	log.Printf("Starting media-reaper on port %d", cfg.Port)

	// Graceful shutdown on interrupt
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
