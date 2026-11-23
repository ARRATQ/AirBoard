//go:build integration

package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"airboard/config"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Groupes hiérarchiques (migration 00010) et groupes liés au trombinoscope.

func hierarchyTag() string { return fmt.Sprintf("gh-%d", time.Now().UnixNano()) }

// groupTree crée racine > enfant > petit-enfant, plus un groupe isolé.
func groupTree(t *testing.T, tx *gorm.DB, tag string) (root, child, grand, other models.Group) {
	t.Helper()
	root = models.Group{Name: tag + "-racine"}
	mustCreate(t, tx, &root)
	child = models.Group{Name: tag + "-enfant", ParentID: &root.ID}
	mustCreate(t, tx, &child)
	grand = models.Group{Name: tag + "-petit", ParentID: &child.ID}
	mustCreate(t, tx, &grand)
	other = models.Group{Name: tag + "-autre"}
	mustCreate(t, tx, &other)
	return
}

func hasID(ids []uint, id uint) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

func TestGroupHierarchyWalks(t *testing.T) {
	tx := rollbackTx(t)
	root, child, grand, other := groupTree(t, tx, hierarchyTag())

	down, err := models.GroupsWithDescendants(tx, []uint{child.ID})
	if err != nil || len(down) != 2 || !hasID(down, child.ID) || !hasID(down, grand.ID) {
		t.Fatalf("descendants de l'enfant = %v (%v), attendu enfant + petit-enfant", down, err)
	}
	up, err := models.GroupsWithAncestors(tx, []uint{grand.ID})
	if err != nil || len(up) != 3 || !hasID(up, root.ID) || hasID(up, other.ID) {
		t.Fatalf("ancêtres du petit-enfant = %v (%v), attendu les trois niveaux", up, err)
	}
	for _, c := range []struct {
		group, parent uint
		cycle         bool
	}{{root.ID, root.ID, true}, {root.ID, grand.ID, true}, {grand.ID, root.ID, false}, {other.ID, child.ID, false}} {
		got, err := models.WouldCreateCycle(tx, c.group, c.parent)
		if err != nil || got != c.cycle {
			t.Errorf("WouldCreateCycle(%d, %d) = %v (%v), attendu %v", c.group, c.parent, got, err, c.cycle)
		}
	}
}

func TestGroupHierarchyVisibilityAndScope(t *testing.T) {
	tx := rollbackTx(t)
	tag := hierarchyTag()
	root, child, grand, _ := groupTree(t, tx, tag)
	member := models.User{Username: tag + "m", Email: tag + "m@example.test", Role: "user", Groups: []models.Group{grand}}
	mustCreate(t, tx, &member)
	manager := models.User{Username: tag + "g", Email: tag + "g@example.test", Role: "user", Groups: []models.Group{child}}
	mustCreate(t, tx, &manager)
	grantPermissions(t, tx, tag+"-role", map[string]string{models.PermNewsManage: models.ScopeGroups}, manager.ID)

	t.Run("un membre d'un sous-groupe voit ce qui cible les groupes au-dessus", func(t *testing.T) {
		got := userGroupIDs(tx, member.ID)
		if !hasID(got, grand.ID) || !hasID(got, child.ID) || !hasID(got, root.ID) {
			t.Fatalf("groupes effectifs = %v, attendu petit-enfant, enfant et racine", got)
		}
	})
	t.Run("un membre d'un groupe parent ne voit pas ce qui cible ses sous-groupes", func(t *testing.T) {
		got := userGroupIDs(tx, manager.ID)
		if hasID(got, grand.ID) || !hasID(got, root.ID) {
			t.Fatalf("groupes effectifs du parent = %v", got)
		}
	})
	t.Run("les membres d'un groupe comprennent ceux des sous-groupes", func(t *testing.T) {
		got := groupMemberIDs(tx, []uint{root.ID})
		if !hasID(got, member.ID) || !hasID(got, manager.ID) {
			t.Fatalf("membres de la racine = %v", got)
		}
		if got := groupMemberIDs(tx, []uint{grand.ID}); hasID(got, manager.ID) {
			t.Fatalf("le membre du parent compte dans le sous-groupe : %v", got)
		}
	})
	t.Run("un rôle limité à ses groupes couvre leurs sous-groupes", func(t *testing.T) {
		groups, err := models.PermissionGroups(tx, manager.ID)
		if err != nil {
			t.Fatal(err)
		}
		got := groups[models.PermNewsManage]
		if !hasID(got, child.ID) || !hasID(got, grand.ID) || hasID(got, root.ID) {
			t.Fatalf("groupes couverts = %v, attendu enfant + petit-enfant", got)
		}
		ids, err := models.PermissionGroupIDs(tx, manager.ID)
		if err != nil || !hasID(ids, grand.ID) || hasID(ids, root.ID) {
			t.Fatalf("PermissionGroupIDs = %v (%v)", ids, err)
		}
	})
}

