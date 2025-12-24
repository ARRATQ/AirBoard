package middleware

import (
	"airboard/config"
	"airboard/services"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SSOMiddleware détecte et traite les headers Authentik pour l'authentification SSO
type SSOMiddleware struct {
	db        *gorm.DB
	config    *config.Config
	ssoMapper *services.SSOMapper
}

// NewSSOMiddleware crée une nouvelle instance de SSOMiddleware
func NewSSOMiddleware(db *gorm.DB, cfg *config.Config) *SSOMiddleware {
	return &SSOMiddleware{
		db:        db,
		config:    cfg,
		ssoMapper: services.NewSSOMapper(db, cfg),
	}
}

// DetectSSO détecte si la requête contient des headers Authentik
func (m *SSOMiddleware) DetectSSO() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Si SSO n'est pas activé, passer
		if !m.config.SSO.Enabled {
			c.Next()
			return
		}

		// Vérifier la présence des headers Authentik
		email := c.GetHeader("X-authentik-email")
		username := c.GetHeader("X-authentik-username")

		// Si pas de headers SSO, continuer normalement (mode classique)
		if email == "" || username == "" {
			c.Next()
			return
		}

		// SECURITY: Valider que les headers SSO proviennent d'une source de confiance
		clientIP := c.ClientIP()
		trustedSources := []string{
			"127.0.0.1",
			"::1",
			"localhost",
			"172.", // Docker networks (commence par 172.)
			"192.168.", // Private networks
			"10.", // Private networks
		}

		isTrusted := false
		for _, trusted := range trustedSources {
			if clientIP == trusted || strings.HasPrefix(clientIP, trusted) {
				isTrusted = true
				break
			}
		}

		if !isTrusted {
			log.Printf("[SECURITY] Tentative de SSO spoofing détectée depuis IP non autorisée: %s (email: %s, username: %s)",
				clientIP, email, username)
			c.JSON(403, gin.H{
				"error": "Forbidden",
				"message": "Headers SSO non autorisés depuis cette source",
			})
			c.Abort()
			return
		}

		log.Printf("[SSO] Headers Authentik détectés pour: %s (%s) depuis IP autorisée: %s", username, email, clientIP)

		// Extraire les informations SSO
		ssoInfo := &services.SSOUserInfo{
			Email:     email,
			Username:  username,
			FirstName: c.GetHeader("X-authentik-name"),
			LastName:  "",
			Groups:    parseGroups(c.GetHeader("X-authentik-groups")),
			SSOID:     c.GetHeader("X-authentik-uid"),
		}

		// Séparer FirstName et LastName si nécessaire
		if ssoInfo.FirstName != "" {
			parts := strings.Split(ssoInfo.FirstName, " ")
			if len(parts) > 1 {
				ssoInfo.FirstName = parts[0]
				ssoInfo.LastName = strings.Join(parts[1:], " ")
			}
		}

		// Synchroniser l'utilisateur
		user, err := m.ssoMapper.SyncUser(ssoInfo)
		if err != nil {
			log.Printf("[SSO] Erreur lors de la synchronisation de l'utilisateur: %v", err)
			c.Next()
			return
		}

		// Stocker les informations SSO dans le contexte pour utilisation ultérieure
		c.Set("sso_user", user)
		c.Set("sso_active", true)

		log.Printf("[SSO] Utilisateur SSO synchronisé: %s (ID: %d, Role: %s)", user.Email, user.ID, user.Role)

		c.Next()
	}
}

// parseGroups parse le header X-authentik-groups qui peut être sous différents formats
func parseGroups(groupsHeader string) []string {
	if groupsHeader == "" {
		return []string{}
	}

	var groups []string

	// Essayer de parser comme liste séparée par des virgules
	for _, group := range strings.Split(groupsHeader, ",") {
		trimmed := strings.TrimSpace(group)
		if trimmed != "" {
			groups = append(groups, trimmed)
		}
	}

	// Si pas de groupes trouvés, essayer avec des pipes (|)
	if len(groups) == 0 {
		for _, group := range strings.Split(groupsHeader, "|") {
			trimmed := strings.TrimSpace(group)
			if trimmed != "" {
				groups = append(groups, trimmed)
			}
		}
	}

	log.Printf("[SSO] Groupes parsés: %v", groups)
	return groups
}
