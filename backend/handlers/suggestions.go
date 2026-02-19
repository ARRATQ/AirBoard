package handlers

import (
	"airboard/models"
	"airboard/services"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

type SuggestionsHandler struct {
	db                  *gorm.DB
	gamificationService *services.GamificationService
}

func NewSuggestionsHandler(db *gorm.DB, gamificationService *services.GamificationService) *SuggestionsHandler {
	return &SuggestionsHandler{db: db, gamificationService: gamificationService}
}

func (h *SuggestionsHandler) GetSuggestions(c *gin.Context) {
	h.listSuggestions(c, false, false)
}

func (h *SuggestionsHandler) GetAdminSuggestions(c *gin.Context) {
	h.listSuggestions(c, true, true)
}

func (h *SuggestionsHandler) GetSuggestionByID(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var suggestion models.Suggestion
	if err := h.db.
		Preload("User").
		Preload("CategoryRef").
		Preload("Poll", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("PollOption", func(db *gorm.DB) *gorm.DB { return db.Select("id", "poll_id", "text") }).
		First(&suggestion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestion"})
		return
	}

	if suggestion.IsArchived && c.GetString("role") != "admin" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
		return
	}

	if err := h.enrichVoteStats([]*models.Suggestion{&suggestion}, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load suggestion votes"})
		return
	}

	h.maskAnonymousSuggestion(&suggestion)

	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionsHandler) CreateSuggestion(c *gin.Context) {
	var req models.SuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	title := strings.TrimSpace(req.Title)
	description := strings.TrimSpace(req.Description)

	if title == "" || description == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title and description are required"})
		return
	}

	if len([]rune(title)) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title must contain at least 3 characters"})
		return
	}

	if len([]rune(description)) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Description must contain at least 3 characters"})
		return
	}

	category, err := h.getActiveCategory(req.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or inactive category_id"})
		return
	}

	if req.PollID != nil {
		var poll models.Poll
		if err := h.db.First(&poll, *req.PollID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll_id"})
			return
		}
	}

	if req.PollOptionID != nil {
		var pollOption models.PollOption
		if err := h.db.First(&pollOption, *req.PollOptionID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll_option_id"})
			return
		}
		if req.PollID != nil && pollOption.PollID != *req.PollID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "poll_option_id does not belong to poll_id"})
			return
		}
		if req.PollID == nil {
			pollID := pollOption.PollID
			req.PollID = &pollID
		}
	}

	suggestion := models.Suggestion{
		Title:        title,
		Description:  description,
		Category:     category.Slug,
		CategoryID:   req.CategoryID,
		Status:       "new",
		IsAnonymous:  req.IsAnonymous,
		UserID:       userID,
		PollID:       req.PollID,
		PollOptionID: req.PollOptionID,
	}

	if err := h.db.Create(&suggestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create suggestion"})
		return
	}

	if h.gamificationService != nil {
		go func(userID uint, suggestionID uint) {
			if err := h.gamificationService.AwardXP(userID, 35, "suggestion_create", fmt.Sprintf("{\"suggestion_id\":%d}", suggestionID)); err != nil {
				log.Printf("[Gamification] Failed to award XP for suggestion creation: %v", err)
			}
		}(userID, suggestion.ID)
	}

	if err := h.db.
		Preload("User").
		Preload("CategoryRef").
		Preload("Poll", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("PollOption", func(db *gorm.DB) *gorm.DB { return db.Select("id", "poll_id", "text") }).
		First(&suggestion, suggestion.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created suggestion"})
		return
	}

	if err := h.enrichVoteStats([]*models.Suggestion{&suggestion}, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load suggestion votes"})
		return
	}

	h.maskAnonymousSuggestion(&suggestion)

	c.JSON(http.StatusCreated, suggestion)
}

