package models

import (
	"time"
)

// SMTPConfig stocke la configuration du serveur SMTP
type SMTPConfig struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Host            string     `json:"host" gorm:"not null"`
	Port            int        `json:"port" gorm:"default:587"`
	Username        string     `json:"username"`
	Password        string     `json:"-" gorm:"type:text"` // Chiffré, jamais exposé en JSON
	FromEmail       string     `json:"from_email" gorm:"not null"`
	FromName        string     `json:"from_name" gorm:"default:'Airboard'"`
	UseTLS          bool       `json:"use_tls" gorm:"default:true"`
	UseSTARTTLS     bool       `json:"use_starttls" gorm:"default:true"`
	IsEnabled       bool       `json:"is_enabled" gorm:"default:false"`
	LastTestedAt    *time.Time `json:"last_tested_at"`
	LastTestSuccess bool       `json:"last_test_success" gorm:"default:false"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// EmailTemplate stocke les templates d'email personnalisables
type EmailTemplate struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Type          string    `json:"type" gorm:"uniqueIndex;not null"` // news, application, event, announcement
	Name          string    `json:"name" gorm:"not null"`
	Subject       string    `json:"subject" gorm:"not null"`
	HTMLBody      string    `json:"html_body" gorm:"type:text;not null"`
	PlainTextBody string    `json:"plain_text_body" gorm:"type:text"`
	IsEnabled     bool      `json:"is_enabled" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// EmailNotificationLog trace les emails envoyés
type EmailNotificationLog struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	TemplateType   string     `json:"template_type" gorm:"index"`
	ContentID      uint       `json:"content_id"`
	ContentTitle   string     `json:"content_title"`
	RecipientCount int        `json:"recipient_count"`
	SuccessCount   int        `json:"success_count"`
	FailureCount   int        `json:"failure_count"`
	Status         string     `json:"status" gorm:"default:'pending'"` // pending, sending, completed, failed
	ErrorMessage   string     `json:"error_message" gorm:"type:text"`
	SentAt         *time.Time `json:"sent_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// SMTPConfigRequest pour la création/modification de la config SMTP
type SMTPConfigRequest struct {
	Host        string `json:"host" binding:"required"`
	Port        int    `json:"port" binding:"required,min=1,max=65535"`
	Username    string `json:"username"`
	Password    string `json:"password"` // Vide = garder l'existant
	FromEmail   string `json:"from_email" binding:"required,email"`
	FromName    string `json:"from_name"`
	UseTLS      bool   `json:"use_tls"`
	UseSTARTTLS bool   `json:"use_starttls"`
	IsEnabled   bool   `json:"is_enabled"`
}

// EmailTemplateRequest pour la modification d'un template
type EmailTemplateRequest struct {
	Subject       string `json:"subject" binding:"required"`
	HTMLBody      string `json:"html_body" binding:"required"`
	PlainTextBody string `json:"plain_text_body"`
	IsEnabled     bool   `json:"is_enabled"`
}

// TestEmailRequest pour l'envoi d'un email de test
type TestEmailRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}

// GetDefaultEmailTemplates retourne les templates par défaut
func GetDefaultEmailTemplates() []EmailTemplate {
	return []EmailTemplate{
		{
			Type:      "news",
			Name:      "Notification Nouvel Article",
			Subject:   "{{.AppName}} - Nouvel article : {{.Title}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.summary { background: #f8fafc; border-left: 4px solid #3B82F6; padding: 15px; margin: 20px 0; border-radius: 0 8px 8px 0; }
.meta { color: #6b7280; font-size: 14px; margin: 15px 0; }
.meta span { margin-right: 20px; }
.button { display: inline-block; padding: 12px 24px; background: #3B82F6; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #2563EB; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>{{.Title}}</h2>
<div class="summary">{{.Summary}}</div>
<div class="meta">
<span>Par <strong>{{.Author}}</strong></span>
<span>Publié le {{.PublishedAt}}</span>
</div>
<a href="{{.Link}}" class="button">Lire l'article</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous êtes membre d'un groupe concerné par cet article.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "application",
			Name:      "Notification Nouvelle Application",
			Subject:   "{{.AppName}} - Nouvelle application : {{.Name}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #10B981 0%, #059669 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.app-card { background: #f0fdf4; border: 1px solid #bbf7d0; border-radius: 12px; padding: 20px; margin: 20px 0; }
.app-card h3 { color: #166534; margin: 0 0 10px 0; }
.app-card p { color: #15803d; margin: 0; }
.category { display: inline-block; background: #dcfce7; color: #166534; padding: 4px 12px; border-radius: 20px; font-size: 13px; margin-top: 10px; }
.button { display: inline-block; padding: 12px 24px; background: #10B981; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #059669; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>Nouvelle Application Disponible</h2>
<div class="app-card">
<h3>{{.Name}}</h3>
<p>{{.Description}}</p>
<span class="category">{{.AppGroup}}</span>
</div>
<a href="{{.URL}}" class="button">Ouvrir l'application</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous avez accès à cette catégorie d'applications.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "event",
			Name:      "Notification Nouvel Événement",
			Subject:   "{{.AppName}} - Nouvel événement : {{.Title}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.event-details { background: #fffbeb; border: 1px solid #fcd34d; border-radius: 12px; padding: 20px; margin: 20px 0; }
.event-details .detail { display: flex; align-items: center; margin: 10px 0; }
.event-details .icon { width: 24px; margin-right: 10px; color: #d97706; }
.event-details .label { color: #92400e; font-weight: 500; }
.event-details .value { color: #78350f; margin-left: 5px; }
.description { color: #4b5563; margin: 20px 0; }
.button { display: inline-block; padding: 12px 24px; background: #F59E0B; color: white; text-decoration: none; border-radius: 8px; font-weight: 500; margin-top: 20px; }
.button:hover { background: #D97706; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<h2>{{.Title}}</h2>
<div class="event-details">
<div class="detail">
<span class="label">Date de début :</span>
<span class="value">{{.StartDate}}</span>
</div>
{{if .EndDate}}
<div class="detail">
<span class="label">Date de fin :</span>
<span class="value">{{.EndDate}}</span>
</div>
{{end}}
{{if .Location}}
<div class="detail">
<span class="label">Lieu :</span>
<span class="value">{{.Location}}</span>
</div>
{{end}}
</div>
<p class="description">{{.Description}}</p>
<a href="{{.Link}}" class="button">Voir les détails</a>
</div>
<div class="footer">
<p>Vous recevez cet email car vous êtes membre d'un groupe concerné par cet événement.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
		{
			Type:      "announcement",
			Name:      "Notification Annonce",
			Subject:   "{{.AppName}} - {{.Type}} : {{.Title}}",
			IsEnabled: true,
			HTMLBody: `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f5f5f5; }
