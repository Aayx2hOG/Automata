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
| 4 | Worker system: job queue, worker pool, concurrency, retries | Complete |
| 5 | Scheduler & Triggers: cron scheduler, webhook triggers | Complete |
| 6 | Frontend: visual workflow editor | Complete |
| 7 | Integrations: GitHub, Discord, Slack, SMTP | Not started |
| 8 | Production readiness: metrics, tracing, secrets | In progress |

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
| POST | `/webhook/{workflowID}` | Trigger a workflow via external HTTP webhook payload |

### Health

| Method | Path | Description |
|---|---|---|
| GET | `/healthz` | Liveness check, including database connectivity |

Authenticated requests carry a JWT in the `Authorization` header: `Authorization: Bearer <access_token>`.

## Testing

Unit tests (pure logic, no external dependencies):

```bash
go test ./tests/unit/... -v
```

Integration tests (require a running PostgreSQL instance with migrations applied):

```bash
export $(grep -v '^#' .env | xargs)
go test ./tests/integration/... -v
```

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
    internal/
        api/        HTTP handlers, router, server lifecycle
        auth/       JWT, password hashing, refresh token logic
        config/     configuration loading
        database/   connection pool, migrations
        middleware/ auth, RBAC, logging, rate limiting
        models/     domain types and typed errors
        nodes/      individual workflow node implementations
        repositories/ database access layer
        services/   business logic
        telemetry/  structured logging
        workflow/   graph validation, execution engine, node registry
    tests/
        unit/       fast, dependency-free tests
        integration/ tests against a real database
docker/             Dockerfile and Compose configuration
docs/               architecture and design documentation
frontend/           planned Next.js application
```
