package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// identity : ce que RequireAuth dépose dans le contexte (rôle relu en base,
// groupes administrés). Un champ nil n'est pas posé.
type identity struct {
	role    interface{}
	managed interface{}
}

func contextFor(who identity) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if who.role != nil {
		c.Set("role", who.role)
	}
	if who.managed != nil {
		c.Set("managed_group_ids", who.managed)
	}
	return c, w
}

// runGuard exécute un middleware puis, s'il laisse passer, un handler témoin.
func runGuard(t *testing.T, guard gin.HandlerFunc, who identity) (status int, passed bool, code string) {
	t.Helper()
	c, w := contextFor(who)
	guard(c)
	if !c.IsAborted() {
		passed = true
		c.Status(http.StatusOK)
	}
	if w.Body.Len() > 0 {
		var body struct {
			Error string `json:"error"`
			Code  string `json:"code"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("corps d'erreur illisible %q : %v", w.Body.String(), err)
		}
		if body.Error == "" {
			t.Errorf("corps d'erreur sans message : %q", w.Body.String())
		}
		code = body.Code
	}
	return w.Code, passed, code
}

var (
	admin       = identity{role: "admin"}
	editor      = identity{role: "editor"}
	plainUser   = identity{role: "user", managed: []uint{}}
	groupAdmin  = identity{role: "user", managed: []uint{3, 7}}
	banned      = identity{role: "banned", managed: []uint{3}}
	emptyRole   = identity{role: "", managed: []uint{3}}
	anonymous   = identity{}
	badRoleType = identity{role: 42}
	badManaged  = identity{role: "user", managed: []int{3}}
)

func TestRoleGuards(t *testing.T) {
	am := &AuthMiddleware{}
	guards := map[string]gin.HandlerFunc{
		"RequireAdmin":         am.RequireAdmin(),
		"RequireEditor":        am.RequireEditor(),
		"RequireEditorOrAdmin": am.RequireEditorOrAdmin(),
		"RequireGroupAdmin":    am.RequireGroupAdmin(),
	}
	cases := []struct {
		guard  string
		name   string
		who    identity
		status int
		code   string
	}{
		{"RequireAdmin", "admin", admin, http.StatusOK, ""},
		{"RequireAdmin", "editor", editor, http.StatusForbidden, "forbidden"},
		{"RequireAdmin", "admin de groupe", groupAdmin, http.StatusForbidden, "forbidden"},
		{"RequireAdmin", "anonyme", anonymous, http.StatusForbidden, "forbidden"},

		{"RequireEditor", "admin", admin, http.StatusOK, ""},
		{"RequireEditor", "editor", editor, http.StatusOK, ""},
		{"RequireEditor", "utilisateur", plainUser, http.StatusForbidden, "forbidden"},
		{"RequireEditor", "admin de groupe", groupAdmin, http.StatusForbidden, "forbidden"},
		{"RequireEditor", "rôle non textuel", badRoleType, http.StatusForbidden, "forbidden"},
		{"RequireEditor", "anonyme", anonymous, http.StatusUnauthorized, "unauthorized"},
		{"RequireEditorOrAdmin", "editor", editor, http.StatusOK, ""},
		{"RequireEditorOrAdmin", "utilisateur", plainUser, http.StatusForbidden, "forbidden"},

		{"RequireGroupAdmin", "admin", admin, http.StatusOK, ""},
		{"RequireGroupAdmin", "admin de groupe", groupAdmin, http.StatusOK, ""},
		{"RequireGroupAdmin", "utilisateur sans groupe géré", plainUser, http.StatusForbidden, "forbidden"},
		{"RequireGroupAdmin", "editor sans groupe géré", editor, http.StatusForbidden, "forbidden"},
		{"RequireGroupAdmin", "banni qui gère un groupe", banned, http.StatusForbidden, "forbidden"},
		{"RequireGroupAdmin", "rôle vide", emptyRole, http.StatusForbidden, "forbidden"},
		{"RequireGroupAdmin", "rôle non textuel", badRoleType, http.StatusForbidden, "forbidden"},
		{"RequireGroupAdmin", "groupes gérés mal typés", badManaged, http.StatusForbidden, "forbidden"},
		{"RequireGroupAdmin", "anonyme", anonymous, http.StatusUnauthorized, "unauthorized"},
	}
	for _, tc := range cases {
		t.Run(tc.guard+"/"+tc.name, func(t *testing.T) {
			status, passed, code := runGuard(t, guards[tc.guard], tc.who)
			if status != tc.status || passed != (tc.status == http.StatusOK) || code != tc.code {
				t.Errorf("statut %d, passé %v, code %q ; attendu %d, %q", status, passed, code, tc.status, tc.code)
			}
		})
	}
}

func TestCanManageGroup(t *testing.T) {
	cases := []struct {
		name  string
		who   identity
		group uint
		want  bool
	}{
		{"admin : tout groupe", admin, 99, true},
		{"admin de groupe : son groupe", groupAdmin, 7, true},
		{"admin de groupe : autre groupe", groupAdmin, 4, false},
		{"utilisateur", plainUser, 3, false},
		{"sans groupes gérés dans le contexte", editor, 3, false},
		{"groupes gérés mal typés", badManaged, 3, false},
	}
	for _, tc := range cases {
		c, _ := contextFor(tc.who)
		if got := CanManageGroup(c, tc.group); got != tc.want {
			t.Errorf("%s : %v, attendu %v", tc.name, got, tc.want)
		}
	}
}

func TestGetManagedGroupIDs(t *testing.T) {
	cases := []struct {
		name string
		who  identity
		want int
	}{
		{"admin de groupe", groupAdmin, 2},
		{"absent", admin, 0},
		{"mal typé", badManaged, 0},
	}
	for _, tc := range cases {
		c, _ := contextFor(tc.who)
		got := GetManagedGroupIDs(c)
		if got == nil || len(got) != tc.want {
			t.Errorf("%s : %v", tc.name, got)
		}
	}
}

// CanManageAppGroupWithDB : les cas tranchés avant toute requête.
func TestCanManageAppGroupWithoutQuery(t *testing.T) {
	cases := []struct {
		name string
		who  identity
		db   interface{}
		want bool
	}{
		{"admin : tout groupe d'applications", admin, nil, true},
		{"aucun groupe géré", plainUser, &gorm.DB{}, false},
		{"base invalide", groupAdmin, "pas une base", false},
	}
	for _, tc := range cases {
		c, _ := contextFor(tc.who)
		if got := CanManageAppGroupWithDB(c, 1, tc.db); got != tc.want {
			t.Errorf("%s : %v, attendu %v", tc.name, got, tc.want)
		}
	}
}
