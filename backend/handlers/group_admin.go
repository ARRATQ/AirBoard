package handlers

import (
	"log"
	"net/http"
	"time"

	"airboard/apierror"
	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type GroupAdminHandler struct {
	db *gorm.DB
}

func NewGroupAdminHandler(db *gorm.DB) *GroupAdminHandler {
	return &GroupAdminHandler{db: db}
}

// groupAdminPageLimits : les écrans de l'admin de groupe filtrent côté client et demandent une
// grande page ; le plafond borne quand même la requête.
var groupAdminPageLimits = pageLimits{Default: 100, Max: 1000}

// privateAppGroupIDs : sous-requête des AppGroups privés liés aux groupes administrés.
func (h *GroupAdminHandler) privateAppGroupIDs(managedGroupIDs []uint) *gorm.DB {
	return h.db.Table("group_app_groups").
		Select("group_app_groups.app_group_id").
		Joins("JOIN app_groups ON app_groups.id = group_app_groups.app_group_id").
		Where("group_app_groups.group_id IN ? AND app_groups.is_private = ? AND app_groups.deleted_at IS NULL", managedGroupIDs, true)
}

// respondGroupAdminListError journalise l'erreur SQL et renvoie un message sans détail technique.
func respondGroupAdminListError(c *gin.Context, what string, err error) {
	log.Printf("group-admin: lecture des %s impossible : %v", what, err)
	apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des "+what)
}

// GetAppGroups retourne une page des AppGroups accessibles : tous pour l'admin global,
// les AppGroups privés liés aux groupes administrés pour un admin de groupe.
func (h *GroupAdminHandler) GetAppGroups(c *gin.Context) {
	page := parsePageParams(c, groupAdminPageLimits)
	query := h.db.Model(&models.AppGroup{})

	if c.GetString("role") != "admin" {
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			c.JSON(http.StatusOK, page.Response([]models.AppGroup{}, 0))
			return
		}
		query = query.Where("app_groups.id IN (?)", h.privateAppGroupIDs(managedGroupIDs))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondGroupAdminListError(c, "groupes d'applications", err)
		return
	}
	var appGroups []models.AppGroup
	if err := query.Order("\"order\" ASC, name ASC, id ASC").
		Preload("Applications").
		Preload("OwnerGroup").
		Limit(page.Size).Offset(page.Offset()).
		Find(&appGroups).Error; err != nil {
		respondGroupAdminListError(c, "groupes d'applications", err)
		return
	}

	c.JSON(http.StatusOK, page.Response(appGroups, total))
}

// managedGroup est un groupe accompagné de son nombre de membres : la liste des utilisateurs
// (adresses e-mail comprises) n'est pas envoyée.
type managedGroup struct {
	models.Group
	UserCount int64 `json:"user_count"`
}

// GetManagedGroups retourne les groupes actifs administrés par l'utilisateur (tous pour l'admin
// global), avec leur nombre de membres.
func (h *GroupAdminHandler) GetManagedGroups(c *gin.Context) {
	query := h.db.Where("is_active = ?", true)
	if c.GetString("role") != "admin" {
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			c.JSON(http.StatusOK, []managedGroup{})
			return
		}
		query = query.Where("id IN ?", managedGroupIDs)
	}

	var groups []models.Group
	if err := query.Order("name ASC").Find(&groups).Error; err != nil {
		respondGroupAdminListError(c, "groupes gérés", err)
		return
	}

	counts, err := h.groupUserCounts(groups)
	if err != nil {
		respondGroupAdminListError(c, "groupes gérés", err)
		return
	}
	result := make([]managedGroup, len(groups))
	for i, g := range groups {
		result[i] = managedGroup{Group: g, UserCount: counts[g.ID]}
	}
	c.JSON(http.StatusOK, result)
}

// groupUserCounts compte les membres (comptes non supprimés) de chaque groupe en une requête.
func (h *GroupAdminHandler) groupUserCounts(groups []models.Group) (map[uint]int64, error) {
	counts := make(map[uint]int64, len(groups))
	if len(groups) == 0 {
		return counts, nil
	}
	ids := make([]uint, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	var rows []struct {
		GroupID uint
		Total   int64
	}
	err := h.db.Table("user_groups").
		Select("user_groups.group_id, COUNT(*) AS total").
		Joins("JOIN users ON users.id = user_groups.user_id AND users.deleted_at IS NULL").
		Where("user_groups.group_id IN ?", ids).
		Group("user_groups.group_id").
		Scan(&rows).Error
	for _, r := range rows {
		counts[r.GroupID] = r.Total
	}
	return counts, err
}

// GetApplications retourne une page des applications : toutes pour l'admin global, celles des
// AppGroups privés liés aux groupes administrés pour un admin de groupe.
func (h *GroupAdminHandler) GetApplications(c *gin.Context) {
	page := parsePageParams(c, groupAdminPageLimits)
	query := h.db.Model(&models.Application{})

	if c.GetString("role") != "admin" {
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			c.JSON(http.StatusOK, page.Response([]models.Application{}, 0))
			return
		}
		query = query.Where("app_group_id IN (?)", h.privateAppGroupIDs(managedGroupIDs))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondGroupAdminListError(c, "applications", err)
		return
	}
	var applications []models.Application
	if err := query.Order("\"order\" ASC, name ASC, id ASC").
		Preload("AppGroup").
		Limit(page.Size).Offset(page.Offset()).
		Find(&applications).Error; err != nil {
		respondGroupAdminListError(c, "applications", err)
		return
	}

	c.JSON(http.StatusOK, page.Response(applications, total))
}

