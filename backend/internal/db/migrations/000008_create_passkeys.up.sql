CREATE TABLE passkeys (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    credential_id    BYTEA NOT NULL UNIQUE,
    public_key       BYTEA NOT NULL,
    attestation_type TEXT NOT NULL DEFAULT '',
    transports       TEXT[] NOT NULL DEFAULT '{}',
    sign_count       BIGINT NOT NULL DEFAULT 0,
    backup_eligible  BOOLEAN NOT NULL DEFAULT false,
    backup_state     BOOLEAN NOT NULL DEFAULT false,
    aaguid           BYTEA,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_passkeys_user_id ON passkeys(user_id);
CREATE INDEX idx_passkeys_credential_id ON passkeys(credential_id);

CREATE TABLE passkey_auth_state (
    id         UUID PRIMARY KEY,
    state_type TEXT NOT NULL,
    state_data JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_passkey_auth_state_expires ON passkey_auth_state(expires_at);
