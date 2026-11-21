package main

import (
	"fmt"
	"log"
	"os"

	"airboard/config"
	"airboard/models"
	"airboard/services"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func createInitialData(db *gorm.DB, cfg *config.Config) (err error) {
	// Commencer une transaction pour garantir la cohérence des données
	tx := db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("panic during data initialization: %v", r)
		}
	}()

	// Créer l'utilisateur admin par défaut s'il n'existe pas.
	// SECURITY: un admin existant n'est JAMAIS modifié au démarrage, sauf si
	// ADMIN_RESET_PASSWORD=true est défini explicitement (procédure de récupération).
	var adminUser models.User
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}
	// Utiliser coût bcrypt sécurisé (12 minimum - OWASP 2025)
	hashedAdminPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(adminPassword), cfg.Security.BcryptCost)
	if bcryptErr != nil {
		return fmt.Errorf("failed to hash admin password: %w", bcryptErr)
	}

	if err = tx.Unscoped().Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Créer l'utilisateur admin
			adminUser = models.User{
				Username:  "admin",
				Email:     "admin@airboard.com",
				Password:  string(hashedAdminPassword),
				FirstName: "Admin",
				LastName:  "Airboard",
				Role:      "admin",
				IsActive:  true,
			}
			if err = tx.Create(&adminUser).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create admin user: %w", err)
			}
			if os.Getenv("ADMIN_PASSWORD") == "" {
				log.Println("⚠️ Utilisateur admin créé avec le mot de passe par défaut 'admin123' - changez-le immédiatement ou définissez ADMIN_PASSWORD")
			} else {
				log.Println("✅ Utilisateur admin créé avec le mot de passe défini dans ADMIN_PASSWORD")
			}
		} else {
			tx.Rollback()
			return fmt.Errorf("failed to check for existing admin user: %w", err)
		}
	} else if os.Getenv("ADMIN_RESET_PASSWORD") == "true" {
		// Récupération explicite : réactive le compte et applique ADMIN_PASSWORD (ou admin123)
		adminUser.Password = string(hashedAdminPassword)
		adminUser.IsActive = true
		adminUser.Role = "admin"
		adminUser.DeletedAt = gorm.DeletedAt{}
		if err = tx.Unscoped().Save(&adminUser).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to reset admin user: %w", err)
		}
		log.Println("🔄 ADMIN_RESET_PASSWORD=true : compte admin réinitialisé. Retirez cette variable après connexion.")
	}

	// Données de démonstration (compte user/user123, applications et groupe exemples) :
	// uniquement si SEED_DEMO_DATA=true, jamais par défaut en production.
	if os.Getenv("SEED_DEMO_DATA") == "true" {
		if err = seedDemoData(tx, cfg); err != nil {
			tx.Rollback()
			return err
		}
	} else if err = disableDemoUserWithDefaultPassword(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Créer un groupe "Common" par défaut pour tous les nouveaux utilisateurs
	var commonGroup models.Group
	if err = tx.Unscoped().Where("LOWER(name) = ?", "common").First(&commonGroup).Error; err == gorm.ErrRecordNotFound {
		commonGroup = models.Group{
			Name:        "Common",
			Description: "Groupe par défaut pour tous les utilisateurs",
			Color:       "#6B7280",
			IsActive:    true,
		}
		if err = tx.Create(&commonGroup).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create common group: %w", err)
		}
		log.Println("✅ Groupe d'utilisateurs Common créé")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return fmt.Errorf("failed to check for common group: %w", err)
	}

	// Créer les fournisseurs OAuth par défaut
	if err = createDefaultOAuthProviders(tx, cfg, adminUser); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create OAuth providers: %w", err)
	}

	// Créer les templates email par défaut
	if err = createDefaultEmailTemplates(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create email templates: %w", err)
	}

	// Valider la transaction
	if err = tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ensureDefaultSuggestionCategories(db *gorm.DB) {
	defaults := []models.SuggestionCategory{
		{Name: "General", Slug: "general", Order: 1, IsActive: true},
		{Name: "Fonctionnalite", Slug: "functionality", Order: 2, IsActive: true},
		{Name: "Interface", Slug: "ui", Order: 3, IsActive: true},
		{Name: "Process", Slug: "process", Order: 4, IsActive: true},
		{Name: "Performance", Slug: "performance", Order: 5, IsActive: true},
		{Name: "Bug", Slug: "bug", Order: 6, IsActive: true},
		{Name: "Autre", Slug: "other", Order: 7, IsActive: true},
	}

	for _, category := range defaults {
		var existing models.SuggestionCategory
		err := db.Where("slug = ?", category.Slug).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if createErr := db.Create(&category).Error; createErr != nil {
				log.Printf("Avertissement: impossible de creer la categorie suggestion %s: %v", category.Slug, createErr)
			}
			continue
		}
		if err != nil {
			log.Printf("Avertissement: impossible de verifier la categorie suggestion %s: %v", category.Slug, err)
		}
	}
}