.container { max-width: 600px; margin: 0 auto; background: white; }
.header { background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%); color: white; padding: 30px; text-align: center; }
.header h1 { margin: 0; font-size: 24px; font-weight: 600; }
.header.info { background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%); }
.header.warning { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
.header.success { background: linear-gradient(135deg, #10B981 0%, #059669 100%); }
.content { padding: 30px; }
.content h2 { color: #1f2937; margin-top: 0; font-size: 22px; }
.announcement { background: #fef2f2; border-left: 4px solid #EF4444; padding: 20px; margin: 20px 0; border-radius: 0 8px 8px 0; }
.announcement.info { background: #eff6ff; border-left-color: #3B82F6; }
.announcement.warning { background: #fffbeb; border-left-color: #F59E0B; }
.announcement.success { background: #f0fdf4; border-left-color: #10B981; }
.type-badge { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: 600; text-transform: uppercase; margin-bottom: 15px; }
.type-badge.error { background: #fee2e2; color: #991b1b; }
.type-badge.info { background: #dbeafe; color: #1e40af; }
.type-badge.warning { background: #fef3c7; color: #92400e; }
.type-badge.success { background: #dcfce7; color: #166534; }
.footer { background: #f8fafc; padding: 20px; text-align: center; color: #6b7280; font-size: 12px; }
</style>
</head>
<body>
<div class="container">
<div class="header {{.Type}}">
<h1>{{.AppName}}</h1>
</div>
<div class="content">
<span class="type-badge {{.Type}}">{{.Type}}</span>
<h2>{{.Title}}</h2>
<div class="announcement {{.Type}}">
{{.Content}}
</div>
</div>
<div class="footer">
<p>Vous recevez cet email car cette annonce concerne tous les utilisateurs.</p>
<p>© {{.AppName}}</p>
</div>
</div>
</body>
</html>`,
		},
	}
}
