package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"safenote/internal/crypto"
	"safenote/internal/domain"
	"safenote/internal/repository"
)

// NoteService orchestrates the encrypt/decrypt flow around NoteRepository.
// It is the only layer allowed to touch plaintext note content, and only
// transiently in memory.
type NoteService struct {
	notes    *repository.NoteRepository
	settings *SettingsService
	lockout  *LockoutTracker
}

func NewNoteService(notes *repository.NoteRepository, settings *SettingsService, lockout *LockoutTracker) *NoteService {
	return &NoteService{notes: notes, settings: settings, lockout: lockout}
}

// CreateInput carries plaintext fields for a new note. It exists only
// on the stack for the duration of the call.
type CreateInput struct {
	Title   string
	Content string
	Tags    []string
	Pin     string
}

func (s *NoteService) Create(ctx context.Context, in CreateInput) (domain.NoteCard, error) {
	if err := ValidatePin(in.Pin); err != nil {
		return domain.NoteCard{}, err
	}
	if err := ValidateTitle(in.Title); err != nil {
		return domain.NoteCard{}, err
	}
	if err := ValidateTags(in.Tags); err != nil {
		return domain.NoteCard{}, err
	}

	secretKey, err := s.settings.GetDecryptedSecretKey(ctx)
	if err != nil {
		return domain.NoteCard{}, err
	}
	defer crypto.Wipe(secretKey)

	salt, err := crypto.NewSalt()
	if err != nil {
		return domain.NoteCard{}, err
	}
	nonce, err := crypto.NewNonce()
	if err != nil {
		return domain.NoteCard{}, err
	}

	key, err := crypto.DeriveKey(string(secretKey), in.Pin, salt)
	if err != nil {
		return domain.NoteCard{}, err
	}
	defer crypto.Wipe(key)

	plaintext := []byte(in.Content)
	defer crypto.Wipe(plaintext)

	ciphertext, err := crypto.Encrypt(key, nonce, plaintext)
	if err != nil {
		return domain.NoteCard{}, fmt.Errorf("service: encrypt note: %w", err)
	}

	now := time.Now().UTC()
	n := domain.Note{
		ID:               uuid.NewString(),
		Title:            in.Title,
		EncryptedContent: ciphertext,
		Salt:             salt,
		Nonce:            nonce,
		Tags:             in.Tags,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.notes.Create(ctx, n); err != nil {
		return domain.NoteCard{}, err
	}

	return domain.NoteCard{
		ID: n.ID, Title: n.Title, Tags: n.Tags,
		CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}, nil
}

// Open decrypts a note's content for viewing. Any failure — wrong PIN,
// wrong Secret Key, or corrupted data — surfaces only as
// domain.ErrInvalidPin, per spec section 4.
func (s *NoteService) Open(ctx context.Context, id, pin string) (domain.DecryptedNote, error) {
	if err := ValidatePin(pin); err != nil {
		return domain.DecryptedNote{}, err
	}

	if locked, remaining := s.lockout.Check(id); locked {
		return domain.DecryptedNote{}, fmt.Errorf("%w (%.0fs)", domain.ErrCooldownActive, remaining.Seconds())
	}

	n, err := s.notes.GetByID(ctx, id)
	if err != nil {
		return domain.DecryptedNote{}, domain.ErrInvalidPin // never reveal "not found" either
	}

	plaintext, err := s.decryptNote(ctx, n, pin)
	if err != nil {
		s.lockout.RecordFailure(id)
		return domain.DecryptedNote{}, domain.ErrInvalidPin
	}
	s.lockout.RecordSuccess(id)
	defer crypto.Wipe(plaintext)

	return domain.DecryptedNote{
		ID: n.ID, Title: n.Title, Content: string(plaintext), Tags: n.Tags,
	}, nil
}

// decryptNote is the shared low-level decrypt step used by Open, Copy and Edit.
func (s *NoteService) decryptNote(ctx context.Context, n domain.Note, pin string) ([]byte, error) {
	secretKey, err := s.settings.GetDecryptedSecretKey(ctx)
	if err != nil {
		return nil, err
	}
	defer crypto.Wipe(secretKey)

	key, err := crypto.DeriveKey(string(secretKey), pin, n.Salt)
	if err != nil {
		return nil, err
	}
	defer crypto.Wipe(key)

	plaintext, err := crypto.Decrypt(key, n.Nonce, n.EncryptedContent)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// Copy decrypts content for clipboard use only, without persisting any
// editor state. Callers must clear the OS clipboard after the
// configured timeout (handled by the app/frontend layer).
func (s *NoteService) Copy(ctx context.Context, id, pin string) (string, error) {
	dn, err := s.Open(ctx, id, pin)
	if err != nil {
		return "", err
	}
	content := dn.Content
	// dn goes out of scope; Go's GC does not guarantee wiping, but we
	// avoid retaining any extra copies beyond this return value.
	return content, nil
}

type EditInput struct {
	ID      string
	Pin     string
	Title   string
	Content string
	Tags    []string
}

func (s *NoteService) Edit(ctx context.Context, in EditInput) (domain.NoteCard, error) {
	if err := ValidateTitle(in.Title); err != nil {
		return domain.NoteCard{}, err
	}
	if err := ValidateTags(in.Tags); err != nil {
		return domain.NoteCard{}, err
	}

	n, err := s.notes.GetByID(ctx, in.ID)
	if err != nil {
		return domain.NoteCard{}, domain.ErrInvalidPin
	}

	// Verify the PIN by attempting a decrypt first (also enforces lockout).
	if _, err := s.Open(ctx, in.ID, in.Pin); err != nil {
		return domain.NoteCard{}, err
	}

	secretKey, err := s.settings.GetDecryptedSecretKey(ctx)
	if err != nil {
		return domain.NoteCard{}, err
	}
	defer crypto.Wipe(secretKey)

	// Re-encrypt with a fresh salt + nonce (never reuse a nonce).
	newSalt, err := crypto.NewSalt()
	if err != nil {
		return domain.NoteCard{}, err
	}
	newNonce, err := crypto.NewNonce()
	if err != nil {
		return domain.NoteCard{}, err
	}
	key, err := crypto.DeriveKey(string(secretKey), in.Pin, newSalt)
	if err != nil {
		return domain.NoteCard{}, err
	}
	defer crypto.Wipe(key)

	plaintext := []byte(in.Content)
	defer crypto.Wipe(plaintext)

	ciphertext, err := crypto.Encrypt(key, newNonce, plaintext)
	if err != nil {
		return domain.NoteCard{}, fmt.Errorf("service: re-encrypt note: %w", err)
	}

	n.Title = in.Title
	n.Tags = in.Tags
	n.EncryptedContent = ciphertext
	n.Salt = newSalt
	n.Nonce = newNonce
	n.UpdatedAt = time.Now().UTC()

	if err := s.notes.Update(ctx, n); err != nil {
		return domain.NoteCard{}, err
	}

	return domain.NoteCard{
		ID: n.ID, Title: n.Title, Tags: n.Tags, Favorite: n.Favorite, Pinned: n.Pinned,
		CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
	}, nil
}

func (s *NoteService) Delete(ctx context.Context, id string) error {
	return s.notes.Delete(ctx, id)
}

func (s *NoteService) SetFavorite(ctx context.Context, id string, favorite bool) error {
	return s.notes.SetFavorite(ctx, id, favorite)
}

func (s *NoteService) SetPinned(ctx context.Context, id string, pinned bool) error {
	return s.notes.SetPinned(ctx, id, pinned)
}

func (s *NoteService) List(ctx context.Context, q domain.ListQuery) ([]domain.NoteCard, int, error) {
	return s.notes.List(ctx, q)
}
