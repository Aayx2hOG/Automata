# Automata

Automata is a self-hosted workflow automation application with a visual DAG editor,
a Go API, and PostgreSQL-backed execution. It supports manual runs, signed webhooks,
cron schedules, conditional branches, and persisted execution results.

## Architecture

```text
React editor → Go HTTP API ──────────────┐
Signed webhooks → Go HTTP API           ├→ PostgreSQL: workflow versions, runs, outbox
Cron scheduler ────────────────────────┘                     ↓
                                        Dispatcher → bounded queue → workers
                                                                        ↓
                                                   DAG engine → node checkpoints
```

The API, scheduler, dispatcher, and workers run in one process. PostgreSQL is the
source of truth; the local queue contains disposable run IDs. Creating a run and
its outbox entry is atomic. Runs retain the workflow version selected at admission.
Workers claim renewable, fenced leases using short transactions and release the
connection while nodes execute. Multiple instances may share one database.

Successful node outputs are checkpointed. Recovery reuses those outputs instead
of repeating completed nodes. Retries apply only to the failing node; attempt
counts survive worker restarts. Every node defaults to **one attempt**. Explicit
retry policies allow up to five attempts with a fixed backoff of 0–30 seconds.
HTTP nodes send a stable `Idempotency-Key` derived from the run and node IDs.

**External effects are not exactly once.** A destination can accept a request
before its checkpoint commits. Enable retries for writes only when the destination
supports idempotency. A node whose attempt budget is exhausted after an interrupted
attempt fails with an unknown-outcome error rather than silently repeating the
write. Separate incoming requests create separate runs and idempotency keys.

See [execution semantics](backend/docs/scheduled-execution.md) for recovery,
retry configuration, migration compatibility, and operational limits.

## Requirements

- Go matching [backend/go.mod](backend/go.mod), currently 1.26.5 or later.
- PostgreSQL 16 or later; Docker Compose v2 can provide the development database.
- Node.js 24+ and npm for the frontend.
- `golang-migrate` v4.18.3 when applying migrations without Compose.

## Local development

From the repository root:

```bash
cp .env.example .env
openssl rand -hex 32
```

Set `JWT_SECRET` in `.env` to the generated value. The example database password is
for local development only. Start the database and apply **all** migrations:

```bash
docker compose -f compose.dev.yml up -d --wait postgres
docker compose -f compose.dev.yml run --rm migrate
```

Start the backend:

```bash
cd backend
go run ./cmd/server
```

In another terminal:

```bash
cd frontend
npm ci
npm run dev
```

Open http://localhost:3000. Vite proxies `/api` to the backend on port 8080. Check
the API with `curl http://localhost:8080/healthz`. Create a blank workflow to test
execution without an external service. See [frontend documentation](frontend/README.md)
for browser tests and editor behavior.

`compose.dev.yml` provides the development database and a migration command.
`docker-compose.yml` is a database-only alternative with its own persistent volume;
use one configuration at a time. Neither starts the application or configures
production backups. Both bind PostgreSQL to loopback.

## Configuration

Existing process environment variables take precedence over `.env` files. The
backend reads `.env` in its working directory, then `../.env`, without overwriting
already-set variables. Environment settings override optional `config.yaml` or
`config.json` values and built-in defaults. Invalid configuration fails startup.
See [the configuration example](backend/internal/config/config.example.yaml).

| Variable | Default | Purpose |
|---|---|---|
| `DATABASE_URL` | Required | PostgreSQL connection string |
| `JWT_SECRET` | Required | Random signing secret, at least 32 characters |
| `ENV` | `development` | Runtime environment; use `production` when deployed |
| `HTTP_PORT` | `8080` | API port |
| `HTTP_ALLOWED_ORIGINS` | Empty | Explicit browser origins; empty disables cross-origin access |
| `HTTP_TRUSTED_PROXIES` | Empty | Comma-separated CIDRs of trusted reverse proxies |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Shutdown deadline |
| `DATABASE_MAX_CONNS` | `10` | Pool capacity per process |
| `WORKER_CONCURRENCY` | `4` | Concurrent workflow executions per process |
| `WORKER_QUEUE_CAPACITY` | `64` | Local run-ID queue capacity |
| `EXECUTION_MAX_PENDING_RUNS` | `10000` | Global queued + running job limit |
| `EXECUTION_MAX_PENDING_PER_USER` | `100` | Queued + running job limit per owner |
| `EXECUTION_RUNS_PER_MINUTE` | `60` | Per-owner rolling admission limit across all triggers |
| `AUTH_ACCESS_TOKEN_TTL` | `15m` | Access-token lifetime |
| `AUTH_REFRESH_TOKEN_TTL` | `168h` | Refresh-token lifetime |

All instances sharing a database must use the same admission limits. Limits are
checked inside the run/outbox transaction under a shared admission lock. Owner
limits return `429`; a full global backlog returns `503`, both with `Retry-After`.
Rejected cron occurrences remain due for a later scheduler poll.

