package crypto

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

// VaultKeyFileName is the file storing the local machine vault key that
// wraps the user's Secret Key at rest. Keeping this key outside the
// SQLite database means database theft alone does not expose the
// Secret Key.
const VaultKeyFileName = "vault.key"

// LoadOrCreateVaultKey reads the local vault key from dir, generating
// and persisting a new random 256-bit key on first run. The file is
// created with 0600 permissions (owner read/write only).
func LoadOrCreateVaultKey(dir string) ([]byte, error) {
	path := filepath.Join(dir, VaultKeyFileName)

	data, err := os.ReadFile(path)
	if err == nil {
		if len(data) != KeySize {
			return nil, fmt.Errorf("crypto: corrupt vault key at %s", path)
		}
		return data, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("crypto: read vault key: %w", err)
	}

	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("crypto: create app dir: %w", err)
	}

	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("crypto: generate vault key: %w", err)
	}
	if err := os.WriteFile(path, key, 0600); err != nil {
		return nil, fmt.Errorf("crypto: write vault key: %w", err)
	}
	return key, nil
}
