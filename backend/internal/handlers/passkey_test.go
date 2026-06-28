package handlers_test

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rssreader/internal/domain"
	"github.com/ekse/rssreader/internal/handlers"
	"github.com/ekse/rssreader/internal/store"
	"github.com/ekse/rssreader/internal/store/mockstore"
)

func TestPasskey_RegisterBegin(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("POST", "/api/auth/passkey/register/begin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		StateID string `json:"state_id"`
		Options any    `json:"options"`
	}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.StateID)
	assert.NotNil(t, resp.Options)
}

func TestPasskey_RegisterBegin_Unauthenticated(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := h.MountRouter()

	req := httptest.NewRequest("POST", "/api/auth/passkey/register/begin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPasskey_LoginBegin(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := h.Router()

	req := httptest.NewRequest("POST", "/api/auth/passkey/login/begin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		StateID string `json:"state_id"`
		Options any    `json:"options"`
	}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.StateID)
	assert.NotNil(t, resp.Options)
}

func TestPasskey_List_Empty(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("GET", "/api/auth/passkeys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]\n", w.Body.String())
}

func TestPasskey_List(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)

	pk, err := store.CreatePasskey(context.Background(), user.ID, "My Key",
		[]byte("cred-id"), []byte("pub-key"), "none", []string{"usb"}, 0, false, false, nil)
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("GET", "/api/auth/passkeys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var passkeys []domain.Passkey
	err = json.NewDecoder(w.Body).Decode(&passkeys)
	require.NoError(t, err)
	require.Len(t, passkeys, 1)
	assert.Equal(t, pk.ID, passkeys[0].ID)
	assert.Equal(t, "My Key", passkeys[0].Name)
	assert.Equal(t, []string{"usb"}, passkeys[0].Transports)
}

