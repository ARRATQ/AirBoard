package handlers

import (
	"log"
	"net/http"
	"time"

	"airboard/config"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EmailHandler gère les endpoints de configuration email
type EmailHandler struct {
	db           *gorm.DB
	emailService *services.EmailService
	config       *config.Config
}

// NewEmailHandler crée une nouvelle instance du handler email
func NewEmailHandler(db *gorm.DB, cfg *config.Config) *EmailHandler {
	return &EmailHandler{
		db:           db,
		emailService: services.NewEmailService(db, cfg),
		config:       cfg,
	}
}

// GetSMTPConfig retourne la configuration SMTP actuelle (mot de passe masqué)
func (h *EmailHandler) GetSMTPConfig(c *gin.Context) {
	var config models.SMTPConfig

	result := h.db.First(&config)
	if result.Error == gorm.ErrRecordNotFound {
		// Retourner une config vide
		c.JSON(http.StatusOK, models.SMTPConfig{
			Port:        587,
			UseTLS:      false,
			UseSTARTTLS: true,
			FromName:    "Airboard",
		})
		return
	} else if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de la configuration SMTP"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateSMTPConfig crée ou met à jour la configuration SMTP
func (h *EmailHandler) UpdateSMTPConfig(c *gin.Context) {
	var req models.SMTPConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var smtpConfig models.SMTPConfig
	h.db.First(&smtpConfig)

	smtpConfig.Host = req.Host
	smtpConfig.Port = req.Port
	smtpConfig.Username = req.Username
	smtpConfig.FromEmail = req.FromEmail
	smtpConfig.FromName = req.FromName
	smtpConfig.UseTLS = req.UseTLS
	smtpConfig.UseSTARTTLS = req.UseSTARTTLS
	smtpConfig.IsEnabled = req.IsEnabled

	// Mettre à jour le mot de passe uniquement s'il est fourni
	if req.Password != "" {
		encryptedPassword, err := h.emailService.EncryptPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors du chiffrement du mot de passe"})
			return
		}
		smtpConfig.Password = encryptedPassword
	}

	if smtpConfig.ID == 0 {
		if err := h.db.Create(&smtpConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de la configuration SMTP"})
			return
		}
	} else {
		if err := h.db.Save(&smtpConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour de la configuration SMTP"})
			return
		}
	}

	c.JSON(http.StatusOK, smtpConfig)
}

// TestSMTPConfig envoie un email de test
func (h *EmailHandler) TestSMTPConfig(c *gin.Context) {
	var req models.TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Email] Erreur binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
	}

	log.Printf("[Email] Test SMTP demandé pour: %s", req.ToEmail)

	var smtpConfig models.SMTPConfig
	if err := h.db.First(&smtpConfig).Error; err != nil {
		log.Printf("[Email] Configuration SMTP non trouvée: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Configuration SMTP non trouvée. Veuillez d'abord configurer le serveur SMTP."})
		return
	}

	if err := h.emailService.TestConnection(&smtpConfig, req.ToEmail); err != nil {
		now := time.Now()
		smtpConfig.LastTestedAt = &now
		smtpConfig.LastTestSuccess = false
		h.db.Save(&smtpConfig)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Test échoué: " + err.Error()})
		return
	}

	now := time.Now()
	smtpConfig.LastTestedAt = &now
	smtpConfig.LastTestSuccess = true
	h.db.Save(&smtpConfig)

	c.JSON(http.StatusOK, gin.H{"message": "Email de test envoyé avec succès"})
}

// GetEmailTemplates retourne tous les templates d'email
func (h *EmailHandler) GetEmailTemplates(c *gin.Context) {
	var templates []models.EmailTemplate

	if err := h.db.Order("type").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des templates"})
		return
	}

	// Si aucun template, créer les templates par défaut
	if len(templates) == 0 {
		defaultTemplates := models.GetDefaultEmailTemplates()
		for _, t := range defaultTemplates {
			h.db.Create(&t)
		}
		h.db.Order("type").Find(&templates)
	}

	c.JSON(http.StatusOK, templates)
}

