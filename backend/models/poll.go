package models

import (
	"time"

	"gorm.io/gorm"
)

// Poll représente un sondage/mini-enquête
type Poll struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title" gorm:"not null"`               // Question du sondage
	Description string         `json:"description" gorm:"type:text"`        // Description détaillée
	PollType    string         `json:"poll_type" gorm:"default:'single'"`   // single (choix unique), multiple (choix multiples)
	IsAnonymous bool           `json:"is_anonymous" gorm:"default:false"`   // Résultats anonymes ou non
	IsActive    bool           `json:"is_active" gorm:"default:true"`       // Sondage actif (ouvert) ou fermé
	ShowResults string         `json:"show_results" gorm:"default:'after'"` // always (toujours), after (après avoir voté), closed (après fermeture)
	StartDate   *time.Time     `json:"start_date"`                          // Date de début (optionnel)
	EndDate     *time.Time     `json:"end_date"`                            // Date de fin (optionnel)
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Auteur
	AuthorID uint `json:"author_id"`
	Author   User `json:"author" gorm:"foreignKey:AuthorID"`

	// Liens optionnels avec news et annonces
	NewsID         *uint         `json:"news_id"`
	News           *News         `json:"news,omitempty" gorm:"foreignKey:NewsID"`
	AnnouncementID *uint         `json:"announcement_id"`
	Announcement   *Announcement `json:"announcement,omitempty" gorm:"foreignKey:AnnouncementID"`

	// Granularité - groupes cibles
	TargetGroups []Group `json:"target_groups" gorm:"many2many:poll_target_groups;"`
	// Si vide, visible par tous (sondage global)

	// Relations
	Options []PollOption `json:"options,omitempty" gorm:"foreignKey:PollID"`
	Votes   []PollVote   `json:"votes,omitempty" gorm:"foreignKey:PollID"`

	// Compteur calculé (non persisté)
	TotalVotes int `json:"total_votes" gorm:"-"`
}

// PollOption représente une option de réponse d'un sondage
type PollOption struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PollID    uint           `json:"poll_id" gorm:"not null;index"`
	Text      string         `json:"text" gorm:"not null"`
	IsOther   bool           `json:"is_other" gorm:"default:false"`
	Order     int            `json:"order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Poll  Poll       `json:"-" gorm:"foreignKey:PollID"`
	Votes []PollVote `json:"votes,omitempty" gorm:"foreignKey:PollOptionID"`
}

// PollVote représente un vote d'un utilisateur sur une option
type PollVote struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	PollID       uint      `json:"poll_id" gorm:"not null;index"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	PollOptionID uint      `json:"poll_option_id" gorm:"not null;index"`
	VotedAt      time.Time `json:"voted_at"`
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Poll       Poll       `json:"poll,omitempty" gorm:"foreignKey:PollID"`
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	PollOption PollOption `json:"poll_option,omitempty" gorm:"foreignKey:PollOptionID"`

	// Index composite pour choix unique : un user peut voter UNE FOIS par sondage sur UNE option
	// Pour choix multiples, on autorise plusieurs votes mais sur des options différentes
}

// PollRequest pour la création/modification de sondages
type PollRequest struct {
	Title          string              `json:"title" binding:"required"`
	Description    string              `json:"description"`
	PollType       string              `json:"poll_type" binding:"required,oneof=single multiple"`
	IsAnonymous    bool                `json:"is_anonymous"`
	IsActive       bool                `json:"is_active"`
	ShowResults    string              `json:"show_results" binding:"oneof=always after closed"`
	StartDate      *time.Time          `json:"start_date"`
	EndDate        *time.Time          `json:"end_date"`
	NewsID         *uint               `json:"news_id"`          // Lier à un article
	AnnouncementID *uint               `json:"announcement_id"`  // Lier à une annonce
	TargetGroupIDs []uint              `json:"target_group_ids"` // IDs des groupes cibles
	Options        []PollOptionRequest `json:"options" binding:"required,min=2,max=10"`
}

// PollOptionRequest pour créer/modifier une option
type PollOptionRequest struct {
	ID      uint   `json:"id"` // Pour modification (si existe déjà)
	Text    string `json:"text" binding:"required,min=1,max=200"`
	IsOther bool   `json:"is_other"`
	Order   int    `json:"order"`
}

// PollVoteRequest pour voter
type PollVoteRequest struct {
	PollOptionIDs []uint `json:"poll_option_ids" binding:"required,min=1"` // Supporte choix multiples
}

// PollListResponse pour les listes avec pagination
type PollListResponse struct {
	Polls      []Poll `json:"polls"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

// PollResultsResponse pour les résultats d'un sondage
type PollResultsResponse struct {
	Poll         Poll               `json:"poll"`
	TotalVotes   int64              `json:"total_votes"`
	UniqueVoters int64              `json:"unique_voters"`
	Options      []PollOptionResult `json:"options"`
	UserVoted    bool               `json:"user_voted"`
	UserVotes    []uint             `json:"user_votes"`              // IDs des options votées par l'utilisateur
	VoterDetails []PollVoterDetail  `json:"voter_details,omitempty"` // Seulement si non anonyme et autorisé
}

// PollOptionResult pour le résultat d'une option
type PollOptionResult struct {
	PollOption PollOption `json:"poll_option"`
	VoteCount  int64      `json:"vote_count"`
	Percentage float64    `json:"percentage"`
}

// PollVoterDetail pour afficher qui a voté (sondages non anonymes)
type PollVoterDetail struct {
	User         User      `json:"user"`
	PollOptionID uint      `json:"poll_option_id"`
	VotedAt      time.Time `json:"voted_at"`
}

// PollStatsResponse pour les statistiques globales
type PollStatsResponse struct {
	TotalPolls  int64           `json:"total_polls"`
	ActivePolls int64           `json:"active_polls"`
	ClosedPolls int64           `json:"closed_polls"`
	TotalVotes  int64           `json:"total_votes"`
	TotalVoters int64           `json:"total_voters"`
	TopPolls    []PollWithStats `json:"top_polls"`
	RecentPolls []Poll          `json:"recent_polls"`
}

// PollWithStats pour les sondages avec stats supplémentaires
type PollWithStats struct {
	Poll
	VoteCount  int64 `json:"vote_count"`
	VoterCount int64 `json:"voter_count"`
}

// TableName spécifie le nom de la table pour Poll
func (Poll) TableName() string {
	return "polls"
}

// TableName spécifie le nom de la table pour PollOption
func (PollOption) TableName() string {
	return "poll_options"
}

// TableName spécifie le nom de la table pour PollVote
func (PollVote) TableName() string {
	return "poll_votes"
}

// BeforeSave hook pour PollVote - définir VotedAt
func (pv *PollVote) BeforeCreate(tx *gorm.DB) error {
	if pv.VotedAt.IsZero() {
		pv.VotedAt = time.Now()
	}
	return nil
}

// AfterDelete hook pour supprimer les votes associés quand on supprime une option
func (po *PollOption) AfterDelete(tx *gorm.DB) error {
	return tx.Where("poll_option_id = ?", po.ID).Delete(&PollVote{}).Error
}

// AfterDelete hook pour supprimer les options et votes quand on supprime un sondage
func (p *Poll) AfterDelete(tx *gorm.DB) error {
	// Supprimer les votes
	if err := tx.Where("poll_id = ?", p.ID).Delete(&PollVote{}).Error; err != nil {
		return err
	}
	// Supprimer les options
	if err := tx.Where("poll_id = ?", p.ID).Delete(&PollOption{}).Error; err != nil {
		return err
	}
	return nil
}
