package handlers

import (
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DirectoryHandler gère le trombinoscope (unités, postes, personnes, champs dynamiques)
type DirectoryHandler struct {
	db *gorm.DB
}

// NewDirectoryHandler crée un nouveau handler de trombinoscope
func NewDirectoryHandler(db *gorm.DB) *DirectoryHandler {
	return &DirectoryHandler{db: db}
}

const dateLayout = "2006-01-02"

var (
	phonePattern    = regexp.MustCompile(`^[0-9+().\-\s]{3,30}$`)
	fieldKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,59}$`)
	keyAccents      = strings.NewReplacer(
		"à", "a", "â", "a", "ä", "a", "á", "a", "ç", "c",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"î", "i", "ï", "i", "í", "i", "ô", "o", "ö", "o", "ó", "o",
		"ù", "u", "û", "u", "ü", "u", "ú", "u", "ñ", "n",
	)
	nonKeyChars = regexp.MustCompile(`[^a-z0-9]+`)
)

// canManageDirectory : admin global ou admin d'au moins un groupe
func canManageDirectory(c *gin.Context) bool {
	if c.GetString("role") == "admin" {
		return true
	}
	if ids, ok := c.Get("managed_group_ids"); ok {
		if list, ok := ids.([]uint); ok && len(list) > 0 {
			return true
		}
	}
	return false
}

func parseIDParam(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Identifiant invalide"})
		return 0, false
	}
	return uint(id), true
}

// GetDirectory renvoie toute la structure du trombinoscope en une fois.
// Les champs réservés aux gestionnaires et les personnes inactives sont masqués aux autres utilisateurs.
func (h *DirectoryHandler) GetDirectory(c *gin.Context) {
	isManager := canManageDirectory(c)

	var units []models.OrgUnit
	if err := h.db.
		Preload("Positions", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC, id ASC") }).
		Preload("Positions.Assignments", func(db *gorm.DB) *gorm.DB { return db.Order("\"order\" ASC, id ASC") }).
		Order("\"order\" ASC, name ASC").
		Find(&units).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger la structure"})
		return
	}

	fieldsQuery := h.db.Order("\"order\" ASC, id ASC")
	if !isManager {
		fieldsQuery = fieldsQuery.Where("visibility = ?", models.DirectoryVisibilityAll)
	}
	var fields []models.DirectoryField
	if err := fieldsQuery.Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger les champs"})
		return
	}

	peopleQuery := h.db.Order("last_name ASC, first_name ASC")
	if !isManager {
		peopleQuery = peopleQuery.Where("is_active = ?", true)
	}
	var people []models.DirectoryPerson
	if err := peopleQuery.Find(&people).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Impossible de charger les personnes"})
		return
	}

	visibleKeys := make(map[string]bool, len(fields))
	for _, f := range fields {
		visibleKeys[f.Key] = true
	}
	visiblePersonIDs := make(map[uint]bool, len(people))
	for i := range people {
		visiblePersonIDs[people[i].ID] = true
		people[i].CustomFields = filterCustomFields(people[i].CustomFields, visibleKeys)
	}

	// Ne pas exposer les affectations de personnes masquées (inactives)
	for ui := range units {
		for pi := range units[ui].Positions {
			pos := &units[ui].Positions[pi]
			kept := make([]models.PositionAssignment, 0, len(pos.Assignments))
			for _, a := range pos.Assignments {
				if visiblePersonIDs[a.PersonID] {
					kept = append(kept, a)
				}
			}
			pos.Assignments = kept
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"units":      units,
		"people":     people,
		"fields":     fields,
		"can_manage": isManager,
		"can_admin":  c.GetString("role") == "admin",
	})
}

func filterCustomFields(values models.StringMap, allowed map[string]bool) models.StringMap {
	result := models.StringMap{}
	for k, v := range values {
		if allowed[k] && v != "" {
			result[k] = v
		}
	}
	return result
}

// validateCustomFields vérifie les valeurs selon les définitions et ne conserve que les clés connues
func validateCustomFields(fields []models.DirectoryField, input map[string]string) (models.StringMap, error) {
	result := models.StringMap{}
	for _, f := range fields {
		value := strings.TrimSpace(input[f.Key])
		if value == "" {
			if f.Required {
				return nil, fmt.Errorf("le champ « %s » est obligatoire", f.Label)
			}
			continue
		}
		if len(value) > 5000 {
			return nil, fmt.Errorf("le champ « %s » est trop long", f.Label)
		}
		if err := validateFieldValue(f, value); err != nil {
			return nil, err
		}
		result[f.Key] = value
	}
	return result, nil
}

func validateFieldValue(f models.DirectoryField, value string) error {
	switch f.Type {
	case models.DirectoryFieldEmail:
		if _, err := mail.ParseAddress(value); err != nil {
			return fmt.Errorf("le champ « %s » doit être un email valide", f.Label)
		}
	case models.DirectoryFieldPhone:
		if !phonePattern.MatchString(value) {
			return fmt.Errorf("le champ « %s » doit être un numéro de téléphone valide", f.Label)
		}
	case models.DirectoryFieldURL:
		u, err := url.ParseRequestURI(value)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("le champ « %s » doit être une URL http(s) valide", f.Label)
		}
	case models.DirectoryFieldDate:
		if _, err := time.Parse(dateLayout, value); err != nil {
			return fmt.Errorf("le champ « %s » doit être une date (AAAA-MM-JJ)", f.Label)
		}
	case models.DirectoryFieldSelect:
		for _, opt := range f.Options {
			if opt == value {
				return nil
			}
		}
		return fmt.Errorf("valeur non autorisée pour le champ « %s »", f.Label)
	case models.DirectoryFieldText:
		if len(value) > 500 {
			return fmt.Errorf("le champ « %s » est trop long", f.Label)
		}
	}
	return nil
}

// normalizeFieldKey dérive une clé technique (snake_case ASCII) depuis un libellé
func normalizeFieldKey(s string) string {
	key := keyAccents.Replace(strings.ToLower(strings.TrimSpace(s)))
	key = strings.Trim(nonKeyChars.ReplaceAllString(key, "_"), "_")
	if key == "" || key[0] < 'a' || key[0] > 'z' {
		key = "f_" + key
	}
	if len(key) > 60 {
		key = strings.TrimRight(key[:60], "_")
	}
	return key
}

// isValidPhotoURL accepte un chemin local (/uploads/...) ou une URL http(s)
func isValidPhotoURL(raw string) bool {
	if raw == "" {
		return true
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return true
	}
	u, err := url.ParseRequestURI(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// SearchUsers permet aux gestionnaires de retrouver un compte Airboard à lier à une fiche
func (h *DirectoryHandler) SearchUsers(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if len(q) > 100 {
		q = q[:100]
	}
	query := h.db.Model(&models.User{}).
		Select("id, username, email, first_name, last_name, phone, avatar_url, job_title").
		Where("is_active = ?", true)
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where(
			"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			like, like, like, like,
		)
	}

	type userLite struct {
		ID        uint   `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Phone     string `json:"phone"`
		AvatarURL string `json:"avatar_url"`
		JobTitle  string `json:"job_title"`
	}
	var users []userLite
	if err := query.Order("last_name ASC, first_name ASC").Limit(20).Scan(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Recherche impossible"})
		return
	}
	c.JSON(http.StatusOK, users)
}
