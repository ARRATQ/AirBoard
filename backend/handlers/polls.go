package handlers

import (
	"fmt"
	"log"
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
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PollsHandler struct {
	db                  *gorm.DB
	gamificationService *services.GamificationService
}

func NewPollsHandler(db *gorm.DB, gamificationService *services.GamificationService) *PollsHandler {
	return &PollsHandler{
		db:                  db,
		gamificationService: gamificationService,
	}
}

// GetPolls - Liste des sondages (accessible à tous les utilisateurs connectés)
func (h *PollsHandler) GetPolls(c *gin.Context) {
	var polls []models.Poll

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	offset := (page - 1) * pageSize

	// Construction de la requête
	query := h.db.Model(&models.Poll{}).
		Preload("Author").
		Preload("TargetGroups").
		Preload("Options").
		Preload("News").
		Preload("Announcement")

	log.Printf("[DEBUG GetPolls] Starting query for role=%s, userID=%d, path=%s", c.GetString("role"), c.GetUint("user_id"), c.Request.URL.Path)

	// Explicit soft delete filter (complex WHERE clauses may bypass GORM's automatic filter)
	query = query.Where("deleted_at IS NULL")

	// Filtres
	if status := c.Query("status"); status != "" {
		if status == "active" {
			query = query.Where("is_active = ?", true)
		} else if status == "closed" {
			query = query.Where("is_active = ?", false)
		}
	}

	if pollType := c.Query("type"); pollType != "" {
		query = query.Where("poll_type = ?", pollType)
	}

	// Filtre par news - DOIT être strict et appliqué en premier
	newsIDFilter := c.Query("news_id")
	if newsIDFilter != "" && newsIDFilter != "0" && newsIDFilter != "null" && newsIDFilter != "undefined" {
		log.Printf("[DEBUG GetPolls] Filtering by news_id=%s", newsIDFilter)
		// Convertir en uint pour validation
		newsIDUint, err := strconv.ParseUint(newsIDFilter, 10, 32)
		if err != nil {
			log.Printf("[WARN GetPolls] Invalid news_id format: %s", newsIDFilter)
			apierror.Respond(c, http.StatusBadRequest, "Invalid news_id format")
			return
		} else {
			// Filtre strict : seulement les sondages avec ce news_id exact (pas NULL)
			query = query.Where("polls.news_id = ?", newsIDUint)
		}
	} else {
		log.Printf("[DEBUG GetPolls] No news_id filter applied (news_id=%s)", newsIDFilter)
	}

	// Filtre par annonce
	if announcementID := c.Query("announcement_id"); announcementID != "" {
		announcementIDUint, err := strconv.ParseUint(announcementID, 10, 32)
		if err != nil {
			log.Printf("[WARN GetPolls] Invalid announcement_id format: %s", announcementID)
			apierror.Respond(c, http.StatusBadRequest, "Invalid announcement_id format")
			return
		}
		query = query.Where("polls.announcement_id = ?", announcementIDUint)
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

		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+sanitizedSearch+"%", "%"+sanitizedSearch+"%")
	}

	// Filtre de visibilité par groupes selon le rôle
	// IMPORTANT : Si on filtre par news_id ou announcement_id, on ne doit PAS appliquer
	// le filtre de visibilité par groupes car on veut UNIQUEMENT les sondages liés à cette entité
	userRole := c.GetString("role")
	userID := c.GetUint("user_id")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	// Si on filtre par news_id ou announcement_id, ne pas appliquer le filtre de visibilité
	// car on veut seulement les sondages explicitement liés à cette entité
	isEntityFilter := newsIDFilter != "" || c.Query("announcement_id") != ""

	// Pour les non-admins sur l'interface publique, forcer is_active = true
	// (les admins peuvent voir les sondages inactifs via l'interface d'administration)
	isAdminInterface := strings.HasPrefix(c.Request.URL.Path, "/api/v1/group-admin/polls") ||
		strings.HasPrefix(c.Request.URL.Path, "/api/v1/admin/polls")
	if userRole != "admin" && !isAdminInterface && c.Query("status") == "" {
		query = query.Where("is_active = ?", true)
	}

	switch {
	case userRole == "admin" || isEntityFilter:
		// Admin voit tout ; un filtre par entité renvoie seulement les sondages liés
	case len(managedGroupIDs) > 0 && strings.HasPrefix(c.Request.URL.Path, "/api/v1/group-admin/polls"):
		// Interface d'administration de groupe : seulement les sondages qu'il peut gérer
		query = query.Scopes(pollTargets.manageableBy(userID, managedGroupIDs))
	default:
		// Lecture publique : sondages globaux + ceux des groupes dont l'utilisateur est
		// membre ou administrateur
		query = query.Scopes(pollTargets.visibleTo(memberOrManagedGroupIDs(h.db, userID, managedGroupIDs)))
	}

	// Tri
	query = query.Order(utils.SafeOrderClause(
		c.Query("sort"), c.Query("order"),
		[]string{"created_at", "updated_at", "title", "start_date", "end_date"},
		"created_at", "desc",
	))

	// Compte total
	var total int64
	query.Count(&total)

	// Récupération avec pagination
	if err := query.Offset(offset).Limit(pageSize).Find(&polls).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch polls")
		return
	}

	// Log des résultats pour débogage
	log.Printf("[DEBUG GetPolls] Found %d polls (total: %d) for user_id=%d, role=%s, news_id=%s", len(polls), total, userID, userRole, c.Query("news_id"))
	for i, poll := range polls {
		var newsIDStr string
		if poll.NewsID != nil {
			newsIDStr = fmt.Sprintf("%d", *poll.NewsID)
		} else {
			newsIDStr = "NULL"
		}
		log.Printf("[DEBUG GetPolls] Poll[%d]: id=%d, title=%s, news_id=%s, target_groups=%d", i, poll.ID, poll.Title, newsIDStr, len(poll.TargetGroups))
	}

	// Calcul du nombre de pages
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PollListResponse{
		Polls:      polls,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetPollByID - Récupérer un sondage par son ID
func (h *PollsHandler) GetPollByID(c *gin.Context) {
	id := c.Param("id")

	var poll models.Poll
	if err := h.db.Preload("Author").
		Preload("TargetGroups").
		Preload("Options").
		Preload("News").
		Preload("Announcement").
		First(&poll, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Poll not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch poll")
		}
		return
	}

	// Vérifier les permissions selon le rôle
	userRole := c.GetString("role")
	userID := c.GetUint("user_id")

	if userRole == "admin" {
		// Admin voit tout
		c.JSON(http.StatusOK, poll)
		return
	}

	// group_admin et user : sondage global, ou ciblant un groupe dont l'utilisateur
	// est membre ou administrateur
	if len(poll.TargetGroups) > 0 &&
		!targetsAnyGroup(poll.TargetGroups, memberOrManagedGroupIDs(h.db, userID, middleware.GetManagedGroupIDs(c))) {
		apierror.Respond(c, http.StatusForbidden, "Access denied to this poll")
		return
	}

	// Les non-admins ne peuvent pas voir un sondage inactif
	if !poll.IsActive {
		apierror.Respond(c, http.StatusNotFound, "Poll not found")
		return
	}

	c.JSON(http.StatusOK, poll)
}

