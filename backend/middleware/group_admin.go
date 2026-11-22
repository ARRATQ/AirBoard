package middleware

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CanManageAppGroup vérifie si un utilisateur peut gérer un AppGroup spécifique
// Un utilisateur peut administrer un AppGroup uniquement si :
// - L'AppGroup est privé (IsPrivate = true)
// - ET l'AppGroup est lié à l'un des groupes de la requête (RequirePermission, via group_app_groups)
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
