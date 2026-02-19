package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sydlexius/media-reaper/internal/repository"
)

type ConnectionRepository struct {
	db *sql.DB
}

func NewConnectionRepository(db *sql.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

func (r *ConnectionRepository) Create(ctx context.Context, conn *repository.Connection) error {
	query := `INSERT INTO connections (id, name, type, url, encrypted_api_key, enabled, status, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, query,
		conn.ID, conn.Name, string(conn.Type), conn.URL, conn.EncryptedAPIKey,
		boolToInt(conn.Enabled), string(conn.Status),
	)
	if err != nil {
		return fmt.Errorf("creating connection: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) GetByID(ctx context.Context, id string) (*repository.Connection, error) {
	query := `SELECT id, name, type, url, encrypted_api_key, enabled, status, last_checked_at, created_at, updated_at
	          FROM connections WHERE id = ?`
	conn, err := r.scanConnection(r.db.QueryRowContext(ctx, query, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting connection by id: %w", err)
	}
	return conn, nil
}

func (r *ConnectionRepository) GetAll(ctx context.Context) ([]*repository.Connection, error) {
	query := `SELECT id, name, type, url, encrypted_api_key, enabled, status, last_checked_at, created_at, updated_at
	          FROM connections ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing connections: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

func (r *ConnectionRepository) GetAllEnabled(ctx context.Context) ([]*repository.Connection, error) {
	query := `SELECT id, name, type, url, encrypted_api_key, enabled, status, last_checked_at, created_at, updated_at
	          FROM connections WHERE enabled = 1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("listing enabled connections: %w", err)
	}
	defer rows.Close()

	return r.scanConnections(rows)
}

func (r *ConnectionRepository) Update(ctx context.Context, conn *repository.Connection) error {
	query := `UPDATE connections
	          SET name = ?, type = ?, url = ?, encrypted_api_key = ?, enabled = ?, updated_at = CURRENT_TIMESTAMP
	          WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query,
		conn.Name, string(conn.Type), conn.URL, conn.EncryptedAPIKey,
		boolToInt(conn.Enabled), conn.ID,
	)
	if err != nil {
		return fmt.Errorf("updating connection: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM connections WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting connection: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) UpdateStatus(ctx context.Context, id string, status repository.ConnectionStatus, checkedAt string) error {
	query := `UPDATE connections SET status = ?, last_checked_at = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, string(status), checkedAt, id)
	if err != nil {
		return fmt.Errorf("updating connection status: %w", err)
	}
	return nil
}

func (r *ConnectionRepository) scanConnection(row *sql.Row) (*repository.Connection, error) {
	conn := &repository.Connection{}
	var enabled int
	var connType, status string
	var lastCheckedAt sql.NullString

	err := row.Scan(
		&conn.ID, &conn.Name, &connType, &conn.URL, &conn.EncryptedAPIKey,
		&enabled, &status, &lastCheckedAt,
		&conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	conn.Type = repository.ConnectionType(connType)
	conn.Status = repository.ConnectionStatus(status)
	conn.Enabled = enabled == 1
	if lastCheckedAt.Valid {
		conn.LastCheckedAt = &lastCheckedAt.String
	}

	return conn, nil
}

func (r *ConnectionRepository) scanConnections(rows *sql.Rows) ([]*repository.Connection, error) {
	var connections []*repository.Connection
	for rows.Next() {
		conn := &repository.Connection{}
		var enabled int
		var connType, status string
		var lastCheckedAt sql.NullString

		err := rows.Scan(
			&conn.ID, &conn.Name, &connType, &conn.URL, &conn.EncryptedAPIKey,
			&enabled, &status, &lastCheckedAt,
			&conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning connection row: %w", err)
		}

		conn.Type = repository.ConnectionType(connType)
		conn.Status = repository.ConnectionStatus(status)
		conn.Enabled = enabled == 1
		if lastCheckedAt.Valid {
			conn.LastCheckedAt = &lastCheckedAt.String
		}

		connections = append(connections, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating connection rows: %w", err)
	}

	return connections, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