// CreatePoll - Créer un sondage (admin/group_admin/editor)
func (h *PollsHandler) CreatePoll(c *gin.Context) {
	var req models.PollRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	if endBeforeStart(req.StartDate, req.EndDate) {
		apierror.Respond(c, http.StatusBadRequest, msgEndBeforeStart)
		return
	}

	// Récupérer l'ID de l'utilisateur connecté
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")

	// Validation : au moins 2 options, max 10
	if len(req.Options) < 2 || len(req.Options) > 10 {
		apierror.Respond(c, http.StatusBadRequest, "Poll must have between 2 and 10 options")
		return
	}

	otherOptionsCount := 0
	for _, optReq := range req.Options {
		if optReq.IsOther {
			otherOptionsCount++
		}
	}
	if otherOptionsCount > 1 {
		apierror.Respond(c, http.StatusBadRequest, "Only one option can be marked as 'other'")
		return
	}

	// Vérifier les permissions pour les groupes cibles (admins de groupe non-global)
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	if userRole != "admin" && len(managedGroupIDs) > 0 && len(req.TargetGroupIDs) > 0 {
		for _, targetGroupID := range req.TargetGroupIDs {
			canManage := false
			for _, managedID := range managedGroupIDs {
				if targetGroupID == managedID {
					canManage = true
					break
				}
			}
			if !canManage {
				apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez cibler que les groupes que vous administrez")
				return
			}
		}
	}

	poll := models.Poll{
		Title:          req.Title,
		Description:    req.Description,
		PollType:       req.PollType,
		IsAnonymous:    req.IsAnonymous,
		IsActive:       req.IsActive,
		ShowResults:    req.ShowResults,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		NewsID:         req.NewsID,
		AnnouncementID: req.AnnouncementID,
		AuthorID:       userID,
	}

	// Créer le sondage
	if err := createKeepingFalse(h.db, c, &poll, "is_active"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to create poll")
		return
	}

	// Créer les options
	for i, optReq := range req.Options {
		option := models.PollOption{
			PollID:  poll.ID,
			Text:    optReq.Text,
			IsOther: optReq.IsOther,
			Order:   i,
		}
		if err := h.db.Create(&option).Error; err != nil {
			// Rollback : supprimer le sondage créé
			h.db.Delete(&poll)
			apierror.Respond(c, http.StatusInternalServerError, "Failed to create poll options")
			return
		}
	}

	// Associer les groupes cibles
	if len(req.TargetGroupIDs) > 0 {
		var groups []models.Group
		h.db.Where("id IN ?", req.TargetGroupIDs).Find(&groups)
		h.db.Model(&poll).Association("TargetGroups").Replace(groups)
	}

	// Recharger avec les relations
	h.db.Preload("Author").
		Preload("TargetGroups").
		Preload("Options").
		Preload("News").
		Preload("Announcement").
		First(&poll, poll.ID)

	// Envoyer des notifications
	if poll.IsActive {
		utils.SafeGo("notify-poll-created", func() {
			notifService := services.NewNotificationService(h.db)

			// Récupérer les utilisateurs à notifier
			var userIDs []uint
			if len(poll.TargetGroups) > 0 {
				// Notifier les membres des groupes cibles
				userIDs = groupMemberIDs(h.db, groupIDsOf(poll.TargetGroups))
			} else {
				// Sondage global : notifier tous les utilisateurs actifs
				h.db.Model(&models.User{}).
					Where("is_active = ?", true).
					Where("id != ?", userID). // Ne pas notifier l'auteur
					Pluck("id", &userIDs)
			}

			// Envoyer la notification
			if len(userIDs) > 0 {
				if err := notifService.NotifyNewPoll(poll.Title, poll.ID, userIDs); err != nil {
					log.Printf("[Notification] Échec de l'envoi de la notification de sondage: %v", err)
				}
			}
		})

		// Accorder de l'XP pour la création du sondage (50 XP)
		utils.SafeGo("xp-poll-create", func() {
			if err := h.gamificationService.AwardXP(userID, 50, "poll_create", fmt.Sprintf("{\"poll_id\": %d}", poll.ID)); err != nil {
				log.Printf("[Gamification] Failed to award XP for poll creation: %v", err)
			}
		})
	}

	c.JSON(http.StatusCreated, poll)
}