// callAdminJSON exécute un handler avec un corps JSON et un paramètre :id.
func callAdminJSON(h gin.HandlerFunc, method string, id uint, body interface{}) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	raw, _ := json.Marshal(body)
	c.Request = httptest.NewRequest(method, "/", bytes.NewReader(raw))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(id), 10)}}
	c.Set("user_id", uint(1))
	c.Set("role", "admin")
	h(c)
	return w
}

func TestUpdateGroupParent(t *testing.T) {
	tx := rollbackTx(t)
	root, child, grand, other := groupTree(t, tx, hierarchyTag())
	h := NewAdminHandler(tx, &config.Config{}, nil)

	if w := callAdminJSON(h.UpdateGroup, http.MethodPut, root.ID, map[string]interface{}{"parent_id": grand.ID}); w.Code != http.StatusBadRequest {
		t.Fatalf("boucle acceptée : %d %s", w.Code, w.Body.String())
	}
	if w := callAdminJSON(h.UpdateGroup, http.MethodPut, other.ID, map[string]interface{}{"parent_id": child.ID}); w.Code != http.StatusOK {
		t.Fatalf("rattachement refusé : %d %s", w.Code, w.Body.String())
	}
	var got models.Group
	tx.First(&got, other.ID)
	if got.ParentID == nil || *got.ParentID != child.ID {
		t.Fatalf("parent = %v, attendu %d", got.ParentID, child.ID)
	}
	// Corps sans parent_id : parent inchangé ; null : groupe racine
	callAdminJSON(h.UpdateGroup, http.MethodPut, other.ID, map[string]interface{}{"description": "x"})
	got = models.Group{}
	tx.First(&got, other.ID)
	if got.ParentID == nil {
		t.Fatal("parent effacé alors que parent_id était absent")
	}
	callAdminJSON(h.UpdateGroup, http.MethodPut, other.ID, map[string]interface{}{"parent_id": nil})
	got = models.Group{}
	tx.First(&got, other.ID)
	if got.ParentID != nil {
		t.Fatalf("parent = %v, attendu aucun", *got.ParentID)
	}

	// Suppression : les sous-groupes remontent d'un niveau
	if w := callAdminJSON(h.DeleteGroup, http.MethodDelete, child.ID, nil); w.Code != http.StatusOK {
		t.Fatalf("suppression : %d %s", w.Code, w.Body.String())
	}
	got = models.Group{}
	tx.First(&got, grand.ID)
	if got.ParentID == nil || *got.ParentID != root.ID {
		t.Fatalf("parent du petit-enfant après suppression = %v, attendu %d", got.ParentID, root.ID)
	}
}

// directoryFixture : entité mère > fille, un poste dans chacune.
type directoryFixture struct {
	parent, sub       models.OrgUnit
	parentPos, subPos models.OrgPosition
}

func newDirectoryFixture(t *testing.T, tx *gorm.DB, tag string) directoryFixture {
	t.Helper()
	var f directoryFixture
	f.parent = models.OrgUnit{Name: tag + " Direction", Code: "DIR"}
	mustCreate(t, tx, &f.parent)
	f.sub = models.OrgUnit{Name: tag + " Service", ParentID: &f.parent.ID}
	mustCreate(t, tx, &f.sub)
	f.parentPos = models.OrgPosition{UnitID: f.parent.ID, Title: "Directeur"}
	mustCreate(t, tx, &f.parentPos)
	f.subPos = models.OrgPosition{UnitID: f.sub.ID, Title: "Agent"}
	mustCreate(t, tx, &f.subPos)
	return f
}

func membershipSource(t *testing.T, tx *gorm.DB, userID, groupID uint) string {
	t.Helper()
	var sources []string
	tx.Table("user_groups").Where("user_id = ? AND group_id = ?", userID, groupID).Pluck("source", &sources)
	if len(sources) == 0 {
		return ""
	}
	return sources[0]
}

