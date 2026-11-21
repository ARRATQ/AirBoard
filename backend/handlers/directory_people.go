package handlers

import (
	"net/http"
	"strings"
	"time"

	"airboard/apierror"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---- Personnes (admin + admin de groupe) ----

// buildPerson valide la requête et renvoie les valeurs à enregistrer
func (h *DirectoryHandler) buildPerson(c *gin.Context) (*models.DirectoryPerson, bool) {
	var req models.DirectoryPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return nil, false
	}
	if !isValidPhotoURL(strings.TrimSpace(req.PhotoURL)) {
		apierror.Respond(c, http.StatusBadRequest, "URL de photo invalide")
		return nil, false
	}
	if req.Phone != "" && !phonePattern.MatchString(req.Phone) {
		apierror.Respond(c, http.StatusBadRequest, "Numéro de téléphone invalide")
		return nil, false
	}
	if req.UserID != nil && !h.exists(&models.User{}, *req.UserID) {
		apierror.Respond(c, http.StatusBadRequest, "Compte utilisateur introuvable")
		return nil, false
	}

	var fields []models.DirectoryField
	if err := h.db.Find(&fields).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de charger les champs")
		return nil, false
	}
	custom, err := validateCustomFields(fields, req.CustomFields)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, err.Error())
		return nil, false
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	return &models.DirectoryPerson{
		FirstName:    strings.TrimSpace(req.FirstName),
		LastName:     strings.TrimSpace(req.LastName),
		Email:        strings.TrimSpace(req.Email),
		Phone:        strings.TrimSpace(req.Phone),
		PhotoURL:     strings.TrimSpace(req.PhotoURL),
		UserID:       req.UserID,
		IsActive:     isActive,
		CustomFields: custom,
	}, true
}

// CreatePerson crée une fiche personne
func (h *DirectoryHandler) CreatePerson(c *gin.Context) {
	person, ok := h.buildPerson(c)
	if !ok {
		return
	}
	// GORM omet la valeur zéro d'un champ avec default:true puis relit le défaut :
	// mémoriser la valeur voulue avant l'insertion et la forcer ensuite.
	wantActive := person.IsActive
	if err := h.db.Create(person).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer la personne")
		return
	}
	if !wantActive {
		if err := h.db.Model(person).Update("is_active", false).Error; err != nil {
			apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer la personne")
			return
		}
		person.IsActive = false
	}
	c.JSON(http.StatusCreated, person)
}

// UpdatePerson modifie une fiche personne
func (h *DirectoryHandler) UpdatePerson(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var existing models.DirectoryPerson
	if err := h.db.First(&existing, id).Error; err != nil {
		respondNotFoundOr500(c, err, "Personne introuvable")
		return
	}
	person, ok := h.buildPerson(c)
	if !ok {
		return
	}
	updates := map[string]interface{}{
		"first_name":    person.FirstName,
		"last_name":     person.LastName,
		"email":         person.Email,
		"phone":         person.Phone,
		"photo_url":     person.PhotoURL,
		"user_id":       person.UserID,
		"is_active":     person.IsActive,
		"custom_fields": person.CustomFields,
	}
	if err := h.db.Model(&existing).Updates(updates).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de modifier la personne")
		return
	}
	h.db.First(&existing, id)
	c.JSON(http.StatusOK, existing)
}

// DeletePerson supprime une fiche personne et ses affectations
func (h *DirectoryHandler) DeletePerson(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("person_id = ?", id).Delete(&models.PositionAssignment{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&models.DirectoryPerson{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		respondNotFoundOr500(c, err, "Personne introuvable")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Personne supprimée"})
}

// ---- Affectations (admin + admin de groupe) ----

func parseStartDate(raw string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateAssignment nomme une personne à un poste
func (h *DirectoryHandler) CreateAssignment(c *gin.Context) {
	var req models.PositionAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if !h.exists(&models.OrgPosition{}, req.PositionID) || !h.exists(&models.DirectoryPerson{}, req.PersonID) {
		apierror.Respond(c, http.StatusBadRequest, "Poste ou personne introuvable")
		return
	}
	var dup int64
	h.db.Model(&models.PositionAssignment{}).
		Where("position_id = ? AND person_id = ?", req.PositionID, req.PersonID).
		Count(&dup)
	if dup > 0 {
		apierror.Respond(c, http.StatusConflict, "Cette personne occupe déjà ce poste")
		return
	}
	start, err := parseStartDate(req.StartDate)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Date de prise de fonction invalide")
		return
	}

	assignment := models.PositionAssignment{
		PositionID: req.PositionID,
		PersonID:   req.PersonID,
		IsInterim:  req.IsInterim,
		StartDate:  start,
		Order:      req.Order,
	}
	if err := h.db.Create(&assignment).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer l'affectation")
		return
	}
	c.JSON(http.StatusCreated, assignment)
}

// UpdateAssignment modifie une affectation (intérim, date, ordre)
func (h *DirectoryHandler) UpdateAssignment(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var assignment models.PositionAssignment
	if err := h.db.First(&assignment, id).Error; err != nil {
		respondNotFoundOr500(c, err, "Affectation introuvable")
		return
	}
	var req models.PositionAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	start, err := parseStartDate(req.StartDate)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Date de prise de fonction invalide")
		return
	}
	updates := map[string]interface{}{
		"is_interim": req.IsInterim,
		"start_date": start,
		"order":      req.Order,
	}
	if err := h.db.Model(&assignment).Updates(updates).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de modifier l'affectation")
		return
	}
	h.db.First(&assignment, id)
	c.JSON(http.StatusOK, assignment)
}

// DeleteAssignment retire une personne d'un poste
func (h *DirectoryHandler) DeleteAssignment(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	res := h.db.Delete(&models.PositionAssignment{}, id)
	if res.Error != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de supprimer l'affectation")
		return
	}
	if res.RowsAffected == 0 {
		apierror.Respond(c, http.StatusNotFound, "Affectation introuvable")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Affectation supprimée"})
}
