package app

import (
	"os"
	"path/filepath"
	"testing"
)

// withFixedConfigRoot temporarily redirects HOME/APPDATA-derived
// os.UserConfigDir() to a temp dir for the duration of the test, so
// tests never touch the real machine's config directory.
func withFixedConfigRoot(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	switch {
	case os.Getenv("APPDATA") != "" || isWindowsGOOS():
		t.Setenv("APPDATA", tmp)
	case os.Getenv("XDG_CONFIG_HOME") != "":
		t.Setenv("XDG_CONFIG_HOME", tmp)
	default:
		// os.UserConfigDir() falls back to $HOME/.config on Linux and
		// $HOME/Library/Application Support on macOS if the more
		// specific env vars aren't set; setting HOME covers both.
		t.Setenv("HOME", tmp)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	}

	return tmp
}

func isWindowsGOOS() bool {
	return os.PathSeparator == '\\'
}

func TestResolveDataDir_DefaultsWhenNeverCustomized(t *testing.T) {
	withFixedConfigRoot(t)

	dir, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("ResolveDataDir: %v", err)
	}
	root, _ := fixedConfigRoot()
	if dir != root {
		t.Fatalf("expected default dir %q, got %q", root, dir)
	}
}

func TestSetDataDirectory_MoveCopiesExistingFiles(t *testing.T) {
	withFixedConfigRoot(t)
	root, _ := fixedConfigRoot()
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}

	// Simulate an existing database + vault key at the default location.
	if err := os.WriteFile(filepath.Join(root, dbFileName), []byte("fake-db"), 0600); err != nil {
		t.Fatalf("write fake db: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, vaultFileNameCopy), []byte("fake-vault-key"), 0600); err != nil {
		t.Fatalf("write fake vault key: %v", err)
	}

	newDir := filepath.Join(t.TempDir(), "synced-folder")
	status, err := SetDataDirectory(newDir)
	if err != nil {
		t.Fatalf("SetDataDirectory: %v", err)
	}

	if !status.Moved || status.Adopted {
		t.Fatalf("expected a MOVE, got status=%+v", status)
	}
	if !status.RestartRequired {
		t.Fatal("expected RestartRequired to be true")
	}

	gotDB, err := os.ReadFile(filepath.Join(newDir, dbFileName))
	if err != nil || string(gotDB) != "fake-db" {
		t.Fatalf("expected database copied to new dir, err=%v content=%q", err, gotDB)
	}
	gotVault, err := os.ReadFile(filepath.Join(newDir, vaultFileNameCopy))
	if err != nil || string(gotVault) != "fake-vault-key" {
		t.Fatalf("expected vault key copied to new dir, err=%v content=%q", err, gotVault)
	}

	resolved, err := ResolveDataDir()
	if err != nil {
		t.Fatalf("ResolveDataDir after move: %v", err)
	}
	if filepath.Clean(resolved) != filepath.Clean(newDir) {
		t.Fatalf("expected ResolveDataDir to return %q, got %q", newDir, resolved)
	}
}

func TestSetDataDirectory_AdoptsExistingDatabaseWithoutOverwriting(t *testing.T) {
	withFixedConfigRoot(t)
	root, _ := fixedConfigRoot()
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, dbFileName), []byte("local-db"), 0600); err != nil {
		t.Fatalf("write local db: %v", err)
	}

	// A folder that already has a SafeNote database (e.g. synced from
	// another device).
	existingDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(existingDir, dbFileName), []byte("remote-db"), 0600); err != nil {
		t.Fatalf("write remote db: %v", err)
	}

	status, err := SetDataDirectory(existingDir)
	if err != nil {
		t.Fatalf("SetDataDirectory: %v", err)
	}
	if !status.Adopted || status.Moved {
		t.Fatalf("expected an ADOPT, got status=%+v", status)
	}
	if !status.OldDataBackedUp {
		t.Fatal("expected the local db to be backed up before adopting")
	}

	// The existing (remote) database must NOT have been overwritten.
	got, err := os.ReadFile(filepath.Join(existingDir, dbFileName))
	if err != nil || string(got) != "remote-db" {
		t.Fatalf("expected existing remote db untouched, err=%v content=%q", err, got)
	}

	// A backup of the local db should exist somewhere in root.
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read root dir: %v", err)
	}
	foundBackup := false
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".db" && e.Name() != dbFileName {
			foundBackup = true
		}
	}
	if !foundBackup {
		t.Fatal("expected a timestamped backup file of the local db in the config root")
	}
}

func TestSetDataDirectory_NoOpWhenSameDirectory(t *testing.T) {
	withFixedConfigRoot(t)
	root, _ := fixedConfigRoot()

	status, err := SetDataDirectory(root)
	if err != nil {
		t.Fatalf("SetDataDirectory: %v", err)
	}
	if status.Moved || status.Adopted || status.RestartRequired {
		t.Fatalf("expected a no-op for identical directory, got %+v", status)
	}
}

func TestSetDataDirectory_RejectsEmptyPath(t *testing.T) {
	withFixedConfigRoot(t)
	if _, err := SetDataDirectory(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}
