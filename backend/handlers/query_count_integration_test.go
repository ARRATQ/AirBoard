//go:build integration

package handlers

// Compte les requêtes SQL émises par les endpoints les plus lourds, pour repérer
// les N+1. Nécessite une base Postgres peuplée ; exclu de `go test ./...`.
//
//	AIRBOARD_TEST_DSN="host=localhost user=airboard password=... dbname=airboard sslmode=disable" \
//	  go test -tags integration -run TestQueryCounts -v ./handlers/
//
// Utilisateurs lus en base : AIRBOARD_TEST_ADMIN (défaut "admin") et
// AIRBOARD_TEST_USER (défaut : premier utilisateur actif de rôle "user").

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	"airboard/config"
	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openCountingDB(t *testing.T) (*gorm.DB, *atomic.Int64) {
	t.Helper()
	dsn := os.Getenv("AIRBOARD_TEST_DSN")
	if dsn == "" {
		t.Skip("AIRBOARD_TEST_DSN non défini")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	var n atomic.Int64
	count := func(*gorm.DB) { n.Add(1) }
	cb := db.Callback()
	register := []error{
		cb.Query().After("gorm:query").Register("test:count_query", count),
		cb.Row().After("gorm:row").Register("test:count_row", count),
		cb.Raw().After("gorm:raw").Register("test:count_raw", count),
		cb.Create().After("gorm:create").Register("test:count_create", count),
		cb.Update().After("gorm:update").Register("test:count_update", count),
		cb.Delete().After("gorm:delete").Register("test:count_delete", count),
	}
	for _, err := range register {
		if err != nil {
			t.Fatalf("callback: %v", err)
		}
	}
	return db, &n
}

type identity struct {
	id      uint
	role    string
	managed []uint
	// perms : permissions des rôles, comme RequireAuth les pose (permission -> portée)
	perms map[string]string
}

func loadIdentity(t *testing.T, db *gorm.DB, username, role string) identity {
	t.Helper()
	var u models.User
	q := db.Where("is_active = ?", true)
	if username != "" {
		q = q.Where("username = ?", username)
	} else {
		q = q.Where("role = ?", role).Order("id")
	}
	if err := q.First(&u).Error; err != nil {
		t.Fatalf("utilisateur %q/%q introuvable: %v", username, role, err)
	}
	// Groupes de la portée « ses groupes » (ceux que RequirePermission poserait)
	managed, err := models.PermissionGroupIDs(db, u.ID)
	if err != nil {
		t.Fatalf("groupes de %q : %v", u.Username, err)
	}
	perms, err := middleware.LoadPermissions(db, u.ID)
	if err != nil {
		t.Fatalf("permissions de %q : %v", u.Username, err)
	}
	return identity{id: u.ID, role: u.Role, managed: managed, perms: perms}
}

func runHandler(h gin.HandlerFunc, who identity, params gin.Params) int {
	return runHandlerAt(h, who, params, "/")
}

func runHandlerAt(h gin.HandlerFunc, who identity, params gin.Params, target string) int {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	c.Params = params
	c.Set("user_id", who.id)
	c.Set("role", who.role)
	c.Set("managed_group_ids", who.managed)
	c.Set(middleware.CtxPermissions, who.perms)
	c.Set(middleware.CtxPermissionGroups, groupsForPerms(who.perms, who.managed))
	h(c)
	return w.Code
}

func TestQueryCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, n := openCountingDB(t)

	adminName := os.Getenv("AIRBOARD_TEST_ADMIN")
	if adminName == "" {
		adminName = "admin"
	}
	admin := loadIdentity(t, db, adminName, "")
	user := loadIdentity(t, db, os.Getenv("AIRBOARD_TEST_USER"), "user")

	var groupIDs []uint
	db.Model(&models.Group{}).Order("id").Limit(1).Pluck("id", &groupIDs)
	if len(groupIDs) == 0 {
		t.Fatal("aucun groupe en base")
	}

	home := NewHomeHandler(db)
	dash := NewDashboardHandler(db)
	reports := NewReportsHandler(db)
	idParam := func(id uint) gin.Params {
		return gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(id), 10)}}
	}

	cases := []struct {
		name   string
		h      gin.HandlerFunc
		who    identity
		params gin.Params
	}{
		{"dashboard/user", dash.GetDashboard, user, nil},
		{"dashboard/admin", dash.GetDashboard, admin, nil},
		{"home/user", home.GetHomeData, user, nil},
		{"home/admin", home.GetHomeData, admin, nil},
		{"reports/roles", reports.GetRoleReport, admin, nil},
		{"reports/groups", reports.GetGroupReport, admin, nil},
		{"reports/groups/:id", reports.GetGroupDetailReport, admin, idParam(groupIDs[0])},
		{"reports/users", reports.GetUserReport, admin, nil},
		{"reports/users/:id", reports.GetUserDetailReport, admin, idParam(user.id)},
	}

	for _, tc := range cases {
		runHandler(tc.h, tc.who, tc.params) // chauffe les caches applicatifs
		n.Store(0)
		code := runHandler(tc.h, tc.who, tc.params)
		if code != http.StatusOK {
			t.Errorf("%s: statut %d", tc.name, code)
		}
		t.Logf("%-20s %4d requêtes", tc.name, n.Load())
	}
}

