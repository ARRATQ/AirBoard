//go:build integration

package main

// Lance deux fois migrations + seed sur une base Postgres DÉDIÉE (jamais la base
// de dev partagée) et vérifie que le second passage ne crée rien et ne renvoie
// aucune erreur (ex. clé unique sur les catégories de suggestion).
//
//	createdb -T template0 airboard_migration_test
//	AIRBOARD_MIGRATION_TEST_DSN="host=localhost user=airboard password=... dbname=airboard_migration_test sslmode=disable" \
//	  go test -tags integration -run TestSeed -v .

import (
	"context"
	"os"
	"testing"

	"airboard/config"
	"airboard/database"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedInitialDataIsIdempotent(t *testing.T) {
	dsn := os.Getenv("AIRBOARD_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("AIRBOARD_MIGRATION_TEST_DSN non défini")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("connexion: %v", err)
	}
	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	cfg := &config.Config{}
	cfg.Security.BcryptCost = bcrypt.MinCost
	cfg.Server.PublicURL = "http://localhost:5173"

	tables := []string{
		"users", "groups", "o_auth_providers", "event_categories", "events",
		"comment_settings", "email_templates", "suggestion_categories", "news_types",
		"achievements", "gamification_rules",
	}
	count := func() map[string]int64 {
		counts := make(map[string]int64, len(tables))
		for _, table := range tables {
			var n int64
			if err := db.Table(table).Count(&n).Error; err != nil {
				t.Fatalf("comptage %s: %v", table, err)
			}
			counts[table] = n
		}
		return counts
	}

	if err := seedInitialData(db, cfg); err != nil {
		t.Fatalf("premier seed: %v", err)
	}
	first := count()
	if err := seedInitialData(db, cfg); err != nil {
		t.Fatalf("second seed: %v", err)
	}
	second := count()

	for _, table := range tables {
		if first[table] == 0 && table != "events" {
			t.Errorf("%s vide après le seed", table)
		}
		if first[table] != second[table] {
			t.Errorf("%s : %d lignes puis %d, le seed n'est pas idempotent", table, first[table], second[table])
		}
	}
}
