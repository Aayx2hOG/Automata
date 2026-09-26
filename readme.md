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
| 6 | Frontend: visual workflow editor | Complete |
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

Each server runs one outbox consumer through the scheduler, even when no cron schedules exist. Consumers use `FOR UPDATE SKIP LOCKED` to coordinate across instances. Workflow errors receive up to two retries with 2-second and 4-second delays, within a shared 60-second execution timeout. Shutdown or connection loss rolls back unfinished consumption so it can be retried.

Execution is **at least once**: external side effects may repeat after a retry or crash, so downstream actions should support idempotency. The consumer holds a database connection and transaction during execution; runs remain visibly pending until the terminal result commits. Keep database idle-in-transaction timeouts longer than the execution timeout. Missed cron intervals are skipped rather than replayed.

See [durable execution details](backend/docs/scheduled-execution.md) for transaction semantics and operational limits.

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

Clone the repository and move into the backend module:

```bash
git clone <repository-url>
cd automata/backend
```

Copy the environment template and fill in a real JWT secret:

```bash
cp .env.example .env
```

Generate a secret:

```bash
openssl rand -base64 32
```

Paste the output into `.env` as the value for `JWT_SECRET`.

### Database

Start PostgreSQL (via the project's `docker-compose.yml` at the repository root, or your own instance):

```bash
docker compose up -d
```

Run migrations:

```bash
export $(grep -v '^#' .env | xargs)
migrate -path internal/database/migrations -database "$DATABASE_URL" up
```

Before starting the updated server, apply migrations through `000010_execution_security`. Migration `000009_schedule_outbox` adds durable dispatch and schedule occurrence identity; `000010_execution_security` adds persisted webhook payloads and generates independent webhook secrets for existing workflows.

### Running the server

```bash
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
| POST | `/workflows/{id}/run` | Manually trigger a workflow run |
| GET | `/runs/{id}` | Get workflow execution status and logs |

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
