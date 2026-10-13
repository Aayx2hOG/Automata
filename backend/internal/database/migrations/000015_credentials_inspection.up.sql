CREATE TABLE credentials (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    header TEXT NOT NULL,
    encrypted_value BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);
CREATE INDEX credentials_owner_id ON credentials(owner_id);
-- Older checkpoints have no reliable start time; leave them unknown.
ALTER TABLE workflow_node_runs ADD COLUMN started_at TIMESTAMPTZ;
ALTER TABLE workflow_node_runs ADD COLUMN terminal_failure BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE workflow_node_runs ADD COLUMN retry_at TIMESTAMPTZ;
