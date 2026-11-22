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
	members, err := models.PermissionGroupIDs(tx, who.id)
	if err != nil {
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
	f := newFixture(t, tx) // f.member est membre de g1 ; f.groupAdmin est éditeur de g2 (ancien admin de groupe converti)
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

	t.Run("éditeur de groupe (ancien admin converti) : son groupe", func(t *testing.T) {
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

// pollsRouter monte les routes Sondages, Modération et rôles d'un compte comme routes.go.
func pollsRouter(t *testing.T, tx *gorm.DB, who identity) *gin.Engine {
	t.Helper()
	perms, err := middleware.LoadPermissions(tx, who.id)
	if err != nil {
		t.Fatalf("permissions : %v", err)
	}
	members, err := models.PermissionGroupIDs(tx, who.id)
	if err != nil {
		t.Fatalf("groupes : %v", err)
	}
	am := middleware.NewAuthMiddleware(&config.Config{}, tx)
	gs := services.NewGamificationService(tx)
	polls := NewPollsHandler(tx, gs)
	comments := NewCommentHandler(tx, gs)
	roles := NewAccessRolesHandler(tx, nil)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", who.id)
		c.Set("role", who.role)
		c.Set("managed_group_ids", who.managed)
		c.Set(middleware.CtxPermissions, perms)
		c.Set(middleware.CtxMemberGroupIDs, members)
	})
	api := r.Group("/api/v1")
	pollsAll := api.Group("/editor", am.RequirePermissionEverywhere(models.PermPollsManage))
	pollsAll.PUT("/polls/:id", polls.UpdatePoll)
	pollsScoped := api.Group("/group-admin", am.RequirePermission(models.PermPollsManage))
	pollsScoped.POST("/polls", polls.CreatePoll)
	pollsScoped.PUT("/polls/:id", polls.UpdatePoll)
	pollsScoped.POST("/polls/:id/close", polls.ClosePoll)
	moderation := api.Group("/editor", am.RequirePermissionEverywhere(models.PermCommentsModerate))
	moderation.GET("/comments/pending", comments.GetPendingComments)
	admin := api.Group("/admin", am.RequireAdmin())
	admin.PUT("/users/:id/access-roles", roles.SetUserRoles)
	admin.GET("/users/:id/access-roles", roles.GetUserRoles)
	return r
}

func TestPollsModerationByPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	f := newFixture(t, tx) // f.member est membre de g1
	g1, g2 := []models.Group{f.g1}, []models.Group{f.g2}

	grantPermissions(t, tx, f.tag+"-sondages-groupes", map[string]string{models.PermPollsManage: models.ScopeGroups}, f.member.id)
	grantPermissions(t, tx, f.tag+"-moderation", map[string]string{
		models.PermPollsManage: models.ScopeAll, models.PermCommentsModerate: models.ScopeAll,
	}, f.outsider.id)

	polls := f.createPolls(t, tx, map[string]content{
		"poll-global": {published: true, author: f.adm},
		"poll-g1":     {published: true, targets: g1, author: f.adm},
		"poll-g2":     {published: true, targets: g2, author: f.adm},
	})

	inGroups := pollsRouter(t, tx, f.member)
	everywhere := pollsRouter(t, tx, f.outsider)
	admin := pollsRouter(t, tx, f.adm)
	status := func(r *gin.Engine, method, path string, body interface{}) int {
		code, _, _ := send(t, r, method, "/api/v1"+path, body)
		return code
	}
	poll := func(title string, groups ...uint) map[string]interface{} {
		return map[string]interface{}{
			"title": f.title(title), "poll_type": "single", "show_results": "always", "is_active": true,
			"options":          []map[string]interface{}{{"text": "oui"}, {"text": "non"}},
			"target_group_ids": groups,
		}
	}

	t.Run("ses groupes : sondage dans son groupe, pas ailleurs", func(t *testing.T) {
		if code := status(inGroups, http.MethodPost, "/group-admin/polls", poll("m-g2", f.g2.ID)); code != http.StatusForbidden {
			t.Errorf("sondage ciblant g2 : %d, attendu 403", code)
		}
		if code := status(inGroups, http.MethodPost, fmt.Sprintf("/group-admin/polls/%d/close", polls["poll-g1"].ID), nil); code != http.StatusOK {
			t.Errorf("clore le sondage de g1 : %d, attendu 200", code)
		}
		if code := status(inGroups, http.MethodPost, fmt.Sprintf("/group-admin/polls/%d/close", polls["poll-g2"].ID), nil); code != http.StatusForbidden {
			t.Errorf("clore le sondage de g2 : %d, attendu 403", code)
		}
		if code := status(inGroups, http.MethodPost, fmt.Sprintf("/group-admin/polls/%d/close", polls["poll-global"].ID), nil); code != http.StatusForbidden {
			t.Errorf("clore le sondage global : %d, attendu 403", code)
		}
	})

	t.Run("modération : permission requise", func(t *testing.T) {
		if code := status(everywhere, http.MethodGet, "/editor/comments/pending", nil); code != http.StatusOK {
			t.Errorf("modérateur : %d, attendu 200", code)
		}
		if code := status(inGroups, http.MethodGet, "/editor/comments/pending", nil); code != http.StatusForbidden {
			t.Errorf("sans permission : %d, attendu 403", code)
		}
	})

	t.Run("rôles d'un compte : remplacés par l'admin", func(t *testing.T) {
		var roleID uint
		if err := tx.Table("access_roles").Where("name = ?", f.tag+"-moderation").Pluck("id", &roleID).Error; err != nil {
			t.Fatalf("rôle : %v", err)
		}
		path := fmt.Sprintf("/admin/users/%d/access-roles", f.member.id)
		if code := status(admin, http.MethodPut, path, map[string]interface{}{"role_ids": []uint{roleID}}); code != http.StatusOK {
			t.Fatalf("attribution : %d, attendu 200", code)
		}
		perms, err := middleware.LoadPermissions(tx, f.member.id)
		if err != nil || perms[models.PermCommentsModerate] != models.ScopeAll || perms[models.PermPollsManage] != models.ScopeAll {
			t.Errorf("permissions après attribution : %v (%v)", perms, err)
		}
		if code := status(admin, http.MethodPut, path, map[string]interface{}{"role_ids": []uint{roleID + 100000}}); code != http.StatusBadRequest {
			t.Errorf("rôle inexistant : %d, attendu 400", code)
		}
		if code := status(inGroups, http.MethodPut, path, map[string]interface{}{"role_ids": []uint{}}); code != http.StatusForbidden {
			t.Errorf("non admin : %d, attendu 403", code)
		}
	})

	// En dernier : une création réussie lance des tâches de fond (notification, XP) qui
	// utilisent la même transaction de test que les requêtes suivantes.
	t.Run("ses groupes : création dans son groupe", func(t *testing.T) {
		if code := status(inGroups, http.MethodPost, "/group-admin/polls", poll("m-g1", f.g1.ID)); code != http.StatusCreated {
			t.Errorf("sondage ciblant g1 : %d, attendu 201", code)
		}
	})
}

