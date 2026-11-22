//go:build integration

package handlers

// Règles d'autorisation des handlers critiques, contre une vraie base Postgres.
// Tout est créé dans une transaction annulée à la fin : la base n'est pas modifiée.
//
//	AIRBOARD_TEST_DSN="host=localhost user=airboard password=... dbname=airboard sslmode=disable" \
//	  go test -tags integration -run 'TestTargetedContent|TestGroupAdminInterface|TestLoginLockout' -v ./handlers/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"testing"
	"time"

	"airboard/config"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// rollbackTx ouvre une transaction annulée à la fin du test.
func rollbackTx(t *testing.T) *gorm.DB {
	t.Helper()
	db, _ := openCountingDB(t)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("transaction: %v", tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	return tx
}

func mustCreate(t *testing.T, tx *gorm.DB, value interface{}) {
	t.Helper()
	if err := tx.Create(value).Error; err != nil {
		t.Fatalf("création %T: %v", value, err)
	}
}

// fixture : deux groupes, des utilisateurs aux rôles variés et des contenus ciblés.
type fixture struct {
	tag                                       string
	member, outsider, groupAdmin, editor, adm identity
	g1, g2                                    models.Group
}

func newFixture(t *testing.T, tx *gorm.DB) fixture {
	t.Helper()
	f := fixture{tag: fmt.Sprintf("c7b-%d", time.Now().UnixNano())}
	f.g1 = models.Group{Name: f.tag + "-g1"}
	f.g2 = models.Group{Name: f.tag + "-g2"}
	mustCreate(t, tx, &f.g1)
	mustCreate(t, tx, &f.g2)

	user := func(name, role string, groups ...models.Group) models.User {
		u := models.User{Username: f.tag + name, Email: f.tag + name + "@example.test", Role: role, Groups: groups}
		mustCreate(t, tx, &u)
		return u
	}
	member := user("member", "user", f.g1)
	outsider := user("outsider", "user")
	ga := user("ga", "user")
	editor := user("editor", "editor", f.g1)
	admin := user("admin", "admin")
	if err := tx.Exec("INSERT INTO group_admins (group_id, user_id) VALUES (?, ?)", f.g2.ID, ga.ID).Error; err != nil {
		t.Fatalf("group_admins: %v", err)
	}

	f.member = identity{id: member.ID, role: "user"}
	f.outsider = identity{id: outsider.ID, role: "user"}
	f.groupAdmin = identity{id: ga.ID, role: "user", managed: []uint{f.g2.ID}}
	// Éditeur : actualités et événements sur tout le site (rôle « Éditeur », migration 00006)
	f.editor = identity{id: editor.ID, role: "editor", perms: map[string]string{
		models.PermNewsManage: models.ScopeAll, models.PermEventsManage: models.ScopeAll,
	}}
	f.adm = identity{id: admin.ID, role: "admin"}
	return f
}

// content : nom court -> (publié, groupes cibles, auteur)
type content struct {
	published bool
	targets   []models.Group
	author    identity
}

func (f fixture) title(name string) string { return f.tag + "-" + name }

func (f fixture) createNews(t *testing.T, tx *gorm.DB, items map[string]content) map[string]models.News {
	t.Helper()
	now := time.Now().Add(-time.Hour)
	out := map[string]models.News{}
	for name, it := range items {
		n := models.News{Title: f.title(name), Slug: f.title(name), AuthorID: it.author.id, IsPublished: it.published, TargetGroups: it.targets}
		if it.published {
			n.PublishedAt = &now
		}
		mustCreate(t, tx, &n)
		out[name] = n
	}
	return out
}

func (f fixture) createEvents(t *testing.T, tx *gorm.DB, items map[string]content) map[string]models.Event {
	t.Helper()
	now := time.Now().Add(-time.Hour)
	out := map[string]models.Event{}
	for name, it := range items {
		e := models.Event{Title: f.title(name), Slug: f.title(name), AuthorID: it.author.id, StartDate: time.Now().Add(48 * time.Hour),
			IsPublished: it.published, TargetGroups: it.targets}
		if it.published {
			e.PublishedAt = &now
		}
		mustCreate(t, tx, &e)
		out[name] = e
	}
	return out
}

// createPolls : « published » signifie ici actif.
func (f fixture) createPolls(t *testing.T, tx *gorm.DB, items map[string]content) map[string]models.Poll {
	t.Helper()
	out := map[string]models.Poll{}
	for name, it := range items {
		p := models.Poll{Title: f.title(name), AuthorID: it.author.id, TargetGroups: it.targets}
		mustCreate(t, tx, &p)
		if !it.published { // is_active a « default:true » : un false explicite serait ignoré à la création
			if err := tx.Model(&p).Update("is_active", false).Error; err != nil {
				t.Fatalf("poll inactif: %v", err)
			}
		}
		out[name] = p
	}
	return out
}

// callJSON exécute un handler en GET et renvoie le statut et le corps décodé.
func callJSON(t *testing.T, h gin.HandlerFunc, who identity, target string, params gin.Params) (int, interface{}) {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, target, nil)
	c.Params = params
	c.Set("user_id", who.id)
	c.Set("role", who.role)
	c.Set("managed_group_ids", who.managed)
	c.Set(middleware.CtxPermissions, who.perms)
	h(c)
	var body interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

// titlesWithPrefix collecte, dans toutes les listes du corps, les titres du fixture.
func titlesWithPrefix(body interface{}, prefix string) []string {
	var out []string
	var walk func(v interface{})
	walk = func(v interface{}) {
		switch x := v.(type) {
		case map[string]interface{}:
			if title, ok := x["title"].(string); ok && len(title) > len(prefix) && title[:len(prefix)] == prefix {
				out = append(out, title[len(prefix)+1:])
			}
			for _, child := range x {
				walk(child)
			}
		case []interface{}:
			for _, child := range x {
				walk(child)
			}
		}
	}
	walk(body)
	sort.Strings(out)
	return dedupe(out)
}

func dedupe(sorted []string) []string {
	var out []string
	for i, s := range sorted {
		if i == 0 || s != sorted[i-1] {
			out = append(out, s)
		}
	}
	return out
}

func expectTitles(t *testing.T, label string, body interface{}, prefix string, want ...string) {
	t.Helper()
	got := titlesWithPrefix(body, prefix)
	sort.Strings(want)
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("%s : %v, attendu %v", label, got, want)
	}
}

func TestTargetedContentVisibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	f := newFixture(t, tx)
	g1, g2 := []models.Group{f.g1}, []models.Group{f.g2}

	news := f.createNews(t, tx, map[string]content{
		"global":       {published: true, author: f.adm},
		"g1":           {published: true, targets: g1, author: f.adm},
		"g1-draft":     {published: false, targets: g1, author: f.adm},
		"g2":           {published: true, targets: g2, author: f.adm},
		"editor-draft": {published: false, targets: g2, author: f.editor},
	})
	events := f.createEvents(t, tx, map[string]content{
		"ev-global":   {published: true, author: f.adm},
		"ev-g1":       {published: true, targets: g1, author: f.adm},
		"ev-g1-draft": {published: false, targets: g1, author: f.adm},
		"ev-g2":       {published: true, targets: g2, author: f.adm},
	})
	polls := f.createPolls(t, tx, map[string]content{
		"poll-global":      {published: true, author: f.adm},
		"poll-g1":          {published: true, targets: g1, author: f.adm},
		"poll-g1-inactive": {published: false, targets: g1, author: f.adm},
		"poll-g2":          {published: true, targets: g2, author: f.adm},
	})

	newsH := NewNewsHandler(tx, &config.Config{}, nil)
	eventsH := NewEventsHandler(tx, nil)
	pollsH := NewPollsHandler(tx, nil)
	searchH := NewSearchHandler(tx)

	t.Run("listes", func(t *testing.T) {
		cases := []struct {
			who                 identity
			news, events, polls []string
		}{
			// Membre de g1 : global + g1 publiés ; ni brouillon, ni sondage inactif, ni g2
			{f.member, []string{"global", "g1"}, []string{"ev-global", "ev-g1"}, []string{"poll-global", "poll-g1"}},
			// Hors groupe : contenus globaux seulement
			{f.outsider, []string{"global"}, []string{"ev-global"}, []string{"poll-global"}},
			// Admin de g2 en lecture publique : global + g2
			{f.groupAdmin, []string{"global", "g2"}, []string{"ev-global", "ev-g2"}, []string{"poll-global", "poll-g2"}},
			// Éditeur (actualités et événements sur tout le site) : tout pour ces deux modules,
			// brouillons compris, comme l'admin ; sondages : ceux de ses groupes
			{f.editor, []string{"global", "g1", "g1-draft", "g2", "editor-draft"},
				[]string{"ev-global", "ev-g1", "ev-g1-draft", "ev-g2"}, []string{"poll-global", "poll-g1"}},
			// Admin : tout, brouillons et sondages inactifs compris
			{f.adm, []string{"global", "g1", "g1-draft", "g2", "editor-draft"},
				[]string{"ev-global", "ev-g1", "ev-g1-draft", "ev-g2"},
				[]string{"poll-global", "poll-g1", "poll-g1-inactive", "poll-g2"}},
		}
		for _, tc := range cases {
			label := fmt.Sprintf("%s#%d", tc.who.role, tc.who.id)
			_, body := callJSON(t, newsH.GetNews, tc.who, "/api/v1/news?page_size=100", nil)
			expectTitles(t, "news "+label, body, f.tag, tc.news...)
			_, body = callJSON(t, eventsH.GetEvents, tc.who, "/api/v1/events?page_size=100", nil)
			expectTitles(t, "events "+label, body, f.tag, tc.events...)
			_, body = callJSON(t, pollsH.GetPolls, tc.who, "/api/v1/polls?page_size=100", nil)
			expectTitles(t, "polls "+label, body, f.tag, tc.polls...)
		}
	})

	t.Run("recherche", func(t *testing.T) {
		_, body := callJSON(t, searchH.GlobalSearch, f.member, "/?q="+f.tag, nil)
		expectTitles(t, "recherche membre", body, f.tag,
			"global", "g1", "ev-global", "ev-g1", "poll-global", "poll-g1", "poll-g1-inactive")
		_, body = callJSON(t, searchH.GlobalSearch, f.outsider, "/?q="+f.tag, nil)
		expectTitles(t, "recherche hors groupe", body, f.tag, "global", "ev-global", "poll-global")
	})

	t.Run("détail", func(t *testing.T) {
		slug := func(s string) gin.Params { return gin.Params{{Key: "slug", Value: s}} }
		id := func(n uint) gin.Params { return gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(n), 10)}} }
		cases := []struct {
			name   string
			h      gin.HandlerFunc
			who    identity
			params gin.Params
			status int
		}{
			{"news g1 / membre", newsH.GetNewsBySlug, f.member, slug(news["g1"].Slug), http.StatusOK},
			{"news g1 / hors groupe", newsH.GetNewsBySlug, f.outsider, slug(news["g1"].Slug), http.StatusForbidden},
			{"news g2 / admin de g2", newsH.GetNewsBySlug, f.groupAdmin, slug(news["g2"].Slug), http.StatusOK},
			{"news brouillon / membre", newsH.GetNewsBySlug, f.member, slug(news["g1-draft"].Slug), http.StatusForbidden},
			{"news brouillon / son auteur editor", newsH.GetNewsBySlug, f.editor, slug(news["editor-draft"].Slug), http.StatusOK},
			{"news globale / hors groupe", newsH.GetNewsBySlug, f.outsider, slug(news["global"].Slug), http.StatusOK},
			{"événement g1 / membre", eventsH.GetEventBySlug, f.member, slug(events["ev-g1"].Slug), http.StatusOK},
			{"événement g1 / hors groupe", eventsH.GetEventBySlug, f.outsider, slug(events["ev-g1"].Slug), http.StatusForbidden},
			{"événement g1 par id / hors groupe", eventsH.GetEventBySlug, f.outsider, slug(strconv.FormatUint(uint64(events["ev-g1"].ID), 10)), http.StatusForbidden},
			{"événement g2 / admin de g2", eventsH.GetEventBySlug, f.groupAdmin, slug(events["ev-g2"].Slug), http.StatusOK},
			{"événement brouillon / membre", eventsH.GetEventBySlug, f.member, slug(events["ev-g1-draft"].Slug), http.StatusForbidden},
			{"sondage g1 / membre", pollsH.GetPollByID, f.member, id(polls["poll-g1"].ID), http.StatusOK},
			{"sondage g1 / hors groupe", pollsH.GetPollByID, f.outsider, id(polls["poll-g1"].ID), http.StatusForbidden},
			{"sondage inactif / membre", pollsH.GetPollByID, f.member, id(polls["poll-g1-inactive"].ID), http.StatusNotFound},
			{"sondage g2 / admin", pollsH.GetPollByID, f.adm, id(polls["poll-g2"].ID), http.StatusOK},
		}
		for _, tc := range cases {
			status, body := callJSON(t, tc.h, tc.who, "/", tc.params)
			if status != tc.status {
				t.Errorf("%s : statut %d, attendu %d (%v)", tc.name, status, tc.status, body)
				continue
			}
			if status >= 400 {
				m, _ := body.(map[string]interface{})
				if m["error"] == "" || m["code"] == nil {
					t.Errorf("%s : corps d'erreur hors format %v", tc.name, body)
				}
			}
		}
	})
}

