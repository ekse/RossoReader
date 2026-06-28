# Passkey (WebAuthn) Support

## Overview

Add WebAuthn passkey authentication as an alternative to password login. Users can register multiple named passkeys, login with them (without a password), list their passkeys, and delete them.

**Library**: `github.com/go-webauthn/webauthn` — the standard Go WebAuthn library.

Only **discoverable credentials** (resident keys) are supported for login — no username-first flow.

---

## Implementation Steps

### 1. Database — Migration 000008

**New tables**:
- `passkeys` — WebAuthn credentials per user
- `passkey_auth_state` — temporary ceremony state (5min TTL)

Includes up/down migrations, indexes.

### 2. Domain Models

Add `domain.Passkey` struct to `internal/domain/models.go`.

### 3. Store Interface

Add 9 passkey-related methods to `store.Store` interface:
- `CreatePasskey`, `GetPasskeysByUserID`, `GetPasskeyByCredentialID`, `UpdatePasskeySignCount`, `DeletePasskey`
- `SaveAuthState`, `GetAuthState`, `DeleteAuthState`, `DeleteExpiredAuthStates`

### 4. SQL Queries

New `sql/queries/passkeys.sql` with sqlc annotations. Run `sqlc generate`.

### 5. Store Implementations

- `pgstore` — wrap generated sqlc queries
- `mockstore` — in-memory maps for tests

### 6. WebAuthn Handler

New `internal/handlers/passkey.go`:
- `PasskeyHandler` struct (store, webauthn instance, cookie config)
- Adapter types implementing `webauthn.User` / `webauthn.Credential`
- 6 handler methods for endpoints

**Endpoints**:

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/auth/passkey/register/begin` | Required | Start registration |
| `POST` | `/api/auth/passkey/register/finish` | Required | Complete registration |
| `POST` | `/api/auth/passkey/login/begin` | Public | Start login (discoverable) |
| `POST` | `/api/auth/passkey/login/finish` | Public | Complete login → create session |
| `GET` | `/api/auth/passkeys` | Required | List user's passkeys |
| `DELETE` | `/api/auth/passkeys/{id}` | Required | Delete a passkey |

### 7. Route Wiring

Wire passkey routes into `handlers.Handler`, add `PasskeyHandler` to struct.

### 8. Server Startup

Create `webauthn.WebAuthn` instance in `main.go` with RP config from env vars. Create `PasskeyHandler`. Update `handlers.New()` signature.

### 9. Frontend — Utilities

New `frontend/src/lib/webauthn.ts` with base64url helpers and credential encoding/decoding functions.

### 10. Frontend — Types & API Client

Add `Passkey` interface to types. Add 6 passkey API functions.

### 11. Frontend — Login Page

Add "Sign in with passkey" button — calls discoverable flow, creates session, redirects.

### 12. Frontend — Settings Page

Add passkey management section — register, list, delete passkeys.

### 13. Docker

Add `RP_ID` and `RP_ORIGIN` env vars to both docker-compose files.

### 14. Tests

- Handler tests: `handlers/passkey_test.go` (mockstore)
- Store integration tests: `pgstore/passkey_pgstore_test.go` (testcontainers)
- Auth tests: extend `auth_test.go` with passkey login test
- Mockstore tests: extend with passkey methods coverage
- Frontend: `lib/webauthn.test.ts` for base64url helpers
