package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"airboard/config"
	"airboard/models"

	"gorm.io/gorm"
)

func createDefaultOAuthProviders(db *gorm.DB, cfg *config.Config, adminUser models.User) error {
	// Construire les redirect URIs basées sur PUBLIC_URL
	publicURL := cfg.Server.PublicURL

	// Google OAuth
	var googleProvider models.OAuthProvider
	err := db.Where("provider_name = ?", "google").First(&googleProvider).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check for existing Google OAuth provider: %w", err)
	}
	if err == gorm.ErrRecordNotFound {
		// Créer si n'existe pas
		googleProvider = models.OAuthProvider{
			ProviderName: "google",
			DisplayName:  "Google",
			Icon:         "mdi:google",
			IsEnabled:    false,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
			Scopes:       "openid email profile",
			RedirectURI:  publicURL + "/auth/oauth/google/callback",
		}
		if err = db.Create(&googleProvider).Error; err != nil {
			return fmt.Errorf("failed to create Google OAuth provider: %w", err)
		}
		log.Printf("✅ Google OAuth provider créé (désactivé par défaut) - Redirect: %s", googleProvider.RedirectURI)
	} else {
		// Mettre à jour le redirect URI seulement si toujours sur localhost (pas encore configuré en prod)
		newRedirectURI := publicURL + "/auth/oauth/google/callback"
		if googleProvider.RedirectURI != newRedirectURI && strings.Contains(googleProvider.RedirectURI, "localhost") {
			googleProvider.RedirectURI = newRedirectURI
			if err = db.Save(&googleProvider).Error; err != nil {
				return fmt.Errorf("failed to update Google OAuth redirect URI: %w", err)
			}
			log.Printf("🔄 Google OAuth redirect URI mis à jour: %s", googleProvider.RedirectURI)
		}
	}

	// Microsoft OAuth
	var microsoftProvider models.OAuthProvider
	err = db.Where("provider_name = ?", "microsoft").First(&microsoftProvider).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check for existing Microsoft OAuth provider: %w", err)
	}
	if err == gorm.ErrRecordNotFound {
		// Créer si n'existe pas
		microsoftProvider = models.OAuthProvider{
			ProviderName: "microsoft",
			DisplayName:  "Microsoft",
			Icon:         "mdi:microsoft",
			IsEnabled:    false,
			AuthURL:      "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			UserInfoURL:  "https://graph.microsoft.com/v1.0/me",
			Scopes:       "openid email profile User.Read",
			RedirectURI:  publicURL + "/auth/oauth/microsoft/callback",
		}
		if err = db.Create(&microsoftProvider).Error; err != nil {
			return fmt.Errorf("failed to create Microsoft OAuth provider: %w", err)
		}
		log.Printf("✅ Microsoft OAuth provider créé (désactivé par défaut) - Redirect: %s", microsoftProvider.RedirectURI)
	} else {
		// Mettre à jour le redirect URI seulement si toujours sur localhost (pas encore configuré en prod)
		newRedirectURI := publicURL + "/auth/oauth/microsoft/callback"
		if microsoftProvider.RedirectURI != newRedirectURI && strings.Contains(microsoftProvider.RedirectURI, "localhost") {
			microsoftProvider.RedirectURI = newRedirectURI
			if err = db.Save(&microsoftProvider).Error; err != nil {
				return fmt.Errorf("failed to update Microsoft OAuth redirect URI: %w", err)
			}
			log.Printf("🔄 Microsoft OAuth redirect URI mis à jour: %s", microsoftProvider.RedirectURI)
		}
	}

	// Créer des catégories d'événements de démonstration
	var meetingsCategory models.EventCategory
	if err = db.Unscoped().Where("slug = ?", "reunions").First(&meetingsCategory).Error; err == gorm.ErrRecordNotFound {
		meetingsCategory = models.EventCategory{
			Name:        "Réunions",
			Slug:        "reunions",
			Description: "Réunions et assemblées",
			Icon:        "lucide:users",
			Color:       "#3B82F6", // Bleu
			Order:       1,
			IsActive:    true,
		}
		if err = db.Create(&meetingsCategory).Error; err != nil {
			return fmt.Errorf("failed to create meetings event category: %w", err)
		}
		log.Println("✅ Catégorie d'événements 'Réunions' créée")
	} else if err != nil {
		return fmt.Errorf("failed to check for meetings event category: %w", err)
	}

	var trainingsCategory models.EventCategory
	if err = db.Unscoped().Where("slug = ?", "formations").First(&trainingsCategory).Error; err == gorm.ErrRecordNotFound {
		trainingsCategory = models.EventCategory{
			Name:        "Formations",
			Slug:        "formations",
			Description: "Formations et ateliers",
			Icon:        "lucide:graduation-cap",
			Color:       "#F59E0B", // Orange
			Order:       2,
			IsActive:    true,
		}
		if err = db.Create(&trainingsCategory).Error; err != nil {
			return fmt.Errorf("failed to create trainings event category: %w", err)
		}
		log.Println("✅ Catégorie d'événements 'Formations' créée")
	} else if err != nil {
		return fmt.Errorf("failed to check for trainings event category: %w", err)
	}

	var socialCategory models.EventCategory
	if err = db.Unscoped().Where("slug = ?", "evenements-sociaux").First(&socialCategory).Error; err == gorm.ErrRecordNotFound {
		socialCategory = models.EventCategory{
			Name:        "Événements Sociaux",
			Slug:        "evenements-sociaux",
			Description: "Événements sociaux et célébrations",
			Icon:        "lucide:party-popper",
			Color:       "#10B981", // Vert
			Order:       3,
			IsActive:    true,
		}
		if err = db.Create(&socialCategory).Error; err != nil {
			return fmt.Errorf("failed to create social event category: %w", err)
		}
		log.Println("✅ Catégorie d'événements 'Événements Sociaux' créée")
	} else if err != nil {
		return fmt.Errorf("failed to check for social event category: %w", err)
	}

	var holidaysCategory models.EventCategory
	if err = db.Unscoped().Where("slug = ?", "jours-feries").First(&holidaysCategory).Error; err == gorm.ErrRecordNotFound {
		holidaysCategory = models.EventCategory{
			Name:        "Jours Fériés",
			Slug:        "jours-feries",
			Description: "Jours fériés et congés",
			Icon:        "lucide:tree-palm",
			Color:       "#EF4444", // Rouge
			Order:       4,
			IsActive:    true,
		}
		if err = db.Create(&holidaysCategory).Error; err != nil {
			return fmt.Errorf("failed to create holidays event category: %w", err)
		}
		log.Println("✅ Catégorie d'événements 'Jours Fériés' créée")
	} else if err != nil {
		return fmt.Errorf("failed to check for holidays event category: %w", err)
	}

	// Créer des événements de démonstration
	var eventCount int64
	if err = db.Model(&models.Event{}).Unscoped().Count(&eventCount).Error; err != nil {
		return fmt.Errorf("failed to count existing events: %w", err)
	}
	if eventCount == 0 {
		now := time.Now()
		nextWeek := now.AddDate(0, 0, 7)
		nextMonth := now.AddDate(0, 1, 0)
		lastMonth := now.AddDate(0, -1, 0)

		// 1. Événement one-time public (semaine prochaine)
		publishedAt := now
		event1 := models.Event{
			Title:       "Assemblée Générale",
			Slug:        "assemblee-generale",
			Description: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Assemblée générale annuelle de l'entreprise. Tous les employés sont invités à participer."}]}]}`,
			StartDate:   nextWeek,
			EndDate:     &nextWeek,
			IsAllDay:    false,
			Timezone:    "UTC",
			Location:    "Salle de conférence A",
			Color:       meetingsCategory.Color,
			Priority:    "important",
			Status:      "confirmed",
			IsPublished: true,
			PublishedAt: &publishedAt,
			AuthorID:    adminUser.ID,
			CategoryID:  &meetingsCategory.ID,
		}
		if err = db.Create(&event1).Error; err != nil {
			return fmt.Errorf("failed to create general assembly event: %w", err)
		}

		// 2. Événement récurrent hebdomadaire (Daily Standup)
		recurrenceRule := `{"type":"weekly","interval":1,"days_of_week":[1,3,5],"end_type":"never"}`
		event2 := models.Event{
			Title:          "Daily Standup",
			Slug:           "daily-standup",
			Description:    `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Réunion quotidienne de synchronisation d'équipe (15 min)."}]}]}`,
			StartDate:      time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC),
			IsAllDay:       false,
			Timezone:       "UTC",
			IsRecurring:    true,
			RecurrenceRule: recurrenceRule,
			Location:       "Salle de réunion B",
			Color:          meetingsCategory.Color,
			Priority:       "normal",
			Status:         "confirmed",
			IsPublished:    true,
			PublishedAt:    &publishedAt,
			AuthorID:       adminUser.ID,
			CategoryID:     &meetingsCategory.ID,
		}
		if err = db.Create(&event2).Error; err != nil {
			return fmt.Errorf("failed to create daily standup event: %w", err)
		}

		// 3. All-day holiday (Jour de l'An 2026)
		newYear := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		event3 := models.Event{
			Title:       "Jour de l'An",
			Slug:        "jour-de-l-an-2026",
			Description: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Jour férié - Bonne année !"}]}]}`,
			StartDate:   newYear,
			IsAllDay:    true,
			Timezone:    "UTC",
			Color:       holidaysCategory.Color,
			Priority:    "normal",
			Status:      "confirmed",
			IsPublished: true,
			PublishedAt: &publishedAt,
			AuthorID:    adminUser.ID,
			CategoryID:  &holidaysCategory.ID,
		}
		if err = db.Create(&event3).Error; err != nil {
			return fmt.Errorf("failed to create new year event: %w", err)
		}

		// 4. Group-scoped event (IT group only)
		// Récupérer le groupe IT s'il existe
		var itGroup models.Group
		if err = db.Where("name = ?", "IT").First(&itGroup).Error; err == nil {
			event4 := models.Event{
				Title:       "Formation Sécurité",
				Slug:        "formation-securite",
				Description: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Formation obligatoire sur les bonnes pratiques de sécurité informatique."}]}]}`,
				StartDate:   nextMonth,
				EndDate:     &nextMonth,
				IsAllDay:    false,
				Timezone:    "UTC",
				Location:    "Salle de formation",
				Color:       trainingsCategory.Color,
				Priority:    "high",
				Status:      "confirmed",
				IsPublished: true,
				PublishedAt: &publishedAt,
				AuthorID:    adminUser.ID,
				CategoryID:  &trainingsCategory.ID,
			}
			if err = db.Create(&event4).Error; err != nil {
				return fmt.Errorf("failed to create security training event: %w", err)
			}
			if err := db.Model(&event4).Association("TargetGroups").Append(&itGroup); err != nil {
				return fmt.Errorf("failed to associate security training with IT group: %v", err)
			}
		}

		// 5. Récurrent mensuel (premier lundi du mois)
		recurrenceMonthly := `{"type":"monthly","interval":1,"day_of_month":1,"end_type":"never"}`
		firstMonday := time.Date(now.Year(), now.Month(), 1, 10, 0, 0, 0, time.UTC)
		for firstMonday.Weekday() != time.Monday {
			firstMonday = firstMonday.AddDate(0, 0, 1)
		}
		event5 := models.Event{
			Title:          "Séminaire Mensuel",
			Slug:           "seminaire-mensuel",
			Description:    `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Séminaire mensuel de présentation des nouveautés et partage de connaissances."}]}]}`,
			StartDate:      firstMonday,
			IsAllDay:       false,
			Timezone:       "UTC",
			IsRecurring:    true,
			RecurrenceRule: recurrenceMonthly,
			Location:       "Auditorium",
			Color:          meetingsCategory.Color,
			Priority:       "normal",
			Status:         "confirmed",
			IsPublished:    true,
			PublishedAt:    &publishedAt,
			AuthorID:       adminUser.ID,
			CategoryID:     &meetingsCategory.ID,
		}
		if err = db.Create(&event5).Error; err != nil {
			return fmt.Errorf("failed to create monthly seminar event: %w", err)
		}

		// 6. Événement passé (pour tester l'indicateur "passé")
		event6 := models.Event{
			Title:       "Revue Trimestrielle Q4 2024",
			Slug:        "revue-trimestrielle-q4-2024",
			Description: `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Revue des résultats du quatrième trimestre 2024."}]}]}`,
			StartDate:   lastMonth,
			EndDate:     &lastMonth,
			IsAllDay:    false,
			Timezone:    "UTC",
			Location:    "Salle de conférence A",
			Color:       meetingsCategory.Color,
			Priority:    "normal",
			Status:      "confirmed",
			IsPublished: true,
			PublishedAt: &lastMonth,
			AuthorID:    adminUser.ID,
			CategoryID:  &meetingsCategory.ID,
		}
		if err = db.Create(&event6).Error; err != nil {
			return fmt.Errorf("failed to create quarterly review event: %w", err)
		}

		log.Println("✅ Événements de démonstration créés (6 événements)")
	}

	// Créer les paramètres de commentaires par défaut
	var commentSettings models.CommentSettings
	if err = db.First(&commentSettings).Error; err == gorm.ErrRecordNotFound {
		commentSettings = models.CommentSettings{
			CommentsEnabled:      true,
			NewsCommentsEnabled:  true,
			AppCommentsEnabled:   false, // Désactivé par défaut pour les applications
			EventCommentsEnabled: true,
			RequireModeration:    false, // Auto-approuvé par défaut
			AllowAnonymous:       false,
			MaxCommentLength:     1000,
		}
		if err = db.Create(&commentSettings).Error; err != nil {
			return fmt.Errorf("failed to create comment settings: %w", err)
		}
		log.Println("✅ Paramètres de commentaires créés (activés par défaut)")
	} else if err != nil {
		return fmt.Errorf("failed to check for comment settings: %w", err)
	}

	return nil
}
