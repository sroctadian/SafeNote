package service

import (
	"context"
	"fmt"
	"time"

	"safenote/internal/crypto"
	"safenote/internal/domain"
	"safenote/internal/repository"
)

// SettingsService manages the application-wide Secret Key and
// preferences. The Secret Key is encrypted at rest with a local
// machine-only vault key (never the SK/PIN itself), so it is never
// stored in plaintext in the database.
type SettingsService struct {
	repo      *repository.SettingsRepository
	vaultKey  []byte // 32-byte local machine key, held only in memory
}

func NewSettingsService(repo *repository.SettingsRepository, vaultKey []byte) *SettingsService {
	return &SettingsService{repo: repo, vaultKey: vaultKey}
}

func (s *SettingsService) IsConfigured(ctx context.Context) (bool, error) {
	return s.repo.Exists(ctx)
}

// SetupSecretKey performs first-run configuration (spec section 1).
func (s *SettingsService) SetupSecretKey(ctx context.Context, secretKey string) error {
	if len(secretKey) < 32 {
		return domain.ErrSecretKeyTooShort
	}

	exists, err := s.repo.Exists(ctx)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("service: secret key already configured, use ChangeSecretKey")
	}

	salt, err := crypto.NewSalt()
	if err != nil {
		return err
	}
	nonce, err := crypto.NewNonce()
	if err != nil {
		return err
	}

	// The vault key itself is combined with the (fixed) salt via a
	// lightweight KDF-free path: the vault key IS already a proper
	// 256-bit random key, so it is used directly as the AES key.
	ciphertext, err := crypto.Encrypt(s.vaultKey, nonce, []byte(secretKey))
	if err != nil {
		return fmt.Errorf("service: encrypt secret key: %w", err)
	}

	now := time.Now().UTC()
	settings := domain.DefaultSettings()
	settings.EncryptedSecret = ciphertext
	settings.SecretSalt = salt
	settings.SecretNonce = nonce
	settings.CreatedAt = now
	settings.UpdatedAt = now

	return s.repo.Create(ctx, settings)
}

// ChangeSecretKey re-wraps the Secret Key. Note: this does NOT
// re-encrypt existing notes, which remain tied to the previous Secret
// Key material used at their creation/edit time. A production rollout
// would re-encrypt all notes transactionally; see docs/ADR.md.
func (s *SettingsService) ChangeSecretKey(ctx context.Context, newSecretKey string) error {
	if len(newSecretKey) < 32 {
		return domain.ErrSecretKeyTooShort
	}

	settings, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	nonce, err := crypto.NewNonce()
	if err != nil {
		return err
	}
	ciphertext, err := crypto.Encrypt(s.vaultKey, nonce, []byte(newSecretKey))
	if err != nil {
		return fmt.Errorf("service: encrypt secret key: %w", err)
	}

	settings.EncryptedSecret = ciphertext
	settings.SecretNonce = nonce
	settings.UpdatedAt = time.Now().UTC()

	return s.repo.Update(ctx, settings)
}

// GetDecryptedSecretKey returns the raw Secret Key bytes for use in note
// key derivation. Callers MUST crypto.Wipe() the result immediately
// after use.
func (s *SettingsService) GetDecryptedSecretKey(ctx context.Context) ([]byte, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.Decrypt(s.vaultKey, settings.SecretNonce, settings.EncryptedSecret)
	if err != nil {
		return nil, fmt.Errorf("service: decrypt secret key: %w", err)
	}
	return plaintext, nil
}

// MaskedSecretKey returns a display-safe representation, never the full
// value, per spec section 1 ("never displayed completely").
func (s *SettingsService) MaskedSecretKey(ctx context.Context) (string, error) {
	raw, err := s.GetDecryptedSecretKey(ctx)
	if err != nil {
		return "", err
	}
	defer crypto.Wipe(raw)

	n := len(raw)
	if n <= 8 {
		return "********", nil
	}
	return string(raw[:4]) + "…" + fmt.Sprintf("(%d chars)", n) + "…" + string(raw[n-4:]), nil
}

func (s *SettingsService) UpdateTheme(ctx context.Context, theme string) error {
	if theme != "light" && theme != "dark" {
		return fmt.Errorf("service: invalid theme %q", theme)
	}
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}
	settings.Theme = theme
	settings.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, settings)
}

func (s *SettingsService) UpdateClipboardTimeout(ctx context.Context, seconds int) error {
	if seconds < 0 || seconds > 300 {
		return fmt.Errorf("service: clipboard timeout out of range")
	}
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}
	settings.ClipboardClearSeconds = seconds
	settings.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, settings)
}

func (s *SettingsService) Get(ctx context.Context) (domain.Settings, error) {
	return s.repo.Get(ctx)
}

// ExportedConfig is the shape written by "Export Configuration". It
// deliberately excludes the Secret Key itself.
type ExportedConfig struct {
	Theme                 string `json:"theme"`
	ClipboardClearSeconds int    `json:"clipboardClearSeconds"`
	FailedAttemptLimit    int    `json:"failedAttemptLimit"`
	CooldownSeconds       int    `json:"cooldownSeconds"`
}

func (s *SettingsService) ExportConfig(ctx context.Context) (ExportedConfig, error) {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return ExportedConfig{}, err
	}
	return ExportedConfig{
		Theme:                 settings.Theme,
		ClipboardClearSeconds: settings.ClipboardClearSeconds,
		FailedAttemptLimit:    settings.FailedAttemptLimit,
		CooldownSeconds:       settings.CooldownSeconds,
	}, nil
}

func (s *SettingsService) ImportConfig(ctx context.Context, cfg ExportedConfig) error {
	settings, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}
	if cfg.Theme == "light" || cfg.Theme == "dark" {
		settings.Theme = cfg.Theme
	}
	if cfg.ClipboardClearSeconds >= 0 && cfg.ClipboardClearSeconds <= 300 {
		settings.ClipboardClearSeconds = cfg.ClipboardClearSeconds
	}
	if cfg.FailedAttemptLimit > 0 {
		settings.FailedAttemptLimit = cfg.FailedAttemptLimit
	}
	if cfg.CooldownSeconds > 0 {
		settings.CooldownSeconds = cfg.CooldownSeconds
	}
	settings.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, settings)
}
