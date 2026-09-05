CREATE TABLE schedules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id      UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    cron_expression  TEXT NOT NULL,
    is_active        BOOLEAN NOT NULL DEFAULT true,
    next_run_at      TIMESTAMPTZ NOT NULL,
    last_run_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_schedules_due ON schedules(next_run_at) WHERE is_active = true;
