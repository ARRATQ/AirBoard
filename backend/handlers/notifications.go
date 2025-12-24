package handlers

import (
	"airboard/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

// GetNotifications récupère les notifications d'un utilisateur
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Paramètres de pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	onlyUnread := c.Query("unread") == "true"

	offset := (page - 1) * pageSize

	query := h.db.Where("user_id = ?", userID)

	if onlyUnread {
		query = query.Where("is_read = ?", false)
	}

	var notifications []models.Notification
	var total int64

	query.Model(&models.Notification{}).Count(&total)

	if err := query.
		Order("priority DESC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "fetch_error",
			Message: "Erreur lors de la récupération des notifications",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       notifications,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetUnreadCount récupère le nombre de notifications non lues
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var count int64
	if err := h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "count_error",
			Message: "Erreur lors du comptage des notifications",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

// MarkAsRead marque une notification comme lue
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "ID de notification invalide",
			Code:    http.StatusBadRequest,
		})
		return
	}

	now := time.Now()
	result := h.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_error",
			Message: "Erreur lors de la mise à jour de la notification",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Notification non trouvée",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Notification marquée comme lue",
	})
}

// MarkAllAsRead marque toutes les notifications comme lues
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	now := time.Now()
	if err := h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_error",
			Message: "Erreur lors de la mise à jour des notifications",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Toutes les notifications ont été marquées comme lues",
	})
}

// DeleteNotification supprime une notification
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	notificationID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "ID de notification invalide",
			Code:    http.StatusBadRequest,
		})
		return
	}

	result := h.db.Where("id = ? AND user_id = ?", notificationID, userID).
		Delete(&models.Notification{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_error",
			Message: "Erreur lors de la suppression de la notification",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "not_found",
			Message: "Notification non trouvée",
			Code:    http.StatusNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Notification supprimée",
	})
}

// DeleteAllRead supprime toutes les notifications lues
func (h *NotificationHandler) DeleteAllRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := h.db.Where("user_id = ? AND is_read = ?", userID, true).
		Delete(&models.Notification{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "delete_error",
			Message: "Erreur lors de la suppression des notifications",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Notifications lues supprimées",
	})
}

// GetNotificationStats récupère les statistiques des notifications
func (h *NotificationHandler) GetNotificationStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var stats models.NotificationStats

	// Total notifications
	h.db.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalNotifications)

	// Unread count
	h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&stats.UnreadCount)

	// Read count
	h.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, true).
		Count(&stats.ReadCount)

	c.JSON(http.StatusOK, stats)
}

// CreateNotification crée une nouvelle notification (usage interne ou admin)
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req models.NotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Données invalides: " + err.Error(),
			Code:    http.StatusBadRequest,
		})
		return
	}

	notification := models.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Category:  req.Category,
		Title:     req.Title,
		Message:   req.Message,
		Icon:      req.Icon,
		Color:     req.Color,
		Priority:  req.Priority,
		ActionURL: req.ActionURL,
		Metadata:  req.Metadata,
	}

	if err := h.db.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "create_error",
			Message: "Erreur lors de la création de la notification",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// CreateNotificationForUser crée une notification pour un utilisateur (fonction helper)
func CreateNotificationForUser(db *gorm.DB, userID uint, notifType, category, title, message, icon, color, actionURL string, priority int) error {
	notification := models.Notification{
		UserID:    userID,
		Type:      notifType,
		Category:  category,
		Title:     title,
		Message:   message,
		Icon:      icon,
		Color:     color,
		Priority:  priority,
		ActionURL: actionURL,
	}

	return db.Create(&notification).Error
}

// CreateNotificationForUsers crée une notification pour plusieurs utilisateurs
func CreateNotificationForUsers(db *gorm.DB, userIDs []uint, notifType, category, title, message, icon, color, actionURL string, priority int) error {
	notifications := make([]models.Notification, 0, len(userIDs))

	for _, userID := range userIDs {
		notifications = append(notifications, models.Notification{
			UserID:    userID,
			Type:      notifType,
			Category:  category,
			Title:     title,
			Message:   message,
			Icon:      icon,
			Color:     color,
			Priority:  priority,
			ActionURL: actionURL,
		})
	}

	return db.Create(&notifications).Error
}
