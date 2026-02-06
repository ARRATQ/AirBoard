package handlers

import (
	"net/http"
	"strconv"
	"time"

	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
)

// CreateEventGroupAdmin - Créer un événement (Group Admin)
// Peut créer des événements publics OU ciblant uniquement ses groupes gérés
func (h *EventsHandler) CreateEventGroupAdmin(c *gin.Context) {
	var req models.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
	}

	// Récupérer les groupes gérés par ce group admin
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	// Validation : si des groupes cibles sont spécifiés, ils doivent être dans les groupes gérés
	if len(req.TargetGroupIDs) > 0 {
		for _, targetGroupID := range req.TargetGroupIDs {
			isManaged := false
			for _, managedID := range managedGroupIDs {
				if targetGroupID == managedID {
					isManaged = true
					break
				}
			}
			if !isManaged {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Vous ne pouvez cibler que les groupes que vous administrez",
				})
				return
			}
		}
	}

	// Récupérer l'utilisateur connecté
	userID := c.GetUint("user_id")

	// Créer l'événement
	event := models.Event{
		Title:                req.Title,
		Description:          req.Description,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		IsAllDay:             req.IsAllDay,
		Timezone:             req.Timezone,
		IsRecurring:          req.IsRecurring,
		RecurrenceRule:       req.RecurrenceRule,
		RecurrenceEnd:        req.RecurrenceEnd,
		RecurrenceExceptions: req.RecurrenceExceptions,
		Location:             req.Location,
		ExternalLinks:        req.ExternalLinks,
		Color:                req.Color,
		Priority:             req.Priority,
		Status:               req.Status,
		CoverImage:           req.CoverImage,
		IsPublished:          req.IsPublished,
		PublishedAt:          req.PublishedAt,
		CategoryID:           req.CategoryID,
		AuthorID:             userID,
	}

	// Auto-set published_at si publié sans date
	if event.IsPublished && event.PublishedAt == nil {
		now := time.Now()
		event.PublishedAt = &now
	}

	// Créer l'événement en BDD
	if err := h.db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de l'événement"})
		return
	}

	// Associer les tags
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		h.db.Where("id IN ?", req.TagIDs).Find(&tags)
		h.db.Model(&event).Association("Tags").Append(tags)
	}

	// Associer les groupes cibles
	if len(req.TargetGroupIDs) > 0 {
		var groups []models.Group
		h.db.Where("id IN ?", req.TargetGroupIDs).Find(&groups)
		h.db.Model(&event).Association("TargetGroups").Append(groups)
	}

	// Recharger avec les relations
	h.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		First(&event, event.ID)

	// Award Contributor XP
	go h.gamificationService.AwardXP(userID, 150, "event_publish", "")

	c.JSON(http.StatusCreated, event)
}

