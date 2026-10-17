# Automata backend guide

This document describes the backend implemented in this repository, including the
workflow deletion endpoint. For architectural diagrams, consistency guarantees,
and design tradeoffs, see [System design](SYSTEM_DESIGN.md).

## 1. What the backend does

Automata is a workflow automation service. Users save directed acyclic graphs
(DAGs), trigger them manually, deliver signed webhook events, or attach cron
schedules. The service validates the graphs, admits execution into a durable
PostgreSQL queue, runs their nodes, and persists results for inspection.

The application is one Go executable containing the HTTP API, scheduler, outbox
dispatcher, and worker pool. PostgreSQL is required. Redis, Kafka, and a separate
message broker are not required. The React frontend is a separate application.

Important behavior:

- Editing creates a new workflow version; admitted runs keep their original version.
- Successful node outputs and consumed attempts survive process restarts.
- Execution is sequential within a graph, with concurrent workflows across workers.
- Incoming webhook requests are authenticated, but separate deliveries are not deduplicated.
- External side effects are not guaranteed to happen exactly once.
- Deleting a workflow removes its versions, schedules, and history, but is blocked while it has unfinished durable work.

## 2. Technology and source map

The Go version and exact dependency versions are declared in [go.mod](go.mod).
The main libraries are Chi for routing, pgx for PostgreSQL, Viper for configuration,
Zap for logging, robfig/cron for scheduling, the Clerk Go SDK, JWT v5, and the
Go crypto packages.

| Path | Responsibility |
|---|---|
| `cmd/server/main.go` | Dependency wiring, startup checks, background loops, shutdown |
| `internal/config` | Environment/file configuration, defaults, validation |
| `internal/api/server.go` | HTTP listener and server timeouts |
| `internal/api/router.go` | Routes, authentication mode, CORS, middleware |
| `internal/api/handlers` | Request validation, status codes, response serialization |
| `internal/services` | Workflow operations, webhook verification, authentication |
| `internal/repositories` | SQL, transactions, admission, outbox leases, checkpoints |
| `internal/database` | Connection pool and numbered SQL migrations |
| `internal/models` | Workflow/user data structures and domain errors |
| `internal/workflow` | DAG validation, traversal, mappings, node registry, retries |
| `internal/node` | Node interfaces, execution context, state and retry contracts |
| `internal/nodes` | Concrete trigger, HTTP, parsing, condition, delay, and logging nodes |
| `internal/queue` | Bounded in-memory queue of durable run IDs |
| `internal/worker` | Dispatcher and concurrent worker pool |
| `internal/scheduler` | Due schedule polling and next-occurrence calculation |
| `internal/auth` | Argon2id password hashing, JWTs, refresh-token helpers |
| `internal/credentials` | Owner-scoped encrypted header credential vault |
| `internal/middleware` | Authentication, IP throttling, body limits, logging, proxy trust |
| `internal/telemetry` | Logger construction |
| `tests/unit`, `tests/integration` | Cross-package unit and database-backed API tests |
| `scripts/verify_execution.py` | Execution/recovery verification tooling |

## 3. Startup and shutdown

Startup loads and validates configuration, initializes logging, and connects to
PostgreSQL. It checks that migration 15's checkpoint fields exist. Clerk mode also
checks for migration 14's user mapping column. Startup does not apply migrations.

The server then constructs repositories, authentication services, the node
registry, execution engine, workflow service, and optional credential vault.
It starts the worker pool, one dispatcher loop, one scheduler loop, and the HTTP
listener. A failure to bind the HTTP port initiates shutdown and exits unsuccessfully.

On SIGINT or SIGTERM, the service stops accepting HTTP work, cancels background
execution, closes the local queue, and waits for background loops and workers
within the shutdown deadline. Unfinished durable work remains recoverable from
PostgreSQL. Forced termination can leave leases that become reclaimable on expiry.

Source: [entry point](cmd/server/main.go).

## 4. Running locally

From the repository root:

```bash
cp .env.example .env
# Set real local values in .env; do not commit secrets.
docker compose -f compose.dev.yml up -d --wait postgres
docker compose -f compose.dev.yml run --rm migrate
cd backend
go run ./cmd/server
```

Use `curl http://localhost:8080/healthz` to check connectivity. The frontend dev
server proxies `/api` to port 8080. Compose starts the database and provides a
migration command; it does not start the Go server or React frontend.

After changing Go code, restart `go run ./cmd/server` or rebuild and restart your
binary. An already-running process continues executing its old code. A frontend
refresh cannot add a route to an outdated backend.

## 5. Configuration

Process environment variables take precedence over loaded dotenv files. The loader
reads local configuration and applies environment overrides. See
[the loader](internal/config/config.go) for the exact file search order and
[the example configuration](internal/config/config.example.yaml) for file syntax.
Never copy deployment secrets into this document.

