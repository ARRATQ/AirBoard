//go:build integration

package handlers

import (
	"net/http"
	"testing"
	"time"

	"airboard/config"
	"airboard/models"
)

// Supprimer une catégorie de ressources dont une ressource a déjà servi (clics, favoris,
// commentaires) : ces traces partent avec elle au lieu de bloquer la suppression (500).
func TestDeleteAppGroupWithUsedResources(t *testing.T) {
	tx := rollbackTx(t)
	tag := hierarchyTag()

	user := models.User{Username: tag, Email: tag + "@example.test", Role: "user"}
	mustCreate(t, tx, &user)
	category := models.AppGroup{Name: tag + "-categorie"}
	mustCreate(t, tx, &category)
	resource := models.Application{Name: tag + "-ressource", URL: "https://example.test", AppGroupID: category.ID}
	mustCreate(t, tx, &resource)

	mustCreate(t, tx, &models.ApplicationClick{UserID: user.ID, ApplicationID: resource.ID, ClickedAt: time.Now()})
	if err := tx.Exec("INSERT INTO user_favorites (user_id, application_id) VALUES (?, ?)", user.ID, resource.ID).Error; err != nil {
		t.Fatalf("favori: %v", err)
	}
	mustCreate(t, tx, &models.Comment{Content: "avis", UserID: user.ID, EntityType: "application", EntityID: resource.ID})

	h := NewAdminHandler(tx, &config.Config{}, nil)
	admin := identity{id: user.ID, role: "admin"}
	if w := callAs(h.DeleteAppGroup, admin, http.MethodDelete, category.ID, "", nil); w.Code != http.StatusOK {
		t.Fatalf("suppression : statut %d %s", w.Code, w.Body.String())
	}

	counts := map[string]string{
		"app_groups":         "SELECT count(*) FROM app_groups WHERE id = ?",
		"applications":       "SELECT count(*) FROM applications WHERE id = ?",
		"application_clicks": "SELECT count(*) FROM application_clicks WHERE application_id = ?",
		"user_favorites":     "SELECT count(*) FROM user_favorites WHERE application_id = ?",
		"comments":           "SELECT count(*) FROM comments WHERE entity_type = 'application' AND entity_id = ?",
	}
	for table, query := range counts {
		id := resource.ID
		if table == "app_groups" {
			id = category.ID
		}
		var n int64
		if err := tx.Raw(query, id).Scan(&n).Error; err != nil {
			t.Fatalf("%s: %v", table, err)
		}
		if n != 0 {
			t.Errorf("%s : %d ligne(s) restante(s)", table, n)
		}
	}
}
