package services

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"airboard/config"
	"airboard/models"

	"gorm.io/gorm"
)

// EmailService gère l'envoi des emails
type EmailService struct {
	db     *gorm.DB
	config *config.Config
}

// NewEmailService crée une nouvelle instance du service email
func NewEmailService(db *gorm.DB, cfg *config.Config) *EmailService {
	return &EmailService{
		db:     db,
		config: cfg,
	}
}

// NewsEmailData contient les données pour le template news
type NewsEmailData struct {
	Title       string
	Summary     string
	Author      string
	Link        string
	AppName     string
	PublishedAt string
}

// ApplicationEmailData contient les données pour le template application
type ApplicationEmailData struct {
	Name        string
	Description string
	URL         string
	AppGroup    string
	AppName     string
}

// EventEmailData contient les données pour le template event
type EventEmailData struct {
	Title       string
	Description string
	StartDate   string
	EndDate     string
	Location    string
	Link        string
	AppName     string
}

// AnnouncementEmailData contient les données pour le template announcement
type AnnouncementEmailData struct {
	Title   string
	Content string
	Type    string
	AppName string
}

// ChatMessageEmailData contient les données pour le template chat_message
type ChatMessageEmailData struct {
	SenderName     string
	MessagePreview string
	SentAt         string
	ChatLink       string
	AppName        string
}

// SendNotification envoie des notifications email aux groupes cibles
func (s *EmailService) SendNotification(templateType string, contentID uint, targetGroupIDs []uint) error {
	// Récupérer la config SMTP avec la config OAuth si disponible
	var smtpConfig models.SMTPConfig
	if err := s.db.Preload("EmailOAuthConfig").First(&smtpConfig).Error; err != nil {
		log.Printf("[Email] SMTP non configuré: %v", err)
		return fmt.Errorf("SMTP non configuré: %w", err)
	}
	if !smtpConfig.IsEnabled {
		log.Println("[Email] SMTP désactivé, notification ignorée")
		return nil
	}

	// Vérifier que OAuth est configuré
	if smtpConfig.EmailOAuthConfig == nil || !smtpConfig.EmailOAuthConfig.IsEnabled {
		log.Printf("[Email] OAuth non configuré ou désactivé")
		return fmt.Errorf("OAuth non configuré. Veuillez configurer OAuth 2.0 dans les paramètres email")
	}
	log.Printf("[Email] Mode OAuth 2.0 (provider: %s, grant: %s)",
		smtpConfig.EmailOAuthConfig.Provider, smtpConfig.EmailOAuthConfig.GrantType)

	// Récupérer le template
	var emailTemplate models.EmailTemplate
	if err := s.db.Where("type = ? AND is_enabled = ?", templateType, true).First(&emailTemplate).Error; err != nil {
		log.Printf("[Email] Template '%s' non trouvé ou désactivé: %v", templateType, err)
		return fmt.Errorf("template non trouvé ou désactivé: %w", err)
	}

	// Récupérer les destinataires selon les groupes cibles
	var recipients []string
	if len(targetGroupIDs) == 0 {
		// Notification globale - envoyer à tous les utilisateurs actifs
		s.db.Model(&models.User{}).Where("is_active = ? AND email != '' AND email IS NOT NULL", true).Pluck("email", &recipients)
	} else {
		// Notification ciblée par groupe
		s.db.Table("users").
			Select("DISTINCT users.email").
			Joins("JOIN user_groups ON user_groups.user_id = users.id").
			Where("user_groups.group_id IN ? AND users.is_active = ? AND users.email != '' AND users.email IS NOT NULL", targetGroupIDs, true).
			Pluck("email", &recipients)
	}

	if len(recipients) == 0 {
		log.Println("[Email] Aucun destinataire trouvé, notification ignorée")
		return nil
	}

	log.Printf("[Email] Envoi de %d notifications de type '%s'", len(recipients), templateType)

	// Créer l'entrée de log
	now := time.Now()
	notifLog := models.EmailNotificationLog{
		TemplateType:   templateType,
		ContentID:      contentID,
		RecipientCount: len(recipients),
		Status:         "sending",
		SentAt:         &now,
	}
	s.db.Create(&notifLog)

	// Préparer les données du template selon le type
	emailData, contentTitle, err := s.prepareEmailData(templateType, contentID)
	if err != nil {
		notifLog.Status = "failed"
		notifLog.ErrorMessage = err.Error()
		s.db.Save(&notifLog)
		return err
	}
	notifLog.ContentTitle = contentTitle

	// Parser et exécuter les templates
	subject, err := s.ExecuteTemplate(emailTemplate.Subject, emailData)
	if err != nil {
		notifLog.Status = "failed"
		notifLog.ErrorMessage = "Erreur template sujet: " + err.Error()
		s.db.Save(&notifLog)
		return err
	}

	htmlBody, err := s.ExecuteTemplate(emailTemplate.HTMLBody, emailData)
	if err != nil {
		notifLog.Status = "failed"
		notifLog.ErrorMessage = "Erreur template corps: " + err.Error()
		s.db.Save(&notifLog)
		return err
	}

	// Envoyer les emails
	successCount := 0
	failureCount := 0
	var lastError string

	for _, recipient := range recipients {
		if err := s.sendEmail(&smtpConfig, recipient, subject, htmlBody); err != nil {
			log.Printf("[Email] Échec envoi à %s: %v", recipient, err)
			failureCount++
			lastError = err.Error()
		} else {
			successCount++
		}
	}

	// Mettre à jour le log
	notifLog.SuccessCount = successCount
	notifLog.FailureCount = failureCount
	completedAt := time.Now()
	notifLog.CompletedAt = &completedAt

	if failureCount > 0 && successCount == 0 {
		notifLog.Status = "failed"
		notifLog.ErrorMessage = lastError
	} else if failureCount > 0 {
		notifLog.Status = "completed"
		notifLog.ErrorMessage = fmt.Sprintf("%d échecs sur %d", failureCount, len(recipients))
	} else {
		notifLog.Status = "completed"
	}

	s.db.Save(&notifLog)

	log.Printf("[Email] Notification terminée: %d succès, %d échecs", successCount, failureCount)
	return nil
}

