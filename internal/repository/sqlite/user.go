package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sydlexius/media-reaper/internal/repository"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *repository.User) error {
	query := `INSERT INTO users (id, username, password_hash, role, created_at, updated_at)
	          VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.Role)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*repository.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at
	          FROM users WHERE username = ?`
	user := &repository.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by username: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*repository.User, error) {
	query := `SELECT id, username, password_hash, role, created_at, updated_at
	          FROM users WHERE id = ?`
	user := &repository.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting users: %w", err)
	}
	return count, nil
}
