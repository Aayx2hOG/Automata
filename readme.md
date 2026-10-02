# Automata

A production-grade, self-hosted workflow automation platform, combining ideas from Temporal (durable workflow execution), n8n (visual workflow building), and GitHub Actions (execution history and logs).

Automata lets users define workflows as directed acyclic graphs connecting triggers, APIs, conditions, and custom logic, then execute and observe those workflows reliably.

## Status

This project is under active development. Current progress by phase:

| Phase | Scope | Status |
|---|---|---|
| 1 | Foundation: config, database, authentication, CI | Complete |
| 2 | Workflow engine: graph model, validation, execution | Complete |
| 3 | Core nodes (HTTP, Delay, Logger, Condition, Webhook, JSON Parser, Manual) | Complete |
| 4 | Durable execution: PostgreSQL outbox, concurrent consumers, retries | Complete |
| 5 | Scheduler & Triggers: cron scheduler, webhook triggers | Complete |
| 6 | Frontend: persistent visual editor, execution history, schedules, signed webhooks | Core flow implemented |
| 7 | Integrations: GitHub, Discord, Slack, SMTP | Not started |
| 8 | Production readiness: execution security implemented; metrics, tracing, secrets management pending | In progress |

## Tech stack

**Backend**
- Go 1.24+
- Chi (HTTP router)
- PostgreSQL via pgx / pgxpool
- JWT authentication with Argon2id password hashing
- Zap (structured logging)
- Robfig Cron (cron expression parsing)
- Go-playground Validator (request validation)

**Frontend**
- Bun runtime & package manager
- React 19 + Vite + TypeScript
- `@xyflow/react` (React Flow v12) visual node DAG editor
- Tailwind CSS v4, Lucide Icons, Glassmorphism design system

**Infrastructure**
- Docker & Docker Compose
- GitHub Actions (CI)

## Architecture

Workflows are stored as directed acyclic graphs: a set of nodes, each performing one action, connected by edges that may carry branching conditions. Every node implements a common interface:

```go
type Node interface {
    Execute(ec *ExecutionContext) (map[string]interface{}, error)
}
```

New node types are added by implementing this interface and registering them in the registry — the execution engine itself never needs to change. Graphs are validated for cycles and dangling references before execution, and every run is persisted with its status, outputs, and timing.

### Durable execution

Manual, webhook, and cron triggers atomically persist a workflow run and a PostgreSQL outbox entry. Webhook payloads are stored with the job, and each run uses the workflow version selected when it was created. Scheduled occurrences also advance their schedule in the same transaction, with a unique occurrence key preventing duplicate run records.

Each server starts a dispatcher and bounded worker pool independently of cron schedules. The local queue contains durable run IDs; workers use short `FOR UPDATE SKIP LOCKED` transactions to claim renewable leases, then release the database connection before executing. Defaults are four workers and 64 queued IDs (`WORKER_CONCURRENCY`, `WORKER_QUEUE_CAPACITY`). Workflow errors receive up to two retries with 2-second and 4-second delays, within a shared 60-second execution timeout.

Runs become visibly `running` when claimed. Leases last 30 seconds and renew every 10 seconds; lease loss cancels execution, and a stale worker cannot acknowledge another worker's job. Shutdown releases unfinished jobs, while crashes leave them reclaimable after lease expiry. Execution remains **at least once**: external side effects may repeat, so downstream actions should support idempotency. Missed cron intervals are skipped rather than replayed.

Workflow creation and its initial version commit atomically. Persisted webhook payloads become the webhook trigger's output. Graph joins wait for every incoming branch to complete or be skipped, then execute once if at least one incoming edge is active; entirely inactive branches propagate their skipped state.

See [durable execution details](backend/docs/scheduled-execution.md) for transaction semantics and operational limits.

Docker Compose includes daily PostgreSQL backups and continuous WAL archiving.
Local backups use a separate Docker volume; production requires configuring
off-server storage. See [database recovery](backend/docs/database-recovery.md)
for setup, monitoring, restore instructions and outage retry behavior.

### Execution security

- Webhooks require a per-workflow HMAC-SHA256 signature and a timestamp within five minutes of server time. Request bodies are limited to 1 MiB.
- HTTP request nodes accept public HTTP(S) destinations without URL credentials. DNS results and redirect destinations are validated, connections use validated IPs, and environment proxies are not used. Internal service URLs are blocked, and response bodies are limited to 4 MiB after decompression.
- Refresh-token rotation revokes the old token and creates its replacement atomically. Revoked, expired, or concurrently reused tokens cannot rotate again, and inactive users cannot refresh.

See [execution security details](backend/docs/execution-security.md) for webhook signing and compatibility changes.

## Getting started

### Prerequisites

- Go 1.24 or later
- PostgreSQL 16 or later (or Docker to run it in a container)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Setup

Clone the repository and move into the project:

```bash
git clone <repository-url>
cd automata
```

Create the shared local environment file:

```bash
cp .env.example .env
```

Generate a secret:

```bash
openssl rand -base64 32
```

Paste the output into `.env` as the value for `JWT_SECRET`.

### Frontend development

