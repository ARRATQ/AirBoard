package middleware

import (
	"airboard/config"
	"airboard/services"
	"crypto/subtle"
	"log"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SSOProxySecretHeader is the header the trusted reverse proxy must set to
// the value of SSO_PROXY_SECRET for X-authentik-* headers to be honoured.
const SSOProxySecretHeader = "X-Airboard-Proxy-Secret"

type ssoProxyCheck int

const (
	ssoProxyNotConfigured ssoProxyCheck = iota
	ssoProxyRejected
	ssoProxyTrusted
)

var warnSSOProxySecretMissing sync.Once

// checkSSOProxySecret compares the secret sent by the proxy with the configured
// one in constant time. With no configured secret, SSO headers are never trusted.
func checkSSOProxySecret(configured, provided string) ssoProxyCheck {
	if configured == "" {
		return ssoProxyNotConfigured
	}
	if subtle.ConstantTimeCompare([]byte(configured), []byte(provided)) != 1 {
		return ssoProxyRejected
	}
	return ssoProxyTrusted
}

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

		// SECURITY: les headers X-authentik-* peuvent être envoyés par n'importe quel client.
		// Seul le reverse proxy connaît SSO_PROXY_SECRET : on exige ce secret partagé.
		clientIP := c.ClientIP()
		switch checkSSOProxySecret(m.config.SSO.ProxySecret, c.GetHeader(SSOProxySecretHeader)) {
		case ssoProxyNotConfigured:
			warnSSOProxySecretMissing.Do(func() {
				log.Printf("[SECURITY] SSO_ENABLED=true mais SSO_PROXY_SECRET est vide : headers SSO ignorés. "+
					"Configurez le secret dans le proxy (header %s) et dans SSO_PROXY_SECRET.", SSOProxySecretHeader)
			})
			c.Next()
			return
		case ssoProxyRejected:
			log.Printf("[SECURITY] Headers SSO rejetés (secret proxy absent ou invalide) depuis %s (email: %s, username: %s)",
				clientIP, email, username)
			c.JSON(403, gin.H{
				"error":   "Forbidden",
				"message": "Headers SSO non autorisés depuis cette source",
			})
			c.Abort()
			return
		}

		log.Printf("[SSO] Headers Authentik détectés pour: %s (%s) depuis le proxy de confiance (%s)", username, email, clientIP)

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
