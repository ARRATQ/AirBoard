package middleware

import (
	"net/http"

	"airboard/apierror"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireGroupAdmin vérifie que l'utilisateur est admin global OU administre au moins un groupe (via table group_admins)
func (am *AuthMiddleware) RequireGroupAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			apierror.Respond(c, http.StatusUnauthorized, "Authentication requise")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			apierror.Respond(c, http.StatusForbidden, "Rôle invalide")
			c.Abort()
			return
		}

		// Vérifier les rôles non autorisés
		if roleStr == "" || roleStr == "banned" {
			apierror.Respond(c, http.StatusForbidden, "Accès non autorisé")
			c.Abort()
			return
		}

		// Admin global peut tout faire
		if roleStr == "admin" {
			c.Next()
			return
		}

		// Vérifier si l'utilisateur administre au moins un groupe (via table group_admins)
		managedGroupIDsInterface, exists := c.Get("managed_group_ids")
		var managedGroupIDs []uint
		if exists {
			if mgids, ok := managedGroupIDsInterface.([]uint); ok {
				managedGroupIDs = mgids
			}
		}

		if len(managedGroupIDs) > 0 {
			// L'utilisateur administre au moins un groupe, donc il peut accéder
			c.Next()
			return
		}

		// Sinon, accès refusé
		apierror.Respond(c, http.StatusForbidden, "Rôle Admin ou Admin de groupe requis")
		c.Abort()
	}
}

// CanManageGroup vérifie si l'utilisateur peut gérer un groupe spécifique
func CanManageGroup(c *gin.Context, groupID uint) bool {
	role := c.GetString("role")
	if role == "admin" {
		return true // Admin global peut tout gérer
	}

	managedGroupIDsInterface, exists := c.Get("managed_group_ids")
	if !exists {
		return false
	}

	managedGroupIDs, ok := managedGroupIDsInterface.([]uint)
	if !ok {
		return false
	}

	for _, id := range managedGroupIDs {
		if id == groupID {
			return true
		}
	}

	return false
}

// GetManagedGroupIDs retourne les IDs des groupes gérés (helper)
func GetManagedGroupIDs(c *gin.Context) []uint {
	managedGroupIDsInterface, exists := c.Get("managed_group_ids")
	if !exists {
		return []uint{}
	}

	managedGroupIDs, ok := managedGroupIDsInterface.([]uint)
	if !ok {
		return []uint{}
	}

	return managedGroupIDs
}

// CanManageAppGroup vérifie si un utilisateur peut gérer un AppGroup spécifique
// Un utilisateur peut administrer un AppGroup uniquement si :
// - L'AppGroup est privé (IsPrivate = true)
// - ET l'AppGroup est lié à l'un des groupes qu'il administre (via group_app_groups)
// Les AppGroups publics (IsPrivate = false) ne peuvent être administrés que par l'admin global
func CanManageAppGroupWithDB(c *gin.Context, appGroupID uint, db interface{}) bool {
	if IsUnrestricted(c) {
		return true // Admin, ou permission « ressources » sur tout le site
	}

	managedGroupIDs := GetManagedGroupIDs(c)
	if len(managedGroupIDs) == 0 {
		return false
	}

	// Cast vers *gorm.DB
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return false
	}

	// Vérifier si l'AppGroup est privé et lié à un des groupes administrés
	var count int64
	gormDB.Table("app_groups").
		Joins("JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
		Where("app_groups.id = ? AND app_groups.is_private = ? AND group_app_groups.group_id IN ? AND app_groups.deleted_at IS NULL",
			appGroupID, true, managedGroupIDs).
		Count(&count)

	return count > 0
}
