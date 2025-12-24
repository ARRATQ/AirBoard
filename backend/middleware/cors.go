package middleware

import (
	"net/url"
	"strings"
	"time"

	"airboard/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SecureCORSConfig configuration CORS sécurisée
type SecureCORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// SetupCORS configure les règles CORS sécurisées pour l'API
func SetupCORS(cfg *config.Config) gin.HandlerFunc {
	// Créer une configuration CORS sécurisée
	secureConfig := createSecureCORSConfig(cfg)

	corsConfig := cors.Config{
		AllowOrigins:     secureConfig.AllowedOrigins,
		AllowMethods:     secureConfig.AllowedMethods,
		AllowHeaders:     secureConfig.AllowedHeaders,
		ExposeHeaders:    secureConfig.ExposedHeaders,
		AllowCredentials: secureConfig.AllowCredentials,
		MaxAge:           secureConfig.MaxAge,
	}

	// Ajouter un validateur d'origine personnalisé
	corsConfig.AllowOriginFunc = func(origin string) bool {
		return validateOrigin(origin, secureConfig.AllowedOrigins)
	}

	return cors.New(corsConfig)
}

// createSecureCORSConfig crée une configuration CORS sécurisée
func createSecureCORSConfig(cfg *config.Config) *SecureCORSConfig {
	// Valider et filtrer les origins autorisées
	allowedOrigins := validateAndFilterOrigins(cfg.Server.Origins)

	// Si aucune origin valide, utiliser seulement les origins de développement sécurisées
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{
			"http://localhost:3000", // React dev
			"http://localhost:5173", // Vite dev
			"http://localhost:8080", // Swagger
		}
	}

	return &SecureCORSConfig{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",     // Lecture sécurisée
			"POST",    // Création (protégée par CSRF)
			"PUT",     // Mise à jour (protégée par CSRF)
			"PATCH",   // Modification partielle (protégée par CSRF)
			"DELETE",  // Suppression (protégée par CSRF)
			"OPTIONS", // Preflight requis
		},
		AllowedHeaders: []string{
			"Origin",           // Requis pour CORS
			"Content-Type",     // JSON/form-data
			"Accept",           // Types de réponse acceptés
			"Authorization",    // Bearer token (sécurisé)
			"X-CSRF-Token",     // Protection CSRF
			"X-Requested-With", // AJAX requests
		},
		ExposedHeaders: []string{
			"Content-Length", // Informations de taille
			"X-Total-Count",  // Pagination info
		},
		AllowCredentials: shouldAllowCredentials(allowedOrigins),
		MaxAge:           time.Duration(3600) * time.Second, // 1 heure max (réduit pour sécurité)
	}
}

// validateAndFilterOrigins valide et filtre les origins pour la sécurité
func validateAndFilterOrigins(origins []string) []string {
	var validOrigins []string

	for _, origin := range origins {
		if isValidSecureOrigin(origin) {
			validOrigins = append(validOrigins, origin)
		}
	}

	return validOrigins
}

// isValidSecureOrigin vérifie si une origin est sécurisée
func isValidSecureOrigin(origin string) bool {
	// Parser l'URL
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Vérifier le scheme (HTTPS requis en production, HTTP acceptable en dev)
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	// Rejeter les origins avec des ports non standards (sauf localhost)
	if u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" {
		if u.Port() != "" && u.Port() != "80" && u.Port() != "443" {
			return false
		}
	}

	// Rejeter les origins avec des sous-domaines dangereux
	hostname := strings.ToLower(u.Hostname())
	dangerousPatterns := []string{
		".tk", ".ml", ".ga", // TLDs souvent utilisés pour phishing
		"localhost", // Autorisé seulement pour développement
	}

	for _, pattern := range dangerousPatterns {
		if strings.HasSuffix(hostname, pattern) && hostname != pattern {
			return false
		}
	}

	// Rejeter les origins avec des caractères dangereux
	dangerousChars := []string{"<", ">", "\"", "'", "(", ")", ";"}
	for _, char := range dangerousChars {
		if strings.Contains(origin, char) {
			return false
		}
	}

	return true
}

// validateOrigin valide une origin contre la whitelist
func validateOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
		// Gérer les wildcards pour les sous-domaines
		if strings.HasPrefix(allowed, "https://*.") {
			domain := strings.TrimPrefix(allowed, "https://*.")
			if strings.HasSuffix(origin, "https://"+domain) {
				return true
			}
		}
		if strings.HasPrefix(allowed, "http://*.") {
			domain := strings.TrimPrefix(allowed, "http://*.")
			if strings.HasSuffix(origin, "http://"+domain) {
				return true
			}
		}
	}
	return false
}

// shouldAllowCredentials détermine si les credentials doivent être autorisés
func shouldAllowCredentials(origins []string) bool {
	// Autoriser les credentials seulement si on a des origins spécifiques et sécurisées
	if len(origins) == 0 {
		return false
	}

	// Vérifier que toutes les origins sont sécurisées
	for _, origin := range origins {
		if !isOriginProductionReady(origin) {
			return false
		}
	}

	return true
}

// isOriginProductionReady vérifie si une origin est prête pour la production
func isOriginProductionReady(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// En production, préférer HTTPS
	if u.Scheme == "https" {
		return true
	}

	// HTTP acceptable seulement pour localhost en développement
	return u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"
}
