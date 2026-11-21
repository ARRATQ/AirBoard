package models

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// AdminNoteMaxLength est la longueur maximale (en caractères) de la note d'une ressource.
const AdminNoteMaxLength = 140

// ErrAdminNoteTooLong est renvoyée quand la note dépasse AdminNoteMaxLength caractères.
var ErrAdminNoteTooLong = errors.New("la note ne doit pas dépasser 140 caractères")

// NormalizeAdminNote nettoie la note d'une ressource : espaces retirés, longueur bornée,
// date de fin effacée quand la note est vide.
func NormalizeAdminNote(note string, until *time.Time) (string, *time.Time, error) {
	trimmed := strings.TrimSpace(note)
	if utf8.RuneCountInString(trimmed) > AdminNoteMaxLength {
		return "", nil, ErrAdminNoteTooLong
	}
	if trimmed == "" {
		return "", nil, nil
	}
	return trimmed, until, nil
}
