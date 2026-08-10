package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"safenote/internal/domain"
	"safenote/internal/repository"
)

const BackupFormatVersion = "1"

type BackupService struct {
	notes   *repository.NoteRepository
	history *repository.BackupHistoryRepository
}

func NewBackupService(notes *repository.NoteRepository, history *repository.BackupHistoryRepository) *BackupService {
	return &BackupService{notes: notes, history: history}
}

// Export writes all notes, still fully encrypted, to path as JSON.
// SafeNote never decrypts note content during backup (spec section 8).
func (s *BackupService) Export(ctx context.Context, path string) (domain.BackupFile, error) {
	notes, err := s.notes.AllForBackup(ctx)
	if err != nil {
		return domain.BackupFile{}, err
	}

	backupNotes := make([]domain.BackupNote, 0, len(notes))
	for _, n := range notes {
		backupNotes = append(backupNotes, domain.BackupNote{
			ID:               n.ID,
			Title:            n.Title,
			EncryptedContent: base64.StdEncoding.EncodeToString(n.EncryptedContent),
			Salt:             base64.StdEncoding.EncodeToString(n.Salt),
			Nonce:            base64.StdEncoding.EncodeToString(n.Nonce),
			Tags:             n.Tags,
			Favorite:         n.Favorite,
			Pinned:           n.Pinned,
			CreatedAt:        n.CreatedAt,
			UpdatedAt:        n.UpdatedAt,
		})
	}

	bf := domain.BackupFile{
		Version:   BackupFormatVersion,
		CreatedAt: time.Now().UTC(),
		Notes:     backupNotes,
	}
	bf.Checksum = checksumOf(backupNotes)

	raw, err := json.MarshalIndent(bf, "", "  ")
	if err != nil {
		return domain.BackupFile{}, fmt.Errorf("service: marshal backup: %w", err)
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		return domain.BackupFile{}, fmt.Errorf("service: write backup file: %w", err)
	}

	_ = s.history.Record(ctx, domain.BackupHistoryEntry{
		ID: uuid.NewString(), Operation: "backup", FilePath: path,
		NoteCount: len(backupNotes), Checksum: bf.Checksum, CreatedAt: time.Now().UTC(),
	})

	return bf, nil
}

// RestorePreview validates a backup file (checksum, version, duplicates)
// without writing anything, so the UI can present a confirmation before
// any import happens (spec section 9: "no automatic overwrite").
type RestorePreview struct {
	Valid           bool     `json:"valid"`
	NoteCount       int      `json:"noteCount"`
	DuplicateIDs    []string `json:"duplicateIds"`
	FormatVersion   string   `json:"formatVersion"`
}

func (s *BackupService) RestorePreview(ctx context.Context, path string) (RestorePreview, error) {
	bf, err := s.readAndValidate(path)
	if err != nil {
		return RestorePreview{}, err
	}

	var duplicates []string
	for _, n := range bf.Notes {
		exists, err := s.notes.ExistsByID(ctx, n.ID)
		if err != nil {
			return RestorePreview{}, err
		}
		if exists {
			duplicates = append(duplicates, n.ID)
		}
	}

	return RestorePreview{
		Valid: true, NoteCount: len(bf.Notes),
		DuplicateIDs: duplicates, FormatVersion: bf.Version,
	}, nil
}

// Restore imports encrypted notes from a validated backup file.
// skipIDs (typically duplicates the user chose not to overwrite) are
// excluded; SafeNote never overwrites silently.
func (s *BackupService) Restore(ctx context.Context, path string, overwriteIDs map[string]bool) (int, error) {
	bf, err := s.readAndValidate(path)
	if err != nil {
		return 0, err
	}

	imported := 0
	for _, bn := range bf.Notes {
		exists, err := s.notes.ExistsByID(ctx, bn.ID)
		if err != nil {
			return imported, err
		}
		if exists && !overwriteIDs[bn.ID] {
			continue // no automatic overwrite
		}

		content, err := base64.StdEncoding.DecodeString(bn.EncryptedContent)
		if err != nil {
			return imported, fmt.Errorf("service: decode backup note %s: %w", bn.ID, err)
		}
		salt, err := base64.StdEncoding.DecodeString(bn.Salt)
		if err != nil {
			return imported, fmt.Errorf("service: decode backup salt %s: %w", bn.ID, err)
		}
		nonce, err := base64.StdEncoding.DecodeString(bn.Nonce)
		if err != nil {
			return imported, fmt.Errorf("service: decode backup nonce %s: %w", bn.ID, err)
		}

		n := domain.Note{
			ID: bn.ID, Title: bn.Title, EncryptedContent: content,
			Salt: salt, Nonce: nonce, Tags: bn.Tags,
			Favorite: bn.Favorite, Pinned: bn.Pinned,
			CreatedAt: bn.CreatedAt, UpdatedAt: bn.UpdatedAt,
		}

		if exists {
			err = s.notes.Update(ctx, n)
		} else {
			err = s.notes.Create(ctx, n)
		}
		if err != nil {
			return imported, fmt.Errorf("service: import note %s: %w", bn.ID, err)
		}
		imported++
	}

	_ = s.history.Record(ctx, domain.BackupHistoryEntry{
		ID: uuid.NewString(), Operation: "restore", FilePath: path,
		NoteCount: imported, Checksum: bf.Checksum, CreatedAt: time.Now().UTC(),
	})

	return imported, nil
}

func (s *BackupService) readAndValidate(path string) (domain.BackupFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return domain.BackupFile{}, fmt.Errorf("service: read backup file: %w", err)
	}

	var bf domain.BackupFile
	if err := json.Unmarshal(raw, &bf); err != nil {
		return domain.BackupFile{}, fmt.Errorf("service: parse backup file: %w", err)
	}

	if bf.Version != BackupFormatVersion {
		return domain.BackupFile{}, domain.ErrUnsupportedBackupVersion
	}

	if checksumOf(bf.Notes) != bf.Checksum {
		return domain.BackupFile{}, domain.ErrChecksumMismatch
	}

	return bf, nil
}

// checksumOf computes a stable SHA-256 checksum over the backup note
// set so tampering or truncation can be detected before import.
func checksumOf(notes []domain.BackupNote) string {
	h := sha256.New()
	for _, n := range notes {
		fmt.Fprintf(h, "%s|%s|%s|%s|%s\n", n.ID, n.Title, n.EncryptedContent, n.Salt, n.Nonce)
	}
	return hex.EncodeToString(h.Sum(nil))
}