// L'interface /group-admin ne montre à un admin de groupe que ce qu'il gère :
// ses contenus et ceux qui ciblent ses groupes, brouillons compris.
func TestGroupAdminInterfaceScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	f := newFixture(t, tx)
	g1, g2 := []models.Group{f.g1}, []models.Group{f.g2}

	f.createNews(t, tx, map[string]content{
		"global":   {published: true, author: f.adm},
		"g1":       {published: true, targets: g1, author: f.adm},
		"g2":       {published: true, targets: g2, author: f.adm},
		"g2-draft": {published: false, targets: g2, author: f.adm},
		"own":      {published: false, author: f.groupAdmin},
	})
	f.createEvents(t, tx, map[string]content{
		"ev-global":   {published: true, author: f.adm},
		"ev-g2-draft": {published: false, targets: g2, author: f.adm},
		"ev-g1":       {published: true, targets: g1, author: f.adm},
	})
	f.createPolls(t, tx, map[string]content{
		"poll-global":      {published: true, author: f.adm},
		"poll-g2-inactive": {published: false, targets: g2, author: f.adm},
		"poll-g1":          {published: true, targets: g1, author: f.adm},
	})

	_, body := callJSON(t, NewNewsHandler(tx, &config.Config{}, nil).GetNews, f.groupAdmin, "/api/v1/group-admin/news?page_size=100", nil)
	expectTitles(t, "group-admin/news", body, f.tag, "g2", "g2-draft", "own")
	_, body = callJSON(t, NewEventsHandler(tx, nil).GetEvents, f.groupAdmin, "/api/v1/group-admin/events?page_size=100", nil)
	expectTitles(t, "group-admin/events", body, f.tag, "ev-g2-draft")
	_, body = callJSON(t, NewPollsHandler(tx, nil).GetPolls, f.groupAdmin, "/api/v1/group-admin/polls?page_size=100", nil)
	expectTitles(t, "group-admin/polls", body, f.tag, "poll-g2-inactive")
}

