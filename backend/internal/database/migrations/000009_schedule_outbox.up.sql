-- Occurrence identity remains on the run after the outbox entry is consumed.
ALTER TABLE workflow_runs ADD COLUMN schedule_id UUID REFERENCES schedules(id) ON DELETE SET NULL;
ALTER TABLE workflow_runs ADD COLUMN scheduled_for TIMESTAMPTZ;
CREATE UNIQUE INDEX workflow_runs_schedule_occurrence ON workflow_runs(schedule_id, scheduled_for);

CREATE TABLE schedule_outbox (
    run_id UUID PRIMARY KEY REFERENCES workflow_runs(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
