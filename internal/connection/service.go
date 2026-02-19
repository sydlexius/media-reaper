package connection

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/sydlexius/media-reaper/internal/repository"
)

// Service provides business logic for connection management.
type Service struct {
	repo      repository.ConnectionRepository
	encryptor *Encryptor
}

// NewService creates a connection service.
func NewService(repo repository.ConnectionRepository, encryptor *Encryptor) *Service {
	return &Service{repo: repo, encryptor: encryptor}
}

// Create creates a new connection with an encrypted API key.
func (s *Service) Create(ctx context.Context, name, connType, url, apiKey string) (*repository.Connection, error) {
	encrypted, err := s.encryptor.Encrypt(apiKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting api key: %w", err)
	}

	conn := &repository.Connection{
		ID:              uuid.New().String(),
		Name:            name,
		Type:            repository.ConnectionType(connType),
		URL:             strings.TrimRight(url, "/"),
		EncryptedAPIKey: encrypted,
		Enabled:         true,
		Status:          repository.ConnectionStatusUnknown,
	}

	if err := s.repo.Create(ctx, conn); err != nil {
		return nil, fmt.Errorf("creating connection: %w", err)
	}

	return conn, nil
}

// GetAll returns all connections with masked API keys.
func (s *Service) GetAll(ctx context.Context) ([]*repository.Connection, error) {
	return s.repo.GetAll(ctx)
}

// GetByID returns a single connection by ID.
func (s *Service) GetByID(ctx context.Context, id string) (*repository.Connection, error) {
	return s.repo.GetByID(ctx, id)
}

// Update updates a connection. If apiKey is non-empty, re-encrypts it.
func (s *Service) Update(ctx context.Context, id, name, connType, url, apiKey string, enabled *bool) (*repository.Connection, error) {
	conn, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching connection: %w", err)
	}
	if conn == nil {
		return nil, nil
	}

	conn.Name = name
	conn.Type = repository.ConnectionType(connType)
	conn.URL = strings.TrimRight(url, "/")

	if apiKey != "" {
		encrypted, err := s.encryptor.Encrypt(apiKey)
		if err != nil {
			return nil, fmt.Errorf("encrypting api key: %w", err)
		}
		conn.EncryptedAPIKey = encrypted
	}

	if enabled != nil {
		conn.Enabled = *enabled
	}

	if err := s.repo.Update(ctx, conn); err != nil {
		return nil, fmt.Errorf("updating connection: %w", err)
	}

	return conn, nil
}

// Delete deletes a connection by ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// DecryptAPIKey decrypts an encrypted API key.
func (s *Service) DecryptAPIKey(encrypted string) (string, error) {
	return s.encryptor.Decrypt(encrypted)
}

// MaskAPIKey returns a masked version of a decrypted API key showing only the last 4 characters.
func MaskAPIKey(encrypted string, encryptor *Encryptor) string {
	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil || len(decrypted) < 4 {
		return "****"
	}
	return strings.Repeat("*", len(decrypted)-4) + decrypted[len(decrypted)-4:]
}
