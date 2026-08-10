package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"safenote/internal/crypto"
	"safenote/internal/domain"
	"safenote/internal/repository"
	"safenote/internal/service"
)

// App is the Wails-bound application backend. Every exported method on
// App is callable from the frontend via the generated JS bindings.
type App struct {
	ctx context.Context

	db *sql.DB

	notes    *service.NoteService
	settings *service.SettingsService
	backup   *service.BackupService
}

// NewApp constructs the App. Actual DB/service wiring happens in
// Startup so we have access to the Wails runtime context (needed for
// resolving the OS-specific app data directory).
func NewApp() *App {
	return &App{}
}

// Startup is called by Wails once the frontend is ready.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	dataDir, err := appDataDir()
	if err != nil {
		runtime.LogFatalf(ctx, "resolve app data dir: %v", err)
		return
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		runtime.LogFatalf(ctx, "create app data dir: %v", err)
		return
	}

	dbPath := filepath.Join(dataDir, "safenote.db")
	db, err := repository.Open(ctx, dbPath)
	if err != nil {
		runtime.LogFatalf(ctx, "open database: %v", err)
		return
	}
	a.db = db

	vaultKey, err := crypto.LoadOrCreateVaultKey(dataDir)
	if err != nil {
		runtime.LogFatalf(ctx, "load vault key: %v", err)
		return
	}

	noteRepo := repository.NewNoteRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	historyRepo := repository.NewBackupHistoryRepository(db)

	a.settings = service.NewSettingsService(settingsRepo, vaultKey)
	lockout := service.NewLockoutTracker(5, 30*time.Second)
	a.notes = service.NewNoteService(noteRepo, a.settings, lockout)
	a.backup = service.NewBackupService(noteRepo, historyRepo)
}

// Shutdown is called by Wails on application exit.
func (a *App) Shutdown(ctx context.Context) {
	if a.db != nil {
		_ = a.db.Close()
	}
}

func appDataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("app: resolve user config dir: %w", err)
	}
	return filepath.Join(base, "SafeNote"), nil
}

// ---- First-run / Settings ----

func (a *App) IsSecretKeyConfigured() (bool, error) {
	return a.settings.IsConfigured(a.ctx)
}

func (a *App) SetupSecretKey(secretKey string) error {
	return a.settings.SetupSecretKey(a.ctx, secretKey)
}

func (a *App) ChangeSecretKey(newSecretKey string) error {
	return a.settings.ChangeSecretKey(a.ctx, newSecretKey)
}

func (a *App) GetMaskedSecretKey() (string, error) {
	return a.settings.MaskedSecretKey(a.ctx)
}

func (a *App) GetSettings() (domain.Settings, error) {
	return a.settings.Get(a.ctx)
}

func (a *App) UpdateTheme(theme string) error {
	return a.settings.UpdateTheme(a.ctx, theme)
}

func (a *App) UpdateClipboardTimeout(seconds int) error {
	return a.settings.UpdateClipboardTimeout(a.ctx, seconds)
}

func (a *App) ExportConfig(path string) error {
	cfg, err := a.settings.ExportConfig(a.ctx)
	if err != nil {
		return err
	}
	return writeJSONFile(path, cfg)
}

func (a *App) ImportConfig(path string) error {
	var cfg service.ExportedConfig
	if err := readJSONFile(path, &cfg); err != nil {
		return err
	}
	return a.settings.ImportConfig(a.ctx, cfg)
}

// ---- Notes ----

func (a *App) CreateNote(title, content, pin string, tags []string) (domain.NoteCard, error) {
	return a.notes.Create(a.ctx, service.CreateInput{Title: title, Content: content, Pin: pin, Tags: tags})
}

func (a *App) OpenNote(id, pin string) (domain.DecryptedNote, error) {
	return a.notes.Open(a.ctx, id, pin)
}

// CopyNoteToClipboard is a legacy convenience method that copies the
// raw decrypted content directly to the clipboard. Since note content
// may be stored as Quill Delta JSON, prefer the frontend flow of
// OpenNote() + deltaToPlainText() + SetClipboardText() for
// human-readable clipboard output.
func (a *App) CopyNoteToClipboard(id, pin string) error {
	content, err := a.notes.Copy(a.ctx, id, pin)
	if err != nil {
		return err
	}
	runtime.ClipboardSetText(a.ctx, content)
	return nil
}

// SetClipboardText writes text (already decrypted/converted to plain
// text on the frontend, e.g. from a rich-text Delta) to the native OS
// clipboard. Kept separate from CopyNoteToClipboard because SafeNote's
// backend stores note content in a format-agnostic way (it may be a
// Quill Delta JSON string, not display-ready text) — converting Delta
// to plain text is a frontend concern.
func (a *App) SetClipboardText(text string) error {
	runtime.ClipboardSetText(a.ctx, text)
	return nil
}

func (a *App) ClearClipboard() {
	runtime.ClipboardSetText(a.ctx, "")
}

func (a *App) EditNote(id, pin, title, content string, tags []string) (domain.NoteCard, error) {
	return a.notes.Edit(a.ctx, service.EditInput{ID: id, Pin: pin, Title: title, Content: content, Tags: tags})
}

func (a *App) DeleteNote(id string) error {
	return a.notes.Delete(a.ctx, id)
}

func (a *App) SetFavorite(id string, favorite bool) error {
	return a.notes.SetFavorite(a.ctx, id, favorite)
}

func (a *App) SetPinned(id string, pinned bool) error {
	return a.notes.SetPinned(a.ctx, id, pinned)
}

// ListNotesResult wraps the paginated note list. Wails v2's JS binding
// generator only reliably supports a (result, error) return shape;
// returning three raw values ([]NoteCard, int, error) does not produce
// an iterable array on the JS side, so we wrap the tuple here instead.
type ListNotesResult struct {
	Notes []domain.NoteCard `json:"notes"`
	Total int               `json:"total"`
}

func (a *App) ListNotes(search string, sort string, page, pageSize int, onlyFavorite bool) (ListNotesResult, error) {
	cards, total, err := a.notes.List(a.ctx, domain.ListQuery{
		Search: search, Sort: domain.SortMode(sort), Page: page, PageSize: pageSize, OnlyFav: onlyFavorite,
	})
	if err != nil {
		return ListNotesResult{}, err
	}
	return ListNotesResult{Notes: cards, Total: total}, nil
}

// ---- Backup / Restore ----

func (a *App) ExportBackup(path string) (domain.BackupFile, error) {
	return a.backup.Export(a.ctx, path)
}

func (a *App) PreviewRestore(path string) (service.RestorePreview, error) {
	return a.backup.RestorePreview(a.ctx, path)
}

func (a *App) RestoreBackup(path string, overwriteIDs []string) (int, error) {
	overwrite := make(map[string]bool, len(overwriteIDs))
	for _, id := range overwriteIDs {
		overwrite[id] = true
	}
	return a.backup.Restore(a.ctx, path, overwrite)
}

// ---- Native dialogs ----

func (a *App) SaveFileDialog(defaultFilename string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "SafeNote Backup (*.json)", Pattern: "*.json"},
		},
	})
}

func (a *App) OpenFileDialog() (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "SafeNote Backup (*.json)", Pattern: "*.json"},
		},
	})
}
