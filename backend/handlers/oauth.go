package handlers

import (
	"airboard/apierror"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OAuthHandler struct {
	db                  *gorm.DB
	authMiddleware      *middleware.AuthMiddleware
	stateManager        *utils.OAuthStateManager
	gamificationService *services.GamificationService
}

func NewOAuthHandler(db *gorm.DB, authMiddleware *middleware.AuthMiddleware, gs *services.GamificationService) *OAuthHandler {
	return &OAuthHandler{
		db:                  db,
		authMiddleware:      authMiddleware,
		stateManager:        utils.NewOAuthStateManager(),
		gamificationService: gs,
	}
}

// GetEnabledProviders retourne la liste des fournisseurs OAuth activés (sans secrets)
func (h *OAuthHandler) GetEnabledProviders(c *gin.Context) {
	var providers []models.OAuthProvider
	if err := h.db.Where("is_enabled = ?", true).Find(&providers).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch OAuth providers")
		return
	}

	// Convertir en version publique (sans secrets)
	publicProviders := make([]models.OAuthProviderPublic, len(providers))
	for i, p := range providers {
		publicProviders[i] = models.OAuthProviderPublic{
			ID:           p.ID,
			ProviderName: p.ProviderName,
			DisplayName:  p.DisplayName,
			Icon:         p.Icon,
			IsEnabled:    p.IsEnabled,
			RedirectURI:  p.RedirectURI,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": publicProviders,
	})
}

// GetAllProviders retourne tous les fournisseurs OAuth (admin uniquement)
func (h *OAuthHandler) GetAllProviders(c *gin.Context) {
	var providers []models.OAuthProvider
	if err := h.db.Find(&providers).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch OAuth providers")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
	})
}

// UpdateProvider met à jour un fournisseur OAuth (admin uniquement)
func (h *OAuthHandler) UpdateProvider(c *gin.Context) {
	providerID := c.Param("id")

	var req models.OAuthProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	var provider models.OAuthProvider
	if err := h.db.First(&provider, providerID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "OAuth provider not found")
		return
	}

	// Mettre à jour les champs
	provider.DisplayName = req.DisplayName
	provider.Icon = req.Icon
	provider.IsEnabled = req.IsEnabled
	provider.ClientID = req.ClientID
	if req.ClientSecret != "" {
		provider.ClientSecret = req.ClientSecret
	}
	provider.RedirectURI = req.RedirectURI
	provider.AuthURL = req.AuthURL
	provider.TokenURL = req.TokenURL
	provider.UserInfoURL = req.UserInfoURL
	provider.Scopes = req.Scopes

	if err := h.db.Save(&provider).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to update OAuth provider")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "OAuth provider updated successfully",
		Data:    provider,
	})
}

// InitiateOAuth démarre le flux OAuth pour un fournisseur avec protection CSRF renforcée
func (h *OAuthHandler) InitiateOAuth(c *gin.Context) {
	providerName := c.Param("provider")

	var provider models.OAuthProvider
	if err := h.db.Where("provider_name = ? AND is_enabled = ?", providerName, true).First(&provider).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "OAuth provider not found or disabled")
		return
	}

	// Générer un state et nonce sécurisés
	state, nonce, err := h.stateManager.GenerateState(providerName, provider.ClientID)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to generate secure state")
		return
	}

	// Construire l'URL d'autorisation sécurisée
	authURL, err := utils.SecureOAuthURL(provider.AuthURL, map[string]string{
		"client_id":     provider.ClientID,
		"redirect_uri":  provider.RedirectURI,
		"response_type": "code",
		"scope":         provider.Scopes,
	}, state, nonce)

	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to construct authorization URL")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state, // Pour debug, à retirer en production
		"nonce":    nonce, // Pour debug, à retirer en production
	})
}

