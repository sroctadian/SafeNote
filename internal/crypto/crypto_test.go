package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKey_Deterministic(t *testing.T) {
	salt, err := NewSalt()
	if err != nil {
		t.Fatalf("NewSalt: %v", err)
	}

	k1, err := DeriveKey("super-secret-key-value-32chars!!", "1234", salt)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}
	k2, err := DeriveKey("super-secret-key-value-32chars!!", "1234", salt)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}

	if !bytes.Equal(k1, k2) {
		t.Fatal("expected identical key material for identical inputs and salt")
	}
	if len(k1) != KeySize {
		t.Fatalf("expected key length %d, got %d", KeySize, len(k1))
	}
}

func TestDeriveKey_DifferentPinDifferentKey(t *testing.T) {
	salt, _ := NewSalt()
	k1, _ := DeriveKey("secret-key-value-must-be-32ch!!!", "1111", salt)
	k2, _ := DeriveKey("secret-key-value-must-be-32ch!!!", "2222", salt)

	if bytes.Equal(k1, k2) {
		t.Fatal("expected different keys for different PINs")
	}
}

func TestDeriveKey_EmptyInputRejected(t *testing.T) {
	salt, _ := NewSalt()
	if _, err := DeriveKey("", "1234", salt); err != ErrEmptyInput {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
	if _, err := DeriveKey("sk", "", salt); err != ErrEmptyInput {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	salt, _ := NewSalt()
	nonce, err := NewNonce()
	if err != nil {
		t.Fatalf("NewNonce: %v", err)
	}
	key, err := DeriveKey("my-secret-key-of-32-characters!!", "0000", salt)
	if err != nil {
		t.Fatalf("DeriveKey: %v", err)
	}

	plaintext := []byte("this is a secret note body")
	ciphertext, err := Encrypt(key, nonce, plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext must not equal plaintext")
	}

	decrypted, err := Decrypt(key, nonce, ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecrypt_WrongKeyFailsGenerically(t *testing.T) {
	salt, _ := NewSalt()
	nonce, _ := NewNonce()
	key1, _ := DeriveKey("secret-key-value-32-characters!!", "1234", salt)
	key2, _ := DeriveKey("secret-key-value-32-characters!!", "9999", salt)

	ciphertext, err := Encrypt(key1, nonce, []byte("data"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(key2, nonce, ciphertext)
	if err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed, got %v", err)
	}
}

func TestDecrypt_TamperedCiphertextFails(t *testing.T) {
	salt, _ := NewSalt()
	nonce, _ := NewNonce()
	key, _ := DeriveKey("secret-key-value-32-characters!!", "1234", salt)

	ciphertext, _ := Encrypt(key, nonce, []byte("data"))
	ciphertext[0] ^= 0xFF // flip a bit

	if _, err := Decrypt(key, nonce, ciphertext); err != ErrDecryptionFailed {
		t.Fatalf("expected ErrDecryptionFailed for tampered data, got %v", err)
	}
}

func TestWipe_ZeroesSlice(t *testing.T) {
	b := []byte{1, 2, 3, 4}
	Wipe(b)
	for i, v := range b {
		if v != 0 {
			t.Fatalf("byte %d not wiped: %d", i, v)
		}
	}
}

func TestNewSalt_NewNonce_UniqueEachCall(t *testing.T) {
	s1, _ := NewSalt()
	s2, _ := NewSalt()
	if bytes.Equal(s1, s2) {
		t.Fatal("expected unique salts across calls")
	}
	n1, _ := NewNonce()
	n2, _ := NewNonce()
	if bytes.Equal(n1, n2) {
		t.Fatal("expected unique nonces across calls")
	}
}
