# RssReader

A self-hosted RSS feed reader with a Go backend and Vue 3 frontend.

Periodically fetches RSS/Atom feeds on a configurable schedule, stores articles in PostgreSQL, and provides a clean web UI for browsing, searching, and managing your subscriptions.

![rosso_logo](logos/rosso_reader_transparent.png)

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
| `SESSION_MAX_AGE_DAYS` | `30` | Session max age in days |
| `SESSION_COOKIE_NAME` | `session` | Session cookie name |
| `COOKIE_SECURE` | `false` | Set cookie Secure flag |
| `DB_PORT` | `5433` (dev) / `5432` (prod) | Host port mapped to PostgreSQL |
| `API_PORT` | `8081` (dev) / `8080` (prod) | Host port mapped to the Go API |
| `WEB_PORT` | `5173` (dev) / `80` (prod) | Host port mapped to the frontend |

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

### Quick start

```bash
# Start PostgreSQL and run the app stack
docker compose up -d
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
