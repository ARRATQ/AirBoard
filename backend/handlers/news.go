package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"airboard/apierror"
	"airboard/config"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NewsHandler struct {
	db                  *gorm.DB
	config              *config.Config
	gamificationService *services.GamificationService
}

func NewNewsHandler(db *gorm.DB, cfg *config.Config, gs *services.GamificationService) *NewsHandler {
	return &NewsHandler{db: db, config: cfg, gamificationService: gs}
}

// GetNews - Liste des news (accessible à tous les utilisateurs connectés)
func (h *NewsHandler) GetNews(c *gin.Context) {
	var news []models.News

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	// Construction de la requête
	query := h.db.Model(&models.News{}).
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("Reactions").
		Preload("TargetGroups")

	log.Printf("[DEBUG GetNews] Starting query for role=%s, userID=%d, path=%s", c.GetString("role"), c.GetUint("user_id"), c.Request.URL.Path)

	// Explicit soft delete filter (complex WHERE clauses may bypass GORM's automatic filter)
	query = query.Where("deleted_at IS NULL")

	// Filtres
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if priority := c.Query("priority"); priority != "" {
		query = query.Where("priority = ?", priority)
	}

	if newsType := c.Query("type"); newsType != "" {
		query = query.Where("type = ?", newsType)
	}

	// Statut de publication (administration) : appliqué à toute la liste, pas à la page courante.
	// Le filtrage par rôle plus bas reste appliqué en plus : il ne peut que restreindre.
	switch c.Query("status") {
	case "published":
		query = query.Where("is_published = ?", true)
	case "draft":
		query = query.Where("is_published = ?", false)
	}

	// Filtre par tags (supporte plusieurs tags séparés par des virgules)
	if tags := c.Query("tags"); tags != "" {
		tagIDs := strings.Split(tags, ",")
		if len(tagIDs) > 0 {
			query = query.Joins("JOIN news_tags ON news_tags.news_id = news.id").
				Where("news_tags.tag_id IN ?", tagIDs).
				Group("news.id").
				Having("COUNT(DISTINCT news_tags.tag_id) = ?", len(tagIDs))
		}
	} else if tagID := c.Query("tag_id"); tagID != "" {
		// Garde la compatibilité avec l'ancien paramètre tag_id
		query = query.Joins("JOIN news_tags ON news_tags.news_id = news.id").
			Where("news_tags.tag_id = ?", tagID)
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

		query = query.Where("title ILIKE ? OR summary ILIKE ?", "%"+sanitizedSearch+"%", "%"+sanitizedSearch+"%")
	}

	// Filtre published only et visibilité par groupes selon les droits
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	now := time.Now()

	switch {
	case middleware.HasPermissionEverywhere(c, models.PermNewsManage):
		// Admin, ou gestion des actualités sur tout le site : tout (publié + brouillons)
	case len(managedGroupIDs) > 0:
		// Administrateur de groupe. Dans l'interface /group-admin/news : seulement ce qu'il
		// gère (ses articles et ceux qui ciblent ses groupes, brouillons compris). En lecture
		// publique : ses brouillons + les news publiées globales ou de ses groupes
		// (appartenance + administration).
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/group-admin/news") {
			query = query.Scopes(newsTargets.manageableBy(userID, managedGroupIDs))
		} else {
			groupIDs := memberOrManagedGroupIDs(h.db, userID, managedGroupIDs)
			query = query.Scopes(newsTargets.ownOrPublishedVisibleTo(userID, groupIDs, now))
		}
	default:
		// User régulier voit : news publiées globales + celles ciblant ses groupes
		query = query.Scopes(newsTargets.publishedVisibleTo(userGroupIDs(h.db, userID), now))
	}

	// Tri
	sortBy := c.Query("sort")
	if sortBy == "pinned" {
		query = query.Order("is_pinned DESC, published_at DESC")
	} else {
		query = query.Order(utils.SafeOrderClause(
			sortBy, c.Query("order"),
			[]string{"published_at", "created_at", "updated_at", "title", "view_count"},
			"published_at", "desc",
		))
	}

	// Compte total
	var total int64
	query.Count(&total)

	// Récupération avec pagination
	if err := query.Offset(offset).Limit(pageSize).Find(&news).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch news")
		return
	}

	// Calcul du nombre de pages
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	log.Printf("[DEBUG GetNews] Returning %d news (total=%d, page=%d, totalPages=%d)", len(news), total, page, totalPages)
	for i, n := range news {
		log.Printf("[DEBUG GetNews] News[%d]: ID=%d, Title=%s, DeletedAt=%v", i, n.ID, n.Title, n.DeletedAt)
	}

	c.JSON(http.StatusOK, models.NewsListResponse{
		News:       news,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetNewsBySlug - Récupérer une news par son slug
func (h *NewsHandler) GetNewsBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var news models.News
	if err := h.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("Reactions").
		Preload("TargetGroups").
		Where("slug = ?", slug).
		First(&news).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "News not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch news")
		}
		return
	}

	// Vérifier les droits de lecture
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	if canManageNews(c, news.AuthorID, news.TargetGroups) {
		// Qui gère l'article le voit, brouillon compris (pour le modifier)
		c.JSON(http.StatusOK, news)
		return
	}

	// Vérifier si publié
	if !news.IsPublished {
		apierror.Respond(c, http.StatusForbidden, "News not published")
		return
	}

	// Article publié ciblant des groupes : l'utilisateur doit être membre OU admin
	// d'au moins un groupe cible (sans groupe cible, la news est globale)
	if len(news.TargetGroups) > 0 &&
		!targetsAnyGroup(news.TargetGroups, memberOrManagedGroupIDs(h.db, userID, managedGroupIDs)) {
		apierror.Respond(c, http.StatusForbidden, "Access denied to this news")
		return
	}

	c.JSON(http.StatusOK, news)
}