// CreateApplication crée une nouvelle application dans un AppGroup privé géré
func (h *GroupAdminHandler) CreateApplication(c *gin.Context) {
	var req struct {
		Name           string     `json:"name" binding:"required"`
		Description    string     `json:"description"`
		URL            string     `json:"url" binding:"required,url"`
		Icon           string     `json:"icon"`
		Color          string     `json:"color"`
		Order          int        `json:"order"`
		OpenInNewTab   bool       `json:"open_in_new_tab"`
		IsActive       bool       `json:"is_active"`
		AppGroupID     uint       `json:"app_group_id" binding:"required"`
		AdminNote      string     `json:"admin_note"`
		AdminNoteUntil *time.Time `json:"admin_note_until"`
	}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		// Vérifier que l'AppGroup est privé et appartient aux groupes administrés
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			apierror.Respond(c, http.StatusForbidden, "You don't manage any groups")
			return
		}

		var count int64
		h.db.Model(&models.AppGroup{}).
			Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
			Where("app_groups.id = ? AND group_app_groups.group_id IN ? AND app_groups.is_private = ?",
				req.AppGroupID, managedGroupIDs, true).
			Count(&count)

		if count == 0 {
			apierror.Respond(c, http.StatusForbidden, "You don't have access to this app group")
			return
		}
	}

	var createdByID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		uid := userIDVal.(uint)
		createdByID = &uid
	}

	note, noteUntil, err := models.NormalizeAdminNote(req.AdminNote, req.AdminNoteUntil)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, err.Error())
		return
	}

	application := models.Application{
		Name:           req.Name,
		Description:    req.Description,
		URL:            req.URL,
		Icon:           req.Icon,
		Color:          req.Color,
		Order:          req.Order,
		OpenInNewTab:   req.OpenInNewTab,
		IsActive:       req.IsActive,
		AppGroupID:     req.AppGroupID,
		CreatedByID:    createdByID,
		AdminNote:      note,
		AdminNoteUntil: noteUntil,
	}

	if err := createKeepingFalse(h.db, c, &application, "is_active", "open_in_new_tab"); err != nil {
		log.Printf("group-admin: création de l'application impossible : %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création de l'application")
		return
	}

	h.db.Preload("AppGroup").First(&application, application.ID)
	c.JSON(http.StatusCreated, application)
}

// UpdateApplication met à jour une application
func (h *GroupAdminHandler) UpdateApplication(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name           string     `json:"name" binding:"required"`
		Description    string     `json:"description"`
		URL            string     `json:"url" binding:"required,url"`
		Icon           string     `json:"icon"`
		Color          string     `json:"color"`
		Order          int        `json:"order"`
		OpenInNewTab   bool       `json:"open_in_new_tab"`
		IsActive       bool       `json:"is_active"`
		AppGroupID     uint       `json:"app_group_id" binding:"required"`
		AdminNote      string     `json:"admin_note"`
		AdminNoteUntil *time.Time `json:"admin_note_until"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	var application models.Application
	if err := h.db.First(&application, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Application not found")
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		// Vérifier que l'application actuelle et le nouveau AppGroup sont accessibles
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			apierror.Respond(c, http.StatusForbidden, "You don't manage any groups")
			return
		}

		// Vérifier l'AppGroup actuel
		var currentCount int64
		h.db.Model(&models.AppGroup{}).
			Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
			Where("app_groups.id = ? AND group_app_groups.group_id IN ? AND app_groups.is_private = ?",
				application.AppGroupID, managedGroupIDs, true).
			Count(&currentCount)

		if currentCount == 0 {
			apierror.Respond(c, http.StatusForbidden, "You don't have access to this application")
			return
		}

		// Vérifier le nouveau AppGroup si différent
		if req.AppGroupID != application.AppGroupID {
			var newCount int64
			h.db.Model(&models.AppGroup{}).
				Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
				Where("app_groups.id = ? AND group_app_groups.group_id IN ? AND app_groups.is_private = ?",
					req.AppGroupID, managedGroupIDs, true).
				Count(&newCount)

			if newCount == 0 {
				apierror.Respond(c, http.StatusForbidden, "You don't have access to the target app group")
				return
			}
		}
	}

	note, noteUntil, err := models.NormalizeAdminNote(req.AdminNote, req.AdminNoteUntil)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, err.Error())
		return
	}

	application.Name = req.Name
	application.Description = req.Description
	application.URL = req.URL
	application.Icon = req.Icon
	application.Color = req.Color
	application.Order = req.Order
	application.OpenInNewTab = req.OpenInNewTab
	application.IsActive = req.IsActive
	application.AppGroupID = req.AppGroupID
	application.AdminNote = note
	application.AdminNoteUntil = noteUntil

	if err := h.db.Save(&application).Error; err != nil {
		log.Printf("group-admin: mise à jour de l'application %d impossible : %v", application.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour de l'application")
		return
	}

	h.db.Preload("AppGroup").First(&application, application.ID)
	c.JSON(http.StatusOK, application)
}

// DeleteApplication supprime une application
func (h *GroupAdminHandler) DeleteApplication(c *gin.Context) {
	id := c.Param("id")

	var application models.Application
	if err := h.db.First(&application, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Application not found")
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		// Vérifier que l'application appartient à un AppGroup privé géré
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			apierror.Respond(c, http.StatusForbidden, "You don't manage any groups")
			return
		}

		var count int64
		h.db.Model(&models.AppGroup{}).
			Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
			Where("app_groups.id = ? AND group_app_groups.group_id IN ? AND app_groups.is_private = ?",
				application.AppGroupID, managedGroupIDs, true).
			Count(&count)

		if count == 0 {
			apierror.Respond(c, http.StatusForbidden, "You don't have access to this application")
			return
		}
	}

	if err := h.db.Delete(&application).Error; err != nil {
		log.Printf("group-admin: suppression de l'application %d impossible : %v", application.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression de l'application")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Application deleted successfully",
	})
}
