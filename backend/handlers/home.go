package handlers

import (
	"log"
	"math"
	"net/http"
	"sync"
	"time"

	"airboard/apierror"
	"airboard/models"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HomeHandler struct {
	db *gorm.DB
}

func NewHomeHandler(db *gorm.DB) *HomeHandler {
	return &HomeHandler{db: db}
}

// Response structures
type HomeStats struct {
	// Admin stats
	TotalUsers     int64 `json:"total_users,omitempty"`
	TotalGroups    int64 `json:"total_groups,omitempty"`
	TotalAppGroups int64 `json:"total_app_groups,omitempty"`
	TotalApps      int64 `json:"total_apps,omitempty"`
	TotalNews      int64 `json:"total_news,omitempty"`
	TotalEvents    int64 `json:"total_events,omitempty"`
	TotalPolls     int64 `json:"total_polls,omitempty"`

	// Group admin stats
	ManagedGroupsCount    int64 `json:"managed_groups_count,omitempty"`
	TotalMembersCount     int64 `json:"total_members_count,omitempty"`
	ManagedAppGroupsCount int64 `json:"managed_app_groups_count,omitempty"`
	ManagedAppsCount      int64 `json:"managed_apps_count,omitempty"`
	ManagedNewsCount      int64 `json:"managed_news_count,omitempty"`
	ManagedPollsCount     int64 `json:"managed_polls_count,omitempty"`

	// Regular user stats
	TotalAccessibleApps   int64 `json:"total_accessible_apps,omitempty"`
	TotalAccessibleEvents int64 `json:"total_accessible_events,omitempty"`
}

type GamificationSummary struct {
	Level           int                  `json:"level"`
	XP              int64                `json:"xp"`
	NextLevelXP     int64                `json:"next_level_xp"`
	ProgressPercent int                  `json:"progress_percent"`
	StreakDays      int                  `json:"streak_days"`
	BadgesCount     int                  `json:"badges_count"`
	Rank            int                  `json:"rank"`
	RecentBadges    []models.Achievement `json:"recent_badges"`
}

type NewsGroup struct {
	Type *models.NewsType `json:"type"`
	News []models.News    `json:"news"`
}

type PendingItems struct {
	UnreadNews      int64 `json:"unread_news"`
	UnansweredPolls int64 `json:"unanswered_polls"`
	NewAppsCount    int64 `json:"new_apps_count"`
}

type HomeResponse struct {
	FavoriteApps     []models.Application  `json:"favorite_apps"`
	NewApps          []models.Application  `json:"new_apps"`
	TodayEvents      []models.Event        `json:"today_events"`
	UpcomingEvents   []models.Event        `json:"upcoming_events"`
	RecentNews       []models.News         `json:"recent_news"`
	RecentNewsByType []NewsGroup           `json:"recent_news_by_type"`
	Polls            []models.Poll         `json:"polls"`
	Announcements    []models.Announcement `json:"announcements"`
	Stats            *HomeStats            `json:"stats"`
	Gamification     *GamificationSummary  `json:"gamification,omitempty"`
	UserRole         string                `json:"user_role"`
	ManagedGroupIDs  []uint                `json:"managed_group_ids,omitempty"`
	AppSettings      *models.AppSettings   `json:"app_settings,omitempty"`
	HeroMessages     []models.HeroMessage  `json:"hero_messages,omitempty"`
	PendingItems     *PendingItems         `json:"pending_items,omitempty"`
}

// Main handler
func (h *HomeHandler) GetHomeData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.Respond(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	role, _ := c.Get("role")
	managedGroupIDs, _ := c.Get("managed_group_ids")

	response := HomeResponse{
		UserRole: role.(string),
	}

	if managedGroupIDs != nil {
		response.ManagedGroupIDs = managedGroupIDs.([]uint)
	}

	// Load settings upfront (needed for newsPerTab before goroutines start)
	settings, _ := h.getAppSettingsWithCache()
	newsPerTab := 5
	if settings != nil && settings.NewsPerTab > 0 {
		newsPerTab = settings.NewsPerTab
	}

	// Groupes du visiteur, lus une fois pour tous les chargements ci-dessous
	groups := h.loadViewerGroups(userID.(uint), role.(string))

	// Use WaitGroup for parallel queries
	var wg sync.WaitGroup
	var mu sync.Mutex // Protect response writes

	// 1. Load Favorite Apps (user-specific, no cache)
	wg.Add(1)
	utils.SafeGo("home-favorites", func() {
		defer wg.Done()
		apps, err := h.getFavoriteApps(userID.(uint))
		if err != nil {
			log.Printf("[HOME] Failed to load favorites: %v", err)
			apps = []models.Application{}
		}
		mu.Lock()
		response.FavoriteApps = apps
		mu.Unlock()
	})

	// 2. Load New Apps (filtered by user groups)
	wg.Add(1)
	utils.SafeGo("home-new-apps", func() {
		defer wg.Done()
		apps, err := h.getNewApps(userID.(uint), role.(string))
		if err != nil {
			log.Printf("[HOME] Failed to load new apps: %v", err)
			apps = []models.Application{}
		}
		mu.Lock()
		response.NewApps = apps
		mu.Unlock()
	})

	// 3. Load Today's Events
	wg.Add(1)
	utils.SafeGo("home-today-events", func() {
		defer wg.Done()
		events, err := h.getTodayEvents(role.(string), groups)
		if err != nil {
			log.Printf("[HOME] Failed to load today's events: %v", err)
			events = []models.Event{}
		}
		mu.Lock()
		response.TodayEvents = events
		mu.Unlock()
	})

	// 4. Load Upcoming Events
	wg.Add(1)
	utils.SafeGo("home-upcoming-events", func() {
		defer wg.Done()
		events, err := h.getUpcomingEvents(role.(string), groups)
		if err != nil {
			log.Printf("[HOME] Failed to load upcoming events: %v", err)
			events = []models.Event{}
		}
		mu.Lock()
		response.UpcomingEvents = events
		mu.Unlock()
	})

	// 5. Load Recent News
	wg.Add(1)
	utils.SafeGo("home-recent-news", func() {
		defer wg.Done()
		news, err := h.getRecentNews(role.(string), groups)
		if err != nil {
			log.Printf("[HOME] Failed to load recent news: %v", err)
			news = []models.News{}
		}
		mu.Lock()
		response.RecentNews = news
		mu.Unlock()
	})

	// 6. Load Recent News By Type (grouped)
	wg.Add(1)
	utils.SafeGo("home-news-by-type", func() {
		defer wg.Done()
		newsByType, err := h.getRecentNewsByType(role.(string), groups, newsPerTab)
		if err != nil {
			log.Printf("[HOME] Failed to load recent news by type: %v", err)
			newsByType = []NewsGroup{}
		}
		mu.Lock()
		response.RecentNewsByType = newsByType
		mu.Unlock()
	})

	// 7. Load Recent Polls
	wg.Add(1)
	utils.SafeGo("home-polls", func() {
		defer wg.Done()
		polls, err := h.getRecentPolls(role.(string), groups)
		if err != nil {
			log.Printf("[HOME] Failed to load recent polls: %v", err)
			polls = []models.Poll{}
		}
		mu.Lock()
		response.Polls = polls
		mu.Unlock()
	})

	// 8. Load Announcements (cached)
	wg.Add(1)
	utils.SafeGo("home-announcements", func() {
		defer wg.Done()
		announcements, err := h.getAnnouncementsWithCache()
		if err != nil {
			log.Printf("[HOME] Failed to load announcements: %v", err)
			announcements = []models.Announcement{}
		}
		mu.Lock()
		response.Announcements = announcements
		mu.Unlock()
	})

	// 9. Load Stats (for all users)
	wg.Add(1)
	utils.SafeGo("home-stats", func() {
		defer wg.Done()
		userIDVal := userID.(uint)
		roleVal := role.(string)
		stats := h.getStats(roleVal, response.ManagedGroupIDs, groups)
		log.Printf("[Home] GetHomeData: userID=%d, role=%s, stats=%+v", userIDVal, roleVal, stats)
		mu.Lock()
		response.Stats = stats
		mu.Unlock()
	})

	// 10. Load App Settings (cached)
	wg.Add(1)
	utils.SafeGo("home-app-settings", func() {
		defer wg.Done()
		settings, err := h.getAppSettingsWithCache()
		if err != nil {
			log.Printf("[HOME] Failed to load app settings: %v", err)
			settings = nil
		}
		mu.Lock()
		response.AppSettings = settings
		mu.Unlock()
	})

	// 11. Load Hero Messages (cached)
	wg.Add(1)
	utils.SafeGo("home-hero-messages", func() {
		defer wg.Done()
		messages, err := h.getHeroMessagesWithCache()
		if err != nil {
			log.Printf("[HOME] Failed to load hero messages: %v", err)
			messages = []models.HeroMessage{}
		}
		mu.Lock()
		response.HeroMessages = messages
		mu.Unlock()
	})

	// 12. Load Gamification Summary
	wg.Add(1)
	utils.SafeGo("home-gamification", func() {
		defer wg.Done()
		summary := h.getGamificationSummary(userID.(uint))
		mu.Lock()
		response.Gamification = summary
		mu.Unlock()
	})

	// 13. Load Pending Items (unread news, unanswered polls, new apps this week)
	wg.Add(1)
	utils.SafeGo("home-pending", func() {
		defer wg.Done()
		pending := h.getPendingItems(userID.(uint), role.(string), groups)
		mu.Lock()
		response.PendingItems = pending
		mu.Unlock()
	})

	wg.Wait()

	// Set cache headers
	c.Header("Cache-Control", "private, max-age=60")
	c.Header("Vary", "Authorization")

	c.JSON(http.StatusOK, response)
}

// Helper: Get favorite apps
func (h *HomeHandler) getFavoriteApps(userID uint) ([]models.Application, error) {
	var apps []models.Application
	err := h.db.Joins("JOIN user_favorites ON applications.id = user_favorites.application_id").
		Preload("AppGroup").
		Where("user_favorites.user_id = ? AND applications.is_active = ?", userID, true).
		Order("applications.name ASC").
		Find(&apps).Error
	return apps, err
}

// Helper: Get new apps (last 5 apps user has access to)
func (h *HomeHandler) getNewApps(userID uint, role string) ([]models.Application, error) {
	var apps []models.Application

	if role == "admin" {
		// Admins see all apps
		err := h.db.Where("is_active = ?", true).
			Preload("AppGroup").
			Order("created_at DESC").
			Limit(5).
			Find(&apps).Error
		return apps, err
	}

	// Users/Group Admins see apps in their groups + managed groups + public apps
	err := h.db.Distinct("applications.*").
		Joins("JOIN app_groups ON applications.app_group_id = app_groups.id").
		Joins("LEFT JOIN group_app_groups ON app_groups.id = group_app_groups.app_group_id").
		Joins("LEFT JOIN user_groups ON group_app_groups.group_id = user_groups.group_id").
		Joins("LEFT JOIN group_admins ON group_app_groups.group_id = group_admins.group_id AND group_admins.user_id = ?", userID).
		Where("applications.is_active = ? AND (app_groups.is_private = ? OR user_groups.user_id = ? OR group_admins.user_id = ?)",
			true, false, userID, userID).
		Preload("AppGroup").
		Order("applications.created_at DESC").
		Limit(5).
		Find(&apps).Error
	return apps, err
}

// viewerGroups : groupes du visiteur, lus une seule fois par requête puis partagés
// par les chargements parallèles de l'accueil (vides pour un admin, qui voit tout).
type viewerGroups struct {
	member []uint // groupes dont il est membre
	all    []uint // groupes dont il est membre ou administrateur
}

func (h *HomeHandler) loadViewerGroups(userID uint, role string) viewerGroups {
	if role == "admin" {
		return viewerGroups{}
	}
	member := userGroupIDs(h.db, userID)
	return viewerGroups{member: member, all: mergeGroupIDs(member, adminGroupIDs(h.db, userID))}
}

// Helper: Get today's events (events starting today only)
func (h *HomeHandler) getTodayEvents(role string, groups viewerGroups) ([]models.Event, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var events []models.Event

	query := h.db.Preload("Author").
		Preload("Category").
		Order("start_date ASC").
		Where("is_published = ? AND ((start_date >= ? AND start_date < ?) OR (start_date < ? AND end_date >= ?))",
			true, startOfDay, endOfDay, startOfDay, startOfDay)

	// Admins see all published events ; les autres : événements globaux ou de leurs groupes
	if role != "admin" {
		query = query.Scopes(eventTargets.visibleTo(groups.member))
	}

	err := query.Find(&events).Error
	return events, err
}

// Helper: Get upcoming events (next 5 events starting after today)
func (h *HomeHandler) getUpcomingEvents(role string, groups viewerGroups) ([]models.Event, error) {
	tomorrow := time.Now().Add(24 * time.Hour)
	startOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	var events []models.Event

	query := h.db.Preload("Author").
		Preload("Category").
		Order("start_date ASC").
		Limit(5).
		Where("is_published = ? AND start_date > ?", true, startOfTomorrow)

	// Admins see all published events ; les autres : événements globaux ou de leurs groupes
	if role != "admin" {
		query = query.Scopes(eventTargets.visibleTo(groups.member))
	}

	err := query.Find(&events).Error
	return events, err
}

// Helper: Get recent news (last 5 published articles filtered by user groups)
func (h *HomeHandler) getRecentNews(role string, groups viewerGroups) ([]models.News, error) {
	var news []models.News

	query := h.db.Where("is_published = ?", true).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		Order("is_pinned DESC, published_at DESC").
		Limit(5)

	// Les admins voient tous les articles publiés ; les autres, les articles globaux
	// et ceux ciblant un groupe dont ils sont membres ou administrateurs
	if role != "admin" {
		query = query.Scopes(newsTargets.visibleTo(groups.all))
	}

	if err := query.Find(&news).Error; err != nil {
		return news, err
	}

	// Compteurs de commentaires et réactions (2 requêtes pour tous les articles)
	applyNewsEngagementCounts(h.db, news)

	return news, nil
}

// Helper: Get recent news grouped by type (3 articles per type)
func (h *HomeHandler) getRecentNewsByType(role string, groups viewerGroups, newsPerTab int) ([]NewsGroup, error) {
	// First, fetch all active news types ordered by order
	var types []models.NewsType
	if err := h.db.Where("is_active = ?", true).Order("\"order\" ASC, name ASC").Find(&types).Error; err != nil {
		return []NewsGroup{}, err
	}

	if len(types) == 0 {
		return nil, nil
	}
	slugs := make([]string, 0, len(types))
	for _, t := range types {
		slugs = append(slugs, t.Slug)
	}

	// Les newsPerTab articles les plus récents de chaque type, en une seule requête
	// (au lieu d'une requête + 4 préchargements par type).
	ranked := h.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.News{}).
		Select("id, type, ROW_NUMBER() OVER (PARTITION BY type ORDER BY is_pinned DESC, published_at DESC) AS rn").
		Where("is_published = ? AND type IN ? AND deleted_at IS NULL", true, slugs)

	// Visibilité pour les non-admins : articles sans groupe cible, ou ciblant un de ses groupes
	if role != "admin" {
		ranked = ranked.Scopes(newsTargets.visibleTo(groups.all))
	}

	var refs []rankedNewsRef
	if err := h.db.Session(&gorm.Session{NewDB: true}).
		Table("(?) AS ranked", ranked).
		Select("id, type").
		Where("rn <= ?", newsPerTab).
		Order("type, rn").
		Scan(&refs).Error; err != nil {
		log.Printf("[HOME] Failed to rank news by type: %v", err)
		return []NewsGroup{}, err
	}
	if len(refs) == 0 {
		return nil, nil
	}

	ids := make([]uint, 0, len(refs))
	for _, ref := range refs {
		ids = append(ids, ref.ID)
	}

	var news []models.News
	if err := h.db.Session(&gorm.Session{NewDB: true}).
		Where("id IN ?", ids).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		Find(&news).Error; err != nil {
		log.Printf("[HOME] Failed to load news by type: %v", err)
		return []NewsGroup{}, err
	}

	applyNewsEngagementCounts(h.db, news)

	byID := make(map[uint]models.News, len(news))
	for _, n := range news {
		byID[n.ID] = n
	}
	return groupNewsByType(types, refs, byID), nil
}