// CreateNews - Créer une news (permission news.manage). Avec une gestion limitée à ses
// groupes, l'article cible au moins un de ses groupes.
func (h *NewsHandler) CreateNews(c *gin.Context) {
	var req models.NewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if !middleware.HasPermissionEverywhere(c, models.PermNewsManage) && len(req.TargetGroupIDs) == 0 {
		apierror.Respond(c, http.StatusBadRequest, msgNeedsTargetGroup)
		return
	}

	// Récupérer l'ID de l'utilisateur connecté
	userID := c.GetUint("user_id")

	news := models.News{
		Title:       req.Title,
		Summary:     req.Summary,
		Content:     req.Content,
		CoverImage:  req.CoverImage,
		Type:        req.Type,
		Priority:    req.Priority,
		IsPinned:    req.IsPinned,
		IsPublished: req.IsPublished,
		PublishedAt: req.PublishedAt,
		ExpiresAt:   req.ExpiresAt,
		CategoryID:  req.CategoryID,
		AuthorID:    userID,
	}

	// Si publié sans date, mettre la date actuelle
	if news.IsPublished && news.PublishedAt == nil {
		now := time.Now()
		news.PublishedAt = &now
	}

	// Créer la news
	if err := h.db.Create(&news).Error; err != nil {
		log.Printf("[ERROR CreateNews] Failed to create news: %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to create news")
		return
	}

	// Associer les tags
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		h.db.Where("id IN ?", req.TagIDs).Find(&tags)
		h.db.Model(&news).Association("Tags").Replace(tags)
	}

	// Associer les groupes cibles ; gestion limitée à ses groupes : seulement les siens
	if len(req.TargetGroupIDs) > 0 {
		managedGroupIDs := middleware.PermissionGroupIDs(c, models.PermNewsManage)

		// Gestion limitée à ses groupes : il ne cible que ses groupes
		if !middleware.HasPermissionEverywhere(c, models.PermNewsManage) {
			for _, targetGroupID := range req.TargetGroupIDs {
				canManage := false
				for _, managedID := range managedGroupIDs {
					if targetGroupID == managedID {
						canManage = true
						break
					}
				}
				if !canManage {
					apierror.Respond(c, http.StatusForbidden, msgTargetOutsideGroups)
					// Supprimer la news créée
					h.db.Delete(&news)
					return
				}
			}
		}

		var groups []models.Group
		h.db.Where("id IN ?", req.TargetGroupIDs).Find(&groups)
		h.db.Model(&news).Association("TargetGroups").Replace(groups)
	}

	// Recharger avec les relations
	h.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		First(&news, news.ID)

	// Envoyer une notification email si l'article est publié
	if news.IsPublished {
		utils.SafeGo("email-news", func() {
			log.Printf("[Email] Tentative d'envoi de notification pour news ID=%d, titre='%s'", news.ID, news.Title)
			// Utiliser la configuration stockée pour éviter de régénérer le secret JWT
			emailService := services.NewEmailService(h.db, h.config)
			var targetGroupIDs []uint
			for _, g := range news.TargetGroups {
				targetGroupIDs = append(targetGroupIDs, g.ID)
			}
			log.Printf("[Email] Groupes cibles pour news %d: %v", news.ID, targetGroupIDs)
			if err := emailService.SendNotification("news", news.ID, targetGroupIDs); err != nil {
				log.Printf("[Email] ❌ ÉCHEC notification news ID=%d: %v", news.ID, err)
			} else {
				log.Printf("[Email] ✅ Notification envoyée avec succès pour news ID=%d", news.ID)
			}
		})

		// Créer des notifications pour les utilisateurs
		utils.SafeGo("notify-news-published", func() {
			notifService := services.NewNotificationService(h.db)

			// Récupérer les utilisateurs à notifier
			var userIDs []uint
			if len(news.TargetGroups) > 0 {
				// Notifier les membres des groupes cibles
				userIDs = groupMemberIDs(h.db, groupIDsOf(news.TargetGroups))
			} else {
				// Article global : notifier tous les utilisateurs actifs
				h.db.Model(&models.User{}).
					Where("is_active = ?", true).
					Where("id != ?", userID). // Ne pas notifier l'auteur
					Pluck("id", &userIDs)
			}

			// Envoyer la notification
			if len(userIDs) > 0 {
				authorName := news.Author.FirstName + " " + news.Author.LastName
				if err := notifService.NotifyNewArticle(news.Title, news.Slug, authorName, userIDs); err != nil {
					log.Printf("[Notification] Échec de l'envoi de la notification: %v", err)
				}
			}
		})
	}

	// Award Contributor XP
	if err := h.gamificationService.AwardXP(userID, 100, "news_publish", fmt.Sprintf("{\"news_id\": %d}", news.ID)); err != nil {
		log.Printf("[Gamification] Error awarding news publish XP for user %d: %v", userID, err)
	}

	c.JSON(http.StatusCreated, news)
}

