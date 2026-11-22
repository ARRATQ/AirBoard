package handlers

import (
	"airboard/apierror"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CommentHandler struct {
	DB           *gorm.DB
	Gamification *services.GamificationService
}

func NewCommentHandler(db *gorm.DB, gamification *services.GamificationService) *CommentHandler {
	return &CommentHandler{
		DB:           db,
		Gamification: gamification,
	}
}

// GetComments récupère une page des commentaires d'une entité (news, app, event, suggestion)
func (h *CommentHandler) GetComments(c *gin.Context) {
	entityType := c.Query("entity_type") // news, application, event, suggestion
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

	// Vérifier les paramètres de commentaires
	var settings models.CommentSettings
	if err := h.DB.First(&settings).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des paramètres")
		return
	}

	// Vérifier si les commentaires sont activés pour ce type d'entité
	if !settings.CommentsEnabled {
		apierror.RespondCode(c, http.StatusForbidden, apierror.CodeCommentsDisabled, "Les commentaires sont désactivés")
		return
	}

	switch entityType {
	case "news":
		if !settings.NewsCommentsEnabled {
			apierror.RespondCode(c, http.StatusForbidden, apierror.CodeCommentsDisabled, "Les commentaires sur les actualités sont désactivés")
			return
		}
	case "application":
		if !settings.AppCommentsEnabled {
			apierror.RespondCode(c, http.StatusForbidden, apierror.CodeCommentsDisabled, "Les commentaires sur les applications sont désactivés")
			return
		}
	case "event":
		if !settings.EventCommentsEnabled {
			apierror.RespondCode(c, http.StatusForbidden, apierror.CodeCommentsDisabled, "Les commentaires sur les événements sont désactivés")
			return
		}
	case "suggestion":
		// Suggestions utilisent uniquement le flag global comments_enabled
	default:
		apierror.Respond(c, http.StatusBadRequest, "entity_type invalide")
		return
	}

	// Récupérer les commentaires (paginés, plus récents d'abord)
	page := parsePageParams(c, commentPageLimits)
	query := h.DB.Model(&models.Comment{}).Where("entity_type = ? AND entity_id = ?", entityType, entityID)

	// Si modération requise, ne montrer que les commentaires approuvés, sauf aux modérateurs
	// (admin compris)
	if settings.RequireModeration && !canModerateComments(c) {
		query = query.Where("is_approved = ?", true)
	}

	comments, total, err := findCommentPage(query, page, "Moderator")
	if err != nil {
		respondCommentsError(c)
		return
	}
	c.JSON(http.StatusOK, commentPageResponse(comments, total, page))
}

var commentPageLimits = pageLimits{Default: 20, Max: 100}

// findCommentPage compte puis charge une page de commentaires, auteur et relations demandées
// préchargés (après le comptage, qui n'en a pas besoin).
func findCommentPage(query *gorm.DB, page pageParams, preloads ...string) ([]models.Comment, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	query = query.Preload("User")
	for _, relation := range preloads {
		query = query.Preload(relation)
	}
	var comments []models.Comment
	err := query.
		Order("created_at DESC, id DESC").
		Limit(page.Size).Offset(page.Offset()).
		Find(&comments).Error
	return comments, total, err
}

// commentPageResponse garde la clé `comments` lue par le front et ajoute la pagination.
func commentPageResponse(comments []models.Comment, total int64, page pageParams) gin.H {
	return gin.H{
		"comments":    comments,
		"total":       total,
		"page":        page.Page,
		"page_size":   page.Size,
		"total_pages": page.TotalPages(total),
	}
}

func respondCommentsError(c *gin.Context) {
	apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des commentaires")
}

