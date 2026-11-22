//go:build integration

package database

// Applique les migrations sur une base Postgres DÉDIÉE (jamais la base de dev
// partagée : les migrations la modifieraient), deux fois de suite, puis vérifie que
// chaque modèle GORM a sa table et ses colonnes. Ce dernier point remplace la
// garantie qu'offrait AutoMigrate : un champ ajouté à un modèle sans migration
// fait échouer le test.
//
//	createdb -T template0 airboard_migration_test
//	AIRBOARD_MIGRATION_TEST_DSN="host=localhost user=airboard password=... dbname=airboard_migration_test sslmode=disable" \
//	  go test -tags integration -run TestMigrate -v ./database/

import (
	"context"
	"os"
	"testing"

	"airboard/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Les modèles persistés, ceux que l'ancien démarrage passait à AutoMigrate.
var persistedModels = []interface{}{
	&models.User{}, &models.Group{}, &models.AppGroup{}, &models.Application{},
	&models.AppSettings{}, &models.OAuthProvider{}, &models.ApplicationClick{},
	&models.Announcement{}, &models.News{}, &models.NewsCategory{}, &models.NewsType{},
	&models.Tag{}, &models.NewsReaction{}, &models.NewsRead{}, &models.Event{},
	&models.EventCategory{}, &models.SMTPConfig{}, &models.EmailOAuthConfig{},
	&models.EmailTemplate{}, &models.EmailNotificationLog{}, &models.Media{},
	&models.Comment{}, &models.Feedback{}, &models.CommentSettings{}, &models.Notification{},
	&models.Poll{}, &models.PollOption{}, &models.PollVote{}, &models.Suggestion{},
	&models.SuggestionCategory{}, &models.SuggestionVote{}, &models.ChatMessage{},
	&models.GamificationProfile{}, &models.Achievement{}, &models.UserAchievement{},
	&models.XPTransaction{}, &models.GamificationRule{}, &models.HeroMessage{},
	&models.Chatbot{}, &models.OrgUnit{}, &models.OrgPosition{}, &models.DirectoryPerson{},
	&models.PositionAssignment{}, &models.DirectoryField{},
	&models.AccessRole{}, &models.AccessRolePermission{},
}

func openMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("AIRBOARD_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("AIRBOARD_MIGRATION_TEST_DSN non défini")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	return db
}

func TestMigrateIsIdempotentAndCoversModels(t *testing.T) {
	db := openMigrationTestDB(t)
	ctx := context.Background()

	for run := 1; run <= 2; run++ {
		if err := Migrate(ctx, db); err != nil {
			t.Fatalf("passage %d : %v", run, err)
		}
	}

	migrator := db.Migrator()
	for _, model := range persistedModels {
		stmt := &gorm.Statement{DB: db}
		if err := stmt.Parse(model); err != nil {
			t.Fatalf("analyse du modèle %T: %v", model, err)
		}
		table := stmt.Schema.Table
		if !migrator.HasTable(table) {
			t.Errorf("table %s absente (modèle %T) : migration manquante", table, model)
			continue
		}
		for _, field := range stmt.Schema.Fields {
			if field.DBName != "" && !migrator.HasColumn(model, field.DBName) {
				t.Errorf("colonne %s.%s absente : migration manquante", table, field.DBName)
			}
		}
		for _, rel := range stmt.Schema.Relationships.Many2Many {
			if !migrator.HasTable(rel.JoinTable.Table) {
				t.Errorf("table de jointure %s absente : migration manquante", rel.JoinTable.Table)
			}
		}
	}

	// Les index uniques partiels (réutilisation des slugs après suppression douce).
	var partial int64
	if err := db.Raw(`SELECT count(*) FROM pg_indexes WHERE schemaname = 'public'
		AND indexdef LIKE 'CREATE UNIQUE INDEX%WHERE (deleted_at IS NULL)'`).Scan(&partial).Error; err != nil {
		t.Fatalf("lecture des index: %v", err)
	}
	if partial != 12 {
		t.Errorf("%d index uniques partiels, 12 attendus", partial)
	}
}
