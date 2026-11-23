package handlers

import (
	"errors"
	"net/http"
	"strings"

	"airboard/apierror"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ---- Unités (admin) ----

// CreateUnit crée une unité organisationnelle
func (h *DirectoryHandler) CreateUnit(c *gin.Context) {
	var req models.OrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if req.ParentID != nil && !h.exists(&models.OrgUnit{}, *req.ParentID) {
		apierror.Respond(c, http.StatusBadRequest, "Unité parente introuvable")
		return
	}

	unit := models.OrgUnit{
		Name:        strings.TrimSpace(req.Name),
		Code:        strings.TrimSpace(req.Code),
		Description: req.Description,
		Color:       defaultColor(req.Color),
		ParentID:    req.ParentID,
		Order:       req.Order,
	}
	if err := h.db.Create(&unit).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer l'unité")
		return
	}
	unit.Positions = []models.OrgPosition{}
	c.JSON(http.StatusCreated, unit)
}

// UpdateUnit modifie une unité (avec contrôle anti-cycle sur le parent)
func (h *DirectoryHandler) UpdateUnit(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var unit models.OrgUnit
	if err := h.db.First(&unit, id).Error; err != nil {
		respondNotFoundOr500(c, err, "Unité introuvable")
		return
	}
	var req models.OrgUnitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if req.ParentID != nil {
		if !h.exists(&models.OrgUnit{}, *req.ParentID) {
			apierror.Respond(c, http.StatusBadRequest, "Unité parente introuvable")
			return
		}
		if h.wouldCreateCycle(id, *req.ParentID) {
			apierror.Respond(c, http.StatusBadRequest, "Une unité ne peut pas être rattachée à elle-même ou à l'une de ses sous-unités")
			return
		}
	}

	updates := map[string]interface{}{
		"name":        strings.TrimSpace(req.Name),
		"code":        strings.TrimSpace(req.Code),
		"description": req.Description,
		"color":       defaultColor(req.Color),
		"parent_id":   req.ParentID,
		"order":       req.Order,
	}
	if err := h.db.Model(&unit).Updates(updates).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de modifier l'unité")
		return
	}
	h.db.First(&unit, id)
	c.JSON(http.StatusOK, unit)
}

// DeleteUnit supprime une unité sans sous-unités, ainsi que ses postes et affectations
func (h *DirectoryHandler) DeleteUnit(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var children int64
	h.db.Model(&models.OrgUnit{}).Where("parent_id = ?", id).Count(&children)
	if children > 0 {
		apierror.Respond(c, http.StatusConflict, "Supprimez ou déplacez d'abord les sous-unités")
		return
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		var positionIDs []uint
		if err := tx.Model(&models.OrgPosition{}).Where("unit_id = ?", id).Pluck("id", &positionIDs).Error; err != nil {
			return err
		}
		if len(positionIDs) > 0 {
			if err := tx.Where("position_id IN ?", positionIDs).Delete(&models.PositionAssignment{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", positionIDs).Delete(&models.OrgPosition{}).Error; err != nil {
				return err
			}
		}
		res := tx.Delete(&models.OrgUnit{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		respondNotFoundOr500(c, err, "Unité introuvable")
		return
	}
	services.SyncDirectoryGroupsLogged(h.db, "entité supprimée")
	c.JSON(http.StatusOK, gin.H{"message": "Unité supprimée"})
}

func (h *DirectoryHandler) wouldCreateCycle(unitID, newParentID uint) bool {
	current := &newParentID
	for depth := 0; current != nil && depth < 100; depth++ {
		if *current == unitID {
			return true
		}
		var parent models.OrgUnit
		if err := h.db.Select("id, parent_id").First(&parent, *current).Error; err != nil {
			return false
		}
		current = parent.ParentID
	}
	return current != nil // profondeur anormale : refuser par prudence
}

// ---- Postes (admin) ----

// CreatePosition crée un poste dans une unité
func (h *DirectoryHandler) CreatePosition(c *gin.Context) {
	var req models.OrgPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if !h.exists(&models.OrgUnit{}, req.UnitID) {
		apierror.Respond(c, http.StatusBadRequest, "Unité introuvable")
		return
	}
	position := models.OrgPosition{
		UnitID: req.UnitID,
		Title:  strings.TrimSpace(req.Title),
		IsHead: req.IsHead,
		Order:  req.Order,
	}
	if err := h.db.Create(&position).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer le poste")
		return
	}
	position.Assignments = []models.PositionAssignment{}
	c.JSON(http.StatusCreated, position)
}

// UpdatePosition modifie un poste
func (h *DirectoryHandler) UpdatePosition(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var position models.OrgPosition
	if err := h.db.First(&position, id).Error; err != nil {
		respondNotFoundOr500(c, err, "Poste introuvable")
		return
	}
	var req models.OrgPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if !h.exists(&models.OrgUnit{}, req.UnitID) {
		apierror.Respond(c, http.StatusBadRequest, "Unité introuvable")
		return
	}
	updates := map[string]interface{}{
		"unit_id": req.UnitID,
		"title":   strings.TrimSpace(req.Title),
		"is_head": req.IsHead,
		"order":   req.Order,
	}
	if err := h.db.Model(&position).Updates(updates).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de modifier le poste")
		return
	}
	h.db.Preload("Assignments").First(&position, id)
	services.SyncDirectoryGroupsLogged(h.db, "poste modifié")
	c.JSON(http.StatusOK, position)
}

// DeletePosition supprime un poste et ses affectations
func (h *DirectoryHandler) DeletePosition(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("position_id = ?", id).Delete(&models.PositionAssignment{}).Error; err != nil {
			return err
		}
		res := tx.Delete(&models.OrgPosition{}, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		respondNotFoundOr500(c, err, "Poste introuvable")
		return
	}
	services.SyncDirectoryGroupsLogged(h.db, "poste supprimé")
	c.JSON(http.StatusOK, gin.H{"message": "Poste supprimé"})
}

// ---- Champs dynamiques (admin) ----

// CreateField crée une définition de champ dynamique
func (h *DirectoryHandler) CreateField(c *gin.Context) {
	var req models.DirectoryFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	key := strings.TrimSpace(req.Key)
	if key == "" {
		key = normalizeFieldKey(req.Label)
	}
	if !fieldKeyPattern.MatchString(key) {
		apierror.Respond(c, http.StatusBadRequest, "Clé invalide (lettres minuscules, chiffres et _)")
		return
	}
	var count int64
	h.db.Model(&models.DirectoryField{}).Where("key = ?", key).Count(&count)
	if count > 0 {
		apierror.Respond(c, http.StatusConflict, "Un champ avec cette clé existe déjà")
		return
	}
	options, err := cleanFieldOptions(req.Type, req.Options)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, err.Error())
		return
	}

	field := models.DirectoryField{
		Key:        key,
		Label:      strings.TrimSpace(req.Label),
		Type:       req.Type,
		Options:    options,
		Required:   req.Required,
		Visibility: defaultVisibility(req.Visibility),
		ShowOnCard: req.ShowOnCard,
		Order:      req.Order,
	}
	if err := h.db.Create(&field).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer le champ")
		return
	}
	c.JSON(http.StatusCreated, field)
}

