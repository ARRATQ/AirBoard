package handlers

import (
	"airboard/apierror"
	"airboard/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FeedbackHandler struct {
	DB *gorm.DB
}

func NewFeedbackHandler(db *gorm.DB) *FeedbackHandler {
	return &FeedbackHandler{DB: db}
}

// GetFeedbackStats récupère les statistiques de feedback pour une entité
func (h *FeedbackHandler) GetFeedbackStats(c *gin.Context) {
	entityType := c.Query("entity_type") // news, application, event
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		apierror.Respond(c, http.StatusBadRequest, "entity_type et entity_id sont requis")
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "entity_id doit être un nombre valide")
		return
	}

	// Compter les feedbacks positifs
	var positiveCount int64
	h.DB.Model(&models.Feedback{}).
		Where("entity_type = ? AND entity_id = ? AND feedback_type = ?", entityType, entityID, "positive").
		Count(&positiveCount)

	// Compter les feedbacks négatifs
	var negativeCount int64
	h.DB.Model(&models.Feedback{}).
		Where("entity_type = ? AND entity_id = ? AND feedback_type = ?", entityType, entityID, "negative").
		Count(&negativeCount)

	stats := models.FeedbackStats{
		EntityType:    entityType,
		EntityID:      uint(entityID),
		PositiveCount: positiveCount,
		NegativeCount: negativeCount,
		TotalCount:    positiveCount + negativeCount,
	}

	// Si l'utilisateur est connecté, récupérer son feedback
	userID, exists := c.Get("user_id")
	if exists {
		var userFeedback models.Feedback
		if err := h.DB.Where("user_id = ? AND entity_type = ? AND entity_id = ?",
			userID, entityType, entityID).First(&userFeedback).Error; err == nil {
			stats.UserFeedback = userFeedback.FeedbackType
		}
	}

	c.JSON(http.StatusOK, stats)
}

// AddFeedback ajoute ou met à jour un feedback
func (h *FeedbackHandler) AddFeedback(c *gin.Context) {
	var req models.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier si l'utilisateur a déjà donné un feedback
	var existingFeedback models.Feedback
	err := h.DB.Where("user_id = ? AND entity_type = ? AND entity_id = ?",
		userID, req.EntityType, req.EntityID).First(&existingFeedback).Error

	if err == nil {
		// Mettre à jour le feedback existant
		if existingFeedback.FeedbackType == req.FeedbackType {
			// Supprimer le feedback si c'est le même (toggle)
			if err := h.DB.Delete(&existingFeedback).Error; err != nil {
				apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression du feedback")
				return
			}
			c.JSON(http.StatusOK, models.SuccessResponse{
				Message: "Feedback supprimé avec succès",
			})
			return
		} else {
			// Changer le type de feedback
			existingFeedback.FeedbackType = req.FeedbackType
			if err := h.DB.Save(&existingFeedback).Error; err != nil {
				apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du feedback")
				return
			}
			c.JSON(http.StatusOK, models.SuccessResponse{
				Message: "Feedback mis à jour avec succès",
				Data:    existingFeedback,
			})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la vérification du feedback existant")
		return
	}

	// Créer un nouveau feedback
	feedback := models.Feedback{
		UserID:       userID.(uint),
		EntityType:   req.EntityType,
		EntityID:     req.EntityID,
		FeedbackType: req.FeedbackType,
	}

	if err := h.DB.Create(&feedback).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création du feedback")
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Feedback ajouté avec succès",
		Data:    feedback,
	})
}

// RemoveFeedback supprime un feedback
func (h *FeedbackHandler) RemoveFeedback(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		apierror.Respond(c, http.StatusBadRequest, "entity_type et entity_id sont requis")
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "entity_id doit être un nombre valide")
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Supprimer le feedback
	result := h.DB.Where("user_id = ? AND entity_type = ? AND entity_id = ?",
		userID, entityType, entityID).Delete(&models.Feedback{})

	if result.Error != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression du feedback")
		return
	}

	if result.RowsAffected == 0 {
		apierror.Respond(c, http.StatusNotFound, "Aucun feedback trouvé")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Feedback supprimé avec succès",
	})
}

// GetAllFeedback récupère tous les feedbacks pour une entité (admin)
func (h *FeedbackHandler) GetAllFeedback(c *gin.Context) {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		apierror.Respond(c, http.StatusBadRequest, "entity_type et entity_id sont requis")
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "entity_id doit être un nombre valide")
		return
	}

	var feedbacks []models.Feedback
	if err := h.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Preload("User").
		Order("created_at DESC").
		Find(&feedbacks).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des feedbacks")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feedbacks": feedbacks,
		"total":     len(feedbacks),
	})
}
