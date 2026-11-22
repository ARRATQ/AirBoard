//go:build integration

package handlers

// Actualités et événements gérés par permission (news.manage, events.manage), à travers
// les vrais middlewares montés comme dans routes.go. Transaction annulée à la fin.
//
//	AIRBOARD_TEST_DSN="host=localhost user=airboard password=... dbname=<copie migrée> sslmode=disable" \
//	  go test -tags integration -run TestNewsEventsByPermission -v ./handlers/

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"airboard/config"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// contentRouter monte les routes Actualités / Événements comme routes.go, derrière une
// authentification simulée qui pose ce que RequireAuth lirait en base pour cet utilisateur.
func contentRouter(t *testing.T, tx *gorm.DB, who identity) *gin.Engine {
	t.Helper()
	perms, err := middleware.LoadPermissions(tx, who.id)
	if err != nil {
		t.Fatalf("permissions : %v", err)
	}
	var members []uint
	if err := tx.Table("user_groups").Where("user_id = ?", who.id).Pluck("group_id", &members).Error; err != nil {
		t.Fatalf("groupes : %v", err)
	}
	am := middleware.NewAuthMiddleware(&config.Config{}, tx)
	gs := services.NewGamificationService(tx)
	news := NewNewsHandler(tx, &config.Config{}, gs)
	events := NewEventsHandler(tx, gs)
	scoped := NewGroupAdminHandler(tx)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", who.id)
		c.Set("role", who.role)
		c.Set("managed_group_ids", who.managed)
		c.Set(middleware.CtxPermissions, perms)
		c.Set(middleware.CtxMemberGroupIDs, members)
	})
	// Préfixe réel : GetNews / GetEvents reconnaissent l'interface /group-admin à leur chemin
	api := r.Group("/api/v1")
	api.GET("/group-admin/managed-groups", am.RequirePermissionParam(), scoped.GetManagedGroups)

	newsAll := api.Group("/editor", am.RequirePermissionEverywhere(models.PermNewsManage))
	newsAll.POST("/news", news.CreateNews)
	newsAll.PUT("/news/:id", news.UpdateNews)
	newsScoped := api.Group("/group-admin", am.RequirePermission(models.PermNewsManage))
	newsScoped.GET("/news", news.GetNews)
	newsScoped.POST("/news", news.CreateNews)
	newsScoped.PUT("/news/:id", news.UpdateNews)
	newsScoped.DELETE("/news/:id", news.DeleteNews)

	eventsAll := api.Group("/editor", am.RequirePermissionEverywhere(models.PermEventsManage))
	eventsAll.DELETE("/events/:id", events.DeleteEvent)
	eventsScoped := api.Group("/group-admin", am.RequirePermission(models.PermEventsManage))
	eventsScoped.POST("/events", events.CreateEventGroupAdmin)
	return r
}

func TestNewsEventsByPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	f := newFixture(t, tx) // f.member est membre de g1 ; f.groupAdmin administre g2 (ancienne table)
	g1, g2 := []models.Group{f.g1}, []models.Group{f.g2}

	grantPermissions(t, tx, f.tag+"-redaction-groupes", map[string]string{
		models.PermNewsManage: models.ScopeGroups, models.PermEventsManage: models.ScopeGroups,
	}, f.member.id)
	grantPermissions(t, tx, f.tag+"-redaction-partout", map[string]string{
		models.PermNewsManage: models.ScopeAll, models.PermEventsManage: models.ScopeAll,
	}, f.outsider.id)

	news := f.createNews(t, tx, map[string]content{
		"global":   {published: true, author: f.adm},
		"g1-draft": {published: false, targets: g1, author: f.adm},
		"g2":       {published: true, targets: g2, author: f.adm},
	})
	events := f.createEvents(t, tx, map[string]content{
		"ev-global": {published: true, author: f.adm},
	})

	inGroups := contentRouter(t, tx, f.member)
	everywhere := contentRouter(t, tx, f.outsider)
	legacy := contentRouter(t, tx, f.groupAdmin)
	noRole := contentRouter(t, tx, identity{id: f.editor.id, role: "editor"}) // ancien rôle, sans permission

	draft := func(title string, groups ...uint) map[string]interface{} {
		return map[string]interface{}{"title": f.title(title), "content": "x", "is_published": false, "target_group_ids": groups}
	}
	status := func(r *gin.Engine, method, path string, body interface{}) int {
		code, _, _ := send(t, r, method, "/api/v1"+path, body)
		return code
	}

	t.Run("ses groupes : article dans son groupe, pas ailleurs", func(t *testing.T) {
		if code := status(inGroups, http.MethodPost, "/group-admin/news", draft("m-g1", f.g1.ID)); code != http.StatusCreated {
			t.Errorf("article ciblant g1 : %d, attendu 201", code)
		}
		if code := status(inGroups, http.MethodPost, "/group-admin/news", draft("m-g2", f.g2.ID)); code != http.StatusForbidden {
			t.Errorf("article ciblant g2 : %d, attendu 403", code)
		}
		if code := status(inGroups, http.MethodPost, "/editor/news", draft("m-partout")); code != http.StatusForbidden {
			t.Errorf("route de tout le site : %d, attendu 403", code)
		}
	})

	t.Run("ses groupes : liste et modification des articles de ses groupes", func(t *testing.T) {
		_, obj, _ := send(t, inGroups, http.MethodGet, "/api/v1/group-admin/news?page_size=100", nil)
		expectTitles(t, "group-admin/news", obj, f.tag, "g1-draft", "m-g1")
		edit := draft("g1-draft-modifié", f.g1.ID)
		if code := status(inGroups, http.MethodPut, "/group-admin/news/"+news["g1-draft"].Slug, edit); code != http.StatusOK {
			t.Errorf("brouillon de g1 rédigé par l'admin : %d, attendu 200", code)
		}
		if code := status(inGroups, http.MethodPut, "/group-admin/news/"+news["g2"].Slug, draft("g2-x", f.g2.ID)); code != http.StatusForbidden {
			t.Errorf("article de g2 : %d, attendu 403", code)
		}
		if code := status(inGroups, http.MethodPut, "/group-admin/news/"+news["global"].Slug, draft("global-x")); code != http.StatusForbidden {
			t.Errorf("article global : %d, attendu 403", code)
		}
	})

	t.Run("groupes gérés selon la permission demandée", func(t *testing.T) {
		code, _, groups := send(t, inGroups, http.MethodGet, "/api/v1/group-admin/managed-groups?permission=news.manage", nil)
		if code != http.StatusOK || len(groups) != 1 {
			t.Errorf("actualités : %d %v, attendu seulement g1", code, groups)
		}
		if code := status(inGroups, http.MethodGet, "/group-admin/managed-groups", nil); code != http.StatusForbidden {
			t.Errorf("ressources (sans permission) : %d, attendu 403", code)
		}
		if code := status(inGroups, http.MethodGet, "/group-admin/managed-groups?permission=inconnue", nil); code != http.StatusBadRequest {
			t.Errorf("permission inconnue : %d, attendu 400", code)
		}
	})

	t.Run("ses groupes : événements", func(t *testing.T) {
		ev := func(groups ...uint) map[string]interface{} {
			return map[string]interface{}{"title": f.title("ev-m"), "start_date": time.Now().Add(72 * time.Hour), "is_published": false, "target_group_ids": groups}
		}
		if code := status(inGroups, http.MethodPost, "/group-admin/events", ev(f.g1.ID)); code != http.StatusCreated {
			t.Errorf("événement ciblant g1 : %d, attendu 201", code)
		}
		if code := status(inGroups, http.MethodPost, "/group-admin/events", ev(f.g2.ID)); code != http.StatusForbidden {
			t.Errorf("événement ciblant g2 : %d, attendu 403", code)
		}
	})

	t.Run("tout le site : modifie l'article et supprime l'événement d'un autre", func(t *testing.T) {
		if code := status(everywhere, http.MethodPut, "/editor/news/"+news["global"].Slug, draft("global-modifié")); code != http.StatusOK {
			t.Errorf("article global de l'admin : %d, attendu 200", code)
		}
		if code := status(everywhere, http.MethodDelete, fmt.Sprintf("/editor/events/%d", events["ev-global"].ID), nil); code != http.StatusOK {
			t.Errorf("événement de l'admin : %d, attendu 200", code)
		}
	})

	t.Run("ancien admin de groupe : inchangé", func(t *testing.T) {
		if code := status(legacy, http.MethodPost, "/group-admin/news", draft("ga-g2", f.g2.ID)); code != http.StatusCreated {
			t.Errorf("article ciblant g2 : %d, attendu 201", code)
		}
	})

	t.Run("ancien rôle éditeur sans permission : refusé", func(t *testing.T) {
		if code := status(noRole, http.MethodPost, "/editor/news", draft("e-partout")); code != http.StatusForbidden {
			t.Errorf("route de tout le site : %d, attendu 403", code)
		}
		if code := status(noRole, http.MethodGet, "/group-admin/news", nil); code != http.StatusForbidden {
			t.Errorf("route de ses groupes : %d, attendu 403", code)
		}
	})
}