// UpdateNews - Modifier une news
func (h *NewsHandler) UpdateNews(c *gin.Context) {
	slug := c.Param("id") // Le paramètre est en fait un slug

	var news models.News
	if err := h.db.Where("slug = ?", slug).First(&news).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "News not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch news")
		}
		return
	}

	// Droits : voir canManageNews (permission partout, auteur, ou groupes cibles gérés)
	if !canManageNews(c, news.AuthorID, h.newsTargetGroups(news.ID)) {
		apierror.Respond(c, http.StatusForbidden, "You don't have permission to edit this news")
		return
	}

	var req models.NewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	// Groupes cibles vérifiés avant toute écriture (nil : inchangés)
	targetIDs, ok := h.updatedNewsTargets(c, news.ID, req.TargetGroupIDs)
	if !ok {
		return
	}

	// Mise à jour des champs
	news.Title = req.Title
	news.Summary = req.Summary
	news.Content = req.Content
	news.CoverImage = req.CoverImage
	news.Type = req.Type
	news.Priority = req.Priority
	news.CategoryID = req.CategoryID
	news.ExpiresAt = req.ExpiresAt

	// Seul admin peut épingler
	if c.GetString("role") == "admin" {
		news.IsPinned = req.IsPinned
	}

	// Gestion de la publication
	wasPublished := news.IsPublished
	news.IsPublished = req.IsPublished
	if news.IsPublished && !wasPublished && news.PublishedAt == nil {
		now := time.Now()
		news.PublishedAt = &now
	}
	if req.PublishedAt != nil {
		news.PublishedAt = req.PublishedAt
	}

	// Sauvegarder
	if err := h.db.Save(&news).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to update news")
		return
	}

	// Mettre à jour les tags
	if req.TagIDs != nil {
		var tags []models.Tag
		h.db.Where("id IN ?", req.TagIDs).Find(&tags)
		h.db.Model(&news).Association("Tags").Replace(tags)
	}

	// Mettre à jour les groupes cibles
	if targetIDs != nil {
		var groups []models.Group
		if len(targetIDs) > 0 {
			h.db.Where("id IN ?", targetIDs).Find(&groups)
		}
		h.db.Model(&news).Association("TargetGroups").Replace(groups)
	}

	// Recharger avec les relations
	h.db.Preload("Author").
		Preload("Category").
		Preload("Tags").
		Preload("TargetGroups").
		First(&news, news.ID)

	c.JSON(http.StatusOK, news)
}

