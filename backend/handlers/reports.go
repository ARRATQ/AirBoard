package handlers

import (
	"net/http"
	"time"

	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportsHandler struct {
	db *gorm.DB
}

func NewReportsHandler(db *gorm.DB) *ReportsHandler {
	return &ReportsHandler{db: db}
}

// parsePeriod extrait les paramètres from/to de la requête (défaut: 30 derniers jours)
func parsePeriod(c *gin.Context) (time.Time, time.Time) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), now.Day()-29, 0, 0, 0, 0, time.UTC)
	to := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		}
	}
	return from, to
}

// ========================
// RAPPORT PAR RÔLE
// ========================

type UserSummary struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Score     int64  `json:"score"`
}

type RoleStat struct {
	Role                 string        `json:"role"`
	MemberCount          int64         `json:"member_count"`
	ActiveMembers        int64         `json:"active_members"`
	AppClicks            int64         `json:"app_clicks"`
	NewsRead             int64         `json:"news_read"`
	ReactionsGiven       int64         `json:"reactions_given"`
	PollVotes            int64         `json:"poll_votes"`
	AppsCreated          int64         `json:"apps_created"`
	AppsCreatedTotal     int64         `json:"apps_created_total"`
	ArticlesPublished    int64         `json:"articles_published"`
	TotalViewsGenerated  int64         `json:"total_views_generated"`
	TotalReactionsEarned int64         `json:"total_reactions_earned"`
	TopContributors      []UserSummary `json:"top_contributors"`
}

type RoleReportResponse struct {
	From      time.Time  `json:"from"`
	To        time.Time  `json:"to"`
	RoleStats []RoleStat `json:"role_stats"`
}

// topContributorsByScore retourne le top 5 des utilisateurs par score d'activité
// whereClause doit filtrer la table "u" (alias de users), ex: "u.role = ?" ou "u.id IN (?)"
func (h *ReportsHandler) topContributorsByScore(from, to time.Time, whereClause string, whereArgs ...interface{}) []UserSummary {
	type TopRow struct {
		UserID    uint
		Username  string
		FirstName string
		LastName  string
		Score     int64
	}
	scoreSQL := `
		SELECT u.id as user_id, u.username, u.first_name, u.last_name,
		(
			COALESCE((SELECT COUNT(*) FROM application_clicks WHERE user_id = u.id AND clicked_at BETWEEN ? AND ?), 0) * 1 +
			COALESCE((SELECT COUNT(*) FROM news_reads WHERE user_id = u.id AND read_at BETWEEN ? AND ?), 0) * 2 +
			COALESCE((SELECT COUNT(*) FROM news_reactions WHERE user_id = u.id AND created_at BETWEEN ? AND ?), 0) * 1 +
			COALESCE((SELECT COUNT(*) FROM poll_votes WHERE user_id = u.id AND voted_at BETWEEN ? AND ?), 0) * 1 +
			COALESCE((SELECT COUNT(*) FROM applications WHERE created_by_id = u.id AND deleted_at IS NULL), 0) * 5 +
			COALESCE((SELECT COUNT(*) FROM news WHERE author_id = u.id AND is_published = true AND created_at BETWEEN ? AND ? AND deleted_at IS NULL), 0) * 10
		) as score
		FROM users u
		WHERE u.deleted_at IS NULL AND ` + whereClause + `
		ORDER BY score DESC
		LIMIT 5`

	args := []interface{}{from, to, from, to, from, to, from, to, from, to}
	args = append(args, whereArgs...)

	var rows []TopRow
	h.db.Raw(scoreSQL, args...).Scan(&rows)

	result := make([]UserSummary, 0, len(rows))
	for _, r := range rows {
		result = append(result, UserSummary{r.UserID, r.Username, r.FirstName, r.LastName, r.Score})
	}
	return result
}

