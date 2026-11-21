package handlers

import (
	"fmt"
	"log"

	"airboard/models"

	"gorm.io/gorm"
)

// commentTarget décrit l'élément commenté, pour la file de modération.
type commentTarget struct {
	Title string `json:"title"`
	Slug  string `json:"slug,omitempty"`
}

// Tables et colonnes lues par type d'élément commentable. Les éléments supprimés
// en douceur restent lus : un commentaire en attente peut viser un article retiré.
var commentTargetSources = map[string]struct{ table, title, slug string }{
	"news":        {"news", "title", "slug"},
	"event":       {"events", "title", "slug"},
	"application": {"applications", "name", ""},
}

func commentTargetKey(entityType string, id uint) string {
	return fmt.Sprintf("%s:%d", entityType, id)
}

// groupCommentTargets regroupe les identifiants commentés par type connu, sans doublon.
func groupCommentTargets(comments []models.Comment) map[string][]uint {
	groups := map[string][]uint{}
	seen := map[string]bool{}
	for _, comment := range comments {
		if _, ok := commentTargetSources[comment.EntityType]; !ok {
			continue
		}
		key := commentTargetKey(comment.EntityType, comment.EntityID)
		if seen[key] {
			continue
		}
		seen[key] = true
		groups[comment.EntityType] = append(groups[comment.EntityType], comment.EntityID)
	}
	return groups
}

// loadCommentTargets lit le titre (et le slug) des éléments commentés : une requête par type.
// Une erreur de lecture n'empêche pas la modération : l'élément s'affiche alors sans titre.
func loadCommentTargets(db *gorm.DB, comments []models.Comment) map[string]commentTarget {
	targets := map[string]commentTarget{}
	for entityType, ids := range groupCommentTargets(comments) {
		source := commentTargetSources[entityType]
		columns := "id, " + source.title + " AS title"
		if source.slug != "" {
			columns += ", " + source.slug + " AS slug"
		}
		var rows []struct {
			ID    uint
			Title string
			Slug  string
		}
		if err := db.Table(source.table).Select(columns).Where("id IN ?", ids).Scan(&rows).Error; err != nil {
			log.Printf("[GetPendingComments] titres des %s : %v", source.table, err)
			continue
		}
		for _, row := range rows {
			targets[commentTargetKey(entityType, row.ID)] = commentTarget{Title: row.Title, Slug: row.Slug}
		}
	}
	return targets
}