// DeleteNews - Supprimer une news (soft delete)
func (h *NewsHandler) DeleteNews(c *gin.Context) {
	id := c.Param("id")

	var news models.News
	if err := h.db.First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "News not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch news")
		}
		return
	}

	// Droits : voir canManageNews (permission partout, auteur, ou groupes cibles gérés)
	userID := c.GetUint("user_id")
	if !canManageNews(c, news.AuthorID, h.newsTargetGroups(news.ID)) {
		log.Printf("[DEBUG DeleteNews] Permission denied for user %d to delete news %d", userID, news.ID)
		apierror.Respond(c, http.StatusForbidden, "You don't have permission to delete this news")
		return
	}

	log.Printf("[DEBUG DeleteNews] Deleting news ID=%d by user %d", news.ID, userID)
	if err := h.db.Delete(&news).Error; err != nil {
		log.Printf("[ERROR DeleteNews] Failed to delete news %d: %v", news.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to delete news")
		return
	}

	log.Printf("[DEBUG DeleteNews] Successfully deleted news ID=%d", news.ID)
	c.JSON(http.StatusOK, gin.H{"message": "News deleted successfully"})
}

// TogglePin - Épingler/désépingler une news (admin uniquement)
func (h *NewsHandler) TogglePin(c *gin.Context) {
	id := c.Param("id")

	var news models.News
	if err := h.db.Preload("TargetGroups").First(&news, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "News not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch news")
		}
		return
	}

	wasNotPinned := !news.IsPinned
	news.IsPinned = !news.IsPinned

	if err := h.db.Save(&news).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to update news")
		return
	}

	// Si l'article est épinglé et publié, envoyer des notifications
	if news.IsPinned && wasNotPinned && news.IsPublished {
		utils.SafeGo("notify-news-pinned", func() {
			notifService := services.NewNotificationService(h.db)

			// Récupérer les utilisateurs à notifier
			var userIDs []uint
			if len(news.TargetGroups) > 0 {
				// Notifier les membres des groupes cibles
				userIDs = groupMemberIDs(h.db, groupIDsOf(news.TargetGroups))
			} else {
				// Article global : notifier tous les utilisateurs actifs
				h.db.Model(&models.User{}).
					Where("is_active = ?", true).
					Pluck("id", &userIDs)
			}

			// Envoyer la notification
			if len(userIDs) > 0 {
				if err := notifService.NotifyPinnedArticle(news.Title, news.Slug, userIDs); err != nil {
					log.Printf("[Notification] Échec de l'envoi de la notification d'épinglage: %v", err)
				}
			}
		})
	}

	c.JSON(http.StatusOK, news)
}

// IncrementView - Incrémenter le compteur de vues et attribuer de l'XP
func (h *NewsHandler) IncrementView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid news ID")
		return
	}
	userID := c.GetUint("user_id")

	// Déjà lu : rien à faire (cas courant, évite une écriture)
	var exists int64
	h.db.Model(&models.NewsRead{}).Where("news_id = ? AND user_id = ?", id, userID).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "View counted"})
		return
	}

	// Enregistrer la lecture avant de compter : avec l'index unique (news_id, user_id),
	// deux requêtes simultanées ne peuvent pas toutes deux compter la vue ni donner l'XP.
	read := models.NewsRead{
		NewsID: uint(id),
		UserID: userID,
		ReadAt: time.Now(),
	}
	result := h.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&read)
	if result.Error != nil {
		log.Printf("[ERROR IncrementView] Failed to record read for news %d: %v", id, result.Error)
		c.JSON(http.StatusOK, gin.H{"message": "View counted"})
		return
	}

	if result.RowsAffected == 1 {
		// Lecture unique : compteur de vues + 20 XP
		if err := h.db.Model(&models.News{}).Where("id = ?", id).
			UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
			log.Printf("[ERROR IncrementView] Failed to increment view for news %d: %v", id, err)
		}
		if err := h.gamificationService.AwardXP(userID, 20, "news_read", fmt.Sprintf("{\"news_id\": %d}", id)); err != nil {
			log.Printf("[Gamification] Error awarding news read XP for user %d: %v", userID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "View counted"})
}

