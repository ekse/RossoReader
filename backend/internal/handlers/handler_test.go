package handlers_test

import (
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/ekse/rossoreader/internal/handlers"
	"github.com/ekse/rossoreader/internal/store"
)

func newTestPasskeyHandler(s store.Store) *handlers.PasskeyHandler {
	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "Test",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:5173"},
	})
	if err != nil {
		panic("failed to create test webauthn: " + err.Error())
	}
	return handlers.NewPasskeyHandler(s, wa)
}
