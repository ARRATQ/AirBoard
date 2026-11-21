package main

import (
	"log"

	"airboard/config"
	"airboard/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// openDatabase ouvre la connexion Postgres et règle le pool de connexions.
func openDatabase(cfg *config.Config) *gorm.DB {
	// Connexion à la base de données
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("Erreur de connexion à la base de données:", err)
	}
	if err := database.ConfigurePool(db); err != nil {
		log.Fatal("Erreur de configuration du pool de connexions:", err)
	}
	return db
}
