package database

import (
	"context"
	"fmt"
	"io/fs"
	"log"

	"airboard/migrations"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
	"gorm.io/gorm"
)

// Migrate applique les migrations SQL embarquées (backend/migrations) qui ne l'ont
// pas encore été, dans l'ordre de leurs numéros. goose note les versions appliquées
// dans la table goose_db_version ; un verrou consultatif Postgres empêche deux
// instances de migrer en même temps.
func Migrate(ctx context.Context, db *gorm.DB) error {
	return migrateFS(ctx, db, migrations.FS)
}

func migrateFS(ctx context.Context, db *gorm.DB, fsys fs.FS) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("accès au pool SQL: %w", err)
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("verrou de migration: %w", err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, fsys, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("lecture des migrations: %w", err)
	}

	results, err := provider.Up(ctx)
	for _, r := range results {
		log.Printf("Migration %s", r)
	}
	if err != nil {
		return fmt.Errorf("application des migrations: %w", err)
	}

	version, err := provider.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("lecture de la version du schéma: %w", err)
	}
	log.Printf("✓ Schéma à jour (version %d, %d migration(s) appliquée(s) maintenant)", version, len(results))
	reportDegradedSchema(db)
	return nil
}

// reportDegradedSchema signale ce que les migrations ont dû laisser de côté sur une
// base existante : index uniques remplacés par un index simple à cause de doublons,
// clés étrangères posées NOT VALID à cause de lignes orphelines. Les avertissements
// Postgres (RAISE WARNING) ne remontent pas dans les journaux de l'application.
func reportDegradedSchema(db *gorm.DB) {
	var fallbacks []string
	if err := db.Raw(`SELECT indexname FROM pg_indexes
		WHERE schemaname = 'public' AND indexname LIKE '%\_nonunique' ORDER BY 1`).
		Scan(&fallbacks).Error; err != nil {
		log.Printf("Avertissement: vérification des index de repli impossible: %v", err)
	}
	for _, name := range fallbacks {
		log.Printf("Avertissement: index unique remplacé par %s (doublons en base)", name)
	}

	var notValid []string
	if err := db.Raw(`SELECT conrelid::regclass || '.' || conname FROM pg_constraint
		WHERE connamespace = 'public'::regnamespace AND NOT convalidated ORDER BY 1`).
		Scan(&notValid).Error; err != nil {
		log.Printf("Avertissement: vérification des contraintes impossible: %v", err)
	}
	for _, name := range notValid {
		log.Printf("Avertissement: contrainte %s non validée (lignes orphelines en base)", name)
	}
}