func (h *SuggestionsHandler) UpdateSuggestionStatus(c *gin.Context) {
	id := c.Param("id")

	var req models.SuggestionStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var suggestion models.Suggestion
	if err := h.db.First(&suggestion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestion"})
		return
	}

	suggestion.Status = req.Status
	if err := h.db.Save(&suggestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update suggestion status"})
		return
	}

	if err := h.db.
		Preload("User").
		Preload("CategoryRef").
		Preload("Poll", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("PollOption", func(db *gorm.DB) *gorm.DB { return db.Select("id", "poll_id", "text") }).
		First(&suggestion, suggestion.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated suggestion"})
		return
	}

	if err := h.enrichVoteStats([]*models.Suggestion{&suggestion}, c.GetUint("user_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load suggestion votes"})
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionsHandler) UpdateSuggestionArchiveState(c *gin.Context) {
	id := c.Param("id")

	var req models.SuggestionArchiveUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var suggestion models.Suggestion
	if err := h.db.First(&suggestion, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestion"})
		return
	}

	suggestion.IsArchived = req.IsArchived
	if err := h.db.Save(&suggestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update archive state"})
		return
	}

	if err := h.db.
		Preload("User").
		Preload("CategoryRef").
		Preload("Poll", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("PollOption", func(db *gorm.DB) *gorm.DB { return db.Select("id", "poll_id", "text") }).
		First(&suggestion, suggestion.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated suggestion"})
		return
	}

	if err := h.enrichVoteStats([]*models.Suggestion{&suggestion}, c.GetUint("user_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load suggestion votes"})
		return
	}

	c.JSON(http.StatusOK, suggestion)
}

func (h *SuggestionsHandler) VoteSuggestion(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid suggestion id"})
		return
	}

	var req models.SuggestionVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	var suggestion models.Suggestion
	if err := h.db.First(&suggestion, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Suggestion not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestion"})
		return
	}

	if req.VoteType == "none" {
		if err := h.db.Where("suggestion_id = ? AND user_id = ?", uint(id), userID).Delete(&models.SuggestionVote{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
			return
		}
	} else {
		isFirstVoteForSuggestion := false
		var vote models.SuggestionVote
		err := h.db.Where("suggestion_id = ? AND user_id = ?", uint(id), userID).First(&vote).Error
		if err == gorm.ErrRecordNotFound {
			isFirstVoteForSuggestion = true
			vote = models.SuggestionVote{SuggestionID: uint(id), UserID: userID, VoteType: req.VoteType}
			if err := h.db.Create(&vote).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save vote"})
				return
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vote"})
			return
		} else {
			if vote.VoteType == req.VoteType {
				if err := h.db.Delete(&vote).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
					return
				}
			} else {
				vote.VoteType = req.VoteType
				if err := h.db.Save(&vote).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote"})
					return
				}
			}
		}

		if isFirstVoteForSuggestion && h.gamificationService != nil {
			hasReward, rewardErr := h.hasAwardedSuggestionVoteXP(userID, uint(id))
			if rewardErr != nil {
				log.Printf("[Gamification] Failed to check suggestion vote reward: %v", rewardErr)
			} else if !hasReward {
				go func(userID uint, suggestionID uint) {
					if err := h.gamificationService.AwardXP(userID, 8, "suggestion_vote", fmt.Sprintf("{\"suggestion_id\":%d}", suggestionID)); err != nil {
						log.Printf("[Gamification] Failed to award XP for suggestion vote: %v", err)
					}
				}(userID, uint(id))
			}
		}
	}

	voteStats, err := h.getSuggestionVoteStats(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load vote stats"})
		return
	}

	c.JSON(http.StatusOK, voteStats)
}

func (h *SuggestionsHandler) GetSuggestionCategories(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "1"

	query := h.db.Model(&models.SuggestionCategory{})
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}

	var categories []models.SuggestionCategory
	if err := query.Order("\"order\" ASC, name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *SuggestionsHandler) GetAdminSuggestionCategories(c *gin.Context) {
	h.GetSuggestionCategories(c)
}

func (h *SuggestionsHandler) CreateSuggestionCategory(c *gin.Context) {
	var req models.SuggestionCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	slug := normalizeSlug(req.Slug)
	if name == "" || slug == "" || !slugRegex.MatchString(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category name or slug"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	category := models.SuggestionCategory{
		Name:     name,
		Slug:     slug,
		Order:    req.Order,
		IsActive: isActive,
	}

	if err := h.db.Create(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create category (name/slug may already exist)"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *SuggestionsHandler) UpdateSuggestionCategory(c *gin.Context) {
	id := c.Param("id")

	var category models.SuggestionCategory
	if err := h.db.First(&category, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category"})
		return
	}

	var req models.SuggestionCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	slug := normalizeSlug(req.Slug)
	if name == "" || slug == "" || !slugRegex.MatchString(slug) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category name or slug"})
		return
	}

	category.Name = name
	category.Slug = slug
	category.Order = req.Order
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}

	if err := h.db.Save(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update category (name/slug may already exist)"})
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *SuggestionsHandler) DeleteSuggestionCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category id"})
		return
	}

	var count int64
	if err := h.db.Model(&models.Suggestion{}).Where("category_id = ?", uint(id)).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check category usage"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is used by existing suggestions"})
		return
	}

	if err := h.db.Delete(&models.SuggestionCategory{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Category deleted"})
}

func (h *SuggestionsHandler) listSuggestions(c *gin.Context, includeAnonymousUser bool, isAdmin bool) {
	page, pageSize := getSuggestionPagination(c)
	search := strings.TrimSpace(c.Query("search"))
	categoryID := strings.TrimSpace(c.Query("category_id"))
	status := strings.TrimSpace(c.Query("status"))
	archived := strings.TrimSpace(c.Query("archived"))

	query := h.db.Model(&models.Suggestion{})

	if !isAdmin {
		query = query.Where("is_archived = ?", false)
	} else if archived == "1" {
		query = query.Where("is_archived = ?", true)
	} else if archived == "0" {
		query = query.Where("is_archived = ?", false)
	}

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ?", like, like)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count suggestions"})
		return
	}

	var suggestions []models.Suggestion
	if err := query.
		Preload("User").
		Preload("CategoryRef").
		Preload("Poll", func(db *gorm.DB) *gorm.DB { return db.Select("id", "title") }).
		Preload("PollOption", func(db *gorm.DB) *gorm.DB { return db.Select("id", "poll_id", "text") }).
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&suggestions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch suggestions"})
		return
	}

	userID := c.GetUint("user_id")
	suggestionPtrs := make([]*models.Suggestion, 0, len(suggestions))
	for i := range suggestions {
		suggestionPtrs = append(suggestionPtrs, &suggestions[i])
	}
	if err := h.enrichVoteStats(suggestionPtrs, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load suggestion votes"})
		return
	}

	if !includeAnonymousUser {
		for i := range suggestions {
			h.maskAnonymousSuggestion(&suggestions[i])
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, models.SuggestionListResponse{
		Suggestions: suggestions,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
	})
}

