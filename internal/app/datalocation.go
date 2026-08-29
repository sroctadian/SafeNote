package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// dbFileName and vaultFileName must match repository.Open's DB filename
// and crypto.VaultKeyFileName respectively — duplicated as constants
// here (rather than importing those packages just for a string) to keep
// this file's file-management logic self-contained and easy to audit.
const (
	dbFileName        = "safenote.db"
	vaultFileNameCopy = "vault.key"
	locationFileName  = "location.json"
)

// locationPointer is a tiny, ALWAYS-fixed-location file that tells
// SafeNote where the real data directory (database + vault key) lives.
// It never moves itself — only the directory it points to is
// configurable — which is what avoids the chicken-and-egg problem of
// "where do I look to find out where to look".
type locationPointer struct {
	DataDir string `json:"dataDir"`
}

// fixedConfigRoot is the one location that is NEVER user-configurable:
// the OS-standard per-user config directory. It holds the location
// pointer file, and is also the default data directory when the user
// has never customized it.
func fixedConfigRoot() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("app: resolve user config dir: %w", err)
	}
	return filepath.Join(base, "SafeNote"), nil
}

// ResolveDataDir returns the directory SafeNote should read/write its
// database and vault key from: the custom directory the user chose via
// SetDataDirectory, or the fixed default if never customized.
func ResolveDataDir() (string, error) {
	root, err := fixedConfigRoot()
	if err != nil {
		return "", err
	}

	pointerPath := filepath.Join(root, locationFileName)
	raw, err := os.ReadFile(pointerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return root, nil // never customized — use the default
		}
		return "", fmt.Errorf("app: read location pointer: %w", err)
	}

	var loc locationPointer
	if err := json.Unmarshal(raw, &loc); err != nil || loc.DataDir == "" {
		return root, nil // corrupt/empty pointer — fail safe to default
	}
	return loc.DataDir, nil
}

// DataDirStatus describes the outcome of a SetDataDirectory call, for
// the frontend to explain to the user what actually happened.
type DataDirStatus struct {
	NewPath         string `json:"newPath"`
	Adopted         bool   `json:"adopted"`         // true: pointed at an EXISTING SafeNote data folder, nothing copied
	Moved           bool   `json:"moved"`           // true: copied the current database + vault key to the new folder
	OldDataBackedUp bool   `json:"oldDataBackedUp"` // true: a safety copy of the previous local DB was kept
	RestartRequired bool   `json:"restartRequired"`
}

// SetDataDirectory points SafeNote at a different folder for its
// database and vault key, going forward. It does NOT hot-swap the
// currently open database connection — the app must be restarted for
// the change to take effect, which is surfaced to the frontend via
// RestartRequired rather than attempted live (swapping an open SQLite
// connection and decryption key mid-session risks corruption or a
// half-old-half-new state).
//
// Two distinct flows, chosen automatically based on what's already at
// newDir:
//   - MOVE: newDir has no existing safenote.db — the current database
//     and vault key are copied there (e.g. relocating storage, or
//     pointing at an empty folder that's about to be synced).
//   - ADOPT: newDir already has a safenote.db — nothing is copied or
//     overwritten; SafeNote will simply read from that existing data
//     the next time it starts (e.g. pointing at a folder that a synced
//     copy from another device already populated). The current LOCAL
//     database is renamed to a timestamped backup file first, purely
//     as a safety net in case this was pointed at the wrong folder by
//     mistake.
func SetDataDirectory(newDir string) (DataDirStatus, error) {
	newDir = filepath.Clean(newDir)
	if newDir == "" {
		return DataDirStatus{}, fmt.Errorf("app: empty directory path")
	}
	if err := os.MkdirAll(newDir, 0700); err != nil {
		return DataDirStatus{}, fmt.Errorf("app: create target directory: %w", err)
	}

	currentDir, err := ResolveDataDir()
	if err != nil {
		return DataDirStatus{}, err
	}
	if filepath.Clean(currentDir) == newDir {
		return DataDirStatus{NewPath: newDir}, nil // no-op, already here
	}

	status := DataDirStatus{NewPath: newDir, RestartRequired: true}

	targetDBPath := filepath.Join(newDir, dbFileName)
	if _, err := os.Stat(targetDBPath); err == nil {
		// ADOPT: an existing SafeNote database is already at the target.
		// Back up (don't delete) whatever is currently local, then just
		// point at the existing one.
		if err := backupLocalDatabase(currentDir); err == nil {
			status.OldDataBackedUp = true
		}
		status.Adopted = true
	} else {
		// MOVE: copy current DB (+ WAL/SHM sidecar files, if present)
		// and the vault key to the new location. See ADR-007: this
		// necessarily copies the vault key alongside the database,
		// which trades away some of the defense-in-depth described in
		// ADR-003 in exchange for the data actually being usable from
		// the new location.
		if err := copyFileIfExists(filepath.Join(currentDir, dbFileName), targetDBPath); err != nil {
			return DataDirStatus{}, fmt.Errorf("app: copy database: %w", err)
		}
		_ = copyFileIfExists(filepath.Join(currentDir, dbFileName+"-wal"), filepath.Join(newDir, dbFileName+"-wal"))
		_ = copyFileIfExists(filepath.Join(currentDir, dbFileName+"-shm"), filepath.Join(newDir, dbFileName+"-shm"))
		if err := copyFileIfExists(filepath.Join(currentDir, vaultFileNameCopy), filepath.Join(newDir, vaultFileNameCopy)); err != nil {
			return DataDirStatus{}, fmt.Errorf("app: copy vault key: %w", err)
		}
		status.Moved = true
	}

	if err := writeLocationPointer(newDir); err != nil {
		return DataDirStatus{}, err
	}

	return status, nil
}

func writeLocationPointer(dataDir string) error {
	root, err := fixedConfigRoot()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return fmt.Errorf("app: create config root: %w", err)
	}

	raw, err := json.MarshalIndent(locationPointer{DataDir: dataDir}, "", "  ")
	if err != nil {
		return fmt.Errorf("app: marshal location pointer: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, locationFileName), raw, 0600); err != nil {
		return fmt.Errorf("app: write location pointer: %w", err)
	}
	return nil
}

func backupLocalDatabase(dir string) error {
	src := filepath.Join(dir, dbFileName)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // nothing local to back up
	}
	backupName := fmt.Sprintf("safenote.pre-adopt-backup-%s.db", time.Now().UTC().Format("20060102-150405"))
	return copyFileIfExists(src, filepath.Join(dir, backupName))
}

func copyFileIfExists(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.WriteFile(dst, data, 0600)
}
