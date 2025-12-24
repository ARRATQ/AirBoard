package handlers

import (
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "MISSING_PARAMETERS",
			Message: "entity_type et entity_id sont requis",
			Code:    http.StatusBadRequest,
		})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_ENTITY_ID",
			Message: "entity_id doit être un nombre valide",
			Code:    http.StatusBadRequest,
		})
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
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
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "DATABASE_ERROR",
					Message: "Erreur lors de la suppression du feedback",
					Code:    http.StatusInternalServerError,
				})
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
				c.JSON(http.StatusInternalServerError, models.ErrorResponse{
					Error:   "DATABASE_ERROR",
					Message: "Erreur lors de la mise à jour du feedback",
					Code:    http.StatusInternalServerError,
				})
				return
			}
			c.JSON(http.StatusOK, models.SuccessResponse{
				Message: "Feedback mis à jour avec succès",
				Data:    existingFeedback,
			})
			return
		}
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Erreur lors de la vérification du feedback existant",
			Code:    http.StatusInternalServerError,
		})
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
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Erreur lors de la création du feedback",
			Code:    http.StatusInternalServerError,
		})
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "MISSING_PARAMETERS",
			Message: "entity_type et entity_id sont requis",
			Code:    http.StatusBadRequest,
		})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_ENTITY_ID",
			Message: "entity_id doit être un nombre valide",
			Code:    http.StatusBadRequest,
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Utilisateur non authentifié",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Supprimer le feedback
	result := h.DB.Where("user_id = ? AND entity_type = ? AND entity_id = ?",
		userID, entityType, entityID).Delete(&models.Feedback{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Erreur lors de la suppression du feedback",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "FEEDBACK_NOT_FOUND",
			Message: "Aucun feedback trouvé",
			Code:    http.StatusNotFound,
		})
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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "MISSING_PARAMETERS",
			Message: "entity_type et entity_id sont requis",
			Code:    http.StatusBadRequest,
		})
		return
	}

	entityID, err := strconv.ParseUint(entityIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "INVALID_ENTITY_ID",
			Message: "entity_id doit être un nombre valide",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var feedbacks []models.Feedback
	if err := h.DB.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Preload("User").
		Order("created_at DESC").
		Find(&feedbacks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "DATABASE_ERROR",
			Message: "Erreur lors de la récupération des feedbacks",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"feedbacks": feedbacks,
		"total":     len(feedbacks),
	})
}
