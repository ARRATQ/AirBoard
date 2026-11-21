package handlers

import (
	"net/http"
	"strconv"

	"airboard/apierror"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FavoritesHandler struct {
	db *gorm.DB
}

func NewFavoritesHandler(db *gorm.DB) *FavoritesHandler {
	return &FavoritesHandler{db: db}
}

// GetUserFavorites retourne la liste des applications favorites de l'utilisateur
func (h *FavoritesHandler) GetUserFavorites(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var user models.User
	if err := h.db.Preload("Favorites", "is_active = ?", true).First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des favoris")
		return
	}

	c.JSON(http.StatusOK, user.Favorites)
}

// AddFavorite ajoute une application aux favoris de l'utilisateur
func (h *FavoritesHandler) AddFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	var requestData struct {
		ApplicationID uint `json:"application_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}

	// Vérifier que l'application existe et est active
	var application models.Application
	if err := h.db.Where("id = ? AND is_active = ?", requestData.ApplicationID, true).First(&application).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Application non trouvée ou inactive")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération de l'utilisateur")
		return
	}

	// Vérifier si l'application est déjà dans les favoris
	var count int64
	h.db.Model(&user).Where("id = ?", application.ID).Association("Favorites").Count()
	if count > 0 {
		apierror.Respond(c, http.StatusConflict, "Cette application est déjà dans vos favoris")
		return
	}

	// Ajouter l'application aux favoris
	if err := h.db.Model(&user).Association("Favorites").Append(&application); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de l'ajout aux favoris")
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Application ajoutée aux favoris avec succès",
		Data:    application,
	})
}

// RemoveFavorite retire une application des favoris de l'utilisateur
func (h *FavoritesHandler) RemoveFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer l'ID de l'application depuis l'URL
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID d'application invalide")
		return
	}

	// Vérifier que l'application existe
	var application models.Application
	if err := h.db.First(&application, appID).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Application non trouvée")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération de l'utilisateur")
		return
	}

	// Retirer l'application des favoris
	if err := h.db.Model(&user).Association("Favorites").Delete(&application); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression du favori")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Application retirée des favoris avec succès",
	})
}

// IsFavorite vérifie si une application est dans les favoris de l'utilisateur
func (h *FavoritesHandler) IsFavorite(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer l'ID de l'application depuis l'URL
	appIDStr := c.Param("id")
	appID, err := strconv.ParseUint(appIDStr, 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID d'application invalide")
		return
	}

	// Récupérer l'utilisateur
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération de l'utilisateur")
		return
	}

	// Vérifier si l'application est dans les favoris
	var count int64
	h.db.Model(&user).Where("id = ?", appID).Association("Favorites").Count()

	c.JSON(http.StatusOK, gin.H{
		"is_favorite": count > 0,
	})
}
