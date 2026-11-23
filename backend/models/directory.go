package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Trombinoscope (annuaire hiérarchique de l'organisation)
//
// Structure :
//   OrgUnit (arbre parent → enfants)
//     └─ OrgPosition (postes ordonnés d'une unité)
//          └─ PositionAssignment (0..n titulaires, éventuellement en intérim)
//               └─ DirectoryPerson (fiche personne, liée optionnellement à un User)
//   DirectoryField : définition des champs dynamiques ; les valeurs sont stockées
//   dans DirectoryPerson.CustomFields (clé du champ → valeur).

// Types de champs dynamiques supportés
const (
	DirectoryFieldText     = "text"
	DirectoryFieldTextarea = "textarea"
	DirectoryFieldEmail    = "email"
	DirectoryFieldPhone    = "phone"
	DirectoryFieldURL      = "url"
	DirectoryFieldDate     = "date"
	DirectoryFieldSelect   = "select"
)

// Visibilité des champs dynamiques
const (
	DirectoryVisibilityAll      = "all"      // tous les utilisateurs connectés
	DirectoryVisibilityManagers = "managers" // administrateurs uniquement
)

// StringMap est une map clé/valeur stockée en JSONB
type StringMap map[string]string

// Value implémente driver.Valuer
func (m StringMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implémente sql.Scanner
func (m *StringMap) Scan(value interface{}) error {
	result := StringMap{}
	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
	}
	*m = result
	return nil
}

// StringList est une liste de chaînes stockée en JSONB
type StringList []string

// Value implémente driver.Valuer
func (l StringList) Value() (driver.Value, error) {
	if l == nil {
		return "[]", nil
	}
	b, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan implémente sql.Scanner
func (l *StringList) Scan(value interface{}) error {
	result := StringList{}
	data, err := scanJSONBytes(value)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
	}
	*l = result
	return nil
}

func scanJSONBytes(value interface{}) ([]byte, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, errors.New("valeur JSON non supportée")
	}
}

// OrgUnit représente une unité organisationnelle (direction, département, service…)
type OrgUnit struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Code        string         `json:"code"` // Sigle optionnel (ex. DSI)
	Description string         `json:"description" gorm:"type:text"`
	Color       string         `json:"color" gorm:"default:'#3B82F6'"`
	ParentID    *uint          `json:"parent_id" gorm:"index"`
	Order       int            `json:"order" gorm:"default:0"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	Positions []OrgPosition `json:"positions" gorm:"foreignKey:UnitID"`
}

// OrgPosition représente un poste au sein d'une unité
type OrgPosition struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UnitID    uint           `json:"unit_id" gorm:"not null;index"`
	Title     string         `json:"title" gorm:"not null"`
	IsHead    bool           `json:"is_head" gorm:"default:false"` // Responsable de l'unité
	Order     int            `json:"order" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Assignments []PositionAssignment `json:"assignments" gorm:"foreignKey:PositionID"`
}

// DirectoryPerson représente une personne du trombinoscope
type DirectoryPerson struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	FirstName    string         `json:"first_name" gorm:"not null"`
	LastName     string         `json:"last_name" gorm:"not null"`
	Email        string         `json:"email"`
	Phone        string         `json:"phone"`
	PhotoURL     string         `json:"photo_url"`
	UserID       *uint          `json:"user_id" gorm:"index"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CustomFields StringMap      `json:"custom_fields" gorm:"type:jsonb;default:'{}'"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// PositionAssignment relie une personne à un poste
type PositionAssignment struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	PositionID uint           `json:"position_id" gorm:"not null;index"`
	PersonID   uint           `json:"person_id" gorm:"not null;index"`
	IsInterim  bool           `json:"is_interim" gorm:"default:false"`
	StartDate  *time.Time     `json:"start_date"`
	Order      int            `json:"order" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// DirectoryField définit un champ dynamique des fiches personnes
type DirectoryField struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Key        string         `json:"key" gorm:"not null"`
	Label      string         `json:"label" gorm:"not null"`
	Type       string         `json:"type" gorm:"not null;default:'text'"`
	Options    StringList     `json:"options" gorm:"type:jsonb;default:'[]'"` // Pour le type select
	Required   bool           `json:"required" gorm:"default:false"`
	Visibility string         `json:"visibility" gorm:"not null;default:'all'"`
	ShowOnCard bool           `json:"show_on_card" gorm:"default:false"`
	Order      int            `json:"order" gorm:"default:0"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// ---- Requêtes ----

// OrgUnitRequest corps de création/modification d'une unité
type OrgUnitRequest struct {
	Name        string `json:"name" binding:"required,max=200"`
	Code        string `json:"code" binding:"max=50"`
	Description string `json:"description" binding:"max=2000"`
	Color       string `json:"color" binding:"max=20"`
	ParentID    *uint  `json:"parent_id"`
	Order       int    `json:"order"`
}

// OrgPositionRequest corps de création/modification d'un poste
type OrgPositionRequest struct {
	UnitID uint   `json:"unit_id" binding:"required"`
	Title  string `json:"title" binding:"required,max=200"`
	IsHead bool   `json:"is_head"`
	Order  int    `json:"order"`
}

// DirectoryPersonRequest corps de création/modification d'une personne
type DirectoryPersonRequest struct {
	FirstName    string            `json:"first_name" binding:"required,max=100"`
	LastName     string            `json:"last_name" binding:"required,max=100"`
	Email        string            `json:"email" binding:"omitempty,email,max=200"`
	Phone        string            `json:"phone" binding:"max=50"`
	PhotoURL     string            `json:"photo_url" binding:"max=1000"`
	UserID       *uint             `json:"user_id"`
	IsActive     *bool             `json:"is_active"`
	CustomFields map[string]string `json:"custom_fields"`
}

// PositionAssignmentRequest corps de création/modification d'une affectation
type PositionAssignmentRequest struct {
	PositionID uint   `json:"position_id" binding:"required"`
	PersonID   uint   `json:"person_id" binding:"required"`
	IsInterim  bool   `json:"is_interim"`
	StartDate  string `json:"start_date"` // YYYY-MM-DD, validé dans le handler
	Order      int    `json:"order"`
}

// DirectoryFieldRequest corps de création/modification d'un champ dynamique
type DirectoryFieldRequest struct {
	Key        string   `json:"key" binding:"max=60"`
	Label      string   `json:"label" binding:"required,max=100"`
	Type       string   `json:"type" binding:"required,oneof=text textarea email phone url date select"`
	Options    []string `json:"options"`
	Required   bool     `json:"required"`
	Visibility string   `json:"visibility" binding:"omitempty,oneof=all managers"`
	ShowOnCard bool     `json:"show_on_card"`
	Order      int      `json:"order"`
}
