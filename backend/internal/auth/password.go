package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	hashMemory      = 64 * 1024
	hashIterations  = 3
	hashParallelism = 2
	hashSaltLength  = 16
	hashKeyLength   = 32
)

// HashPassword returns an argon2id encoded hash for the given password.
// Format: $argon2id$v=19$m=65536,t=3,p=2$<base64-salt>$<base64-key>
func HashPassword(password string) (string, error) {
	salt := make([]byte, hashSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	key := argon2.IDKey(
		[]byte(password), salt,
		hashIterations, hashMemory, hashParallelism, hashKeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		hashMemory, hashIterations, hashParallelism, b64Salt, b64Key), nil
}

// VerifyPassword checks the encoded argon2id hash against the plaintext password.
// Returns true if the password matches.
func VerifyPassword(encoded, password string) (bool, error) {
	parts, err := parseArgon2id(encoded)
	if err != nil {
		return false, err
	}

	key := argon2.IDKey(
		[]byte(password), parts.salt,
		parts.iterations, parts.memory, parts.parallelism, parts.keyLength,
	)

	otherKey, err := base64.RawStdEncoding.DecodeString(parts.b64Key)
	if err != nil {
		return false, fmt.Errorf("decode stored key: %w", err)
	}

	if len(key) != len(otherKey) {
		return false, nil
	}

	return subtle.ConstantTimeCompare(key, otherKey) == 1, nil
}

type argon2idParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	keyLength   uint32
	salt        []byte
	b64Key      string
}

func parseArgon2id(encoded string) (*argon2idParams, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return nil, errors.New("invalid argon2id hash format")
	}
	if parts[1] != "argon2id" {
		return nil, errors.New("not an argon2id hash")
	}
	if !strings.HasPrefix(parts[2], "v=") {
		return nil, errors.New("missing version")
	}

	var memory, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return nil, fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}

	keyBytes, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}

	return &argon2idParams{
		memory:      memory,
		iterations:  iterations,
		parallelism: parallelism,
		keyLength:   uint32(len(keyBytes)),
		salt:        salt,
		b64Key:      parts[5],
	}, nil
}
