ALTER TABLE workflow_node_runs DROP COLUMN started_at;
ALTER TABLE workflow_node_runs DROP COLUMN terminal_failure;
ALTER TABLE workflow_node_runs DROP COLUMN retry_at;
DROP TABLE credentials;
