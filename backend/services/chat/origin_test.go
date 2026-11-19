package chat

import (
	"net/http/httptest"
	"testing"
)

func TestCheckOrigin(t *testing.T) {
	SetAllowedOrigins([]string{"http://localhost:3000/", " https://intranet.example.com "})
	t.Cleanup(func() { SetAllowedOrigins(nil) })

	tests := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{"no Origin header (non-browser client)", "airboard.example.com", "", true},
		{"same origin behind nginx", "airboard.example.com", "https://airboard.example.com", true},
		{"same origin, different case", "Airboard.example.com", "https://airboard.EXAMPLE.com", true},
		{"allowed dev origin", "localhost:8080", "http://localhost:3000", true},
		{"allowed origin from config", "backend:8080", "https://intranet.example.com", true},
		{"third-party site is rejected", "airboard.example.com", "https://evil.example.net", false},
		{"look-alike subdomain is rejected", "airboard.example.com", "https://airboard.example.com.evil.net", false},
		{"null origin is rejected", "airboard.example.com", "null", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/ws", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if got := checkOrigin(req); got != tt.want {
				t.Errorf("checkOrigin(host=%q, origin=%q) = %v, want %v", tt.host, tt.origin, got, tt.want)
			}
		})
	}
}
