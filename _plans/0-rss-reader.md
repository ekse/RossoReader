# RSS Reader — Implementation Plan

## Technology Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22+, `chi` router, `pgx` driver |
| Queries | `sqlc` (type-safe code generation from SQL) |
| Migrations | `golang-migrate/migrate` |
| RSS Parser | `gofeed` |
| Scheduling | `robfig/cron` (default: every 30 min, env-configurable) |
| Database | PostgreSQL (dev and prod) |
| Frontend | Vue 3 (Composition API + `<script setup>`), Vite, Tailwind CSS |
| State | Pinia |
| Routing | Vue Router 4 |
| Package Manager | pnpm |
| Go Tests | `testing` + `stretchr/testify`, `testcontainers-go` (integration) |
| Frontend Tests | `vitest` + `@vue/test-utils` + `happy-dom` |
| Dev Env | VS Code Devcontainers (Docker Compose) |
| Prod | Docker Compose with multi-stage builds |

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Docker Compose (dev / prod)                        │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐   │
│  │ Vue 3    │  │  Go API  │  │  PostgreSQL      │   │
│  │ (Vite)   │─▶│  Server  │──│  Database        │   │
│  │ :5173    │  │  :8080   │  │  :5432           │   │
│  └──────────┘  └──────────┘  └──────────────────┘   │
│                   │                                 │
│                   ▼                                 │
│           ┌──────────────┐                          │
│           │ Scheduler    │                          │
│           │ (background) │                          │
│           │ fetches RSS  │                          │
│           │ every N min  │                          │
│           └──────────────┘                          │
└─────────────────────────────────────────────────────┘
```

---

## Project Structure

```
RssReader/
├── .devcontainer/
│   ├── devcontainer.json
│   └── docker-compose.dev.yml
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── domain/
│   │   │   └── models.go           # Feed, Item, Settings structs (no deps)
│   │   ├── store/
│   │   │   ├── store.go            # Store interface (all DB operations)
│   │   │   ├── pgstore/
│   │   │   │   ├── pgstore.go      # Postgres + sqlc implementation
│   │   │   │   └── pgstore_test.go # Integration tests (testcontainers)
│   │   │   └── mockstore/
│   │   │       └── mockstore.go    # In-memory mock for unit tests
│   │   ├── db/
│   │   │   ├── db.go               # pgx connection pool
│   │   │   ├── migrate.go          # golang-migrate runner
│   │   │   └── migrations/
│   │   │       ├── 000001_create_feeds.up.sql
│   │   │       ├── 000001_create_feeds.down.sql
│   │   │       ├── 000002_create_items.up.sql
│   │   │       ├── 000002_create_items.down.sql
│   │   │       ├── 000003_create_settings.up.sql
│   │   │       └── 000003_create_settings.down.sql
│   │   ├── handlers/
│   │   │   ├── feeds.go
│   │   │   ├── feeds_test.go       # Unit tests (mockstore)
│   │   │   ├── items.go
│   │   │   ├── items_test.go
│   │   │   ├── settings.go
│   │   │   ├── settings_test.go
│   │   │   └── handlers.go         # Shared handler setup, router
│   │   ├── fetcher/
│   │   │   ├── fetcher.go          # Fetcher interface + gofeed impl
│   │   │   └── fetcher_test.go     # Unit tests (httptest server)
│   │   ├── scheduler/
│   │   │   ├── scheduler.go
│   │   │   └── scheduler_test.go   # Unit tests (mock store + mock fetcher)
│   │   └── middleware/
│   │       ├── middleware.go
│   │       └── middleware_test.go
│   ├── sql/
│   │   ├── sqlc.yaml
│   │   └── queries/
│   │       ├── feeds.sql
│   │       ├── items.sql
│   │       └── settings.sql
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── main.ts
│   │   ├── App.vue
│   │   ├── api/
│   │   │   └── client.ts
│   │   ├── router/
│   │   │   └── index.ts
│   │   ├── stores/
│   │   │   ├── feeds.ts
│   │   │   ├── items.ts
│   │   │   └── settings.ts
│   │   ├── views/
│   │   │   ├── AllItems.vue
│   │   │   ├── FeedItems.vue
│   │   │   ├── StarredItems.vue
│   │   │   └── Settings.vue
│   │   └── components/
│   │       ├── FeedList.vue
│   │       ├── ItemList.vue
│   │       ├── ItemDetail.vue
│   │       └── AddFeedDialog.vue
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── postcss.config.js
│   ├── tsconfig.json
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
└── README.md
```

---

## Database Schema

```sql
-- Feeds table (RSS sources the user subscribes to)
CREATE TABLE feeds (
    id              SERIAL PRIMARY KEY,
    url             TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    description     TEXT,
    site_link       TEXT,
    last_fetched_at TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Items (individual articles/posts from feeds)
CREATE TABLE items (
    id           SERIAL PRIMARY KEY,
    feed_id      INTEGER NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
    guid         TEXT NOT NULL,
    title        TEXT NOT NULL,
    url          TEXT NOT NULL,
    content      TEXT,
    description  TEXT,
    author       TEXT,
    published_at TIMESTAMPTZ,
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read         BOOLEAN NOT NULL DEFAULT FALSE,
    starred      BOOLEAN NOT NULL DEFAULT FALSE,
    UNIQUE(feed_id, guid)
);

-- Indexes
CREATE INDEX idx_items_feed_id ON items(feed_id);
CREATE INDEX idx_items_published_at ON items(published_at DESC);
CREATE INDEX idx_items_read ON items(read);
CREATE INDEX idx_items_starred ON items(starred);

-- Settings (key-value pairs for app configuration)
CREATE TABLE settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
```

---

## API Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/api/feeds` | List all feed subscriptions |
| POST | `/api/feeds` | Add a new feed (body: `{url}`) |
| DELETE | `/api/feeds/:id` | Remove a feed subscription |
| POST | `/api/feeds/:id/refresh` | Force-refresh a single feed |
| GET | `/api/items` | List items (query: `?page=&per_page=&feed_id=&read=&starred=`) |
| PATCH | `/api/items/:id` | Update item (body: `{read, starred}`) |
| GET | `/api/settings` | Get all settings |
| PATCH | `/api/settings` | Update settings (body: `{key: value}`) |
| GET | `/api/health` | Health check |

---

## Testing Strategy

### Architecture for Testability

Dependency injection via interfaces is used throughout. Every component depends on interfaces, not concrete types, enabling easy mocking in unit tests and real implementations in integration tests.

```go
// internal/store/store.go — central data access interface
type Store interface {
    // Feeds
    GetFeeds(ctx context.Context) ([]domain.Feed, error)
    GetFeed(ctx context.Context, id int64) (domain.Feed, error)
    CreateFeed(ctx context.Context, url, title, description, siteLink string) (domain.Feed, error)
    DeleteFeed(ctx context.Context, id int64) error
    UpdateFeedLastFetched(ctx context.Context, id int64) error

    // Items
    GetItems(ctx context.Context, opts domain.ItemsQuery) ([]domain.Item, int64, error)
    GetItem(ctx context.Context, id int64) (domain.Item, error)
    CreateOrUpdateItem(ctx context.Context, item domain.Item) (domain.Item, error)
    MarkItemRead(ctx context.Context, id int64, read bool) error
    MarkItemStarred(ctx context.Context, id int64, starred bool) error

    // Settings
    GetSettings(ctx context.Context) (map[string]string, error)
    SetSetting(ctx context.Context, key, value string) error
}

// internal/fetcher/fetcher.go — RSS fetching interface
type Fetcher interface {
    Fetch(ctx context.Context, url string) (*gofeed.Feed, error)
}
```

sqlc generates a `Querier` interface via `emit_interface: true` which `pgstore` wraps to implement `Store`:

```yaml
# sql/sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "queries/"
    schema: "../internal/db/migrations/"
    gen:
      go:
        package: "generated"
        out: "../internal/db/generated"
        sql_package: "pgx/v5"
        emit_interface: true
        emit_json_tags: true
        emit_pointers_for_null_types: true
```

Startup wiring in `cmd/server/main.go`:

```go
pool := db.Connect(databaseURL)
db.RunMigrations(pool)

store := pgstore.New(pool)             // implements Store
fetcher := fetcher.NewHTTPFetcher()     // implements Fetcher
sched := scheduler.New(store, fetcher, cron)
handler := handlers.New(store, sched)
```

### Test Coverage Plan

| Test Type | Tool | What It Covers |
|---|---|---|
| **Go unit tests** | `testing` + `stretchr/testify` | Handlers (with mock store), scheduler (mock store + mock fetcher), fetcher (httptest server), middleware |
| **Go integration tests** | `testcontainers-go` + Postgres module | Real Postgres container → run migrations → test pgstore end-to-end (CRUD, filtering, pagination) |
| **Frontend unit tests** | `vitest` + `@vue/test-utils` + `happy-dom` | Component rendering, Pinia store logic, API client |
| **Frontend e2e** (future) | Playwright | Full browser workflows |

### Go Test Examples

**Handler unit test (mock store):**

```go
func TestGetFeeds(t *testing.T) {
    store := mockstore.New()
    store.Feeds = []domain.Feed{{ID: 1, Title: "Example Blog", URL: "https://example.com/rss"}}
    h := handlers.New(store, nil)

    req := httptest.NewRequest("GET", "/api/feeds", nil)
    w := httptest.NewRecorder()
    h.Router().ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
    var feeds []domain.Feed
    json.Unmarshal(w.Body.Bytes(), &feeds)
    assert.Len(t, feeds, 1)
    assert.Equal(t, "Example Blog", feeds[0].Title)
}
```

**Scheduler unit test (mock store + mock fetcher):**

```go
func TestScheduler_FetchAll(t *testing.T) {
    store := mockstore.New()
    store.Feeds = []domain.Feed{{ID: 1, URL: "https://example.com/rss"}}
    fetcher := &mockFetcher{
        feed: &gofeed.Feed{Title: "Test", Items: []*gofeed.Item{{Title: "Post"}}},
    }
    s := scheduler.New(store, fetcher, nil)

    s.FetchAll(context.Background())

    assert.Len(t, store.Items, 1)
    assert.Equal(t, "Post", store.Items[0].Title)
}
```

**Integration test (testcontainers + pgstore):**

```go
func TestStore_FeedCRUD(t *testing.T) {
    ctx := context.Background()
    pgContainer, _ := postgres.Run(ctx, "postgres:16-alpine",
        postgres.WithDatabase("rssreader"),
        postgres.WithUsername("postgres"),
        postgres.WithPassword("postgres"),
    )
    defer pgContainer.Terminate(ctx)

    connStr, _ := pgContainer.ConnectionString(ctx)
    store, _ := pgstore.New(ctx, connStr)
    store.MigrateUp()

    feed, err := store.CreateFeed(ctx, "https://example.com/rss", "Example", "", "")
    require.NoError(t, err)
    assert.Equal(t, "Example", feed.Title)

    feeds, err := store.GetFeeds(ctx)
    require.NoError(t, err)
    assert.Len(t, feeds, 1)
}
```

### Frontend Test Examples

```ts
// stores/feeds.test.ts
import { setActivePinia, createPinia } from 'pinia'
import { useFeedsStore } from './feeds'
import { describe, it, expect, beforeEach } from 'vitest'

describe('useFeedsStore', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('starts with an empty feed list', () => {
    const store = useFeedsStore()
    expect(store.feeds).toEqual([])
  })
})
```

```ts
// components/ItemList.test.ts
import { mount } from '@vue/test-utils'
import { describe, it, expect } from 'vitest'
import ItemList from './ItemList.vue'

describe('ItemList', () => {
  it('renders a list of items', () => {
    const wrapper = mount(ItemList, { props: { items: [{ id: 1, title: 'Test' }] } })
    expect(wrapper.text()).toContain('Test')
  })
})
```

---

## Implementation Phases

### Phase 1 — Project Scaffolding

- Create monorepo directory structure
- `.devcontainer/devcontainer.json`:
  - Docker Compose-based
  - Services: `app` (Go + Node dev container)
  - VS Code extensions: Go, Vue Language Features, Tailwind CSS IntelliSense, ESLint, Prettier
  - Forward ports: 5173 (Vite), 8080 (API), 5432 (Postgres)
- `docker-compose.yml` for development:
  - `api` — Go backend with hot-reload (air or CompileDaemon)
  - `web` — Vite dev server
  - `db` — PostgreSQL 16
- `backend/go.mod` initialized with `chi`, `pgx`, `gofeed`, `robfig/cron`, `golang-migrate`, `testify`, `testcontainers-go`
- `frontend/` initialized with Vue 3 + TypeScript + Vite + Router + Pinia + Tailwind, then add `vitest`, `@vue/test-utils`, `happy-dom`

### Phase 2 — Database & Migrations

- Create `internal/domain/models.go` — pure struct definitions (`Feed`, `Item`, `Settings`, `ItemsQuery`) with no dependencies
- Write golang-migrate SQL migration files for feeds, items, settings
- Write sqlc query definitions in `sql/queries/` for feeds, items, settings
- Configure `sqlc.yaml` with `emit_interface: true` so sqlc generates a `Querier` interface for mocking
- Run `sqlc generate` to produce `internal/db/generated/` Go code

### Phase 3 — Backend Core

- **Store interface** (`internal/store/store.go`):
  - Define `Store` interface with all data access methods (feeds CRUD, items CRUD with filtering/pagination, settings)
- **Domain models** (`internal/domain/models.go`) already created in Phase 2:
  - `Feed`, `Item`, `Settings` structs, `ItemsQuery` for filter/pagination params
- **Postgres store** (`internal/store/pgstore/pgstore.go`):
  - Wraps the sqlc-generated `Querier` interface to implement `Store`
  - Maps between domain models and generated types
- **Mock store** (`internal/store/mockstore/mockstore.go`):
  - In-memory implementation of `Store` using slices/maps
  - Used by all handler and scheduler unit tests
  - `New()` constructor, call `.Reset()` between tests
- **Store integration tests** (`internal/store/pgstore/pgstore_test.go`):
  - Spin up Postgres via testcontainers
  - Run migrations, test every Store method (CRUD, pagination, filters)
- **DB connection** (`internal/db/db.go`):
  - Connect to Postgres via `pgx` pool
  - Env vars: `DATABASE_URL`
- **Migration runner** (`internal/db/migrate.go`):
  - Run `golang-migrate` on startup using embedded migration files (`//go:embed`)
- **Fetcher** (`internal/fetcher/fetcher.go`):
  - Define `Fetcher` interface: `Fetch(ctx, url) (*gofeed.Feed, error)`
  - `HTTPFetcher` struct implements it using `gofeed`
  - Parse entries, map to domain items
- **Fetcher unit tests** (`internal/fetcher/fetcher_test.go`):
  - Spin up `httptest.Server` serving a static RSS XML
  - Verify `Fetch` parses title, items, links correctly
- **Scheduler** (`internal/scheduler/scheduler.go`):
  - Depends on `Store` and `Fetcher` interfaces (not concrete types)
  - Use `robfig/cron` with interval from `FETCH_INTERVAL_MINUTES` env var (default `30`)
  - `FetchAll(ctx)`: iterate feeds → fetch each → upsert items → log errors per-feed
- **Scheduler unit tests** (`internal/scheduler/scheduler_test.go`):
  - Mock `Store` with preset feeds, mock `Fetcher` returning known RSS data
  - Call `FetchAll(ctx)`, assert items were upserted correctly
- **HTTP handlers** (`internal/handlers/`):
  - `handlers.go` — shared setup: `Handler` struct holding `Store` + `Scheduler`, creates `chi.Router`
  - `feeds.go` — GET/POST/DELETE `/api/feeds`, POST `/api/feeds/:id/refresh`
  - `items.go` — GET `/api/items` (paginated, filterable), PATCH `/api/items/:id`
  - `settings.go` — GET/PATCH `/api/settings`
- **Handler unit tests** (`internal/handlers/*_test.go`):
  - Each handler tested with `mockstore` and `httptest`
  - Test all endpoints: success cases, empty states, not found, validation errors
- **Middleware** (`internal/middleware/middleware.go`):
  - CORS (allow configured origin from env)
  - Request logging (method, path, status, duration)
  - JSON content-type
- **Middleware unit tests** (`internal/middleware/middleware_test.go`):
  - Verify CORS headers, JSON content-type, logging output
- **Entry point** (`cmd/server/main.go`):
  - Load config from env
  - Connect DB, run migrations
  - Wire: `pgstore` → `Store` interface, `HTTPFetcher` → `Fetcher` interface
  - Create scheduler with injected dependencies
  - Build chi router via handlers, apply middleware
  - Start HTTP server on `:8080`

### Phase 4 — Frontend

- **Layout**: Sidebar (FeedList + navigation) + main content area
- **Router** (`src/router/index.ts`):
  ```
  /              → AllItems.vue
  /feed/:id      → FeedItems.vue
  /starred       → StarredItems.vue
  /settings      → Settings.vue
  ```
- **API client** (`src/api/client.ts`):
  - Fetch wrapper with base URL from env
  - Typed functions for each endpoint
- **Stores** (Pinia):
  - `useFeedsStore` — feed list, add/remove feeds
  - `useItemsStore` — paginated items, mark read/starred, filter state
  - `useSettingsStore` — fetch interval, app preferences
- **Views**:
  - `AllItems.vue` — paginated list of all items, latest first, filter tabs (all/unread/starred)
  - `FeedItems.vue` — items filtered by feed ID
  - `StarredItems.vue` — all starred items
  - `Settings.vue` — feed management table + fetch interval setting
- **Components**:
  - `FeedList.vue` — sidebar list of feeds with unread count badges, active highlight
  - `ItemList.vue` — list of article cards, click to read, buttons for read/starred toggle
  - `ItemDetail.vue` — expandable article reader with full content
  - `AddFeedDialog.vue` — modal with URL input, validates and adds feed

### Phase 5 — Deployment

- **Backend Dockerfile**:
  - Stage 1: `golang:1.22-alpine` — build Go binary
  - Stage 2: `alpine:3.20` — copy binary, run migrations on startup
- **Frontend Dockerfile**:
  - Stage 1: `node:22-alpine` + pnpm — build Vite app
  - Stage 2: `nginx:alpine` — serve static files, proxy `/api` to backend
- **docker-compose.prod.yml**:
  - `db` — PostgreSQL with volume for data persistence, healthcheck
  - `api` — Go backend, depends on db, restart always
  - `web` — nginx reverse proxy, depends on api, restart always
  - Environment variables for production config
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | `postgres://postgres:postgres@db:5432/rssreader?sslmode=disable` | PostgreSQL connection string |
| `SERVER_PORT` | `8080` | Backend HTTP port |
| `FETCH_INTERVAL_MINUTES` | `30` | RSS fetch interval |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed CORS origin (dev) |
| `VITE_API_URL` | `http://localhost:8080` | Backend URL for frontend |
```

---

## Dev Workflow

```bash
# Start all services
docker compose up -d

# Backend hot-reload (air)
cd backend && air

# Frontend dev server
cd frontend && pnpm dev

# Run migrations manually
go run cmd/server/main.go -migrate

# Generate sqlc code
cd backend && sqlc generate

# Run all tests
cd backend && go test ./...
cd frontend && pnpm test

# Run tests with coverage
cd backend && go test -coverprofile=coverage.out ./...
cd frontend && pnpm coverage

# Run only unit tests (skip integration)
cd backend && go test -short ./...

# Run only integration tests
cd backend && go test -run Integration ./internal/store/pgstore/

# Frontend test watch mode
cd frontend && pnpm test:watch
```
