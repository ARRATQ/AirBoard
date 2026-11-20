package handlers

import "testing"

func TestUserSortClause(t *testing.T) {
	tests := []struct{ sort, order, want string }{
		{"", "", "username ASC, id ASC"},
		{"email", "desc", "email DESC, id ASC"},
		{"last_login", "DESC", "last_login DESC NULLS LAST, id ASC"},
		{"password; DROP TABLE users", "asc", "username ASC, id ASC"},
		{"role", "sideways", "role ASC, id ASC"},
	}
	for _, tt := range tests {
		if got := userSortClause(tt.sort, tt.order); got != tt.want {
			t.Errorf("userSortClause(%q, %q) = %q, want %q", tt.sort, tt.order, got, tt.want)
		}
	}
}
