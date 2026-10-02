package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkflowRepository interface {
	UpdateWithVersion(ctx context.Context, id, ownerID uuid.UUID, name, description string, graph models.WorkflowGraph) (*models.WorkflowVersion, error)
	Create(ctx context.Context, w *models.Workflow) error
	CreateWithInitialVersion(ctx context.Context, w *models.Workflow, graph models.WorkflowGraph) (*models.WorkflowVersion, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.Workflow, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]models.Workflow, error)
}

type pgWorkflowRepository struct {
	pool *pgxpool.Pool
}

func NewWorkflowRepository(pool *pgxpool.Pool) WorkflowRepository {
	return &pgWorkflowRepository{pool: pool}
}

// CreateWithInitialVersion publishes a workflow only when its first version exists.
func (r *pgWorkflowRepository) CreateWithInitialVersion(ctx context.Context, w *models.Workflow, graph models.WorkflowGraph) (*models.WorkflowVersion, error) {
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow graph: %w", err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	created := *w
	if err = tx.QueryRow(ctx, `INSERT INTO workflows (owner_id, name, description, is_active)
 VALUES ($1,$2,$3,$4) RETURNING id, created_at, updated_at, webhook_secret`,
		w.OwnerID, w.Name, w.Description, w.IsActive).Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt, &created.WebhookSecret); err != nil {
		return nil, fmt.Errorf("insert workflow: %w", err)
	}
	version := &models.WorkflowVersion{WorkflowID: created.ID, Version: 1, Graph: graph}
	if err = tx.QueryRow(ctx, `INSERT INTO workflow_versions (workflow_id, version, graph)
 VALUES ($1,1,$2::jsonb) RETURNING id, created_at`, created.ID, graphJSON).Scan(&version.ID, &version.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert initial version: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	*w = created
	return version, nil
}

func (r *pgWorkflowRepository) Create(ctx context.Context, w *models.Workflow) error {
	query := `
		INSERT INTO workflows (owner_id, name, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at, webhook_secret
	`
	err := r.pool.QueryRow(ctx, query, w.OwnerID, w.Name, w.Description, w.IsActive).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt, &w.WebhookSecret)
	if err != nil {
		return fmt.Errorf("insert workflow: %w", err)
	}
	return nil
}

func (r *pgWorkflowRepository) GetById(ctx context.Context, id uuid.UUID) (*models.Workflow, error) {
	query := `
SELECT id, owner_id, name, description, is_active, created_at, updated_at, webhook_secret
		FROM workflows
		WHERE id = $1
	`
	w := models.Workflow{}
	var description *string

	err := r.pool.QueryRow(ctx, query, id).Scan(&w.ID, &w.OwnerID, &w.Name, &description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt, &w.WebhookSecret)

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
		SELECT id, owner_id, name, description, is_active, created_at, updated_at, webhook_secret
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
		if err := rows.Scan(&w.ID, &w.OwnerID, &w.Name, &description, &w.IsActive, &w.CreatedAt, &w.UpdatedAt, &w.WebhookSecret); err != nil {
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

func (r *pgWorkflowRepository) UpdateWithVersion(ctx context.Context, id, ownerID uuid.UUID, name, description string, graph models.WorkflowGraph) (*models.WorkflowVersion, error) {
	data, err := json.Marshal(graph)
	if err != nil {
		return nil, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	result, err := tx.Exec(ctx, `UPDATE workflows SET name=$3, description=$4, updated_at=now() WHERE id=$1 AND owner_id=$2`, id, ownerID, name, description)
	if err != nil {
		return nil, err
	}
	if result.RowsAffected() == 0 {
		return nil, models.ErrWorkflowNotFound
	}
	v := &models.WorkflowVersion{WorkflowID: id, Graph: graph}
	err = tx.QueryRow(ctx, `INSERT INTO workflow_versions(workflow_id,version,graph)
 SELECT $1,COALESCE(MAX(version),0)+1,$2::jsonb FROM workflow_versions WHERE workflow_id=$1
 RETURNING id,version,created_at`, id, data).Scan(&v.ID, &v.Version, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return v, nil
}
