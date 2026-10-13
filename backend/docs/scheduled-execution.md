# Execution, checkpoints, and admission

## Upgrade compatibility

Stop all application instances, apply all migrations through
`000013_node_execution_state`, and deploy the new binary. Do not mix old workers
with checkpoint-aware workers or old API instances with quota-aware admission.
Existing queued runs have no historical checkpoints, so their first execution
under this version starts from the beginning. Reconcile any external effects from
previous interrupted executions before resuming those jobs.

Rollback requires stopping workers first. Reverting migration 13 removes
checkpoints and their attempt budgets. Do not resume unfinished work under the old
binary without assessing duplicate-delivery risks.

## Admission and version identity

Manual, webhook, and scheduled runs atomically create a run and outbox entry.
Each run pins its workflow version. Scheduled admission also locks the due
occurrence, advances the schedule, and uses a unique `(schedule_id, scheduled_for)`
key. A rollback leaves the occurrence due.

Every trigger passes the same PostgreSQL admission check. A transaction-scoped
advisory lock serializes checks and inserts across all instances. It is held only
for the short admission transaction, never during execution. Pending limits count
all outbox rows, including running jobs. The rolling one-minute quota counts all
runs admitted for an owner, including already completed runs.

| Setting | Default |
|---|---|
| `EXECUTION_MAX_PENDING_RUNS` | 10000 |
| `EXECUTION_MAX_PENDING_PER_USER` | 100 |
| `EXECUTION_RUNS_PER_MINUTE` | 60 |

Use identical limits across replicas. Owner quota failures return `429`; global
capacity failures return `503`. Both carry `Retry-After: 60`. A scheduled occurrence
rejected by admission remains due; it is retried on a later scheduler poll. These
quotas bound accepted execution load, not the number of incoming HTTP connections.
The shared admission lock favors correctness and simplicity; measure transaction
latency before increasing admission throughput significantly.

## Claims and checkpoints

The dispatcher polls every second and copies durable IDs into a bounded local
queue. Losing this queue loses no committed jobs. Workers claim after dequeue with
`FOR UPDATE SKIP LOCKED`, a random token, and a 30-second lease. Heartbeats renew
every ten seconds. Lease loss cancels execution; crashes become reclaimable after
expiry. Claims, checkpoints, and acknowledgements use short database transactions
with five-second deadlines. No connection is held while a node executes.

Before executing a node, the worker commits its attempt count to
`workflow_node_runs`. After success, it commits the output before visiting a
successor. Recovery traverses the graph using completed outputs, preserving branch
and join decisions. Joins wait for all incoming edges to resolve and execute once
if any incoming edge is active. Fully inactive branches remain skipped.

Checkpoint mutations lock and verify the current outbox lease. A stale worker
cannot overwrite a replacement worker's progress. Checkpoint storage failures keep
the job pending rather than acknowledge it as a terminal execution failure. Final
acknowledgement atomically stores the workflow result and removes the outbox entry.
If that acknowledgement fails, completed node checkpoints still survive.

Graceful shutdown cancels active execution and releases unfinished jobs. Recovery
keeps completed outputs and consumed attempt counts. A started but uncheckpointed
node may have an unknown outcome and has already consumed an attempt.

## Node retry policy

Every node defaults to one attempt. To enable retries, publish a graph node with:

```json
{
  "id": "fetch",
  "type": "http_request",
  "config": { "method": "GET", "url": "https://example.com/data" },
  "retry": { "max_attempts": 3, "backoff_seconds": 2 }
}
```

`max_attempts` includes the first attempt and must be 1–5. Fixed backoff must be
0–30 seconds. Only the failed node repeats. The retry budget persists across worker
restarts. Fixed backoff is in memory; recovery may skip that delay. HTTP
`Retry-After` deadlines are persisted and cannot be skipped by recovery. All nodes
and retries share the existing 60-second deadline for that worker execution;
recovery starts a new deadline. Delay and backoff currently occupy a worker.

Retry policies apply to node execution errors. HTTP nodes default to
`error_policy: "output"`, returning non-2xx responses as outputs with
`success=false`. With `error_policy: "fail"`, non-2xx responses fail the node;
408, 429, and 5xx are retryable, while other statuses stop immediately, including
after worker recovery. Retry-After seconds and HTTP dates set the minimum wait.
Waits beyond the execution deadline fail without retrying early. The visual
inspector edits both retry settings and HTTP error policy. Migration 15 is required;
stop old workers before applying it and deploying the new release.

## External effects and idempotency

HTTP nodes send `Idempotency-Key: automata-<sha256(run-id:node-id)>`. The engine-owned
header overrides a configured header of the same name, is stable across node
attempts, and differs between runs and nodes. The destination must actually honor
this header, retain keys for the required recovery window, and reject inconsistent
payload reuse for it to prevent duplicates. Providers requiring a different header
or request field need an integration-specific adapter.

There is no distributed transaction between PostgreSQL and an external API. An
external request can succeed before its output checkpoint commits. With remaining
attempts, recovery can repeat that request using the same key. With an exhausted
budget, the run fails and reports that the outcome may be unknown. Inspect the
external system before manually triggering another run: a new run has a new key.

The API does not deduplicate separate manual requests or webhook deliveries.
Webhook signatures authenticate delivery but allow replay within their timestamp
window. Schedule occurrence uniqueness prevents duplicate run records for a single
occurrence; it does not guarantee exactly-once external effects.

## Operations and tests

Tune worker concurrency against node resource usage and database capacity.
Defaults are four workers and 64 local queued IDs. Cron polls every 30 seconds;
missed intervals are skipped by calculating the next occurrence from current time.
Inactive workflows or missing versions leave a schedule due and log an error.

For an individual run, inspect checkpoints with a parameterized query equivalent
to `SELECT node_id, status, attempts, error_message FROM workflow_node_runs WHERE
run_id = $1`. Do not modify checkpoint rows or leases while workers are active.
Outputs may contain sensitive business data; restrict database access accordingly.

Run `go test -race ./internal/repositories ./internal/workflow ./internal/queue
./internal/worker` from `backend/` with `TEST_DATABASE_URL` pointing to a dedicated
test database. Tests cover concurrent admission, rollback, scheduled quota
rejection, rolling windows, resumed node outputs, persisted attempts, stale-worker
fencing, checkpoint failure, heartbeat recovery, and final acknowledgement.
