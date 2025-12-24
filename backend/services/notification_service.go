package services

import (
	"airboard/models"
	"fmt"
	"gorm.io/gorm"
)

// NotificationService gère la création de notifications système
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService crée une nouvelle instance du service
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// NotifyLogin crée une notification de connexion
func (s *NotificationService) NotifyLogin(userID uint, isFirstLogin bool) error {
	title := "Connexion réussie"
	message := "Vous vous êtes connecté avec succès"
	icon := "mdi:login"

	if isFirstLogin {
		title = "Bienvenue sur Airboard"
		message = "Première connexion réussie. Bienvenue !"
		icon = "mdi:account-check"
	}

	return s.createNotification(userID, "system", "login", title, message, icon, "#10B981", "", 0)
}

// NotifyRoleChange crée une notification de changement de rôle
func (s *NotificationService) NotifyRoleChange(userID uint, oldRole, newRole string) error {
	title := "Changement de rôle"
	message := fmt.Sprintf("Votre rôle a été modifié de '%s' à '%s'", oldRole, newRole)
	icon := "mdi:account-switch"

	return s.createNotification(userID, "system", "role_change", title, message, icon, "#F59E0B", "", 1)
}

// NotifyAccessGranted crée une notification d'accès accordé à une application
func (s *NotificationService) NotifyAccessGranted(userID uint, appName string, appID uint) error {
	title := "Nouvel accès"
	message := fmt.Sprintf("Vous avez maintenant accès à l'application '%s'", appName)
	icon := "mdi:check-circle"
	actionURL := "/dashboard"

	return s.createNotification(userID, "system", "access_granted", title, message, icon, "#10B981", actionURL, 0)
}

// NotifyAccessRevoked crée une notification d'accès révoqué à une application
func (s *NotificationService) NotifyAccessRevoked(userID uint, appName string) error {
	title := "Accès révoqué"
	message := fmt.Sprintf("Votre accès à l'application '%s' a été révoqué", appName)
	icon := "mdi:close-circle"
	actionURL := "/dashboard"

	return s.createNotification(userID, "system", "access_revoked", title, message, icon, "#EF4444", actionURL, 1)
}

// NotifyNewArticle crée une notification pour un nouvel article
func (s *NotificationService) NotifyNewArticle(title, slug string, authorName string, userIDs []uint) error {
	notifTitle := "Nouvel article"
	message := fmt.Sprintf("'%s' par %s", title, authorName)
	icon := "mdi:newspaper"
	actionURL := fmt.Sprintf("/news/%s", slug)

	return s.createNotificationForUsers(userIDs, "news", "new_article", notifTitle, message, icon, "#3B82F6", actionURL, 0)
}

// NotifyPinnedArticle crée une notification pour un article épinglé
func (s *NotificationService) NotifyPinnedArticle(title, slug string, userIDs []uint) error {
	notifTitle := "Article important épinglé"
	message := fmt.Sprintf("À ne pas manquer: '%s'", title)
	icon := "mdi:pin"
	actionURL := fmt.Sprintf("/news/%s", slug)

	return s.createNotificationForUsers(userIDs, "news", "pinned_article", notifTitle, message, icon, "#EF4444", actionURL, 2)
}

// NotifyNewAnnouncement crée une notification pour une nouvelle annonce
func (s *NotificationService) NotifyNewAnnouncement(title string, announcementType string, userIDs []uint) error {
	notifTitle := "Nouvelle annonce"
	message := title
	icon := "mdi:bullhorn"
	priority := 0
	color := "#3B82F6"

	if announcementType == "warning" {
		priority = 1
		color = "#F59E0B"
	} else if announcementType == "error" {
		priority = 2
		color = "#EF4444"
	}

	return s.createNotificationForUsers(userIDs, "announcement", "new_announcement", notifTitle, message, icon, color, "/dashboard", priority)
}

// NotifyUrgentAnnouncement crée une notification pour une annonce urgente
func (s *NotificationService) NotifyUrgentAnnouncement(title string, userIDs []uint) error {
	notifTitle := "Annonce urgente"
	message := title
	icon := "mdi:alert"

	return s.createNotificationForUsers(userIDs, "announcement", "urgent_announcement", notifTitle, message, icon, "#EF4444", "/dashboard", 2)
}

// NotifyAnnouncementExpiring crée une notification pour une annonce expirant bientôt
func (s *NotificationService) NotifyAnnouncementExpiring(title string, userIDs []uint) error {
	notifTitle := "Annonce expirant bientôt"
	message := fmt.Sprintf("L'annonce '%s' expire bientôt", title)
	icon := "mdi:clock-alert"

	return s.createNotificationForUsers(userIDs, "announcement", "expiring_announcement", notifTitle, message, icon, "#F59E0B", "/dashboard", 1)
}

// NotifyNewEvent crée une notification pour un nouvel événement
func (s *NotificationService) NotifyNewEvent(title, slug string, userIDs []uint) error {
	notifTitle := "Nouvel événement"
	message := title
	icon := "mdi:calendar-plus"
	actionURL := fmt.Sprintf("/events/%s", slug)

	return s.createNotificationForUsers(userIDs, "event", "new_event", notifTitle, message, icon, "#8B5CF6", actionURL, 0)
}

