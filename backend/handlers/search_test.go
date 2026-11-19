package handlers

import "testing"

func TestEscapeLikePattern(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"l'équipe", "l'équipe"},
		{"100%", `100\%`},
		{"snake_case", `snake\_case`},
		{`C:\temp`, `C:\\temp`},
		{`a\%b`, `a\\\%b`},
	}
	for _, tt := range tests {
		if got := escapeLikePattern(tt.in); got != tt.want {
			t.Errorf("escapeLikePattern(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