See [frontend setup](frontend/README.md) for the reusable Docker development database, API and frontend startup commands, and browser tests.

### Database

Start PostgreSQL with the shared root environment:

```bash
docker compose -f compose.dev.yml up -d --wait postgres
docker compose -f compose.dev.yml run --rm migrate
```

Stop older server instances, apply migrations through `000011_outbox_leases`, then start the updated server. Older consumers do not honor leases and must not run alongside the new workers. Migration `000009_schedule_outbox` adds durable dispatch and schedule occurrence identity; `000010_execution_security` adds persisted webhook payloads and independent webhook secrets; `000011_outbox_leases` adds renewable execution ownership.

### Running the server

```bash
cd backend
go run ./cmd/server
```

The server starts on the port configured in `.env` (default `8080`). Verify it is running:

```bash
curl http://localhost:8080/healthz
```

## API overview

### Authentication

| Method | Path | Description |
|---|---|---|
| POST | `/auth/register` | Create a new user account |
| POST | `/auth/login` | Authenticate and receive tokens |
| POST | `/auth/refresh` | Rotate a refresh token for a new access token |
| POST | `/auth/logout` | Revoke a refresh token |

### Users

| Method | Path | Description |
|---|---|---|
| GET | `/users/me` | Return the authenticated user (requires a Bearer token) |

### Workflows & Execution

| Method | Path | Description |
|---|---|---|
| POST | `/workflows` | Create a new workflow |
| GET | `/workflows` | List workflows owned by the authenticated user |
| GET | `/workflows/{id}` | Get workflow details and active version |
| PUT | `/workflows/{id}` | Save metadata and publish a new graph version atomically |
| POST | `/workflows/{id}/run` | Manually trigger a workflow run |
| GET | `/runs/{id}` | Get workflow execution status and outputs |
| GET | `/runs?limit=50&offset=0` | List the authenticated owner's runs, newest first (maximum limit 100) |

### Schedules (Cron Triggers)

| Method | Path | Description |
|---|---|---|
| POST | `/workflows/{id}/schedules` | Create a new cron schedule for a workflow |
| GET | `/workflows/{id}/schedules` | List active/inactive schedules for a workflow |
| DELETE | `/schedules/{scheduleID}` | Deactivate a schedule |

### Webhook Triggers

| Method | Path | Description |
|---|---|---|
| POST | `/webhook/{workflowID}` | Trigger a workflow with a signed JSON webhook payload |

Retrieve `webhook_secret` from an authenticated workflow create, get, or list response. Send these headers with each webhook request:

| Header | Value |
|---|---|
| `X-Webhook-Timestamp` | Unix timestamp in seconds, within ±300 seconds of server time |
| `X-Webhook-Signature` | Hex-encoded HMAC-SHA256 of `timestamp + "." + exact_request_body`, using the workflow's secret string as the key |

Use the secret string directly, without hex-decoding it, and sign the exact bytes sent as the body. Unsigned or invalidly signed requests receive `401`; bodies larger than 1 MiB receive `413`. Valid signatures can be replayed within the timestamp window, so deduplicate event IDs when repeated delivery must not repeat business effects.

### Health

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Liveness check, including database connectivity |

Authenticated requests carry a JWT in the `Authorization` header: `Authorization: Bearer <access_token>`.

## Testing

Run these commands from `backend/`. Unit tests and HTTP/webhook security tests (no external dependencies):

```bash
go test ./tests/unit/... ./internal/nodes/http_request ./internal/services -v
```

Integration tests (require a running PostgreSQL instance with migrations applied):

```bash
export $(grep -v '^#' .env | xargs)
go test ./tests/integration/... -v
```

Durable execution regression tests use a separate test database and create and remove isolated schemas:

```bash
TEST_DATABASE_URL=postgres://... go test ./internal/repositories -v
```

These cover transactional rollback, competing schedulers and consumers, interrupted consumption, and terminal failure acknowledgement. They are skipped when `TEST_DATABASE_URL` is unset; API integration tests are skipped when `DATABASE_URL` is unset. Set both variables to include both database suites in a full run.

Run everything with race detection and coverage:

```bash
go test ./... -race -cover
```

## Continuous integration

Every push and pull request runs `go vet`, `go build`, and the full test suite against a real PostgreSQL service container via GitHub Actions. See `.github/workflows/ci.yml`.

## Project structure

```
backend/
    cmd/            entry points (server, migrate)
    docs/           durable execution and security documentation
    internal/
        api/        HTTP handlers, router, server lifecycle
        auth/       JWT, password hashing, refresh token logic
        config/     configuration loading
        database/   connection pool, migrations
        middleware/ auth, RBAC, logging, rate limiting
        models/     domain types and typed errors
        nodes/      individual workflow node implementations
        repositories/ database access layer
        scheduler/  cron scheduling and durable outbox consumption
        services/   business logic
        telemetry/  structured logging
        workflow/   graph validation, execution engine, node registry
    tests/
        unit/       fast, dependency-free tests
        integration/ tests against a real database
docker-compose.yml  local PostgreSQL service
frontend/           React + Vite visual workflow editor
```
