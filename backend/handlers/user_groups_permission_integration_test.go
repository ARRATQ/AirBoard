//go:build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Permission groups.manage : page « Mes groupes » (/group-admin/user-groups).

// callAs exécute un handler pour cette identité, avec un corps JSON, :id et une requête.
func callAs(h gin.HandlerFunc, who identity, method string, id uint, target string, body interface{}) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var raw []byte
	if body != nil {
		raw, _ = json.Marshal(body)
	}
	if target == "" {
		target = "/"
	}
	c.Request = httptest.NewRequest(method, target, bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	if id != 0 {
		c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(id), 10)}}
	}
	c.Set("user_id", who.id)
	c.Set("role", who.role)
	c.Set(middleware.CtxPermissions, who.perms)
	c.Set(middleware.CtxPermissionGroups, groupsForPerms(who.perms, who.managed))
	h(c)
	return w
}

type userGroupsFixture struct {
	root, sub, linked, other models.Group
	holder, member, outsider identity
	allHolder                identity
}

func newUserGroupsFixture(t *testing.T, tx *gorm.DB) userGroupsFixture {
	t.Helper()
	tag := hierarchyTag()
	var f userGroupsFixture
	f.root = models.Group{Name: tag + "-equipe"}
	mustCreate(t, tx, &f.root)
	f.sub = models.Group{Name: tag + "-sous", ParentID: &f.root.ID}
	mustCreate(t, tx, &f.sub)
	f.linked = models.Group{Name: tag + "-lie", ParentID: &f.root.ID}
	mustCreate(t, tx, &f.linked)
	mustCreate(t, tx, &models.OrgUnit{Name: tag + " Entité", GroupID: &f.linked.ID})
	f.other = models.Group{Name: tag + "-autre"}
	mustCreate(t, tx, &f.other)

	user := func(name string, groups ...models.Group) models.User {
		u := models.User{Username: tag + name, Email: tag + name + "@example.test", Role: "user", IsActive: true, Groups: groups}
		mustCreate(t, tx, &u)
		return u
	}
	holder := user("holder", f.root)
	member := user("member", f.sub) // membre du groupe couvert par son sous-groupe
	outsider := user("outsider", f.other)
	all := user("all")
	perm := map[string]string{models.PermGroupsManage: models.ScopeGroups}
	grantPermissions(t, tx, tag+"-sous-groupes", perm, holder.ID)
	permAll := map[string]string{models.PermGroupsManage: models.ScopeAll}
	grantPermissions(t, tx, tag+"-tous-groupes", permAll, all.ID)

	covered, err := models.PermissionGroupIDs(tx, holder.ID)
	if err != nil {
		t.Fatal(err)
	}
	f.holder = identity{id: holder.ID, role: "user", managed: covered, perms: perm}
	f.member = identity{id: member.ID, role: "user"}
	f.outsider = identity{id: outsider.ID, role: "user"}
	f.allHolder = identity{id: all.ID, role: "user", perms: permAll}
	return f
}

func decodeGroup(t *testing.T, w *httptest.ResponseRecorder) models.Group {
	t.Helper()
	var g models.Group
	if err := json.Unmarshal(w.Body.Bytes(), &g); err != nil {
		t.Fatalf("réponse illisible : %s", w.Body.String())
	}
	return g
}

