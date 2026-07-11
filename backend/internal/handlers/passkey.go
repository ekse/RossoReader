package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"

	"github.com/ekse/rossoreader/internal/domain"
	"github.com/ekse/rossoreader/internal/middleware"
	"github.com/ekse/rossoreader/internal/store"
)

type PasskeyHandler struct {
	Store    store.Store
	WebAuthn *webauthn.WebAuthn
}

func NewPasskeyHandler(s store.Store, wa *webauthn.WebAuthn) *PasskeyHandler {
	return &PasskeyHandler{Store: s, WebAuthn: wa}
}

func (h *PasskeyHandler) sessionMaxAge() int {
	if v := os.Getenv("SESSION_MAX_AGE_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultSessionMaxAgeDays
}

func (h *PasskeyHandler) cookieName() string {
	if v := os.Getenv("SESSION_COOKIE_NAME"); v != "" {
		return v
	}
	return defaultCookieName
}

func (h *PasskeyHandler) cookieSecure() bool {
	v := os.Getenv("COOKIE_SECURE")
	return v == "true" || v == "1"
}

type passkeyUser struct {
	user        domain.User
	credentials []webauthn.Credential
}

func (u *passkeyUser) WebAuthnID() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(u.user.ID))
	return b
}

func (u *passkeyUser) WebAuthnName() string {
	return u.user.Username
}

func (u *passkeyUser) WebAuthnDisplayName() string {
	return u.user.Username
}

func (u *passkeyUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func toWebAuthnCredentials(passkeys []domain.Passkey) []webauthn.Credential {
	creds := make([]webauthn.Credential, 0, len(passkeys))
	for _, pk := range passkeys {
		transports := make([]protocol.AuthenticatorTransport, len(pk.Transports))
		for i, t := range pk.Transports {
			transports[i] = protocol.AuthenticatorTransport(t)
		}
		creds = append(creds, webauthn.Credential{
			ID:              pk.CredentialID,
			PublicKey:       pk.PublicKey,
			AttestationType: pk.AttestationType,
			Transport:       transports,
			Flags: webauthn.CredentialFlags{
				BackupEligible: pk.BackupEligible,
				BackupState:    pk.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:    pk.AAGUID,
				SignCount: uint32(pk.SignCount),
			},
		})
	}
	return creds
}

// RegisterBegin starts the WebAuthn registration ceremony.
// POST /api/auth/passkey/register/begin (authenticated)
func (h *PasskeyHandler) RegisterBegin(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	passkeys, err := h.Store.GetPasskeysByUserID(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "failed to load passkeys", http.StatusInternalServerError)
		return
	}

	wuser := &passkeyUser{user: u, credentials: toWebAuthnCredentials(passkeys)}

	creation, sessionData, err := h.WebAuthn.BeginRegistration(wuser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			ResidentKey:      protocol.ResidentKeyRequirementRequired,
			UserVerification: protocol.VerificationPreferred,
		}),
	)
	if err != nil {
		http.Error(w, "failed to begin registration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		http.Error(w, "failed to encode session data", http.StatusInternalServerError)
		return
	}

	stateID := uuid.New()
	err = h.Store.SaveAuthState(r.Context(), stateID, "registration", sessionDataJSON, time.Now().Add(5*time.Minute))
	if err != nil {
		http.Error(w, "failed to save session state", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"state_id": stateID.String(),
		"options":  creation,
	})
}

type registerFinishRequest struct {
	StateID    string          `json:"state_id"`
	Name       string          `json:"name"`
	Credential json.RawMessage `json:"credential"`
}

