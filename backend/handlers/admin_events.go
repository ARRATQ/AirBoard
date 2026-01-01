package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"airboard/config"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
)

// CreateEvent - Créer un événement (Admin/Editor)
func (h *EventsHandler) CreateEvent(c *gin.Context) {
	var req models.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
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

	// Envoyer une notification email si l'événement est publié
	if event.IsPublished {
		go func() {
			emailService := services.NewEmailService(h.db, config.LoadConfig())
			var targetGroupIDs []uint
			for _, g := range event.TargetGroups {
				targetGroupIDs = append(targetGroupIDs, g.ID)
			}
			if err := emailService.SendNotification("event", event.ID, targetGroupIDs); err != nil {
				log.Printf("[Email] Échec de l'envoi de la notification événement: %v", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, event)
}

// UpdateEvent - Modifier un événement (Admin/Editor - propriétaire)
func (h *EventsHandler) UpdateEvent(c *gin.Context) {
	identifier := c.Param("id")

	var event models.Event
	query := h.db.Model(&models.Event{})

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
	userRole := c.GetString("role")

	if userRole != "admin" && event.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Vous ne pouvez modifier que vos propres événements"})
		return
	}

	var req models.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
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

// DeleteEvent - Supprimer un événement (Admin/Editor - propriétaire)
func (h *EventsHandler) DeleteEvent(c *gin.Context) {
	identifier := c.Param("id")

	var event models.Event
	query := h.db.Model(&models.Event{})

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
	userRole := c.GetString("role")

	if userRole != "admin" && event.AuthorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Vous ne pouvez supprimer que vos propres événements"})
		return
	}

	// Soft delete (GORM)
	if err := h.db.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Événement supprimé avec succès"})
}

// ListEvents - Liste des événements pour l'admin (Admin uniquement)
func (h *EventsHandler) ListEvents(c *gin.Context) {
	var events []models.Event

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * pageSize

	// Construction de la requête avec toutes les relations
	query := h.db.Model(&models.Event{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups")

	// Filtres
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR description LIKE ? OR location LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if status := c.Query("status"); status != "" {
		if status == "published" {
			query = query.Where("is_published = ?", true)
		} else if status == "draft" {
			query = query.Where("is_published = ?", false)
		}
	}

	if eventType := c.Query("event_type"); eventType != "" {
		switch eventType {
		case "recurring":
			query = query.Where("is_recurring = ?", true)
		case "all_day":
			query = query.Where("is_all_day = ?", true)
		case "upcoming":
			query = query.Where("(end_date IS NULL AND start_date >= ?) OR (end_date IS NOT NULL AND end_date >= ?)",
				time.Now(), time.Now())
		case "past":
			query = query.Where("(end_date IS NOT NULL AND end_date < ?) OR (end_date IS NULL AND start_date < ?)",
				time.Now(), time.Now())
		}
	}

	// Tri
	sortBy := c.DefaultQuery("sort", "start_date")
	sortOrder := c.DefaultQuery("order", "desc")
	if sortBy == "start_date" || sortBy == "created_at" || sortBy == "priority" {
		query = query.Order(sortBy + " " + sortOrder)
	} else {
		query = query.Order("start_date desc")
	}

	// Compter le total
	var total int64
	query.Count(&total)

	// Récupérer les événements avec pagination
	query.Offset(offset).Limit(pageSize).Find(&events)

	// Calculer le nombre total de pages
	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.EventListResponse{
		Events:     events,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetAnalytics - Statistiques des événements (Admin uniquement)
func (h *EventsHandler) GetAnalytics(c *gin.Context) {
	var stats models.EventStatsResponse

	// Total événements
	h.db.Model(&models.Event{}).Count(&stats.TotalEvents)

	// Événements publiés
	h.db.Model(&models.Event{}).Where("is_published = ?", true).Count(&stats.PublishedEvents)

	// Événements brouillons
	stats.DraftEvents = stats.TotalEvents - stats.PublishedEvents

	// Événements à venir
	now := time.Now()
	h.db.Model(&models.Event{}).
		Where("(end_date IS NULL AND start_date >= ?) OR (end_date IS NOT NULL AND end_date >= ?)", now, now).
		Count(&stats.UpcomingEvents)

	// Événements passés
	h.db.Model(&models.Event{}).
		Where("(end_date IS NOT NULL AND end_date < ?) OR (end_date IS NULL AND start_date < ?)", now, now).
		Count(&stats.PastEvents)

	// Événements récurrents
	h.db.Model(&models.Event{}).Where("is_recurring = ?", true).Count(&stats.RecurringEvents)

	// Événements par catégorie
	stats.EventsByCategory = make(map[string]int64)
	var categoryStats []struct {
		CategoryName string
		Count        int64
	}
	h.db.Model(&models.Event{}).
		Select("event_categories.name as category_name, COUNT(events.id) as count").
		Joins("LEFT JOIN event_categories ON events.category_id = event_categories.id").
		Group("event_categories.name").
		Scan(&categoryStats)

	for _, stat := range categoryStats {
		if stat.CategoryName != "" {
			stats.EventsByCategory[stat.CategoryName] = stat.Count
		} else {
			stats.EventsByCategory["Sans catégorie"] = stat.Count
		}
	}

	// Événements par priorité
	stats.EventsByPriority = make(map[string]int64)
	var priorityStats []struct {
		Priority string
		Count    int64
	}
	h.db.Model(&models.Event{}).
		Select("priority, COUNT(*) as count").
		Group("priority").
		Scan(&priorityStats)

	for _, stat := range priorityStats {
		stats.EventsByPriority[stat.Priority] = stat.Count
	}

	c.JSON(http.StatusOK, stats)
}

// CreateCategory - Créer une catégorie d'événements (Admin uniquement)
func (h *EventsHandler) CreateCategory(c *gin.Context) {
	var req models.EventCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
	}

	category := models.EventCategory{
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Order:       req.Order,
		IsActive:    req.IsActive,
	}

	if err := h.db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de la catégorie"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory - Modifier une catégorie d'événements (Admin uniquement)
func (h *EventsHandler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.EventCategory
	if err := h.db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Catégorie non trouvée"})
		return
	}

	var req models.EventCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
	}

	category.Name = req.Name
	category.Description = req.Description
	category.Icon = req.Icon
	category.Color = req.Color
	category.Order = req.Order
	category.IsActive = req.IsActive

	if err := h.db.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory - Supprimer une catégorie d'événements (Admin uniquement)
func (h *EventsHandler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.EventCategory
	if err := h.db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Catégorie non trouvée"})
		return
	}

	// Vérifier si des événements utilisent cette catégorie
	var eventCount int64
	h.db.Model(&models.Event{}).Where("category_id = ?", id).Count(&eventCount)

	if eventCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Impossible de supprimer : des événements utilisent cette catégorie"})
		return
	}

	if err := h.db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Catégorie supprimée avec succès"})
}

// ============================================
// Gestion des jours fériés
// ============================================

// GetAvailableCountries - Liste des pays disponibles pour l'import des jours fériés
func (h *EventsHandler) GetAvailableCountries(c *gin.Context) {
	holidayService := services.NewHolidayService(h.db)

	countries, err := holidayService.GetAvailableCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"countries": countries})
}

// ImportHolidaysRequest représente la requête d'import des jours fériés
type ImportHolidaysRequest struct {
	CountryCode string `json:"country_code" binding:"required"`
	Year        int    `json:"year" binding:"required"`
	CategoryID  *uint  `json:"category_id"`
}

// ImportHolidays - Importe les jours fériés pour un pays et une année
func (h *EventsHandler) ImportHolidays(c *gin.Context) {
	var req ImportHolidaysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Données invalides: " + err.Error()})
		return
	}

	// Validation de l'année
	currentYear := time.Now().Year()
	if req.Year < currentYear-5 || req.Year > currentYear+5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "L'année doit être comprise entre " + strconv.Itoa(currentYear-5) + " et " + strconv.Itoa(currentYear+5)})
		return
	}

	// Récupérer l'utilisateur connecté comme auteur
	authorID := c.GetUint("user_id")

	holidayService := services.NewHolidayService(h.db)
	result, err := holidayService.ImportHolidays(req.CountryCode, req.Year, authorID, req.CategoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Import terminé avec succès",
		"imported": result.Imported,
		"skipped":  result.Skipped,
		"errors":   result.Errors,
	})
}

// DeleteHolidays - Supprime les jours fériés d'un pays pour une année
func (h *EventsHandler) DeleteHolidays(c *gin.Context) {
	countryCode := c.Query("country_code")
	yearStr := c.Query("year")

	if countryCode == "" || yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country_code et year sont requis"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Année invalide"})
		return
	}

	holidayService := services.NewHolidayService(h.db)
	deleted, err := holidayService.DeleteHolidaysByCountryAndYear(countryCode, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Suppression terminée",
		"deleted": deleted,
	})
}

// PreviewHolidays - Prévisualise les jours fériés sans les importer
func (h *EventsHandler) PreviewHolidays(c *gin.Context) {
	countryCode := c.Query("country_code")
	yearStr := c.Query("year")

	if countryCode == "" || yearStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country_code et year sont requis"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Année invalide"})
		return
	}

	holidayService := services.NewHolidayService(h.db)
	holidays, err := holidayService.FetchHolidays(countryCode, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"holidays": holidays,
		"count":    len(holidays),
	})
}
