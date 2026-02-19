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
