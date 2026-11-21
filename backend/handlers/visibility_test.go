package handlers

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"airboard/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// dryRunDB construit les requêtes sans se connecter : on inspecte le SQL produit.
func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=dry sslmode=disable"}),
		&gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("dry run: %v", err)
	}
	return db
}

func compactSQL(sql string) string {
	return strings.Join(strings.Fields(sql), " ")
}

func newsSQL(t *testing.T, scopes ...func(*gorm.DB) *gorm.DB) (string, []interface{}) {
	t.Helper()
	var rows []models.News
	stmt := dryRunDB(t).Model(&models.News{}).Where("category_id = ?", 7).Scopes(scopes...).Find(&rows).Statement
	return compactSQL(stmt.SQL.String()), stmt.Vars
}

func TestVisibleToSQL(t *testing.T) {
	notExists := "NOT EXISTS (SELECT 1 FROM news_target_groups WHERE news_target_groups.news_id = news.id)"
	exists := "EXISTS (SELECT 1 FROM news_target_groups WHERE news_target_groups.news_id = news.id AND news_target_groups.group_id IN ($2,$3))"

	cases := []struct {
		name   string
		groups []uint
		want   string
	}{
		{"sans groupe : contenus globaux seulement", nil, "category_id = $1 AND " + notExists + " AND \"news\".\"deleted_at\" IS NULL"},
		{"avec groupes : globaux ou ciblant un groupe", []uint{4, 9},
			"category_id = $1 AND ((" + notExists + " OR " + exists + ")) AND \"news\".\"deleted_at\" IS NULL"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sql, _ := newsSQL(t, newsTargets.visibleTo(tc.groups))
			if !strings.HasSuffix(sql, "WHERE "+tc.want) {
				t.Errorf("SQL :\n%s\nattendu en fin :\nWHERE %s", sql, tc.want)
			}
		})
	}
}

// Le OR de la visibilité ne doit jamais s'échapper des autres filtres : la
// suppression logique et le statut publié s'appliquent à toutes les lignes.
func TestVisibilityConditionsStayParenthesized(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	scopes := map[string]func(*gorm.DB) *gorm.DB{
		"visibleTo":               newsTargets.visibleTo([]uint{1}),
		"publishedVisibleTo":      newsTargets.publishedVisibleTo([]uint{1}, now),
		"ownOrPublishedVisibleTo": newsTargets.ownOrPublishedVisibleTo(5, []uint{1}, now),
		"manageableBy":            newsTargets.manageableBy(5, []uint{1}),
	}
	for name, scope := range scopes {
		t.Run(name, func(t *testing.T) {
			sql, _ := newsSQL(t, scope)
			where := sql[strings.Index(sql, "WHERE ")+len("WHERE "):]
			// Après « category_id = $1 AND », la condition de visibilité forme un seul
			// groupe parenthésé, suivi du filtre deleted_at.
			rest := strings.TrimPrefix(where, "category_id = $1 AND ")
			if !strings.HasPrefix(rest, "(") || !strings.HasSuffix(rest, ") AND \"news\".\"deleted_at\" IS NULL") {
				t.Fatalf("condition non parenthésée : %s", where)
			}
			inner := strings.TrimSuffix(rest, " AND \"news\".\"deleted_at\" IS NULL")
			if depth := parenDepthAtTopLevelOr(inner); depth == 0 {
				t.Errorf("OR au niveau racine : %s", inner)
			}
		})
	}
}