func TestPasskey_List_Unauthenticated(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := h.MountRouter()

	req := httptest.NewRequest("GET", "/api/auth/passkeys", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPasskey_Delete(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)

	pk, err := store.CreatePasskey(context.Background(), user.ID, "Delete Me",
		[]byte("cred-del"), []byte("pub-key"), "none", nil, 0, false, false, nil)
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/auth/passkeys/%d", pk.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestPasskey_Delete_Ownership(t *testing.T) {
	store := mockstore.New()
	alice := makeUser(t, store, "alice", false)
	bob := makeUser(t, store, "bob", false)

	_, err := store.CreatePasskey(context.Background(), bob.ID, "Bob's Key",
		[]byte("cred-bob"), []byte("pub-key"), "none", nil, 0, false, false, nil)
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, alice)

	req := httptest.NewRequest("DELETE", "/api/auth/passkeys/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Alice cannot delete Bob's passkey — should be an error.
	assert.NotEqual(t, http.StatusNoContent, w.Code)
	assert.NotEqual(t, http.StatusOK, w.Code)
}

func TestPasskey_RegisterFinish_MissingFields(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	body := `{"state_id":"","name":"","credential":null}`
	req := httptest.NewRequest("POST", "/api/auth/passkey/register/finish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPasskey_RegisterFinish_InvalidStateID(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	body := `{"state_id":"not-a-uuid","name":"My Key","credential":{}}`
	req := httptest.NewRequest("POST", "/api/auth/passkey/register/finish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPasskey_RegisterFinish_ExpiredState(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := authedRouter(h, user)

	beginReq := httptest.NewRequest("POST", "/api/auth/passkey/register/begin", nil)
	beginW := httptest.NewRecorder()
	r.ServeHTTP(beginW, beginReq)
	require.Equal(t, http.StatusOK, beginW.Code)

	var beginResp struct {
		StateID string `json:"state_id"`
	}
	json.NewDecoder(beginW.Body).Decode(&beginResp)

	stateUUID, err := uuid.Parse(beginResp.StateID)
	require.NoError(t, err)
	store.DeleteAuthState(context.Background(), stateUUID)

	body := `{"state_id":"` + beginResp.StateID + `","name":"My Key","credential":{"id":"x","type":"public-key","rawId":"x","response":{}}}`
	req := httptest.NewRequest("POST", "/api/auth/passkey/register/finish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPasskey_LoginFinish_MissingFields(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := h.Router()

	body := `{"state_id":""}`
	req := httptest.NewRequest("POST", "/api/auth/passkey/login/finish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPasskey_LoginFinish_InvalidStateID(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := h.Router()

	body := `{"state_id":"not-a-uuid","credential":{}}`
	req := httptest.NewRequest("POST", "/api/auth/passkey/login/finish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPasskey_LoginFinish_ExpiredState(t *testing.T) {
	store := mockstore.New()
	h := handlers.New(store, nil, nil, newTestPasskeyHandler(store))
	r := h.Router()

	beginReq := httptest.NewRequest("POST", "/api/auth/passkey/login/begin", nil)
	beginW := httptest.NewRecorder()
	r.ServeHTTP(beginW, beginReq)
	require.Equal(t, http.StatusOK, beginW.Code)

	var beginResp struct {
		StateID string `json:"state_id"`
	}
	json.NewDecoder(beginW.Body).Decode(&beginResp)

	stateUUID, err := uuid.Parse(beginResp.StateID)
	require.NoError(t, err)
	store.DeleteAuthState(context.Background(), stateUUID)

	body := `{"state_id":"` + beginResp.StateID + `","credential":{"id":"x","type":"public-key","rawId":"x","response":{}}}`
	req := httptest.NewRequest("POST", "/api/auth/passkey/login/finish",
		strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---- Spec-Vector-Based Login Tests ----
//
// These use the W3C WebAuthn test vectors (NoneES256) to create a valid
// authenticator assertion that can be verified by the webauthn library
// without requiring a real browser.

const (
	specCredentialIDHex     = "f91f391db4c9b2fde0ea70189cba3fb63f579ba6122b33ad94ff3ec330084be4"
	specCredentialPubKeyHex = "a5010203262001215820afefa16f97ca9b2d23eb86ccb64098d20db90856062eb249c33a9b672f26df61225820930a56b87a2fca66334b03458abf879717c12cc68ed73290af2e2664796b9220"
	specLoginChallengeHex   = "39c0e7521417ba54d43e8dc95174f423dee9bf3cd804ff6d65c857c9abf4d408"

	specAuthenticatorDataHex = "bfabc37432958b063360d3ad6461c9c4735ae7f8edd46592a5e0f01452b2e4b51900000000"
	specClientDataJSONHex    = "7b2274797065223a22776562617574686e2e676574222c226368616c6c656e6765223a224f63446e55685158756c5455506f334a5558543049393770767a7a59425039745a63685879617630314167222c226f726967696e223a2268747470733a2f2f6578616d706c652e6f7267222c2263726f73734f726967696e223a66616c73657d"
	specSignatureHex         = "3046022100f50a4e2e4409249c4a853ba361282f09841df4dd4547a13a87780218deffcd380221008480ac0f0b93538174f575bf11a1dd5d78c6e486013f937295ea13653e331e87"
)

func mustDecodeHex(t *testing.T, s string) []byte {
	t.Helper()
	data, err := hex.DecodeString(s)
	require.NoError(t, err)
	return data
}

func specLoginSessionData(t *testing.T) []byte {
	t.Helper()
	challenge := base64.RawURLEncoding.EncodeToString(mustDecodeHex(t, specLoginChallengeHex))
	sd := webauthn.SessionData{
		Challenge:            challenge,
		UserVerification:     protocol.VerificationPreferred,
	}
	data, err := json.Marshal(sd)
	require.NoError(t, err)
	return data
}

func specAssertionJSON(t *testing.T, credentialID []byte, userHandle []byte) []byte {
	t.Helper()
	body := map[string]any{
		"id":    base64.RawURLEncoding.EncodeToString(credentialID),
		"rawId": base64.RawURLEncoding.EncodeToString(credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"authenticatorData": base64.RawURLEncoding.EncodeToString(mustDecodeHex(t, specAuthenticatorDataHex)),
			"clientDataJSON":    base64.RawURLEncoding.EncodeToString(mustDecodeHex(t, specClientDataJSONHex)),
			"signature":         base64.RawURLEncoding.EncodeToString(mustDecodeHex(t, specSignatureHex)),
			"userHandle":        base64.RawURLEncoding.EncodeToString(userHandle),
		},
	}
	data, err := json.Marshal(body)
	require.NoError(t, err)
	return data
}

func newSpecPasskeyHandler(s store.Store) *handlers.PasskeyHandler {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "Test",
		RPID:          "example.org",
		RPOrigins:     []string{"https://example.org"},
	})
	if err != nil {
		panic("failed to create spec webauthn: " + err.Error())
	}
	return handlers.NewPasskeyHandler(s, wa)
}

func TestPasskey_LoginFinish_Success(t *testing.T) {
	store := mockstore.New()
	user := makeUser(t, store, "alice", false)

	credentialID := mustDecodeHex(t, specCredentialIDHex)
	pubKey := mustDecodeHex(t, specCredentialPubKeyHex)

	_, err := store.CreatePasskey(context.Background(), user.ID, "Test Passkey",
		credentialID, pubKey, "none", nil, 0, true, true, nil)
	require.NoError(t, err)

	stateID := uuid.New()
	sessionData := specLoginSessionData(t)
	err = store.SaveAuthState(context.Background(), stateID, "authentication", sessionData, time.Now().Add(5*time.Minute))
	require.NoError(t, err)

	userHandle := make([]byte, 8)
	binary.BigEndian.PutUint64(userHandle, uint64(user.ID))
	credentialJSON := specAssertionJSON(t, credentialID, userHandle)
	finishBody := map[string]any{
		"state_id":   stateID.String(),
		"credential": json.RawMessage(credentialJSON),
	}
	bodyBytes, err := json.Marshal(finishBody)
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newSpecPasskeyHandler(store))
	r := h.Router()

	req := httptest.NewRequest("POST", "/api/auth/passkey/login/finish",
		strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "response body: %s", w.Body.String())

	setCookie := w.Header().Get("Set-Cookie")
	assert.Contains(t, setCookie, "session=", "session cookie should be set")

	var resp struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
	}
	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, user.ID, resp.ID)
	assert.Equal(t, "alice", resp.Username)
}

func TestPasskey_LoginFinish_UnknownCredential(t *testing.T) {
	store := mockstore.New()

	stateID := uuid.New()
	sessionData := specLoginSessionData(t)
	err := store.SaveAuthState(context.Background(), stateID, "authentication", sessionData, time.Now().Add(5*time.Minute))
	require.NoError(t, err)

	unknownCredentialID := mustDecodeHex(t, "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	userHandle := []byte("unknown-user")

	credentialJSON := specAssertionJSON(t, unknownCredentialID, userHandle)
	finishBody := map[string]any{
		"state_id":   stateID.String(),
		"credential": json.RawMessage(credentialJSON),
	}
	bodyBytes, err := json.Marshal(finishBody)
	require.NoError(t, err)

	h := handlers.New(store, nil, nil, newSpecPasskeyHandler(store))
	r := h.Router()

	req := httptest.NewRequest("POST", "/api/auth/passkey/login/finish",
		strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code, "response body: %s", w.Body.String())
	assert.Empty(t, w.Header().Get("Set-Cookie"), "no session cookie on failure")
}