// RegisterFinish completes the WebAuthn registration ceremony.
// POST /api/auth/passkey/register/finish (authenticated)
func (h *PasskeyHandler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req registerFinishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.StateID == "" || req.Name == "" || req.Credential == nil {
		http.Error(w, "state_id, name, and credential are required", http.StatusBadRequest)
		return
	}

	stateUUID, err := uuid.Parse(req.StateID)
	if err != nil {
		http.Error(w, "invalid state_id", http.StatusBadRequest)
		return
	}

	stateType, stateData, err := h.Store.GetAuthState(r.Context(), stateUUID)
	if err != nil {
		http.Error(w, "invalid or expired session state", http.StatusUnauthorized)
		return
	}
	defer h.Store.DeleteAuthState(r.Context(), stateUUID)

	if stateType != "registration" {
		http.Error(w, "invalid session state type", http.StatusBadRequest)
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(stateData, &sessionData); err != nil {
		http.Error(w, "failed to decode session data", http.StatusInternalServerError)
		return
	}

	if binary.BigEndian.Uint64(sessionData.UserID) != uint64(u.ID) {
		http.Error(w, "session state does not match user", http.StatusUnauthorized)
		return
	}

	passkeys, err := h.Store.GetPasskeysByUserID(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "failed to load passkeys", http.StatusInternalServerError)
		return
	}

	wuser := &passkeyUser{user: u, credentials: toWebAuthnCredentials(passkeys)}

	r.Body = io.NopCloser(bytes.NewReader(req.Credential))
	r.Header.Set("Content-Type", "application/json")

	cred, err := h.WebAuthn.FinishRegistration(wuser, sessionData, r)
	if err != nil {
		http.Error(w, "failed to verify credential: "+err.Error(), http.StatusBadRequest)
		return
	}

	transports := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transports[i] = string(t)
	}

	created, err := h.Store.CreatePasskey(r.Context(), u.ID, req.Name,
		cred.ID, cred.PublicKey, cred.AttestationType, transports,
		int64(cred.Authenticator.SignCount), cred.Flags.BackupEligible,
		cred.Flags.BackupState, cred.Authenticator.AAGUID,
	)
	if err != nil {
		http.Error(w, "failed to save passkey: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// LoginBegin starts the discoverable passkey login ceremony.
// POST /api/auth/passkey/login/begin (public)
func (h *PasskeyHandler) LoginBegin(w http.ResponseWriter, r *http.Request) {
	assertion, sessionData, err := h.WebAuthn.BeginDiscoverableLogin(
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		http.Error(w, "failed to begin login: "+err.Error(), http.StatusInternalServerError)
		return
	}

	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		http.Error(w, "failed to encode session data", http.StatusInternalServerError)
		return
	}

	stateID := uuid.New()
	err = h.Store.SaveAuthState(r.Context(), stateID, "authentication", sessionDataJSON, time.Now().Add(5*time.Minute))
	if err != nil {
		http.Error(w, "failed to save session state", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"state_id": stateID.String(),
		"options":  assertion,
	})
}

type loginFinishRequest struct {
	StateID    string          `json:"state_id"`
	Credential json.RawMessage `json:"credential"`
}

// LoginFinish completes the discoverable passkey login ceremony and creates a session.
// POST /api/auth/passkey/login/finish (public)
func (h *PasskeyHandler) LoginFinish(w http.ResponseWriter, r *http.Request) {
	var req loginFinishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.StateID == "" || req.Credential == nil {
		http.Error(w, "state_id and credential are required", http.StatusBadRequest)
		return
	}

	stateUUID, err := uuid.Parse(req.StateID)
	if err != nil {
		http.Error(w, "invalid state_id", http.StatusBadRequest)
		return
	}

	stateType, stateData, err := h.Store.GetAuthState(r.Context(), stateUUID)
	if err != nil {
		http.Error(w, "invalid or expired session state", http.StatusUnauthorized)
		return
	}
	defer h.Store.DeleteAuthState(r.Context(), stateUUID)

	if stateType != "authentication" {
		http.Error(w, "invalid session state type", http.StatusBadRequest)
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(stateData, &sessionData); err != nil {
		http.Error(w, "failed to decode session data", http.StatusInternalServerError)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(req.Credential))
	r.Header.Set("Content-Type", "application/json")

	waUser, cred, err := h.WebAuthn.FinishPasskeyLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			pk, err := h.Store.GetPasskeyByCredentialID(r.Context(), rawID)
			if err != nil {
				return nil, err
			}

			user, _, err := h.Store.GetUserByID(r.Context(), pk.UserID)
			if err != nil {
				return nil, err
			}

			wu := &passkeyUser{user: user}
			if !bytes.Equal(userHandle, wu.WebAuthnID()) {
				return nil, errors.New("user handle does not match user ID")
			}

			passkeys, err := h.Store.GetPasskeysByUserID(r.Context(), pk.UserID)
			if err != nil {
				return nil, err
			}
			wu.credentials = toWebAuthnCredentials(passkeys)

			return wu, nil
		},
		sessionData,
		r,
	)
	if err != nil {
		http.Error(w, "passkey verification failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	wu, ok := waUser.(*passkeyUser)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	pk, err := h.Store.GetPasskeyByCredentialID(r.Context(), cred.ID)
	if err == nil {
		h.Store.UpdatePasskeySignCount(r.Context(), pk.ID, int64(cred.Authenticator.SignCount))
	}

	maxAgeDays := h.sessionMaxAge()
	expiresAt := time.Now().Add(time.Duration(maxAgeDays) * 24 * time.Hour)
	sessionID := uuid.New()

	if err := h.Store.CreateSession(r.Context(), sessionID, wu.user.ID, expiresAt); err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName(),
		Value:    sessionID.String(),
		Path:     "/",
		MaxAge:   maxAgeDays * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   h.cookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, toUserResponse(wu.user))
}

// ListPasskeys returns the authenticated user's passkeys.
// GET /api/auth/passkeys (authenticated)
func (h *PasskeyHandler) ListPasskeys(w http.ResponseWriter, r *http.Request) {
	_ = h.Store.DeleteExpiredAuthStates(r.Context())

	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	passkeys, err := h.Store.GetPasskeysByUserID(r.Context(), u.ID)
	if err != nil {
		http.Error(w, "failed to load passkeys", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, passkeys)
}

// DeletePasskey deletes a passkey by ID (scoped to the authenticated user).
// DELETE /api/auth/passkeys/{id} (authenticated)
func (h *PasskeyHandler) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid passkey id", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeletePasskey(r.Context(), u.ID, id); err != nil {
		http.Error(w, "failed to delete passkey", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
