package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"airboard/config"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db                  *gorm.DB
	authMiddleware      *middleware.AuthMiddleware
	signupEnabled       bool
	notificationService *services.NotificationService
	authSecurity        *utils.AuthSecurityManager
	bcryptCost          int
}

func NewAuthHandler(db *gorm.DB, authMiddleware *middleware.AuthMiddleware, signupEnabled bool, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:                  db,
		authMiddleware:      authMiddleware,
		signupEnabled:       signupEnabled,
		notificationService: services.NewNotificationService(db),
		authSecurity:        utils.NewAuthSecurityManager(),
		bcryptCost:          cfg.Security.BcryptCost,
	}
}

// @Summary Connexion utilisateur
// @Description Authentifie un utilisateur avec username/password
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body models.LoginRequest true "Informations de connexion"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Données de connexion invalides",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Vérifier le lockout avant de tenter la connexion
	clientIP := c.ClientIP()
	identifier := fmt.Sprintf("%s:%s", clientIP, req.Username)

	// Vérifier si l'identifiant est locké
	if isLocked, remaining := h.authSecurity.CheckFailedLogin(identifier); isLocked {
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Error:   "Too Many Requests",
			Message: fmt.Sprintf("Trop de tentatives échouées. Réessayez dans %.0f minutes", remaining.Minutes()),
			Code:    http.StatusTooManyRequests,
		})
		return
	}

	// Rechercher l'utilisateur avec ses relations
	var user models.User
	if err := h.db.Preload("Groups").Preload("AdminOfGroups").Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Nom d'utilisateur ou mot de passe incorrect",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Vérifier le compte actif
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Compte désactivé",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Vérifier le mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Enregistrer la tentative échouée
		if isLocked, remaining := h.authSecurity.RecordFailedLogin(identifier); isLocked {
			log.Printf("[Auth] Account locked for %s after failed attempts: %v", identifier, remaining)
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error:   "Too Many Requests",
				Message: fmt.Sprintf("Compte verrouillé après trop de tentatives échouées. Réessayez dans %.0f minutes", remaining.Minutes()),
				Code:    http.StatusTooManyRequests,
			})
			return
		}

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Nom d'utilisateur ou mot de passe incorrect",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Enregistrer la connexion réussie et nettoyer les tentatives échouées
	h.authSecurity.RecordSuccessfulLogin(identifier)
	log.Printf("[Auth] Successful login for %s from IP %s", req.Username, clientIP)

	// Mettre à jour la date de dernière connexion
	now := time.Now()
	if err := h.db.Model(&user).Update("last_login", now).Error; err != nil {
		log.Printf("Erreur lors de la mise à jour de la dernière connexion: %v", err)
		// Ne pas bloquer la connexion pour cette erreur
	}

	// Créer une notification de connexion (en arrière-plan) - DÉSACTIVÉ
	/*
		go func() {
			if err := h.notificationService.NotifyLogin(user.ID, isFirstLogin); err != nil {
				log.Printf("Erreur lors de la création de la notification de connexion: %v", err)
			}
		}()
	*/

	// Générer les tokens
	token, err := h.authMiddleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du refresh token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// @Summary Inscription utilisateur