// Helper: Get recent polls (last 10 polls - active and closed)
func (h *HomeHandler) getRecentPolls(role string, groups viewerGroups) ([]models.Poll, error) {
	var polls []models.Poll

	query := h.db.Preload("Author").
		Preload("Options").
		Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(10)

	// Les admins voient tous les sondages actifs ; les autres, les sondages globaux
	// et ceux ciblant un groupe dont ils sont membres ou administrateurs
	if role != "admin" {
		query = query.Scopes(pollTargets.visibleTo(groups.all))
	}

	if err := query.Find(&polls).Error; err != nil {
		return polls, err
	}

	// Nombre total de votes de chaque sondage (1 requête)
	applyPollVoteCounts(h.db, polls)

	return polls, nil
}

// Helper: Get announcements with cache
func (h *HomeHandler) getAnnouncementsWithCache() ([]models.Announcement, error) {
	homeCache.mu.RLock()
	cached := homeCache.announcements
	homeCache.mu.RUnlock()

	if cached.Data != nil && time.Now().Before(cached.ExpiresAt) {
		return cached.Data.([]models.Announcement), nil
	}

	// Cache miss - query database
	var announcements []models.Announcement
	now := time.Now()
	err := h.db.Where("is_active = ? AND (start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
		true, now, now).
		Order("priority DESC, created_at DESC").
		Find(&announcements).Error

	if err != nil {
		return nil, err
	}

	// Update cache
	homeCache.mu.Lock()
	homeCache.announcements = &CachedData{
		Data:      announcements,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	homeCache.mu.Unlock()

	return announcements, nil
}

// Helper: Get app settings with cache
func (h *HomeHandler) getAppSettingsWithCache() (*models.AppSettings, error) {
	homeCache.mu.RLock()
	cached := homeCache.appSettings
	homeCache.mu.RUnlock()

	if cached.Data != nil && time.Now().Before(cached.ExpiresAt) {
		settings := cached.Data.(models.AppSettings)
		return &settings, nil
	}

	// Cache miss - query database
	var settings models.AppSettings
	err := h.db.First(&settings).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default settings if none exist
			settings = models.AppSettings{
				AppName:         "Airboard",
				AppIcon:         "lucide:layout-dashboard",
				DashboardTitle:  "Dashboard",
				WelcomeMessage:  "Welcome to your application portal",
				HomePageMessage: "Discover your personalized workspace",
			}
		} else {
			return nil, err
		}
	}

	// Update cache
	homeCache.mu.Lock()
	homeCache.appSettings = &CachedData{
		Data:      settings,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	homeCache.mu.Unlock()

	return &settings, nil
}

