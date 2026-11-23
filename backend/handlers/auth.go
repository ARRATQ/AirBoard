package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"time"

	"airboard/apierror"
	"airboard/config"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Verrouillage d'un compte toutes IP confondues : freine une attaque distribuée.
// Il ne s'applique qu'aux IP ayant déjà échoué sur ce compte (voir Login).
const (
	accountLockoutMaxAttempts = 10
	accountLockoutDuration    = 15 * time.Minute
)

type AuthHandler struct {
	db                  *gorm.DB
	authMiddleware      *middleware.AuthMiddleware
	signupEnabled       bool
	notificationService *services.NotificationService
	authSecurity        *utils.AuthSecurityManager // échecs par IP + identifiant
	accountLockout      *utils.AuthSecurityManager // échecs par identifiant, toutes IP confondues
	bcryptCost          int
	gamificationService *services.GamificationService
	dummyHashOnce       sync.Once
	dummyHash           []byte
}

func NewAuthHandler(db *gorm.DB, authMiddleware *middleware.AuthMiddleware, signupEnabled bool, cfg *config.Config, gs *services.GamificationService) *AuthHandler {
	return &AuthHandler{
		db:                  db,
		authMiddleware:      authMiddleware,
		signupEnabled:       signupEnabled,
		notificationService: services.NewNotificationService(db),
		authSecurity:        utils.NewAuthSecurityManager(),
		accountLockout:      utils.NewAuthSecurityManagerWithLimits(accountLockoutMaxAttempts, accountLockoutDuration),
		bcryptCost:          cfg.Security.BcryptCost,
		gamificationService: gs,
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
		apierror.Respond(c, http.StatusBadRequest, "Données de connexion invalides")
		return
	}

	// Rechercher l'utilisateur avec ses relations
	var user models.User
	userErr := h.db.Preload("Groups").Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error

	// Le compteur suit le compte réel (nom ou email donnent la même clé), sinon la saisie
	clientIP := c.ClientIP()
	accountKey := "name:" + strings.ToLower(strings.TrimSpace(req.Username))
	if userErr == nil {
		accountKey = fmt.Sprintf("id:%d", user.ID)
	}
	identifier := clientIP + "|" + accountKey

	// Verrou IP + compte : cette IP a trop échoué sur ce compte
	if isLocked, remaining := h.authSecurity.CheckFailedLogin(identifier); isLocked {
		respondLoginLocked(c, remaining)
		return
	}
	// Verrou du compte : ne bloque que les IP ayant déjà échoué sur ce compte, pour
	// qu'un attaquant ne puisse pas empêcher le vrai titulaire de se connecter
	if isLocked, remaining := h.accountLockout.CheckFailedLogin(accountKey); isLocked && h.authSecurity.HasRecentFailures(identifier) {
		respondLoginLocked(c, remaining)
		return
	}

	// Compte inconnu et mauvais mot de passe sont traités pareil : même réponse, même
	// compteur, et même coût bcrypt (pas d'indice de timing sur l'existence du compte)
	passwordHash := []byte(user.Password)
	if userErr != nil {
		passwordHash = h.dummyPasswordHash()
	}
	if bcrypt.CompareHashAndPassword(passwordHash, []byte(req.Password)) != nil || userErr != nil {
		_, ipRemaining := h.authSecurity.RecordFailedLogin(identifier)
		_, accountRemaining := h.accountLockout.RecordFailedLogin(accountKey)
		if remaining := max(ipRemaining, accountRemaining); remaining > 0 {
			log.Printf("[Auth] Login locked for %q from IP %s for %v", accountKey, clientIP, remaining)
			respondLoginLocked(c, remaining)
			return
		}

		apierror.Respond(c, http.StatusUnauthorized, "Nom d'utilisateur ou mot de passe incorrect")
		return
	}

	// Vérifier le compte actif (après le mot de passe, pour ne rien révéler aux inconnus)
	if !user.IsActive {
		apierror.Respond(c, http.StatusUnauthorized, "Compte désactivé")
		return
	}

	// Enregistrer la connexion réussie et nettoyer les tentatives échouées
	h.authSecurity.RecordSuccessfulLogin(identifier)
	h.accountLockout.RecordSuccessfulLogin(accountKey)
	log.Printf("[Auth] Successful login for %s from IP %s", req.Username, clientIP)

	// Mettre à jour la date de dernière connexion
	now := time.Now()
	if err := h.db.Model(&user).Update("last_login", now).Error; err != nil {
		log.Printf("Erreur lors de la mise à jour de la dernière connexion: %v", err)
		// Ne pas bloquer la connexion pour cette erreur
	}

	// Gamification XP (Daily Login)
	if err := h.gamificationService.AwardXP(user.ID, 50, "daily_login", "", req.Timezone); err != nil {
		log.Printf("[Gamification] Error awarding daily login XP for user %d: %v", user.ID, err)
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
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du token")
		return
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du refresh token")
		return
	}

	// Permissions et groupes couverts (nécessaires au frontend)
	attachPermissions(h.db, &user)

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// dummyPasswordHash renvoie un hash bcrypt jetable, au même coût que les vrais,
// comparé quand le compte n'existe pas.
func (h *AuthHandler) dummyPasswordHash() []byte {
	h.dummyHashOnce.Do(func() {
		hash, err := bcrypt.GenerateFromPassword([]byte("airboard-dummy-password"), h.bcryptCost)
		if err != nil {
			log.Printf("[Auth] Impossible de générer le hash factice: %v", err)
			return
		}
		h.dummyHash = hash
	})
	return h.dummyHash
}

// respondLoginLocked répond 429 avec le temps d'attente restant.
func respondLoginLocked(c *gin.Context, remaining time.Duration) {
	minutes := int(math.Ceil(remaining.Minutes()))
	apierror.Respond(c, http.StatusTooManyRequests, fmt.Sprintf("Trop de tentatives échouées. Réessayez dans %d minute(s)", minutes))
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
		apierror.Respond(c, http.StatusForbidden, "L'inscription est actuellement désactivée")
		return
	}

	// Vérifier si l'inscription est activée dans les paramètres de l'application
	var settings models.AppSettings
	if err := h.db.First(&settings).Error; err == nil {
		if !settings.SignupEnabled {
			apierror.Respond(c, http.StatusForbidden, "L'inscription est actuellement désactivée")
			return
		}
	}

	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données d'inscription invalides")
		return
	}

	// Valider la force du mot de passe
	if err := h.authSecurity.ValidatePassword(req.Password); err != nil {
		apierror.RespondCode(c, http.StatusBadRequest, apierror.CodeWeakPassword, fmt.Sprintf("Mot de passe trop faible: %v", err))
		return
	}

	// Vérifier si l'utilisateur existe déjà
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		apierror.Respond(c, http.StatusConflict, "Nom d'utilisateur ou email déjà utilisé")
		return
	}

	// Hasher le mot de passe avec coût sécurisé (12 minimum - OWASP 2025)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.bcryptCost)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors du hashage du mot de passe")
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
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création de l'utilisateur")
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
	h.db.Preload("Groups").First(&user, user.ID)

	// Générer les tokens
	token, err := h.authMiddleware.GenerateToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du token")
		return
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du refresh token")
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
		apierror.Respond(c, http.StatusBadRequest, "Refresh token manquant")
		return
	}

	// Vérifier le refresh token
	claims, err := h.authMiddleware.VerifyRefreshToken(req.RefreshToken)
	if err != nil {
		// Si c'est une erreur de secret JWT changé, c'est probablement un redémarrage du serveur
		if strings.Contains(err.Error(), "secret_jwt_changed") {
			log.Printf("[Auth] Refresh token signé avec un ancien secret JWT (redémarrage serveur détecté)")
			apierror.RespondCode(c, http.StatusUnauthorized, apierror.CodeJWTSecretChanged, "Le serveur a redémarré avec un nouveau secret JWT. Veuillez vous reconnecter.")
			return
		}

		apierror.Respond(c, http.StatusUnauthorized, "Refresh token invalide ou expiré")
		return
	}

	// Récupérer l'utilisateur avec ses relations
	var user models.User
	if err := h.db.Preload("Groups").First(&user, claims.UserID).Error; err != nil {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non trouvé")
		return
	}

	// Vérifier que le compte est actif
	if !user.IsActive {
		apierror.Respond(c, http.StatusUnauthorized, "Compte désactivé")
		return
	}

	// Générer de nouveaux tokens
	newToken, err := h.authMiddleware.GenerateToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du token")
		return
	}

	newRefreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du refresh token")
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
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var user models.User
	if err := h.db.Preload("Groups").First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Permissions et groupes couverts (nécessaires au frontend)
	attachPermissions(h.db, &user)

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
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier l'ancien mot de passe
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		apierror.Respond(c, http.StatusUnauthorized, "Ancien mot de passe incorrect")
		return
	}

	// Hasher le nouveau mot de passe avec coût sécurisé (12 minimum - OWASP 2025)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), h.bcryptCost)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors du hashage du mot de passe")
		return
	}

	// Mettre à jour le mot de passe
	user.Password = string(hashedPassword)
	if err := h.db.Save(&user).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du mot de passe")
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
		apierror.Respond(c, http.StatusUnauthorized, "SSO non disponible ou non configuré")
		return
	}

	// Récupérer l'utilisateur SSO du contexte
	ssoUserInterface, exists := c.Get("sso_user")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur SSO non trouvé dans le contexte")
		return
	}

	ssoUser, ok := ssoUserInterface.(*models.User)
	if !ok {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des informations SSO")
		return
	}

	// Vérifier que le compte est actif
	if !ssoUser.IsActive {
		apierror.Respond(c, http.StatusUnauthorized, "Compte désactivé")
		return
	}

	// Générer les tokens JWT pour l'utilisateur SSO
	token, err := h.authMiddleware.GenerateToken(ssoUser)
	if err != nil {
		log.Printf("[SSO] Erreur lors de la génération du token: %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du token")
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
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la génération du refresh token")
		return
	}

	// Recharger l'utilisateur avec ses groupes
	var user models.User
	if err := h.db.Preload("Groups").First(&user, ssoUser.ID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des informations utilisateur")
		return
	}

	// Permissions et groupes couverts (nécessaires au frontend)
	attachPermissions(h.db, &user)

	// Gamification XP (Daily Login)
	ssoTimezone := c.Query("timezone")
	if err := h.gamificationService.AwardXP(user.ID, 50, "daily_login", "", ssoTimezone); err != nil {
		log.Printf("[Gamification] Error awarding daily login XP for SSO user %d: %v", user.ID, err)
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
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
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
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du profil")
		return
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").First(&user, user.ID)

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
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer le fichier
	file, err := c.FormFile("avatar")
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Aucun fichier fourni")
		return
	}

	// Vérifier la taille (max 5MB)
	if file.Size > 5*1024*1024 {
		apierror.Respond(c, http.StatusBadRequest, "Le fichier est trop volumineux (max 5MB)")
		return
	}

	// Vérifier le type réel du fichier (octets de tête), pas le Content-Type envoyé par le client
	contentType, err := sniffImageType(file)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Type de fichier non autorisé (JPG, PNG, GIF, WEBP uniquement)")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Créer le dossier avatars s'il n'existe pas
	avatarDir := "./uploads/avatars"
	if err := utils.EnsureDir(avatarDir); err != nil {
		log.Printf("Erreur lors de la création du dossier avatars: %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la sauvegarde de l'avatar")
		return
	}

	// Générer un nom de fichier unique (extension déduite du contenu réel)
	ext := avatarImageTypes[contentType]
	filename := fmt.Sprintf("avatar_%d_%d%s", user.ID, time.Now().Unix(), ext)
	filepath := fmt.Sprintf("%s/%s", avatarDir, filename)

	// Sauvegarder le fichier
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		log.Printf("Erreur lors de la sauvegarde du fichier avatar: %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la sauvegarde de l'avatar")
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
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du profil")
		return
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").First(&user, user.ID)

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// avatarImageTypes associe les types d'image acceptés pour un avatar à leur extension.
var avatarImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// sniffImageType détecte le type d'image à partir du contenu du fichier.
func sniffImageType(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("ouverture du fichier: %w", err)
	}
	defer f.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", fmt.Errorf("lecture du fichier: %w", err)
	}
	contentType := http.DetectContentType(head[:n])
	if _, ok := avatarImageTypes[contentType]; !ok {
		return "", fmt.Errorf("type d'image non autorisé: %s", contentType)
	}
	return contentType, nil
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
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
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
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du profil")
		return
	}

	// Recharger l'utilisateur avec ses relations
	h.db.Preload("Groups").First(&user, user.ID)

	// Masquer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, user)
}
