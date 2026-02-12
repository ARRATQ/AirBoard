package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"airboard/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchHandler struct {
	db *gorm.DB
}

func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

// Structures de réponse pour la recherche
type SearchResultApp struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Icon         string `json:"icon"`
	URL          string `json:"url"`
	Color        string `json:"color"`
	AppGroupName string `json:"app_group_name"`
}

type SearchResultNews struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Slug         string    `json:"slug"`
	Summary      string    `json:"summary"`
	CategoryName string    `json:"category_name"`
	AuthorName   string    `json:"author_name"`
	CreatedAt    time.Time `json:"created_at"`
}

type SearchResultEvent struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Slug      string     `json:"slug"`
	Location  string     `json:"location"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type SearchResultPoll struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	IsActive    bool       `json:"is_active"`
	EndDate     *time.Time `json:"end_date"`
}

type SearchResultAnnouncement struct {
	ID        uint       `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type SearchResponse struct {
	Applications  []SearchResultApp          `json:"applications"`
	News          []SearchResultNews         `json:"news"`
	Events        []SearchResultEvent        `json:"events"`
	Polls         []SearchResultPoll         `json:"polls"`
	Announcements []SearchResultAnnouncement `json:"announcements"`
	TotalCount    int                        `json:"total_count"`
}

// GlobalSearch - Recherche globale à travers toutes les entités
func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	q := c.Query("q")
	typeFilter := c.Query("type") // app, news, event, poll, announcement

	// Validation de la requête
	if len(q) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La recherche doit contenir au moins 2 caractères"})
		return
	}
	if len(q) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "La recherche ne peut pas dépasser 100 caractères"})
		return
	}

	// Vérifier les caractères interdits
	forbiddenChars := []string{"<", ">", "&", "\"", "'", "(", ")", "=", "+", ";", "--", "/*", "*/", "union", "select", "insert", "update", "delete", "drop", "create"}
	lowerSearch := strings.ToLower(q)
	for _, char := range forbiddenChars {
		if strings.Contains(lowerSearch, char) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Caractère ou terme interdit dans la recherche: %s", char)})
			return
		}
	}

	// Assainir l'input pour les requêtes LIKE
	sanitized := strings.ReplaceAll(q, "%", "\\%")
	sanitized = strings.ReplaceAll(sanitized, "_", "\\_")
	likePattern := "%" + sanitized + "%"

	// Récupérer le contexte utilisateur
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	// Récupérer les groupes d'appartenance de l'utilisateur
	var userGroupIDs []uint
	h.db.Table("user_groups").Where("user_id = ?", userID).Pluck("group_id", &userGroupIDs)

	// Combiner groupes d'appartenance + groupes administrés
	allGroupIDs := make(map[uint]bool)
	for _, id := range userGroupIDs {
		allGroupIDs[id] = true
	}
	for _, id := range managedGroupIDs {
		allGroupIDs[id] = true
	}
	var combinedGroupIDs []uint
	for id := range allGroupIDs {
		combinedGroupIDs = append(combinedGroupIDs, id)
	}

	response := SearchResponse{
		Applications:  []SearchResultApp{},
		News:          []SearchResultNews{},
		Events:        []SearchResultEvent{},
		Polls:         []SearchResultPoll{},
		Announcements: []SearchResultAnnouncement{},
	}

	// --- Applications ---
	if typeFilter == "" || typeFilter == "app" {
		response.Applications = h.searchApplications(userID, userRole, combinedGroupIDs, likePattern)
	}

	// --- News ---
	if typeFilter == "" || typeFilter == "news" {
		response.News = h.searchNews(userID, userRole, combinedGroupIDs, likePattern)
	}

	// --- Events ---
	if typeFilter == "" || typeFilter == "event" {
		response.Events = h.searchEvents(userID, userRole, combinedGroupIDs, likePattern)
	}

	// --- Polls ---
	if typeFilter == "" || typeFilter == "poll" {
		response.Polls = h.searchPolls(userID, userRole, combinedGroupIDs, likePattern)
	}

	// --- Announcements ---
	if typeFilter == "" || typeFilter == "announcement" {
		response.Announcements = h.searchAnnouncements(likePattern)
	}

	response.TotalCount = len(response.Applications) + len(response.News) + len(response.Events) + len(response.Polls) + len(response.Announcements)

	c.JSON(http.StatusOK, response)
}

