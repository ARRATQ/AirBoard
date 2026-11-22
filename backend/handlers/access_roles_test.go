package handlers

import (
	"strings"
	"testing"

	"airboard/models"
)

func roleReq(name string, perms ...[2]string) accessRoleRequest {
	req := accessRoleRequest{Name: name}
	for _, p := range perms {
		req.Permissions = append(req.Permissions, struct {
			Permission string `json:"permission"`
			Scope      string `json:"scope"`
		}{p[0], p[1]})
	}
	return req
}

func TestValidateAccessRole(t *testing.T) {
	ok := [2]string{models.PermResourcesManage, models.ScopeGroups}
	cases := []struct {
		name    string
		req     accessRoleRequest
		wantErr string
	}{
		{"valide", roleReq("  Éditeur  ", ok), ""},
		{"sans permission", roleReq("Lecteur"), ""},
		{"nom vide", roleReq("   ", ok), "obligatoire"},
		{"nom trop long", roleReq(strings.Repeat("é", accessRoleNameMax+1)), "80 caractères"},
		{"nom réservé", roleReq("Administrateur"), "réservé"},
		{"permission inconnue", roleReq("X", [2]string{"users.delete", models.ScopeAll}), "inconnue"},
		{"portée inconnue", roleReq("X", [2]string{models.PermResourcesManage, "partout"}), "Portée invalide"},
		{"permission en double", roleReq("X", ok, ok), "en double"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			role, perms, msg := validateAccessRole(tc.req)
			if tc.wantErr == "" {
				if msg != "" {
					t.Fatalf("erreur inattendue : %s", msg)
				}
				if role.Name != strings.TrimSpace(tc.req.Name) || len(perms) != len(tc.req.Permissions) {
					t.Fatalf("rôle = %+v, permissions = %+v", role, perms)
				}
				return
			}
			if !strings.Contains(msg, tc.wantErr) {
				t.Fatalf("erreur = %q, attendu qui contient %q", msg, tc.wantErr)
			}
		})
	}
}

func TestUniqueUint(t *testing.T) {
	got := uniqueUint([]uint{3, 0, 3, 1})
	if len(got) != 2 || got[0] != 3 || got[1] != 1 {
		t.Fatalf("uniqueUint = %v", got)
	}
}
