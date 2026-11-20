package handlers

import (
	"net/http"

	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
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

func respondGroupAdminListError(c *gin.Context, what string) {
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error:   "Internal Server Error",
		Message: "Erreur lors de la récupération des " + what,
		Code:    http.StatusInternalServerError,
	})
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
		respondGroupAdminListError(c, "groupes d'applications")
		return
	}
	var appGroups []models.AppGroup
	if err := query.Order("\"order\" ASC, name ASC, id ASC").
		Preload("Applications").
		Preload("OwnerGroup").
		Limit(page.Size).Offset(page.Offset()).
		Find(&appGroups).Error; err != nil {
		respondGroupAdminListError(c, "groupes d'applications")
		return
	}

	c.JSON(http.StatusOK, page.Response(appGroups, total))
}

// GetManagedGroups retourne les groupes administrés par l'utilisateur
func (h *GroupAdminHandler) GetManagedGroups(c *gin.Context) {
	role := c.GetString("role")

	if role == "admin" {
		// Admin global voit tous les groupes
		var groups []models.Group
		h.db.Where("is_active = ?", true).Preload("Users").Preload("AppGroups").Find(&groups)
		c.JSON(http.StatusOK, groups)
		return
	}

	// Group admin voit uniquement ses groupes
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	if len(managedGroupIDs) == 0 {
		c.JSON(http.StatusOK, []models.Group{})
		return
	}

	var groups []models.Group
	h.db.Where("id IN ? AND is_active = ?", managedGroupIDs, true).
		Preload("Users").
		Preload("AppGroups").
		Find(&groups)

	c.JSON(http.StatusOK, groups)
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
		respondGroupAdminListError(c, "applications")
		return
	}
	var applications []models.Application
	if err := query.Order("\"order\" ASC, name ASC, id ASC").
		Preload("AppGroup").
		Limit(page.Size).Offset(page.Offset()).
		Find(&applications).Error; err != nil {
		respondGroupAdminListError(c, "applications")
		return
	}

	c.JSON(http.StatusOK, page.Response(applications, total))
}

// CreateApplication crée une nouvelle application dans un AppGroup privé géré
func (h *GroupAdminHandler) CreateApplication(c *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		URL          string `json:"url" binding:"required,url"`
		Icon         string `json:"icon"`
		Color        string `json:"color"`
		Order        int    `json:"order"`
		OpenInNewTab bool   `json:"open_in_new_tab"`
		IsActive     bool   `json:"is_active"`
		AppGroupID   uint   `json:"app_group_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		// Vérifier que l'AppGroup est privé et appartient aux groupes administrés
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You don't manage any groups",
				Code:    http.StatusForbidden,
			})
			return
		}

		var count int64
		h.db.Model(&models.AppGroup{}).
			Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
			Where("app_groups.id = ? AND group_app_groups.group_id IN ? AND app_groups.is_private = ?",
				req.AppGroupID, managedGroupIDs, true).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You don't have access to this app group",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	var createdByID *uint
	if userIDVal, exists := c.Get("user_id"); exists {
		uid := userIDVal.(uint)
		createdByID = &uid
	}

	application := models.Application{
		Name:        req.Name,
		Description: req.Description,
		URL:         req.URL,
		Icon:        req.Icon,
		Color:       req.Color,
		Order:       req.Order,
		OpenInNewTab: req.OpenInNewTab,
		IsActive:    req.IsActive,
		AppGroupID:  req.AppGroupID,
		CreatedByID: createdByID,
	}

	if err := h.db.Create(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	h.db.Preload("AppGroup").First(&application, application.ID)
	c.JSON(http.StatusCreated, application)
}

// UpdateApplication met à jour une application
func (h *GroupAdminHandler) UpdateApplication(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		URL          string `json:"url" binding:"required,url"`
		Icon         string `json:"icon"`
		Color        string `json:"color"`
		Order        int    `json:"order"`
		OpenInNewTab bool   `json:"open_in_new_tab"`
		IsActive     bool   `json:"is_active"`
		AppGroupID   uint   `json:"app_group_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	var application models.Application
	if err := h.db.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not found",
			Message: "Application not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		// Vérifier que l'application actuelle et le nouveau AppGroup sont accessibles
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You don't manage any groups",
				Code:    http.StatusForbidden,
			})
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
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You don't have access to this application",
				Code:    http.StatusForbidden,
			})
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
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error:   "Forbidden",
					Message: "You don't have access to the target app group",
					Code:    http.StatusForbidden,
				})
				return
			}
		}
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

	if err := h.db.Save(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
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
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not found",
			Message: "Application not found",
			Code:    http.StatusNotFound,
		})
		return
	}

	role := c.GetString("role")
	if role != "admin" {
		// Vérifier que l'application appartient à un AppGroup privé géré
		managedGroupIDs := middleware.GetManagedGroupIDs(c)
		if len(managedGroupIDs) == 0 {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You don't manage any groups",
				Code:    http.StatusForbidden,
			})
			return
		}

		var count int64
		h.db.Model(&models.AppGroup{}).
			Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
			Where("app_groups.id = ? AND group_app_groups.group_id IN ? AND app_groups.is_private = ?",
				application.AppGroupID, managedGroupIDs, true).
			Count(&count)

		if count == 0 {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "You don't have access to this application",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	if err := h.db.Delete(&application).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Database error",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Application deleted successfully",
	})
}
