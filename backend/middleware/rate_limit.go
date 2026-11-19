package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"airboard/models"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitMiddleware crée un middleware de rate limiting
func RateLimitMiddleware(rate limiter.Rate) gin.HandlerFunc {
	store := memory.NewStore()
	instance := limiter.New(store, rate)

	return func(c *gin.Context) {
		key := rateLimitKey(c)

		context, err := instance.Get(c.Request.Context(), key)
		if err != nil {
			log.Printf("[RateLimit] Erreur lors de la vérification: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "Internal Server Error",
				Message: "Erreur lors de la vérification du rate limit",
				Code:    http.StatusInternalServerError,
			})
			c.Abort()
			return
		}

		// Ajouter les headers de rate limit
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		// Si la limite est dépassée, retourner une erreur 429
		if context.Reached {
			log.Printf("[RateLimit] Limite atteinte pour %s - IP: %s", key, c.ClientIP())
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Too Many Requests",
				Message: "Trop de requêtes. Veuillez réessayer plus tard.",
				Code:    http.StatusTooManyRequests,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// rateLimitKey identifie le client : l'utilisateur s'il est authentifié, sinon son IP.
// user_id est un uint dans le contexte ; fmt.Sprint évite toute assertion de type.
func rateLimitKey(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	return "ip:" + c.ClientIP()
}

// GlobalRateLimit limite globale pour toutes les requêtes
// 100 requêtes par minute par IP
func GlobalRateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}
	return RateLimitMiddleware(rate)
}

// AuthRateLimit limite les endpoints de connexion par mot de passe et d'inscription :
// 20 requêtes par minute par IP. Le seuil reste large car plusieurs collègues
// peuvent partager l'IP publique du bureau ; le verrouillage par compte
// (handlers.AuthHandler.Login) bloque le brute-force sur un compte précis.
func AuthRateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  20,
	}
	return RateLimitMiddleware(rate)
}

// APIRateLimit limite pour les endpoints API généraux
// 60 requêtes par minute par utilisateur
func APIRateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  60,
	}
	return RateLimitMiddleware(rate)
}

// StrictAPIRateLimit limite stricte pour les opérations sensibles
// 10 requêtes par minute (création, modification, suppression)
func StrictAPIRateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  10,
	}
	return RateLimitMiddleware(rate)
}