// UpdatePoll - Modifier un sondage
func (h *PollsHandler) UpdatePoll(c *gin.Context) {
	id := c.Param("id")

	var poll models.Poll
	if err := h.db.Preload("Options").Preload("TargetGroups").First(&poll, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Poll not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch poll")
		}
		return
	}

	// Vérifier les permissions
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	canEdit := false
	if userRole == "admin" {
		canEdit = true
	} else if poll.AuthorID == userID {
		// L'auteur peut modifier son propre sondage
		canEdit = true
	} else if len(managedGroupIDs) > 0 {
		// Admin de groupe peut modifier les sondages ciblant ses groupes administrés
		if len(poll.TargetGroups) > 0 {
			for _, targetGroup := range poll.TargetGroups {
				for _, managedID := range managedGroupIDs {
					if targetGroup.ID == managedID {
						canEdit = true
						break
					}
				}
				if canEdit {
					break
				}
			}
		}
	}

	if !canEdit {
		apierror.Respond(c, http.StatusForbidden, "You don't have permission to edit this poll")
		return
	}

	var req models.PollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	if endBeforeStart(req.StartDate, req.EndDate) {
		apierror.Respond(c, http.StatusBadRequest, msgEndBeforeStart)
		return
	}

	otherOptionsCount := 0
	for _, optReq := range req.Options {
		if optReq.IsOther {
			otherOptionsCount++
		}
	}
	if otherOptionsCount > 1 {
		apierror.Respond(c, http.StatusBadRequest, "Only one option can be marked as 'other'")
		return
	}

	// Vérifier les permissions pour les nouveaux groupes cibles (admins de groupe non-global)
	if userRole != "admin" && len(managedGroupIDs) > 0 && len(req.TargetGroupIDs) > 0 {
		for _, targetGroupID := range req.TargetGroupIDs {
			canManage := false
			for _, managedID := range managedGroupIDs {
				if targetGroupID == managedID {
					canManage = true
					break
				}
			}
			if !canManage {
				apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez cibler que les groupes que vous administrez")
				return
			}
		}
	}

	// Mise à jour des champs du sondage
	poll.Title = req.Title
	poll.Description = req.Description
	poll.PollType = req.PollType
	poll.IsAnonymous = req.IsAnonymous
	poll.ShowResults = req.ShowResults
	poll.StartDate = req.StartDate
	poll.EndDate = req.EndDate
	poll.NewsID = req.NewsID
	poll.AnnouncementID = req.AnnouncementID

	// Seul admin peut activer/désactiver
	if userRole == "admin" || poll.AuthorID == userID {
		poll.IsActive = req.IsActive
	}

	// Sauvegarder le sondage et ses options dans une transaction. Les options gardées
	// sont mises à jour sur place (leurs votes restent) au lieu d'être recréées ;
	// les votes d'une option retirée sont supprimés, ce qui permet de revoter.
	plan := planPollOptions(poll.ID, poll.Options, req.Options)
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Omit : Save réenregistrerait sinon les options et groupes préchargés
		if err := tx.Omit(clause.Associations).Save(&poll).Error; err != nil {
			return fmt.Errorf("sondage : %w", err)
		}
		for _, opt := range plan.Update {
			if err := tx.Model(&models.PollOption{}).Where("id = ? AND poll_id = ?", opt.ID, poll.ID).
				Updates(map[string]interface{}{"text": opt.Text, "is_other": opt.IsOther, "order": opt.Order}).Error; err != nil {
				return fmt.Errorf("option %d : %w", opt.ID, err)
			}
		}
		for i := range plan.Create {
			if err := tx.Create(&plan.Create[i]).Error; err != nil {
				return fmt.Errorf("nouvelle option : %w", err)
			}
		}
		if len(plan.DeleteIDs) > 0 {
			if err := tx.Where("poll_option_id IN ?", plan.DeleteIDs).Delete(&models.PollVote{}).Error; err != nil {
				return fmt.Errorf("votes des options retirées : %w", err)
			}
			if err := tx.Where("id IN ? AND poll_id = ?", plan.DeleteIDs, poll.ID).Delete(&models.PollOption{}).Error; err != nil {
				return fmt.Errorf("options retirées : %w", err)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("[UpdatePoll] sondage %d : %v", poll.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to update poll")
		return
	}

	// Mettre à jour les groupes cibles
	if req.TargetGroupIDs != nil {
		var groups []models.Group
		h.db.Where("id IN ?", req.TargetGroupIDs).Find(&groups)
		h.db.Model(&poll).Association("TargetGroups").Replace(groups)
	}

	// Recharger avec les relations
	h.db.Preload("Author").
		Preload("TargetGroups").
		Preload("Options").
		Preload("News").
		Preload("Announcement").
		First(&poll, poll.ID)

	c.JSON(http.StatusOK, poll)
}

// DeletePoll - Supprimer un sondage (soft delete)
func (h *PollsHandler) DeletePoll(c *gin.Context) {
	id := c.Param("id")

	var poll models.Poll
	if err := h.db.Preload("TargetGroups").First(&poll, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Poll not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch poll")
		}
		return
	}

	// Vérifier les permissions
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")
	managedGroupIDs := middleware.GetManagedGroupIDs(c)

	canDelete := false
	if userRole == "admin" {
		canDelete = true
	} else if poll.AuthorID == userID {
		canDelete = true
	} else if len(managedGroupIDs) > 0 {
		// Admin de groupe peut supprimer les sondages ciblant ses groupes administrés
		if len(poll.TargetGroups) > 0 {
			for _, targetGroup := range poll.TargetGroups {
				for _, managedID := range managedGroupIDs {
					if targetGroup.ID == managedID {
						canDelete = true
						break
					}
				}
				if canDelete {
					break
				}
			}
		}
	}

	if !canDelete {
		apierror.Respond(c, http.StatusForbidden, "You don't have permission to delete this poll")
		return
	}

	if err := h.db.Delete(&poll).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to delete poll")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll deleted successfully"})
}

