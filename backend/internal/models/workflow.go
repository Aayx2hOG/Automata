package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkflowGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

type Workflow struct {
	ID          uuid.UUID `json:"id"`
	OwnerID     uuid.UUID `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type GraphNode struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

type GraphEdge struct {
	FromNodeID string `json:"from_node_id"`
	ToNodeID   string `json:"to_node_id"`
	Condition  string `json:"condition,omitempty"`
}

type WorkflowVersion struct {
	ID         uuid.UUID     `json:"id"`
	WorkflowID uuid.UUID     `json:"workflow_id"`
	Version    int           `json:"version"`
	Graph      WorkflowGraph `json:"graph"`
	CreatedAt  time.Time     `json:"created_at"`
}

type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusSucceeded RunStatus = "succeeded"
	RunStatusFailed    RunStatus = "failed"
)

type TriggerType string

const (
	TriggerManual   TriggerType = "manual"
	TriggerWebhook  TriggerType = "webhook"
	TriggerCron     TriggerType = "cron"
	TriggerInterval TriggerType = "interval"
	TriggerAPI      TriggerType = "api"
)

type WorkflowRun struct {
	ID                uuid.UUID      `json:"id"`
	WorkflowID        uuid.UUID      `json:"workflow_id"`
	WorkflowVersionID uuid.UUID      `json:"workflow_version_id"`
	Status            RunStatus      `json:"status"`
	TriggerType       TriggerType    `json:"trigger_type"`
	Outputs           map[string]any `json:"outputs,omitempty"`
	ErrorMessage      string         `json:"error_message,omitempty"`
	StartedAt         *time.Time     `json:"started_at,omitempty"`
	FinishedAt        *time.Time     `json:"finished_at,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
}

type Schedule struct {
	ID             uuid.UUID  `json:"id"`
	WorkflowID     uuid.UUID  `json:"workflow_id"`
	CronExpression string     `json:"cron_expression"`
	IsActive       bool       `json:"is_active"`
	NextRunAt      time.Time  `json:"next_run_at"`
	LastRunAt      *time.Time `json:"last_run_at"`
	CreatedAt      time.Time  `json:"created_at"`
}