// GetRoleReport retourne les métriques d'activité par rôle utilisateur
func (h *ReportsHandler) GetRoleReport(c *gin.Context) {
	from, to := parsePeriod(c)

	// Helper : construit les conditions WHERE selon le type de rôle
	// Pour "group_admin", on identifie les utilisateurs via la table group_admins (junction),
	// pas via le champ role des utilisateurs.
	type roleFilter struct {
		isGroupAdmin bool
		roleValue    string
	}

	buildFilters := []struct {
		key    string
		filter roleFilter
	}{
		{"admin", roleFilter{roleValue: "admin"}},
		{"group_admin", roleFilter{isGroupAdmin: true}},
		{"editor", roleFilter{roleValue: "editor"}},
		{"user", roleFilter{roleValue: "user"}},
	}

	stats := make([]RoleStat, 0, len(buildFilters))

	for _, rf := range buildFilters {
		var stat RoleStat
		stat.Role = rf.key

		// Sous-requête ou condition selon le type de filtre
		gaSubQ := h.db.Table("group_admins").Select("DISTINCT user_id")

		if rf.filter.isGroupAdmin {
			// Admins de groupe : identifiés par leur présence dans group_admins
			h.db.Table("users").
				Where("id IN (?) AND deleted_at IS NULL", gaSubQ).
				Count(&stat.MemberCount)

			var clickUsers, readUsers []uint
			h.db.Model(&models.ApplicationClick{}).
				Where("user_id IN (?) AND clicked_at BETWEEN ? AND ?", gaSubQ, from, to).
				Distinct("user_id").Pluck("user_id", &clickUsers)
			h.db.Model(&models.NewsRead{}).
				Where("user_id IN (?) AND read_at BETWEEN ? AND ?", gaSubQ, from, to).
				Distinct("user_id").Pluck("user_id", &readUsers)
			seenIDs := map[uint]bool{}
			for _, id := range clickUsers {
				seenIDs[id] = true
			}
			for _, id := range readUsers {
				seenIDs[id] = true
			}
			stat.ActiveMembers = int64(len(seenIDs))

			h.db.Model(&models.ApplicationClick{}).
				Where("user_id IN (?) AND clicked_at BETWEEN ? AND ?", gaSubQ, from, to).
				Count(&stat.AppClicks)
			h.db.Model(&models.NewsRead{}).
				Where("user_id IN (?) AND read_at BETWEEN ? AND ?", gaSubQ, from, to).
				Count(&stat.NewsRead)
			h.db.Model(&models.NewsReaction{}).
				Where("user_id IN (?) AND created_at BETWEEN ? AND ?", gaSubQ, from, to).
				Count(&stat.ReactionsGiven)
			h.db.Model(&models.PollVote{}).
				Where("user_id IN (?) AND voted_at BETWEEN ? AND ?", gaSubQ, from, to).
				Count(&stat.PollVotes)

			// Applications créées dans la période
			h.db.Model(&models.Application{}).
				Where("created_by_id IN (?) AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", gaSubQ, from, to).
				Count(&stat.AppsCreated)
			// Total toutes périodes confondues
			h.db.Model(&models.Application{}).
				Where("created_by_id IN (?) AND deleted_at IS NULL", gaSubQ).
				Count(&stat.AppsCreatedTotal)

			// Production : articles publiés par les admins de groupe
			h.db.Model(&models.News{}).
				Where("author_id IN (?) AND is_published = true AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", gaSubQ, from, to).
				Count(&stat.ArticlesPublished)
			type ViewResult struct{ TotalViews int64 }
			var vr ViewResult
			h.db.Model(&models.News{}).Select("COALESCE(SUM(view_count), 0) as total_views").
				Where("author_id IN (?) AND is_published = true AND deleted_at IS NULL", gaSubQ).Scan(&vr)
			stat.TotalViewsGenerated = vr.TotalViews
			h.db.Model(&models.NewsReaction{}).
				Joins("JOIN news ON news.id = news_reactions.news_id").
				Where("news.author_id IN (?) AND news.deleted_at IS NULL AND news_reactions.created_at BETWEEN ? AND ?", gaSubQ, from, to).
				Count(&stat.TotalReactionsEarned)

			stat.TopContributors = h.topContributorsByScore(from, to, "u.id IN (SELECT DISTINCT user_id FROM group_admins)")
		} else {
			role := rf.filter.roleValue
			h.db.Model(&models.User{}).Where("role = ? AND deleted_at IS NULL", role).Count(&stat.MemberCount)

			var clickUsers, readUsers []uint
			h.db.Model(&models.ApplicationClick{}).
				Joins("JOIN users ON users.id = application_clicks.user_id").
				Where("users.role = ? AND users.deleted_at IS NULL AND application_clicks.clicked_at BETWEEN ? AND ?", role, from, to).
				Distinct("application_clicks.user_id").Pluck("application_clicks.user_id", &clickUsers)
			h.db.Model(&models.NewsRead{}).
				Joins("JOIN users ON users.id = news_reads.user_id").
				Where("users.role = ? AND users.deleted_at IS NULL AND news_reads.read_at BETWEEN ? AND ?", role, from, to).
				Distinct("news_reads.user_id").Pluck("news_reads.user_id", &readUsers)
			seenIDs := map[uint]bool{}
			for _, id := range clickUsers {
				seenIDs[id] = true
			}
			for _, id := range readUsers {
				seenIDs[id] = true
			}
			stat.ActiveMembers = int64(len(seenIDs))

			h.db.Model(&models.ApplicationClick{}).
				Joins("JOIN users ON users.id = application_clicks.user_id").
				Where("users.role = ? AND users.deleted_at IS NULL AND application_clicks.clicked_at BETWEEN ? AND ?", role, from, to).
				Count(&stat.AppClicks)
			h.db.Model(&models.NewsRead{}).
				Joins("JOIN users ON users.id = news_reads.user_id").
				Where("users.role = ? AND users.deleted_at IS NULL AND news_reads.read_at BETWEEN ? AND ?", role, from, to).
				Count(&stat.NewsRead)
			h.db.Model(&models.NewsReaction{}).
				Joins("JOIN users ON users.id = news_reactions.user_id").
				Where("users.role = ? AND users.deleted_at IS NULL AND news_reactions.created_at BETWEEN ? AND ?", role, from, to).
				Count(&stat.ReactionsGiven)
			h.db.Model(&models.PollVote{}).
				Joins("JOIN users ON users.id = poll_votes.user_id").
				Where("users.role = ? AND users.deleted_at IS NULL AND poll_votes.voted_at BETWEEN ? AND ?", role, from, to).
				Count(&stat.PollVotes)

			if role == "admin" || role == "editor" {
				// Applications créées dans la période
				h.db.Model(&models.Application{}).
					Joins("JOIN users ON users.id = applications.created_by_id").
					Where("users.role = ? AND users.deleted_at IS NULL AND applications.created_at BETWEEN ? AND ? AND applications.deleted_at IS NULL", role, from, to).
					Count(&stat.AppsCreated)
				// Total toutes périodes confondues
				h.db.Model(&models.Application{}).
					Joins("JOIN users ON users.id = applications.created_by_id").
					Where("users.role = ? AND users.deleted_at IS NULL AND applications.deleted_at IS NULL", role).
					Count(&stat.AppsCreatedTotal)

				h.db.Model(&models.News{}).
					Joins("JOIN users ON users.id = news.author_id").
					Where("users.role = ? AND users.deleted_at IS NULL AND news.is_published = true AND news.created_at BETWEEN ? AND ? AND news.deleted_at IS NULL", role, from, to).
					Count(&stat.ArticlesPublished)
				type ViewResult struct{ TotalViews int64 }
				var vr ViewResult
				h.db.Model(&models.News{}).Select("COALESCE(SUM(news.view_count), 0) as total_views").
					Joins("JOIN users ON users.id = news.author_id").
					Where("users.role = ? AND users.deleted_at IS NULL AND news.is_published = true AND news.deleted_at IS NULL", role).Scan(&vr)
				stat.TotalViewsGenerated = vr.TotalViews
				h.db.Model(&models.NewsReaction{}).
					Joins("JOIN news ON news.id = news_reactions.news_id").
					Joins("JOIN users ON users.id = news.author_id").
					Where("users.role = ? AND users.deleted_at IS NULL AND news.deleted_at IS NULL AND news_reactions.created_at BETWEEN ? AND ?", role, from, to).
					Count(&stat.TotalReactionsEarned)

				stat.TopContributors = h.topContributorsByScore(from, to, "u.role = ?", role)
			} else {
				stat.TopContributors = h.topContributorsByScore(from, to, "u.role = ?", role)
			}
		}

		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, RoleReportResponse{From: from, To: to, RoleStats: stats})
}