// searchApplications - Recherche d'applications avec visibilité par groupe
func (h *SearchHandler) searchApplications(userID uint, userRole string, combinedGroupIDs []uint, likePattern string) []SearchResultApp {
	var results []SearchResultApp

	if userRole == "admin" {
		// Admin voit toutes les applications actives
		h.db.Table("applications").
			Select("applications.id, applications.name, applications.description, applications.icon, applications.url, applications.color, app_groups.name as app_group_name").
			Joins("LEFT JOIN app_groups ON app_groups.id = applications.app_group_id").
			Where("applications.is_active = ? AND applications.deleted_at IS NULL", true).
			Where("applications.name ILIKE ? OR applications.description ILIKE ?", likePattern, likePattern).
			Limit(10).
			Find(&results)
	} else if len(combinedGroupIDs) > 0 {
		// Utilisateur avec groupes : voir les apps des AppGroups accessibles via ses groupes
		h.db.Table("applications").
			Select("applications.id, applications.name, applications.description, applications.icon, applications.url, applications.color, app_groups.name as app_group_name").
			Joins("JOIN app_groups ON app_groups.id = applications.app_group_id").
			Joins("JOIN group_app_groups ON app_groups.id = group_app_groups.app_group_id").
			Where("group_app_groups.group_id IN ?", combinedGroupIDs).
			Where("applications.is_active = ? AND app_groups.is_active = ? AND applications.deleted_at IS NULL", true, true).
			Where("applications.name ILIKE ? OR applications.description ILIKE ?", likePattern, likePattern).
			Group("applications.id, app_groups.name").
			Limit(10).
			Find(&results)
	}
	// Si pas de groupes et pas admin : aucune application visible

	return results
}

// searchNews - Recherche de news avec visibilité par groupe
func (h *SearchHandler) searchNews(userID uint, userRole string, combinedGroupIDs []uint, likePattern string) []SearchResultNews {
	type newsRow struct {
		ID           uint      `json:"id"`
		Title        string    `json:"title"`
		Slug         string    `json:"slug"`
		Summary      string    `json:"summary"`
		CategoryName *string   `json:"category_name"`
		FirstName    string    `json:"first_name"`
		LastName     string    `json:"last_name"`
		CreatedAt    time.Time `json:"created_at"`
	}
	var rows []newsRow

	query := h.db.Table("news").
		Select("news.id, news.title, news.slug, news.summary, news_categories.name as category_name, users.first_name, users.last_name, news.created_at").
		Joins("LEFT JOIN news_categories ON news_categories.id = news.category_id").
		Joins("LEFT JOIN users ON users.id = news.author_id").
		Where("news.deleted_at IS NULL").
		Where("news.title ILIKE ? OR news.summary ILIKE ?", likePattern, likePattern)

	now := time.Now()

	if userRole == "admin" {
		// Admin voit tout (publié + brouillons)
	} else {
		// Tous les autres rôles : uniquement les news publiées
		query = query.Where("news.is_published = ? AND (news.published_at IS NULL OR news.published_at <= ?)", true, now)

		// Filtre de visibilité par groupes
		if len(combinedGroupIDs) > 0 {
			query = query.Where(`
				(SELECT COUNT(*) FROM news_target_groups WHERE news_target_groups.news_id = news.id) = 0
				OR EXISTS (
					SELECT 1 FROM news_target_groups
					WHERE news_target_groups.news_id = news.id
					AND news_target_groups.group_id IN (?)
				)
			`, combinedGroupIDs)
		} else {
			// Pas de groupes : uniquement les news globales (sans target groups)
			query = query.Where("(SELECT COUNT(*) FROM news_target_groups WHERE news_target_groups.news_id = news.id) = 0")
		}
	}

	query.Order("news.created_at DESC").Limit(10).Find(&rows)

	results := make([]SearchResultNews, len(rows))
	for i, r := range rows {
		catName := ""
		if r.CategoryName != nil {
			catName = *r.CategoryName
		}
		results[i] = SearchResultNews{
			ID:           r.ID,
			Title:        r.Title,
			Slug:         r.Slug,
			Summary:      r.Summary,
			CategoryName: catName,
			AuthorName:   strings.TrimSpace(r.FirstName + " " + r.LastName),
			CreatedAt:    r.CreatedAt,
		}
	}
	return results
}

