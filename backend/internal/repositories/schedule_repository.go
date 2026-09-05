package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepository interface {
	Create(ctx context.Context, s *models.Schedule) error
	ListByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Schedule, error)
	ListDue(ctx context.Context, now time.Time) ([]models.Schedule, error)
	MarkRun(ctx context.Context, id uuid.UUID, ranAt, nextRunAt time.Time) error
	Deactivate(ctx context.Context, id, workflowOwnerID uuid.UUID) error
}

type pgScheduleRepository struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) ScheduleRepository {
	return &pgScheduleRepository{pool: pool}
}

func (r *pgScheduleRepository) Create(ctx context.Context, s *models.Schedule) error {
	query := `
		INSERT INTO schedules (workflow_id, cron_expression, is_active, next_run_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query, s.WorkflowID, s.CronExpression, s.IsActive, s.NextRunAt).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert schedule: %w", err)
	}
	return nil
}

func (r *pgScheduleRepository) ListByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]models.Schedule, error) {
	query := `
		SELECT id, workflow_id, cron_expression, is_active, next_run_at, last_run_at, created_at
		FROM schedules
		WHERE workflow_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("list schedules by workflow: %w", err)
	}
	defer rows.Close()

	schedules := make([]models.Schedule, 0)
	for rows.Next() {
		s := models.Schedule{}
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.CronExpression, &s.IsActive, &s.NextRunAt, &s.LastRunAt, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan schedule row: %w", err)
		}

		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *pgScheduleRepository) ListDue(ctx context.Context, now time.Time) ([]models.Schedule, error) {
	query := `
		SELECT id, workflow_id, cron_expression, is_active, next_run_at, last_run_at, created_at
		FROM schedules
		WHERE is_active = true AND next_run_at <= $1
		ORDER BY next_run_at ASC
	`

	rows, err := r.pool.Query(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("list due schedules: %w", err)
	}
	defer rows.Close()

	schedules := make([]models.Schedule, 0)
	for rows.Next() {
		s := models.Schedule{}
		if err := rows.Scan(&s.ID, &s.WorkflowID, &s.CronExpression, &s.IsActive, &s.NextRunAt, &s.LastRunAt, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan due schedule row: %w", err)
		}

		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *pgScheduleRepository) MarkRun(ctx context.Context, id uuid.UUID, ranAt, nextRunAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE schedules SET last_run_at = $1, next_run_at = $2 WHERE id = $3`,
		ranAt, nextRunAt, id,
	)
	if err != nil {
		return fmt.Errorf("mark schedule run: %w", err)
	}

	return nil
}

func (r *pgScheduleRepository) Deactivate(ctx context.Context, id, workflowOwnerID uuid.UUID) error {
	query := `
		UPDATE schedules
		SET is_active = false
		WHERE id = $1 AND workflow_id IN (SELECT id FROM workflows WHERE owner_id = $2)
	`
	tag, err := r.pool.Exec(ctx, query, id, workflowOwnerID)
	if err != nil {
		return fmt.Errorf("deactivate schedule: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrWorkflowNotFound
	}
	return nil
}
