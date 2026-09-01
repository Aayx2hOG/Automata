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

type WorkflowRepository interface {
	Create(ctx context.Context, w *models.Workflow) error
	GetById(ctx context.Context, id uuid.UUID) (*models.Workflow, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Workflow, error)
}

type pgWorkflowRepository struct {
	pool *pgxpool.Pool
}

func NewWorkflowRepository(pool *pgxpool.Pool) WorkflowRepository {
	return &pgWorkflowRepository{pool: pool}
}

func (r *pgWorkflowRepository) Create(ctx context.Context, w *models.Workflow) error {
	query := `
		INSERT INTO workflows (owner_id, name, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, query, w.OwnerID, w.Name, w.Description, w.IsActive).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert workflow: %w", err)
	}
	return nil
}

func (r *pgWorkflowRepository) GetById(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	query := `
SELECT id, owner_id, name, description, is_active, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`
	w := models.Workflow{}
	var description *string

	err := r.pool.QueryRow(ctx, query, id).Scan(&w.ID, &w.OwnerID, &w.Name, &description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrWorkflowNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("query workflow by id: %w", err)
	}

	if description != nil {
		w.Description = *description
	}

	return &w, nil
}

func (r *pgWorkflowRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Workflow, error) {
	query := `
		SELECT id, owner_id, name, description, is_active, created_at, updated_at
		FROM workflows
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("query workflows by owner: %w", err)
	}
	defer rows.Close()

	workflows := make([]models.Workflow, 0)

	for rows.Next() {
		w := models.Workflow{}
		var description *string
		if err := rows.Scan(&w.ID, &w.OwnerID, &w.Name, &description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan workflow row: %w", err)
		}
		if description != nil {
			w.Description = *description
		}
		workflows = append(workflows, w)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workflow rows: %w", err)
	}
	return workflows, nil
}