// OAuthCallback gère le callback OAuth
func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	providerName := c.Param("provider")

	// Accepter le code, state et nonce depuis query params (GET) ou body (POST)
	var code, state, nonce, timezone string
	if c.Request.Method == "POST" {
		var req struct {
			Code     string `json:"code" binding:"required"`
			State    string `json:"state"`
			Nonce    string `json:"nonce"`
			Timezone string `json:"timezone"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("[OAuth] Error parsing POST body: %v", err)
			apierror.Respond(c, http.StatusBadRequest, "Invalid request body")
			return
		}
		code = req.Code
		state = req.State
		nonce = req.Nonce
		timezone = req.Timezone
	} else {
		code = c.Query("code")
		state = c.Query("state")
		nonce = c.Query("nonce")
		timezone = c.Query("timezone")
	}

	log.Printf("[OAuth] Callback received for provider %s - code present: %v, state present: %v, nonce present: %v",
		providerName, code != "", state != "", nonce != "")

	if code == "" {
		apierror.Respond(c, http.StatusBadRequest, "Missing authorization code")
		return
	}

	// Récupérer le provider pour la validation
	var provider models.OAuthProvider
	if err := h.db.Where("provider_name = ? AND is_enabled = ?", providerName, true).First(&provider).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "OAuth provider not found or disabled")
		return
	}

	// SECURITY: state et nonce sont OBLIGATOIRES (protection CSRF / login forcé).
	// Le frontend les conserve en sessionStorage et Microsoft renvoie toujours le state.
	if state == "" || nonce == "" {
		log.Printf("[OAuth] Callback rejected for %s: missing state or nonce", providerName)
		apierror.RespondCode(c, http.StatusBadRequest, apierror.CodeOAuthStateMismatch, "Missing state or nonce parameter")
		return
	}

	validatedNonce, err := h.stateManager.ValidateState(state, providerName, provider.ClientID)
	if err != nil {
		log.Printf("[OAuth] State validation failed: %v", err)
		apierror.RespondCode(c, http.StatusForbidden, apierror.CodeOAuthStateMismatch, "Invalid or expired state parameter")
		return
	}

	if err := h.stateManager.ValidateNonce(state, nonce); err != nil {
		log.Printf("[OAuth] Nonce validation failed: %v", err)
		apierror.RespondCode(c, http.StatusForbidden, apierror.CodeOAuthStateMismatch, "Invalid nonce parameter")
		return
	}

	if validatedNonce != nonce {
		log.Printf("[OAuth] Nonce mismatch for %s", providerName)
		apierror.RespondCode(c, http.StatusForbidden, apierror.CodeOAuthStateMismatch, "Nonce validation failed")
		return
	}

	log.Printf("[OAuth] State and nonce validation successful for %s", providerName)

	// Valider la callback URL pour sécurité supplémentaire
	if err := utils.ValidateOAuthCallbackURL(provider.RedirectURI, c.Request.Host); err != nil {
		log.Printf("[OAuth] Callback URL validation failed: %v", err)
		// Pour le développement, on peut logger mais continuer
		if gin.Mode() == gin.DebugMode {
			log.Printf("[OAuth] Continuing in debug mode despite callback URL validation failure")
		} else {
			apierror.Respond(c, http.StatusForbidden, "Callback URL validation failed")
			return
		}
	}

	// Échanger le code contre un token
	log.Printf("[OAuth] Exchanging code for token with %s (redirect_uri: %s)...", provider.ProviderName, provider.RedirectURI)
	token, err := h.exchangeCodeForToken(provider, code)
	if err != nil {
		log.Printf("[OAuth] ❌ Error exchanging code for token: %v", err)
		errMsg := err.Error()
		// Extract a meaningful message from Microsoft's error response
		if strings.Contains(errMsg, "redirect_uri") || strings.Contains(errMsg, "AADSTS") {
			log.Printf("[OAuth] ❌ Likely redirect_uri mismatch. Stored redirect_uri: %s", provider.RedirectURI)
		}
		apierror.Respond(c, http.StatusInternalServerError, "Failed to exchange code for token")
		return
	}
	log.Printf("[OAuth] ✅ Token exchange successful")

	// Récupérer les informations utilisateur
	log.Printf("[OAuth] Fetching user info from %s...", provider.ProviderName)
	userInfo, err := h.getUserInfo(provider, token)
	if err != nil {
		log.Printf("[OAuth] ❌ Error getting user info: %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to get user information")
		return
	}
	log.Printf("[OAuth] ✅ User info retrieved: %v", userInfo["mail"])

	// Créer ou récupérer l'utilisateur
	log.Printf("[OAuth] Finding or creating user...")
	user, err := h.findOrCreateOAuthUser(provider.ProviderName, userInfo)
	if err != nil {
		log.Printf("[OAuth] ❌ Error finding or creating user: %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to create or find user")
		return
	}
	log.Printf("[OAuth] ✅ User found/created: %s (%s)", user.Username, user.Email)

	// Mettre à jour la date de dernière connexion
	now := time.Now()
	if err := h.db.Model(&user).Update("last_login", now).Error; err != nil {
		log.Printf("[OAuth] Erreur lors de la mise à jour de la dernière connexion: %v", err)
		// Ne pas bloquer la connexion pour cette erreur
	}

	// Générer les tokens JWT
	jwtToken, err := h.authMiddleware.GenerateToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to generate JWT token")
		return
	}

	refreshToken, err := h.authMiddleware.GenerateRefreshToken(&user)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	// Recharger l'utilisateur avec les groupes et les groupes administrés
	if err := h.db.Preload("Groups").First(&user, user.ID).Error; err != nil {
		log.Printf("Error loading user groups: %v", err)
	}

	// Groupes administrés (nécessaires au frontend)
	attachPermissions(h.db, &user)

	user.Password = ""

	// Gamification XP (Daily Login)
	if err := h.gamificationService.AwardXP(user.ID, 50, "daily_login", "", timezone); err != nil {
		log.Printf("[Gamification] Error awarding daily login XP for OAuth user %d: %v", user.ID, err)
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:        jwtToken,
		RefreshToken: refreshToken,
		User:         user,
	})
}

// exchangeCodeForToken échange le code d'autorisation contre un token d'accès
func (h *OAuthHandler) exchangeCodeForToken(provider models.OAuthProvider, code string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", provider.ClientID)
	data.Set("client_secret", provider.ClientSecret)
	data.Set("redirect_uri", provider.RedirectURI)

	req, err := http.NewRequest("POST", provider.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access_token in response")
	}

	return accessToken, nil
}

// getUserInfo récupère les informations utilisateur depuis le provider OAuth
func (h *OAuthHandler) getUserInfo(provider models.OAuthProvider, accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", provider.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed: %s", string(body))
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

// findOrCreateOAuthUser trouve ou crée un utilisateur OAuth
func (h *OAuthHandler) findOrCreateOAuthUser(providerName string, userInfo map[string]interface{}) (models.User, error) {
	var email, firstName, lastName, ssoID string

	// Extraire les informations selon le provider
	switch providerName {
	case "google":
		email, _ = userInfo["email"].(string)
		firstName, _ = userInfo["given_name"].(string)
		lastName, _ = userInfo["family_name"].(string)
		ssoID, _ = userInfo["sub"].(string)
	case "microsoft":
		email, _ = userInfo["mail"].(string)
		if email == "" {
			email, _ = userInfo["userPrincipalName"].(string)
		}
		firstName, _ = userInfo["givenName"].(string)
		lastName, _ = userInfo["surname"].(string)
		ssoID, _ = userInfo["id"].(string)
	}

	if email == "" || ssoID == "" {
		return models.User{}, fmt.Errorf("missing required user information")
	}

	// Chercher l'utilisateur existant par email d'abord, puis par SSO
	var user models.User
	err := h.db.Where("email = ?", email).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Chercher par SSO ID si pas trouvé par email
		err = h.db.Where("sso_provider = ? AND sso_id = ?", providerName, ssoID).First(&user).Error
	}

	if err == gorm.ErrRecordNotFound {
		// Créer un nouvel utilisateur
		username := strings.Split(email, "@")[0]

		// S'assurer que le username est unique
		baseUsername := username
		counter := 1
		for {
			var existingUser models.User
			if err := h.db.Where("username = ?", username).First(&existingUser).Error; err == gorm.ErrRecordNotFound {
				break
			}
			username = fmt.Sprintf("%s%d", baseUsername, counter)
			counter++
		}

		user = models.User{
			Username:    username,
			Email:       email,
			FirstName:   firstName,
			LastName:    lastName,
			Role:        "user",
			IsActive:    true,
			SSOProvider: providerName,
			SSOID:       ssoID,
		}

		if err := h.db.Create(&user).Error; err != nil {
			return models.User{}, err
		}

		// Ajouter au groupe par défaut (depuis les paramètres de l'application)
		var settings models.AppSettings
		if err := h.db.First(&settings).Error; err == nil && settings.DefaultGroupID != nil {
			var defaultGroup models.Group
			if err := h.db.First(&defaultGroup, *settings.DefaultGroupID).Error; err == nil {
				h.db.Model(&user).Association("Groups").Append(&defaultGroup)
				log.Printf("[OAuth] User added to default group: %s", defaultGroup.Name)
			}
		} else {
			// Fallback: chercher un groupe "Common" ou "common" si les paramètres ne définissent pas de groupe par défaut
			var commonGroup models.Group
			if err := h.db.Where("LOWER(name) = ?", "common").First(&commonGroup).Error; err == nil {
				h.db.Model(&user).Association("Groups").Append(&commonGroup)
				log.Printf("[OAuth] User added to fallback group: %s", commonGroup.Name)
			}
		}

		log.Printf("[OAuth] New user created: %s (%s) via %s", user.Email, user.Username, providerName)
	} else if err != nil {
		return models.User{}, err
	} else {
		// Utilisateur existant trouvé - mettre à jour les informations
		updated := false

		// Mettre à jour les infos SSO si l'utilisateur était créé manuellement
		if user.SSOProvider == "" || user.SSOID == "" {
			user.SSOProvider = providerName
			user.SSOID = ssoID
			updated = true
			log.Printf("[OAuth] Linking existing user %s to %s SSO", user.Email, providerName)
		}

		if user.FirstName != firstName && firstName != "" {
			user.FirstName = firstName
			updated = true
		}
		if user.LastName != lastName && lastName != "" {
			user.LastName = lastName
			updated = true
		}

		if updated {
			h.db.Save(&user)
		}
		log.Printf("[OAuth] Existing user logged in: %s (%s) via %s", user.Email, user.Username, providerName)
	}

	return user, nil
}

// generateRandomState function removed - replaced by OAuthStateManager
