package main

import (
	"context"
	"log"

	"airboard/config"
	"airboard/database"
	"airboard/services"
	"airboard/services/chat"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// Charger la configuration
	cfg := config.LoadConfig()

	// Configuration Gin
	gin.SetMode(cfg.Server.Mode)

	db := openDatabase(cfg)
	if err := database.Migrate(context.Background(), db); err != nil {
		log.Fatalf("Erreur lors des migrations: %v", err)
	}

	if err := seedInitialData(db, cfg); err != nil {
		log.Fatalf("Erreur lors de la création des données initiales: %v", err)
	}

	// Initialiser le service email global
	InitEmailService(db, cfg)

	// Initialiser le service de stockage
	storageService, err := services.NewLocalStorage(cfg.Storage.UploadDir, cfg.Storage.BaseURL)
	if err != nil {
		log.Fatal("Erreur d'initialisation du service de stockage:", err)
	}

	// Gamification
	gamificationService := services.NewGamificationService(db)

	// Initialisation du Chat
	chat.SetAllowedOrigins(cfg.Server.Origins)
	chatHub := chat.NewHub()
	chatHub.EmailService = GetEmailService()
	go chatHub.Run()

	router := newRouter(cfg, db, storageService, gamificationService, chatHub)

	log.Printf("🚀 Serveur Airboard démarré sur le port %s", cfg.Server.Port)
	log.Printf("📊 Dashboard: http://localhost:%s/health", cfg.Server.Port)
	log.Printf("📚 Mode: %s", cfg.Server.Mode)

	// Démarrer le serveur (arrêt propre sur SIGINT/SIGTERM)
	if err := runServer(newHTTPServer(":"+cfg.Server.Port, router), db); err != nil {
		log.Fatalf("Erreur du serveur HTTP: %v", err)
	}
}

// Variable globale pour le service email (utilisée par les handlers)
var emailService *services.EmailService

// InitEmailService initialise le service email global
func InitEmailService(db *gorm.DB, cfg *config.Config) {
	emailService = services.NewEmailService(db, cfg)
}

// GetEmailService retourne le service email global
func GetEmailService() *services.EmailService {
	return emailService
}