Forwarded IP headers are ignored unless the directly connected peer is trusted.
For `X-Forwarded-For`, the API walks the chain from right to left to find the first
untrusted hop. Configure only actual proxy addresses and have the edge proxy
sanitize incoming forwarding headers. IP-based auth/webhook throttles are local
to each process; database admission quotas are shared across instances.

All API request bodies are limited to 1 MiB. Graphs allow at most 100 nodes,
500 edges, and 255 bytes per node ID. Oversized bodies return `413`; invalid graphs
return `400` with validation details. Rate-limiter tracking is bounded.

The former `CREDENTIALS_ENCRYPTION_KEY` setting was unused and is no longer required
or consumed. Credential storage and encryption are not implemented; do not assume
workflow configuration or output is encrypted at the application layer.

## API

Authenticated routes require `Authorization: Bearer <access_token>` and enforce
workflow ownership.

| Method | Path | Purpose |
|---|---|---|
| POST | `/auth/register`, `/auth/login` | Create an account or authenticate |
| POST | `/auth/refresh`, `/auth/logout` | Rotate or revoke refresh tokens |
| GET | `/users/me` | Current user |
| POST, GET | `/workflows` | Create or list workflows |
| GET, PUT | `/workflows/{id}` | Read or publish a new workflow version |
| POST | `/workflows/{id}/run` | Queue a manual run |
| GET | `/runs/{id}` | Execution status and outputs |
| GET | `/runs?limit=50&offset=0` | Owner's run history, maximum page size 100 |
| POST, GET | `/workflows/{id}/schedules` | Create or list cron schedules |
| DELETE | `/schedules/{scheduleID}` | Deactivate a schedule |
| POST | `/webhook/{workflowID}` | Queue a signed webhook run |
| GET | `/healthz` | Process and database connectivity check |

Webhooks require `X-Webhook-Timestamp` and `X-Webhook-Signature`. Sign the exact
request bytes with HMAC-SHA256 over `timestamp + "." + body`, using the workflow's
secret string directly. The signature is hex encoded, and the timestamp must be
within five minutes of server time. Signatures authenticate requests but do not
deduplicate delivery. See [security details](backend/docs/execution-security.md).

## Deployment and upgrades

1. Provision PostgreSQL, restricted networking, TLS termination, and runtime secrets.
2. Configure explicit browser origins, trusted proxies, and shared admission limits.
3. Stop existing application instances before applying migration
   `000013_node_execution_state` and deploying this version. Old workers cannot
   honor checkpoints or quotas. Already pending runs have no historical node
   checkpoints; reconcile possible external effects before resuming them.
4. Apply all migrations once using a migration role, then start the application
   with an appropriately restricted runtime database role. Do not run migrations
   independently from every application replica.
5. Verify health, a workflow run, a signed webhook, and a scheduled occurrence.
6. Configure backup scheduling, off-host storage, monitoring, and restore drills.

Database schema changes are additive, but rollback of migration 13 deletes node
checkpoints. Stop all workers before rollback and reconcile unfinished external
operations. Never mix old and new workers during this upgrade.

The repository does not provision production hosting, automatic backups, metrics,
or tracing. [Backup and recovery procedures](backend/docs/database-recovery.md)
describe the supported logical backup workflow and its limits. HTTP bind errors
terminate the process with a failure exit status after worker shutdown.

## Verification

From `backend/`, fast tests require no database:

```bash
go test ./tests/unit ./internal/workflow ./internal/queue ./internal/worker \
  ./internal/config ./internal/middleware ./internal/nodes/http_request \
  ./internal/services ./internal/api/handlers
```

For full verification, use a **dedicated test database** with all migrations applied:

```bash
export DATABASE_URL='postgres://user:password@localhost:5433/automata_test?sslmode=disable'
export TEST_DATABASE_URL="$DATABASE_URL"
export JWT_SECRET='test-only-secret-at-least-32-characters'
go vet ./...
go build ./...
go test ./... -race -cover
```

Repository tests create isolated schemas and remove them afterwards. API tests
write test users to the configured database. Database suites skip if their URLs
are unset; a passing run without these variables does not validate persistence.

CI runs migrations, vet, build, race-enabled Go tests, a logical backup/restore
smoke test, frontend checks, and browser regressions. Test-only IP-limit bypass
requires both `ENV=test` and `TEST_DISABLE_RATE_LIMIT=1`; durable run quotas remain
enforced. No external shell-script files are required.

## Current boundaries

Execution is sequential within each DAG and has a 60-second deadline per worker
execution. Delay nodes occupy workers. HTTP non-2xx responses are returned as node
outputs (`success=false`), so workflows must branch on the result. Long-lived waits,
configurable HTTP error classification, credential management, output retention,
metrics, and tracing remain future work.
