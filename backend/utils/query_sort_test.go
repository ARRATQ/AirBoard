package utils

import "testing"

func TestSafeOrderClause(t *testing.T) {
	allowed := []string{"created_at", "title"}

	tests := []struct {
		name         string
		sortBy       string
		order        string
		defaultSort  string
		defaultOrder string
		want         string
	}{
		{"allowed column and order", "title", "asc", "created_at", "desc", "title ASC"},
		{"order is case-insensitive", "title", "DeSc", "created_at", "desc", "title DESC"},
		{"unknown column falls back", "password", "asc", "created_at", "desc", "created_at ASC"},
		{"injection in column falls back", "created_at; DROP TABLE users--", "asc", "created_at", "desc", "created_at ASC"},
		{"injection in order falls back", "title", "asc, (SELECT 1)", "created_at", "desc", "title DESC"},
		{"empty values use defaults", "", "", "created_at", "desc", "created_at DESC"},
		{"invalid default order becomes ASC", "", "", "created_at", "sideways", "created_at ASC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeOrderClause(tt.sortBy, tt.order, allowed, tt.defaultSort, tt.defaultOrder)
			if got != tt.want {
				t.Errorf("SafeOrderClause(%q, %q) = %q, want %q", tt.sortBy, tt.order, got, tt.want)
			}
		})
	}
}
