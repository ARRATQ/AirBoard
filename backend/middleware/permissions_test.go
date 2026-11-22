package middleware

import (
	"reflect"
	"testing"

	"airboard/models"

	"github.com/gin-gonic/gin"
)

// permContext prépare un contexte comme RequireAuth le laisse : rôle, permissions des rôles et
// groupes dont l'utilisateur est membre (managed : groupes déjà posés, ignorés désormais).
func permContext(role string, managed []uint, perms map[string]string, members []uint) *gin.Context {
	c, _ := contextFor(identity{role: role, managed: managed})
	c.Set(CtxPermissions, perms)
	c.Set(CtxMemberGroupIDs, members)
	return c
}

func TestRequirePermission(t *testing.T) {
	am := &AuthMiddleware{}
	perm := models.PermResourcesManage
	cases := []struct {
		name           string
		role           string
		managed        []uint
		perms          map[string]string
		members        []uint
		wantPass       bool
		wantUnrestrict bool
		wantManaged    []uint
	}{
		{"admin", "admin", nil, nil, nil, true, true, []uint{}},
		{"permission sur tout le site", "user", nil, map[string]string{perm: models.ScopeAll}, []uint{4}, true, true, []uint{}},
		{"permission limitée à ses groupes", "user", nil, map[string]string{perm: models.ScopeGroups}, []uint{4, 5}, true, false, []uint{4, 5}},
		{"limitée à ses groupes mais membre d'aucun", "user", nil, map[string]string{perm: models.ScopeGroups}, nil, false, false, nil},
		{"ancien admin de groupe sans rôle : plus rien", "user", []uint{7}, nil, nil, false, false, nil},
		{"autre permission seulement", "user", nil, map[string]string{"news.manage": models.ScopeAll}, []uint{4}, false, false, nil},
		{"aucun droit", "editor", nil, nil, []uint{4}, false, false, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := permContext(tc.role, tc.managed, tc.perms, tc.members)
			am.RequirePermission(perm)(c)
			if passed := !c.IsAborted(); passed != tc.wantPass {
				t.Fatalf("passe = %v, attendu %v", passed, tc.wantPass)
			}
			if !tc.wantPass {
				if c.Writer.Status() != 403 {
					t.Errorf("statut = %d, attendu 403", c.Writer.Status())
				}
				return
			}
			if got := IsUnrestricted(c); got != tc.wantUnrestrict {
				t.Errorf("IsUnrestricted = %v, attendu %v", got, tc.wantUnrestrict)
			}
			if got := GetManagedGroupIDs(c); !tc.wantUnrestrict && !reflect.DeepEqual(got, tc.wantManaged) {
				t.Errorf("groupes gérés = %v, attendu %v", got, tc.wantManaged)
			}
		})
	}
}

func TestRequirePermissionEverywhere(t *testing.T) {
	am := &AuthMiddleware{}
	perm := models.PermResourcesManage
	cases := []struct {
		name     string
		role     string
		perms    map[string]string
		managed  []uint
		wantPass bool
	}{
		{"admin", "admin", nil, nil, true},
		{"tout le site", "user", map[string]string{perm: models.ScopeAll}, nil, true},
		{"ses groupes seulement", "user", map[string]string{perm: models.ScopeGroups}, nil, false},
		{"ancien admin de groupe", "user", nil, []uint{3}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := permContext(tc.role, tc.managed, tc.perms, []uint{1})
			am.RequirePermissionEverywhere(perm)(c)
			if passed := !c.IsAborted(); passed != tc.wantPass {
				t.Fatalf("passe = %v, attendu %v", passed, tc.wantPass)
			}
			if tc.wantPass && !IsUnrestricted(c) {
				t.Error("la requête devrait être sans restriction de groupe")
			}
		})
	}
}

func TestMergePermissionKeepsWidestScope(t *testing.T) {
	perms := map[string]string{}
	models.MergePermission(perms, "p", models.ScopeGroups)
	models.MergePermission(perms, "p", models.ScopeAll)
	models.MergePermission(perms, "p", models.ScopeGroups)
	if perms["p"] != models.ScopeAll {
		t.Fatalf("portée = %q, attendu %q", perms["p"], models.ScopeAll)
	}
}
