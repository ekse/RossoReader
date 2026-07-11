package auth_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ekse/rossoreader/internal/auth"
)

func TestHashPassword_VerifyPassword_RoundTrip(t *testing.T) {
	hash, err := auth.HashPassword("s3cret-passphrase!")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)

	ok, err := auth.VerifyPassword(hash, "s3cret-passphrase!")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("correct horse")
	require.NoError(t, err)

	ok, err := auth.VerifyPassword(hash, "battery staple")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	h1, _ := auth.HashPassword("same-pass")
	h2, _ := auth.HashPassword("same-pass")
	assert.NotEqual(t, h1, h2, "two hashes of same password must differ due to salt")
}

func TestVerifyPassword_MalformedHash(t *testing.T) {
	_, err := auth.VerifyPassword("not-a-hash", "anything")
	assert.Error(t, err)
}