// Le groupe par défaut réunit tout le monde : la portée « ses groupes » ne le couvre pas.
func TestGroupsScopeExcludesEveryoneGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	f := newFixture(t, tx) // f.member est membre de g1

	// g1 devient le groupe par défaut
	res := tx.Exec("UPDATE app_settings SET default_group_id = ?", f.g1.ID)
	if res.Error != nil {
		t.Fatalf("paramètres : %v", res.Error)
	}
	if res.RowsAffected == 0 {
		if err := tx.Create(&models.AppSettings{DefaultGroupID: &f.g1.ID}).Error; err != nil {
			t.Fatalf("paramètres : %v", err)
		}
	}
	grantPermissions(t, tx, f.tag+"-sondages-groupes", map[string]string{models.PermPollsManage: models.ScopeGroups}, f.member.id)

	ids, err := models.PermissionGroupIDs(tx, f.member.id)
	if err != nil {
		t.Fatalf("groupes : %v", err)
	}
	for _, id := range ids {
		if id == f.g1.ID {
			t.Fatalf("le groupe par défaut %d est couvert : %v", f.g1.ID, ids)
		}
	}

	r := pollsRouter(t, tx, f.member)
	code, _, _ := send(t, r, http.MethodPost, "/api/v1/group-admin/polls", map[string]interface{}{
		"title": f.title("defaut"), "poll_type": "single", "show_results": "always", "is_active": true,
		"options":          []map[string]interface{}{{"text": "oui"}, {"text": "non"}},
		"target_group_ids": []uint{f.g1.ID},
	})
	if code != http.StatusForbidden {
		t.Errorf("sondage ciblant le groupe par défaut : %d, attendu 403", code)
	}
}