// CreateComment crée un nouveau commentaire
func (h *CommentHandler) CreateComment(c *gin.Context) {
	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	// Récupérer l'utilisateur connecté
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Vérifier les paramètres de commentaires
	var settings models.CommentSettings
	if err := h.DB.First(&settings).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des paramètres")
		return
	}

	if !settings.CommentsEnabled {
		apierror.RespondCode(c, http.StatusForbidden, apierror.CodeCommentsDisabled, "Les commentaires sont désactivés")
		return
	}

	// Vérifier la limite de longueur
	if len(req.Content) > settings.MaxCommentLength {
		apierror.RespondCode(c, http.StatusBadRequest, apierror.CodeCommentTooLong, "Le commentaire est trop long")
		return
	}

	// Créer le commentaire
	comment := models.Comment{
		Content:    req.Content,
		UserID:     userID.(uint),
		EntityType: req.EntityType,
		EntityID:   req.EntityID,
		IsApproved: !settings.RequireModeration, // Auto-approuvé si pas de modération
		IsFlagged:  false,
		ParentID:   req.ParentID,
	}

	// IsApproved porte default:true : GORM omettrait le false à l'insertion et le
	// commentaire serait publié malgré la modération. Le false est réécrit dans la
	// même transaction.
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}
		if settings.RequireModeration {
			return tx.Model(&comment).Update("is_approved", false).Error
		}
		return nil
	})
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création du commentaire")
		return
	}

	// Recharger avec les relations
	if err := h.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération du commentaire")
		return
	}

	// Accorder des points XP pour la création d'un commentaire
	if h.Gamification != nil {
		// 10 XP pour un commentaire
		utils.SafeGo("xp-comment-create", func() {
			err := h.Gamification.AwardXP(userID.(uint), 10, "comment_create", "{}")
			if err != nil {
				// Log error but don't fail request
				log.Printf("[Gamification] Failed to award XP for comment creation: %v", err)
			}

			if req.EntityType == "suggestion" {
				err = h.Gamification.AwardXP(userID.(uint), 8, "suggestion_comment", fmt.Sprintf("{\"suggestion_id\":%d}", req.EntityID))
				if err != nil {
					log.Printf("[Gamification] Failed to award XP for suggestion comment: %v", err)
				}
			}
		})
	}

	c.JSON(http.StatusCreated, models.SuccessResponse{
		Message: "Commentaire créé avec succès",
		Data:    comment,
	})
}

// UpdateComment met à jour un commentaire (uniquement par l'auteur)
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	commentID := c.Param("id")

	var req models.CommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer le commentaire
	var comment models.Comment
	if err := h.DB.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Commentaire non trouvé")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération du commentaire")
		return
	}

	// Vérifier que l'utilisateur est l'auteur ou modérateur
	if comment.UserID != userID.(uint) && !canModerateComments(c) {
		apierror.Respond(c, http.StatusForbidden, "Vous n'avez pas la permission de modifier ce commentaire")
		return
	}

	// Mettre à jour le contenu
	comment.Content = req.Content

	if err := h.DB.Save(&comment).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du commentaire")
		return
	}

	// Recharger avec les relations
	if err := h.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération du commentaire")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Commentaire mis à jour avec succès",
		Data:    comment,
	})
}

// DeleteComment supprime un commentaire (par l'auteur ou admin)
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("id")

	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer le commentaire
	var comment models.Comment
	if err := h.DB.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Commentaire non trouvé")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération du commentaire")
		return
	}

	// Vérifier que l'utilisateur est l'auteur ou modérateur
	if comment.UserID != userID.(uint) && !canModerateComments(c) {
		apierror.Respond(c, http.StatusForbidden, "Vous n'avez pas la permission de supprimer ce commentaire")
		return
	}

	// Supprimer le commentaire
	if err := h.DB.Delete(&comment).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression du commentaire")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Commentaire supprimé avec succès",
	})
}

