// Package database regroupe le réglage du pool de connexions et l'application
// des migrations versionnées du schéma.
package database

import (
	"fmt"
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
