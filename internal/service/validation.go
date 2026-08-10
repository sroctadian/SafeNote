package service

import (
	"regexp"
	"unicode/utf8"

	"safenote/internal/domain"
)

const (
	MaxTitleLength = 75
	MaxTagLength   = 25
	PinLength      = 6
)

var pinPattern = regexp.MustCompile(`^\d{6}$`)

// ValidatePin enforces the PIN format: exactly 6 numeric digits.
// Applied on every create/open/edit/copy flow so no note can ever be
// created with a weak or malformed PIN, and no caller can bypass the
// rule by calling the backend directly (e.g. via devtools).
func ValidatePin(pin string) error {
	if !pinPattern.MatchString(pin) {
		return domain.ErrInvalidPinFormat
	}
	return nil
}

// ValidateTitle enforces a non-empty title of at most MaxTitleLength
// characters (counted in runes, not bytes, so multi-byte characters
// such as emoji or non-Latin scripts count as one character each).
func ValidateTitle(title string) error {
	if title == "" {
		return domain.ErrTitleRequired
	}
	if utf8.RuneCountInString(title) > MaxTitleLength {
		return domain.ErrTitleTooLong
	}
	return nil
}

// ValidateTags enforces that each tag is at most MaxTagLength characters.
func ValidateTags(tags []string) error {
	for _, t := range tags {
		if utf8.RuneCountInString(t) > MaxTagLength {
			return domain.ErrTagTooLong
		}
	}
	return nil
}