func ensureDefaultNewsTypes(db *gorm.DB) {
	defaults := []models.NewsType{
		{Name: "Article", Slug: "article", Icon: "mdi:text-box", Color: "#3B82F6", Order: 1, IsActive: true},
		{Name: "Tutorial", Slug: "tutorial", Icon: "mdi:book-open-variant", Color: "#10B981", Order: 2, IsActive: true},
		{Name: "Announcement", Slug: "announcement", Icon: "mdi:bullhorn", Color: "#8B5CF6", Order: 3, IsActive: true},
		{Name: "FAQ", Slug: "faq", Icon: "mdi:help-circle", Color: "#F59E0B", Order: 4, IsActive: true},
	}

	for _, newsType := range defaults {
		var existing models.NewsType
		err := db.Where("slug = ?", newsType.Slug).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if createErr := db.Create(&newsType).Error; createErr != nil {
				log.Printf("Avertissement: impossible de creer le type de news %s: %v", newsType.Slug, createErr)
			}
			continue
		}
		if err != nil {
			log.Printf("Avertissement: impossible de verifier le type de news %s: %v", newsType.Slug, err)
		}
	}
	log.Println("✓ Types d'articles par défaut créés/vérifiés")
}

func createDefaultEmailTemplates(db *gorm.DB) error {
	// Vérifier si des templates existent déjà
	var count int64
	if err := db.Model(&models.EmailTemplate{}).Count(&count).Error; err != nil {
		return fmt.Errorf("failed to count email templates: %w", err)
	}

	if count > 0 {
		return nil // Templates déjà créés
	}

	// Créer les templates par défaut
	templates := models.GetDefaultEmailTemplates()
	for _, t := range templates {
		if err := db.Create(&t).Error; err != nil {
			return fmt.Errorf("failed to create email template %s: %w", t.Type, err)
		}
	}

	log.Println("✅ Templates d'email par défaut créés (4 templates)")
	return nil
}

// ensureChatMessageTemplate crée le template email chat_message s'il n'existe pas encore
// (pour les installations existantes qui ont déjà les 4 templates de base)
func ensureChatMessageTemplate(db *gorm.DB) {
	var count int64
	db.Model(&models.EmailTemplate{}).Where("type = ?", "chat_message").Count(&count)
	if count > 0 {
		return
	}

	defaults := models.GetDefaultEmailTemplates()
	for _, t := range defaults {
		if t.Type == "chat_message" {
			if err := db.Create(&t).Error; err != nil {
				log.Printf("Avertissement: impossible de créer le template email chat_message: %v", err)
			} else {
				log.Println("✅ Template email chat_message créé")
			}
			return
		}
	}
}

// seedGamification crée les badges et règles d'XP par défaut.
func seedGamification(gamificationService *services.GamificationService) {
	// Seeding gamification
	if err := gamificationService.SeedAchievements(); err != nil {
		log.Printf("Erreur lors du seeding des achievements: %v", err)
	}
	if err := gamificationService.SeedGamificationRules(); err != nil {
		log.Printf("Erreur lors du seeding des règles de gamification: %v", err)
	}
}