// searchEvents - Recherche d'événements avec visibilité par groupe
func (h *SearchHandler) searchEvents(userID uint, userRole string, combinedGroupIDs []uint, likePattern string) []SearchResultEvent {
	var results []SearchResultEvent

	query := h.db.Table("events").
		Select("events.id, events.title, events.slug, events.location, events.start_date, events.end_date").
		Where("events.deleted_at IS NULL").
		Where("events.title ILIKE ? OR events.description ILIKE ? OR events.location ILIKE ?", likePattern, likePattern, likePattern)

	now := time.Now()

	if userRole == "admin" {
		// Admin voit tout
	} else {
		// Uniquement les événements publiés
		query = query.Where("events.is_published = ? AND (events.published_at IS NULL OR events.published_at <= ?)", true, now)

		// Filtre de visibilité par groupes
		if len(combinedGroupIDs) > 0 {
			query = query.Where(`
				(SELECT COUNT(*) FROM event_target_groups WHERE event_target_groups.event_id = events.id) = 0
				OR EXISTS (
					SELECT 1 FROM event_target_groups
					WHERE event_target_groups.event_id = events.id
					AND event_target_groups.group_id IN (?)
				)
			`, combinedGroupIDs)
		} else {
			query = query.Where("(SELECT COUNT(*) FROM event_target_groups WHERE event_target_groups.event_id = events.id) = 0")
		}
	}

	query.Order("events.start_date DESC").Limit(10).Find(&results)

	return results
}

// searchPolls - Recherche de sondages avec visibilité par groupe
func (h *SearchHandler) searchPolls(userID uint, userRole string, combinedGroupIDs []uint, likePattern string) []SearchResultPoll {
	var results []SearchResultPoll

	query := h.db.Table("polls").
		Select("polls.id, polls.title, polls.description, polls.is_active, polls.end_date").
		Where("polls.deleted_at IS NULL").
		Where("polls.title ILIKE ? OR polls.description ILIKE ?", likePattern, likePattern)

	if userRole == "admin" {
		// Admin voit tout
	} else {
		// Filtre de visibilité par groupes
		if len(combinedGroupIDs) > 0 {
			query = query.Where(`
				(SELECT COUNT(*) FROM poll_target_groups WHERE poll_target_groups.poll_id = polls.id) = 0
				OR EXISTS (
					SELECT 1 FROM poll_target_groups
					WHERE poll_target_groups.poll_id = polls.id
					AND poll_target_groups.group_id IN (?)
				)
			`, combinedGroupIDs)
		} else {
			query = query.Where("(SELECT COUNT(*) FROM poll_target_groups WHERE poll_target_groups.poll_id = polls.id) = 0")
		}
	}

	query.Order("polls.created_at DESC").Limit(10).Find(&results)

	return results
}

// searchAnnouncements - Recherche d'annonces actives (pas de filtre par groupe)
func (h *SearchHandler) searchAnnouncements(likePattern string) []SearchResultAnnouncement {
	var results []SearchResultAnnouncement
	now := time.Now()

	h.db.Table("announcements").
		Select("announcements.id, announcements.title, announcements.content, announcements.type, announcements.start_date, announcements.end_date").
		Where("announcements.is_active = ? AND announcements.deleted_at IS NULL", true).
		Where("(announcements.start_date IS NULL OR announcements.start_date <= ?)", now).
		Where("(announcements.end_date IS NULL OR announcements.end_date >= ?)", now).
		Where("announcements.title ILIKE ? OR announcements.content ILIKE ?", likePattern, likePattern).
		Order("announcements.priority DESC, announcements.created_at DESC").
		Limit(10).
		Find(&results)

	return results
}
