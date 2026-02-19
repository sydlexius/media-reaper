package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func New(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// SQLite only supports one writer at a time
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	if _, err := database.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if _, err := database.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	if err := runMigrations(database); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return database, nil
}

func runMigrations(database *sql.DB) error {
	fsys, err := fs.Sub(embedMigrations, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration sub-filesystem: %w", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectSQLite3,
		database,
		fsys,
	)
	if err != nil {
		return fmt.Errorf("creating goose provider: %w", err)
	}

	results, err := provider.Up(context.Background())
	if err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	for _, r := range results {
		log.Printf("Migration applied: %s (duration: %s)", r.Source.Path, r.Duration)
	}

	return nil
}