// Exécute les listes filtrées par groupes cibles (NOT EXISTS sur *_target_groups),
// pour un utilisateur et un admin : une erreur SQL donnerait un statut 500.
//
//	go test -tags integration -run TestVisibilityQueries -v ./handlers/
func TestVisibilityQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := openCountingDB(t)

	adminName := os.Getenv("AIRBOARD_TEST_ADMIN")
	if adminName == "" {
		adminName = "admin"
	}
	admin := loadIdentity(t, db, adminName, "")
	user := loadIdentity(t, db, os.Getenv("AIRBOARD_TEST_USER"), "user")

	news := NewNewsHandler(db, &config.Config{}, nil)
	events := NewEventsHandler(db, nil)
	polls := NewPollsHandler(db, nil)
	search := NewSearchHandler(db)
	home := NewHomeHandler(db)

	cases := []struct {
		name   string
		h      gin.HandlerFunc
		target string
	}{
		{"news", news.GetNews, "/"},
		{"events", events.GetEvents, "/"},
		{"events/calendar", events.GetCalendarView, "/?start_date=2020-01-01&end_date=2030-12-31"},
		{"polls", polls.GetPolls, "/"},
		{"search", search.GlobalSearch, "/?q=te"},
		{"home", home.GetHomeData, "/"},
	}
	for _, tc := range cases {
		for _, who := range []identity{user, admin} {
			if code := runHandlerAt(tc.h, who, nil, tc.target); code != http.StatusOK {
				t.Errorf("%s (%s): statut %d", tc.name, who.role, code)
			}
		}
	}
}