// ClosePoll - Fermer un sondage (désactiver)
func (h *PollsHandler) ClosePoll(c *gin.Context) {
	id := c.Param("id")

	var poll models.Poll
	if err := h.db.First(&poll, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Poll not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch poll")
		}
		return
	}

	// Vérifier les permissions (admin ou auteur)
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")

	if userRole != "admin" && poll.AuthorID != userID {
		apierror.Respond(c, http.StatusForbidden, "You don't have permission to close this poll")
		return
	}

	poll.IsActive = false
	if err := h.db.Save(&poll).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to close poll")
		return
	}

	c.JSON(http.StatusOK, poll)
}

// Vote - Voter pour un sondage
func (h *PollsHandler) Vote(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var req models.PollVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	// Vérifier que le sondage existe et est actif
	var poll models.Poll
	if err := h.db.Preload("Options").Preload("TargetGroups").First(&poll, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Poll not found")
		return
	}

	if !poll.IsActive {
		apierror.Respond(c, http.StatusBadRequest, "This poll is closed")
		return
	}

	// Vérifier la période de vote
	now := time.Now()
	if poll.StartDate != nil && now.Before(*poll.StartDate) {
		apierror.Respond(c, http.StatusBadRequest, "This poll has not started yet")
		return
	}
	if poll.EndDate != nil && now.After(*poll.EndDate) {
		apierror.Respond(c, http.StatusBadRequest, "This poll has ended")
		return
	}

	// Vérifier que l'utilisateur a accès au sondage
	hasAccess := canSeeTargeted(h.db, userID, poll.TargetGroups)

	if !hasAccess {
		apierror.Respond(c, http.StatusForbidden, "You don't have access to this poll")
		return
	}

	// Vérifier que les options existent et appartiennent au sondage
	var validOptions []models.PollOption
	if err := h.db.Where("poll_id = ? AND id IN ?", poll.ID, req.PollOptionIDs).Find(&validOptions).Error; err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid poll options")
		return
	}

	if len(validOptions) != len(req.PollOptionIDs) {
		apierror.Respond(c, http.StatusBadRequest, "Some poll options are invalid")
		return
	}

	// Vérifier le type de sondage
	if poll.PollType == "single" && len(req.PollOptionIDs) > 1 {
		apierror.Respond(c, http.StatusBadRequest, "This poll allows only one choice")
		return
	}

	// Vérifier si l'utilisateur a déjà voté
	var existingVotes []models.PollVote
	h.db.Where("poll_id = ? AND user_id = ?", poll.ID, userID).Find(&existingVotes)

	if len(existingVotes) > 0 {
		apierror.Respond(c, http.StatusBadRequest, "You have already voted in this poll")
		return
	}

	// Créer les nouveaux votes
	votesCreated := 0
	for _, optionID := range req.PollOptionIDs {
		vote := models.PollVote{
			PollID:       poll.ID,
			UserID:       userID,
			PollOptionID: optionID,
			VotedAt:      time.Now(),
		}
		if err := h.db.Create(&vote).Error; err != nil {
			log.Printf("[Vote] Erreur lors de la création du vote: %v", err)
			apierror.Respond(c, http.StatusInternalServerError, "Failed to record vote")
			return
		}
		votesCreated++
	}

	if votesCreated == 0 {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to record vote")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Vote recorded successfully",
		"votes_count": votesCreated,
		"poll_id":     poll.ID,
		"option_ids":  req.PollOptionIDs,
	})

	// Accorder de l'XP pour le vote (10 XP)
	// On le fait de manière asynchrone pour ne pas ralentir la réponse
	if h.gamificationService != nil {
		hasReward, rewardErr := h.hasAwardedPollVoteXP(userID, poll.ID)
		if rewardErr != nil {
			log.Printf("[Gamification] Failed to check poll vote reward: %v", rewardErr)
		} else if !hasReward {
			utils.SafeGo("xp-poll-vote", func() {
				if err := h.gamificationService.AwardXP(userID, 10, "poll_vote", fmt.Sprintf("{\"poll_id\": %d}", poll.ID)); err != nil {
					log.Printf("[Gamification] Failed to award XP for poll vote: %v", err)
				}
			})
		}
	}
}

