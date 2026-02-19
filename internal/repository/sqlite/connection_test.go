package sqlite

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/sydlexius/media-reaper/internal/repository"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		t.Fatalf("enabling foreign keys: %v", err)
	}

	schema := `
	CREATE TABLE connections (
		id              TEXT PRIMARY KEY,
		name            TEXT NOT NULL,
		type            TEXT NOT NULL CHECK(type IN ('sonarr', 'radarr', 'emby')),
		url             TEXT NOT NULL,
		encrypted_api_key TEXT NOT NULL,
		enabled         INTEGER NOT NULL DEFAULT 1,
		status          TEXT NOT NULL DEFAULT 'unknown' CHECK(status IN ('healthy', 'unhealthy', 'unknown')),
		last_checked_at TIMESTAMP,
		created_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX idx_connections_type ON connections(type);
	CREATE INDEX idx_connections_enabled ON connections(enabled);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("creating schema: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}

func testConnection() *repository.Connection {
	return &repository.Connection{
		ID:              "test-id-1",
		Name:            "Test Sonarr",
		Type:            repository.ConnectionTypeSonarr,
		URL:             "http://localhost:8989",
		EncryptedAPIKey: "encrypted-key-data",
		Enabled:         true,
		Status:          repository.ConnectionStatusUnknown,
	}
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	conn := testConnection()
	if err := repo.Create(ctx, conn); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("GetByID after create: %v", err)
	}
	if got == nil {
		t.Fatal("expected connection, got nil")
	}
	if got.Name != conn.Name {
		t.Errorf("name: got %q, want %q", got.Name, conn.Name)
	}
	if got.Type != conn.Type {
		t.Errorf("type: got %q, want %q", got.Type, conn.Type)
	}
	if got.URL != conn.URL {
		t.Errorf("url: got %q, want %q", got.URL, conn.URL)
	}
	if got.EncryptedAPIKey != conn.EncryptedAPIKey {
		t.Errorf("encrypted key: got %q, want %q", got.EncryptedAPIKey, conn.EncryptedAPIKey)
	}
	if !got.Enabled {
		t.Error("expected enabled=true")
	}
	if got.Status != repository.ConnectionStatusUnknown {
		t.Errorf("status: got %q, want %q", got.Status, repository.ConnectionStatusUnknown)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)

	got, err := repo.GetByID(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestGetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	// Empty list
	conns, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll empty: %v", err)
	}
	if len(conns) != 0 {
		t.Errorf("expected 0 connections, got %d", len(conns))
	}

	// Add two connections
	c1 := testConnection()
	c2 := &repository.Connection{
		ID:              "test-id-2",
		Name:            "Test Radarr",
		Type:            repository.ConnectionTypeRadarr,
		URL:             "http://localhost:7878",
		EncryptedAPIKey: "encrypted-key-2",
		Enabled:         true,
		Status:          repository.ConnectionStatusUnknown,
	}
	if err := repo.Create(ctx, c1); err != nil {
		t.Fatalf("Create c1: %v", err)
	}
	if err := repo.Create(ctx, c2); err != nil {
		t.Fatalf("Create c2: %v", err)
	}

	conns, err = repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	if len(conns) != 2 {
		t.Errorf("expected 2 connections, got %d", len(conns))
	}
}

func TestGetAllEnabled(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	enabled := testConnection()
	disabled := &repository.Connection{
		ID:              "test-id-disabled",
		Name:            "Disabled Connection",
		Type:            repository.ConnectionTypeEmby,
		URL:             "http://localhost:8096",
		EncryptedAPIKey: "encrypted-key-disabled",
		Enabled:         false,
		Status:          repository.ConnectionStatusUnknown,
	}

	if err := repo.Create(ctx, enabled); err != nil {
		t.Fatalf("Create enabled: %v", err)
	}
	if err := repo.Create(ctx, disabled); err != nil {
		t.Fatalf("Create disabled: %v", err)
	}

	conns, err := repo.GetAllEnabled(ctx)
	if err != nil {
		t.Fatalf("GetAllEnabled: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("expected 1 enabled connection, got %d", len(conns))
	}
	if conns[0].ID != enabled.ID {
		t.Errorf("expected enabled connection ID %q, got %q", enabled.ID, conns[0].ID)
	}
}

func TestUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	conn := testConnection()
	if err := repo.Create(ctx, conn); err != nil {
		t.Fatalf("Create: %v", err)
	}

	conn.Name = "Updated Name"
	conn.URL = "http://localhost:9999"
	conn.EncryptedAPIKey = "new-encrypted-key"
	conn.Enabled = false

	if err := repo.Update(ctx, conn); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Name != "Updated Name" {
		t.Errorf("name: got %q, want %q", got.Name, "Updated Name")
	}
	if got.URL != "http://localhost:9999" {
		t.Errorf("url: got %q, want %q", got.URL, "http://localhost:9999")
	}
	if got.EncryptedAPIKey != "new-encrypted-key" {
		t.Errorf("key: got %q, want %q", got.EncryptedAPIKey, "new-encrypted-key")
	}
	if got.Enabled {
		t.Error("expected enabled=false after update")
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	conn := testConnection()
	if err := repo.Create(ctx, conn); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, conn.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := repo.GetByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("GetByID after delete: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestUpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	conn := testConnection()
	if err := repo.Create(ctx, conn); err != nil {
		t.Fatalf("Create: %v", err)
	}

	checkedAt := "2025-01-15T10:30:00Z"
	if err := repo.UpdateStatus(ctx, conn.ID, repository.ConnectionStatusHealthy, checkedAt); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.GetByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("GetByID after status update: %v", err)
	}
	if got.Status != repository.ConnectionStatusHealthy {
		t.Errorf("status: got %q, want %q", got.Status, repository.ConnectionStatusHealthy)
	}
	if got.LastCheckedAt == nil {
		t.Fatal("expected last_checked_at to be set")
	}
	if *got.LastCheckedAt != checkedAt {
		t.Errorf("last_checked_at: got %q, want %q", *got.LastCheckedAt, checkedAt)
	}
}

func TestLastCheckedAtNullable(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)
	ctx := context.Background()

	conn := testConnection()
	if err := repo.Create(ctx, conn); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, conn.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.LastCheckedAt != nil {
		t.Errorf("expected nil last_checked_at on new connection, got %q", *got.LastCheckedAt)
	}
}

func TestTypeConstraint(t *testing.T) {
	db := setupTestDB(t)
	repo := NewConnectionRepository(db)

	conn := &repository.Connection{
		ID:              "bad-type",
		Name:            "Bad Type",
		Type:            "invalid",
		URL:             "http://localhost:1234",
		EncryptedAPIKey: "key",
		Enabled:         true,
		Status:          repository.ConnectionStatusUnknown,
	}

	err := repo.Create(context.Background(), conn)
	if err == nil {
		t.Error("expected error for invalid connection type")
	}
}
