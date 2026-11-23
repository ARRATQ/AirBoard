package models

import (
	"time"
)

// SMTPConfig stocke la configuration email (OAuth 2.0 uniquement)
// Note: Les champs SMTP legacy (Host, Port, Username, Password, UseTLS, UseSTARTTLS) sont conservés
// pour la compatibilité DB mais ne sont plus utilisés. L'envoi se fait via Microsoft Graph API.
type SMTPConfig struct {
	ID              uint       `json:"id" gorm:"primaryKey"`
	Host            string     `json:"host,omitempty"`                    // Obsolète - conservé pour compatibilité
	Port            int        `json:"port,omitempty" gorm:"default:587"` // Obsolète - conservé pour compatibilité
	Username        string     `json:"username,omitempty"`                // Obsolète - conservé pour compatibilité
	Password        string     `json:"-" gorm:"type:text"`                // Obsolète - conservé pour compatibilité
	FromEmail       string     `json:"from_email" gorm:"not null"`
	FromName        string     `json:"from_name" gorm:"default:'Airboard'"`
	UseTLS          bool       `json:"use_tls,omitempty"`             // Obsolète - conservé pour compatibilité
	UseSTARTTLS     bool       `json:"use_starttls,omitempty"`        // Obsolète - conservé pour compatibilité
	UseOAuth        bool       `json:"use_oauth" gorm:"default:true"` // Toujours true maintenant
	IsEnabled       bool       `json:"is_enabled" gorm:"default:false"`
	LastTestedAt    *time.Time `json:"last_tested_at"`
	LastTestSuccess bool       `json:"last_test_success" gorm:"default:false"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// Relations
	EmailOAuthConfig *EmailOAuthConfig `json:"oauth_config,omitempty" gorm:"foreignKey:SMTPConfigID"`
}

// EmailOAuthConfig stocke la configuration OAuth 2.0 pour l'authentification SMTP
type EmailOAuthConfig struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	SMTPConfigID uint   `json:"smtp_config_id" gorm:"uniqueIndex"`   // One-to-one with SMTPConfig
	Provider     string `json:"provider" gorm:"default:'microsoft'"` // microsoft, google (future)
	TenantID     string `json:"tenant_id"`                           // Azure AD tenant ID
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"-" gorm:"type:text"` // Encrypted, never exposed
	Scopes       string `json:"scopes" gorm:"default:'https://graph.microsoft.com/.default'"`
	AuthURL      string `json:"auth_url" gorm:"default:'https://login.microsoftonline.com/{tenant}/oauth2/v2.0/authorize'"`
	TokenURL     string `json:"token_url" gorm:"default:'https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token'"`

	// Token storage (all encrypted)
	AccessToken  string     `json:"-" gorm:"type:text"`                 // Current access token (encrypted)
	RefreshToken string     `json:"-" gorm:"type:text"`                 // Refresh token (encrypted)
	TokenType    string     `json:"token_type" gorm:"default:'Bearer'"` // Usually "Bearer"
	ExpiresAt    *time.Time `json:"expires_at"`                         // Token expiration time

	// Flow type: "client_credentials" or "refresh_token"
	GrantType string `json:"grant_type" gorm:"default:'client_credentials'"`

	// Status tracking
	IsEnabled        bool       `json:"is_enabled" gorm:"default:false"`
	LastTokenRefresh *time.Time `json:"last_token_refresh"`
	LastRefreshError string     `json:"last_refresh_error" gorm:"type:text"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relation
	SMTPConfig *SMTPConfig `json:"smtp_config,omitempty" gorm:"foreignKey:SMTPConfigID"`
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

// SMTPConfigRequest pour la création/modification de la config email
// Simplifié pour OAuth uniquement
type SMTPConfigRequest struct {
	FromEmail string `json:"from_email" binding:"required,email"`
	FromName  string `json:"from_name"`
	IsEnabled bool   `json:"is_enabled"`
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

// EmailOAuthConfigRequest pour la création/modification de la config OAuth
type EmailOAuthConfigRequest struct {
	Provider     string `json:"provider" binding:"required,oneof=microsoft google"`
	TenantID     string `json:"tenant_id"`
	ClientID     string `json:"client_id" binding:"required"`
	ClientSecret string `json:"client_secret"` // Empty = keep existing
	Scopes       string `json:"scopes"`
	GrantType    string `json:"grant_type" binding:"required,oneof=client_credentials refresh_token"`
	RefreshToken string `json:"refresh_token"` // For refresh_token flow
	IsEnabled    bool   `json:"is_enabled"`
}

// EmailOAuthTestRequest pour tester la connexion OAuth
type EmailOAuthTestRequest struct {
	ToEmail string `json:"to_email" binding:"required,email"`
}
