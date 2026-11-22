package middleware

import (
	"net/http"

	"airboard/apierror"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Clés de contexte posées par RequireAuth et RequirePermission (exportées pour les tests).
const (
	CtxPermissions    = "permissions"
	CtxMemberGroupIDs = "member_group_ids"
	ctxUnrestricted   = "permission_unrestricted"
	// ctxManagedGroupIDs : groupes où la requête agit (RequirePermission, portée « ses groupes »)
	ctxManagedGroupIDs = "managed_group_ids"
)

// LoadPermissions lit les permissions effectives d'un utilisateur : l'union des permissions
// de ses rôles, la portée « tout le site » l'emportant sur « ses groupes ».
func LoadPermissions(db *gorm.DB, userID uint) (map[string]string, error) {
	var rows []struct {
		Permission string
		Scope      string
	}
	err := db.Table("access_role_permissions").
		Select("access_role_permissions.permission, access_role_permissions.scope").
		Joins("JOIN user_access_roles ON user_access_roles.access_role_id = access_role_permissions.role_id").
		Where("user_access_roles.user_id = ?", userID).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	perms := make(map[string]string, len(rows))
	for _, r := range rows {
		models.MergePermission(perms, r.Permission, r.Scope)
	}
	return perms, nil
}

// PermissionScope renvoie la portée de la permission pour l'utilisateur de la requête :
// models.ScopeAll, models.ScopeGroups ou "" s'il ne l'a pas. L'admin a tout, partout.
func PermissionScope(c *gin.Context, permission string) string {
	if c.GetString("role") == "admin" {
		return models.ScopeAll
	}
	perms, _ := c.Get(CtxPermissions)
	if m, ok := perms.(map[string]string); ok {
		return m[permission]
	}
	return ""
}

// IsUnrestricted : la requête agit sans restriction de groupe, parce que l'utilisateur est
// admin ou que RequirePermission lui a reconnu la permission sur tout le site.
func IsUnrestricted(c *gin.Context) bool {
	return c.GetString("role") == "admin" || c.GetBool(ctxUnrestricted)
}

// HasPermissionEverywhere : l'utilisateur a la permission sur tout le site (l'admin l'a).
func HasPermissionEverywhere(c *gin.Context, permission string) bool {
	return PermissionScope(c, permission) == models.ScopeAll
}

// PermissionGroupIDs renvoie les groupes où l'utilisateur exerce la permission de portée
// « ses groupes » : ceux dont il est membre, hors groupe par défaut. Vide sans la permission ;
// sans objet pour la portée « tout le site » (voir HasPermissionEverywhere).
func PermissionGroupIDs(c *gin.Context, permission string) []uint {
	if PermissionScope(c, permission) == models.ScopeGroups {
		return memberGroupIDs(c)
	}
	return []uint{}
}

// GetManagedGroupIDs renvoie les groupes de la requête posés par RequirePermission (portée
// « ses groupes ») ; vide ailleurs.
func GetManagedGroupIDs(c *gin.Context) []uint {
	if ids, ok := c.Get(ctxManagedGroupIDs); ok {
		if list, ok := ids.([]uint); ok {
			return list
		}
	}
	return []uint{}
}

func memberGroupIDs(c *gin.Context) []uint {
	ids, _ := c.Get(CtxMemberGroupIDs)
	if list, ok := ids.([]uint); ok {
		return list
	}
	return nil
}

// RequirePermission laisse passer l'admin et les utilisateurs qui ont la permission.
//   - Portée « tout le site » : la requête est marquée sans restriction (IsUnrestricted).
//   - Portée « ses groupes » : les groupes de la requête (GetManagedGroupIDs) sont ceux dont
//     l'utilisateur est membre, hors groupe par défaut ; sans aucun, il n'a rien à gérer.
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch PermissionScope(c, permission) {
		case models.ScopeAll:
			c.Set(ctxUnrestricted, true)
			c.Next()
			return
		case models.ScopeGroups:
			groups := PermissionGroupIDs(c, permission)
			if len(groups) > 0 {
				c.Set(ctxManagedGroupIDs, groups)
				c.Next()
				return
			}
		}

		apierror.Respond(c, http.StatusForbidden, "Permission requise")
		c.Abort()
	}
}

// RequireAnyPermission laisse passer l'admin et qui a au moins une des permissions, quelle qu'en soit la portée. Pour les routes sans notion de groupe
// (envoi d'un média, par exemple) utiles à plusieurs modules.
func (am *AuthMiddleware) RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") == "admin" {
			c.Next()
			return
		}
		for _, p := range permissions {
			if PermissionScope(c, p) != "" {
				c.Next()
				return
			}
		}
		apierror.Respond(c, http.StatusForbidden, "Permission requise")
		c.Abort()
	}
}

// RequirePermissionParam applique RequirePermission à la permission nommée par le paramètre
// de requête « permission » (resources.manage par défaut) : une même route sert plusieurs
// modules, chacun avec ses propres groupes (liste des groupes gérés, par exemple).
func (am *AuthMiddleware) RequirePermissionParam() gin.HandlerFunc {
	return func(c *gin.Context) {
		permission := c.DefaultQuery("permission", models.PermResourcesManage)
		if _, ok := models.LookupPermission(permission); !ok {
			apierror.Respond(c, http.StatusBadRequest, "Permission inconnue")
			c.Abort()
			return
		}
		am.RequirePermission(permission)(c)
	}
}

// RequirePermissionEverywhere exige la permission sur tout le site (ou le rôle admin) :
// routes /admin dont les handlers ne filtrent pas par groupe.
func (am *AuthMiddleware) RequirePermissionEverywhere(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if PermissionScope(c, permission) != models.ScopeAll {
			apierror.Respond(c, http.StatusForbidden, "Permission requise sur tout le site")
			c.Abort()
			return
		}
		c.Set(ctxUnrestricted, true)
		c.Next()
	}
}
