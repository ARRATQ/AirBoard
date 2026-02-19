package models

import (
	"time"

	"gorm.io/gorm"
)

type Suggestion struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"type:text;not null"`
	Category    string `json:"category" gorm:"default:''"`
	CategoryID  *uint  `json:"category_id,omitempty"`
	Status      string `json:"status" gorm:"default:'new'"`
	IsArchived  bool   `json:"is_archived" gorm:"default:false"`
	IsAnonymous bool   `json:"is_anonymous" gorm:"default:false"`

	UserID uint `json:"user_id"`
	User   User `json:"user,omitempty" gorm:"foreignKey:UserID"`

	PollID       *uint               `json:"poll_id,omitempty"`
	Poll         *Poll               `json:"poll,omitempty" gorm:"foreignKey:PollID"`
	PollOptionID *uint               `json:"poll_option_id,omitempty"`
	PollOption   *PollOption         `json:"poll_option,omitempty" gorm:"foreignKey:PollOptionID"`
	CategoryRef  *SuggestionCategory `json:"category_ref,omitempty" gorm:"foreignKey:CategoryID"`
	Votes        []SuggestionVote    `json:"-" gorm:"foreignKey:SuggestionID"`

	Upvotes   int64  `json:"upvotes" gorm:"-"`
	Downvotes int64  `json:"downvotes" gorm:"-"`
	UserVote  string `json:"user_vote,omitempty" gorm:"-"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type SuggestionRequest struct {
	Title        string `json:"title" binding:"required,min=3,max=180"`
	Description  string `json:"description" binding:"required,max=5000"`
	CategoryID   *uint  `json:"category_id" binding:"required"`
	IsAnonymous  bool   `json:"is_anonymous"`
	PollID       *uint  `json:"poll_id"`
	PollOptionID *uint  `json:"poll_option_id"`
}

type SuggestionStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=new reviewing planned in_progress implemented rejected"`
}

type SuggestionArchiveUpdateRequest struct {
	IsArchived bool `json:"is_archived"`
}

type SuggestionListResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Total       int64        `json:"total"`
	Page        int          `json:"page"`
	PageSize    int          `json:"page_size"`
	TotalPages  int          `json:"total_pages"`
}

type SuggestionCategory struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null;uniqueIndex"`
	Slug      string         `json:"slug" gorm:"not null;uniqueIndex"`
	Order     int            `json:"order" gorm:"default:0"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type SuggestionCategoryRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=80"`
	Slug     string `json:"slug" binding:"required,min=2,max=80"`
	Order    int    `json:"order"`
	IsActive *bool  `json:"is_active"`
}

type SuggestionVote struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	SuggestionID uint      `json:"suggestion_id" gorm:"not null;index"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	VoteType     string    `json:"vote_type" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SuggestionVoteRequest struct {
	VoteType string `json:"vote_type" binding:"required,oneof=up down none"`
}

type SuggestionVoteResponse struct {
	SuggestionID uint   `json:"suggestion_id"`
	Upvotes      int64  `json:"upvotes"`
	Downvotes    int64  `json:"downvotes"`
	UserVote     string `json:"user_vote,omitempty"`
}

func (Suggestion) TableName() string {
	return "suggestions"
}

func (SuggestionCategory) TableName() string {
	return "suggestion_categories"
}

func (SuggestionVote) TableName() string {
	return "suggestion_votes"
}