// RemoveVote - Retirer le vote d'un sondage (utilisateur connecté)
func (h *PollsHandler) RemoveVote(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var poll models.Poll
	if err := h.db.Preload("TargetGroups").First(&poll, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Poll not found")
		return
	}

	if !poll.IsActive {
		apierror.Respond(c, http.StatusBadRequest, "This poll is closed")
		return
	}

	now := time.Now()
	if poll.EndDate != nil && now.After(*poll.EndDate) {
		apierror.Respond(c, http.StatusBadRequest, "This poll has ended")
		return
	}

	// Vérifier l'accès utilisateur au sondage
	hasAccess := canSeeTargeted(h.db, userID, poll.TargetGroups)

	if !hasAccess {
		apierror.Respond(c, http.StatusForbidden, "You don't have access to this poll")
		return
	}

	result := h.db.Where("poll_id = ? AND user_id = ?", poll.ID, userID).Delete(&models.PollVote{})
	if result.Error != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to remove vote")
		return
	}

	if result.RowsAffected == 0 {
		apierror.Respond(c, http.StatusBadRequest, "You have not voted in this poll")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Vote removed successfully",
		"removed_votes": result.RowsAffected,
		"poll_id":       poll.ID,
	})
}

// GetPollResults - Récupérer les résultats d'un sondage
func (h *PollsHandler) GetPollResults(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")
	userRole := c.GetString("role")

	var poll models.Poll
	if err := h.db.Preload("Author").
		Preload("TargetGroups").
		Preload("Options").
		First(&poll, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Poll not found")
		return
	}

	// Vérifier que l'utilisateur a accès au sondage
	hasAccess := userRole == "admin" || poll.AuthorID == userID ||
		canSeeTargeted(h.db, userID, poll.TargetGroups)

	if !hasAccess {
		apierror.Respond(c, http.StatusForbidden, "You don't have access to this poll")
		return
	}

	// Vérifier si l'utilisateur peut voir les résultats selon ShowResults
	var userVotes []models.PollVote
	h.db.Where("poll_id = ? AND user_id = ?", poll.ID, userID).Find(&userVotes)
	userVoted := len(userVotes) > 0

	canSeeResults := false
	if poll.ShowResults == "always" {
		canSeeResults = true
	} else if poll.ShowResults == "after" && userVoted {
		canSeeResults = true
	} else if poll.ShowResults == "closed" && !poll.IsActive {
		canSeeResults = true
	}

	// Admin et auteur peuvent toujours voir les résultats
	if userRole == "admin" || poll.AuthorID == userID {
		canSeeResults = true
	}

	if !canSeeResults {
		apierror.Respond(c, http.StatusForbidden, "Results not available yet")
		return
	}

	// Compter les votes totaux et les votants uniques
	var totalVotes int64
	var uniqueVoters int64
	h.db.Model(&models.PollVote{}).Where("poll_id = ?", poll.ID).Count(&totalVotes)
	h.db.Model(&models.PollVote{}).Where("poll_id = ?", poll.ID).Distinct("user_id").Count(&uniqueVoters)

	// Calculer les résultats par option
	var optionResults []models.PollOptionResult
	for _, option := range poll.Options {
		var voteCount int64
		h.db.Model(&models.PollVote{}).Where("poll_option_id = ?", option.ID).Count(&voteCount)

		percentage := 0.0
		if uniqueVoters > 0 {
			percentage = (float64(voteCount) / float64(uniqueVoters)) * 100
		}

		optionResults = append(optionResults, models.PollOptionResult{
			PollOption: option,
			VoteCount:  voteCount,
			Percentage: percentage,
		})
	}

	// Récupérer les votes de l'utilisateur
	var userVoteIDs []uint
	for _, vote := range userVotes {
		userVoteIDs = append(userVoteIDs, vote.PollOptionID)
	}

	// Détails des votants (seulement si non anonyme ET admin/auteur)
	var voterDetails []models.PollVoterDetail
	if !poll.IsAnonymous && (userRole == "admin" || poll.AuthorID == userID) {
		var votes []models.PollVote
		h.db.Preload("User").Where("poll_id = ?", poll.ID).Find(&votes)
		for _, vote := range votes {
			voterDetails = append(voterDetails, models.PollVoterDetail{
				User:         vote.User,
				PollOptionID: vote.PollOptionID,
				VotedAt:      vote.VotedAt,
			})
		}
	}

	response := models.PollResultsResponse{
		Poll:         poll,
		TotalVotes:   totalVotes,
		UniqueVoters: uniqueVoters,
		Options:      optionResults,
		UserVoted:    userVoted,
		UserVotes:    userVoteIDs,
		VoterDetails: voterDetails,
	}

	c.JSON(http.StatusOK, response)
}