// GetEmailTemplate retourne un template spécifique par type
func (h *EmailHandler) GetEmailTemplate(c *gin.Context) {
	templateType := c.Param("type")

	var template models.EmailTemplate
	if err := h.db.Where("type = ?", templateType).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template non trouvé"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateEmailTemplate met à jour un template d'email
func (h *EmailHandler) UpdateEmailTemplate(c *gin.Context) {
	templateType := c.Param("type")

	var req models.EmailTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var template models.EmailTemplate
	if err := h.db.Where("type = ?", templateType).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template non trouvé"})
		return
	}

	template.Subject = req.Subject
	template.HTMLBody = req.HTMLBody
	template.PlainTextBody = req.PlainTextBody
	template.IsEnabled = req.IsEnabled

	if err := h.db.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour du template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ResetEmailTemplate réinitialise un template à sa valeur par défaut
func (h *EmailHandler) ResetEmailTemplate(c *gin.Context) {
	templateType := c.Param("type")

	// Trouver le template par défaut
	defaultTemplates := models.GetDefaultEmailTemplates()
	var defaultTemplate *models.EmailTemplate
	for _, t := range defaultTemplates {
		if t.Type == templateType {
			defaultTemplate = &t
			break
		}
	}

	if defaultTemplate == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Type de template inconnu"})
		return
	}

	var template models.EmailTemplate
	if err := h.db.Where("type = ?", templateType).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template non trouvé"})
		return
	}

	template.Subject = defaultTemplate.Subject
	template.HTMLBody = defaultTemplate.HTMLBody
	template.PlainTextBody = defaultTemplate.PlainTextBody
	template.Name = defaultTemplate.Name

	if err := h.db.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la réinitialisation du template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// PreviewTemplate retourne un aperçu du template avec des données exemple
func (h *EmailHandler) PreviewTemplate(c *gin.Context) {
	templateType := c.Param("type")

	var template models.EmailTemplate
	if err := h.db.Where("type = ?", templateType).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template non trouvé"})
		return
	}

	// Récupérer les données exemple
	sampleData := h.emailService.GetSampleData(templateType)
	if sampleData == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type de template inconnu"})
		return
	}

	// Exécuter les templates
	subject, err := h.emailService.ExecuteTemplate(template.Subject, sampleData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur dans le sujet: " + err.Error()})
		return
	}

	body, err := h.emailService.ExecuteTemplate(template.HTMLBody, sampleData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur dans le corps: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subject": subject,
		"body":    body,
	})
}

// GetEmailLogs retourne l'historique des notifications envoyées
func (h *EmailHandler) GetEmailLogs(c *gin.Context) {
	var logs []models.EmailNotificationLog

	if err := h.db.Order("created_at DESC").Limit(100).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetTemplateVariables retourne les variables disponibles pour chaque type de template
func (h *EmailHandler) GetTemplateVariables(c *gin.Context) {
	variables := map[string][]map[string]string{
		"news": {
			{"name": "{{.Title}}", "description": "Titre de l'article"},
			{"name": "{{.Summary}}", "description": "Résumé de l'article"},
			{"name": "{{.Author}}", "description": "Nom de l'auteur"},
			{"name": "{{.Link}}", "description": "Lien vers l'article"},
			{"name": "{{.AppName}}", "description": "Nom de l'application"},
			{"name": "{{.PublishedAt}}", "description": "Date de publication"},
		},
		"application": {
			{"name": "{{.Name}}", "description": "Nom de l'application"},
			{"name": "{{.Description}}", "description": "Description de l'application"},
			{"name": "{{.URL}}", "description": "URL de l'application"},
			{"name": "{{.AppGroup}}", "description": "Catégorie de l'application"},
			{"name": "{{.AppName}}", "description": "Nom de l'application Airboard"},
		},
		"event": {
			{"name": "{{.Title}}", "description": "Titre de l'événement"},
			{"name": "{{.Description}}", "description": "Description de l'événement"},
			{"name": "{{.StartDate}}", "description": "Date de début"},
			{"name": "{{.EndDate}}", "description": "Date de fin"},
			{"name": "{{.Location}}", "description": "Lieu de l'événement"},
			{"name": "{{.Link}}", "description": "Lien vers l'événement"},
			{"name": "{{.AppName}}", "description": "Nom de l'application"},
		},
		"announcement": {
			{"name": "{{.Title}}", "description": "Titre de l'annonce"},
			{"name": "{{.Content}}", "description": "Contenu de l'annonce"},
			{"name": "{{.Type}}", "description": "Type d'annonce (info, warning, success, error)"},
			{"name": "{{.AppName}}", "description": "Nom de l'application"},
		},
	}

	c.JSON(http.StatusOK, variables)
}
