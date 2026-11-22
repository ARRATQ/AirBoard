package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"airboard/apierror"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EventsHandler struct {
	db                  *gorm.DB
	gamificationService *services.GamificationService
}

func NewEventsHandler(db *gorm.DB, gs *services.GamificationService) *EventsHandler {
	return &EventsHandler{db: db, gamificationService: gs}
}

// GetEvents - Liste des événements (accessible à tous les utilisateurs connectés)
func (h *EventsHandler) GetEvents(c *gin.Context) {
	var events []models.Event

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	offset := (page - 1) * pageSize

	// Construction de la requête
	query := h.db.Model(&models.Event{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups")

	// Explicit soft delete filter (complex WHERE clauses may bypass GORM's automatic filter)
	query = query.Where("deleted_at IS NULL")

	// Filtres
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filtres publié / brouillon, récurrent et journée entière de la liste d'administration
	for _, column := range []string{"is_published", "is_recurring", "is_all_day"} {
		if value, ok := parseBoolFilter(c.Query(column)); ok {
			query = query.Where(column+" = ?", value)
		}
	}

	// Filtre par tags (supporte plusieurs tags séparés par des virgules)
	if tags := c.Query("tags"); tags != "" {
		tagIDs := strings.Split(tags, ",")
		if len(tagIDs) > 0 {
			query = query.Joins("JOIN event_tags ON event_tags.event_id = events.id").
				Where("event_tags.tag_id IN ?", tagIDs).
				Group("events.id").
				Having("COUNT(DISTINCT event_tags.tag_id) = ?", len(tagIDs))
		}
	}

	// Filtre par plage de dates
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("start_date >= ?", parsed)
		}
	}

	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			// Inclure les événements qui commencent avant end_date
			query = query.Where("start_date <= ?", parsed)
		}
	}

	// Filtre événements à venir / passés
	if upcoming := c.Query("upcoming"); upcoming == "true" {
		query = query.Where("(end_date IS NULL AND start_date >= ?) OR (end_date IS NOT NULL AND end_date >= ?)",
			time.Now(), time.Now())
	} else if past := c.Query("past"); past == "true" {
		query = query.Where("(end_date IS NOT NULL AND end_date < ?) OR (end_date IS NULL AND start_date < ?)",
			time.Now(), time.Now())
	}

	// Recherche sécurisée avec validation et assainissement
	if search := c.Query("search"); search != "" {
		// Valider l'input de recherche (longueur en caractères ; un seul caractère est permis)
		if listSearchTooLong(search) {
			apierror.Respond(c, http.StatusBadRequest, "La recherche ne peut pas dépasser 100 caractères")
			return
		}

		// Requête paramétrée : seuls les jokers de LIKE (\, % et _) sont échappés
		sanitizedSearch := strings.ReplaceAll(search, "\\", "\\\\")
		sanitizedSearch = strings.ReplaceAll(sanitizedSearch, "%", "\\%")
		sanitizedSearch = strings.ReplaceAll(sanitizedSearch, "_", "\\_")

		query = query.Where("title ILIKE ? OR description ILIKE ? OR location ILIKE ?",
			"%"+sanitizedSearch+"%", "%"+sanitizedSearch+"%", "%"+sanitizedSearch+"%")
	}

	// Filtre published only et visibilité par groupes selon les droits
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	now := time.Now()

	switch {
	case middleware.HasPermissionEverywhere(c, models.PermEventsManage):
		// Admin, ou gestion des événements sur tout le site : tout (publié + brouillons)
	case len(managedGroupIDs) > 0:
		// Administrateur de groupe. Interface /group-admin/events : les événements qu'il
		// gère. Lecture publique : événements publiés globaux ou des groupes dont il est
		// membre ou administrateur.
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/group-admin/events") {
			query = query.Scopes(eventTargets.manageableBy(userID, managedGroupIDs))
		} else {
			groupIDs := memberOrManagedGroupIDs(h.db, userID, managedGroupIDs)
			query = query.Scopes(eventTargets.publishedVisibleTo(groupIDs, now))
		}
	default:
		// User régulier voit : événements publiés globaux + ceux ciblant ses groupes
		query = query.Scopes(eventTargets.publishedVisibleTo(userGroupIDs(h.db, userID), now))
	}

	// Tri
	query = query.Order(utils.SafeOrderClause(
		c.Query("sort"), c.Query("order"),
		[]string{"start_date", "created_at", "priority"},
		"start_date", "asc",
	))

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