// Verrouillage des connexions : 5 échecs depuis une IP verrouillent le couple
// IP + compte, même avec le bon mot de passe ; une autre IP n'est pas bloquée
// tant que le seuil du compte (10) n'est pas atteint.
func TestLoginLockout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tx := rollbackTx(t)
	tag := fmt.Sprintf("c7blogin%d", time.Now().UnixNano())

	const password = "Correct-horse-42"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	active := models.User{Username: tag, Email: tag + "@example.test", Password: string(hash), Role: "user"}
	inactive := models.User{Username: tag + "off", Email: tag + "off@example.test", Password: string(hash), Role: "user"}
	mustCreate(t, tx, &active)
	mustCreate(t, tx, &inactive)
	if err := tx.Model(&inactive).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	cfg.JWT.Secret = "integration-secret-with-enough-length-0123456789"
	cfg.JWT.TokenExpirationHours = 1
	cfg.JWT.RefreshExpirationDays = 1
	cfg.Security.BcryptCost = bcrypt.MinCost
	h := NewAuthHandler(tx, middleware.NewAuthMiddleware(cfg, tx), false, cfg, services.NewGamificationService(tx))

	login := func(username, pwd, ip string) (int, map[string]interface{}) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(map[string]string{"username": username, "password": pwd})
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Request.RemoteAddr = ip + ":40000"
		h.Login(c)
		var out map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &out)
		return w.Code, out
	}

	steps := []struct {
		name     string
		username string
		password string
		ip       string
		status   int
		code     string
	}{
		{"compte inconnu : même réponse qu'un mauvais mot de passe", tag + "absent", "x", "10.9.0.9", http.StatusUnauthorized, "unauthorized"},
		{"compte désactivé, bon mot de passe", tag + "off", password, "10.9.0.8", http.StatusUnauthorized, "unauthorized"},
		{"échec 1", tag, "faux", "10.9.0.1", http.StatusUnauthorized, "unauthorized"},
		{"échec 2", tag, "faux", "10.9.0.1", http.StatusUnauthorized, "unauthorized"},
		{"échec 3", tag, "faux", "10.9.0.1", http.StatusUnauthorized, "unauthorized"},
		{"échec 4", tag, "faux", "10.9.0.1", http.StatusUnauthorized, "unauthorized"},
		{"échec 5 : verrouillé", tag, "faux", "10.9.0.1", http.StatusTooManyRequests, "too_many_requests"},
		{"bon mot de passe depuis l'IP verrouillée", tag, password, "10.9.0.1", http.StatusTooManyRequests, "too_many_requests"},
		{"bon mot de passe par l'email, IP verrouillée", tag + "@example.test", password, "10.9.0.1", http.StatusTooManyRequests, "too_many_requests"},
		{"bon mot de passe depuis une autre IP", tag, password, "10.9.0.2", http.StatusOK, ""},
	}
	for _, s := range steps {
		status, body := login(s.username, s.password, s.ip)
		if status != s.status {
			t.Fatalf("%s : statut %d, attendu %d (%v)", s.name, status, s.status, body)
		}
		if s.code != "" && body["code"] != s.code {
			t.Errorf("%s : code %v, attendu %q", s.name, body["code"], s.code)
		}
		if s.status == http.StatusOK && (body["token"] == nil || body["refresh_token"] == nil) {
			t.Errorf("%s : jetons absents %v", s.name, body)
		}
	}

	_, unknown := login(tag+"absent", "x", "10.9.0.7")
	_, wrong := login(tag, "faux", "10.9.0.7")
	if unknown["error"] != wrong["error"] {
		t.Errorf("messages distincts : compte inconnu %q, mauvais mot de passe %q", unknown["error"], wrong["error"])
	}
}
