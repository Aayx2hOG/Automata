package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository interface {
	Rotate(ctx context.Context, oldHash, newHash string, userID uuid.UUID, expiresAt time.Time) error
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type pgRefreshTokenRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshTokenRepository(pool *pgxpool.Pool) RefreshTokenRepository {
	return &pgRefreshTokenRepository{pool: pool}
}

func (r *pgRefreshTokenRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.pool.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}
	return nil
}

func (r *pgRefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`
	var t RefreshToken
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrorInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("query refresh token: %w", err)
	}
	return &t, nil
}

func (r *pgRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (r *pgRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens for user: %w", err)
	}
	return nil
}

func (r *pgRefreshTokenRepository) Rotate(ctx context.Context, oldHash, newHash string, userID uuid.UUID, expiresAt time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	var active bool
	if err = tx.QueryRow(ctx, `SELECT is_active FROM users WHERE id=$1 FOR SHARE`, userID).Scan(&active); err != nil {
		return err
	}
	if !active {
		return models.ErrorInvalidToken
	}
	tag, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=clock_timestamp() WHERE token_hash=$1 AND user_id=$2 AND revoked_at IS NULL AND expires_at>clock_timestamp()`, oldHash, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return models.ErrorInvalidToken
	}
	if _, err = tx.Exec(ctx, `INSERT INTO refresh_tokens(user_id,token_hash,expires_at) VALUES($1,$2,$3)`, userID, newHash, expiresAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
