package connection

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	sqliterepo "github.com/sydlexius/media-reaper/internal/repository/sqlite"
)

func setupTestService(t *testing.T) *Service {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

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
	)`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("creating schema: %v", err)
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generating key: %v", err)
	}
	encryptor, err := NewEncryptor(hex.EncodeToString(key))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	repo := sqliterepo.NewConnectionRepository(db)
	return NewService(repo, encryptor)
}

func newTestContext(method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

func TestCreateHandler(t *testing.T) {
	svc := setupTestService(t)

	body := `{"name":"My Sonarr","type":"sonarr","url":"http://localhost:8989","apiKey":"test-key-123"}`
	c, rec := newTestContext(http.MethodPost, "/api/connections", body)

	if err := svc.CreateHandler(c); err != nil {
		t.Fatalf("CreateHandler: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp connectionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	if resp.Name != "My Sonarr" {
		t.Errorf("name: got %q, want %q", resp.Name, "My Sonarr")
	}
	if resp.Type != "sonarr" {
		t.Errorf("type: got %q, want %q", resp.Type, "sonarr")
	}
	if resp.URL != "http://localhost:8989" {
		t.Errorf("url: got %q, want %q", resp.URL, "http://localhost:8989")
	}
	if !resp.Enabled {
		t.Error("expected enabled=true")
	}
	if resp.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateHandlerMasksAPIKey(t *testing.T) {
	svc := setupTestService(t)

	body := `{"name":"Test","type":"radarr","url":"http://localhost:7878","apiKey":"abcdef123456"}`
	c, rec := newTestContext(http.MethodPost, "/api/connections", body)

	if err := svc.CreateHandler(c); err != nil {
		t.Fatalf("CreateHandler: %v", err)
	}

	var resp connectionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}

	// Key should be masked: only last 4 chars visible
	if !strings.HasSuffix(resp.MaskedAPIKey, "3456") {
		t.Errorf("masked key should end with '3456', got %q", resp.MaskedAPIKey)
	}
	if !strings.HasPrefix(resp.MaskedAPIKey, "****") {
		t.Errorf("masked key should start with asterisks, got %q", resp.MaskedAPIKey)
	}
	// Key should NOT contain the full plaintext
	if resp.MaskedAPIKey == "abcdef123456" {
		t.Error("API key should be masked, not returned in plaintext")
	}
}

func TestCreateHandlerValidation(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		name string
		body string
	}{
		{"missing name", `{"type":"sonarr","url":"http://localhost","apiKey":"key"}`},
		{"missing type", `{"name":"Test","url":"http://localhost","apiKey":"key"}`},
		{"missing url", `{"name":"Test","type":"sonarr","apiKey":"key"}`},
		{"missing apiKey", `{"name":"Test","type":"sonarr","url":"http://localhost"}`},
		{"invalid type", `{"name":"Test","type":"invalid","url":"http://localhost","apiKey":"key"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newTestContext(http.MethodPost, "/api/connections", tt.body)
			if err := svc.CreateHandler(c); err != nil {
				t.Fatalf("CreateHandler: %v", err)
			}
			if rec.Code != http.StatusBadRequest {
				t.Errorf("status: got %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestListHandler(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// Empty list
	c, rec := newTestContext(http.MethodGet, "/api/connections", "")
	if err := svc.ListHandler(c); err != nil {
		t.Fatalf("ListHandler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	var emptyList []connectionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &emptyList); err != nil {
		t.Fatalf("decoding empty list: %v", err)
	}
	if len(emptyList) != 0 {
		t.Errorf("expected 0 connections, got %d", len(emptyList))
	}

	// Create one, then list
	if _, err := svc.Create(ctx, "Test", "sonarr", "http://localhost:8989", "key"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	c2, rec2 := newTestContext(http.MethodGet, "/api/connections", "")
	if err := svc.ListHandler(c2); err != nil {
		t.Fatalf("ListHandler: %v", err)
	}
	var list []connectionResponse
	if err := json.Unmarshal(rec2.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding list: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 connection, got %d", len(list))
	}
}

func TestGetHandler(t *testing.T) {
	svc := setupTestService(t)

	conn, err := svc.Create(context.Background(), "Test", "emby", "http://localhost:8096", "key")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, rec := newTestContext(http.MethodGet, "/api/connections/"+conn.ID, "")
	c.SetParamNames("id")
	c.SetParamValues(conn.ID)

	if err := svc.GetHandler(c); err != nil {
		t.Fatalf("GetHandler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestGetHandlerNotFound(t *testing.T) {
	svc := setupTestService(t)

	c, rec := newTestContext(http.MethodGet, "/api/connections/nonexistent", "")
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	if err := svc.GetHandler(c); err != nil {
		t.Fatalf("GetHandler: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateHandler(t *testing.T) {
	svc := setupTestService(t)

	conn, err := svc.Create(context.Background(), "Original", "sonarr", "http://localhost:8989", "original-key")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Update name and URL, omit apiKey to keep current
	body := `{"name":"Updated","type":"sonarr","url":"http://localhost:9999"}`
	c, rec := newTestContext(http.MethodPut, "/api/connections/"+conn.ID, body)
	c.SetParamNames("id")
	c.SetParamValues(conn.ID)

	if err := svc.UpdateHandler(c); err != nil {
		t.Fatalf("UpdateHandler: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusOK)
	}

	var resp connectionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Name != "Updated" {
		t.Errorf("name: got %q, want %q", resp.Name, "Updated")
	}
	if resp.URL != "http://localhost:9999" {
		t.Errorf("url: got %q, want %q", resp.URL, "http://localhost:9999")
	}
	// Masked key should still show last 4 of original key
	if !strings.HasSuffix(resp.MaskedAPIKey, "key") {
		t.Errorf("masked key should preserve original key suffix, got %q", resp.MaskedAPIKey)
	}
}

func TestUpdateHandlerNewAPIKey(t *testing.T) {
	svc := setupTestService(t)

	conn, err := svc.Create(context.Background(), "Test", "radarr", "http://localhost:7878", "old-api-key")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	body := `{"name":"Test","type":"radarr","url":"http://localhost:7878","apiKey":"brand-new-key"}`
	c, rec := newTestContext(http.MethodPut, "/api/connections/"+conn.ID, body)
	c.SetParamNames("id")
	c.SetParamValues(conn.ID)

	if err := svc.UpdateHandler(c); err != nil {
		t.Fatalf("UpdateHandler: %v", err)
	}

	var resp connectionResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)

	// Should show last 4 of new key
	if !strings.HasSuffix(resp.MaskedAPIKey, "-key") {
		t.Errorf("masked key should show new key suffix, got %q", resp.MaskedAPIKey)
	}
}

func TestDeleteHandler(t *testing.T) {
	svc := setupTestService(t)

	conn, err := svc.Create(context.Background(), "To Delete", "emby", "http://localhost:8096", "key")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	c, rec := newTestContext(http.MethodDelete, "/api/connections/"+conn.ID, "")
	c.SetParamNames("id")
	c.SetParamValues(conn.ID)

	if err := svc.DeleteHandler(c); err != nil {
		t.Fatalf("DeleteHandler: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Verify deleted
	got, err := svc.GetByID(context.Background(), conn.ID)
	if err != nil {
		t.Fatalf("GetByID after delete: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestAPIKeyEncryptedAtRest(t *testing.T) {
	svc := setupTestService(t)

	conn, err := svc.Create(context.Background(), "Test", "sonarr", "http://localhost:8989", "my-secret-api-key")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Fetch raw from repo (encrypted key)
	raw, err := svc.repo.GetByID(context.Background(), conn.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	// Encrypted key should NOT be the plaintext
	if raw.EncryptedAPIKey == "my-secret-api-key" {
		t.Error("API key stored in plaintext, expected encrypted")
	}

	// But should decrypt back to the original
	decrypted, err := svc.DecryptAPIKey(raw.EncryptedAPIKey)
	if err != nil {
		t.Fatalf("DecryptAPIKey: %v", err)
	}
	if decrypted != "my-secret-api-key" {
		t.Errorf("decrypted key: got %q, want %q", decrypted, "my-secret-api-key")
	}
}

func TestURLTrailingSlashTrimmed(t *testing.T) {
	svc := setupTestService(t)

	conn, err := svc.Create(context.Background(), "Test", "sonarr", "http://localhost:8989/", "key")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if conn.URL != "http://localhost:8989" {
		t.Errorf("URL should have trailing slash trimmed, got %q", conn.URL)
	}
}