// @Description Crée un nouveau compte utilisateur
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body models.RegisterRequest true "Informations d'inscription"
// @Success 201 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// Vérifier si l'inscription est activée via la variable d'environnement
	if !h.signupEnabled {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "Forbidden",
			Message: "L'inscription est actuellement désactivée",
			Code:    http.StatusForbidden,
		})
		return
	}

	// Vérifier si l'inscription est activée dans les paramètres de l'application
	var settings models.AppSettings
	if err := h.db.First(&settings).Error; err == nil {
		if !settings.SignupEnabled {
			c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "Forbidden",
				Message: "L'inscription est actuellement désactivée",
				Code:    http.StatusForbidden,
			})
			return
		}
	}

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Données d'inscription invalides",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Valider la force du mot de passe
	if err := h.authSecurity.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Weak Password",
			Message: fmt.Sprintf("Mot de passe trop faible: %v", err),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Vérifier si l'utilisateur existe déjà
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "Conflict",
			Message: "Nom d'utilisateur ou email déjà utilisé",
			Code:    http.StatusConflict,
		})
		return
	}

	// Hasher le mot de passe avec coût sécurisé (12 minimum - OWASP 2025)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.bcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors du hashage du mot de passe",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Créer l'utilisateur
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      "user", // Par défaut, les nouveaux utilisateurs sont des users normaux
		IsActive:  true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la création de l'utilisateur",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Ajouter l'utilisateur au groupe par défaut configuré
	defaultGroup := GetDefaultGroupFromDB(h.db)
	if defaultGroup != nil {
		// Un groupe par défaut est configuré, l'ajouter à l'utilisateur
		if err := h.db.Model(&user).Association("Groups").Append(defaultGroup); err != nil {
			log.Printf("Erreur lors de l'ajout de l'utilisateur au groupe par défaut: %v", err)
		}
	} else {
		// Aucun groupe par défaut configuré, utiliser "common" comme fallback
		var commonGroup models.Group
		if err := h.db.Where("name = ?", "common").First(&commonGroup).Error; err == nil {
			// Le groupe "common" existe, l'ajouter à l'utilisateur
			if err := h.db.Model(&user).Association("Groups").Append(&commonGroup); err != nil {
				log.Printf("Erreur lors de l'ajout de l'utilisateur au groupe common: %v", err)
			}
		} else {
			// Le groupe "common" n'existe pas, le créer puis ajouter l'utilisateur
			commonGroup = models.Group{
				Name:        "common",
				Description: "Groupe par défaut pour tous les nouveaux utilisateurs",
				Color:       "#3B82F6",
				IsActive:    true,
			}
			if err := h.db.Create(&commonGroup).Error; err == nil {
				if err := h.db.Model(&user).Association("Groups").Append(&commonGroup); err != nil {
					log.Printf("Erreur lors de l'ajout de l'utilisateur au groupe common: %v", err)
				}
			} else {
				log.Printf("Erreur lors de la création du groupe common: %v", err)
			}
		}
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, user.ID)

	// Générer les tokens
	token, err := h.authMiddleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du refresh token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusCreated, models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// @Summary Rafraîchir le token
// @Description Génère un nouveau token à partir d'un refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body map[string]string true "Refresh token"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Refresh token manquant",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Vérifier le refresh token
	claims, err := h.authMiddleware.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		// Si c'est une erreur de secret JWT changé, c'est probablement un redémarrage du serveur
		if strings.Contains(err.Error(), "secret_jwt_changed") {
			log.Printf("[Auth] Refresh token signé avec un ancien secret JWT (redémarrage serveur détecté)")
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "JWT_SECRET_CHANGED",
				Message: "Le serveur a redémarré avec un nouveau secret JWT. Veuillez vous reconnecter.",
				Code:    http.StatusUnauthorized,
			})
			return
		}

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Refresh token invalide ou expiré",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Récupérer l'utilisateur avec ses relations
	var user models.User
	if err := h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Vérifier que le compte est actif
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Compte désactivé",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Générer de nouveaux tokens
	newToken, err := h.authMiddleware.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	newRefreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du refresh token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
		User:         user,
	})
}

// @Summary Profil utilisateur
// @Description Récupère les informations du profil de l'utilisateur connecté
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var user models.User
	if err := h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// @Summary Changer le mot de passe
// @Description Change le mot de passe de l'utilisateur connecté
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changePassword body models.ChangePasswordRequest true "Nouveau mot de passe"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Données invalides",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Vérifier l'ancien mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Ancien mot de passe incorrect",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Hasher le nouveau mot de passe avec coût sécurisé (12 minimum - OWASP 2025)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), h.bcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors du hashage du mot de passe",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Mettre à jour le mot de passe
	user.Password = string(hashedPassword)
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la mise à jour du mot de passe",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Mot de passe changé avec succès",
	})
}

// @Summary Statut de l'inscription
// @Description Récupère le statut de l'inscription (activée/désactivée)
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /auth/signup/status [get]
func (h *AuthHandler) GetSignupStatus(c *gin.Context) {
	signupEnabled := h.signupEnabled

	// Vérifier également les paramètres de l'application
	var settings models.AppSettings
	if err := h.db.First(&settings).Error; err == nil {
		signupEnabled = signupEnabled && settings.SignupEnabled
	}

	c.JSON(http.StatusOK, gin.H{
		"signup_enabled": signupEnabled,
	})
}

