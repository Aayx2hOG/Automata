package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type pgUserRepository struct {
	pool *pgxpool.Pool
}

func (r *pgUserRepository) Create(ctx context.Context, u *models.User) error {
	query := `
	INSERT INTO users (username, email, password_hash, role, is_active)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, u.Username, u.Email, u.PasswordHash, u.Role, u.IsActive).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *pgUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
	FROM users
	WHERE email = $1
	`
	u := models.User{}
	err := r.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrorUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	return &u, nil
}

func (r *pgUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
	FROM users
	WHERE username = $1
	`
	u := models.User{}
	err := r.pool.QueryRow(ctx, query, username).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrorUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by username: %w", err)
	}
	return &u, nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
	SELECT id, username, email, password_hash, role, is_active, created_at, updated_at
	FROM users
	WHERE id = $1
	`
	u := models.User{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrorUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}

	return &u, nil
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &pgUserRepository{pool: pool}
}