// ========================
// RAPPORT PAR GROUPE
// ========================

type AppSummary struct {
	AppID   uint   `json:"app_id"`
	AppName string `json:"app_name"`
	Icon    string `json:"icon"`
	Color   string `json:"color"`
	Clicks  int64  `json:"clicks"`
}

type NewsSummary struct {
	NewsID    uint   `json:"news_id"`
	Title     string `json:"title"`
	ReadCount int64  `json:"read_count"`
}

type GroupStat struct {
	GroupID          uint          `json:"group_id"`
	GroupName        string        `json:"group_name"`
	GroupColor       string        `json:"group_color"`
	MemberCount      int64         `json:"member_count"`
	ActiveMembers    int64         `json:"active_members"`
	EngagementRate   float64       `json:"engagement_rate"`
	AppClicks        int64         `json:"app_clicks"`
	NewsRead         int64         `json:"news_read"`
	ReactionsGiven   int64         `json:"reactions_given"`
	PollVotes        int64         `json:"poll_votes"`
	AppsCreated      int64         `json:"apps_created"`
	AppsCreatedTotal int64         `json:"apps_created_total"`
	TopApps          []AppSummary  `json:"top_apps"`
	TopNews          []NewsSummary `json:"top_news"`
}

type GroupReportResponse struct {
	From       time.Time   `json:"from"`
	To         time.Time   `json:"to"`
	GroupStats []GroupStat `json:"group_stats"`
}

