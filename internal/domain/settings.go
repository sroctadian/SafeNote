package domain

import "time"

// Settings holds the single application-wide configuration row.
// EncryptedSecret stores the Secret Key wrapped by an OS-level secret
// store when available; SafeNote never writes the raw Secret Key.
type Settings struct {
	ID                    string    `json:"id"`
	EncryptedSecret       []byte    `json:"-"`
	SecretSalt            []byte    `json:"-"`
	SecretNonce           []byte    `json:"-"`
	Theme                 string    `json:"theme"` // "light" | "dark"
	ClipboardClearSeconds int       `json:"clipboardClearSeconds"`
	FailedAttemptLimit    int       `json:"failedAttemptLimit"`
	CooldownSeconds       int       `json:"cooldownSeconds"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

// DefaultSettings returns sane defaults applied on first run.
func DefaultSettings() Settings {
	return Settings{
		Theme:                 "dark",
		ClipboardClearSeconds: 20,
		FailedAttemptLimit:    5,
		CooldownSeconds:       30,
	}
}

// BackupFile is the on-disk JSON structure exported by the Backup module.
// Notes remain fully encrypted; SafeNote never decrypts during backup.
type BackupFile struct {
	Version    string           `json:"version"`
	CreatedAt  time.Time        `json:"created_at"`
	Checksum   string           `json:"checksum"`
	Notes      []BackupNote     `json:"notes"`
}

// BackupNote is the encrypted, portable form of a note.
type BackupNote struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent string    `json:"encrypted_content"` // base64
	Salt             string    `json:"salt"`              // base64
	Nonce            string    `json:"nonce"`              // base64
	Tags             []string  `json:"tags"`
	Favorite         bool      `json:"favorite"`
	Pinned           bool      `json:"pinned"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// BackupHistoryEntry records metadata about a past backup or restore
// operation for audit purposes (never contains note content).
type BackupHistoryEntry struct {
	ID        string    `json:"id"`
	Operation string    `json:"operation"` // "backup" | "restore"
	FilePath  string    `json:"filePath"`
	NoteCount int       `json:"noteCount"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"createdAt"`
}
