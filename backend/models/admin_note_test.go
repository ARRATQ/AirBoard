package models

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeAdminNote(t *testing.T) {
	until := time.Date(2026, 12, 1, 23, 59, 59, 0, time.UTC)

	t.Run("trims and keeps the end date", func(t *testing.T) {
		note, end, err := NormalizeAdminNote("  Maintenance jeudi 18 h  ", &until)
		if err != nil || note != "Maintenance jeudi 18 h" || end == nil || !end.Equal(until) {
			t.Fatalf("got %q %v %v", note, end, err)
		}
	})

	t.Run("empty note clears the end date", func(t *testing.T) {
		note, end, err := NormalizeAdminNote("   ", &until)
		if err != nil || note != "" || end != nil {
			t.Fatalf("got %q %v %v", note, end, err)
		}
	})

	t.Run("counts characters, not bytes", func(t *testing.T) {
		arabic := strings.Repeat("ص", AdminNoteMaxLength)
		if _, _, err := NormalizeAdminNote(arabic, nil); err != nil {
			t.Fatalf("140 Arabic letters must pass: %v", err)
		}
	})

	t.Run("rejects more than 140 characters", func(t *testing.T) {
		if _, _, err := NormalizeAdminNote(strings.Repeat("a", AdminNoteMaxLength+1), nil); err != ErrAdminNoteTooLong {
			t.Fatalf("expected ErrAdminNoteTooLong, got %v", err)
		}
	})
}
