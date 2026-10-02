# Automata frontend

React + TypeScript + React Flow, backed by the Go API. Requires Node.js 24+ and Docker Compose v2.

## Start the development database

From the repository root:

```sh
docker compose -f compose.dev.yml up -d --wait postgres
docker compose -f compose.dev.yml run --rm migrate
```

Use `sudo docker` if your account cannot access the Docker socket. This development database listens on `127.0.0.1:5433` and keeps its data in a named volume. Database credentials and backend settings are shared through the repository-root `.env` file.

```sh
# Stop without removing data
docker compose -f compose.dev.yml stop
# Start it again
docker compose -f compose.dev.yml up -d --wait postgres
# Open a SQL shell
docker compose -f compose.dev.yml exec postgres psql -U automata -d automata_dev
```

## Start the API

The backend reads the repository-root `.env` file automatically.

Generate a JWT secret with `openssl rand -hex 32` and set `JWT_SECRET` in `.env`. Then:

```sh
cd backend
go run ./cmd/server
```

## Start the frontend

In another terminal:

```sh
cd frontend
npm ci
npm run dev
```

Open http://localhost:5173 and create an account. The development server proxies `/api` to port 8080. For a deployed frontend, configure a reverse proxy for `/api` or set `VITE_API_BASE_URL` at build time.

Create a **Blank Canvas** workflow to try the complete flow without external services: edit, save, execute, inspect outputs, and reload. Schedules are configured from the Schedules tab, not as graph nodes. Webhook testing signs the exact JSON body using the workflow's secret; generated shell commands prompt for that secret.

The execution screen shows the latest 50 runs and refreshes automatically. Per-node outputs are shown after execution; streaming node logs are not implemented.

## Checks

```sh
npm run build
npm run lint
npx playwright install chromium
# Requires the API and migrated database running:
npm run test:e2e
```

Playwright starts Vite automatically. The browser test uses a fresh account and workflow, checks persistence and successful execution, schedules, signed webhooks, and failed-save feedback. It leaves its test data in the database.

For database-backed Go tests, run from `backend/` against the development database:

```sh
TEST_DATABASE_URL='postgres://automata:automata_dev_password@localhost:5433/automata_dev?sslmode=disable' go test ./internal/repositories -race
```
