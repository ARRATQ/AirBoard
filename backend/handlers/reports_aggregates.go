package handlers

import (
	"log"
	"time"
)

// Agrégats des rapports : une requête GROUP BY par métrique, quel que soit le nombre
// d'utilisateurs ou de groupes (au lieu de ~9 requêtes par utilisateur et ~12 par groupe).

// countBy exécute une requête qui renvoie des colonnes (id, n) et les range par id.
func (h *ReportsHandler) countBy(label, query string, args ...interface{}) map[uint]int64 {
	var rows []idCount
	if err := h.db.Raw(query, args...).Scan(&rows).Error; err != nil {
		log.Printf("[REPORTS] %s: %v", label, err)
	}
	return countsByID(rows)
}

// ---- Rapport par utilisateur ----

type userActivityTotals struct {
	appClicks, newsRead, reactionsGiven, pollVotes       map[uint]int64
	appsCreated, appsCreatedTotal, articles, views, earn map[uint]int64
}

func (h *ReportsHandler) userActivityTotals(from, to time.Time) userActivityTotals {
	return userActivityTotals{
		appClicks: h.countBy("clicks by user",
			`SELECT user_id AS id, COUNT(*) AS n FROM application_clicks
			 WHERE clicked_at BETWEEN ? AND ? GROUP BY user_id`, from, to),
		newsRead: h.countBy("reads by user",
			`SELECT user_id AS id, COUNT(*) AS n FROM news_reads
			 WHERE read_at BETWEEN ? AND ? GROUP BY user_id`, from, to),
		reactionsGiven: h.countBy("reactions by user",
			`SELECT user_id AS id, COUNT(*) AS n FROM news_reactions
			 WHERE created_at BETWEEN ? AND ? GROUP BY user_id`, from, to),
		pollVotes: h.countBy("votes by user",
			`SELECT user_id AS id, COUNT(*) AS n FROM poll_votes
			 WHERE voted_at BETWEEN ? AND ? GROUP BY user_id`, from, to),
		appsCreated: h.countBy("apps created by user",
			`SELECT created_by_id AS id, COUNT(*) AS n FROM applications
			 WHERE created_by_id IS NOT NULL AND created_at BETWEEN ? AND ? AND deleted_at IS NULL
			 GROUP BY created_by_id`, from, to),
		appsCreatedTotal: h.countBy("apps total by user",
			`SELECT created_by_id AS id, COUNT(*) AS n FROM applications
			 WHERE created_by_id IS NOT NULL AND deleted_at IS NULL GROUP BY created_by_id`),
		articles: h.countBy("articles by author",
			`SELECT author_id AS id, COUNT(*) AS n FROM news
			 WHERE is_published = true AND created_at BETWEEN ? AND ? AND deleted_at IS NULL
			 GROUP BY author_id`, from, to),
		views: h.countBy("views by author",
			`SELECT author_id AS id, COALESCE(SUM(view_count), 0) AS n FROM news
			 WHERE is_published = true AND deleted_at IS NULL GROUP BY author_id`),
		earn: h.countBy("reactions earned by author",
			`SELECT news.author_id AS id, COUNT(*) AS n FROM news_reactions
			 JOIN news ON news.id = news_reactions.news_id
			 WHERE news.deleted_at IS NULL AND news_reactions.created_at BETWEEN ? AND ?
			 GROUP BY news.author_id`, from, to),
	}
}

// ---- Rapport par groupe (membres via user_groups) ----

type groupActivityTotals struct {
	members, activeMembers, appClicks, newsRead, reactions, pollVotes map[uint]int64
	appsCreated, appsCreatedTotal, articles                           map[uint]int64
	topApps                                                           map[uint][]AppSummary
	topNews                                                           map[uint][]NewsSummary
}