func TestCreateUnitGroupsAndSync(t *testing.T) {
	tx := rollbackTx(t)
	tag := hierarchyTag()
	f := newDirectoryFixture(t, tx, tag)
	h := NewDirectoryHandler(tx)

	// Un groupe libre porte déjà le nom de la sous-entité : il est lié, pas recréé
	existing := models.Group{Name: f.sub.Name}
	mustCreate(t, tx, &existing)

	// Compte au même e-mail (casse différente) qu'une fiche sans compte
	alice := models.User{Username: tag + "alice", Email: tag + "Alice@example.test", Role: "user"}
	mustCreate(t, tx, &alice)
	person := models.DirectoryPerson{FirstName: "Alice", LastName: tag, Email: tag + "alice@EXAMPLE.test", IsActive: true}
	mustCreate(t, tx, &person)
	mustCreate(t, tx, &models.PositionAssignment{PositionID: f.subPos.ID, PersonID: person.ID, IsInterim: true})

	w := callAdminJSON(h.CreateUnitGroups, http.MethodPost, f.parent.ID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("création des groupes : %d %s", w.Code, w.Body.String())
	}
	var res unitGroupsResult
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res.Created != 1 || res.Linked != 1 {
		t.Fatalf("résultat = %+v, attendu 1 créé et 1 lié", res)
	}

	var parent, sub models.OrgUnit
	tx.First(&parent, f.parent.ID)
	tx.First(&sub, f.sub.ID)
	if parent.GroupID == nil || sub.GroupID == nil || *sub.GroupID != existing.ID {
		t.Fatalf("liens : mère %v, fille %v (attendu %d)", parent.GroupID, sub.GroupID, existing.ID)
	}
	var subGroup models.Group
	tx.First(&subGroup, existing.ID)
	if subGroup.ParentID == nil || *subGroup.ParentID != *parent.GroupID {
		t.Fatalf("hiérarchie : parent du groupe de la fille = %v, attendu %d", subGroup.ParentID, *parent.GroupID)
	}

	t.Run("la fiche est liée au compte par l'e-mail et le compte rejoint le groupe de son entité", func(t *testing.T) {
		var p models.DirectoryPerson
		tx.First(&p, person.ID)
		if p.UserID == nil || *p.UserID != alice.ID {
			t.Fatalf("fiche liée à %v, attendu %d", p.UserID, alice.ID)
		}
		if s := membershipSource(t, tx, alice.ID, existing.ID); s != "directory" {
			t.Fatalf("appartenance au groupe de la fille : %q", s)
		}
		if s := membershipSource(t, tx, alice.ID, *parent.GroupID); s != "" {
			t.Fatalf("membre direct du groupe de la mère (%q) : la hiérarchie suffit", s)
		}
		if !hasID(userGroupIDs(tx, alice.ID), *parent.GroupID) {
			t.Fatal("le groupe de la mère n'est pas acquis par la hiérarchie")
		}
	})

	t.Run("un changement d'entité déplace l'appartenance ; les ajouts manuels restent", func(t *testing.T) {
		manual := models.Group{Name: tag + "-manuel"}
		mustCreate(t, tx, &manual)
		tx.Exec("INSERT INTO user_groups (user_id, group_id) VALUES (?, ?)", alice.ID, manual.ID)
		tx.Model(&models.PositionAssignment{}).Where("person_id = ?", person.ID).Update("position_id", f.parentPos.ID)

		r, err := services.SyncDirectoryGroups(tx)
		if err != nil {
			t.Fatal(err)
		}
		if r.Added != 1 || r.Removed != 1 {
			t.Fatalf("synchronisation = %+v, attendu +1 / -1", r)
		}
		if s := membershipSource(t, tx, alice.ID, existing.ID); s != "" {
			t.Fatalf("toujours membre de l'ancienne entité (%q)", s)
		}
		if s := membershipSource(t, tx, alice.ID, *parent.GroupID); s != "directory" {
			t.Fatalf("pas membre de la nouvelle entité (%q)", s)
		}
		if s := membershipSource(t, tx, alice.ID, manual.ID); s != "manual" {
			t.Fatalf("ajout manuel perdu (%q)", s)
		}
		again, _ := services.SyncDirectoryGroups(tx)
		if again.Added+again.Removed+again.LinkedPeople != 0 {
			t.Fatalf("relance non idempotente : %+v", again)
		}
	})

	t.Run("une fiche désactivée quitte le groupe", func(t *testing.T) {
		tx.Model(&models.DirectoryPerson{}).Where("id = ?", person.ID).Update("is_active", false)
		if _, err := services.SyncDirectoryGroups(tx); err != nil {
			t.Fatal(err)
		}
		if s := membershipSource(t, tx, alice.ID, *parent.GroupID); s != "" {
			t.Fatalf("fiche inactive toujours membre (%q)", s)
		}
	})

	t.Run("lier un groupe déjà lié à une autre entité est refusé", func(t *testing.T) {
		w := callAdminJSON(h.LinkUnitGroup, http.MethodPut, f.parent.ID, map[string]interface{}{"group_id": existing.ID})
		if w.Code != http.StatusConflict {
			t.Fatalf("statut %d %s, attendu 409", w.Code, w.Body.String())
		}
		w = callAdminJSON(h.LinkUnitGroup, http.MethodPut, f.sub.ID, map[string]interface{}{"group_id": nil})
		if w.Code != http.StatusOK {
			t.Fatalf("retrait du lien : %d %s", w.Code, w.Body.String())
		}
		sub = models.OrgUnit{}
		tx.First(&sub, f.sub.ID)
		if sub.GroupID != nil {
			t.Fatalf("lien toujours présent : %d", *sub.GroupID)
		}
	})
}
