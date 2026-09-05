package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/Aayx2hOG/automata/internal/models"
	"github.com/Aayx2hOG/automata/internal/repositories"
	"github.com/Aayx2hOG/automata/internal/services"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Scheduler struct {
	schedules        repositories.ScheduleRepository
	workflowServices *services.WorkflowService
	parser           cron.Parser
	interval         time.Duration
	logger           *zap.Logger
}

func New(schedules repositories.ScheduleRepository, workflowService *services.WorkflowService, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		schedules:        schedules,
		workflowServices: workflowService,
		parser:           cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
		interval:         30 * time.Second,
		logger:           logger,
	}
}

func (s *Scheduler) ParseNext(cronExpr string, after time.Time) (time.Time, error) {
	schedule, err := s.parser.Parse(cronExpr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression: %w", err)
	}
	return schedule.Next(after), nil
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.logger.Info("Scheduler started", zap.Duration("interval", s.interval))

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			s.logger.Info("Scheduler stopping: context cancelled")
			return
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	now := time.Now()

	due, err := s.schedules.ListDue(ctx, now)
	if err != nil {
		s.logger.Error("failed to list due to schedules", zap.Error(err))
		return
	}
	for _, schedule := range due {
		s.trigger(ctx, schedule, now)
	}
}

func (s *Scheduler) trigger(ctx context.Context, schedule models.Schedule, now time.Time) {
	nextRun, err := s.ParseNext(schedule.CronExpression, now)
	if err != nil {
		s.logger.Error("schedule has invalid cron expression, skipping...", zap.String("schedule_id", schedule.ID.String()), zap.Error(err))
		return
	}

	if err := s.schedules.MarkRun(ctx, schedule.ID, now, nextRun); err != nil {
		s.logger.Error("scheduled run failed to enqueue",
			zap.String("schedule_id", schedule.ID.String()),
			zap.String("workflow_id", schedule.WorkflowID.String()),
			zap.Error(err),
		)
	}
}
