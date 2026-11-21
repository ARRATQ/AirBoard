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
