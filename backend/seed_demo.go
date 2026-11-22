package main

import (
	"errors"
	"fmt"
	"log"

	"airboard/config"
	"airboard/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// seedDemoData crée le compte de démonstration user/user123, des groupes
// d'applications exemples et le groupe "Développeurs". Idempotent.
func seedDemoData(tx *gorm.DB, cfg *config.Config) (err error) {

	// Créer un utilisateur normal par défaut
	var normalUser models.User
	var userExists bool
	if err = tx.Unscoped().Where("username = ?", "user").First(&normalUser).Error; err == nil {
		userExists = true
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check for existing user: %w", err)
	}

	if !userExists {
		// Utiliser coût bcrypt sécurisé (12 minimum - OWASP 2025)
		hashedUserPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte("user123"), cfg.Security.BcryptCost)
		if bcryptErr != nil {
			return fmt.Errorf("failed to hash user password: %w", bcryptErr)
		}

		user := models.User{
			Username:  "user",
			Email:     "user@airboard.com",
			Password:  string(hashedUserPassword),
			FirstName: "User",
			LastName:  "Demo",
			Role:      "user",
			IsActive:  true,
		}
		if err = tx.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create default user: %w", err)
		}
		log.Println("✅ Utilisateur demo créé: user@airboard.com / user123")
	} else {
		// Récupérer l'utilisateur existant (y compris les soft-deleted)
		if err = tx.Unscoped().Where("username = ?", "user").First(&normalUser).Error; err != nil {
			return fmt.Errorf("failed to retrieve existing user: %w", err)
		}
	}

	// Créer des groupes d'applications de démonstration
	var devGroup models.AppGroup
	if err = tx.Unscoped().Where("name = ?", "Développement").First(&devGroup).Error; err == gorm.ErrRecordNotFound {
		devGroup = models.AppGroup{
			Name:        "Développement",
			Description: "Applications de développement",
			Color:       "#3B82F6",
			Icon:        "lucide:code",
			Order:       1,
			IsActive:    true,
		}
		if err = tx.Create(&devGroup).Error; err != nil {
			return fmt.Errorf("failed to create development app group: %w", err)
		}

		// Applications de développement
		apps := []models.Application{
			{
				Name:         "GitLab",
				Description:  "Gestion de code source",
				URL:          "https://gitlab.com",
				Icon:         "mdi:gitlab",
				Color:        "#FC6D26",
				Order:        1,
				IsActive:     true,
				OpenInNewTab: true,
				AppGroupID:   devGroup.ID,
			},
			{
				Name:         "Jenkins",
				Description:  "Intégration continue",
				URL:          "https://jenkins.io",
				Icon:         "lucide:bot",
				Color:        "#D33833",
				Order:        2,
				IsActive:     true,
				OpenInNewTab: true,
				AppGroupID:   devGroup.ID,
			},
		}
		for _, app := range apps {
			if err = tx.Create(&app).Error; err != nil {
				return fmt.Errorf("failed to create application %s: %w", app.Name, err)
			}
		}
		log.Println("✅ Groupe Développement créé avec applications de demo")
	} else if err != nil {
		return fmt.Errorf("failed to check for development app group: %w", err)
	}

	// Créer un groupe Production
	var prodGroup models.AppGroup
	if err = tx.Unscoped().Where("name = ?", "Production").First(&prodGroup).Error; err == gorm.ErrRecordNotFound {
		prodGroup = models.AppGroup{
			Name:        "Production",
			Description: "Applications de production",
			Color:       "#10B981",
			Icon:        "lucide:server",
			Order:       2,
			IsActive:    true,
		}
		if err = tx.Create(&prodGroup).Error; err != nil {
			return fmt.Errorf("failed to create production app group: %w", err)
		}

		// Applications de production
		apps := []models.Application{
			{
				Name:         "Grafana",
				Description:  "Monitoring et métriques",
				URL:          "https://grafana.com",
				Icon:         "lucide:chart-line",
				Color:        "#F46800",
				Order:        1,
				IsActive:     true,
				OpenInNewTab: true,
				AppGroupID:   prodGroup.ID,
			},
			{
				Name:         "Prometheus",
				Description:  "Collecte de métriques",
				URL:          "https://prometheus.io",
				Icon:         "lucide:database-search",
				Color:        "#E6522C",
				Order:        2,
				IsActive:     true,
				OpenInNewTab: true,
				AppGroupID:   prodGroup.ID,
			},
		}
		for _, app := range apps {
			if err = tx.Create(&app).Error; err != nil {
				return fmt.Errorf("failed to create application %s: %w", app.Name, err)
			}
		}
		log.Println("✅ Groupe Production créé avec applications de demo")
	} else if err != nil {
		return fmt.Errorf("failed to check for production app group: %w", err)
	}

	// Créer un groupe d'utilisateurs de démonstration
	var demoGroup models.Group
	if err = tx.Unscoped().Where("name = ?", "Développeurs").First(&demoGroup).Error; err == gorm.ErrRecordNotFound {
		demoGroup = models.Group{
			Name:        "Développeurs",
			Description: "Équipe de développement",
			Color:       "#8B5CF6",
			IsActive:    true,
		}
		if err = tx.Create(&demoGroup).Error; err != nil {
			return fmt.Errorf("failed to create developers group: %w", err)
		}

		// Associer l'utilisateur normal au groupe
		if err = tx.Model(&demoGroup).Association("Users").Append(&normalUser); err != nil {
			return fmt.Errorf("failed to associate user with developers group: %w", err)
		}

		// Associer le groupe aux groupes d'applications
		if err = tx.Model(&demoGroup).Association("AppGroups").Append(&devGroup); err != nil {
			return fmt.Errorf("failed to associate dev group with developers group: %w", err)
		}

		if err = tx.Model(&demoGroup).Association("AppGroups").Append(&prodGroup); err != nil {
			return fmt.Errorf("failed to associate prod group with developers group: %w", err)
		}

		log.Println("✅ Groupe d'utilisateurs Développeurs créé avec permissions")
	} else if err != nil {
		return fmt.Errorf("failed to check for developers group: %w", err)
	}

	return nil
}

// disableDemoUserWithDefaultPassword désactive le compte de démonstration "user"
// s'il a encore le mot de passe public "user123" (créé par une ancienne version).
func disableDemoUserWithDefaultPassword(tx *gorm.DB) error {
	var demo models.User
	err := tx.Where("username = ? AND is_active = ?", "user", true).First(&demo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check demo user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(demo.Password), []byte("user123")) != nil {
		return nil // mot de passe changé : compte réellement utilisé, on n'y touche pas
	}
	if err := tx.Model(&demo).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to disable demo user: %w", err)
	}
	log.Println("🔒 Compte de démonstration 'user' désactivé (mot de passe public user123). Définissez SEED_DEMO_DATA=true pour le garder.")
	return nil
}
