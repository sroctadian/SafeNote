-- SafeNote initial schema.
-- No column here may ever hold plaintext note content or a PIN.

CREATE TABLE IF NOT EXISTS settings (
    id                       TEXT PRIMARY KEY,
    encrypted_secret         BLOB NOT NULL,
    secret_salt              BLOB NOT NULL,
    secret_nonce             BLOB NOT NULL,
    theme                    TEXT NOT NULL DEFAULT 'dark',
    clipboard_clear_seconds  INTEGER NOT NULL DEFAULT 20,
    failed_attempt_limit     INTEGER NOT NULL DEFAULT 5,
    cooldown_seconds         INTEGER NOT NULL DEFAULT 30,
    created_at               TIMESTAMP NOT NULL,
    updated_at               TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS notes (
    id                TEXT PRIMARY KEY,
    title             TEXT NOT NULL,
    encrypted_content BLOB NOT NULL,
    nonce             BLOB NOT NULL,
    salt              BLOB NOT NULL,
    tags              TEXT NOT NULL DEFAULT '[]',
    favorite          INTEGER NOT NULL DEFAULT 0,
    pinned            INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_notes_title ON notes(title);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at);
CREATE INDEX IF NOT EXISTS idx_notes_updated_at ON notes(updated_at);
CREATE INDEX IF NOT EXISTS idx_notes_pinned ON notes(pinned);
CREATE INDEX IF NOT EXISTS idx_notes_favorite ON notes(favorite);

CREATE TABLE IF NOT EXISTS backup_history (
    id         TEXT PRIMARY KEY,
    operation  TEXT NOT NULL,
    file_path  TEXT NOT NULL,
    note_count INTEGER NOT NULL,
    checksum   TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS audit_log (
    id          TEXT PRIMARY KEY,
    event       TEXT NOT NULL,
    detail      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS schema_migrations (
    version     INTEGER PRIMARY KEY,
    applied_at  TIMESTAMP NOT NULL
);
