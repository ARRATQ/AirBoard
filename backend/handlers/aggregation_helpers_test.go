package handlers

import (
	"testing"

	"airboard/models"
)

func TestCountsByID(t *testing.T) {
	got := countsByID([]idCount{{ID: 1, N: 3}, {ID: 7, N: 0}, {ID: 2, N: 12}})
	want := map[uint]int64{1: 3, 7: 0, 2: 12}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for id, n := range want {
		if got[id] != n {
			t.Errorf("counts[%d] = %d, want %d", id, got[id], n)
		}
	}
	if got[99] != 0 {
		t.Errorf("id absent: %d, want 0", got[99])
	}
}

func TestGroupApplications(t *testing.T) {
	groups := []models.AppGroup{{ID: 10}, {ID: 20}, {ID: 30}}
	apps := []models.Application{
		{ID: 1, AppGroupID: 20},
		{ID: 2, AppGroupID: 10},
		{ID: 3, AppGroupID: 20},
		{ID: 4, AppGroupID: 99}, // groupe non demandé : ignoré
	}

	got := groupApplications(groups, apps)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	wantOrder := []uint{10, 20, 30}
	wantApps := [][]uint{{2}, {1, 3}, {}}
	for i, g := range got {
		if g.AppGroup.ID != wantOrder[i] {
			t.Errorf("groupe %d: id %d, want %d", i, g.AppGroup.ID, wantOrder[i])
		}
		if g.Applications == nil {
			t.Errorf("groupe %d: liste nil (JSON null), want []", g.AppGroup.ID)
		}
		if len(g.Applications) != len(wantApps[i]) {
			t.Fatalf("groupe %d: %d apps, want %d", g.AppGroup.ID, len(g.Applications), len(wantApps[i]))
		}
		for j, a := range g.Applications {
			if a.ID != wantApps[i][j] {
				t.Errorf("groupe %d app %d: id %d, want %d", g.AppGroup.ID, j, a.ID, wantApps[i][j])
			}
		}
	}
}

func TestGroupNewsByType(t *testing.T) {
	types := []models.NewsType{{ID: 1, Slug: "article"}, {ID: 2, Slug: "faq"}, {ID: 3, Slug: "tutorial"}}
	refs := []rankedNewsRef{
		{ID: 5, Type: "article"},
		{ID: 3, Type: "article"},
		{ID: 8, Type: "tutorial"},
		{ID: 4, Type: "article"}, // absent de byID (supprimé entre-temps) : ignoré
	}
	byID := map[uint]models.News{
		3: {ID: 3, Type: "article"},
		5: {ID: 5, Type: "article"},
		8: {ID: 8, Type: "tutorial"},
	}

	got := groupNewsByType(types, refs, byID)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2 (faq sans article omis)", len(got))
	}
	if got[0].Type.Slug != "article" || got[1].Type.Slug != "tutorial" {
		t.Errorf("ordre des types = %s, %s", got[0].Type.Slug, got[1].Type.Slug)
	}
	if len(got[0].News) != 2 || got[0].News[0].ID != 5 || got[0].News[1].ID != 3 {
		t.Errorf("article: ordre de rang non respecté: %+v", got[0].News)
	}
	if got[0].Type == got[1].Type {
		t.Error("les onglets partagent le même pointeur de type")
	}
}

func TestGroupNewsByTypeEmpty(t *testing.T) {
	if got := groupNewsByType([]models.NewsType{{Slug: "article"}}, nil, nil); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}
