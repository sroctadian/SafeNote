package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"safenote/internal/domain"
)

// NoteRepository persists Note entities. It never sees plaintext content:
// all encryption/decryption happens in the service layer.
type NoteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Create(ctx context.Context, n domain.Note) error {
	tagsJSON, err := json.Marshal(n.Tags)
	if err != nil {
		return fmt.Errorf("repository: marshal tags: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO notes (id, title, encrypted_content, nonce, salt, tags, favorite, pinned, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.ID, n.Title, n.EncryptedContent, n.Nonce, n.Salt, string(tagsJSON),
		boolToInt(n.Favorite), boolToInt(n.Pinned), n.CreatedAt, n.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("repository: insert note: %w", err)
	}
	return nil
}

func (r *NoteRepository) GetByID(ctx context.Context, id string) (domain.Note, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, encrypted_content, nonce, salt, tags, favorite, pinned, created_at, updated_at
		FROM notes WHERE id = ?`, id)
	return scanNote(row)
}

func (r *NoteRepository) Update(ctx context.Context, n domain.Note) error {
	tagsJSON, err := json.Marshal(n.Tags)
	if err != nil {
		return fmt.Errorf("repository: marshal tags: %w", err)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE notes SET title = ?, encrypted_content = ?, nonce = ?, salt = ?,
			tags = ?, favorite = ?, pinned = ?, updated_at = ?
		WHERE id = ?`,
		n.Title, n.EncryptedContent, n.Nonce, n.Salt, string(tagsJSON),
		boolToInt(n.Favorite), boolToInt(n.Pinned), n.UpdatedAt, n.ID,
	)
	if err != nil {
		return fmt.Errorf("repository: update note: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NoteRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM notes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("repository: delete note: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NoteRepository) SetFavorite(ctx context.Context, id string, favorite bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notes SET favorite = ?, updated_at = ? WHERE id = ?`,
		boolToInt(favorite), time.Now().UTC(), id)
	return err
}

func (r *NoteRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE notes SET pinned = ?, updated_at = ? WHERE id = ?`,
		boolToInt(pinned), time.Now().UTC(), id)
	return err
}

// List returns note cards (no ciphertext) matching the given query,
// honoring search, sort, favorite filter, and pagination. Pinned notes
// are always surfaced first regardless of sort mode.
func (r *NoteRepository) List(ctx context.Context, q domain.ListQuery) ([]domain.NoteCard, int, error) {
	where := "WHERE 1 = 1"
	args := []any{}

	if q.Search != "" {
		where += " AND title LIKE ? ESCAPE '\\'"
		args = append(args, "%"+escapeLike(q.Search)+"%")
	}
	if q.OnlyFav {
		where += " AND favorite = 1"
	}

	var total int
	countQuery := "SELECT COUNT(1) FROM notes " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("repository: count notes: %w", err)
	}

	orderBy := "pinned DESC, created_at DESC"
	switch q.Sort {
	case domain.SortOldest:
		orderBy = "pinned DESC, created_at ASC"
	case domain.SortAlphabet:
		orderBy = "pinned DESC, title COLLATE NOCASE ASC"
	case domain.SortNewest, "":
		orderBy = "pinned DESC, created_at DESC"
	}

	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	listQuery := fmt.Sprintf(`
		SELECT id, title, tags, favorite, pinned, created_at, updated_at
		FROM notes %s ORDER BY %s LIMIT ? OFFSET ?`, where, orderBy)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list notes: %w", err)
	}
	defer rows.Close()

	var cards []domain.NoteCard
	for rows.Next() {
		var c domain.NoteCard
		var tagsJSON string
		var fav, pinned int
		if err := rows.Scan(&c.ID, &c.Title, &tagsJSON, &fav, &pinned, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("repository: scan note card: %w", err)
		}
		_ = json.Unmarshal([]byte(tagsJSON), &c.Tags)
		c.Favorite = fav == 1
		c.Pinned = pinned == 1
		cards = append(cards, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return cards, total, nil
}

// AllForBackup returns every note in raw (encrypted) form, used only by
// the backup module. It never decrypts anything.
func (r *NoteRepository) AllForBackup(ctx context.Context) ([]domain.Note, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, encrypted_content, nonce, salt, tags, favorite, pinned, created_at, updated_at
		FROM notes ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("repository: list all notes: %w", err)
	}
	defer rows.Close()

	var notes []domain.Note
	for rows.Next() {
		n, err := scanNoteRows(rows)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM notes WHERE id = ?`, id).Scan(&count)
	return count > 0, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNote(s scanner) (domain.Note, error) {
	var n domain.Note
	var tagsJSON string
	var fav, pinned int
	err := s.Scan(&n.ID, &n.Title, &n.EncryptedContent, &n.Nonce, &n.Salt, &tagsJSON, &fav, &pinned, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return domain.Note{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Note{}, fmt.Errorf("repository: scan note: %w", err)
	}
	_ = json.Unmarshal([]byte(tagsJSON), &n.Tags)
	n.Favorite = fav == 1
	n.Pinned = pinned == 1
	return n, nil
}

func scanNoteRows(rows *sql.Rows) (domain.Note, error) {
	return scanNote(rows)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// escapeLike escapes SQLite LIKE wildcard characters in user input so a
// search for e.g. "50%" or "a_b" behaves as a literal substring match.
func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
