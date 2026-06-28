-- name: CreatePasskey :one
INSERT INTO passkeys (user_id, name, credential_id, public_key, attestation_type, transports, sign_count, backup_eligible, backup_state, aaguid)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, user_id, name, credential_id, public_key, attestation_type, transports, sign_count, backup_eligible, backup_state, aaguid, created_at, updated_at;

-- name: GetPasskeysByUserID :many
SELECT id, user_id, name, credential_id, public_key, attestation_type, transports, sign_count, backup_eligible, backup_state, aaguid, created_at, updated_at
FROM passkeys
WHERE user_id = $1
ORDER BY created_at;

-- name: GetPasskeyByCredentialID :one
SELECT id, user_id, name, credential_id, public_key, attestation_type, transports, sign_count, backup_eligible, backup_state, aaguid, created_at, updated_at
FROM passkeys
WHERE credential_id = $1;

-- name: UpdatePasskeySignCount :exec
UPDATE passkeys
SET sign_count = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeletePasskey :exec
DELETE FROM passkeys
WHERE id = $1 AND user_id = $2;

-- name: SaveAuthState :exec
INSERT INTO passkey_auth_state (id, state_type, state_data, expires_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET state_type = $2, state_data = $3, expires_at = $4;

-- name: GetAuthState :one
SELECT state_type, state_data
FROM passkey_auth_state
WHERE id = $1 AND expires_at > NOW();

-- name: DeleteAuthState :exec
DELETE FROM passkey_auth_state
WHERE id = $1;

-- name: DeleteExpiredAuthStates :exec
DELETE FROM passkey_auth_state
WHERE expires_at < NOW();