// NotifyEventModified crée une notification pour un événement modifié
func (s *NotificationService) NotifyEventModified(title, slug string, userIDs []uint) error {
	notifTitle := "Événement modifié"
	message := fmt.Sprintf("L'événement '%s' a été modifié", title)
	icon := "mdi:calendar-edit"
	actionURL := fmt.Sprintf("/events/%s", slug)

	return s.createNotificationForUsers(userIDs, "event", "event_modified", notifTitle, message, icon, "#F59E0B", actionURL, 1)
}

// NotifyEventReminder crée une notification de rappel d'événement
func (s *NotificationService) NotifyEventReminder(title, slug string, reminderType string, userIDs []uint) error {
	notifTitle := "Rappel d'événement"
	message := fmt.Sprintf("Rappel: '%s' - %s", title, reminderType)
	icon := "mdi:calendar-clock"
	actionURL := fmt.Sprintf("/events/%s", slug)

	return s.createNotificationForUsers(userIDs, "event", "event_reminder", notifTitle, message, icon, "#8B5CF6", actionURL, 1)
}

// NotifyNewComment crée une notification pour un nouveau commentaire
func (s *NotificationService) NotifyNewComment(userID uint, entityType, entityTitle string, authorName string, actionURL string) error {
	title := "Nouveau commentaire"
	message := fmt.Sprintf("%s a commenté '%s'", authorName, entityTitle)
	icon := "mdi:comment"

	return s.createNotification(userID, "comment", "new_comment", title, message, icon, "#06B6D4", actionURL, 0)
}

// NotifyReaction crée une notification pour une réaction
func (s *NotificationService) NotifyReaction(userID uint, reactionType, entityTitle, userName string, actionURL string) error {
	title := "Nouvelle réaction"
	message := fmt.Sprintf("%s a réagi %s à '%s'", userName, reactionType, entityTitle)
	icon := "mdi:heart"

	return s.createNotification(userID, "reaction", "new_reaction", title, message, icon, "#EC4899", actionURL, 0)
}

// NotifyNewPoll crée une notification pour un nouveau sondage
func (s *NotificationService) NotifyNewPoll(title string, pollID uint, userIDs []uint) error {
	notifTitle := "Nouveau sondage"
	message := fmt.Sprintf("Donnez votre avis: '%s'", title)
	icon := "mdi:poll"
	actionURL := fmt.Sprintf("/polls/%d", pollID)

	return s.createNotificationForUsers(userIDs, "poll", "new_poll", notifTitle, message, icon, "#8B5CF6", actionURL, 0)
}

// NotifyPollClosing crée une notification pour un sondage qui va bientôt se fermer
func (s *NotificationService) NotifyPollClosing(title string, pollID uint, userIDs []uint) error {
	notifTitle := "Sondage bientôt fermé"
	message := fmt.Sprintf("Dernière chance de voter: '%s'", title)
	icon := "mdi:poll-clock"
	actionURL := fmt.Sprintf("/polls/%d", pollID)

	return s.createNotificationForUsers(userIDs, "poll", "poll_closing", notifTitle, message, icon, "#F59E0B", actionURL, 1)
}

// NotifyPollClosed crée une notification pour un sondage fermé avec résultats
func (s *NotificationService) NotifyPollClosed(title string, pollID uint, userIDs []uint) error {
	notifTitle := "Résultats du sondage disponibles"
	message := fmt.Sprintf("Découvrez les résultats de: '%s'", title)
	icon := "mdi:poll-check"
	actionURL := fmt.Sprintf("/polls/%d", pollID)

	return s.createNotificationForUsers(userIDs, "poll", "poll_closed", notifTitle, message, icon, "#10B981", actionURL, 0)
}

// createNotification crée une notification pour un utilisateur
func (s *NotificationService) createNotification(userID uint, notifType, category, title, message, icon, color, actionURL string, priority int) error {
	notification := models.Notification{
		UserID:    userID,
		Type:      notifType,
		Category:  category,
		Title:     title,
		Message:   message,
		Icon:      icon,
		Color:     color,
		Priority:  priority,
		ActionURL: actionURL,
		Metadata:  "{}",
	}

	return s.db.Create(&notification).Error
}

// createNotificationForUsers crée une notification pour plusieurs utilisateurs
func (s *NotificationService) createNotificationForUsers(userIDs []uint, notifType, category, title, message, icon, color, actionURL string, priority int) error {
	notifications := make([]models.Notification, 0, len(userIDs))

	for _, userID := range userIDs {
		notifications = append(notifications, models.Notification{
			UserID:    userID,
			Type:      notifType,
			Category:  category,
			Title:     title,
			Message:   message,
			Icon:      icon,
			Color:     color,
			Priority:  priority,
			ActionURL: actionURL,
			Metadata:  "{}",
		})
	}

	return s.db.Create(&notifications).Error
}