// GetGroupReport retourne les métriques d'engagement par groupe utilisateur
func (h *ReportsHandler) GetGroupReport(c *gin.Context) {
	from, to := parsePeriod(c)

	var groups []models.Group
	h.db.Where("deleted_at IS NULL AND is_active = true").Find(&groups)

	stats := make([]GroupStat, 0, len(groups))

	for _, group := range groups {
		var stat GroupStat
		stat.GroupID = group.ID
		stat.GroupName = group.Name
		stat.GroupColor = group.Color

		h.db.Table("user_groups").Where("group_id = ?", group.ID).Count(&stat.MemberCount)

		// Membres actifs dans la période
		var clickUsers, readUsers []uint
		h.db.Model(&models.ApplicationClick{}).
			Joins("JOIN user_groups ug ON ug.user_id = application_clicks.user_id").
			Where("ug.group_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", group.ID, from, to).
			Distinct("application_clicks.user_id").
			Pluck("application_clicks.user_id", &clickUsers)
		h.db.Model(&models.NewsRead{}).
			Joins("JOIN user_groups ug ON ug.user_id = news_reads.user_id").
			Where("ug.group_id = ? AND news_reads.read_at BETWEEN ? AND ?", group.ID, from, to).
			Distinct("news_reads.user_id").
			Pluck("news_reads.user_id", &readUsers)
		seenIDs := map[uint]bool{}
		for _, id := range clickUsers {
			seenIDs[id] = true
		}
		for _, id := range readUsers {
			seenIDs[id] = true
		}
		stat.ActiveMembers = int64(len(seenIDs))
		if stat.MemberCount > 0 {
			stat.EngagementRate = float64(stat.ActiveMembers) / float64(stat.MemberCount) * 100
		}

		h.db.Model(&models.ApplicationClick{}).
			Joins("JOIN user_groups ug ON ug.user_id = application_clicks.user_id").
			Where("ug.group_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", group.ID, from, to).
			Count(&stat.AppClicks)

		h.db.Model(&models.NewsRead{}).
			Joins("JOIN user_groups ug ON ug.user_id = news_reads.user_id").
			Where("ug.group_id = ? AND news_reads.read_at BETWEEN ? AND ?", group.ID, from, to).
			Count(&stat.NewsRead)

		h.db.Model(&models.NewsReaction{}).
			Joins("JOIN user_groups ug ON ug.user_id = news_reactions.user_id").
			Where("ug.group_id = ? AND news_reactions.created_at BETWEEN ? AND ?", group.ID, from, to).
			Count(&stat.ReactionsGiven)

		h.db.Model(&models.PollVote{}).
			Joins("JOIN user_groups ug ON ug.user_id = poll_votes.user_id").
			Where("ug.group_id = ? AND poll_votes.voted_at BETWEEN ? AND ?", group.ID, from, to).
			Count(&stat.PollVotes)

		// Applications créées dans la période
		h.db.Model(&models.Application{}).
			Joins("JOIN user_groups ug ON ug.user_id = applications.created_by_id").
			Where("ug.group_id = ? AND applications.created_at BETWEEN ? AND ? AND applications.deleted_at IS NULL", group.ID, from, to).
			Count(&stat.AppsCreated)
		// Total toutes périodes confondues
		h.db.Model(&models.Application{}).
			Joins("JOIN user_groups ug ON ug.user_id = applications.created_by_id").
			Where("ug.group_id = ? AND applications.deleted_at IS NULL", group.ID).
			Count(&stat.AppsCreatedTotal)

		// Top applications du groupe
		type TopAppRow struct {
			AppID   uint
			AppName string
			Icon    string
			Color   string
			Clicks  int64
		}
		var topApps []TopAppRow
		h.db.Model(&models.ApplicationClick{}).
			Select("applications.id as app_id, applications.name as app_name, applications.icon, applications.color, COUNT(application_clicks.id) as clicks").
			Joins("JOIN user_groups ug ON ug.user_id = application_clicks.user_id").
			Joins("JOIN applications ON applications.id = application_clicks.application_id").
			Where("ug.group_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", group.ID, from, to).
			Group("applications.id, applications.name, applications.icon, applications.color").
			Order("clicks DESC").Limit(5).Scan(&topApps)
		for _, a := range topApps {
			stat.TopApps = append(stat.TopApps, AppSummary{a.AppID, a.AppName, a.Icon, a.Color, a.Clicks})
		}

		// Top news lues par le groupe
		type TopNewsRow struct {
			NewsID    uint
			Title     string
			ReadCount int64
		}
		var topNews []TopNewsRow
		h.db.Model(&models.NewsRead{}).
			Select("news.id as news_id, news.title, COUNT(news_reads.id) as read_count").
			Joins("JOIN user_groups ug ON ug.user_id = news_reads.user_id").
			Joins("JOIN news ON news.id = news_reads.news_id").
			Where("ug.group_id = ? AND news_reads.read_at BETWEEN ? AND ? AND news.deleted_at IS NULL", group.ID, from, to).
			Group("news.id, news.title").
			Order("read_count DESC").Limit(5).Scan(&topNews)
		for _, n := range topNews {
			stat.TopNews = append(stat.TopNews, NewsSummary{n.NewsID, n.Title, n.ReadCount})
		}

		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, GroupReportResponse{From: from, To: to, GroupStats: stats})
}

