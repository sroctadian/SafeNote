package domain

import "errors"

var (
	// ErrNotFound covers any missing entity (note, settings row, etc).
	ErrNotFound = errors.New("not found")

	// ErrInvalidPin is the ONLY error ever surfaced to the UI on a
	// failed unlock attempt. It intentionally does not distinguish
	// between "wrong PIN", "wrong Secret Key", or "corrupted data" —
	// spec section 4 requires that wrong-PIN failures never reveal
	// the underlying reason.
	ErrInvalidPin = errors.New("invalid PIN or unable to open note")

	// ErrCooldownActive is returned when too many failed unlock
	// attempts have triggered a temporary lockout.
	ErrCooldownActive = errors.New("too many failed attempts, please wait before trying again")

	// ErrSecretKeyNotConfigured indicates first-run setup has not
	// completed yet.
	ErrSecretKeyNotConfigured = errors.New("secret key has not been configured")

	// ErrSecretKeyTooShort enforces the 32-character minimum.
	ErrSecretKeyTooShort = errors.New("secret key must be at least 32 characters")

	// ErrDuplicateNote is returned by restore when a note with the
	// same ID already exists and the caller has not requested overwrite.
	ErrDuplicateNote = errors.New("note already exists")

	// ErrChecksumMismatch indicates a backup file failed integrity
	// validation during restore.
	ErrChecksumMismatch = errors.New("backup checksum mismatch")

	// ErrUnsupportedBackupVersion indicates a backup file format
	// version this build of SafeNote does not know how to import.
	ErrUnsupportedBackupVersion = errors.New("unsupported backup format version")

	// ErrInvalidPinFormat indicates the PIN does not meet the required
	// format: exactly 6 numeric digits.
	ErrInvalidPinFormat = errors.New("PIN must be exactly 6 digits (0-9)")

	// ErrTitleRequired indicates an empty title was submitted.
	ErrTitleRequired = errors.New("title is required")

	// ErrTitleTooLong indicates the title exceeds the 75 character limit.
	ErrTitleTooLong = errors.New("title must be at most 75 characters")

	// ErrTagTooLong indicates a tag exceeds the 25 character limit.
	ErrTagTooLong = errors.New("each tag must be at most 25 characters")
)
