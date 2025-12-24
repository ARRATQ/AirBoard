package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"airboard/models"

	"github.com/gin-gonic/gin"
)

// CSRFToken représente un token CSRF avec son expiration
type CSRFToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// CSRFManager gère les tokens CSRF
type CSRFManager struct {
	tokens map[string]CSRFToken // userID -> CSRFToken
}

// NewCSRFManager crée un nouveau gestionnaire CSRF
func NewCSRFManager() *CSRFManager {
	return &CSRFManager{
		tokens: make(map[string]CSRFToken),
	}
}

// generateToken génère un token CSRF aléatoire
func (csm *CSRFManager) generateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback: utiliser le temps actuel si rand.Read échoue
		hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		return hex.EncodeToString(hash[:])
	}
	return hex.EncodeToString(bytes)
}

// generateCSRFForUser génère un token CSRF pour un utilisateur
func (csm *CSRFManager) generateCSRFForUser(userID uint) string {
	// Nettoyer les anciens tokens expirés
	csm.cleanupExpiredTokens()

	// Générer un nouveau token
	token := csm.generateToken()

	// Stocker avec expiration (1 heure)
	csm.tokens[fmt.Sprintf("%d", userID)] = CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	return token
}

// validateCSRFToken valide un token CSRF
func (csm *CSRFManager) validateCSRFToken(userID uint, token string) bool {
	if token == "" {
		return false
	}

	userToken, exists := csm.tokens[fmt.Sprintf("%d", userID)]
	if !exists {
		return false
	}

	// Vérifier l'expiration
	if time.Now().After(userToken.ExpiresAt) {
		delete(csm.tokens, fmt.Sprintf("%d", userID))
		return false
	}

	// Vérifier la correspondance
	return userToken.Token == token
}

// cleanupExpiredTokens supprime les tokens expirés
func (csm *CSRFManager) cleanupExpiredTokens() {
	now := time.Now()
	for userID, tokenData := range csm.tokens {
		if now.After(tokenData.ExpiresAt) {
			delete(csm.tokens, userID)
		}
	}
}

// CSRFProtection middleware pour protéger les requêtes de mutation
func CSRFProtection(csrfManager *CSRFManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Méthodes qui nécessitent une protection CSRF
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == http.MethodPatch {

			// Ignorer certains endpoints (webhooks, API externes)
			if strings.Contains(c.Request.URL.Path, "/webhook") ||
				strings.Contains(c.Request.URL.Path, "/api/v1/oauth/callback") {
				c.Next()
				return
			}

			// Vérifier la présence du token CSRF
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   "Unauthorized",
					Message: "Token CSRF requis - utilisateur non authentifié",
					Code:    http.StatusUnauthorized,
				})
				c.Abort()
				return
			}

			// Récupérer le token CSRF depuis l'en-tête ou le formulaire
			csrfToken := c.GetHeader("X-CSRF-Token")
			if csrfToken == "" {
				csrfToken = c.PostForm("csrf_token")
			}
			if csrfToken == "" {
				csrfToken = c.Query("csrf_token")
			}

			// Valider le token CSRF
			if !csrfManager.validateCSRFToken(userID.(uint), csrfToken) {
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error:   "Forbidden",
					Message: "Token CSRF invalide ou expiré",
					Code:    http.StatusForbidden,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// CSRFTokenHandler génère un nouveau token CSRF
func CSRFTokenHandler(csrfManager *CSRFManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Vérifier l'authentification
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Authentification requise pour générer un token CSRF",
				Code:    http.StatusUnauthorized,
			})
			return
		}

		// Générer un nouveau token CSRF
		token := csrfManager.generateCSRFForUser(userID.(uint))

		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
			"expires_at": time.Now().Add(time.Hour),
		})
	}
}

// OptionalCSRFProtection middleware optionnel pour CSRF (ne bloque pas si absent)
func OptionalCSRFProtection(csrfManager *CSRFManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Seulement pour les méthodes de mutation
		if c.Request.Method == http.MethodPost ||
			c.Request.Method == http.MethodPut ||
			c.Request.Method == http.MethodDelete ||
			c.Request.Method == http.MethodPatch {

			userID, exists := c.Get("user_id")
			if !exists {
				c.Next()
				return
			}

			// Récupérer le token CSRF
			csrfToken := c.GetHeader("X-CSRF-Token")
			if csrfToken == "" {
				csrfToken = c.PostForm("csrf_token")
			}
			if csrfToken == "" {
				csrfToken = c.Query("csrf_token")
			}

			// Si token présent, le valider
			if csrfToken != "" && !csrfManager.validateCSRFToken(userID.(uint), csrfToken) {
				c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error:   "Forbidden",
					Message: "Token CSRF invalide",
					Code:    http.StatusForbidden,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