// GetUsers : pagination, filtres et tri exécutés contre une vraie base.
func TestUsersPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := openCountingDB(t)
	admin := loadIdentity(t, db, "admin", "")
	h := NewAdminHandler(db, &config.Config{}, nil)

	var all int64
	db.Model(&models.User{}).Count(&all)

	for _, target := range []string{
		"/?page=1&page_size=1",
		"/?search=adm&role=admin&status=active&sort=last_login&order=desc",
		"/?search=%25_&sort=email",
	} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, target, nil)
		c.Set("user_id", admin.id)
		c.Set("role", admin.role)
		h.GetUsers(c)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: statut %d", target, w.Code)
		}
		var body struct {
			Data  []models.User `json:"data"`
			Total int64         `json:"total"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s: %v", target, err)
		}
		if target == "/?page=1&page_size=1" && (len(body.Data) != 1 || body.Total != all) {
			t.Errorf("page 1 : %d lignes, total %d (attendu 1 / %d)", len(body.Data), body.Total, all)
		}
		for _, u := range body.Data {
			if u.Password != "" {
				t.Errorf("%s: mot de passe renvoyé", target)
			}
		}
	}
}

// GetComments / GetPendingComments : pages bornées et total exact.
func TestCommentsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := openCountingDB(t)
	admin := loadIdentity(t, db, "admin", "")
	h := NewCommentHandler(db, nil)

	var entity struct {
		EntityType string
		EntityID   uint
		N          int64
	}
	db.Model(&models.Comment{}).Select("entity_type, entity_id, COUNT(*) AS n").
		Group("entity_type, entity_id").Order("n DESC").Limit(1).Scan(&entity)
	if entity.N == 0 {
		t.Skip("aucun commentaire en base")
	}

	call := func(handler gin.HandlerFunc, target string) (int, map[string]interface{}) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, target, nil)
		c.Set("user_id", admin.id)
		c.Set("role", admin.role)
		handler(c)
		var body map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		return w.Code, body
	}

	target := "/?page_size=1&entity_type=" + entity.EntityType + "&entity_id=" + strconv.FormatUint(uint64(entity.EntityID), 10)
	code, body := call(h.GetComments, target)
	if code != http.StatusOK {
		t.Fatalf("GetComments: statut %d", code)
	}
	if got := len(body["comments"].([]interface{})); got != 1 {
		t.Errorf("GetComments: %d commentaires sur la page, attendu 1", got)
	}
	if body["total_pages"] == nil || body["page"].(float64) != 1 {
		t.Errorf("GetComments: pagination absente %v", body)
	}

	code, body = call(h.GetPendingComments, "/?page=1&page_size=5")
	if code != http.StatusOK {
		t.Fatalf("GetPendingComments: statut %d", code)
	}
	var pending int64
	db.Model(&models.Comment{}).Where("is_approved = ?", false).Count(&pending)
	if int64(body["total"].(float64)) != pending {
		t.Errorf("GetPendingComments: total %v, attendu %d", body["total"], pending)
	}
}

// Listes de l'admin de groupe : paginées pour l'admin global et pour un admin de groupe.
func TestGroupAdminListsPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := openCountingDB(t)
	admin := loadIdentity(t, db, "admin", "")
	h := NewGroupAdminHandler(db)

	who := []identity{admin}
	// Un compte qui gère les ressources limitées à ses groupes, s'il y en a un
	var groupAdminID uint
	db.Table("user_access_roles").
		Joins("JOIN access_role_permissions arp ON arp.role_id = user_access_roles.access_role_id").
		Joins("JOIN users ON users.id = user_access_roles.user_id").
		Where("arp.permission = ? AND arp.scope = ? AND users.role <> ? AND users.deleted_at IS NULL",
			models.PermResourcesManage, models.ScopeGroups, "admin").
		Limit(1).Pluck("user_access_roles.user_id", &groupAdminID)
	if groupAdminID != 0 {
		var u models.User
		db.First(&u, groupAdminID)
		who = append(who, loadIdentity(t, db, u.Username, ""))
	}

	for _, id := range who {
		for name, handler := range map[string]gin.HandlerFunc{"app-groups": h.GetAppGroups, "applications": h.GetApplications} {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=1", nil)
			c.Set("user_id", id.id)
			c.Set("role", id.role)
			c.Set("managed_group_ids", id.managed)
			handler(c)
			var body models.PaginatedResponse
			if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &body) != nil {
				t.Fatalf("%s (%s): statut %d", name, id.role, w.Code)
			}
			if body.PageSize != 1 || len(body.Data.([]interface{})) > 1 {
				t.Errorf("%s (%s): page non bornée %+v", name, id.role, body)
			}
			t.Logf("%s (%s, %d groupes gérés): total %d", name, id.role, len(id.managed), body.Total)
		}
	}
}

// Groupes gérés : un nombre de membres calculé en SQL, sans renvoyer les utilisateurs eux-mêmes.
func TestManagedGroupsUserCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, _ := openCountingDB(t)
	admin := loadIdentity(t, db, "admin", "")
	h := NewGroupAdminHandler(db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", admin.id)
	c.Set("role", admin.role)
	h.GetManagedGroups(c)

	var groups []map[string]interface{}
	if w.Code != http.StatusOK || json.Unmarshal(w.Body.Bytes(), &groups) != nil {
		t.Fatalf("statut %d : %s", w.Code, w.Body.String())
	}
	for _, g := range groups {
		if _, leaked := g["users"]; leaked {
			t.Fatalf("groupe %v : la liste des utilisateurs est renvoyée", g["name"])
		}
		count, ok := g["user_count"].(float64)
		if !ok {
			t.Fatalf("groupe %v : user_count absent", g["name"])
		}
		var want int64
		db.Table("user_groups").Joins("JOIN users ON users.id = user_groups.user_id AND users.deleted_at IS NULL").
			Where("user_groups.group_id = ?", g["id"]).Count(&want)
		if int64(count) != want {
			t.Errorf("groupe %v : user_count %v, attendu %d", g["name"], count, want)
		}
	}
	t.Logf("%d groupes", len(groups))
}

// groupsForPerms : les mêmes groupes pour chaque permission de portée « ses groupes »
// (contexte posé à la main par les tests, comme le ferait RequireAuth).
func groupsForPerms(perms map[string]string, groups []uint) map[string][]uint {
	out := map[string][]uint{}
	for perm, scope := range perms {
		if scope == models.ScopeGroups {
			out[perm] = groups
		}
	}
	return out
}
