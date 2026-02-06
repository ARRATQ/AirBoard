package handlers

import (
	"airboard/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SettingsHandler struct {
	DB        *gorm.DB
	homeCache *HomeCache
}

func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{
		DB:        db,
		homeCache: homeCache,
	}
}

// GetAppSettings récupère les paramètres de l'application
func (h *SettingsHandler) GetAppSettings(c *gin.Context) {
	var settings models.AppSettings

	// Récupérer les paramètres (il ne devrait y en avoir qu'un seul)
	result := h.DB.First(&settings)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Créer les paramètres par défaut s'ils n'existent pas
			settings = models.AppSettings{
				AppName:         "Airboard",
				AppIcon:         "mdi:view-dashboard",
				DashboardTitle:  "Dashboard",
				WelcomeMessage:  "Welcome to your application portal",
				HomePageMessage: "Discover your personalized workspace",
				SignupEnabled:   true,
			}

			if err := h.DB.Create(&settings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "database_error",
					Message: "Failed to create default settings",
					Code:    http.StatusInternalServerError,
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to fetch app settings",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "App settings retrieved successfully",
		Data:    settings,
	})
}

// UpdateAppSettings met à jour les paramètres de l'application
func (h *SettingsHandler) UpdateAppSettings(c *gin.Context) {
	var request models.AppSettingsRequest

	// Valider les données de la requête
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Récupérer les paramètres existants
	var settings models.AppSettings
	result := h.DB.First(&settings)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Créer de nouveaux paramètres
			signupEnabled := true
			if request.SignupEnabled != nil {
				signupEnabled = *request.SignupEnabled
			}
			settings = models.AppSettings{
				AppName:         request.AppName,
				AppIcon:         request.AppIcon,
				DashboardTitle:  request.DashboardTitle,
				WelcomeMessage:  request.WelcomeMessage,
				HomePageMessage: request.HomePageMessage,
				SignupEnabled:   signupEnabled,
				DefaultGroupID:  request.DefaultGroupID,
			}

			if err := h.DB.Create(&settings).Error; err != nil {
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "database_error",
					Message: "Failed to create app settings",
					Code:    http.StatusInternalServerError,
				})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to fetch app settings",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	} else {
		// Mettre à jour les paramètres existants
		settings.AppName = request.AppName
		settings.AppIcon = request.AppIcon
		settings.DashboardTitle = request.DashboardTitle
		settings.WelcomeMessage = request.WelcomeMessage
		settings.HomePageMessage = request.HomePageMessage
		if request.SignupEnabled != nil {
			settings.SignupEnabled = *request.SignupEnabled
		}
		settings.DefaultGroupID = request.DefaultGroupID

		if err := h.DB.Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to update app settings",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	}

	// Invalidate home cache
	if h.homeCache != nil {
		h.homeCache.InvalidateAppSettings()
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "App settings updated successfully",
		Data:    settings,
	})
}

// ResetAppSettings remet les paramètres aux valeurs par défaut
func (h *SettingsHandler) ResetAppSettings(c *gin.Context) {
	var settings models.AppSettings

	// Récupérer les paramètres existants
	result := h.DB.First(&settings)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch app settings",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Remettre aux valeurs par défaut
	settings.AppName = "Airboard"
	settings.AppIcon = "mdi:view-dashboard"
	settings.DashboardTitle = "Dashboard"
	settings.WelcomeMessage = "Welcome to your application portal"
	settings.HomePageMessage = "Discover your personalized workspace"
	settings.SignupEnabled = true

	if result.Error == gorm.ErrRecordNotFound {
		// Créer de nouveaux paramètres avec les valeurs par défaut
		if err := h.DB.Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to create default settings",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	} else {
		// Mettre à jour les paramètres existants
		if err := h.DB.Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to reset app settings",
				Code:    http.StatusInternalServerError,
			})
			return
		}
	}

	// Invalidate home cache
	if h.homeCache != nil {
		h.homeCache.InvalidateAppSettings()
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "App settings reset to defaults successfully",
		Data:    settings,
	})
}

// GetDefaultGroupFromDB récupère le groupe par défaut depuis les settings
// Retourne nil si aucun groupe par défaut n'est configuré
func GetDefaultGroupFromDB(db *gorm.DB) *models.Group {
	var settings models.AppSettings
	if err := db.First(&settings).Error; err != nil {
		return nil
	}

	if settings.DefaultGroupID == nil {
		return nil
	}

	var group models.Group
	if err := db.First(&group, *settings.DefaultGroupID).Error; err != nil {
		return nil
	}

	return &group
}

// GetHeroMessages récupère tous les messages Hero
func (h *SettingsHandler) GetHeroMessages(c *gin.Context) {
	var messages []models.HeroMessage
	if err := h.DB.Order("created_at DESC").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch hero messages",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Hero messages retrieved successfully",
		Data:    messages,
	})
}

// CreateHeroMessage crée un nouveau message Hero
func (h *SettingsHandler) CreateHeroMessage(c *gin.Context) {
	var request models.HeroMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	message := models.HeroMessage{
		Content:  request.Content,
		Author:   request.Author,
		IsActive: isActive,
	}

	if err := h.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to create hero message",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Invalider le cache
	homeCache.InvalidateHeroMessages()

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Hero message created successfully",
		Data:    message,
	})
}

// UpdateHeroMessage met à jour un message Hero existant
func (h *SettingsHandler) UpdateHeroMessage(c *gin.Context) {
	id := c.Param("id")
	var message models.HeroMessage
	if err := h.DB.First(&message, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Hero message not found",
				Code:    http.StatusNotFound,
			})
		} else {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "database_error",
				Message: "Failed to fetch hero message",
				Code:    http.StatusInternalServerError,
			})
		}
		return
	}

	var request models.HeroMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	message.Content = request.Content
	message.Author = request.Author
	if request.IsActive != nil {
		message.IsActive = *request.IsActive
	}

	if err := h.DB.Save(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update hero message",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Invalider le cache
	homeCache.InvalidateHeroMessages()

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Hero message updated successfully",
		Data:    message,
	})
}

// DeleteHeroMessage supprime un message Hero
func (h *SettingsHandler) DeleteHeroMessage(c *gin.Context) {
	id := c.Param("id")
	if err := h.DB.Delete(&models.HeroMessage{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete hero message",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Invalider le cache
	homeCache.InvalidateHeroMessages()

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Hero message deleted successfully",
	})
}