// GetAnalytics - Statistiques des sondages (admin)
func (h *PollsHandler) GetAnalytics(c *gin.Context) {
	var stats models.PollStatsResponse

	// Total sondages
	h.db.Model(&models.Poll{}).Count(&stats.TotalPolls)

	// Sondages actifs
	h.db.Model(&models.Poll{}).Where("is_active = ?", true).Count(&stats.ActivePolls)

	// Sondages fermés
	h.db.Model(&models.Poll{}).Where("is_active = ?", false).Count(&stats.ClosedPolls)

	// Total votes
	h.db.Model(&models.PollVote{}).Count(&stats.TotalVotes)

	// Total votants uniques
	h.db.Model(&models.PollVote{}).Distinct("user_id").Count(&stats.TotalVoters)

	// Top 5 sondages (par nombre de votes)
	var topPolls []models.Poll
	h.db.Preload("Author").
		Order("created_at DESC").
		Limit(5).
		Find(&topPolls)

	stats.TopPolls = make([]models.PollWithStats, len(topPolls))
	for i, poll := range topPolls {
		var voteCount int64
		var voterCount int64
		h.db.Model(&models.PollVote{}).Where("poll_id = ?", poll.ID).Count(&voteCount)
		h.db.Model(&models.PollVote{}).Where("poll_id = ?", poll.ID).Distinct("user_id").Count(&voterCount)
		stats.TopPolls[i] = models.PollWithStats{
			Poll:       poll,
			VoteCount:  voteCount,
			VoterCount: voterCount,
		}
	}

	// Sondages récents
	h.db.Preload("Author").
		Order("created_at DESC").
		Limit(10).
		Find(&stats.RecentPolls)

	c.JSON(http.StatusOK, stats)
}

func (h *PollsHandler) hasAwardedPollVoteXP(userID uint, pollID uint) (bool, error) {
	metadata := fmt.Sprintf("{\"poll_id\": %d}", pollID)
	var count int64
	err := h.db.Model(&models.XPTransaction{}).
		Where("user_id = ? AND reason = ? AND metadata = ?", userID, "poll_vote", metadata).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