func (h *ReportsHandler) groupActivityTotals(from, to time.Time) groupActivityTotals {
	t := groupActivityTotals{
		members: h.countBy("members by group",
			`SELECT group_id AS id, COUNT(*) AS n FROM user_groups GROUP BY group_id`),
		// Actif = au moins un clic ou une lecture dans la période
		activeMembers: h.countBy("active members by group",
			`SELECT ug.group_id AS id, COUNT(DISTINCT ug.user_id) AS n FROM user_groups ug
			 WHERE ug.user_id IN (SELECT user_id FROM application_clicks WHERE clicked_at BETWEEN ? AND ?)
			    OR ug.user_id IN (SELECT user_id FROM news_reads WHERE read_at BETWEEN ? AND ?)
			 GROUP BY ug.group_id`, from, to, from, to),
		appClicks: h.countBy("clicks by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM application_clicks
			 JOIN user_groups ug ON ug.user_id = application_clicks.user_id
			 WHERE application_clicks.clicked_at BETWEEN ? AND ? GROUP BY ug.group_id`, from, to),
		newsRead: h.countBy("reads by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM news_reads
			 JOIN user_groups ug ON ug.user_id = news_reads.user_id
			 WHERE news_reads.read_at BETWEEN ? AND ? GROUP BY ug.group_id`, from, to),
		reactions: h.countBy("reactions by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM news_reactions
			 JOIN user_groups ug ON ug.user_id = news_reactions.user_id
			 WHERE news_reactions.created_at BETWEEN ? AND ? GROUP BY ug.group_id`, from, to),
		pollVotes: h.countBy("votes by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM poll_votes
			 JOIN user_groups ug ON ug.user_id = poll_votes.user_id
			 WHERE poll_votes.voted_at BETWEEN ? AND ? GROUP BY ug.group_id`, from, to),
		appsCreated: h.countBy("apps created by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM applications
			 JOIN user_groups ug ON ug.user_id = applications.created_by_id
			 WHERE applications.created_at BETWEEN ? AND ? AND applications.deleted_at IS NULL
			 GROUP BY ug.group_id`, from, to),
		appsCreatedTotal: h.countBy("apps total by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM applications
			 JOIN user_groups ug ON ug.user_id = applications.created_by_id
			 WHERE applications.deleted_at IS NULL GROUP BY ug.group_id`),
		articles: h.countBy("articles by group",
			`SELECT ug.group_id AS id, COUNT(*) AS n FROM news
			 JOIN user_groups ug ON ug.user_id = news.author_id
			 WHERE news.is_published = true AND news.created_at BETWEEN ? AND ? AND news.deleted_at IS NULL
			 GROUP BY ug.group_id`, from, to),
	}
	t.topApps = h.topAppsByGroup(from, to)
	t.topNews = h.topNewsByGroup(from, to)
	return t
}

// Top 5 des applications cliquées par les membres de chaque groupe.
func (h *ReportsHandler) topAppsByGroup(from, to time.Time) map[uint][]AppSummary {
	type row struct {
		GroupID uint
		AppID   uint
		AppName string
		Icon    string
		Color   string
		Clicks  int64
	}
	var rows []row
	if err := h.db.Raw(`
		SELECT group_id, app_id, app_name, icon, color, clicks FROM (
			SELECT ug.group_id, applications.id AS app_id, applications.name AS app_name,
			       applications.icon, applications.color, COUNT(application_clicks.id) AS clicks,
			       ROW_NUMBER() OVER (PARTITION BY ug.group_id ORDER BY COUNT(application_clicks.id) DESC) AS rn
			FROM application_clicks
			JOIN user_groups ug ON ug.user_id = application_clicks.user_id
			JOIN applications ON applications.id = application_clicks.application_id
			WHERE application_clicks.clicked_at BETWEEN ? AND ?
			GROUP BY ug.group_id, applications.id, applications.name, applications.icon, applications.color
		) ranked
		WHERE rn <= 5
		ORDER BY group_id, rn`, from, to).Scan(&rows).Error; err != nil {
		log.Printf("[REPORTS] top apps by group: %v", err)
	}
	out := make(map[uint][]AppSummary)
	for _, r := range rows {
		out[r.GroupID] = append(out[r.GroupID], AppSummary{r.AppID, r.AppName, r.Icon, r.Color, r.Clicks})
	}
	return out
}

// Top 5 des actualités lues par les membres de chaque groupe.
func (h *ReportsHandler) topNewsByGroup(from, to time.Time) map[uint][]NewsSummary {
	type row struct {
		GroupID   uint
		NewsID    uint
		Title     string
		ReadCount int64
	}
	var rows []row
	if err := h.db.Raw(`
		SELECT group_id, news_id, title, read_count FROM (
			SELECT ug.group_id, news.id AS news_id, news.title, COUNT(news_reads.id) AS read_count,
			       ROW_NUMBER() OVER (PARTITION BY ug.group_id ORDER BY COUNT(news_reads.id) DESC) AS rn
			FROM news_reads
			JOIN user_groups ug ON ug.user_id = news_reads.user_id
			JOIN news ON news.id = news_reads.news_id
			WHERE news_reads.read_at BETWEEN ? AND ? AND news.deleted_at IS NULL
			GROUP BY ug.group_id, news.id, news.title
		) ranked
		WHERE rn <= 5
		ORDER BY group_id, rn`, from, to).Scan(&rows).Error; err != nil {
		log.Printf("[REPORTS] top news by group: %v", err)
	}
	out := make(map[uint][]NewsSummary)
	for _, r := range rows {
		out[r.GroupID] = append(out[r.GroupID], NewsSummary{r.NewsID, r.Title, r.ReadCount})
	}
	return out
}
