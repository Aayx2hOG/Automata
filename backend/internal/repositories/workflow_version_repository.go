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

type WorkflowVersionRepository interface {
	Create(ctx context.Context, workflowID uuid.UUID, graph models.WorkflowGraph) (*models.WorkflowVersion, error)
	GetLatestByWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.WorkflowVersion, error)
}

type pgWorkflowVersionRepository struct {
	pool *pgxpool.Pool
}

func NewWorkflowVersionRepository(pool *pgxpool.Pool) WorkflowVersionRepository {
	return &pgWorkflowVersionRepository{pool: pool}
}

func (r *pgWorkflowVersionRepository) Create(ctx context.Context, workflowID uuid.UUID, graph models.WorkflowGraph) (*models.WorkflowVersion, error) {
	graphJSON, err := json.Marshal(graph)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow graph: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT id FROM workflows WHERE id = $1 FOR UPDATE`, workflowID); err != nil {
		return nil, fmt.Errorf("lock workflow row: %w", err)
	}

	var nextVersion int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(version), 0) + 1 FROM workflow_versions WHERE workflow_id = $1`,
		workflowID,
	).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("compute next version: %w", err)
	}

	version := &models.WorkflowVersion{
		WorkflowID: workflowID,
		Version:    nextVersion,
		Graph:      graph,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO workflow_versions (workflow_id, version, graph)
		 VALUES ($1, $2, $3::jsonb)
		 RETURNING id, created_at`,
		workflowID, nextVersion, graphJSON,
	).Scan(&version.ID, &version.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("insert workflow version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return version, nil
}

func (r *pgWorkflowVersionRepository) GetLatestByWorkflow(ctx context.Context, workflowID uuid.UUID) (*models.WorkflowVersion, error) {
	query := `
		SELECT id, workflow_id, version, graph, created_at
		FROM workflow_versions
		WHERE workflow_id = $1
		ORDER BY version DESC
		LIMIT 1
	`
	version := models.WorkflowVersion{}
	graphByte := []byte{}

	err := r.pool.QueryRow(ctx, query, workflowID).Scan(
		&version.ID, &version.WorkflowID, &version.Version, &graphByte, &version.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, models.ErrWorkflowVersionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query latest workflow version: %w", err)
	}
	if err := json.Unmarshal(graphByte, &version.Graph); err != nil {
		return nil, fmt.Errorf("unmarshal workflow graph: %w", err)
	}
	return &version, nil
}

