package handlers

import (
	"net/http"
	"strconv"

	"airboard/apierror"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

// ChatbotHandler gère les endpoints liés aux chatbots n8n
type ChatbotHandler struct {
	db                  *gorm.DB
	gamificationService *services.GamificationService
}

// NewChatbotHandler crée un nouveau ChatbotHandler
func NewChatbotHandler(db *gorm.DB, gs *services.GamificationService) *ChatbotHandler {
	return &ChatbotHandler{db: db, gamificationService: gs}
}

// GetChatbots retourne la liste paginée de tous les chatbots (admin)
func (h *ChatbotHandler) GetChatbots(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var chatbots []models.Chatbot
	var total int64

	h.db.Model(&models.Chatbot{}).Count(&total)
	if err := h.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&chatbots).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des chatbots")
		return
	}

	totalPages := int(total+int64(limit)-1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       chatbots,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	})
}

// CreateChatbot crée un nouveau chatbot (admin)
func (h *ChatbotHandler) CreateChatbot(c *gin.Context) {
	var req models.ChatbotRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	initialMessages := req.InitialMessages
	if initialMessages == "" {
		initialMessages = "[]"
	}
	icon := req.Icon
	if icon == "" {
		icon = "mdi:robot-outline"
	}
	color := req.Color
	if color == "" {
		color = "#4f46e5"
	}

	hideHeader := false
	if req.HideHeader != nil {
		hideHeader = *req.HideHeader
	}
	showAvatarIntro := false
	if req.ShowAvatarIntro != nil {
		showAvatarIntro = *req.ShowAvatarIntro
	}

	chatbot := models.Chatbot{
		Name:            req.Name,
		Description:     req.Description,
		WebhookURL:      req.WebhookURL,
		Icon:            icon,
		Color:           color,
		WelcomeTitle:    req.WelcomeTitle,
		WelcomeSubtitle: req.WelcomeSubtitle,
		HideHeader:      hideHeader,
		ShowAvatarIntro: showAvatarIntro,
		IsActive:        isActive,
		InitialMessages: initialMessages,
	}

	if err := createKeepingFalse(h.db, c, &chatbot, "is_active"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création du chatbot")
		return
	}

	c.JSON(http.StatusCreated, chatbot)
}

// UpdateChatbot met à jour un chatbot existant (admin)
// Gère aussi le toggle is_active
func (h *ChatbotHandler) UpdateChatbot(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var chatbot models.Chatbot
	if err := h.db.First(&chatbot, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Chatbot non trouvé")
		return
	}

	var req models.ChatbotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	chatbot.Name = req.Name
	chatbot.Description = req.Description
	chatbot.WebhookURL = req.WebhookURL
	chatbot.WelcomeTitle = req.WelcomeTitle
	chatbot.WelcomeSubtitle = req.WelcomeSubtitle
	if req.Icon != "" {
		chatbot.Icon = req.Icon
	}
	if req.Color != "" {
		chatbot.Color = req.Color
	}
	if req.HideHeader != nil {
		chatbot.HideHeader = *req.HideHeader
	}
	if req.ShowAvatarIntro != nil {
		chatbot.ShowAvatarIntro = *req.ShowAvatarIntro
	}
	if req.IsActive != nil {
		chatbot.IsActive = *req.IsActive
	}
	if req.InitialMessages != "" {
		chatbot.InitialMessages = req.InitialMessages
	}

	if err := h.db.Save(&chatbot).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour du chatbot")
		return
	}

	c.JSON(http.StatusOK, chatbot)
}

// DeleteChatbot supprime (soft delete) un chatbot (admin)
func (h *ChatbotHandler) DeleteChatbot(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.db.Delete(&models.Chatbot{}, id).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression du chatbot")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Chatbot supprimé avec succès"})
}

// TrackInteraction enregistre l'utilisation d'un chatbot pour la gamification
func (h *ChatbotHandler) TrackInteraction(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		apierror.Respond(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var chatbot models.Chatbot
	if err := h.db.Where("id = ? AND is_active = ?", id, true).First(&chatbot).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Chatbot non trouvé")
		return
	}

	if err := h.gamificationService.AwardXP(userID, 10, "chatbot_use", ""); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur gamification")
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetActiveChatbots retourne uniquement les chatbots actifs (tous les utilisateurs connectés)
func (h *ChatbotHandler) GetActiveChatbots(c *gin.Context) {
	var chatbots []models.Chatbot
	if err := h.db.Where("is_active = ?", true).Order("created_at ASC").Find(&chatbots).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des chatbots")
		return
	}

	c.JSON(http.StatusOK, chatbots)
}
