package models

import (
	"time"

	"gorm.io/gorm"
)

// Comment représente un commentaire sur un article, une application ou un événement
type Comment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Content   string         `json:"content" gorm:"type:text;not null"`
	UserID    uint           `json:"user_id"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`

	// Polymorphic relation - can be attached to News, Application, or Event
	EntityType string `json:"entity_type" gorm:"not null"` // "news", "application", "event"
	EntityID   uint   `json:"entity_id" gorm:"not null"`

	// Moderation
	IsApproved   bool       `json:"is_approved" gorm:"default:true"` // Auto-approved by default
	IsFlagged    bool       `json:"is_flagged" gorm:"default:false"` // Flagged for review
	ModeratedBy  *uint      `json:"moderated_by"`
	Moderator    *User      `json:"moderator,omitempty" gorm:"foreignKey:ModeratedBy"`
	ModeratedAt  *time.Time `json:"moderated_at"`

	// Parent comment for threaded replies (optional - for future)
	ParentID     *uint      `json:"parent_id"`
	Parent       *Comment   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`

	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// Feedback représente un feedback simple (👍/👎) sur un article, app ou événement
type Feedback struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	UserID     uint   `json:"user_id"`
	User       User   `json:"user" gorm:"foreignKey:UserID"`

	// Polymorphic relation
	EntityType string `json:"entity_type" gorm:"not null"` // "news", "application", "event"
	EntityID   uint   `json:"entity_id" gorm:"not null"`

	// Feedback type: positive (thumbs up) or negative (thumbs down)
	FeedbackType string    `json:"feedback_type" gorm:"not null"` // "positive", "negative"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Contrainte unique composite pour éviter les doublons (un user = un feedback par entité)
	// Géré par l'index unique en DB
}

// CommentSettings représente les paramètres de configuration pour les commentaires
type CommentSettings struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	CommentsEnabled       bool      `json:"comments_enabled" gorm:"default:true"`
	NewsCommentsEnabled   bool      `json:"news_comments_enabled" gorm:"default:true"`
	AppCommentsEnabled    bool      `json:"app_comments_enabled" gorm:"default:false"`
	EventCommentsEnabled  bool      `json:"event_comments_enabled" gorm:"default:true"`
	RequireModeration     bool      `json:"require_moderation" gorm:"default:false"` // Si true, commentaires nécessitent approbation
	AllowAnonymous        bool      `json:"allow_anonymous" gorm:"default:false"`
	MaxCommentLength      int       `json:"max_comment_length" gorm:"default:1000"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Request/Response structures

// CommentRequest pour la création/modification de commentaires
type CommentRequest struct {
	Content    string `json:"content" binding:"required,min=1,max=1000"`
	EntityType string `json:"entity_type" binding:"required,oneof=news application event"`
	EntityID   uint   `json:"entity_id" binding:"required"`
	ParentID   *uint  `json:"parent_id"` // Pour les réponses (future feature)
}

// FeedbackRequest pour ajouter un feedback
type FeedbackRequest struct {
	EntityType   string `json:"entity_type" binding:"required,oneof=news application event"`
	EntityID     uint   `json:"entity_id" binding:"required"`
	FeedbackType string `json:"feedback_type" binding:"required,oneof=positive negative"`
}

// CommentWithReplies pour afficher les commentaires avec leurs réponses
type CommentWithReplies struct {
	Comment
	Replies []Comment `json:"replies,omitempty"`
}

// CommentsListResponse pour les listes de commentaires
type CommentsListResponse struct {
	Comments   []CommentWithReplies `json:"comments"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// FeedbackStats pour les statistiques de feedback
type FeedbackStats struct {
	EntityType    string `json:"entity_type"`
	EntityID      uint   `json:"entity_id"`
	PositiveCount int64  `json:"positive_count"`
	NegativeCount int64  `json:"negative_count"`
	TotalCount    int64  `json:"total_count"`
	UserFeedback  string `json:"user_feedback,omitempty"` // "positive", "negative", ou vide
}

// ModerationRequest pour modérer un commentaire
type ModerationRequest struct {
	CommentID  uint `json:"comment_id" binding:"required"`
	IsApproved bool `json:"is_approved"`
	IsFlagged  bool `json:"is_flagged"`
}

// CommentSettingsRequest pour mettre à jour les paramètres
type CommentSettingsRequest struct {
	CommentsEnabled      *bool `json:"comments_enabled"`
	NewsCommentsEnabled  *bool `json:"news_comments_enabled"`
	AppCommentsEnabled   *bool `json:"app_comments_enabled"`
	EventCommentsEnabled *bool `json:"event_comments_enabled"`
	RequireModeration    *bool `json:"require_moderation"`
	AllowAnonymous       *bool `json:"allow_anonymous"`
	MaxCommentLength     *int  `json:"max_comment_length"`
}

// TableName spécifie le nom de la table pour Comment
func (Comment) TableName() string {
	return "comments"
}

// TableName spécifie le nom de la table pour Feedback
func (Feedback) TableName() string {
	return "feedbacks"
}

// TableName spécifie le nom de la table pour CommentSettings
func (CommentSettings) TableName() string {
	return "comment_settings"
}
