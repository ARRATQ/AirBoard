package middleware

import (
	"log"
	"net/http"
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
		// Utiliser l'IP du client comme identifiant
		key := c.ClientIP()

		// Si l'utilisateur est authentifié, utiliser son ID pour un meilleur contrôle
		if userID, exists := c.Get("user_id"); exists {
			key = userID.(string)
		}

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
		c.Header("X-RateLimit-Limit", string(rune(context.Limit)))
		c.Header("X-RateLimit-Remaining", string(rune(context.Remaining)))
		c.Header("X-RateLimit-Reset", string(rune(context.Reset)))

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

// GlobalRateLimit limite globale pour toutes les requêtes
// 100 requêtes par minute par IP
func GlobalRateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  100,
	}
	return RateLimitMiddleware(rate)
}

// AuthRateLimit limite stricte pour les endpoints d'authentification
// 5 tentatives par minute par IP pour éviter le brute-force
func AuthRateLimit() gin.HandlerFunc {
	rate := limiter.Rate{
		Period: 1 * time.Minute,
		Limit:  5,
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