| Variable | Default / requirement | Purpose |
|---|---|---|
| `DATABASE_URL` | Required | Database used by this process; migrations must target the same DB |
| `JWT_SECRET` | Required, minimum 32 characters | Legacy access-token signing |
| `ENV` | `development` | Logging/runtime environment |
| `HTTP_PORT` | `8080` | HTTP listener port |
| `HTTP_ALLOWED_ORIGINS` | Empty | Explicit permitted browser origins |
| `HTTP_TRUSTED_PROXIES` | Empty | Trusted proxy CIDRs for forwarded client addresses |
| `HTTP_SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown deadline |
| `DATABASE_MAX_CONNS` | `10` | Pool size per application process |
| `WORKER_CONCURRENCY` | `4` | Parallel workflow executions per process |
| `WORKER_QUEUE_CAPACITY` | `64` | Local queue size; not the durable backlog limit |
| `EXECUTION_MAX_PENDING_RUNS` | `10000` | Global outstanding outbox limit |
| `EXECUTION_MAX_PENDING_PER_USER` | `100` | Outstanding outbox limit per owner |
| `EXECUTION_RUNS_PER_MINUTE` | `60` | Rolling per-owner run admission limit |
| `AUTH_ACCESS_TOKEN_TTL` | `15m` | Legacy JWT lifetime |
| `AUTH_REFRESH_TOKEN_TTL` | `168h` | Legacy refresh-token lifetime |
| `CLERK_SECRET_KEY` | Optional | Enables Clerk authentication on protected API routes |
| `CLERK_AUTHORIZED_PARTIES` | Local frontend origins | Permitted Clerk token authorized parties |
| `CREDENTIALS_ENCRYPTION_KEY` | Optional, 64 hex characters | AES-256-GCM credential encryption key |

Replicas must share admission settings and credential keys. An empty CORS origin
list does not permit arbitrary cross-origin browser requests. The credential API
returns `503` if the vault is not configured. The JWT setting is still validated
by configuration loading when Clerk is enabled.

## 6. Authentication and authorization

### Clerk mode

With `CLERK_SECRET_KEY`, protected routes verify Clerk session tokens and their
authorized party. A valid subject and session ID are required. Local accounts
are looked up by `clerk_user_id`, not by an email supplied by a caller.

First-time mapping fetches the Clerk profile and requires a verified primary
email. New mappings receive the normal user role. Existing local accounts are not
automatically merged by matching email; migration must be deliberate.

The frontend uses Clerk's sign-in/sign-up components. Password, email-code, and
Google authentication options are controlled by the Clerk instance. An initial
email field is not itself backend authorization. The API requires a verified token.

### Legacy JWT mode

Without Clerk configuration, protected routes use the local JWT middleware.
Registration validates a username, email, and password, then stores an Argon2id
password hash. Login verifies the supplied password before issuing tokens.
Refresh tokens are stored hashed, rotated atomically, and checked for revocation,
expiry, and active user status. Logout revokes a refresh token; an already-issued
access token remains valid until expiry.

The `/auth/*` routes remain registered in Clerk mode, but their legacy tokens do
not replace Clerk authentication on protected routes.

### Resource ownership

Workflow and run operations enforce the authenticated owner's identity. Access to
another owner's workflow is generally represented as `404`, avoiding resource
existence disclosure. Credential access is also owner-scoped. An RBAC middleware
exists, but the current application is primarily a per-owner workspace, not a
shared-team permissions system. The API-key and audit-log tables do not imply
implemented public API-key authentication or comprehensive audit capture.

## 7. HTTP API reference

All paths below are backend paths without the frontend `/api` proxy prefix.
Protected endpoints require a bearer token in the configured authentication mode.

| Method | Path | Behavior |
|---|---|---|
| GET | `/healthz` | Process/database connectivity; not proof that workers are progressing |
| POST | `/auth/register` | Local account creation with username, email, password |
| POST | `/auth/login` | Local password verification |
| POST | `/auth/refresh` | Rotate a legacy refresh token |
| POST | `/auth/logout` | Revoke a legacy refresh token |
| GET | `/users/me` | Resolve current application user |
| GET | `/workflows` | List caller's workflows |
| POST | `/workflows` | Create workflow and initial graph version atomically |
| GET | `/workflows/{id}` | Workflow metadata and active graph version |
| PUT | `/workflows/{id}` | Update metadata and publish another graph version |
| DELETE | `/workflows/{id}` | Owner-only permanent deletion; `204` on success |
| POST | `/workflows/{id}/run` | Admit a manual run and return its run record |
| GET | `/runs?limit=50&offset=0` | Paginated owner history; maximum page size 100 |
| GET | `/runs/{id}` | Run result and persisted node execution details |
| GET, POST | `/workflows/{id}/schedules` | List schedules or create a cron schedule |
| DELETE | `/schedules/{scheduleID}` | Deactivate, rather than delete, a schedule |
| POST | `/webhook/{workflowID}` | Authenticate and enqueue webhook payload using HMAC |
| GET, POST | `/credentials` | List metadata or create an encrypted header credential |
| PUT, DELETE | `/credentials/{id}` | Replace or delete an owned credential |

A create/update workflow body contains `name`, `description`, and `graph`.
A schedule body contains `cron_expression`. Create-workflow responses currently
use `Workflow` and `Version` keys; workflow detail responses use `workflow` and
`active_version`. Clients should follow the actual contract, not assume uniform casing.

Validation failures generally return `400`; missing/invalid authentication returns
`401`; oversized bodies return `413`; owner admission limits return `429`; global
backlog exhaustion returns `503`. Admission-limit responses carry `Retry-After`.
Graph errors include structured node validation details.

Deletion returns `409` when queued/running executions or unacknowledged outbox
entries remain. The transaction checks ownership, locks the workflow, coordinates
with schedule locks, checks outstanding work, and deletes the parent. Foreign-key
cascades remove child versions, runs, checkpoints, schedules, and outbox rows.
A deleted workflow's webhook URL stops accepting new runs. External actions that
already occurred are not undone; shared credentials are not deleted with a workflow.

## 8. Database model and migration history

| Table | Role |
|---|---|
| `users` | Local identity, role, active state, password hash, optional Clerk subject |
| `refresh_tokens` | Hashed refresh credentials and revocation/expiry state |
| `api_keys` | Schema foundation; no current API-key middleware wired into routing |
| `audit_logs` | Schema foundation; not a complete implemented audit trail |
| `workflows` | Owner, name, description, active flag, webhook signing secret |
| `workflow_versions` | Versioned JSON graph snapshots |
| `workflow_runs` | Version-pinned execution, trigger, status, timestamps, result/error |
| `schedules` | Cron expression, active state, last/next occurrence |
| `schedule_outbox` | Durable jobs for all triggers, payload seed, lease token/expiry |
| `workflow_node_runs` | Per-node attempts, checkpoints, errors, retry metadata |
| `credentials` | Owner-scoped host/header metadata and encrypted secret values |

Migrations 1–8 introduce the core identity/workflow/run/schedule tables.
Migration 9 adds the transactional schedule outbox; 10 adds webhook secrets and
payload seeds; 11 adds leases; 12 adds usernames; 13 adds node execution state;
14 adds Clerk subjects; 15 adds the credential vault and inspection/retry fields.
All migration SQL is in [internal/database/migrations](internal/database/migrations).

Apply all migrations before startup. The startup schema probes are targeted
compatibility checks, not a replacement for migration version management.
Do not mix incompatible old and new workers. Rollbacks of checkpoint/credential
migrations destroy the data introduced by those migrations.

## 9. Graphs, nodes, and execution

Graphs contain nodes with stable IDs, type, config, optional visual position,
input mappings, and retry policy; edges reference source/target IDs and may carry
branch conditions. Validation rejects cycles, invalid references, oversized graphs,
and invalid policies. Limits are 100 nodes, 500 edges, and 255 bytes per node ID.

Registered node types:

| Type | Purpose |
|---|---|
| `manual_trigger` | Manual entry into a graph |
| `webhook_trigger` | Entry that receives persisted webhook seed data |
| `http_request` | Outbound public HTTP(S) request with optional credential reference |
| `json_parser` | JSON parsing/transformation into usable output values |
| `condition` | Evaluate a predicate for conditional edges |
| `delay` | Wait while respecting execution context cancellation |
| `logger` | Emit configured logging output |

The engine resolves nodes in dependency order. A join waits until every incoming
edge resolves and executes if at least one incoming edge is active. Skipped
branches still resolve edges so joins do not hang. Each graph runs sequentially.

Input mappings reference an upstream node and a path through its output. They
preserve JSON types and override mapped top-level configuration fields. Missing
fields or skipped sources fail the consuming node before its external action.
Credential IDs and HTTP error policies cannot be dynamically mapped.

A worker execution has a 60-second deadline. Recovery starts a fresh execution
deadline, so this is not a strict end-to-end lifetime limit. Delay nodes and retry
waits occupy a worker slot.

## 10. Durable jobs, schedules, and retries

Admission commits the run and outbox row together. A shared transaction-scoped
advisory lock serializes backlog/quota checks across processes. Counts include
running jobs still present in the outbox; the rolling quota includes completed
runs within the time window.

The dispatcher polls every second and copies eligible IDs into a bounded local
queue. A worker claims a job after dequeue using a short transaction, a random
lease token, and a 30-second lease. Heartbeats renew approximately every ten
seconds. Stale tokens cannot checkpoint, acknowledge, or release newer claims.
Database operations in the outbox machinery have five-second deadlines.

Before a node runs, its attempt is recorded. After success, its output is committed
before advancing. Recovery reuses completed outputs and attempt budgets. Final
acknowledgement stores the result and removes the durable job atomically.
Checkpoint-storage failures leave durable work available for recovery.

Cron uses five fields: minute, hour, day-of-month, month, day-of-week. Scheduling
polls immediately on startup and then every 30 seconds. Claiming an occurrence,
creating its run/outbox, and advancing the next time are transactional. A unique
schedule/occurrence key prevents duplicate run records for one occurrence.
Missed intervals are skipped by computing the next occurrence from current time.

Retries default to one attempt. Explicit policies allow 1–5 total attempts and
0–30 seconds fixed backoff. Only the failing node repeats. Fixed backoff is not a
fully persisted durable timer; HTTP `Retry-After` deadlines are persisted.

HTTP defaults to `error_policy: "output"`: non-2xx responses become outputs with
`success=false`. With `"fail"`, 408, 429, and 5xx are retryable; other non-2xx
responses are permanent failures. Retry budgets still apply. A stable engine-owned
idempotency key is sent for each run/node, but the destination must honor it.

## 11. Security controls and boundaries

- Webhooks require a Unix timestamp and hex HMAC-SHA256 signature over
  `timestamp + "." + exact_body`, using the workflow secret string as the key.
  The timestamp window is five minutes; valid deliveries can be replayed within it.
- Request bodies are capped at 1 MiB. Auth and webhook IP rate limits are local to
  each process; durable run admission limits are shared through PostgreSQL.
- Forwarded IP headers are trusted only from configured proxy CIDRs.
- Outbound HTTP validates DNS answers and destination addresses, including redirects,
  rejects private/internal targets and URL credentials, bypasses environment proxies,
  and limits decompressed response bodies to 4 MiB.
- The credential vault uses AES-256-GCM, binding ciphertext to owner and credential
  IDs. Credentials are restricted to an exact HTTPS host/port and cannot follow redirects.
- Credential APIs return metadata only. Selected-secret strings and supported token
  representations are redacted from HTTP response bodies before downstream use.
- Ordinary graph config, payloads, outputs, and webhook secrets are not encrypted at
  the application layer. Redaction is not a general sensitive-data scanner.
- Production TLS termination, database access policy, secret storage, and backups
  are deployment responsibilities.

## 12. Tests and operational checks

Run fast checks from `backend/`:

```bash
go test ./tests/unit ./internal/workflow ./internal/queue ./internal/worker \
  ./internal/config ./internal/middleware ./internal/nodes/http_request \
  ./internal/services ./internal/api ./internal/api/handlers
go vet ./...
go build ./...
```

For full persistence coverage, provision a dedicated test database, apply migrations,
and set `DATABASE_URL`, `TEST_DATABASE_URL`, and a test-only `JWT_SECRET` before:

```bash
go test ./... -race -cover
```

Repository fixtures create isolated schemas; integration tests write users into the
configured database. Do not point them at production. Unset database variables cause
suites to skip: a green summary alone does not prove transaction/recovery coverage.

CI also exercises a logical backup/restore smoke test and frontend/browser checks.
Clerk/Google sign-in requires separate configured authentication testing. Browser
regressions need Chromium system libraries as well as its downloaded executable.

For troubleshooting, start with health, process version, schema compatibility,
request status/error, and the run's persisted node details. A healthy database check
does not establish that the dispatcher, scheduler, or workers are making progress.

## 13. Current limitations and maintenance

There is no implemented general run-cancellation API, retention service, distributed
rate-limit store, OAuth connection manager, streaming execution log service, metrics
exporter, or tracing pipeline. API-key and audit tables remain foundations rather
than completed features. Graph parallelism and long-lived durable waits are not implemented.

Backup both PostgreSQL and the credential encryption key, storing the key separately.
A database backup without its encryption key cannot recover stored credential values.
Restore drills and off-host storage must be configured by the deployment operator.

Further references:

- [System design](SYSTEM_DESIGN.md)
- [Execution and recovery semantics](docs/scheduled-execution.md)
- [Execution security](docs/execution-security.md)
- [Database recovery](docs/database-recovery.md)
- [Project setup and deployment](../readme.md)

This guide contains implementation documentation, not deployment secrets, and is
intended to be version-controlled alongside the backend.