// AddReaction - Ajouter une réaction à une news
func (h *NewsHandler) AddReaction(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var req models.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	// Vérifier si la news existe
	var news models.News
	if err := h.db.First(&news, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "News not found")
		return
	}

	// Vérifier si l'utilisateur a déjà réagi
	var existingReaction models.NewsReaction
	err := h.db.Where("news_id = ? AND user_id = ?", id, userID).First(&existingReaction).Error

	if err == nil {
		// Déjà une réaction, la mettre à jour
		existingReaction.ReactionType = req.ReactionType
		if err := h.db.Save(&existingReaction).Error; err != nil {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to update reaction")
			return
		}
		c.JSON(http.StatusOK, existingReaction)
		return
	}

	// Créer une nouvelle réaction
	newsID, _ := strconv.ParseUint(id, 10, 32)
	reaction := models.NewsReaction{
		NewsID:       uint(newsID),
		UserID:       userID,
		ReactionType: req.ReactionType,
	}

	if err := h.db.Create(&reaction).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to add reaction")
		return
	}

	c.JSON(http.StatusCreated, reaction)
}

// RemoveReaction - Retirer une réaction
func (h *NewsHandler) RemoveReaction(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	result := h.db.Where("news_id = ? AND user_id = ?", id, userID).Delete(&models.NewsReaction{})

	if result.Error != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to remove reaction")
		return
	}

	if result.RowsAffected == 0 {
		apierror.Respond(c, http.StatusNotFound, "Reaction not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

// GetReactions - Récupérer les réactions d'une news avec compteurs
func (h *NewsHandler) GetReactions(c *gin.Context) {
	id := c.Param("id")

	var reactions []models.NewsReaction
	if err := h.db.Where("news_id = ?", id).Find(&reactions).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch reactions")
		return
	}

	// Compter par type
	reactionCounts := make(map[string]int)
	for _, reaction := range reactions {
		reactionCounts[reaction.ReactionType]++
	}

	// Vérifier si l'utilisateur a réagi
	userID := c.GetUint("user_id")
	var userReaction *models.NewsReaction
	for _, reaction := range reactions {
		if reaction.UserID == userID {
			userReaction = &reaction
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"counts":        reactionCounts,
		"total":         len(reactions),
		"user_reaction": userReaction,
	})
}

// GetUnreadCount - Nombre de news non lues (pour le badge)
func (h *NewsHandler) GetUnreadCount(c *gin.Context) {
	// Compter les news publiées depuis la dernière visite
	// Pour MVP simple: toutes les news publiées dans les 30 derniers jours
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

	var count int64
	h.db.Model(&models.News{}).
		Where("is_published = ?", true).
		Where("published_at >= ?", thirtyDaysAgo).
		Count(&count)

	// TODO: Implémenter le système de tracking de lecture (NewsRead)
	// Pour le moment, on retourne juste le compte des news récentes

	c.JSON(http.StatusOK, gin.H{
		"unread_count": count,
	})
}

// GetAnalytics - Statistiques du News Hub (admin/editor)
func (h *NewsHandler) GetAnalytics(c *gin.Context) {
	var stats models.NewsStatsResponse

	// Total news
	h.db.Model(&models.News{}).Count(&stats.TotalNews)

	// News publiées
	h.db.Model(&models.News{}).Where("is_published = ?", true).Count(&stats.PublishedNews)

	// News en brouillon
	h.db.Model(&models.News{}).Where("is_published = ?", false).Count(&stats.DraftNews)

	// Total vues
	h.db.Model(&models.News{}).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.TotalViews)

	// Total réactions
	h.db.Model(&models.NewsReaction{}).Count(&stats.TotalReactions)

	// Top 5 news (par vues)
	var topNews []models.News
	h.db.Preload("Author").
		Preload("Category").
		Order("view_count DESC").
		Limit(5).
		Find(&topNews)

	// Convertir en NewsWithStats avec compteur de réactions
	stats.TopNews = make([]models.NewsWithStats, len(topNews))
	for i, news := range topNews {
		var reactionCount int64
		h.db.Model(&models.NewsReaction{}).Where("news_id = ?", news.ID).Count(&reactionCount)
		stats.TopNews[i] = models.NewsWithStats{
			News:          news,
			ReactionCount: reactionCount,
		}
	}

	// Réactions par type
	rows, err := h.db.Model(&models.NewsReaction{}).
		Select("reaction_type, COUNT(*) as count").
		Group("reaction_type").
		Rows()

	if err == nil {
		defer rows.Close()
		stats.ReactionsByType = make(map[string]int64)
		for rows.Next() {
			var reactionType string
			var count int64
			rows.Scan(&reactionType, &count)
			stats.ReactionsByType[reactionType] = count
		}
	}

	c.JSON(http.StatusOK, stats)
}