// SendChatNotificationEmail envoie un email à un utilisateur hors ligne pour l'informer d'un nouveau message chat
func (s *EmailService) SendChatNotificationEmail(senderID, recipientID uint) error {
	// Récupérer la config SMTP avec OAuth
	var smtpConfig models.SMTPConfig
	if err := s.db.Preload("EmailOAuthConfig").First(&smtpConfig).Error; err != nil {
		return fmt.Errorf("SMTP non configuré: %w", err)
	}
	if !smtpConfig.IsEnabled {
		return nil
	}
	if smtpConfig.EmailOAuthConfig == nil || !smtpConfig.EmailOAuthConfig.IsEnabled {
		return nil
	}

	// Récupérer le template chat_message
	var emailTemplate models.EmailTemplate
	if err := s.db.Where("type = ? AND is_enabled = ?", "chat_message", true).First(&emailTemplate).Error; err != nil {
		return nil // Template non trouvé ou désactivé
	}

	// Récupérer l'expéditeur
	var sender models.User
	if err := s.db.First(&sender, senderID).Error; err != nil {
		return fmt.Errorf("expéditeur non trouvé: %w", err)
	}

	// Récupérer le destinataire
	var recipient models.User
	if err := s.db.First(&recipient, recipientID).Error; err != nil {
		return fmt.Errorf("destinataire non trouvé: %w", err)
	}

	// Le destinataire doit avoir un email valide
	if recipient.Email == "" {
		return nil
	}

	// Récupérer le dernier message de l'expéditeur vers le destinataire
	var lastMessage models.ChatMessage
	if err := s.db.Where("sender_id = ? AND recipient_id = ?", senderID, recipientID).
		Order("created_at desc").First(&lastMessage).Error; err != nil {
		return fmt.Errorf("message non trouvé: %w", err)
	}

	// Nom d'affichage de l'expéditeur
	senderName := strings.TrimSpace(sender.FirstName + " " + sender.LastName)
	if senderName == "" {
		senderName = sender.Username
	}

	// Tronquer l'aperçu du message
	messagePreview := lastMessage.Content
	if len(messagePreview) > 100 {
		messagePreview = messagePreview[:100] + "..."
	}

	// Nom de l'application
	var appSettings models.AppSettings
	s.db.First(&appSettings)
	appName := appSettings.AppName
	if appName == "" {
		appName = "Airboard"
	}

	// Construire les données du template
	emailData := ChatMessageEmailData{
		SenderName:     senderName,
		MessagePreview: messagePreview,
		SentAt:         lastMessage.CreatedAt.Format("02/01/2006 à 15:04"),
		ChatLink:       fmt.Sprintf("%s/chat", s.config.Server.PublicURL),
		AppName:        appName,
	}

	// Exécuter les templates
	subject, err := s.ExecuteTemplate(emailTemplate.Subject, emailData)
	if err != nil {
		return fmt.Errorf("erreur template sujet: %w", err)
	}

	htmlBody, err := s.ExecuteTemplate(emailTemplate.HTMLBody, emailData)
	if err != nil {
		return fmt.Errorf("erreur template corps: %w", err)
	}

	// Envoyer l'email
	if err := s.sendEmail(&smtpConfig, recipient.Email, subject, htmlBody); err != nil {
		log.Printf("[Chat Email] Échec envoi à %s: %v", recipient.Email, err)
		return err
	}

	log.Printf("[Chat Email] Notification envoyée à %s (de %s)", recipient.Email, senderName)
	return nil
}