func (h *SuggestionsHandler) getActiveCategory(categoryID *uint) (*models.SuggestionCategory, error) {
	if categoryID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var category models.SuggestionCategory
	if err := h.db.First(&category, *categoryID).Error; err != nil {
		return nil, err
	}
	if !category.IsActive {
		return nil, gorm.ErrRecordNotFound
	}

	return &category, nil
}

func (h *SuggestionsHandler) enrichVoteStats(suggestions []*models.Suggestion, userID uint) error {
	if len(suggestions) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(suggestions))
	for _, suggestion := range suggestions {
		ids = append(ids, suggestion.ID)
	}

	type voteAggregate struct {
		SuggestionID uint
		VoteType     string
		Count        int64
	}

	var aggregates []voteAggregate
	if err := h.db.Model(&models.SuggestionVote{}).
		Select("suggestion_id, vote_type, COUNT(*) as count").
		Where("suggestion_id IN ?", ids).
		Group("suggestion_id, vote_type").
		Scan(&aggregates).Error; err != nil {
		return err
	}

	upvoteMap := make(map[uint]int64)
	downvoteMap := make(map[uint]int64)
	for _, aggregate := range aggregates {
		if aggregate.VoteType == "up" {
			upvoteMap[aggregate.SuggestionID] = aggregate.Count
		}
		if aggregate.VoteType == "down" {
			downvoteMap[aggregate.SuggestionID] = aggregate.Count
		}
	}

	userVoteMap := make(map[uint]string)
	if userID > 0 {
		var votes []models.SuggestionVote
		if err := h.db.Where("suggestion_id IN ? AND user_id = ?", ids, userID).Find(&votes).Error; err != nil {
			return err
		}
		for _, vote := range votes {
			userVoteMap[vote.SuggestionID] = vote.VoteType
		}
	}

	for _, suggestion := range suggestions {
		suggestion.Upvotes = upvoteMap[suggestion.ID]
		suggestion.Downvotes = downvoteMap[suggestion.ID]
		suggestion.UserVote = userVoteMap[suggestion.ID]
	}

	return nil
}

func (h *SuggestionsHandler) getSuggestionVoteStats(suggestionID uint, userID uint) (models.SuggestionVoteResponse, error) {
	var upvotes int64
	if err := h.db.Model(&models.SuggestionVote{}).
		Where("suggestion_id = ? AND vote_type = ?", suggestionID, "up").
		Count(&upvotes).Error; err != nil {
		return models.SuggestionVoteResponse{}, err
	}

	var downvotes int64
	if err := h.db.Model(&models.SuggestionVote{}).
		Where("suggestion_id = ? AND vote_type = ?", suggestionID, "down").
		Count(&downvotes).Error; err != nil {
		return models.SuggestionVoteResponse{}, err
	}

	response := models.SuggestionVoteResponse{
		SuggestionID: suggestionID,
		Upvotes:      upvotes,
		Downvotes:    downvotes,
	}

	if userID > 0 {
		var userVote models.SuggestionVote
		if err := h.db.Where("suggestion_id = ? AND user_id = ?", suggestionID, userID).First(&userVote).Error; err == nil {
			response.UserVote = userVote.VoteType
		}
	}

	return response, nil
}

func (h *SuggestionsHandler) maskAnonymousSuggestion(suggestion *models.Suggestion) {
	if suggestion.IsAnonymous {
		suggestion.UserID = 0
		suggestion.User = models.User{}
	}
}

func normalizeSlug(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = strings.ReplaceAll(slug, "_", "-")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func (h *SuggestionsHandler) hasAwardedSuggestionVoteXP(userID uint, suggestionID uint) (bool, error) {
	metadata := fmt.Sprintf("{\"suggestion_id\":%d}", suggestionID)
	var count int64
	err := h.db.Model(&models.XPTransaction{}).
		Where("user_id = ? AND reason = ? AND metadata = ?", userID, "suggestion_vote", metadata).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func getSuggestionPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	return page, pageSize
}