// updatedNewsTargets calcule les groupes cibles demandés par une modification (nil :
// inchangés). Avec une gestion limitée à ses groupes, voir mergeGroupAdminTargets, et
// l'article garde au moins un groupe cible. Répond et renvoie false si la demande est refusée.
func (h *NewsHandler) updatedNewsTargets(c *gin.Context, newsID uint, requested []uint) ([]uint, bool) {
	if requested == nil || middleware.HasPermissionEverywhere(c, models.PermNewsManage) {
		return requested, true
	}
	existing := groupIDsOf(h.newsTargetGroups(newsID))
	merged, ok := mergeGroupAdminTargets(existing, requested, middleware.PermissionGroupIDs(c, models.PermNewsManage))
	if !ok {
		apierror.Respond(c, http.StatusForbidden, msgTargetOutsideGroups)
		return nil, false
	}
	if len(merged) == 0 {
		apierror.Respond(c, http.StatusBadRequest, msgNeedsTargetGroup)
		return nil, false
	}
	return merged, true
}

// mergeGroupAdminTargets calcule les groupes cibles d'une actualité modifiée avec une
// gestion limitée à ses groupes. La personne peut cibler ses groupes ; elle peut aussi
// renvoyer une cible déjà présente sur un groupe qu'elle ne gère pas, mais pas en ajouter
// une. Les cibles existantes hors de ses groupes sont toujours conservées. ok vaut false
// si la demande ajoute un groupe qu'elle ne gère pas.
func mergeGroupAdminTargets(existing, requested, managed []uint) ([]uint, bool) {
	isManaged := make(map[uint]bool, len(managed))
	for _, id := range managed {
		isManaged[id] = true
	}
	isExisting := make(map[uint]bool, len(existing))
	for _, id := range existing {
		isExisting[id] = true
	}

	result := make([]uint, 0, len(requested)+len(existing))
	seen := make(map[uint]bool, len(requested)+len(existing))
	add := func(id uint) {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	for _, id := range requested {
		if !isManaged[id] && !isExisting[id] {
			return nil, false
		}
		add(id)
	}
	for _, id := range existing {
		if !isManaged[id] {
			add(id)
		}
	}
	return result, true
}

// canManageNews : l'utilisateur peut modifier ou supprimer l'article — gestion des actualités
// sur tout le site (admin compris), auteur, ou article ciblant un groupe où il gère les
// actualités (portée « ses groupes »). Un article sans groupe
// cible est global : seule la gestion sur tout le site (ou l'auteur) y touche.
func canManageNews(c *gin.Context, authorID uint, targets []models.Group) bool {
	if middleware.HasPermissionEverywhere(c, models.PermNewsManage) {
		return true
	}
	if authorID != 0 && authorID == c.GetUint("user_id") {
		return true
	}
	// targetsAnyGroup traite « aucune cible » comme visible de tous : ici, c'est un refus
	return len(targets) > 0 && targetsAnyGroup(targets, middleware.PermissionGroupIDs(c, models.PermNewsManage))
}

// newsTargetGroups lit les groupes cibles d'un article sans les attacher à l'article chargé
// (un Save ultérieur ne doit pas réécrire l'association).
func (h *NewsHandler) newsTargetGroups(newsID uint) []models.Group {
	var groups []models.Group
	if err := h.db.Table("groups").
		Joins("JOIN news_target_groups ON news_target_groups.group_id = groups.id").
		Where("news_target_groups.news_id = ? AND groups.deleted_at IS NULL", newsID).
		Find(&groups).Error; err != nil {
		log.Printf("news: groupes cibles de %d illisibles : %v", newsID, err)
	}
	return groups
}
