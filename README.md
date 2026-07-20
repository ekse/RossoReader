# Rosso Reader

A self-hosted RSS feed reader. Periodically fetches RSS/Atom feeds on a configurable schedule, stores articles in PostgreSQL, and provides a clean web UI for browsing, searching, and managing your subscriptions.

<img src="logos/rosso_reader_transparent.png" width="600"/>

![screenshot](logos/Screenshot_Rosso%20Reader.png)

### Quick start (dev setup)

Docker and docker-compose are needed to run Rosso Reader. To try Rosso Reader locally, the development server can be started with `docker compose`. **This setup is not recommended for real use, see the Production section below.**

```bash
docker compose up
```

Rosso Reader will now be available on http://localhost:5173/.

### Production

Use `docker-compose.prod.yml` to deploy and installation of Rosso Reader.

```bash
# Set database credentials
export POSTGRES_USER=rssreader
export POSTGRES_PASSWORD=your-secure-password
export DOMAIN=yourdomain.com
export LETSENCRYPT_EMAIL=you@email.com

# Start all services
docker compose -f docker-compose.prod.yml up -d
```

Access Rosso Reader on https://yourdomain.com/.

### Environment Variables

#### Backend (Go application)

| Variable                 | Default                                                                   | Description                     |
| ------------------------ | ------------------------------------------------------------------------- | ------------------------------- |
| `DATABASE_URL`           | `postgres://postgres:postgres@localhost:5433/rssreader?sslmode=disable`   | PostgreSQL connection string    |
| `SERVER_PORT`            | `8080`                                                                    | Backend HTTP listen port        |
| `FETCH_INTERVAL_MINUTES` | `30`                                                                      | RSS feed fetch interval         |
| `CORS_ORIGIN`            | `http://localhost:5173`                                                   | Allowed CORS origin             |
| `RP_ID`                  | `localhost`                                                               | WebAuthn RP ID                  |
| `RP_ORIGIN`              | `http://localhost:5173`                                                   | WebAuthn RP origin              |
| `SESSION_MAX_AGE_DAYS`   | `30`                                                                      | Session max age in days         |
| `SESSION_COOKIE_NAME`    | `session`                                                                 | Session cookie name             |
| `COOKIE_SECURE`          | `false`                                                                   | Set cookie `Secure` flag        |

#### Frontend (Vite build-time)

| Variable       | Default                 | Description              |
| -------------- | ----------------------- | ------------------------ |
| `VITE_API_URL` | `http://localhost:8080` | Backend URL for API calls |

#### Docker Compose host port mappings

| Variable   | Default          | Description                   |
| ---------- | -----------------| ----------------------------- |
| `DB_PORT`  | `5433` (dev)     | Host port mapped to PostgreSQL |
| `API_PORT` | `8081` (dev)     | Host port mapped to the Go API |
| `WEB_PORT` | `5173` (dev)     | Host port mapped to the frontend |

#### Production (docker-compose.prod.yml)

| Variable             | Default    | Required | Description                           |
| -------------------- | ---------- | -------- | ------------------------------------- |
| `POSTGRES_USER`      | `postgres` | No       | PostgreSQL user                       |
| `POSTGRES_PASSWORD`  | `postgres` | **Yes**  | PostgreSQL password                   |
| `POSTGRES_DB`        | `rssreader`| No       | PostgreSQL database name              |
| `DOMAIN`             | –          | **Yes**  | Domain for Traefik routing & TLS      |
| `RP_ORIGIN`          | `https://${DOMAIN}:443` | No   | WebAuthn RP origin |
| `LETSENCRYPT_EMAIL`  | –          | **Yes**  | Email for Let's Encrypt ACME          |
| `FETCH_INTERVAL_MINUTES` | `30`   | No       | RSS feed fetch interval               |
| `CORS_ORIGIN`         | –          | No       | Allowed CORS origin                   |
| `SESSION_MAX_AGE_DAYS` | `30`      | No       | Session max age in days               |
| `SESSION_COOKIE_NAME`  | `session` | No       | Session cookie name                   |
| `COOKIE_SECURE`        | `true`    | No       | Set cookie `Secure` flag              |

## Development Setup

### Starting the development containers

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