// GetCalendarView - Vue calendrier avec événements expandus (CRITIQUE)
func (h *EventsHandler) GetCalendarView(c *gin.Context) {
	// Paramètres de plage de dates (requis)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		apierror.Respond(c, http.StatusBadRequest, "start_date et end_date sont requis (format YYYY-MM-DD)")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Format start_date invalide (YYYY-MM-DD requis)")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Format end_date invalide (YYYY-MM-DD requis)")
		return
	}

	// Récupérer tous les événements dans la plage (+ filtres de visibilité)
	var allEvents []models.Event

	query := h.db.Model(&models.Event{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		Where("is_published = ?", true).
		Where("published_at IS NULL OR published_at <= ?", time.Now())

	// Filtrer par plage de dates
	// Inclure les événements qui ont une partie dans la période demandée
	query = query.Where(
		"(start_date <= ? AND (end_date IS NULL OR end_date >= ?)) OR (end_date IS NOT NULL AND end_date >= ? AND start_date <= ?)",
		endDate, startDate, endDate, startDate)

	// Appliquer filtres de visibilité selon rôle : l'admin voit tout ; un administrateur
	// de groupe, les événements globaux et ceux des groupes dont il est membre ou
	// administrateur ; un utilisateur, les événements globaux et ceux de ses groupes.
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	if !middleware.HasPermissionEverywhere(c, models.PermEventsManage) {
		query = query.Scopes(eventTargets.visibleTo(memberOrManagedGroupIDs(h.db, userID, managedGroupIDs)))
	}

	// Filtres optionnels
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Find(&allEvents)

	// Séparer événements récurrents et non-récurrents
	var normalEvents []models.Event
	var recurringEvents []models.Event

	for _, event := range allEvents {
		if event.IsRecurring {
			recurringEvents = append(recurringEvents, event)
		} else {
			normalEvents = append(normalEvents, event)
		}
	}

	// Expander les événements récurrents
	recurringInstances := services.ExpandRecurringEvents(recurringEvents, startDate, endDate)

	// Calculer le résumé
	publicCount := 0
	privateCount := 0

	for _, event := range allEvents {
		var targetGroupCount int64
		h.db.Model(&event).Association("TargetGroups").Count()
		if targetGroupCount == 0 {
			publicCount++
		} else {
			privateCount++
		}
	}

	summary := models.EventCalendarSummary{
		TotalEvents:    len(allEvents),
		PublicEvents:   publicCount,
		PrivateEvents:  privateCount,
		RecurringCount: len(recurringEvents),
	}

	c.JSON(http.StatusOK, models.EventCalendarResponse{
		Events:             normalEvents,
		RecurringInstances: recurringInstances,
		Summary:            summary,
	})
}

// GetEventBySlug - Détail d'un événement par slug ou ID
func (h *EventsHandler) GetEventBySlug(c *gin.Context) {
	identifier := c.Param("slug")

	var event models.Event
	query := h.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups")

	// Essayer de traiter l'identifiant comme un ID numérique d'abord
	if eventID, err := strconv.Atoi(identifier); err == nil {
		// C'est un ID numérique valide
		query = query.Where("id = ?", eventID)
	} else {
		// Sinon, c'est un slug
		query = query.Where("slug = ?", identifier)
	}

	if err := query.First(&event).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Événement non trouvé")
		return
	}

	// Vérifier la visibilité selon les droits
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	if !middleware.HasPermissionEverywhere(c, models.PermEventsManage) {
		// Vérifier si l'événement est publié
		if !event.IsPublished {
			// Seul l'auteur peut voir un brouillon
			if event.AuthorID != userID {
				apierror.Respond(c, http.StatusForbidden, "Accès refusé")
				return
			}
		}

		// Vérifier la visibilité par groupe : un événement ciblé n'est visible que des
		// membres ou administrateurs d'un groupe cible (et de son auteur), comme dans
		// les listes. Le compte des groupes cibles n'était jusqu'ici jamais lu, ce qui
		// laissait tout événement ciblé accessible par son identifiant.
		if len(event.TargetGroups) > 0 && event.AuthorID != userID &&
			!targetsAnyGroup(event.TargetGroups, memberOrManagedGroupIDs(h.db, userID, managedGroupIDs)) {
			apierror.Respond(c, http.StatusForbidden, "Accès refusé")
			return
		}
	}

	c.JSON(http.StatusOK, event)
}

// GetCategories - Liste des catégories d'événements
func (h *EventsHandler) GetCategories(c *gin.Context) {
	var categories []models.EventCategory

	h.db.Where("is_active = ?", true).Order("\"order\" asc").Find(&categories)

	c.JSON(http.StatusOK, categories)
}