// ModerateComment modère un commentaire (admin/editor seulement)
func (h *CommentHandler) ModerateComment(c *gin.Context) {
	var req models.ModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "Utilisateur non authentifié")
		return
	}

	// Récupérer le commentaire
	var comment models.Comment
	if err := h.DB.First(&comment, req.CommentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Commentaire non trouvé")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération du commentaire")
		return
	}

	// Mettre à jour le statut de modération
	comment.IsApproved = req.IsApproved
	comment.IsFlagged = req.IsFlagged
	moderatorID := userID.(uint)
	comment.ModeratedBy = &moderatorID
	now := time.Now()
	comment.ModeratedAt = &now

	if err := h.DB.Save(&comment).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la modération du commentaire")
		return
	}

	// Recharger avec les relations
	if err := h.DB.Preload("User").Preload("Moderator").First(&comment, comment.ID).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération du commentaire")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Commentaire modéré avec succès",
		Data:    comment,
	})
}

// GetPendingComments récupère une page des commentaires en attente de modération (admin/editor)
func (h *CommentHandler) GetPendingComments(c *gin.Context) {
	page := parsePageParams(c, commentPageLimits)
	query := h.DB.Model(&models.Comment{}).Where("is_approved = ?", false)
	comments, total, err := findCommentPage(query, page)
	if err != nil {
		respondCommentsError(c)
		return
	}
	response := commentPageResponse(comments, total, page)
	response["targets"] = loadCommentTargets(h.DB, comments)
	c.JSON(http.StatusOK, response)
}

// GetCommentSettings récupère les paramètres de commentaires
func (h *CommentHandler) GetCommentSettings(c *gin.Context) {
	var settings models.CommentSettings
	if err := h.DB.First(&settings).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Créer les paramètres par défaut
			settings = models.CommentSettings{
				CommentsEnabled:      true,
				NewsCommentsEnabled:  true,
				AppCommentsEnabled:   false,
				EventCommentsEnabled: true,
				RequireModeration:    false,
				AllowAnonymous:       false,
				MaxCommentLength:     1000,
			}
			if err := h.DB.Create(&settings).Error; err != nil {
				apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création des paramètres")
				return
			}
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des paramètres")
			return
		}
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateCommentSettings met à jour les paramètres de commentaires (admin seulement)
func (h *CommentHandler) UpdateCommentSettings(c *gin.Context) {
	var req models.CommentSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	var settings models.CommentSettings
	if err := h.DB.First(&settings).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des paramètres")
		return
	}

	// Mettre à jour uniquement les champs fournis
	if req.CommentsEnabled != nil {
		settings.CommentsEnabled = *req.CommentsEnabled
	}
	if req.NewsCommentsEnabled != nil {
		settings.NewsCommentsEnabled = *req.NewsCommentsEnabled
	}
	if req.AppCommentsEnabled != nil {
		settings.AppCommentsEnabled = *req.AppCommentsEnabled
	}
	if req.EventCommentsEnabled != nil {
		settings.EventCommentsEnabled = *req.EventCommentsEnabled
	}
	if req.RequireModeration != nil {
		settings.RequireModeration = *req.RequireModeration
	}
	if req.AllowAnonymous != nil {
		settings.AllowAnonymous = *req.AllowAnonymous
	}
	if req.MaxCommentLength != nil {
		// 0 ou une valeur trop basse rendrait tout commentaire impossible
		if !validCommentLength(*req.MaxCommentLength) {
			apierror.Respond(c, http.StatusBadRequest, fmt.Sprintf("La longueur maximale doit être comprise entre %d et %d caractères", minCommentLengthSetting, maxCommentLengthSetting))
			return
		}
		settings.MaxCommentLength = *req.MaxCommentLength
	}

	if err := h.DB.Save(&settings).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour des paramètres")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Paramètres mis à jour avec succès",
		Data:    settings,
	})
}

// canModerateComments : l'utilisateur modère les commentaires (permission comments.moderate,
// admin compris).
func canModerateComments(c *gin.Context) bool {
	return middleware.HasPermissionEverywhere(c, models.PermCommentsModerate)
}