// prepareEmailData prépare les données selon le type de contenu
func (s *EmailService) prepareEmailData(templateType string, contentID uint) (interface{}, string, error) {
	// Récupérer le nom de l'application
	var appSettings models.AppSettings
	s.db.First(&appSettings)
	appName := appSettings.AppName
	if appName == "" {
		appName = "Airboard"
	}

	switch templateType {
	case "news":
		var news models.News
		if err := s.db.Preload("Author").First(&news, contentID).Error; err != nil {
			return nil, "", fmt.Errorf("article non trouvé: %w", err)
		}
		authorName := strings.TrimSpace(news.Author.FirstName + " " + news.Author.LastName)
		if authorName == "" {
			authorName = news.Author.Username
		}
		publishedAt := ""
		if news.PublishedAt != nil {
			publishedAt = news.PublishedAt.Format("02/01/2006 à 15:04")
		} else {
			publishedAt = time.Now().Format("02/01/2006 à 15:04")
		}
		return NewsEmailData{
			Title:       news.Title,
			Summary:     news.Summary,
			Author:      authorName,
			Link:        fmt.Sprintf("%s/news/%s", s.config.Server.PublicURL, news.Slug),
			AppName:     appName,
			PublishedAt: publishedAt,
		}, news.Title, nil

	case "application":
		var app models.Application
		if err := s.db.Preload("AppGroup").First(&app, contentID).Error; err != nil {
			return nil, "", fmt.Errorf("application non trouvée: %w", err)
		}
		appGroupName := ""
		if app.AppGroup != nil {
			appGroupName = app.AppGroup.Name
		}
		return ApplicationEmailData{
			Name:        app.Name,
			Description: app.Description,
			URL:         app.URL,
			AppGroup:    appGroupName,
			AppName:     appName,
		}, app.Name, nil

	case "event":
		var event models.Event
		if err := s.db.First(&event, contentID).Error; err != nil {
			return nil, "", fmt.Errorf("événement non trouvé: %w", err)
		}
		endDate := ""
		if event.EndDate != nil {
			endDate = event.EndDate.Format("02/01/2006 à 15:04")
		}
		return EventEmailData{
			Title:       event.Title,
			Description: event.Description,
			StartDate:   event.StartDate.Format("02/01/2006 à 15:04"),
			EndDate:     endDate,
			Location:    event.Location,
			Link:        fmt.Sprintf("%s/events/%s", s.config.Server.PublicURL, event.Slug),
			AppName:     appName,
		}, event.Title, nil

	case "announcement":
		var announcement models.Announcement
		if err := s.db.First(&announcement, contentID).Error; err != nil {
			return nil, "", fmt.Errorf("annonce non trouvée: %w", err)
		}
		return AnnouncementEmailData{
			Title:   announcement.Title,
			Content: announcement.Content,
			Type:    announcement.Type,
			AppName: appName,
		}, announcement.Title, nil
	}

	return nil, "", fmt.Errorf("type de template inconnu: %s", templateType)
}

// ExecuteTemplate exécute un template Go avec les données fournies
func (s *EmailService) ExecuteTemplate(templateStr string, data interface{}) (string, error) {
	tmpl, err := template.New("email").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("erreur parsing template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("erreur exécution template: %w", err)
	}
	return buf.String(), nil
}