// @Summary Auto-login SSO
// @Description Authentifie automatiquement un utilisateur via les headers Authentik
// @Tags Auth
// @Produce json
// @Success 200 {object} models.LoginResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 503 {object} models.ErrorResponse
// @Router /auth/sso/auto-login [get]
func (h *AuthHandler) SSOAutoLogin(c *gin.Context) {
	// Vérifier si SSO est actif dans le contexte (injecté par le middleware SSO)
	ssoActive, exists := c.Get("sso_active")
	if !exists || !ssoActive.(bool) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "SSO non disponible ou non configuré",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Récupérer l'utilisateur SSO du contexte
	ssoUserInterface, exists := c.Get("sso_user")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur SSO non trouvé dans le contexte",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	ssoUser, ok := ssoUserInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la récupération des informations SSO",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Vérifier que le compte est actif
	if !ssoUser.IsActive {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Compte désactivé",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Générer les tokens JWT pour l'utilisateur SSO
	token, err := h.authMiddleware.GenerateToken(ssoUser)
	if err != nil {
		log.Printf("[SSO] Erreur lors de la génération du token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Mettre à jour la date de dernière connexion
	now := time.Now()
	if err := h.db.Model(ssoUser).Update("last_login", now).Error; err != nil {
		log.Printf("[SSO] Erreur lors de la mise à jour de la dernière connexion: %v", err)
		// Ne pas bloquer la connexion pour cette erreur
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(ssoUser)
	if err != nil {
		log.Printf("[SSO] Erreur lors de la génération du refresh token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la génération du refresh token",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Recharger l'utilisateur avec les groupes et les groupes administrés
	var user models.User
	if err := h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, ssoUser.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la récupération des informations utilisateur",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Masquer le mot de passe (même si vide pour SSO)
	user.Password = ""

	log.Printf("[SSO] Auto-login réussi pour: %s (%s)", user.Email, user.Username)

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// @Summary Mettre à jour le profil
// @Description Met à jour les informations du profil de l'utilisateur connecté
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body models.UpdateProfileRequest true "Informations du profil"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Données invalides",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Mettre à jour les champs
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone
	user.Department = req.Department
	user.JobTitle = req.JobTitle
	user.Location = req.Location

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la mise à jour du profil",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, user.ID)

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// @Summary Upload avatar
// @Description Upload un nouvel avatar pour l'utilisateur connecté
// @Tags Auth
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formance file true "Fichier avatar (JPG, PNG, GIF, max 5MB)"
// @Success 200 {object} models.User
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/avatar [post]
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Récupérer le fichier
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Aucun fichier fourni",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Vérifier la taille (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Le fichier est trop volumineux (max 5MB)",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Vérifier le type MIME
	contentType := file.Header.Get("Content-Type")
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Bad Request",
			Message: "Type de fichier non autorisé (JPG, PNG, GIF, WEBP uniquement)",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Créer le dossier avatars s'il n'existe pas
	avatarDir := "./uploads/avatars"
	if err := utils.EnsureDir(avatarDir); err != nil {
		log.Printf("Erreur lors de la création du dossier avatars: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la sauvegarde de l'avatar",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Générer un nom de fichier unique
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	}
	filename := fmt.Sprintf("avatar_%d_%d%s", user.ID, time.Now().Unix(), ext)
	filepath := fmt.Sprintf("%s/%s", avatarDir, filename)

	// Sauvegarder le fichier
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		log.Printf("Erreur lors de la sauvegarde du fichier avatar: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la sauvegarde de l'avatar",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Supprimer l'ancien avatar si existant (et si c'est un fichier local)
	if user.AvatarURL != "" && strings.HasPrefix(user.AvatarURL, "/uploads/avatars/") {
		oldPath := "." + user.AvatarURL
		if err := utils.RemoveFile(oldPath); err != nil {
			log.Printf("Erreur lors de la suppression de l'ancien avatar: %v", err)
			// On continue quand même
		}
	}

	// Mettre à jour l'URL de l'avatar
	user.AvatarURL = fmt.Sprintf("/uploads/avatars/%s", filename)
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la mise à jour du profil",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, user.ID)

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// @Summary Supprimer l'avatar
// @Description Supprime l'avatar de l'utilisateur connecté
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/avatar [delete]
func (h *AuthHandler) DeleteAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Unauthorized",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Supprimer le fichier avatar si existant (et si c'est un fichier local)
	if user.AvatarURL != "" && strings.HasPrefix(user.AvatarURL, "/uploads/avatars/") {
		oldPath := "." + user.AvatarURL
		if err := utils.RemoveFile(oldPath); err != nil {
			log.Printf("Erreur lors de la suppression de l'avatar: %v", err)
			// On continue quand même
		}
	}

	// Réinitialiser l'URL de l'avatar
	user.AvatarURL = ""
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Internal Server Error",
			Message: "Erreur lors de la mise à jour du profil",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").Preload("AdminOfGroups").First(&user, user.ID)

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, user)
}
