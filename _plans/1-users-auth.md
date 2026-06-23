# Users, Authentication & Sessions

## Overview

Add multi-user support with username/password authentication, server-side sessions (HttpOnly cookie, 30-day expiry), per-user ownership of feeds/settings, and per-user item read/starred state. Bootstrap an `admin` user with a random password on first run (logged to API logs). Admin user management and self-service password change in the Settings page.

No new Go libraries. `golang.org/x/crypto` (argon2id) and `github.com/google/uuid` are already indirect deps — both promoted to direct.

## Decisions (confirmed)

- Password hashing: **argon2id**.
- Settings scope: **per-user**.
- Login identifier: **username**.
- Feeds: **keep separate per user** (no shared-feed model).
- User creation: **admin only**.
- Session cleanup: on every `/api/auth/me` call (`DeleteExpiredSessions`).
- Drop `items.read`/`items.starred` columns in bootstrap backfill (after migrating states).

## Phase 1 — Database Migrations

New files in `backend/internal/db/migrations/`:

### 000005_create_users.up.sql / .down.sql
```sql
CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  username      TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  is_admin      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 000006_create_sessions.up.sql / .down.sql
```sql
CREATE TABLE sessions (
  id         UUID PRIMARY KEY,
  user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

### 000007_add_user_ownership.up.sql / .down.sql
- Add `user_id BIGINT REFERENCES users(id) ON DELETE CASCADE` (nullable) to `feeds`, `settings`.
- Drop `feeds_url_key`, add `UNIQUE(user_id, url)`.
- Drop `settings_pkey` (key), add `PRIMARY KEY(user_id, key)`.
- Create `user_item_states(user_id, item_id, read BOOLEAN, starred BOOLEAN, PK(user_id,item_id))` with FKs to users+items.

### Bootstrap (in Go, after migrations run)
1. `CountUsers == 0` → create admin user (username=admin, is_admin=true, random argon2id password). Log username + password.
2. `UPDATE feeds SET user_id = (SELECT id FROM users LIMIT 1) WHERE user_id IS NULL`
3. `UPDATE settings SET user_id = (SELECT id FROM users LIMIT 1) WHERE user_id IS NULL`
4. `INSERT INTO user_item_states SELECT 1, id, read, starred FROM items WHERE read OR starred ON CONFLICT DO NOTHING`
5. `ALTER TABLE items DROP COLUMN read, DROP COLUMN starred` (raw exec).
6. Log: `Created admin user — username: admin password: <random>`

## Phase 2 — Backend

### 2.1 Password hashing (`internal/auth/password.go`, new package)
- `HashPassword(password) (string, error)` — argon2id: memory=64MiB, iterations=3, parallelism=2, salt=16B, key=32B; standard `$argon2id$v=19$m=...,t=...,p=...$salt$hash` encoding.
- `VerifyPassword(encoded, password) (bool, error)` — decode params, recompute, constant-time compare.

### 2.2 Domain models (`internal/domain/models.go`)
- `User{ID, Username, IsAdmin, CreatedAt, UpdatedAt}` (no PasswordHash in JSON).
- `Session` for internal use if needed.
- Add `UserID int64` to `Feed` and `ItemsQuery`.

### 2.3 sqlc queries (`backend/sql/queries/`)
- New `users.sql`: `CreateUser`, `GetUserByUsername`, `GetUserByID`, `ListUsers`, `UpdateUserPassword`, `DeleteUser`, `CountUsers`.
- New `sessions.sql`: `CreateSession`, `GetSessionWithUser` (JOIN users), `DeleteSession`, `DeleteExpiredSessions`.
- Modify `feeds.sql`: all queries take `user_id`; `AddFeed` includes `user_id`; delete/get/refresh filter by `user_id`.
- Modify `items.sql`: `ListItems` LEFT JOIN `user_item_states` for caller's read/starred; `MarkItemRead`/`MarkItemStarred`/`MarkAllItemsRead`/`MarkFeedRead` become UPSERTs into `user_item_states`. `ItemsQuery` gains `UserID`.
- Modify `settings.sql`: all queries take `user_id`; composite key.
- Run `sqlc generate -f sql/sqlc.yaml`.

### 2.4 Store (`internal/store/store.go`, pgstore, mockstore)
- Add user/session methods to interface.
- Modify feed/item/settings signatures to accept `userID int64`.
- pgstore wraps generated code; converts generated/domain.
- mockstore: add `Users`, `Sessions` slices + counters; store `passwordHash` internally.
- Update `pgstore_test.go` inline schema with new tables (or refactor to use `RunMigrations`).

### 2.5 Auth handlers (`internal/handlers/auth.go`, new)
- `POST /api/auth/login` — body `{username, password}`; verify; create session UUID; set HttpOnly `session` cookie (SameSite=Lax, Secure when `COOKIE_SECURE=true`, MaxAge=30d, Path=/); return user JSON.
- `POST /api/auth/logout` — delete session; clear cookie.
- `GET /api/auth/me` — `DeleteExpiredSessions`, then return current user from context.
- `PUT /api/auth/password` — body `{current_password, new_password}`; verify current; update hash.
- `POST /api/users` (admin only) — body `{username, password, is_admin}`.
- `GET /api/users` (admin only) — list users (no hashes).
- `DELETE /api/users/:id` (admin only) — delete user.

### 2.6 Auth middleware (`internal/middleware/auth.go`, new)
- `Authenticate(store)` — read `session` cookie → lookup session+user (not expired) → set user in context. On failure: 401.
- `RequireAdmin` — check `IsAdmin`; else 403.
- `UserFromContext(ctx) (domain.User, bool)`.
- Update `CORS`: add `Access-Control-Allow-Credentials: true`.

### 2.7 Router wiring (`internal/handlers/handlers.go`, `cmd/server/main.go`)
- Pass admin/secret/config into `Handlers`.
- Routes:
  - Public: `/api/auth/login`, `/api/auth/logout`, `/api/health`.
  - Authenticated group: `/api/feeds`, `/api/items`, `/api/settings`, `/api/auth/me`, `/api/auth/password`.
  - Admin group: `/api/users` (GET, POST), `/api/users/:id` (DELETE).
- Existing handlers read `UserID` from context and pass to all store calls.
- Run bootstrap after migrations.

### 2.8 Scheduler (`internal/scheduler/scheduler.go`)
- Fetching stays global; item inserts no longer set `read`/`starred` (gone). New items default to unread/unstarred via LEFT JOIN on `user_item_states`.

### 2.9 Config (`cmd/server/main.go`, docker-compose files)
- New env: `COOKIE_SECURE` (dev false / prod true), `SESSION_MAX_AGE_DAYS` (30), `SESSION_COOKIE_NAME` (`session`).
- Add to `docker-compose.yml` and `docker-compose.prod.yml`.

## Phase 3 — Frontend

### 3.1 API client (`frontend/src/api/client.ts`)
- `withCredentials: true` on axios instance.
- `401` interceptor → redirect `/login`.
- New endpoints: login, logout, getMe, changePassword, listUsers, createUser, deleteUser.

### 3.2 Auth store (`frontend/src/stores/auth.ts`, new)
- State `user`, `loading`. Actions: `fetchMe`, `login`, `logout`. Getters `isAuthenticated`, `isAdmin`.

### 3.3 Login view (`frontend/src/views/Login.vue`, new)
- Username + password form → `auth.login()` → redirect `/unread`. Themed.

### 3.4 Router guards (`frontend/src/router/index.ts`)
- Add `/login`.
- `beforeEach`: if not authed and route not `/login` → `/login`. If authed and `/login` → `/unread`.
- `auth.fetchMe()` before app mount in `main.ts`.

### 3.5 Settings page (`frontend/src/views/Settings.vue`)
- **Account**: change-password form (current, new, confirm).
- **Administration** (visible if `auth.user.is_admin`): user table + create-user form + delete (not self/last admin).

### 3.6 FeedList sidebar
- Show current username near ThemeToggle; route to `/settings` on click.

## Phase 4 — Testing

- Password hash/verify round-trip + wrong password.
- Auth handlers (mockstore): login success/failure, logout, me, change password, admin create user (non-admin 403), list users.
- Middleware: missing cookie / expired / valid.
- Existing handler tests: inject `UserID` into context.
- pgstore integration: add new tables to inline schema.
- Frontend: auth store, Login, Settings admin section.