// sendEmail envoie un email via OAuth 2.0 (Microsoft Graph API)
func (s *EmailService) sendEmail(config *models.SMTPConfig, to, subject, htmlBody string) error {
	// Vérifier que OAuth est configuré
	if config.EmailOAuthConfig == nil || !config.EmailOAuthConfig.IsEnabled {
		return fmt.Errorf("OAuth non configuré ou désactivé. Veuillez configurer OAuth 2.0 dans les paramètres email")
	}

	log.Printf("[Email] Envoi via OAuth 2.0 (Microsoft Graph API) à %s", to)
	return s.SendEmailWithOAuth(config, to, subject, htmlBody)
}


// SendEmailWithOAuth envoie un email via OAuth 2.0 Microsoft Graph API
func (s *EmailService) SendEmailWithOAuth(config *models.SMTPConfig, to, subject, htmlBody string) error {
	if config.EmailOAuthConfig == nil {
		return fmt.Errorf("OAuth config not found")
	}

	// Get valid access token (auto-refresh if needed)
	oauthService := NewEmailOAuthService(s.db, s.config)
	accessToken, err := oauthService.GetValidAccessToken(config.EmailOAuthConfig)
	if err != nil {
		return fmt.Errorf("failed to get OAuth token: %w", err)
	}

	log.Printf("[Email OAuth] Token acquis, envoi email à %s via Microsoft Graph API", to)
	return s.sendEmailViaGraphAPI(config, accessToken, to, subject, htmlBody)
}

// sendEmailViaGraphAPI envoie un email via Microsoft Graph API
// C'est la seule méthode supportée pour client credentials flow avec Microsoft 365
func (s *EmailService) sendEmailViaGraphAPI(config *models.SMTPConfig, accessToken, to, subject, htmlBody string) error {
	log.Printf("[Email OAuth] Envoi via Microsoft Graph API à %s", to)

	// Construire le payload JSON pour Graph API
	emailPayload := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": subject,
			"body": map[string]interface{}{
				"contentType": "HTML",
				"content":     htmlBody,
			},
			"from": map[string]interface{}{
				"emailAddress": map[string]interface{}{
					"address": config.FromEmail,
					"name":    config.FromName,
				},
			},
			"toRecipients": []map[string]interface{}{
				{
					"emailAddress": map[string]interface{}{
						"address": to,
					},
				},
			},
		},
		"saveToSentItems": "false",
	}

	jsonPayload, err := json.Marshal(emailPayload)
	if err != nil {
		return fmt.Errorf("erreur sérialisation JSON: %w", err)
	}

	// Appeler l'API Microsoft Graph
	// URL: https://graph.microsoft.com/v1.0/users/{from_email}/sendMail
	graphURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", config.FromEmail)

	req, err := http.NewRequest("POST", graphURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("erreur création requête: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("erreur appel Graph API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[Email OAuth] Graph API erreur: %s", string(body))
		return fmt.Errorf("Graph API erreur (HTTP %d): %s", resp.StatusCode, string(body))
	}

	log.Printf("[Email OAuth] Email envoyé via Graph API à %s", to)
	return nil
}



