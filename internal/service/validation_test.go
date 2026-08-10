package service

import (
	"context"
	"strings"
	"testing"

	"safenote/internal/domain"
)

func TestValidatePin(t *testing.T) {
	cases := map[string]bool{
		"123456":  true,
		"000000":  true,
		"999999":  true,
		"12345":   false, // too short
		"1234567": false, // too long
		"12345a":  false, // non-numeric
		"":        false,
		"abcdef":  false,
		" 123456": false, // whitespace not allowed
	}
	for pin, wantValid := range cases {
		err := ValidatePin(pin)
		gotValid := err == nil
		if gotValid != wantValid {
			t.Errorf("ValidatePin(%q) valid=%v, want %v (err=%v)", pin, gotValid, wantValid, err)
		}
		if !wantValid && err != domain.ErrInvalidPinFormat {
			t.Errorf("ValidatePin(%q) expected ErrInvalidPinFormat, got %v", pin, err)
		}
	}
}

func TestValidateTitle(t *testing.T) {
	if err := ValidateTitle(""); err != domain.ErrTitleRequired {
		t.Errorf("expected ErrTitleRequired for empty title, got %v", err)
	}

	exactly75 := strings.Repeat("a", 75)
	if err := ValidateTitle(exactly75); err != nil {
		t.Errorf("expected 75-char title to be valid, got %v", err)
	}

	tooLong := strings.Repeat("a", 76)
	if err := ValidateTitle(tooLong); err != domain.ErrTitleTooLong {
		t.Errorf("expected ErrTitleTooLong for 76-char title, got %v", err)
	}
}

func TestValidateTags(t *testing.T) {
	ok := []string{strings.Repeat("t", 25), "short"}
	if err := ValidateTags(ok); err != nil {
		t.Errorf("expected valid tags, got %v", err)
	}

	tooLong := []string{strings.Repeat("t", 26)}
	if err := ValidateTags(tooLong); err != domain.ErrTagTooLong {
		t.Errorf("expected ErrTagTooLong, got %v", err)
	}
}

func TestNoteService_CreateRejectsInvalidPinFormat(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	_, err := noteSvc.Create(ctx, CreateInput{Title: "N", Content: "c", Pin: "12"})
	if err != domain.ErrInvalidPinFormat {
		t.Fatalf("expected ErrInvalidPinFormat, got %v", err)
	}
}

func TestNoteService_CreateRejectsTitleTooLong(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	_, err := noteSvc.Create(ctx, CreateInput{Title: strings.Repeat("x", 76), Content: "c", Pin: "123456"})
	if err != domain.ErrTitleTooLong {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}

func TestNoteService_CreateRejectsTagTooLong(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	_, err := noteSvc.Create(ctx, CreateInput{
		Title: "N", Content: "c", Pin: "123456", Tags: []string{strings.Repeat("t", 26)},
	})
	if err != domain.ErrTagTooLong {
		t.Fatalf("expected ErrTagTooLong, got %v", err)
	}
}

func TestNoteService_OpenRejectsInvalidPinFormat(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	card, _ := noteSvc.Create(ctx, CreateInput{Title: "N", Content: "c", Pin: "123456"})

	_, err := noteSvc.Open(ctx, card.ID, "12")
	if err != domain.ErrInvalidPinFormat {
		t.Fatalf("expected ErrInvalidPinFormat, got %v", err)
	}
}

func TestNoteService_EditRejectsTitleTooLong(t *testing.T) {
	ctx := context.Background()
	noteSvc, _ := newTestServices(t)

	card, _ := noteSvc.Create(ctx, CreateInput{Title: "N", Content: "c", Pin: "123456"})

	_, err := noteSvc.Edit(ctx, EditInput{
		ID: card.ID, Pin: "123456", Title: strings.Repeat("x", 76), Content: "c2",
	})
	if err != domain.ErrTitleTooLong {
		t.Fatalf("expected ErrTitleTooLong, got %v", err)
	}
}
