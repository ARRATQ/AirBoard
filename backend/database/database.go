// Package database regroupe le réglage du pool de connexions et les index
// de performance créés au démarrage.
package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// Pool de connexions Postgres. Sans réglage, database/sql n'a aucune limite de
// connexions ouvertes et ne garde que 2 connexions inactives : chaque pic de trafic
// ouvre puis ferme des connexions (coûteux), au risque d'atteindre max_connections.
const (
	dbMaxOpenConns    = 25
	dbMaxIdleConns    = 10
	dbConnMaxLifetime = 30 * time.Minute
	dbConnMaxIdleTime = 5 * time.Minute
)

// ConfigurePool règle le pool de connexions de database/sql.
func ConfigurePool(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("accès au pool SQL: %w", err)
	}
	sqlDB.SetMaxOpenConns(dbMaxOpenConns)
	sqlDB.SetMaxIdleConns(dbMaxIdleConns)
	sqlDB.SetConnMaxLifetime(dbConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbConnMaxIdleTime)
	return nil
}

type indexSpec struct {
	name    string
	table   string
	columns string
}

// Index de performance absents du schéma AutoMigrate : clés étrangères et colonnes
// filtrées par les pages les plus fréquentées (accueil, ressources, actualités,
// rapports, commentaires, visibilité par groupe).
var performanceIndexes = []indexSpec{
	{"idx_application_clicks_app_clicked", "application_clicks", "application_id, clicked_at"},
	{"idx_application_clicks_user", "application_clicks", "user_id"},
	{"idx_applications_app_group", "applications", "app_group_id"},
	{"idx_comments_entity", "comments", "entity_type, entity_id"},
	{"idx_comments_user", "comments", "user_id"},
	{"idx_comments_parent", "comments", "parent_id"},
	{"idx_news_author", "news", "author_id"},
	{"idx_news_category", "news", "category_id"},
	{"idx_events_category", "events", "category_id"},
	{"idx_polls_author", "polls", "author_id"},
	// Tables de jointure : la clé primaire commence par l'autre colonne,
	// la recherche « par groupe » / « par utilisateur » n'était pas indexée.
	{"idx_user_groups_user", "user_groups", "user_id"},
	{"idx_group_admins_group", "group_admins", "group_id"},
	{"idx_group_app_groups_group", "group_app_groups", "group_id"},
	{"idx_news_target_groups_group", "news_target_groups", "group_id"},
	{"idx_event_target_groups_group", "event_target_groups", "group_id"},
	{"idx_poll_target_groups_group", "poll_target_groups", "group_id"},
}

// Une lecture et une réaction par utilisateur et par article.
var uniquePerformanceIndexes = []indexSpec{
	{"idx_news_reads_news_user", "news_reads", "news_id, user_id"},
	{"idx_news_reactions_news_user", "news_reactions", "news_id, user_id"},
}

func indexStatement(spec indexSpec, unique bool) string {
	kind := "INDEX"
	if unique {
		kind = "UNIQUE INDEX"
	}
	return fmt.Sprintf("CREATE %s IF NOT EXISTS %s ON %s (%s)", kind, spec.name, spec.table, spec.columns)
}

// EnsurePerformanceIndexes crée les index manquants. Un échec n'empêche pas le
// démarrage : il est seulement journalisé.
func EnsurePerformanceIndexes(db *gorm.DB) {
	for _, spec := range performanceIndexes {
		if err := db.Exec(indexStatement(spec, false)).Error; err != nil {
			log.Printf("Avertissement: index %s non créé: %v", spec.name, err)
		}
	}
	for _, spec := range uniquePerformanceIndexes {
		err := db.Exec(indexStatement(spec, true)).Error
		if err == nil {
			continue
		}
		// Des doublons existants empêchent l'unicité : aucune donnée n'est supprimée
		// au démarrage, on se contente d'un index simple pour les lectures.
		log.Printf("Avertissement: index unique %s impossible (doublons ?): %v — index simple à la place", spec.name, err)
		fallback := indexSpec{spec.name + "_nonunique", spec.table, spec.columns}
		if err := db.Exec(indexStatement(fallback, false)).Error; err != nil {
			log.Printf("Avertissement: index %s non créé: %v", fallback.name, err)
		}
	}
	log.Printf("✓ Index de performance vérifiés (%d + %d uniques)", len(performanceIndexes), len(uniquePerformanceIndexes))
}
