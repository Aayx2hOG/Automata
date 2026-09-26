# Durable scheduled execution

Apply migrations through `000010_execution_security` before starting the updated server.

Scheduled, manual, and webhook runs share a PostgreSQL transactional outbox.
Manual and webhook requests commit the run and outbox entry together.

For each due occurrence, one transaction locks the schedule with `FOR UPDATE
SKIP LOCKED`, verifies its original due time, inserts a pending workflow run and
outbox entry, and advances the schedule. Rollback leaves the occurrence due.
A unique `(schedule_id, scheduled_for)` index protects occurrence identity.
Concurrent schedulers with a stale due-list entry cannot advance it again.
The workflow version is captured when the run is created.

Each scheduler process also runs one outbox consumer. It polls once per second
when idle and drains available work immediately. It locks an outbox row with
`SKIP LOCKED`, executes the graph directly, then commits the result and removes
the outbox entry in the same transaction. PostgreSQL releases the lock when a
failed connection is detected, making unfinished work available again. Shutdown
cancellation rolls back consumption instead of recording a terminal failure.
Workflow errors receive up to two retries within a shared 60-second execution
timeout; exhausted retries are recorded as failed and acknowledged.

This deliberately holds one database connection and a transaction while a
scheduled workflow executes. The running status is not visible outside that
transaction; the run remains visibly pending until its terminal result commits.
Keep database idle-in-transaction timeouts longer than the execution timeout.
For higher throughput or longer workflows, replace the consumer lock with
renewable leases and fenced acknowledgements, or relay to a durable broker.

Execution is at least once across crashes: an external HTTP request may succeed
before the result transaction commits. Nodes that perform non-repeatable actions
need downstream idempotency. The unique occurrence key prevents duplicate run
records, not duplicate external effects.

The existing policy skips missed cron intervals by computing the next occurrence
from the current time. Inactive workflows or missing versions leave the schedule
due and log an error.

Run the PostgreSQL regression tests against a test database:

```sh
cd backend
TEST_DATABASE_URL=postgres://... go test ./internal/repositories -v
```

The tests create and remove isolated schemas and cover transactional rollback,
competing schedulers, competing consumers, interrupted consumption, and terminal
failure acknowledgement. Without `TEST_DATABASE_URL`, these tests are skipped.
