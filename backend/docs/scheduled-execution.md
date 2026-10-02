# Durable scheduled execution

Stop older server instances, apply migrations through `000011_outbox_leases`,
then start the updated server. Do not mix old lock-based consumers with leased
workers: old consumers do not respect lease ownership. Rollback also requires
stopping all new workers before reverting the migration and application.

Scheduled, manual, and webhook runs share a PostgreSQL transactional outbox.
Manual and webhook requests commit the run and outbox entry together.

For each due occurrence, one transaction locks the schedule with `FOR UPDATE
SKIP LOCKED`, verifies its original due time, inserts a pending workflow run and
outbox entry, and advances the schedule. Rollback leaves the occurrence due.
A unique `(schedule_id, scheduled_for)` index protects occurrence identity.
Concurrent schedulers with a stale due-list entry cannot advance it again.
The workflow version is captured when the run is created.

Startup launches the cron scheduler, outbox dispatcher, and worker pool separately.
The dispatcher polls once per second and places only run IDs in a bounded in-memory
queue. Full queues and lost process memory leave the database outbox untouched.
Local hints are deduplicated until execution finishes. Workers claim after dequeue,
so queued IDs consume no lease. Across instances, competing claims coordinate with
`FOR UPDATE SKIP LOCKED` in a short transaction. Graph/version and webhook payload
are loaded from the database while claiming, and `running` status commits immediately.

The claim has a random ownership token and 30-second lease, renewed every 10 seconds.
Execution holds no database connection between heartbeats. Database operations have
five-second deadlines. A failed heartbeat cancels the executor; expired leases are
reclaimable. Completion atomically stores the terminal result and removes the outbox
row, conditional on the current unexpired token. A stale worker cannot acknowledge
or release a replacement worker's job. Graceful cancellation releases ownership and
returns the run to pending; process loss is recovered after lease expiry. If release
fails during a database outage, lease expiry still recovers the job.

Workflow errors receive up to two retries within a shared 60-second execution
timeout; exhausted retries are recorded as failed and acknowledged.

`WORKER_CONCURRENCY` defaults to 4; `WORKER_QUEUE_CAPACITY` defaults to 64.
Tune concurrency for node resource usage and brief database claim/renewal demand.
Node implementations must honor context cancellation. Lease fencing protects
database results, but cannot undo an HTTP request or other external side effect.

Workflow and initial version creation is a single transaction. During traversal,
persisted webhook data replaces the webhook trigger's fallback output. Each join
waits for all incoming edges to resolve. It runs once if any incoming edge is active;
otherwise it is skipped and resolves its outgoing edges as inactive. Errors abort
the attempt rather than treating a failed branch as skipped.

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
TEST_DATABASE_URL=postgres://... go test -race ./internal/repositories ./internal/workflow ./internal/queue ./internal/worker
```

The tests create and remove isolated schemas and cover transactional rollback,
competing schedulers/consumers, webhook replay through the engine, visible running
status with a single database connection, renewal, stale-owner fencing, interrupted
consumption, and atomic creation/acknowledgement. Database tests are skipped without
`TEST_DATABASE_URL`; CI sets it explicitly. Engine and worker tests need no database.
