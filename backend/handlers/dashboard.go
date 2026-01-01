package handlers

import (
	"net/http"

	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// @Summary Dashboard utilisateur
// @Description Récupère les groupes d'applications et applications accessibles à l'utilisateur
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.DashboardResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	role, _ := c.Get("role")

	var appGroups []models.AppGroupWithApps

	if role == "admin" {
		// Les admins voient tout
		var allAppGroups []models.AppGroup
		if err := h.db.Where("is_active = ?", true).
			Order("\"order\" ASC, name ASC").
			Find(&allAppGroups).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Erreur lors de la récupération des groupes d'applications",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		for _, ag := range allAppGroups {
			var apps []models.Application
			h.db.Where("app_group_id = ? AND is_active = ?", ag.ID, true).
				Order("\"order\" ASC, name ASC").
				Find(&apps)

			appGroups = append(appGroups, models.AppGroupWithApps{
				AppGroup:     ag,
				Applications: apps,
			})
		}
	} else {
		// Récupérer les groupes administrés par l'utilisateur
		var managedGroupIDs []uint
		h.db.Table("group_admins").Where("user_id = ?", userID).Pluck("group_id", &managedGroupIDs)

		// Récupérer les groupes auxquels l'utilisateur appartient
		var userGroupIDs []uint
		h.db.Table("user_groups").Where("user_id = ?", userID).Pluck("group_id", &userGroupIDs)

		// Combiner les deux : groupes d'appartenance + groupes administrés
		allGroupIDs := make(map[uint]bool)
		for _, id := range userGroupIDs {
			allGroupIDs[id] = true
		}
		for _, id := range managedGroupIDs {
			allGroupIDs[id] = true
		}

		// Convertir en slice
		var combinedGroupIDs []uint
		for id := range allGroupIDs {
			combinedGroupIDs = append(combinedGroupIDs, id)
		}

		if len(combinedGroupIDs) > 0 {
			// Récupérer les AppGroups accessibles via les groupes combinés
			var userAppGroups []models.AppGroup
			if err := h.db.Table("app_groups").
				Joins("JOIN group_app_groups ON app_groups.id = group_app_groups.app_group_id").
				Where("group_app_groups.group_id IN ? AND app_groups.is_active = ?", combinedGroupIDs, true).
				Group("app_groups.id").
				Order("app_groups.\"order\" ASC, app_groups.name ASC").
				Find(&userAppGroups).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "Internal Server Error",
					Message: "Erreur lors de la récupération des groupes d'applications",
					Code:    http.StatusInternalServerError,
				})
				return
			}

			for _, ag := range userAppGroups {
				var apps []models.Application
				h.db.Where("app_group_id = ? AND is_active = ?", ag.ID, true).
					Order("\"order\" ASC, name ASC").
					Find(&apps)

				appGroups = append(appGroups, models.AppGroupWithApps{
					AppGroup:     ag,
					Applications: apps,
				})
			}
		}
	}

	// Statistiques (pour les admins seulement)
	stats := models.DashboardStats{}
	if role == "admin" {
		h.db.Model(&models.AppGroup{}).Where("is_active = ?", true).Count(&stats.TotalAppGroups)
		h.db.Model(&models.Application{}).Where("is_active = ?", true).Count(&stats.TotalApplications)
		h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.TotalUsers)
		h.db.Model(&models.Group{}).Where("is_active = ?", true).Count(&stats.TotalGroups)
	}

	response := models.DashboardResponse{
		AppGroups: appGroups,
		Stats:     stats,
	}

	c.JSON(http.StatusOK, response)
}