// ========================
// RAPPORT PAR UTILISATEUR - LISTE
// ========================

type UserReportItem struct {
	UserID            uint       `json:"user_id"`
	Username          string     `json:"username"`
	FirstName         string     `json:"first_name"`
	LastName          string     `json:"last_name"`
	Email             string     `json:"email"`
	Role              string     `json:"role"`
	Groups            []string   `json:"groups"`
	IsGroupAdmin      bool       `json:"is_group_admin"`
	MemberSince       time.Time  `json:"member_since"`
	LastLogin         *time.Time `json:"last_login"`
	AppClicks         int64      `json:"app_clicks"`
	NewsRead          int64      `json:"news_read"`
	ReactionsGiven    int64      `json:"reactions_given"`
	PollVotes         int64      `json:"poll_votes"`
	AppsCreated       int64      `json:"apps_created"`
	AppsCreatedTotal  int64      `json:"apps_created_total"`
	ArticlesPublished int64      `json:"articles_published"`
	ViewsGenerated    int64      `json:"views_generated"`
	ReactionsEarned   int64      `json:"reactions_earned"`
	ActivityScore     int64      `json:"activity_score"`
}

type UserReportListResponse struct {
	From  time.Time        `json:"from"`
	To    time.Time        `json:"to"`
	Users []UserReportItem `json:"users"`
	Total int64            `json:"total"`
}

