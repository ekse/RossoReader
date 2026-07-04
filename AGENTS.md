# AGENTS.md

## Project

RSS feed reader (Rosso Reader). Go backend (`backend/`), Vue 3 frontend (`frontend/`), PostgreSQL storage.

## Making changes

Changes must include unit tests.

Use constants for default or constant values. Do not duplicate constants, centraize them to avoid the values becoming desynchronized.

## Build & Run

```bash
# Development (Docker, recommended)
docker compose up -d
# API: localhost:8081, Web: localhost:5173

# Backend (local)
cd backend && go build -o server ./cmd/server && ./server

# Frontend (local)
cd frontend && pnpm install && pnpm dev
```

## Code Generation

After changing SQL queries in `backend/sql/queries/`:

```bash
cd backend && sqlc generate -f sql/sqlc.yaml
```

Output goes to `backend/internal/db/generated/`.

## Tests

### Backend

Tests live alongside source in `*_test.go` files.

| Layer          | Location                            | Type                          |
| -------------- | ----------------------------------- | ----------------------------- |
| Handlers       | `backend/internal/handlers/`        | Unit (mockstore)              |
| Middleware     | `backend/internal/middleware/`      | Unit                          |
| Scheduler      | `backend/internal/scheduler/`       | Unit                          |
| Fetcher        | `backend/internal/fetcher/`         | Unit (httptest)               |
| OPML           | `backend/internal/opml/`            | Unit                          |
| Auth/passwords | `backend/internal/auth/`            | Unit                          |
| Store (PG)     | `backend/internal/store/pgstore/`   | Integration (testcontainers)  |
| Store (mock)   | `backend/internal/store/mockstore/` | In-memory fake for unit tests |

```bash
# Unit tests (fast)
cd backend && go test -short ./...

# Integration tests (requires Docker)
cd backend && go test ./internal/store/pgstore/

# All tests
cd backend && go test ./...
```

New handler/feature tests go in `backend/internal/handlers/`. Use `mockstore.New()` to create test store instances. See existing `handler_test.go` for test setup patterns.

### Frontend

Tests live in `__tests__/` directories next to the code they test.

| Layer         | Location                              |
| ------------- | ------------------------------------- |
| Stores        | `frontend/src/stores/__tests__/`      |
| Components    | `frontend/src/components/__tests__/`  |
| Composables   | `frontend/src/composables/__tests__/` |
| Lib utilities | `frontend/src/lib/*.test.ts`          |

```bash
cd frontend && pnpm test          # Single run
cd frontend && pnpm test:watch    # Watch mode
cd frontend && pnpm coverage      # With coverage
```

Use vitest + @vue/test-utils + happy-dom. See `vitest.config.ts` and existing tests for patterns.

## Lint & Format

```bash
cd frontend && pnpm lint          # oxlint
cd frontend && pnpm fmt:check     # oxfmt check
cd frontend && pnpm fmt           # oxfmt fix
```

The backend has no lint/format tooling configured.

## Architecture

- **Backend**: Chi router, Store interface (65 methods) with `pgstore` (PostgreSQL via sqlc) and `mockstore` (in-memory) implementations. Dependency injection wired in `cmd/server/main.go`. Middleware sets user on request context from session cookie. All data scoped to `user_id`.
- **Frontend**: Vue 3 + Pinia (stores) + Vue Router + Tailwind CSS (dark mode via `class`). API client at `src/api/client.ts`. Composable functions in `src/composables/` for reusable logic.
- **Multi-user**: Session-based auth (UUID cookie), argon2id passwords, WebAuthn passkeys. Admin user bootstrapped on first run.
