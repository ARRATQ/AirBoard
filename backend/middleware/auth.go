package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"airboard/config"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type AuthMiddleware struct {
	config    *config.Config
	db        *gorm.DB
	userCache *userStatusCache
}

func NewAuthMiddleware(cfg *config.Config, db *gorm.DB) *AuthMiddleware {
	am := &AuthMiddleware{config: cfg, db: db}
	am.userCache = newUserStatusCache(userStatusTTL, am.loadUserStatus)
	return am
}

// loadUserStatus relit en base le rôle, l'état actif et les groupes administrés.
func (am *AuthMiddleware) loadUserStatus(ctx context.Context, userID uint) (userStatus, error) {
	var user models.User
	err := am.db.WithContext(ctx).
		Select("id", "role", "is_active").
		First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return userStatus{}, errUserUnavailable
	}
	if err != nil {
		return userStatus{}, fmt.Errorf("lecture de l'utilisateur %d: %w", userID, err)
	}

	managedGroupIDs := []uint{}
	if err := am.db.WithContext(ctx).Table("group_admins").
		Where("user_id = ?", userID).
		Pluck("group_id", &managedGroupIDs).Error; err != nil {
		return userStatus{}, fmt.Errorf("lecture des groupes administrés de %d: %w", userID, err)
	}

	return userStatus{Role: user.Role, IsActive: user.IsActive, ManagedGroupIDs: managedGroupIDs}, nil
}

// RequireAuth middleware pour vérifier l'authentification
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		var tokenString string

		if authHeader != "" {
			// Extraire le token (format: "Bearer <token>")
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   "Unauthorized",
					Message: "Format de token invalide",
					Code:    http.StatusUnauthorized,
				})
				c.Abort()
				return
			}
		} else {
			// Essayer de récupérer le token depuis les query params (pour WebSocket)
			tokenString = c.Query("token")
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error:   "Unauthorized",
					Message: "Token d'autorisation manquant",
					Code:    http.StatusUnauthorized,
				})
				c.Abort()
				return
			}
		}

		// Vérifier le token
		claims, err := am.verifyToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Token invalide ou expiré",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Le rôle et l'état du compte viennent de la base (cache court), pas du JWT :
		// un utilisateur rétrogradé ou désactivé perd ses droits sans attendre l'expiration.
		status, err := am.userCache.Get(c.Request.Context(), claims.UserID)
		if err != nil && !errors.Is(err, errUserUnavailable) {
			log.Printf("[Auth] Impossible de vérifier l'utilisateur %d: %v", claims.UserID, err)
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Error:   "Service Unavailable",
				Message: "Vérification du compte impossible, réessayez",
				Code:    http.StatusServiceUnavailable,
			})
			c.Abort()
			return
		}
		if err != nil || !status.IsActive {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "Unauthorized",
				Message: "Compte désactivé ou supprimé",
				Code:    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", status.Role)
		c.Set("email", claims.Email)
		c.Set("managed_group_ids", status.ManagedGroupIDs)

		c.Next()
	}
}

// RequireAdmin middleware pour vérifier les droits administrateur
func (am *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "Droits administrateur requis",
				Code:    http.StatusForbidden,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GenerateToken génère un token JWT
func (am *AuthMiddleware) GenerateToken(user *models.User) (string, error) {
	// Charger les groupes administrés pour tous les utilisateurs
	var managedGroupIDs []uint
	am.db.Table("group_admins").
		Where("user_id = ?", user.ID).
		Pluck("group_id", &managedGroupIDs)

	claims := jwt.MapClaims{
		"user_id":           user.ID,
		"username":          user.Username,
		"role":              user.Role,
		"email":             user.Email,
		"managed_group_ids": managedGroupIDs,
		"exp":               time.Now().Add(time.Hour * time.Duration(am.config.JWT.TokenExpirationHours)).Unix(),
		"iat":               time.Now().Unix(),
		"type":              tokenTypeAccess,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.JWT.Secret))
}

// GenerateRefreshToken génère un refresh token
func (am *AuthMiddleware) GenerateRefreshToken(user *models.User) (string, error) {
	// Charger les groupes administrés pour tous les utilisateurs
	var managedGroupIDs []uint
	am.db.Table("group_admins").
		Where("user_id = ?", user.ID).
		Pluck("group_id", &managedGroupIDs)

	claims := jwt.MapClaims{
		"user_id":           user.ID,
		"username":          user.Username,
		"role":              user.Role,
		"email":             user.Email,
		"managed_group_ids": managedGroupIDs,
		"exp":               time.Now().Add(time.Hour * 24 * time.Duration(am.config.JWT.RefreshExpirationDays)).Unix(),
		"iat":               time.Now().Unix(),
		"type":              tokenTypeRefresh,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.JWT.Secret))
}

// verifyToken vérifie un access token. Un refresh token est refusé, même bien signé.
// Les access tokens émis avant l'ajout du champ "type" restent acceptés.
func (am *AuthMiddleware) verifyToken(tokenString string) (*models.Claims, error) {
	claims, err := am.parseSignedClaims(tokenString)
	if err != nil {
		return nil, err
	}
	if tokenType, _ := claims["type"].(string); tokenType != "" && tokenType != tokenTypeAccess {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claimsToUser(claims)
}

// VerifyRefreshToken vérifie un refresh token.
func (am *AuthMiddleware) VerifyRefreshToken(tokenString string) (*models.Claims, error) {
	claims, err := am.parseSignedClaims(tokenString)
	if err != nil {
		// Signature invalide : le secret JWT a probablement changé depuis l'émission.
		if errors.Is(err, jwt.ErrSignatureInvalid) || errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, fmt.Errorf("secret_jwt_changed: refresh token signé avec un ancien secret")
		}
		return nil, err
	}
	if tokenType, _ := claims["type"].(string); tokenType != tokenTypeRefresh {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claimsToUser(claims)
}

// parseSignedClaims vérifie la signature HMAC et l'expiration d'un JWT.
func (am *AuthMiddleware) parseSignedClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(am.config.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	// jwt v5 vérifie "exp" seulement s'il est présent : un jeton sans expiration est refusé.
	if _, ok := claims["exp"].(float64); !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// claimsToUser extrait les informations utilisateur sans paniquer sur un champ absent.
func claimsToUser(claims jwt.MapClaims) (*models.Claims, error) {
	rawID, ok := claims["user_id"].(float64)
	if !ok || rawID <= 0 {
		return nil, jwt.ErrTokenInvalidClaims
	}

	managedGroupIDs := []uint{}
	if mgids, ok := claims["managed_group_ids"].([]interface{}); ok {
		for _, id := range mgids {
			if idFloat, ok := id.(float64); ok {
				managedGroupIDs = append(managedGroupIDs, uint(idFloat))
			}
		}
	}

	username, _ := claims["username"].(string)
	role, _ := claims["role"].(string)
	email, _ := claims["email"].(string)

	return &models.Claims{
		UserID:          uint(rawID),
		Username:        username,
		Role:            role,
		Email:           email,
		ManagedGroupIDs: managedGroupIDs,
	}, nil
}
