package handlers

import (
	"log"

	"airboard/models"

	"gorm.io/gorm"
)

// Compteurs agrégés en une requête GROUP BY au lieu d'un COUNT par élément (N+1).

type idCount struct {
	ID uint
	N  int64
}

func countsByID(rows []idCount) map[uint]int64 {
	out := make(map[uint]int64, len(rows))
	for _, r := range rows {
		out[r.ID] = r.N
	}
	return out
}

func newsIDs(news []models.News) []uint {
	ids := make([]uint, 0, len(news))
	for _, n := range news {
		ids = append(ids, n.ID)
	}
	return ids
}

// applyNewsEngagementCounts remplit CommentCount (commentaires approuvés) et
// ReactionCount (feedbacks) de chaque article : 2 requêtes quel que soit le nombre d'articles.
func applyNewsEngagementCounts(db *gorm.DB, news []models.News) {
	if len(news) == 0 {
		return
	}
	ids := newsIDs(news)

	var commentRows []idCount
	if err := db.Model(&models.Comment{}).
		Select("entity_id AS id, COUNT(*) AS n").
		Where("entity_type = ? AND entity_id IN ? AND is_approved = ?", "news", ids, true).
		Group("entity_id").
		Scan(&commentRows).Error; err != nil {
		log.Printf("[HOME] Failed to count news comments: %v", err)
	}

	var reactionRows []idCount
	if err := db.Model(&models.Feedback{}).
		Select("entity_id AS id, COUNT(*) AS n").
		Where("entity_type = ? AND entity_id IN ?", "news", ids).
		Group("entity_id").
		Scan(&reactionRows).Error; err != nil {
		log.Printf("[HOME] Failed to count news reactions: %v", err)
	}

	comments := countsByID(commentRows)
	reactions := countsByID(reactionRows)
	for i := range news {
		news[i].CommentCount = int(comments[news[i].ID])
		news[i].ReactionCount = int(reactions[news[i].ID])
	}
}

// applyPollVoteCounts remplit TotalVotes de chaque sondage en une requête.
func applyPollVoteCounts(db *gorm.DB, polls []models.Poll) {
	if len(polls) == 0 {
		return
	}
	ids := make([]uint, 0, len(polls))
	for _, p := range polls {
		ids = append(ids, p.ID)
	}

	var rows []idCount
	if err := db.Model(&models.PollVote{}).
		Select("poll_id AS id, COUNT(*) AS n").
		Where("poll_id IN ?", ids).
		Group("poll_id").
		Scan(&rows).Error; err != nil {
		log.Printf("[HOME] Failed to count poll votes: %v", err)
	}

	votes := countsByID(rows)
	for i := range polls {
		polls[i].TotalVotes = int(votes[polls[i].ID])
	}
}

// rankedNewsRef : un article retenu pour un onglet, dans l'ordre d'affichage.
type rankedNewsRef struct {
	ID   uint
	Type string
}

// groupNewsByType reconstruit les onglets dans l'ordre des types, chaque onglet gardant
// l'ordre de refs. Les types sans article sont omis.
func groupNewsByType(types []models.NewsType, refs []rankedNewsRef, byID map[uint]models.News) []NewsGroup {
	perType := make(map[string][]models.News, len(types))
	for _, ref := range refs {
		if n, ok := byID[ref.ID]; ok {
			perType[ref.Type] = append(perType[ref.Type], n)
		}
	}

	var result []NewsGroup
	for _, t := range types {
		news := perType[t.Slug]
		if len(news) == 0 {
			continue
		}
		nt := t
		result = append(result, NewsGroup{Type: &nt, News: news})
	}
	return result
}