// parenDepthAtTopLevelOr renvoie 0 si un « OR » apparaît hors de toute parenthèse.
func parenDepthAtTopLevelOr(sql string) int {
	depth := 0
	for i := 0; i < len(sql); i++ {
		switch sql[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 && strings.HasPrefix(sql[i:], " OR ") {
			return 0
		}
	}
	return 1
}

func TestScopeArguments(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	cases := []struct {
		name     string
		scope    func(*gorm.DB) *gorm.DB
		wantSQL  string
		wantVars []interface{}
	}{
		{
			"publiées sans groupe",
			newsTargets.publishedVisibleTo(nil, now),
			"(is_published = $2 AND (published_at IS NULL OR published_at <= $3) AND NOT EXISTS",
			[]interface{}{7, true, now},
		},
		{
			"auteur ou publiées des groupes",
			newsTargets.ownOrPublishedVisibleTo(5, []uint{3}, now),
			"(author_id = $2 OR (is_published = $3 AND (published_at IS NULL OR published_at <= $4) AND (NOT EXISTS",
			[]interface{}{7, uint(5), true, now, uint(3)},
		},
		{
			"gérables sans groupe géré : ses contenus",
			newsTargets.manageableBy(5, nil),
			"category_id = $1 AND author_id = $2 AND",
			[]interface{}{7, uint(5)},
		},
		{
			"gérables : ses contenus ou ceux de ses groupes",
			newsTargets.manageableBy(5, []uint{3, 8}),
			"(author_id = $2 OR EXISTS (SELECT 1 FROM news_target_groups WHERE news_target_groups.news_id = news.id AND news_target_groups.group_id IN ($3,$4)))",
			[]interface{}{7, uint(5), uint(3), uint(8)},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sql, vars := newsSQL(t, tc.scope)
			if !strings.Contains(sql, tc.wantSQL) {
				t.Errorf("SQL :\n%s\nne contient pas :\n%s", sql, tc.wantSQL)
			}
			if !reflect.DeepEqual(vars, tc.wantVars) {
				t.Errorf("arguments %#v, attendu %#v", vars, tc.wantVars)
			}
		})
	}
}

func TestTargetTablesMatchModels(t *testing.T) {
	cases := []struct {
		targets targetGroups
		model   interface{}
		join    string
	}{
		{newsTargets, &models.News{}, "news_target_groups"},
		{eventTargets, &models.Event{}, "event_target_groups"},
		{pollTargets, &models.Poll{}, "poll_target_groups"},
	}
	db := dryRunDB(t)
	for _, tc := range cases {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(tc.model); err != nil {
			t.Fatalf("%T: %v", tc.model, err)
		}
		if tc.targets.parent != stmt.Schema.Table {
			t.Errorf("%T : table %q, attendu %q", tc.model, tc.targets.parent, stmt.Schema.Table)
		}
		rel, ok := stmt.Schema.Relationships.Relations["TargetGroups"]
		if !ok || rel.JoinTable == nil {
			t.Fatalf("%T : relation TargetGroups absente", tc.model)
		}
		if rel.JoinTable.Table != tc.join || tc.targets.table != tc.join {
			t.Errorf("%T : table de liaison %q / %q, attendu %q", tc.model, rel.JoinTable.Table, tc.targets.table, tc.join)
		}
	}
}

func TestMergeGroupIDs(t *testing.T) {
	cases := []struct {
		name  string
		lists [][]uint
		want  []uint
	}{
		{"vide", nil, nil},
		{"une liste", [][]uint{{3, 1}}, []uint{1, 3}},
		{"union dédoublonnée et triée", [][]uint{{5, 2, 2}, {2, 9}, nil}, []uint{2, 5, 9}},
	}
	for _, tc := range cases {
		if got := mergeGroupIDs(tc.lists...); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s : %v, attendu %v", tc.name, got, tc.want)
		}
	}
}

func TestTargetsAnyGroup(t *testing.T) {
	groups := func(ids ...uint) []models.Group {
		out := make([]models.Group, len(ids))
		for i, id := range ids {
			out[i].ID = id
		}
		return out
	}
	cases := []struct {
		name    string
		targets []models.Group
		mine    []uint
		want    bool
	}{
		{"contenu global, sans groupe", nil, nil, true},
		{"contenu global, avec groupes", nil, []uint{1}, true},
		{"ciblé, utilisateur sans groupe", groups(1), nil, false},
		{"ciblé, autre groupe", groups(1, 2), []uint{3}, false},
		{"ciblé, un groupe commun", groups(1, 2), []uint{3, 2}, true},
	}
	for _, tc := range cases {
		if got := targetsAnyGroup(tc.targets, tc.mine); got != tc.want {
			t.Errorf("%s : %v, attendu %v", tc.name, got, tc.want)
		}
	}
}

func TestGroupIDsOf(t *testing.T) {
	gs := []models.Group{{}, {}}
	gs[0].ID, gs[1].ID = 4, 2
	if got := groupIDsOf(gs); !reflect.DeepEqual(got, []uint{4, 2}) {
		t.Errorf("groupIDsOf : %v", got)
	}
	if got := groupIDsOf(nil); len(got) != 0 {
		t.Errorf("groupIDsOf(nil) : %v", got)
	}
}
