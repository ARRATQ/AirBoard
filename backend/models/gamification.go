package models

import (
	"time"

	"gorm.io/gorm"
)

// GamificationProfile stocke les informations de progression d'un utilisateur
type GamificationProfile struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	UserID         uint           `json:"user_id" gorm:"uniqueIndex"`
	XP             int64          `json:"xp" gorm:"default:0"`
	Level          int            `json:"level" gorm:"default:1"`
	LastDailyLogin *time.Time     `json:"last_login_reward"`
	LoginStreak    int            `json:"login_streak" gorm:"default:0"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User User `json:"user,omitempty"`
}

// Achievement représente un badge à débloquer
type Achievement struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Code          string         `json:"code" gorm:"uniqueIndex"` // ex: "early_bird", "explorer"
	Name          string         `json:"name" gorm:"not null"`
	Description   string         `json:"description"`
	Icon          string         `json:"icon" gorm:"default:'mdi:medal'"`
	Color         string         `json:"color" gorm:"default:'#FBBF24'"`
	XPReward      int64          `json:"xp_reward" gorm:"default:0"`
	Category      string         `json:"category" gorm:"default:'user'"` // user, contributor
	TriggerReason string         `json:"trigger_reason" gorm:"default:''"`
	Metric        string         `json:"metric" gorm:"default:''"`
	Threshold     int64          `json:"threshold" gorm:"default:0"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	IsSecret      bool           `json:"is_secret" gorm:"default:false"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type GamificationRule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Reason    string         `json:"reason" gorm:"uniqueIndex;not null"`
	Points    int64          `json:"points" gorm:"not null"`
	Enabled   bool           `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserAchievement lie un utilisateur à un badge débloqué
type UserAchievement struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"index"`
	AchievementID uint      `json:"achievement_id" gorm:"index"`
	UnlockedAt    time.Time `json:"unlocked_at"`

	// Relations
	User        User        `json:"user,omitempty"`
	Achievement Achievement `json:"achievement,omitempty"`
}

// XPTransaction trace l'origine des gains de points
type XPTransaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Amount    int64     `json:"amount"`
	Reason    string    `json:"reason"` // login, click, news, achievement, etc.
	Metadata  string    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}
