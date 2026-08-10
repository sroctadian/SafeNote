package service

import (
	"context"
	"testing"
	"time"

	"safenote/internal/domain"
	"safenote/internal/repository"
)

func newTestServices(t *testing.T) (*NoteService, *SettingsService) {
	t.Helper()
	ctx := context.Background()

	db, err := repository.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	noteRepo := repository.NewNoteRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	vaultKey := make([]byte, 32)
	for i := range vaultKey {
		vaultKey[i] = byte(i)
	}
	settingsSvc := NewSettingsService(settingsRepo, vaultKey)
	if err := settingsSvc.SetupSecretKey(ctx, "this-is-a-32-character-secret-key!!"); err != nil {
		t.Fatalf("setup secret key: %v", err)
	}

	lockout := NewLockoutTracker(3, 200*time.Millisecond)
	noteSvc := NewNoteService(noteRepo, settingsSvc, lockout)

	return noteSvc, settingsSvc
}

func TestNoteService_CreateAndOpen(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	card, err := noteSvc.Create(ctx, CreateInput{
		Title: "My Note", Content: "top secret body", Pin: "123456",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if card.ID == "" {
		t.Fatal("expected generated ID")
	}

	opened, err := noteSvc.Open(ctx, card.ID, "123456")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if opened.Content != "top secret body" {
		t.Fatalf("unexpected content: %q", opened.Content)
	}
}

func TestNoteService_WrongPinGenericError(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	card, err := noteSvc.Create(ctx, CreateInput{Title: "N", Content: "body", Pin: "111111"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, err = noteSvc.Open(ctx, card.ID, "999999")
	if err != domain.ErrInvalidPin {
		t.Fatalf("expected ErrInvalidPin, got %v", err)
	}
}

func TestNoteService_CooldownAfterFailures(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	card, _ := noteSvc.Create(ctx, CreateInput{Title: "N", Content: "body", Pin: "111111"})

	for i := 0; i < 3; i++ {
		_, _ = noteSvc.Open(ctx, card.ID, "654321")
	}

	_, err := noteSvc.Open(ctx, card.ID, "111111") // even correct PIN should be blocked
	if err == nil {
		t.Fatal("expected cooldown error")
	}
}

func TestNoteService_EditReEncrypts(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	card, _ := noteSvc.Create(ctx, CreateInput{Title: "N", Content: "v1", Pin: "123456"})

	_, err := noteSvc.Edit(ctx, EditInput{ID: card.ID, Pin: "123456", Title: "N2", Content: "v2"})
	if err != nil {
		t.Fatalf("edit: %v", err)
	}

	opened, err := noteSvc.Open(ctx, card.ID, "123456")
	if err != nil {
		t.Fatalf("open after edit: %v", err)
	}
	if opened.Content != "v2" || opened.Title != "N2" {
		t.Fatalf("expected updated content, got title=%q content=%q", opened.Title, opened.Content)
	}
}

func TestNoteService_DeleteAndList(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	c1, _ := noteSvc.Create(ctx, CreateInput{Title: "Alpha", Content: "a", Pin: "123456"})
	_, _ = noteSvc.Create(ctx, CreateInput{Title: "Beta", Content: "b", Pin: "123456"})

	cards, total, err := noteSvc.List(ctx, domain.ListQuery{Sort: domain.SortAlphabet, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(cards) != 2 {
		t.Fatalf("expected 2 notes, got total=%d len=%d", total, len(cards))
	}

	if err := noteSvc.Delete(ctx, c1.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, total, err = noteSvc.List(ctx, domain.ListQuery{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 note after delete, got %d", total)
	}
}

func TestNoteService_SearchByTitle(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	_, _ = noteSvc.Create(ctx, CreateInput{Title: "Shopping List", Content: "milk", Pin: "123456"})
	_, _ = noteSvc.Create(ctx, CreateInput{Title: "Meeting Notes", Content: "agenda", Pin: "123456"})

	cards, total, err := noteSvc.List(ctx, domain.ListQuery{Search: "shop", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || cards[0].Title != "Shopping List" {
		t.Fatalf("expected 1 match 'Shopping List', got total=%d", total)
	}
}

func TestNoteService_ContentIsFormatAgnostic(t *testing.T) {
	// The service layer never inspects note content structure — it
	// encrypts whatever string it's given. This test documents that
	// contract, since the frontend now stores rich text as Quill Delta
	// JSON rather than plain text, and no backend change was needed to
	// support that.
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	deltaJSON := `{"ops":[{"insert":"Hello "},{"insert":"world","attributes":{"bold":true}},{"insert":"\n"}]}`

	card, err := noteSvc.Create(ctx, CreateInput{Title: "Rich Note", Content: deltaJSON, Pin: "123456"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	opened, err := noteSvc.Open(ctx, card.ID, "123456")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if opened.Content != deltaJSON {
		t.Fatalf("expected content to round-trip unchanged, got %q", opened.Content)
	}
}

func TestSettingsService_SecretKeyTooShortRejected(t *testing.T) {
	ctx := context.Background()
	db, err := repository.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := repository.NewSettingsRepository(db)
	svc := NewSettingsService(repo, make([]byte, 32))

	if err := svc.SetupSecretKey(ctx, "too-short"); err != domain.ErrSecretKeyTooShort {
		t.Fatalf("expected ErrSecretKeyTooShort, got %v", err)
	}
}

func TestSettingsService_ChangeSecretKey(t *testing.T) {
	ctx := context.Background()
	_, settingsSvc := newTestServices(t)

	newKey := "a-brand-new-32-character-secret!!"
	if err := settingsSvc.ChangeSecretKey(ctx, newKey); err != nil {
		t.Fatalf("change secret key: %v", err)
	}

	got, err := settingsSvc.GetDecryptedSecretKey(ctx)
	if err != nil {
		t.Fatalf("get decrypted secret key: %v", err)
	}
	if string(got) != newKey {
		t.Fatalf("expected %q, got %q", newKey, string(got))
	}
}
