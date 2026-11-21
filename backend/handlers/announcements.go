package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"airboard/apierror"
	"airboard/config"
	"airboard/models"
	"airboard/services"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type AnnouncementHandler struct {
	db *gorm.DB
}

func NewAnnouncementHandler(db *gorm.DB) *AnnouncementHandler {
	return &AnnouncementHandler{db: db}
}

// GetAllAnnouncements retourne toutes les annonces (admin)
func (h *AnnouncementHandler) GetAllAnnouncements(c *gin.Context) {
	var announcements []models.Announcement

	if err := h.db.Order("priority DESC, created_at DESC").Find(&announcements).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des annonces")
		return
	}

	c.JSON(http.StatusOK, announcements)
}

// GetActiveAnnouncements retourne uniquement les annonces actives et valides (pour les users)
func (h *AnnouncementHandler) GetActiveAnnouncements(c *gin.Context) {
	var announcements []models.Announcement
	now := time.Now()

	query := h.db.Where("is_active = ?", true)

	// Filtrer par dates si elles sont définies
	query = query.Where("(start_date IS NULL OR start_date <= ?)", now)
	query = query.Where("(end_date IS NULL OR end_date >= ?)", now)

	if err := query.Order("priority DESC, created_at DESC").Find(&announcements).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des annonces")
		return
	}

	c.JSON(http.StatusOK, announcements)
}

// GetAnnouncement retourne une annonce spécifique
func (h *AnnouncementHandler) GetAnnouncement(c *gin.Context) {
	id := c.Param("id")
	var announcement models.Announcement

	if err := h.db.First(&announcement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Annonce non trouvée")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération de l'annonce")
		return
	}

	c.JSON(http.StatusOK, announcement)
}

// CreateAnnouncement crée une nouvelle annonce
func (h *AnnouncementHandler) CreateAnnouncement(c *gin.Context) {
	var req models.AnnouncementRequest

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	if endBeforeStart(req.StartDate, req.EndDate) {
		apierror.Respond(c, http.StatusBadRequest, msgEndBeforeStart)
		return
	}

	announcement := models.Announcement{
		Title:     req.Title,
		Content:   req.Content,
		Type:      req.Type,
		Priority:  req.Priority,
		IsActive:  req.IsActive,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	if err := createKeepingFalse(h.db, c, &announcement, "is_active"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création de l'annonce")
		return
	}

	// Envoyer une notification email si l'annonce est active
	if announcement.IsActive {
		utils.SafeGo("email-announcement", func() {
			emailService := services.NewEmailService(h.db, config.LoadConfig())
			// Les annonces sont globales - pas de groupes cibles
			if err := emailService.SendNotification("announcement", announcement.ID, []uint{}); err != nil {
				log.Printf("[Email] Échec de l'envoi de la notification annonce: %v", err)
			}
		})
	}

	c.JSON(http.StatusCreated, announcement)
}

// UpdateAnnouncement met à jour une annonce existante
func (h *AnnouncementHandler) UpdateAnnouncement(c *gin.Context) {
	id := c.Param("id")
	var announcement models.Announcement

	if err := h.db.First(&announcement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Annonce non trouvée")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération de l'annonce")
		return
	}

	var req models.AnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	if endBeforeStart(req.StartDate, req.EndDate) {
		apierror.Respond(c, http.StatusBadRequest, msgEndBeforeStart)
		return
	}

	// Mise à jour des champs
	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.Type = req.Type
	announcement.Priority = req.Priority
	announcement.IsActive = req.IsActive
	announcement.StartDate = req.StartDate
	announcement.EndDate = req.EndDate

	if err := h.db.Save(&announcement).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la mise à jour de l'annonce")
		return
	}

	c.JSON(http.StatusOK, announcement)
}

// DeleteAnnouncement supprime une annonce (soft delete)
func (h *AnnouncementHandler) DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	if err := h.db.Delete(&models.Announcement{}, idInt).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression de l'annonce")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Annonce supprimée avec succès",
	})
}
