package handlers

import (
	"net/http"
	"strings"

	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GamificationHandler struct {
	db                  *gorm.DB
	gamificationService *services.GamificationService
}

type GamificationRuleRequest struct {
	Reason  string `json:"reason" binding:"required,min=2,max=80"`
	Points  int64  `json:"points" binding:"required"`
	Enabled *bool  `json:"enabled"`
}

type AchievementRequest struct {
	Code          string `json:"code" binding:"required,min=2,max=100"`
	Name          string `json:"name" binding:"required,min=2,max=100"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	XPReward      int64  `json:"xp_reward"`
	Category      string `json:"category" binding:"required,oneof=user contributor"`
	IsSecret      bool   `json:"is_secret"`
	TriggerReason string `json:"trigger_reason"`
	Metric        string `json:"metric"`
	Threshold     int64  `json:"threshold"`
	IsActive      *bool  `json:"is_active"`
}

func NewGamificationHandler(db *gorm.DB, gs *services.GamificationService) *GamificationHandler {
	return &GamificationHandler{
		db:                  db,
		gamificationService: gs,
	}
}

// GetMyProfile récupère le profil de gamification de l'utilisateur connecté
func (h *GamificationHandler) GetMyProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var profile models.GamificationProfile
	if err := h.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			profile = models.GamificationProfile{UserID: userID, Level: 1, XP: 0}
			h.db.Create(&profile)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du profil"})
			return
		}
	}

	// Rank = nombre d'utilisateurs avec plus de XP + 1
	var higherCount int64
	h.db.Model(&models.GamificationProfile{}).
		Where("xp > ? AND deleted_at IS NULL", profile.XP).
		Count(&higherCount)

	// Nombre de badges débloqués
	var badgesCount int64
	h.db.Model(&models.UserAchievement{}).Where("user_id = ?", userID).Count(&badgesCount)

	c.JSON(http.StatusOK, gin.H{
		"id":                profile.ID,
		"user_id":           profile.UserID,
		"xp":                profile.XP,
		"level":             profile.Level,
		"login_streak":      profile.LoginStreak,
		"last_login_reward": profile.LastDailyLogin,
		"created_at":        profile.CreatedAt,
		"updated_at":        profile.UpdatedAt,
		"rank":              int(higherCount) + 1,
		"badges_count":      int(badgesCount),
	})
}

// GetMyAchievements récupère les badges débloqués par l'utilisateur
func (h *GamificationHandler) GetMyAchievements(c *gin.Context) {
	userID := c.GetUint("user_id")

	var userAchievements []models.UserAchievement
	if err := h.db.Preload("Achievement").Where("user_id = ?", userID).Find(&userAchievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des badges"})
		return
	}

	c.JSON(http.StatusOK, userAchievements)
}

// GetAllAchievements récupère tous les badges configurés
func (h *GamificationHandler) GetAllAchievements(c *gin.Context) {
	role := c.GetString("role")
	managedGroupIDs, _ := c.Get("managed_group_ids")

	// Un utilisateur est considéré comme "contributeur" s'il est admin, editor,
	// ou s'il administre au moins un groupe.
	isContributor := role == "admin" || role == "editor"
	if ids, ok := managedGroupIDs.([]uint); ok && len(ids) > 0 {
		isContributor = true
	}

	var achievements []models.Achievement
	query := h.db

	// Si l'utilisateur n'est pas un contributeur, on cache les badges de catégorie "contributor"
	if !isContributor {
		query = query.Where("category != ?", "contributor")
	}

	if err := query.Find(&achievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des badges"})
		return
	}

	c.JSON(http.StatusOK, achievements)
}

type LeaderboardBadge struct {
	Code  string `json:"code"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Name  string `json:"name"`
}

type LeaderboardEntry struct {
	UserID    uint               `json:"user_id"`
	Username  string             `json:"username"`
	XP        int64              `json:"xp"`
	Level     int                `json:"level"`
	FirstName string             `json:"first_name"`
	LastName  string             `json:"last_name"`
	Badges    []LeaderboardBadge `json:"badges"`
}

// GetLeaderboard récupère le classement XP avec les badges de chaque utilisateur
func (h *GamificationHandler) GetLeaderboard(c *gin.Context) {
	var rows []struct {
		UserID    uint   `json:"user_id"`
		Username  string `json:"username"`
		XP        int64  `json:"xp"`
		Level     int    `json:"level"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	h.db.Table("gamification_profiles").
		Select("gamification_profiles.user_id, users.username, users.first_name, users.last_name, gamification_profiles.xp, gamification_profiles.level").
		Joins("JOIN users ON users.id = gamification_profiles.user_id").
		Where("users.is_active = ? AND users.deleted_at IS NULL", true).
		Order("gamification_profiles.xp DESC").
		Limit(10).
		Scan(&rows)

	if len(rows) == 0 {
		c.JSON(http.StatusOK, []LeaderboardEntry{})
		return
	}

	userIDs := make([]uint, len(rows))
	for i, r := range rows {
		userIDs[i] = r.UserID
	}

	var userAchievements []struct {
		UserID      uint   `json:"user_id"`
		Code        string `json:"code"`
		Icon        string `json:"icon"`
		Color       string `json:"color"`
		Name        string `json:"name"`
		UnlockedAt  string `json:"unlocked_at"`
	}
	h.db.Table("user_achievements").
		Select("user_achievements.user_id, achievements.code, achievements.icon, achievements.color, achievements.name, user_achievements.unlocked_at").
		Joins("JOIN achievements ON achievements.id = user_achievements.achievement_id").
		Where("user_achievements.user_id IN ? AND achievements.deleted_at IS NULL", userIDs).
		Order("user_achievements.unlocked_at DESC").
		Scan(&userAchievements)

	badgesMap := make(map[uint][]LeaderboardBadge)
	for _, ua := range userAchievements {
		if len(badgesMap[ua.UserID]) < 4 {
			badgesMap[ua.UserID] = append(badgesMap[ua.UserID], LeaderboardBadge{
				Code:  ua.Code,
				Icon:  ua.Icon,
				Color: ua.Color,
				Name:  ua.Name,
			})
		}
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, r := range rows {
		badges := badgesMap[r.UserID]
		if badges == nil {
			badges = []LeaderboardBadge{}
		}
		entries[i] = LeaderboardEntry{
			UserID:    r.UserID,
			Username:  r.Username,
			XP:        r.XP,
			Level:     r.Level,
			FirstName: r.FirstName,
			LastName:  r.LastName,
			Badges:    badges,
		}
	}

	c.JSON(http.StatusOK, entries)
}

// GetMyTransactions récupère l'historique des gains de points de l'utilisateur
func (h *GamificationHandler) GetMyTransactions(c *gin.Context) {
	userID := c.GetUint("user_id")

	var transactions []models.XPTransaction
	if err := h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(20).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de l'historique"})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

func (h *GamificationHandler) GetRules(c *gin.Context) {
	var rules []models.GamificationRule
	if err := h.db.Order("reason ASC").Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des règles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rules": rules})
}

func (h *GamificationHandler) UpsertRule(c *gin.Context) {
	var req GamificationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	var rule models.GamificationRule
	err := h.db.Where("reason = ?", reason).First(&rule).Error
	if err == gorm.ErrRecordNotFound {
		rule = models.GamificationRule{Reason: reason, Points: req.Points, Enabled: enabled}
		if createErr := h.db.Create(&rule).Error; createErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de la règle"})
			return
		}
		c.JSON(http.StatusCreated, rule)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération de la règle"})
		return
	}

	rule.Points = req.Points
	rule.Enabled = enabled
	if err := h.db.Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la mise à jour de la règle"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *GamificationHandler) GetAdminAchievements(c *gin.Context) {
	var achievements []models.Achievement
	if err := h.db.Order("created_at DESC").Find(&achievements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération des badges"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"achievements": achievements})
}

func (h *GamificationHandler) CreateAchievement(c *gin.Context) {
	var req AchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	achievement := models.Achievement{
		Code:          strings.TrimSpace(req.Code),
		Name:          strings.TrimSpace(req.Name),
		Description:   strings.TrimSpace(req.Description),
		Icon:          strings.TrimSpace(req.Icon),
		Color:         strings.TrimSpace(req.Color),
		XPReward:      req.XPReward,
		Category:      req.Category,
		IsSecret:      req.IsSecret,
		TriggerReason: strings.TrimSpace(req.TriggerReason),
		Metric:        strings.TrimSpace(req.Metric),
		Threshold:     req.Threshold,
		IsActive:      isActive,
	}

	if achievement.Icon == "" {
		achievement.Icon = "mdi:medal"
	}
	if achievement.Color == "" {
		achievement.Color = "#FBBF24"
	}

	if err := h.db.Create(&achievement).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur lors de la création du badge (code peut déjà exister)"})
		return
	}

	c.JSON(http.StatusCreated, achievement)
}

func (h *GamificationHandler) UpdateAchievement(c *gin.Context) {
	id := c.Param("id")

	var achievement models.Achievement
	if err := h.db.First(&achievement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Badge introuvable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la récupération du badge"})
		return
	}

	var req AchievementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	achievement.Code = strings.TrimSpace(req.Code)
	achievement.Name = strings.TrimSpace(req.Name)
	achievement.Description = strings.TrimSpace(req.Description)
	achievement.Icon = strings.TrimSpace(req.Icon)
	achievement.Color = strings.TrimSpace(req.Color)
	achievement.XPReward = req.XPReward
	achievement.Category = req.Category
	achievement.IsSecret = req.IsSecret
	achievement.TriggerReason = strings.TrimSpace(req.TriggerReason)
	achievement.Metric = strings.TrimSpace(req.Metric)
	achievement.Threshold = req.Threshold
	if req.IsActive != nil {
		achievement.IsActive = *req.IsActive
	}

	if achievement.Icon == "" {
		achievement.Icon = "mdi:medal"
	}
	if achievement.Color == "" {
		achievement.Color = "#FBBF24"
	}

	if err := h.db.Save(&achievement).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur lors de la mise à jour du badge"})
		return
	}

	c.JSON(http.StatusOK, achievement)
}

func (h *GamificationHandler) DeleteAchievement(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Delete(&models.Achievement{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la suppression du badge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Badge supprimé"})
}
