package handlers

import (
	"log"
	"net/http"

	"airboard/apierror"
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
	_, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
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
			apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des groupes d'applications")
			return
		}

		appGroups = h.withApplications(allAppGroups)
	} else {
		uid := c.GetUint("user_id")
		// Groupes d'appartenance + groupes administrés (lus en base)
		combinedGroupIDs := mergeGroupIDs(userGroupIDs(h.db, uid), adminGroupIDs(h.db, uid))

		if len(combinedGroupIDs) > 0 {
			// Récupérer les AppGroups accessibles via les groupes combinés
			var userAppGroups []models.AppGroup
			if err := h.db.Table("app_groups").
				Joins("JOIN group_app_groups ON app_groups.id = group_app_groups.app_group_id").
				Where("group_app_groups.group_id IN ? AND app_groups.is_active = ?", combinedGroupIDs, true).
				Group("app_groups.id").
				Order("app_groups.\"order\" ASC, app_groups.name ASC").
				Find(&userAppGroups).Error; err != nil {
				apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des groupes d'applications")
				return
			}

			appGroups = h.withApplications(userAppGroups)
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

// withApplications charge les applications actives de tous les groupes en une requête
// (au lieu d'une requête par groupe) et les range sous leur groupe, dans l'ordre donné.
func (h *DashboardHandler) withApplications(groups []models.AppGroup) []models.AppGroupWithApps {
	if len(groups) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(groups))
	for _, ag := range groups {
		ids = append(ids, ag.ID)
	}

	var apps []models.Application
	if err := h.db.Where("app_group_id IN ? AND is_active = ?", ids, true).
		Order("\"order\" ASC, name ASC").
		Find(&apps).Error; err != nil {
		log.Printf("[DASHBOARD] Failed to load applications: %v", err)
	}
	return groupApplications(groups, apps)
}

// groupApplications range les applications (déjà triées) sous leur groupe.
// Chaque groupe a une liste non nulle, vide s'il n'a pas d'application.
func groupApplications(groups []models.AppGroup, apps []models.Application) []models.AppGroupWithApps {
	byGroup := make(map[uint][]models.Application, len(groups))
	for _, app := range apps {
		byGroup[app.AppGroupID] = append(byGroup[app.AppGroupID], app)
	}

	result := make([]models.AppGroupWithApps, 0, len(groups))
	for _, ag := range groups {
		list := byGroup[ag.ID]
		if list == nil {
			list = []models.Application{}
		}
		result = append(result, models.AppGroupWithApps{AppGroup: ag, Applications: list})
	}
	return result
}
