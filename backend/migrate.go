package main

import (
	"fmt"
	"log"

	"airboard/database"
	"airboard/models"

	"gorm.io/gorm"
)

// migrateSchema crée ou met à jour le schéma : AutoMigrate des modèles, puis
// index SQL que GORM ne sait pas exprimer.
func migrateSchema(db *gorm.DB) {
	// Migrations
	if err := db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.AppGroup{},
		&models.Application{},
		&models.AppSettings{},
		&models.OAuthProvider{},
		&models.ApplicationClick{},
		&models.Announcement{},
		&models.News{},
		&models.NewsCategory{},
		&models.NewsType{},
		&models.Tag{},
		&models.NewsReaction{},
		&models.NewsRead{},
		&models.Event{},
		&models.EventCategory{},
		&models.SMTPConfig{},
		&models.EmailOAuthConfig{},
		&models.EmailTemplate{},
		&models.EmailNotificationLog{},
		&models.Media{},
		&models.Comment{},
		&models.Feedback{},
		&models.CommentSettings{},
		&models.Notification{},
		&models.Poll{},
		&models.PollOption{},
		&models.PollVote{},
		&models.Suggestion{},
		&models.SuggestionCategory{},
		&models.SuggestionVote{},
		&models.ChatMessage{},         // Chat
		&models.GamificationProfile{}, // Gamification
		&models.Achievement{},
		&models.UserAchievement{},
		&models.XPTransaction{},
		&models.GamificationRule{},
		&models.HeroMessage{}, // Dynamic Hero Messages
		&models.Chatbot{},     // Chatbots n8n
		// Trombinoscope
		&models.OrgUnit{},
		&models.OrgPosition{},
		&models.DirectoryPerson{},
		&models.PositionAssignment{},
		&models.DirectoryField{},
	); err != nil {
		log.Fatal("Erreur lors des migrations:", err)
	}

	// Créer les index uniques pour éviter les doublons
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_feedback_user_entity ON feedbacks(user_id, entity_type, entity_id)").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique pour feedbacks: %v", err)
	}

	// Index unique pour éviter qu'un utilisateur vote plusieurs fois sur la même option
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_poll_vote_user_option ON poll_votes(poll_id, user_id, poll_option_id)").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique pour poll_votes: %v", err)
	}

	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_suggestion_vote_user_unique ON suggestion_votes(suggestion_id, user_id)").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique pour suggestion_votes: %v", err)
	}

	// Fix: Corriger les contraintes d'unicité sur les slugs pour permettre la réutilisation après soft delete
	// News slug
	db.Exec("DROP INDEX IF EXISTS idx_news_slug")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_news_slug ON news(slug) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour news.slug: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour news.slug")
	}

	// NewsCategory slug
	db.Exec("DROP INDEX IF EXISTS idx_category_slug")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_category_slug ON news_categories(slug) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour news_categories.slug: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour news_categories.slug")
	}

	// NewsType slug
	db.Exec("DROP INDEX IF EXISTS idx_news_type_slug")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_news_type_slug ON news_types(slug) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour news_types.slug: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour news_types.slug")
	}

	// Tag slug
	db.Exec("DROP INDEX IF EXISTS idx_tag_slug")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_slug ON tags(slug) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour tags.slug: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour tags.slug")
	}

	// Tag name
	db.Exec("DROP INDEX IF EXISTS idx_tag_name")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tag_name ON tags(name) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour tags.name: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour tags.name")
	}

	// Event slug
	db.Exec("DROP INDEX IF EXISTS idx_event_slug")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_event_slug ON events(slug) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour events.slug: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour events.slug")
	}

	// EventCategory name
	db.Exec("DROP INDEX IF EXISTS idx_event_category_name")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_event_category_name ON event_categories(name) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour event_categories.name: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour event_categories.name")
	}

	// EventCategory slug
	db.Exec("DROP INDEX IF EXISTS idx_event_category_slug")
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_event_category_slug ON event_categories(slug) WHERE deleted_at IS NULL").Error; err != nil {
		log.Printf("Avertissement: Impossible de créer l'index unique partiel pour event_categories.slug: %v", err)
	} else {
		log.Println("✓ Index unique partiel créé/vérifié pour event_categories.slug")
	}

	// Noms et slugs réutilisables après suppression douce : les anciens index uniques
	// complets (noms par défaut de GORM) bloquaient la recréation d'un élément supprimé.
	// Leurs remplaçants partiels sont posés ci-dessus et ci-dessous.
	// Selon la version de GORM qui les a créés, ce sont des contraintes ou de simples index.
	for _, legacy := range []struct{ table, name string }{
		{"news_categories", "idx_news_categories_name"}, {"news_categories", "idx_news_categories_slug"},
		{"news_types", "idx_news_types_name"}, {"news_types", "idx_news_types_slug"},
		{"suggestion_categories", "idx_suggestion_categories_name"}, {"suggestion_categories", "idx_suggestion_categories_slug"},
		{"events", "idx_events_slug"},
		{"event_categories", "idx_event_categories_name"}, {"event_categories", "idx_event_categories_slug"},
		{"tags", "idx_tags_name"}, {"tags", "idx_tags_slug"},
	} {
		if err := db.Exec(fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", legacy.table, legacy.name)).Error; err != nil {
			log.Printf("Avertissement: Impossible de supprimer l'ancienne contrainte %s: %v", legacy.name, err)
		}
		if err := db.Exec("DROP INDEX IF EXISTS " + legacy.name).Error; err != nil {
			log.Printf("Avertissement: Impossible de supprimer l'ancien index %s: %v", legacy.name, err)
		}
	}
	for _, idx := range []struct{ name, table, column string }{
		{"idx_news_category_name", "news_categories", "name"},
		{"idx_category_slug", "news_categories", "slug"},
		{"idx_news_type_name", "news_types", "name"},
		{"idx_news_type_slug", "news_types", "slug"},
		{"idx_suggestion_category_name", "suggestion_categories", "name"},
		{"idx_suggestion_category_slug", "suggestion_categories", "slug"},
	} {
		query := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s(%s) WHERE deleted_at IS NULL", idx.name, idx.table, idx.column)
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Avertissement: Impossible de créer l'index unique partiel %s: %v", idx.name, err)
		}
	}

	database.EnsurePerformanceIndexes(db)
}
