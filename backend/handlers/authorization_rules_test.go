package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// L'annuaire est partagé par toute l'organisation : seul l'admin global le gère,
// pas un admin de groupe ni un editor.
func TestCanManageDirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		role    string
		managed []uint
		want    bool
	}{
		{"admin", nil, true},
		{"editor", nil, false},
		{"user", []uint{1, 2}, false},
		{"user", nil, false},
		{"", nil, false},
	}
	for _, tc := range cases {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("role", tc.role)
		c.Set("managed_group_ids", tc.managed)
		if got := canManageDirectory(c); got != tc.want {
			t.Errorf("rôle %q (groupes gérés %v) : %v, attendu %v", tc.role, tc.managed, got, tc.want)
		}
	}
}
