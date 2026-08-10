package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"safenote/internal/domain"
	"safenote/internal/repository"
)

func newBackupTestEnv(t *testing.T) (*NoteService, *BackupService) {
	t.Helper()
	ctx := context.Background()

	db, err := repository.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	noteRepo := repository.NewNoteRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	historyRepo := repository.NewBackupHistoryRepository(db)

	settingsSvc := NewSettingsService(settingsRepo, make([]byte, 32))
	if err := settingsSvc.SetupSecretKey(ctx, "this-is-a-32-character-secret-key!!"); err != nil {
		t.Fatalf("setup secret key: %v", err)
	}

	noteSvc := NewNoteService(noteRepo, settingsSvc, NewLockoutTracker(5, 0))
	backupSvc := NewBackupService(noteRepo, historyRepo)

	return noteSvc, backupSvc
}

func TestBackupService_ExportNeverDecrypts(t *testing.T) {
	ctx := context.Background()
	noteSvc, backupSvc := newBackupTestEnv(t)

	_, _ = noteSvc.Create(ctx, CreateInput{Title: "Secret", Content: "sensitive data", Pin: "123456"})

	dir := t.TempDir()
	path := filepath.Join(dir, "backup.json")

	bf, err := backupSvc.Export(ctx, path)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(bf.Notes) != 1 {
		t.Fatalf("expected 1 note in backup, got %d", len(bf.Notes))
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read backup file: %v", err)
	}
	if containsPlaintext(raw, "sensitive data") {
		t.Fatal("backup file must never contain plaintext note content")
	}
}

func TestBackupService_RestorePreviewDetectsDuplicates(t *testing.T) {
	ctx := context.Background()
	noteSvc, backupSvc := newBackupTestEnv(t)

	_, _ = noteSvc.Create(ctx, CreateInput{Title: "N1", Content: "c1", Pin: "123456"})

	dir := t.TempDir()
	path := filepath.Join(dir, "backup.json")
	if _, err := backupSvc.Export(ctx, path); err != nil {
		t.Fatalf("export: %v", err)
	}

	preview, err := backupSvc.RestorePreview(ctx, path)
	if err != nil {
		t.Fatalf("restore preview: %v", err)
	}
	if !preview.Valid || preview.NoteCount != 1 {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	if len(preview.DuplicateIDs) != 1 {
		t.Fatalf("expected 1 duplicate (re-importing same file), got %d", len(preview.DuplicateIDs))
	}
}

func TestBackupService_RestoreSkipsDuplicatesWithoutOverwrite(t *testing.T) {
	ctx := context.Background()
	noteSvc, backupSvc := newBackupTestEnv(t)

	_, _ = noteSvc.Create(ctx, CreateInput{Title: "N1", Content: "c1", Pin: "123456"})

	dir := t.TempDir()
	path := filepath.Join(dir, "backup.json")
	backupSvc.Export(ctx, path)

	imported, err := backupSvc.Restore(ctx, path, map[string]bool{})
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if imported != 0 {
		t.Fatalf("expected 0 imported (duplicate skipped, no overwrite), got %d", imported)
	}
}

func TestBackupService_ChecksumMismatchRejected(t *testing.T) {
	ctx := context.Background()
	noteSvc, backupSvc := newBackupTestEnv(t)

	_, _ = noteSvc.Create(ctx, CreateInput{Title: "N1", Content: "c1", Pin: "123456"})

	dir := t.TempDir()
	path := filepath.Join(dir, "backup.json")
	backupSvc.Export(ctx, path)

	raw, _ := os.ReadFile(path)
	tampered := []byte(replaceFirst(string(raw), `"N1"`, `"N1-tampered"`))
	os.WriteFile(path, tampered, 0600)

	_, err := backupSvc.RestorePreview(ctx, path)
	if err == nil {
		t.Fatal("expected error for tampered/corrupted backup file")
	}
}

func replaceFirst(s, old, new string) string {
	idx := indexOf(s, old)
	if idx < 0 {
		return s
	}
	return s[:idx] + new + s[idx+len(old):]
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func containsPlaintext(data []byte, needle string) bool {
	return len(needle) > 0 && string(data) != "" && bytesContains(data, []byte(needle))
}

func bytesContains(haystack, needle []byte) bool {
	if len(needle) == 0 || len(haystack) < len(needle) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

var _ = domain.ErrChecksumMismatch // ensure domain import used if future assertions need it
