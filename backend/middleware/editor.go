package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"airboard/apierror"
)

// RequireEditor vérifie que l'utilisateur a le rôle editor ou admin
func (m *AuthMiddleware) RequireEditor() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			apierror.Respond(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok || (roleStr != "editor" && roleStr != "admin") {
			apierror.Respond(c, http.StatusForbidden, "Editor or Admin role required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireEditorOrAdmin - Alias pour RequireEditor (même comportement)
func (m *AuthMiddleware) RequireEditorOrAdmin() gin.HandlerFunc {
	return m.RequireEditor()
}

// RequireEditorOr laisse passer l'ancien rôle éditeur, sinon délègue au contrôle donné.
// Transitoire : tant que des modules (sondages, modération) dépendent encore de ce rôle.
func (m *AuthMiddleware) RequireEditorOr(next gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") == "editor" {
			c.Next()
			return
		}
		next(c)
	}
}