// Helper: Get hero messages with cache
func (h *HomeHandler) getHeroMessagesWithCache() ([]models.HeroMessage, error) {
	homeCache.mu.RLock()
	cached := homeCache.heroMessages
	homeCache.mu.RUnlock()

	if cached != nil && cached.Data != nil && time.Now().Before(cached.ExpiresAt) {
		return cached.Data.([]models.HeroMessage), nil
	}

	// Cache miss - query database
	var messages []models.HeroMessage
	err := h.db.Where("is_active = ?", true).Order("created_at DESC").Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Update cache
	homeCache.mu.Lock()
	homeCache.heroMessages = &CachedData{
		Data:      messages,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	homeCache.mu.Unlock()

	return messages, nil
}

// Helper: Get stats based on role
func (h *HomeHandler) getStats(role string, managedGroupIDs []uint, groups viewerGroups) *HomeStats {
	stats := &HomeStats{}

	if role == "admin" {
		// Count all active users
		h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.TotalUsers)

		// Count all active groups
		h.db.Model(&models.Group{}).Where("is_active = ?", true).Count(&stats.TotalGroups)

		// Count all app groups
		h.db.Model(&models.AppGroup{}).Count(&stats.TotalAppGroups)

		// Count all active applications
		h.db.Model(&models.Application{}).Where("is_active = ?", true).Count(&stats.TotalApps)

		// Count all published news
		h.db.Model(&models.News{}).Where("is_published = ?", true).Count(&stats.TotalNews)

		// Count all published events
		h.db.Model(&models.Event{}).Where("is_published = ?", true).Count(&stats.TotalEvents)

		// Count all polls
		h.db.Model(&models.Poll{}).Count(&stats.TotalPolls)
	} else if len(managedGroupIDs) > 0 {
		stats.ManagedGroupsCount = int64(len(managedGroupIDs))

		// Count total members in managed groups (distinct users)
		h.db.Model(&models.User{}).
			Joins("JOIN user_groups ON users.id = user_groups.user_id").
			Where("user_groups.group_id IN ? AND users.is_active = ?", managedGroupIDs, true).
			Distinct("users.id").
			Count(&stats.TotalMembersCount)

		// Count private app groups owned by managed groups
		h.db.Model(&models.AppGroup{}).
			Where("is_private = ? AND owner_group_id IN ?", true, managedGroupIDs).
			Count(&stats.ManagedAppGroupsCount)

		// Count apps in managed private app groups
		h.db.Model(&models.Application{}).
			Joins("JOIN app_groups ON applications.app_group_id = app_groups.id").
			Where("app_groups.owner_group_id IN ? AND applications.is_active = ?", managedGroupIDs, true).
			Count(&stats.ManagedAppsCount)

		// Count published news (group admins can manage all news)
		h.db.Model(&models.News{}).Where("is_published = ?", true).Count(&stats.ManagedNewsCount)

		// Count polls accessible to managed groups
		h.db.Model(&models.Poll{}).
			Joins("JOIN poll_target_groups ON polls.id = poll_target_groups.poll_id").
			Where("poll_target_groups.group_id IN ?", managedGroupIDs).
			Distinct("polls.id").
			Count(&stats.ManagedPollsCount)
	} else {
		// Regular user stats
		userGroupIDs := groups.member

		// Count accessible apps (public + in user's groups)
		h.db.Model(&models.Application{}).
			Joins("LEFT JOIN app_groups ON applications.app_group_id = app_groups.id").
			Joins("LEFT JOIN group_app_groups ON group_app_groups.app_group_id = app_groups.id").
			Where("(app_groups.is_private = ? OR group_app_groups.group_id IN ?) AND applications.is_active = ?", false, userGroupIDs, true).
			Distinct("applications.id").
			Count(&stats.TotalAccessibleApps)

		// Count accessible events (public + in user's groups)
		h.db.Model(&models.Event{}).
			Where("is_published = ?", true).
			Count(&stats.TotalAccessibleEvents)

		// Count all published news
		h.db.Model(&models.News{}).Where("is_published = ?", true).Count(&stats.TotalNews)

		// Count polls (basic count - frontend filters by accessibility)
		h.db.Model(&models.Poll{}).Count(&stats.TotalPolls)
	}

	return stats
}