// TestConnection teste la connexion OAuth avec un email de test
func (s *EmailService) TestConnection(smtpConfig *models.SMTPConfig, testEmail string) error {
	log.Printf("[Email] TestConnection OAuth appelé - Destinataire: %s", testEmail)

	// Vérifier que OAuth est configuré
	if smtpConfig.EmailOAuthConfig == nil || !smtpConfig.EmailOAuthConfig.IsEnabled {
		return fmt.Errorf("OAuth non configuré ou désactivé. Veuillez configurer OAuth 2.0")
	}

	// Récupérer le nom de l'application
	var appSettings models.AppSettings
	if err := s.db.First(&appSettings).Error; err != nil {
		log.Printf("[Email] Impossible de récupérer les paramètres: %v", err)
	}
	appName := appSettings.AppName
	if appName == "" {
		appName = "Airboard"
	}

	testSubject := fmt.Sprintf("%s - Email de test OAuth", appName)
	testBody := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 20px auto; background: white; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
.header { background: linear-gradient(135deg, #10B981 0%%, #059669 100%%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; }
.content { padding: 30px; text-align: center; }
.success-icon { font-size: 48px; margin-bottom: 20px; }
.badge { display: inline-block; background: #dcfce7; color: #166534; padding: 8px 16px; border-radius: 8px; font-weight: 600; margin: 10px 0; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>%s</h1>
</div>
<div class="content">
<div class="success-icon">✅</div>
<h2>Configuration OAuth réussie !</h2>
<p>Cet email confirme que votre configuration OAuth 2.0 est correcte.</p>
<p>Les notifications email fonctionneront correctement via Microsoft Graph API.</p>
<div class="badge">🔐 OAuth 2.0 - Microsoft Graph API</div>
</div>
<div class="footer">
<p>Email de test envoyé le %s</p>
</div>
</div>
</body>
</html>`, appName, time.Now().Format("02/01/2006 à 15:04"))

	// Envoyer via OAuth (Microsoft Graph API)
	if err := s.SendEmailWithOAuth(smtpConfig, testEmail, testSubject, testBody); err != nil {
		log.Printf("[Email] Échec envoi email de test OAuth: %v", err)
		return err
	}

	log.Printf("[Email] Email de test OAuth envoyé avec succès à %s", testEmail)
	return nil
}

// EncryptPassword chiffre un mot de passe avec AES-256
func (s *EmailService) EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", nil
	}

	// Utiliser les 32 premiers octets du secret JWT comme clé
	secret := s.config.JWT.Secret
	if len(secret) < 32 {
		// Padding si le secret est trop court
		secret = secret + strings.Repeat("0", 32-len(secret))
	}
	key := []byte(secret[:32])

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("erreur création cipher: %w", err)
	}

	plaintext := []byte(password)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("erreur génération IV: %w", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword déchiffre un mot de passe chiffré avec AES-256
func (s *EmailService) DecryptPassword(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}

	// Utiliser les 32 premiers octets du secret JWT comme clé
	secret := s.config.JWT.Secret
	if len(secret) < 32 {
		// Padding si le secret est trop court
		secret = secret + strings.Repeat("0", 32-len(secret))
	}
	key := []byte(secret[:32])

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("erreur décodage base64: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("erreur création cipher: %w", err)
	}

	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext trop court")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// GetSampleData retourne des données exemple pour un type de template
func (s *EmailService) GetSampleData(templateType string) interface{} {
	// Récupérer le nom de l'application
	var appSettings models.AppSettings
	s.db.First(&appSettings)
	appName := appSettings.AppName
	if appName == "" {
		appName = "Airboard"
	}

	switch templateType {
	case "news":
		return NewsEmailData{
			Title:       "Exemple d'article",
			Summary:     "Ceci est un résumé exemple pour prévisualiser le template d'email.",
			Author:      "Jean Dupont",
			Link:        fmt.Sprintf("%s/news/exemple-article", s.config.Server.PublicURL),
			AppName:     appName,
			PublishedAt: time.Now().Format("02/01/2006 à 15:04"),
		}
	case "application":
		return ApplicationEmailData{
			Name:        "Application Exemple",
			Description: "Ceci est une description exemple pour prévisualiser le template.",
			URL:         "https://example.com/app",
			AppGroup:    "Développement",
			AppName:     appName,
		}
	case "event":
		return EventEmailData{
			Title:       "Événement Exemple",
			Description: "Ceci est une description exemple pour prévisualiser le template d'événement.",
			StartDate:   time.Now().Format("02/01/2006 à 15:04"),
			EndDate:     time.Now().Add(2 * time.Hour).Format("02/01/2006 à 15:04"),
			Location:    "Salle de conférence A",
			Link:        fmt.Sprintf("%s/events/exemple-evenement", s.config.Server.PublicURL),
			AppName:     appName,
		}
	case "announcement":
		return AnnouncementEmailData{
			Title:   "Annonce Exemple",
			Content: "Ceci est un contenu d'annonce exemple pour prévisualiser le template.",
			Type:    "info",
			AppName: appName,
		}
	case "chat_message":
		return ChatMessageEmailData{
			SenderName:     "Jean Dupont",
			MessagePreview: "Bonjour, avez-vous vu le dernier rapport ? Je l'ai partagé dans le dossier commun...",
			SentAt:         time.Now().Format("02/01/2006 à 15:04"),
			ChatLink:       fmt.Sprintf("%s/chat", s.config.Server.PublicURL),
			AppName:        appName,
		}
	}
	return nil
}
