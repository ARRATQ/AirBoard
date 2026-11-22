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

func memberGroupIDs(c *gin.Context) []uint {
	ids, _ := c.Get(CtxMemberGroupIDs)
	if list, ok := ids.([]uint); ok {
		return list
	}
	return nil
}

// RequirePermission laisse passer l'admin et les utilisateurs qui ont la permission.
//   - Portée « tout le site » : la requête est marquée sans restriction (IsUnrestricted).
//   - Portée « ses groupes » : les groupes gérés de la requête (GetManagedGroupIDs) deviennent
//     les groupes dont l'utilisateur est membre, ajoutés aux groupes qu'il administre encore
//     (ancienne table group_admins, gardée pendant la transition).
//
// Les anciens admins de groupe passent toujours, avec leurs groupes administrés.
func (am *AuthMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch PermissionScope(c, permission) {
		case models.ScopeAll:
			c.Set(ctxUnrestricted, true)
			c.Next()
			return
		case models.ScopeGroups:
			groups := mergeUintSets(GetManagedGroupIDs(c), memberGroupIDs(c))
			if len(groups) > 0 {
				c.Set("managed_group_ids", groups)
				c.Next()
				return
			}
		}

		if len(GetManagedGroupIDs(c)) > 0 {
			c.Next()
			return
		}

		apierror.Respond(c, http.StatusForbidden, "Permission requise")
		c.Abort()
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

// mergeUintSets renvoie l'union dédoublonnée des deux listes, dans l'ordre d'apparition.
func mergeUintSets(a, b []uint) []uint {
	seen := make(map[uint]struct{}, len(a)+len(b))
	out := make([]uint, 0, len(a)+len(b))
	for _, list := range [][]uint{a, b} {
		for _, id := range list {
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}