func TestUserGroupsPermissionScopedToOwnGroups(t *testing.T) {
	tx := rollbackTx(t)
	f := newUserGroupsFixture(t, tx)
	h := NewGroupAdminHandler(tx)
	tag := hierarchyTag()

	t.Run("la liste montre ses groupes et dit lesquels sont modifiables", func(t *testing.T) {
		w := callAs(h.GetUserGroups, f.holder, http.MethodGet, 0, "", nil)
		var res struct {
			Groups []struct {
				ID       uint `json:"id"`
				Editable bool `json:"editable"`
			} `json:"groups"`
			CanCreateRoot bool `json:"can_create_root"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		editable := map[uint]bool{}
		for _, g := range res.Groups {
			editable[g.ID] = g.Editable
		}
		if _, ok := editable[f.other.ID]; ok {
			t.Fatal("un groupe extérieur est listé")
		}
		if editable[f.root.ID] || !editable[f.sub.ID] || editable[f.linked.ID] || res.CanCreateRoot {
			t.Fatalf("modifiables = %v (racine %d, sous %d, lié %d), can_create_root %v", editable, f.root.ID, f.sub.ID, f.linked.ID, res.CanCreateRoot)
		}
	})

	var created models.Group
	t.Run("crée un sous-groupe de son groupe avec des membres de ce groupe", func(t *testing.T) {
		w := callAs(h.CreateUserGroup, f.holder, http.MethodPost, 0, "", map[string]interface{}{
			"name": tag + "-nouveau", "parent_id": f.root.ID, "user_ids": []uint{f.member.id, f.holder.id},
		})
		if w.Code != http.StatusCreated {
			t.Fatalf("statut %d %s", w.Code, w.Body.String())
		}
		created = decodeGroup(t, w)
		if created.ParentID == nil || *created.ParentID != f.root.ID {
			t.Fatalf("parent = %v", created.ParentID)
		}
		if s := membershipSource(t, tx, f.member.id, created.ID); s != "manual" {
			t.Fatalf("membre absent (%q)", s)
		}
	})

	refusals := []struct {
		name string
		body map[string]interface{}
		code int
	}{
		{"un membre extérieur", map[string]interface{}{"name": tag + "-x1", "parent_id": f.root.ID, "user_ids": []uint{f.outsider.id}}, http.StatusForbidden},
		{"sous un groupe qu'il ne couvre pas", map[string]interface{}{"name": tag + "-x2", "parent_id": f.other.ID}, http.StatusForbidden},
		{"un groupe principal", map[string]interface{}{"name": tag + "-x3", "parent_id": nil}, http.StatusForbidden},
		{"un nom déjà pris", map[string]interface{}{"name": f.other.Name, "parent_id": f.root.ID}, http.StatusConflict},
		{"un nom vide", map[string]interface{}{"name": "  ", "parent_id": f.root.ID}, http.StatusBadRequest},
	}
	for _, r := range refusals {
		t.Run("refuse "+r.name, func(t *testing.T) {
			if w := callAs(h.CreateUserGroup, f.holder, http.MethodPost, 0, "", r.body); w.Code != r.code {
				t.Fatalf("statut %d %s, attendu %d", w.Code, w.Body.String(), r.code)
			}
		})
	}

	t.Run("ne modifie ni son groupe couvert, ni un groupe lié au trombinoscope, ni un groupe extérieur", func(t *testing.T) {
		for id, want := range map[uint]int{f.root.ID: http.StatusForbidden, f.linked.ID: http.StatusForbidden, f.other.ID: http.StatusNotFound} {
			if w := callAs(h.UpdateUserGroup, f.holder, http.MethodPut, id, "", map[string]interface{}{"name": tag + "-renomme"}); w.Code != want {
				t.Errorf("groupe %d : statut %d, attendu %d", id, w.Code, want)
			}
			if w := callAs(h.DeleteUserGroup, f.holder, http.MethodDelete, id, "", nil); w.Code != want {
				t.Errorf("suppression du groupe %d : statut %d, attendu %d", id, w.Code, want)
			}
		}
	})

	t.Run("modifie son sous-groupe ; ne le déplace pas hors de ses groupes", func(t *testing.T) {
		w := callAs(h.UpdateUserGroup, f.holder, http.MethodPut, created.ID, "", map[string]interface{}{"name": tag + "-renomme", "parent_id": f.sub.ID, "user_ids": []uint{f.member.id}})
		if w.Code != http.StatusOK {
			t.Fatalf("statut %d %s", w.Code, w.Body.String())
		}
		if s := membershipSource(t, tx, f.holder.id, created.ID); s != "" {
			t.Fatal("membre retiré toujours présent")
		}
		if w := callAs(h.UpdateUserGroup, f.holder, http.MethodPut, created.ID, "", map[string]interface{}{"parent_id": f.other.ID}); w.Code != http.StatusForbidden {
			t.Fatalf("déplacement hors de ses groupes : statut %d", w.Code)
		}
		if w := callAs(h.UpdateUserGroup, f.holder, http.MethodPut, created.ID, "", map[string]interface{}{"user_ids": []uint{f.member.id, f.outsider.id}}); w.Code != http.StatusForbidden {
			t.Fatalf("ajout d'un extérieur : statut %d", w.Code)
		}
	})

	t.Run("les candidats sont les membres de ses groupes", func(t *testing.T) {
		w := callAs(h.GetUserGroupCandidates, f.holder, http.MethodGet, 0, "/?parent_id="+strconv.Itoa(int(f.root.ID)), nil)
		var users []struct {
			ID uint `json:"id"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &users)
		ids := []uint{}
		for _, u := range users {
			ids = append(ids, u.ID)
		}
		if !hasID(ids, f.member.id) || !hasID(ids, f.holder.id) || hasID(ids, f.outsider.id) {
			t.Fatalf("candidats = %v", ids)
		}
		if w := callAs(h.GetUserGroupCandidates, f.holder, http.MethodGet, 0, "/?parent_id="+strconv.Itoa(int(f.other.ID)), nil); w.Code != http.StatusForbidden {
			t.Fatalf("candidats d'un groupe extérieur : statut %d", w.Code)
		}
	})

	t.Run("supprimer un sous-groupe fait remonter les siens", func(t *testing.T) {
		if w := callAs(h.DeleteUserGroup, f.holder, http.MethodDelete, f.sub.ID, "", nil); w.Code != http.StatusOK {
			t.Fatalf("statut %d %s", w.Code, w.Body.String())
		}
		var g models.Group
		tx.First(&g, created.ID)
		if g.ParentID == nil || *g.ParentID != f.root.ID {
			t.Fatalf("parent après suppression = %v, attendu %d", g.ParentID, f.root.ID)
		}
	})

	t.Run("sans la permission, rien n'est permis", func(t *testing.T) {
		if w := callAs(h.CreateUserGroup, f.member, http.MethodPost, 0, "", map[string]interface{}{"name": tag + "-x4", "parent_id": f.root.ID}); w.Code != http.StatusForbidden {
			t.Fatalf("statut %d", w.Code)
		}
	})
}

