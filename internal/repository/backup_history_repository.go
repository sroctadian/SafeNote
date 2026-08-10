package repository

import (
	"context"
	"database/sql"
	"fmt"

	"safenote/internal/domain"
)

type BackupHistoryRepository struct {
	db *sql.DB
}

func NewBackupHistoryRepository(db *sql.DB) *BackupHistoryRepository {
	return &BackupHistoryRepository{db: db}
}

func (r *BackupHistoryRepository) Record(ctx context.Context, e domain.BackupHistoryEntry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO backup_history (id, operation, file_path, note_count, checksum, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.Operation, e.FilePath, e.NoteCount, e.Checksum, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: record backup history: %w", err)
	}
	return nil
}

func (r *BackupHistoryRepository) List(ctx context.Context, limit int) ([]domain.BackupHistoryEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, operation, file_path, note_count, checksum, created_at
		FROM backup_history ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("repository: list backup history: %w", err)
	}
	defer rows.Close()

	var out []domain.BackupHistoryEntry
	for rows.Next() {
		var e domain.BackupHistoryEntry
		if err := rows.Scan(&e.ID, &e.Operation, &e.FilePath, &e.NoteCount, &e.Checksum, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// AuditLogger writes non-sensitive, append-only audit events (e.g.
// "note_created", "unlock_failed", "backup_exported"). Never logs
// PINs, Secret Keys, derived keys, or note content.
type AuditLogger struct {
	db *sql.DB
}

func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

func (a *AuditLogger) Log(ctx context.Context, id, event, detail string, createdAt any) error {
	_, err := a.db.ExecContext(ctx, `
		INSERT INTO audit_log (id, event, detail, created_at) VALUES (?, ?, ?, ?)`,
		id, event, detail, createdAt,
	)
	return err
}
