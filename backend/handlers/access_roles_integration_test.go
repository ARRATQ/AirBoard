//go:build integration

package handlers

// Ressources gérées par permission (resources.manage), de la base au handler, à travers
// le vrai middleware RequirePermission. Transaction annulée à la fin.
//
//	AIRBOARD_TEST_DSN="host=localhost user=airboard password=... dbname=<copie migrée> sslmode=disable" \
//	  go test -tags integration -run TestResourcesByPermission -v ./handlers/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"airboard/config"
	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// grantRole crée un rôle « gérer les ressources » et l'attribue aux utilisateurs.
func grantRole(t *testing.T, tx *gorm.DB, name, scope string, userIDs ...uint) {
	t.Helper()
	grantPermissions(t, tx, name, map[string]string{models.PermResourcesManage: scope}, userIDs...)
}

// grantPermissions crée un rôle avec ces permissions (permission -> portée) et l'attribue.
func grantPermissions(t *testing.T, tx *gorm.DB, name string, perms map[string]string, userIDs ...uint) {
	t.Helper()
	role := models.AccessRole{Name: name}
	if err := tx.Omit("Permissions").Create(&role).Error; err != nil {
		t.Fatalf("rôle %s : %v", name, err)
	}
	for perm, scope := range perms {
		mustCreate(t, tx, &models.AccessRolePermission{RoleID: role.ID, Permission: perm, Scope: scope})
	}
	for _, id := range userIDs {
		if err := tx.Exec("INSERT INTO user_access_roles (user_id, access_role_id) VALUES (?, ?)", id, role.ID).Error; err != nil {
			t.Fatalf("attribution : %v", err)
		}
	}
}

// resourcesRouter monte les routes Ressources comme routes.go, derrière une authentification
// simulée qui pose ce que RequireAuth lirait en base pour cet utilisateur.
func resourcesRouter(t *testing.T, tx *gorm.DB, userID uint, role string, managed []uint) *gin.Engine {
	t.Helper()
	perms, err := middleware.LoadPermissions(tx, userID)
	if err != nil {
		t.Fatalf("permissions : %v", err)
	}
	members, err := models.PermissionGroups(tx, userID)
	if err != nil {
		t.Fatalf("groupes : %v", err)
	}
	am := middleware.NewAuthMiddleware(&config.Config{}, tx)
	admin := NewAdminHandler(tx, &config.Config{}, nil)
	scoped := NewGroupAdminHandler(tx)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role", role)
		c.Set("managed_group_ids", managed)
		c.Set(middleware.CtxPermissions, perms)
		c.Set(middleware.CtxPermissionGroups, members)
	})
	everywhere := r.Group("/admin", am.RequirePermissionEverywhere(models.PermResourcesManage))
	everywhere.GET("/app-groups", admin.GetAppGroups)
	everywhere.POST("/app-groups", admin.CreateAppGroup)
	inGroups := r.Group("/group-admin", am.RequirePermission(models.PermResourcesManage))
	inGroups.GET("/app-groups", scoped.GetAppGroups)
	inGroups.POST("/app-groups", admin.CreateAppGroup)
	inGroups.POST("/applications", scoped.CreateApplication)
	inGroups.GET("/managed-groups", scoped.GetManagedGroups)
	return r
}

func send(t *testing.T, r *gin.Engine, method, path string, body interface{}) (int, map[string]interface{}, []interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var obj map[string]interface{}
	var list []interface{}
	if json.Unmarshal(w.Body.Bytes(), &obj) != nil {
		_ = json.Unmarshal(w.Body.Bytes(), &list)
	}
	return w.Code, obj, list
}

func TestResourcesByPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	f := newFixture(t, tx) // f.editor est membre de g1 ; f.groupAdmin est éditeur de g2 (ancien admin de groupe converti)

	grantRole(t, tx, f.tag+"-éditeurs", models.ScopeGroups, f.editor.id)
	grantRole(t, tx, f.tag+"-partout", models.ScopeAll, f.member.id)

	editor := resourcesRouter(t, tx, f.editor.id, "editor", nil)
	everywhere := resourcesRouter(t, tx, f.member.id, "user", nil)
	outsider := resourcesRouter(t, tx, f.outsider.id, "user", nil)
	legacy := resourcesRouter(t, tx, f.groupAdmin.id, "user", f.groupAdmin.managed)

	t.Run("éditeur : catégorie dans son groupe", func(t *testing.T) {
		code, obj, _ := send(t, editor, http.MethodPost, "/group-admin/app-groups", map[string]interface{}{"name": f.tag + "-cat-g1", "owner_group_id": f.g1.ID})
		if code != http.StatusCreated && code != http.StatusOK {
			t.Fatalf("création : %d %v", code, obj)
		}
		catID := uint(obj["id"].(float64))
		if obj["is_private"] != true {
			t.Errorf("la catégorie d'un groupe doit être privée : %v", obj["is_private"])
		}
		code, obj, _ = send(t, editor, http.MethodPost, "/group-admin/applications", map[string]interface{}{"name": f.tag + "-app", "url": "https://example.test", "app_group_id": catID})
		if code != http.StatusCreated && code != http.StatusOK {
			t.Fatalf("ressource : %d %v", code, obj)
		}
		_, obj, _ = send(t, editor, http.MethodGet, "/group-admin/app-groups?page_size=1000", nil)
		if got := fmt.Sprint(namesWithPrefix(obj["data"], f.tag)); got != "[cat-g1]" {
			t.Errorf("catégories visibles : %s", got)
		}
		_, _, groups := send(t, editor, http.MethodGet, "/group-admin/managed-groups", nil)
		if len(groups) != 1 {
			t.Errorf("groupes gérés : %v, attendu seulement g1", groups)
		}
	})

	t.Run("éditeur : pas dans un groupe dont il n'est pas membre", func(t *testing.T) {
		code, _, _ := send(t, editor, http.MethodPost, "/group-admin/app-groups", map[string]interface{}{"name": f.tag + "-cat-g2", "owner_group_id": f.g2.ID})
		if code != http.StatusForbidden {
			t.Errorf("statut %d, attendu 403", code)
		}
	})

	t.Run("éditeur : pas les routes de tout le site", func(t *testing.T) {
		if code, _, _ := send(t, editor, http.MethodGet, "/admin/app-groups", nil); code != http.StatusForbidden {
			t.Errorf("statut %d, attendu 403", code)
		}
	})

	t.Run("permission sur tout le site : catégorie publique", func(t *testing.T) {
		code, obj, _ := send(t, everywhere, http.MethodPost, "/admin/app-groups", map[string]interface{}{"name": f.tag + "-cat-publique"})
		if code != http.StatusCreated && code != http.StatusOK {
			t.Fatalf("création : %d %v", code, obj)
		}
		if obj["is_private"] == true {
			t.Error("une catégorie créée sans restriction de groupe ne doit pas être forcée en privé")
		}
		if code, _, _ := send(t, everywhere, http.MethodGet, "/admin/app-groups", nil); code != http.StatusOK {
			t.Errorf("liste : %d", code)
		}
	})

	t.Run("sans rôle : refusé", func(t *testing.T) {
		if code, _, _ := send(t, outsider, http.MethodGet, "/group-admin/app-groups", nil); code != http.StatusForbidden {
			t.Errorf("statut %d, attendu 403", code)
		}
	})

	t.Run("éditeur de groupe (ancien admin converti) : son groupe", func(t *testing.T) {
		code, obj, _ := send(t, legacy, http.MethodPost, "/group-admin/app-groups", map[string]interface{}{"name": f.tag + "-cat-legacy", "owner_group_id": f.g2.ID})
		if code != http.StatusCreated && code != http.StatusOK {
			t.Fatalf("création : %d %v", code, obj)
		}
	})
}

// namesWithPrefix renvoie les noms (sans le préfixe du fixture) des éléments d'une liste.
func namesWithPrefix(list interface{}, prefix string) []string {
	var out []string
	items, _ := list.([]interface{})
	for _, it := range items {
		m, _ := it.(map[string]interface{})
		if name, ok := m["name"].(string); ok && len(name) > len(prefix) && name[:len(prefix)] == prefix {
			out = append(out, name[len(prefix)+1:])
		}
	}
	return out
}