// GetUserReport retourne la liste des utilisateurs avec leurs métriques d'activité
func (h *ReportsHandler) GetUserReport(c *gin.Context) {
	from, to := parsePeriod(c)
	roleFilter := c.Query("role")
	groupIDFilter := c.Query("group_id")

	query := h.db.Model(&models.User{}).Where("users.deleted_at IS NULL AND users.is_active = true")
	if roleFilter != "" {
		query = query.Where("users.role = ?", roleFilter)
	}
	if groupIDFilter != "" {
		query = query.Joins("JOIN user_groups ug_filter ON ug_filter.user_id = users.id").
			Where("ug_filter.group_id = ?", groupIDFilter)
	}

	var users []models.User
	query.Preload("Groups").Find(&users)

	result := make([]UserReportItem, 0, len(users))

	// Pré-calculer les IDs des admins de groupe (table group_admins)
	var groupAdminIDs []uint
	h.db.Table("group_admins").Distinct("user_id").Pluck("user_id", &groupAdminIDs)
	groupAdminSet := map[uint]bool{}
	for _, id := range groupAdminIDs {
		groupAdminSet[id] = true
	}

	for _, user := range users {
		item := UserReportItem{
			UserID:       user.ID,
			Username:     user.Username,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Email:        user.Email,
			Role:         user.Role,
			IsGroupAdmin: groupAdminSet[user.ID],
			MemberSince:  user.CreatedAt,
			LastLogin:    user.LastLogin,
			Groups:       []string{},
		}
		for _, g := range user.Groups {
			item.Groups = append(item.Groups, g.Name)
		}

		h.db.Model(&models.ApplicationClick{}).
			Where("user_id = ? AND clicked_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.AppClicks)
		h.db.Model(&models.NewsRead{}).
			Where("user_id = ? AND read_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.NewsRead)
		h.db.Model(&models.NewsReaction{}).
			Where("user_id = ? AND created_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.ReactionsGiven)
		h.db.Model(&models.PollVote{}).
			Where("user_id = ? AND voted_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.PollVotes)

		// Apps créées dans la période
		h.db.Model(&models.Application{}).
			Where("created_by_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", user.ID, from, to).
			Count(&item.AppsCreated)
		// Total toutes périodes confondues
		h.db.Model(&models.Application{}).
			Where("created_by_id = ? AND deleted_at IS NULL", user.ID).
			Count(&item.AppsCreatedTotal)

		if user.Role == "admin" || user.Role == "editor" || item.IsGroupAdmin {
			h.db.Model(&models.News{}).
				Where("author_id = ? AND is_published = true AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", user.ID, from, to).
				Count(&item.ArticlesPublished)
			type ViewResult struct{ TotalViews int64 }
			var vr ViewResult
			h.db.Model(&models.News{}).Select("COALESCE(SUM(view_count), 0) as total_views").
				Where("author_id = ? AND is_published = true AND deleted_at IS NULL", user.ID).Scan(&vr)
			item.ViewsGenerated = vr.TotalViews
			h.db.Model(&models.NewsReaction{}).
				Joins("JOIN news ON news.id = news_reactions.news_id").
				Where("news.author_id = ? AND news.deleted_at IS NULL AND news_reactions.created_at BETWEEN ? AND ?", user.ID, from, to).
				Count(&item.ReactionsEarned)
		}

		// Score d'activité pondéré
		item.ActivityScore = item.AppClicks + item.NewsRead*2 + item.ReactionsGiven + item.PollVotes + item.AppsCreated*5 + item.ArticlesPublished*10

		result = append(result, item)
	}

	c.JSON(http.StatusOK, UserReportListResponse{
		From:  from,
		To:    to,
		Users: result,
		Total: int64(len(result)),
	})
}

// ========================
// RAPPORT PAR UTILISATEUR - DÉTAIL
// ========================

type MonthlyActivity struct {
	Month             string `json:"month"`
	AppClicks         int64  `json:"app_clicks"`
	NewsRead          int64  `json:"news_read"`
	ReactionsGiven    int64  `json:"reactions_given"`
	AppsCreated       int64  `json:"apps_created"`
	ArticlesPublished int64  `json:"articles_published"`
}

type ArticleSummary struct {
	NewsID        uint      `json:"news_id"`
	Title         string    `json:"title"`
	ViewCount     int       `json:"view_count"`
	ReactionCount int64     `json:"reaction_count"`
	IsPublished   bool      `json:"is_published"`
	CreatedAt     time.Time `json:"created_at"`
}

type UserDetailReportResponse struct {
	User             UserReportItem    `json:"user"`
	MonthlyActivity  []MonthlyActivity `json:"monthly_activity"`
	TopApps          []AppSummary      `json:"top_apps"`
	TopNews          []NewsSummary     `json:"top_news"`
	ArticlesAuthored []ArticleSummary  `json:"articles_authored"`
}

// GetUserDetailReport retourne le profil d'activité complet d'un utilisateur
func (h *ReportsHandler) GetUserDetailReport(c *gin.Context) {
	userIDParam := c.Param("id")
	from, to := parsePeriod(c)

	var user models.User
	if err := h.db.Preload("Groups").Where("deleted_at IS NULL").First(&user, userIDParam).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Utilisateur non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Vérifier si cet utilisateur est admin de groupe
	var gaCount int64
	h.db.Table("group_admins").Where("user_id = ?", user.ID).Count(&gaCount)
	isGroupAdmin := gaCount > 0

	item := UserReportItem{
		UserID:       user.ID,
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		Role:         user.Role,
		IsGroupAdmin: isGroupAdmin,
		MemberSince:  user.CreatedAt,
		LastLogin:    user.LastLogin,
		Groups:       []string{},
	}
	for _, g := range user.Groups {
		item.Groups = append(item.Groups, g.Name)
	}

	h.db.Model(&models.ApplicationClick{}).Where("user_id = ? AND clicked_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.AppClicks)
	h.db.Model(&models.NewsRead{}).Where("user_id = ? AND read_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.NewsRead)
	h.db.Model(&models.NewsReaction{}).Where("user_id = ? AND created_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.ReactionsGiven)
	h.db.Model(&models.PollVote{}).Where("user_id = ? AND voted_at BETWEEN ? AND ?", user.ID, from, to).Count(&item.PollVotes)

	// Apps créées dans la période
	h.db.Model(&models.Application{}).
		Where("created_by_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", user.ID, from, to).
		Count(&item.AppsCreated)
	// Total toutes périodes confondues
	h.db.Model(&models.Application{}).
		Where("created_by_id = ? AND deleted_at IS NULL", user.ID).
		Count(&item.AppsCreatedTotal)

	if user.Role == "admin" || user.Role == "editor" || isGroupAdmin {
		h.db.Model(&models.News{}).
			Where("author_id = ? AND is_published = true AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", user.ID, from, to).
			Count(&item.ArticlesPublished)
		type ViewResult struct{ TotalViews int64 }
		var vr ViewResult
		h.db.Model(&models.News{}).Select("COALESCE(SUM(view_count), 0) as total_views").
			Where("author_id = ? AND is_published = true AND deleted_at IS NULL", user.ID).Scan(&vr)
		item.ViewsGenerated = vr.TotalViews
		h.db.Model(&models.NewsReaction{}).
			Joins("JOIN news ON news.id = news_reactions.news_id").
			Where("news.author_id = ? AND news.deleted_at IS NULL AND news_reactions.created_at BETWEEN ? AND ?", user.ID, from, to).
			Count(&item.ReactionsEarned)
	}
	item.ActivityScore = item.AppClicks + item.NewsRead*2 + item.ReactionsGiven + item.PollVotes + item.AppsCreated*5 + item.ArticlesPublished*10

	// Activité mensuelle (12 derniers mois)
	monthly := make([]MonthlyActivity, 0, 12)
	for i := 11; i >= 0; i-- {
		t := time.Now().AddDate(0, -i, 0)
		mStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		mEnd := mStart.AddDate(0, 1, 0).Add(-time.Second)

		var ma MonthlyActivity
		ma.Month = mStart.Format("2006-01")
		h.db.Model(&models.ApplicationClick{}).Where("user_id = ? AND clicked_at BETWEEN ? AND ?", user.ID, mStart, mEnd).Count(&ma.AppClicks)
		h.db.Model(&models.NewsRead{}).Where("user_id = ? AND read_at BETWEEN ? AND ?", user.ID, mStart, mEnd).Count(&ma.NewsRead)
		h.db.Model(&models.NewsReaction{}).Where("user_id = ? AND created_at BETWEEN ? AND ?", user.ID, mStart, mEnd).Count(&ma.ReactionsGiven)
		h.db.Model(&models.Application{}).
			Where("created_by_id = ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", user.ID, mStart, mEnd).
			Count(&ma.AppsCreated)
		if user.Role == "admin" || user.Role == "editor" || isGroupAdmin {
			h.db.Model(&models.News{}).
				Where("author_id = ? AND is_published = true AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", user.ID, mStart, mEnd).
				Count(&ma.ArticlesPublished)
		}
		monthly = append(monthly, ma)
	}

	// Top applications
	type TopAppRow struct {
		AppID   uint
		AppName string
		Icon    string
		Color   string
		Clicks  int64
	}
	var topAppsRows []TopAppRow
	h.db.Model(&models.ApplicationClick{}).
		Select("applications.id as app_id, applications.name as app_name, applications.icon, applications.color, COUNT(application_clicks.id) as clicks").
		Joins("JOIN applications ON applications.id = application_clicks.application_id").
		Where("application_clicks.user_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", user.ID, from, to).
		Group("applications.id, applications.name, applications.icon, applications.color").
		Order("clicks DESC").Limit(5).Scan(&topAppsRows)
	topApps := make([]AppSummary, 0, len(topAppsRows))
	for _, a := range topAppsRows {
		topApps = append(topApps, AppSummary{a.AppID, a.AppName, a.Icon, a.Color, a.Clicks})
	}

	// Top news lues
	type TopNewsRow struct {
		NewsID    uint
		Title     string
		ReadCount int64
	}
	var topNewsRows []TopNewsRow
	h.db.Model(&models.NewsRead{}).
		Select("news.id as news_id, news.title, COUNT(news_reads.id) as read_count").
		Joins("JOIN news ON news.id = news_reads.news_id").
		Where("news_reads.user_id = ? AND news_reads.read_at BETWEEN ? AND ? AND news.deleted_at IS NULL", user.ID, from, to).
		Group("news.id, news.title").Order("read_count DESC").Limit(5).Scan(&topNewsRows)
	topNews := make([]NewsSummary, 0, len(topNewsRows))
	for _, n := range topNewsRows {
		topNews = append(topNews, NewsSummary{n.NewsID, n.Title, n.ReadCount})
	}

	// Articles rédigés par cet utilisateur
	var articles []models.News
	h.db.Where("author_id = ? AND deleted_at IS NULL", user.ID).Order("created_at DESC").Limit(10).Find(&articles)
	articlesAuthored := make([]ArticleSummary, 0, len(articles))
	for _, a := range articles {
		var rc int64
		h.db.Model(&models.NewsReaction{}).Where("news_id = ?", a.ID).Count(&rc)
		articlesAuthored = append(articlesAuthored, ArticleSummary{
			NewsID:        a.ID,
			Title:         a.Title,
			ViewCount:     a.ViewCount,
			ReactionCount: rc,
			IsPublished:   a.IsPublished,
			CreatedAt:     a.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, UserDetailReportResponse{
		User:             item,
		MonthlyActivity:  monthly,
		TopApps:          topApps,
		TopNews:          topNews,
		ArticlesAuthored: articlesAuthored,
	})
}

// ========================
// RAPPORT PAR GROUPE - DÉTAIL
// ========================

type GroupMonthlyActivity struct {
	Month          string `json:"month"`
	AppClicks      int64  `json:"app_clicks"`
	NewsRead       int64  `json:"news_read"`
	ReactionsGiven int64  `json:"reactions_given"`
	AppsCreated    int64  `json:"apps_created"`
}

type GroupDetailReportResponse struct {
	Group           GroupStat              `json:"group"`
	MonthlyActivity []GroupMonthlyActivity `json:"monthly_activity"`
}

// GetGroupDetailReport retourne le profil d'activité mensuel d'un groupe
func (h *ReportsHandler) GetGroupDetailReport(c *gin.Context) {
	groupIDParam := c.Param("id")
	from, to := parsePeriod(c)

	var group models.Group
	if err := h.db.Where("deleted_at IS NULL").First(&group, groupIDParam).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Not Found",
			Message: "Groupe non trouvé",
			Code:    http.StatusNotFound,
		})
		return
	}

	// Recalculer les stats du groupe pour la période demandée
	var stat GroupStat
	stat.GroupID = group.ID
	stat.GroupName = group.Name
	stat.GroupColor = group.Color

	h.db.Table("user_groups").Where("group_id = ?", group.ID).Count(&stat.MemberCount)

	var clickUsers, readUsers []uint
	h.db.Model(&models.ApplicationClick{}).
		Joins("JOIN user_groups ug ON ug.user_id = application_clicks.user_id").
		Where("ug.group_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", group.ID, from, to).
		Distinct("application_clicks.user_id").Pluck("application_clicks.user_id", &clickUsers)
	h.db.Model(&models.NewsRead{}).
		Joins("JOIN user_groups ug ON ug.user_id = news_reads.user_id").
		Where("ug.group_id = ? AND news_reads.read_at BETWEEN ? AND ?", group.ID, from, to).
		Distinct("news_reads.user_id").Pluck("news_reads.user_id", &readUsers)
	seenIDs := map[uint]bool{}
	for _, id := range clickUsers {
		seenIDs[id] = true
	}
	for _, id := range readUsers {
		seenIDs[id] = true
	}
	stat.ActiveMembers = int64(len(seenIDs))
	if stat.MemberCount > 0 {
		stat.EngagementRate = float64(stat.ActiveMembers) / float64(stat.MemberCount) * 100
	}

	h.db.Model(&models.ApplicationClick{}).
		Joins("JOIN user_groups ug ON ug.user_id = application_clicks.user_id").
		Where("ug.group_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", group.ID, from, to).
		Count(&stat.AppClicks)
	h.db.Model(&models.NewsRead{}).
		Joins("JOIN user_groups ug ON ug.user_id = news_reads.user_id").
		Where("ug.group_id = ? AND news_reads.read_at BETWEEN ? AND ?", group.ID, from, to).
		Count(&stat.NewsRead)
	h.db.Model(&models.NewsReaction{}).
		Joins("JOIN user_groups ug ON ug.user_id = news_reactions.user_id").
		Where("ug.group_id = ? AND news_reactions.created_at BETWEEN ? AND ?", group.ID, from, to).
		Count(&stat.ReactionsGiven)
	h.db.Model(&models.PollVote{}).
		Joins("JOIN user_groups ug ON ug.user_id = poll_votes.user_id").
		Where("ug.group_id = ? AND poll_votes.voted_at BETWEEN ? AND ?", group.ID, from, to).
		Count(&stat.PollVotes)
	h.db.Model(&models.Application{}).
		Joins("JOIN user_groups ug ON ug.user_id = applications.created_by_id").
		Where("ug.group_id = ? AND applications.created_at BETWEEN ? AND ? AND applications.deleted_at IS NULL", group.ID, from, to).
		Count(&stat.AppsCreated)
	h.db.Model(&models.Application{}).
		Joins("JOIN user_groups ug ON ug.user_id = applications.created_by_id").
		Where("ug.group_id = ? AND applications.deleted_at IS NULL", group.ID).
		Count(&stat.AppsCreatedTotal)

	// Activité mensuelle (12 derniers mois)
	monthly := make([]GroupMonthlyActivity, 0, 12)
	for i := 11; i >= 0; i-- {
		t := time.Now().AddDate(0, -i, 0)
		mStart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		mEnd := mStart.AddDate(0, 1, 0).Add(-time.Second)

		var ma GroupMonthlyActivity
		ma.Month = mStart.Format("2006-01")
		h.db.Model(&models.ApplicationClick{}).
			Joins("JOIN user_groups ug ON ug.user_id = application_clicks.user_id").
			Where("ug.group_id = ? AND application_clicks.clicked_at BETWEEN ? AND ?", group.ID, mStart, mEnd).
			Count(&ma.AppClicks)
		h.db.Model(&models.NewsRead{}).
			Joins("JOIN user_groups ug ON ug.user_id = news_reads.user_id").
			Where("ug.group_id = ? AND news_reads.read_at BETWEEN ? AND ?", group.ID, mStart, mEnd).
			Count(&ma.NewsRead)
		h.db.Model(&models.NewsReaction{}).
			Joins("JOIN user_groups ug ON ug.user_id = news_reactions.user_id").
			Where("ug.group_id = ? AND news_reactions.created_at BETWEEN ? AND ?", group.ID, mStart, mEnd).
			Count(&ma.ReactionsGiven)
		h.db.Model(&models.Application{}).
			Joins("JOIN user_groups ug ON ug.user_id = applications.created_by_id").
			Where("ug.group_id = ? AND applications.created_at BETWEEN ? AND ? AND applications.deleted_at IS NULL", group.ID, mStart, mEnd).
			Count(&ma.AppsCreated)
		monthly = append(monthly, ma)
	}

	c.JSON(http.StatusOK, GroupDetailReportResponse{
		Group:           stat,
		MonthlyActivity: monthly,
	})
}