// UpdateField modifie un champ dynamique (la clé est immuable pour préserver les valeurs)
func (h *DirectoryHandler) UpdateField(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var field models.DirectoryField
	if err := h.db.First(&field, id).Error; err != nil {
		respondNotFoundOr500(c, err, "Champ introuvable")
		return
	}
	var req models.DirectoryFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	options, err := cleanFieldOptions(req.Type, req.Options)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, err.Error())
		return
	}
	updates := map[string]interface{}{
		"label":        strings.TrimSpace(req.Label),
		"type":         req.Type,
		"options":      options,
		"required":     req.Required,
		"visibility":   defaultVisibility(req.Visibility),
		"show_on_card": req.ShowOnCard,
		"order":        req.Order,
	}
	if err := h.db.Model(&field).Updates(updates).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de modifier le champ")
		return
	}
	h.db.First(&field, id)
	c.JSON(http.StatusOK, field)
}

// DeleteField supprime une définition de champ (les valeurs stockées sont ignorées ensuite)
func (h *DirectoryHandler) DeleteField(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	res := h.db.Delete(&models.DirectoryField{}, id)
	if res.Error != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de supprimer le champ")
		return
	}
	if res.RowsAffected == 0 {
		apierror.Respond(c, http.StatusNotFound, "Champ introuvable")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Champ supprimé"})
}

func cleanFieldOptions(fieldType string, raw []string) (models.StringList, error) {
	if fieldType != models.DirectoryFieldSelect {
		return models.StringList{}, nil
	}
	seen := map[string]bool{}
	options := models.StringList{}
	for _, o := range raw {
		o = strings.TrimSpace(o)
		if o == "" || seen[o] {
			continue
		}
		if len(o) > 200 {
			return nil, errors.New("une option de liste est trop longue")
		}
		seen[o] = true
		options = append(options, o)
	}
	if len(options) == 0 {
		return nil, errors.New("une liste de choix doit contenir au moins une option")
	}
	return options, nil
}

// ---- Utilitaires ----

func (h *DirectoryHandler) exists(model interface{}, id uint) bool {
	var count int64
	h.db.Model(model).Where("id = ?", id).Count(&count)
	return count > 0
}

func respondNotFoundOr500(c *gin.Context, err error, notFoundMsg string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		apierror.Respond(c, http.StatusNotFound, notFoundMsg)
		return
	}
	apierror.Respond(c, http.StatusInternalServerError, "Erreur interne")
}

func defaultColor(color string) string {
	color = strings.TrimSpace(color)
	if color == "" {
		return "#3B82F6"
	}
	return color
}

func defaultVisibility(v string) string {
	if v == models.DirectoryVisibilityManagers {
		return v
	}
	return models.DirectoryVisibilityAll
}
