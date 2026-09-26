DROP TABLE schedule_outbox;
DROP INDEX workflow_runs_schedule_occurrence;
ALTER TABLE workflow_runs DROP COLUMN scheduled_for;
ALTER TABLE workflow_runs DROP COLUMN schedule_id;
