package handlers

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestEndBeforeStart(t *testing.T) {
	start := time.Date(2026, 12, 10, 9, 0, 0, 0, time.UTC)
	before := start.Add(-time.Hour)
	after := start.Add(time.Hour)
	tests := []struct {
		name       string
		start, end *time.Time
		want       bool
	}{
		{"fin avant début", &start, &before, true},
		{"fin après début", &start, &after, false},
		{"fin égale au début", &start, &start, false},
		{"sans fin", &start, nil, false},
		{"sans début", nil, &before, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endBeforeStart(tt.start, tt.end); got != tt.want {
				t.Errorf("endBeforeStart = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestValidCommentLength(t *testing.T) {
	for _, tt := range []struct {
		value int
		want  bool
	}{{0, false}, {99, false}, {100, true}, {1000, true}, {5000, true}, {5001, false}} {
		if got := validCommentLength(tt.value); got != tt.want {
			t.Errorf("validCommentLength(%d) = %v, attendu %v", tt.value, got, tt.want)
		}
	}
}

func TestIsUniqueViolation(t *testing.T) {
	unique := &pgconn.PgError{Code: "23505"}
	if !isUniqueViolation(fmt.Errorf("création : %w", unique)) {
		t.Error("une violation d'unicité enveloppée doit être reconnue")
	}
	if isUniqueViolation(&pgconn.PgError{Code: "23503"}) {
		t.Error("une violation de clé étrangère n'est pas une violation d'unicité")
	}
	if isUniqueViolation(errors.New("autre")) || isUniqueViolation(nil) {
		t.Error("une erreur quelconque ou nil n'est pas une violation d'unicité")
	}
}
