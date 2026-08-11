-- Each save creates a new version rather than mutating the graph in
-- place, matching the "Workflow versioning" goal from the vision doc
-- and giving executions a stable, immutable definition to run against.
CREATE TABLE workflow_versions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id  UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    version      INTEGER NOT NULL,
    graph        JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (workflow_id, version)
);

CREATE INDEX idx_workflow_versions_workflow_id ON workflow_versions(workflow_id);
