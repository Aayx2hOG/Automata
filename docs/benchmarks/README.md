# Recovery and latency verification

[Measured results](verification.json) · [Application demo](../demo/README.md)

## Reproduce the experiment

Build the API from `backend/`:

```sh
go build -o /tmp/automata-verification-server ./cmd/server
```

From the repository root, set `TEST_DATABASE_URL` to a **disposable, fully migrated
test database**, then run:

```sh
python3 backend/scripts/verify_execution.py \
  --binary /tmp/automata-verification-server \
  --samples 20 \
  --output docs/benchmarks/verification.json
```

The script starts its own API on port 18080, registers a fresh account, and creates
its own workflows. It kills only its own API process after the first node has
checkpointed and a delay node has started, then restarts that process against the
same database. It asserts that the completed node retains its output and one
attempt, the interrupted delay completes on its second allowed attempt, and the
final node runs once. The script shuts down its process afterward and leaves the
sample account and runs in the disposable database for inspection.

The delay explicitly permits two attempts because repeating a delay is safe.
Do not generalize this experiment to exactly-once external writes: destinations
must support idempotency, and a write can succeed before its checkpoint commits.
The default lease lasts 30 seconds, so a hard crash can require waiting for lease
expiry before recovery begins.

After recovery, the script measures 20 **sequential, single-node** runs. Admission
latency is the HTTP request duration. Observed completion includes dispatch and
100 ms status polling; it is not precise node execution time. This is a small
local baseline, not a saturation, concurrency, or production throughput test.

Captured on 2026-10-15 with PostgreSQL 18 on this development machine, default
worker concurrency (4), and one API process:

| Measurement | Result |
|---|---:|
| Recovery after restart | 35.11 s |
| Completed / interrupted / final node attempts | 1 / 2 / 1 |
| Admission median / p95 | 9.10 / 12.81 ms |
| Observed completion median / p95 | 957.48 / 1058.15 ms |

Results depend on hardware, database placement, load, and dispatch timing. Rerun
the script before quoting numbers for a different environment.

## Interview discussion

Explain why run creation and outbox insertion share a transaction; how lease
fencing prevents stale workers from committing; what checkpoints can recover; and
why an external effect cannot be guaranteed exactly once by this application.
The UI makes these behaviors visible through saved workflow versions, per-node
attempts, outputs, and errors.

Possible résumé wording:

> Built a workflow automation platform with Go, PostgreSQL, and React, featuring
> a visual DAG editor, signed webhook triggers, and cron scheduling. Implemented
> transactional job admission, renewable worker leases, node checkpoints, and
> configurable retries for recoverable execution.
