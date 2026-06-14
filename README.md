# RssReader

A self-hosted RSS feed reader with a Go backend and Vue 3 frontend.

Periodically fetches RSS/Atom feeds on a configurable schedule, stores articles in PostgreSQL, and provides a clean web UI for browsing, searching, and managing your subscriptions.

## Tech Stack

**Backend:** Go 1.22+, chi router, sqlc (type-safe queries), golang-migrate, pgx, gofeed, robfig/cron  
**Frontend:** Vue 3, Vite, Pinia, Vue Router, Tailwind CSS, pnpm  
**Database:** PostgreSQL 16  
**Testing:** testify, testcontainers-go (Go), vitest + @vue/test-utils + happy-dom (frontend)

## Usage

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/rssreader?sslmode=disable` | PostgreSQL connection string |
| `SERVER_PORT` | `8080` | Backend HTTP port |
| `FETCH_INTERVAL_MINUTES` | `30` | RSS fetch interval |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin (dev) |
| `VITE_API_URL` | `http://localhost:8080` | Backend URL for frontend |

### Production (Docker Compose)

```bash
# Set database credentials
export POSTGRES_USER=rssreader
export POSTGRES_PASSWORD=your-secure-password

# Start all services
docker compose -f docker-compose.prod.yml up -d

# The web UI is served on port 80
```

### API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/feeds` | List subscriptions |
| POST | `/api/feeds` | Add a feed `{"url": "..."}` |
| DELETE | `/api/feeds/:id` | Remove a feed |
| POST | `/api/feeds/:id/refresh` | Force-refresh a feed |
| GET | `/api/items` | List items (paginated, filterable) |
| PATCH | `/api/items/:id` | Mark read/starred |
| GET | `/api/settings` | Get app settings |
| PATCH | `/api/settings` | Update settings |
| GET | `/api/health` | Health check |

## Development Setup

### Prerequisites

- Docker & Docker Compose
- VS Code (recommended) with Dev Containers extension

### Quick start (Dev Containers)

```bash
# Open the project in VS Code
code .

# Reopen in container when prompted
# (or Cmd+Shift+P → "Dev Containers: Reopen in Container")
```

The dev container includes Go, Node.js, pnpm, and connects to a PostgreSQL service automatically. Ports 5173 (Vite), 8080 (API), and 5432 (Postgres) are forwarded.

### Manual start (without Dev Containers)

```bash
# Start PostgreSQL and run the app stack
docker compose up -d

# Backend hot-reload (standalone)
cd backend
go mod tidy
air

# Frontend dev server (standalone)
cd frontend
pnpm install
pnpm dev
```

### Running tests

```bash
# Backend unit tests
cd backend && go test -short ./...

# Backend integration tests (requires Docker)
cd backend && go test ./internal/store/pgstore/

# Backend all tests
cd backend && go test ./...

# Frontend tests
cd frontend && pnpm test
cd frontend && pnpm test:watch  # watch mode
```

### Code generation

```bash
# Regenerate sqlc code after changing queries
cd backend && sqlc generate -f sql/sqlc.yaml
```
