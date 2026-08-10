package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

// ErrDecryptionFailed is returned for any failure during decryption
// (wrong key, tampered ciphertext, wrong nonce, etc). It is intentionally
// generic: SafeNote must never leak *why* decryption failed, since that
// can be used to distinguish "wrong PIN" from "corrupted data" and aid
// an attacker.
var ErrDecryptionFailed = errors.New("crypto: decryption failed")

// Encrypt seals plaintext with AES-256-GCM using the given key and nonce.
// The key must be exactly KeySize (32) bytes and the nonce exactly
// NonceSize (12) bytes. Returns ciphertext with the GCM auth tag appended.
func Encrypt(key, nonce, plaintext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("crypto: invalid key length %d", len(key))
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("crypto: invalid nonce length %d", len(nonce))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, NonceSize)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt opens ciphertext sealed by Encrypt. Any failure (wrong key,
// tampered data, wrong nonce) collapses to ErrDecryptionFailed.
func Decrypt(key, nonce, ciphertext []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, ErrDecryptionFailed
	}
	if len(nonce) != NonceSize {
		return nil, ErrDecryptionFailed
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, NonceSize)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plaintext, nil
}