// UpdateEventGroupAdmin - Modifier un événement (Group Admin)
// Peut modifier ses propres événements OU événements ciblant ses groupes gérés
func (h *EventsHandler) UpdateEventGroupAdmin(c *gin.Context) {
	identifier := c.Param("id")

	var event models.Event
	query := h.db.Preload("TargetGroups")

	// Essayer de traiter l'identifiant comme un ID numérique d'abord
	if eventID, err := strconv.Atoi(identifier); err == nil {
		// C'est un ID numérique valide
		query = query.Where("id = ?", eventID)
	} else {
		// Sinon, c'est un slug
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&event).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Événement non trouvé"})
		return
	}

	// Vérification des permissions
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	// Vérifier si le group admin peut gérer cet événement
	canManage := false

	// 1. Si c'est son propre événement
	if event.AuthorID == userID {
		canManage = true
	} else {
		// 2. Si l'événement cible un de ses groupes gérés
		for _, targetGroup := range event.TargetGroups {
			for _, managedID := range managedGroupIDs {
				if targetGroup.ID == managedID {
					canManage = true
					break
				}
			}
			if canManage {
				break
			}
		}
	}

	if !canManage {
		c.JSON(http.StatusForbidden, gin.H{"error": "Vous ne pouvez modifier que les événements de vos groupes gérés"})
		return
	}

	var req models.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
	}

	// Validation : si des groupes cibles sont spécifiés, ils doivent être dans les groupes gérés
	if len(req.TargetGroupIDs) > 0 {
		for _, targetGroupID := range req.TargetGroupIDs {
			isManaged := false
			for _, managedID := range managedGroupIDs {
				if targetGroupID == managedID {
					isManaged = true
					break
				}
			}
			if !isManaged {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Vous ne pouvez cibler que les groupes que vous administrez",
				})
				return
			}
		}
	}

	// Mettre à jour les champs
	event.Title = req.Title
	event.Description = req.Description
	event.StartDate = req.StartDate
	event.EndDate = req.EndDate
	event.IsAllDay = req.IsAllDay
	event.Timezone = req.Timezone
	event.IsRecurring = req.IsRecurring
	event.RecurrenceRule = req.RecurrenceRule
	event.RecurrenceEnd = req.RecurrenceEnd
	event.RecurrenceExceptions = req.RecurrenceExceptions
	event.Location = req.Location
	event.ExternalLinks = req.ExternalLinks
	event.Color = req.Color
	event.Priority = req.Priority
	event.Status = req.Status
	event.CoverImage = req.CoverImage
	event.IsPublished = req.IsPublished
	event.PublishedAt = req.PublishedAt

	if req.CategoryID != nil {
		event.CategoryID = req.CategoryID
	}

	// Auto-set published_at si publié sans date
	if event.IsPublished && event.PublishedAt == nil {
		now := time.Now()
		event.PublishedAt = &now
	}

	// Sauvegarder
	if err := h.db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour"})
		return
	}

	// Mettre à jour les tags
	h.db.Model(&event).Association("Tags").Clear()
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		h.db.Where("id IN ?", req.TagIDs).Find(&tags)
		h.db.Model(&event).Association("Tags").Append(tags)
	}

	// Mettre à jour les groupes cibles
	h.db.Model(&event).Association("TargetGroups").Clear()
	if len(req.TargetGroupIDs) > 0 {
		var groups []models.Group
		h.db.Where("id IN ?", req.TargetGroupIDs).Find(&groups)
		h.db.Model(&event).Association("TargetGroups").Append(groups)
	}

	// Recharger avec les relations
	h.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		First(&event, event.ID)

	c.JSON(http.StatusOK, event)
}

// DeleteEventGroupAdmin - Supprimer un événement (Group Admin)
// Peut supprimer ses propres événements OU événements ciblant ses groupes gérés
func (h *EventsHandler) DeleteEventGroupAdmin(c *gin.Context) {
	identifier := c.Param("id")

	var event models.Event
	query := h.db.Preload("TargetGroups")

	// Essayer de traiter l'identifiant comme un ID numérique d'abord
	if eventID, err := strconv.Atoi(identifier); err == nil {
		// C'est un ID numérique valide
		query = query.Where("id = ?", eventID)
	} else {
		// Sinon, c'est un slug
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&event).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Événement non trouvé"})
		return
	}

	// Vérification des permissions
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	// Vérifier si le group admin peut gérer cet événement
	canManage := false

	// 1. Si c'est son propre événement
	if event.AuthorID == userID {
		canManage = true
	} else {
		// 2. Si l'événement cible un de ses groupes gérés
		for _, targetGroup := range event.TargetGroups {
			for _, managedID := range managedGroupIDs {
				if targetGroup.ID == managedID {
					canManage = true
					break
				}
			}
			if canManage {
				break
			}
		}
	}

	if !canManage {
		c.JSON(http.StatusForbidden, gin.H{"error": "Vous ne pouvez supprimer que les événements de vos groupes gérés"})
		return
	}

	// Soft delete (GORM)
	if err := h.db.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Événement supprimé avec succès"})
}
