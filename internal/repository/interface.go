package repository

import "context"

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Role         string
	CreatedAt    string
	UpdatedAt    string
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	Count(ctx context.Context) (int, error)
}

type ConnectionType string

const (
	ConnectionTypeSonarr ConnectionType = "sonarr"
	ConnectionTypeRadarr ConnectionType = "radarr"
	ConnectionTypeEmby   ConnectionType = "emby"
)

type ConnectionStatus string

const (
	ConnectionStatusHealthy   ConnectionStatus = "healthy"
	ConnectionStatusUnhealthy ConnectionStatus = "unhealthy"
	ConnectionStatusUnknown   ConnectionStatus = "unknown"
)

type Connection struct {
	ID              string
	Name            string
	Type            ConnectionType
	URL             string
	EncryptedAPIKey string
	Enabled         bool
	Status          ConnectionStatus
	LastCheckedAt   *string
	CreatedAt       string
	UpdatedAt       string
}

type ConnectionRepository interface {
	Create(ctx context.Context, conn *Connection) error
	GetByID(ctx context.Context, id string) (*Connection, error)
	GetAll(ctx context.Context) ([]*Connection, error)
	GetAllEnabled(ctx context.Context) ([]*Connection, error)
	Update(ctx context.Context, conn *Connection) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status ConnectionStatus, checkedAt string) error
}