// Helper: Get gamification summary for a user
func (h *HomeHandler) getGamificationSummary(userID uint) *GamificationSummary {
	var profile models.GamificationProfile
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			profile = models.GamificationProfile{UserID: userID, Level: 1, XP: 0}
			h.db.Create(&profile)
		} else {
			return nil
		}
	}

	// Calculate progress
	currentXP := profile.XP
	currentLevel := profile.Level
	xpForCurrentLevel := int64(math.Pow(float64(currentLevel-1), 2) * 100)
	xpForNextLevel := int64(math.Pow(float64(currentLevel), 2) * 100)

	rangeNeed := xpForNextLevel - xpForCurrentLevel
	xpInRange := currentXP - xpForCurrentLevel

	progressPercent := 0
	if rangeNeed > 0 {
		progressPercent = int((float64(xpInRange) / float64(rangeNeed)) * 100)
	}

	// Fetch recent badges + total count
	var userAchievements []models.UserAchievement
	h.db.Preload("Achievement").
		Where("user_id = ?", userID).
		Order("user_achievements.unlocked_at DESC").
		Limit(5).
		Find(&userAchievements)

	recentBadges := make([]models.Achievement, 0)
	for _, ua := range userAchievements {
		recentBadges = append(recentBadges, ua.Achievement)
	}

	// Total count (sans limite) : un COUNT plutôt que charger toutes les lignes
	var totalAchievements int64
	h.db.Model(&models.UserAchievement{}).Where("user_id = ?", userID).Count(&totalAchievements)

	// Compute rank (position in leaderboard by XP descending)
	var rank int64
	h.db.Model(&models.GamificationProfile{}).
		Where("xp > ? AND deleted_at IS NULL", profile.XP).
		Count(&rank)

	return &GamificationSummary{
		Level:           profile.Level,
		XP:              profile.XP,
		NextLevelXP:     xpForNextLevel,
		ProgressPercent: progressPercent,
		StreakDays:      profile.LoginStreak,
		BadgesCount:     int(totalAchievements),
		Rank:            int(rank) + 1,
		RecentBadges:    recentBadges,
	}
}

