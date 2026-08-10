package repository

import (
	"context"
	"database/sql"
	"fmt"

	"safenote/internal/domain"
)

const settingsRowID = "singleton"

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context) (domain.Settings, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, encrypted_secret, secret_salt, secret_nonce, theme,
			clipboard_clear_seconds, failed_attempt_limit, cooldown_seconds,
			created_at, updated_at
		FROM settings WHERE id = ?`, settingsRowID)

	var s domain.Settings
	err := row.Scan(&s.ID, &s.EncryptedSecret, &s.SecretSalt, &s.SecretNonce, &s.Theme,
		&s.ClipboardClearSeconds, &s.FailedAttemptLimit, &s.CooldownSeconds,
		&s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Settings{}, domain.ErrSecretKeyNotConfigured
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("repository: get settings: %w", err)
	}
	return s, nil
}

func (r *SettingsRepository) Exists(ctx context.Context) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM settings WHERE id = ?`, settingsRowID).Scan(&count)
	return count > 0, err
}

func (r *SettingsRepository) Create(ctx context.Context, s domain.Settings) error {
	s.ID = settingsRowID
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO settings (id, encrypted_secret, secret_salt, secret_nonce, theme,
			clipboard_clear_seconds, failed_attempt_limit, cooldown_seconds, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.EncryptedSecret, s.SecretSalt, s.SecretNonce, s.Theme,
		s.ClipboardClearSeconds, s.FailedAttemptLimit, s.CooldownSeconds, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: create settings: %w", err)
	}
	return nil
}

func (r *SettingsRepository) Update(ctx context.Context, s domain.Settings) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE settings SET encrypted_secret = ?, secret_salt = ?, secret_nonce = ?,
			theme = ?, clipboard_clear_seconds = ?, failed_attempt_limit = ?,
			cooldown_seconds = ?, updated_at = ?
		WHERE id = ?`,
		s.EncryptedSecret, s.SecretSalt, s.SecretNonce, s.Theme,
		s.ClipboardClearSeconds, s.FailedAttemptLimit, s.CooldownSeconds, s.UpdatedAt, settingsRowID,
	)
	if err != nil {
		return fmt.Errorf("repository: update settings: %w", err)
	}
	return nil
}
