package crypto

import (
	"crypto/rand"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

var ErrEmptyInput = errors.New("crypto: secret key and pin must not be empty")

// NewSalt generates a cryptographically secure random salt.
func NewSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("crypto: generate salt: %w", err)
	}
	return salt, nil
}

// NewNonce generates a cryptographically secure random nonce for AES-GCM.
func NewNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("crypto: generate nonce: %w", err)
	}
	return nonce, nil
}

// DeriveKey combines the user's Secret Key (SK) and per-note PIN with a
// random salt to derive a 256-bit encryption key via Argon2id.
//
// Neither the PIN, the Secret Key, nor the derived key are ever persisted.
// The caller is responsible for wiping the returned key with Wipe() once
// it is no longer needed.
func DeriveKey(secretKey, pin string, salt []byte) ([]byte, error) {
	if secretKey == "" || pin == "" {
		return nil, ErrEmptyInput
	}
	if len(salt) != SaltSize {
		return nil, fmt.Errorf("crypto: invalid salt length %d", len(salt))
	}

	// Combine PIN + Secret Key as the Argon2id password material.
	material := make([]byte, 0, len(pin)+len(secretKey))
	material = append(material, []byte(pin)...)
	material = append(material, []byte(secretKey)...)
	defer Wipe(material)

	key := argon2.IDKey(material, salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)
	return key, nil
}