func TestUserGroupsPermissionEverywhere(t *testing.T) {
	tx := rollbackTx(t)
	f := newUserGroupsFixture(t, tx)
	h := NewGroupAdminHandler(tx)
	tag := hierarchyTag()

	w := callAs(h.CreateUserGroup, f.allHolder, http.MethodPost, 0, "", map[string]interface{}{"name": tag + "-principal", "parent_id": nil, "user_ids": []uint{f.outsider.id}})
	if w.Code != http.StatusCreated {
		t.Fatalf("groupe principal avec n'importe quel compte : statut %d %s", w.Code, w.Body.String())
	}
	if w := callAs(h.UpdateUserGroup, f.allHolder, http.MethodPut, f.root.ID, "", map[string]interface{}{"description": "x"}); w.Code != http.StatusOK {
		t.Fatalf("modification d'un groupe quelconque : statut %d", w.Code)
	}
	if w := callAs(h.UpdateUserGroup, f.allHolder, http.MethodPut, f.linked.ID, "", map[string]interface{}{"description": "x"}); w.Code != http.StatusForbidden {
		t.Fatalf("groupe lié au trombinoscope : statut %d, attendu 403", w.Code)
	}
	if everyone, err := models.EveryoneGroupID(tx); err == nil && everyone != 0 {
		if w := callAs(h.DeleteUserGroup, f.allHolder, http.MethodDelete, everyone, "", nil); w.Code != http.StatusForbidden {
			t.Fatalf("groupe par défaut : statut %d, attendu 403", w.Code)
		}
	}
}
