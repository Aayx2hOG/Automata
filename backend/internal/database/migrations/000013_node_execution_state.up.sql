CREATE TABLE workflow_node_runs (
    run_id UUID NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    node_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('running', 'succeeded', 'failed')),
    attempts INTEGER NOT NULL CHECK (attempts > 0),
    output JSONB,
    error_message TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    PRIMARY KEY (run_id, node_id)
);

-- Supports the rolling admission window without scanning all historical runs.
CREATE INDEX workflow_runs_created_at ON workflow_runs(created_at);
