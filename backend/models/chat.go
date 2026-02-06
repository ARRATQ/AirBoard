package models

import (
	"time"

	"gorm.io/gorm"
)

// ChatMessage représente un message dans le système de chat
type ChatMessage struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Content string `json:"content" gorm:"type:text;not null"`
	Type    string `json:"type" gorm:"default:'text'"` // text, image, file, system

	// Expéditeur
	SenderID uint `json:"sender_id" gorm:"not null;index"`
	Sender   User `json:"sender" gorm:"foreignKey:SenderID"`

	// Destinataire (Soit un User (DM), soit un Groupe)
	RecipientID *uint `json:"recipient_id,omitempty" gorm:"index"` // Pour DM
	Recipient   *User `json:"recipient,omitempty" gorm:"foreignKey:RecipientID"`

	GroupID *uint  `json:"group_id,omitempty" gorm:"index"` // Pour Group Chat
	Group   *Group `json:"group,omitempty" gorm:"foreignKey:GroupID"`

	IsRead    bool           `json:"is_read" gorm:"default:false;index"` // Read status
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName spécifie le nom de la table
func (ChatMessage) TableName() string {
	return "chat_messages"
}