// Cache implementation
type CachedData struct {
	Data      interface{}
	ExpiresAt time.Time
}

type HomeCache struct {
	mu            sync.RWMutex
	announcements *CachedData
	appSettings   *CachedData
	heroMessages  *CachedData
}

var homeCache = &HomeCache{
	announcements: &CachedData{},
	appSettings:   &CachedData{},
	heroMessages:  &CachedData{},
}

// InvalidateAnnouncements can be called from admin handlers when announcements are modified
func (cache *HomeCache) InvalidateAnnouncements() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.announcements = &CachedData{}
}

// getPendingItems calculates personalized pending counts for the hero banner:
// - unread news notifications
// - active polls the user hasn't voted on
// - new apps added in the last 7 days (accessible to the user)
func (h *HomeHandler) getPendingItems(userID uint, role string, groups viewerGroups) *PendingItems {
	pending := &PendingItems{}

	// Unread news notifications
	h.db.Model(&models.Notification{}).
		Where("user_id = ? AND type = ? AND is_read = ?", userID, "news", false).
		Count(&pending.UnreadNews)

	// Active polls the user hasn't voted on
	var activePollIDs []uint
	pollQuery := h.db.Model(&models.Poll{}).Where("is_active = ?", true)
	if role != "admin" {
		// Respect poll visibility: global polls + polls targeting user's groups
		pollQuery = pollQuery.Scopes(pollTargets.visibleTo(groups.member))
	}
	pollQuery.Pluck("id", &activePollIDs)

	if len(activePollIDs) > 0 {
		var votedPollIDs []uint
		h.db.Table("poll_votes").
			Where("user_id = ? AND poll_id IN ?", userID, activePollIDs).
			Distinct("poll_id").
			Pluck("poll_id", &votedPollIDs)
		pending.UnansweredPolls = int64(len(activePollIDs) - len(votedPollIDs))
		if pending.UnansweredPolls < 0 {
			pending.UnansweredPolls = 0
		}
	}

	// New apps added in the last 7 days accessible to the user
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if role == "admin" {
		h.db.Model(&models.Application{}).
			Where("is_active = ? AND created_at >= ?", true, sevenDaysAgo).
			Count(&pending.NewAppsCount)
	} else {
		h.db.Model(&models.Application{}).
			Distinct("applications.id").
			Joins("JOIN app_groups ON applications.app_group_id = app_groups.id").
			Joins("LEFT JOIN group_app_groups ON app_groups.id = group_app_groups.app_group_id").
			Joins("LEFT JOIN user_groups ON group_app_groups.group_id = user_groups.group_id AND user_groups.user_id = ?", userID).
			Joins("LEFT JOIN group_admins ON group_app_groups.group_id = group_admins.group_id AND group_admins.user_id = ?", userID).
			Where("applications.is_active = ? AND applications.created_at >= ? AND (app_groups.is_private = ? OR user_groups.user_id = ? OR group_admins.user_id = ?)",
				true, sevenDaysAgo, false, userID, userID).
			Count(&pending.NewAppsCount)
	}

	return pending
}

// InvalidateAppSettings can be called from admin handlers when settings are modified
func (cache *HomeCache) InvalidateAppSettings() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.appSettings = &CachedData{}
}

// InvalidateHeroMessages can be called from admin handlers when hero messages are modified
func (cache *HomeCache) InvalidateHeroMessages() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.heroMessages = &CachedData{}
}
