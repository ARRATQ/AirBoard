package handlers

import (
	"net/http"
	"strconv"

	"airboard/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetNewsTypes - Liste tous les types d'articles (accessible à tous)
func (h *NewsHandler) GetNewsTypes(c *gin.Context) {
	var types []models.NewsType

	query := h.db.Model(&models.NewsType{})

	// Filtre actif/inactif
	if active := c.Query("active"); active != "" {
		query = query.Where("is_active = ?", active == "true")
	}

	if err := query.Order("\"order\" ASC, name ASC").Find(&types).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch types"})
		return
	}

	c.JSON(http.StatusOK, types)
}

// GetNewsType - Récupérer un type d'article par ID
func (h *NewsHandler) GetNewsType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var newsType models.NewsType
	if err := h.db.First(&newsType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Type not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch type"})
		}
		return
	}

	c.JSON(http.StatusOK, newsType)
}

// CreateNewsType - Créer un type d'article (admin uniquement)
func (h *NewsHandler) CreateNewsType(c *gin.Context) {
	var req models.NewsTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newsType := models.NewsType{
		Name:     req.Name,
		Icon:     req.Icon,
		Color:    req.Color,
		Order:    req.Order,
		IsActive: req.IsActive,
	}

	if err := h.db.Create(&newsType).Error; err != nil {
		if err.Error() == "UNIQUE constraint failed: news_types.name" || err.Error() == "UNIQUE constraint failed: news_types.slug" {
			c.JSON(http.StatusConflict, gin.H{"error": "A type with this name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create type"})
		}
		return
	}

	c.JSON(http.StatusCreated, newsType)
}

// UpdateNewsType - Modifier un type d'article (admin uniquement)
func (h *NewsHandler) UpdateNewsType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var newsType models.NewsType
	if err := h.db.First(&newsType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Type not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch type"})
		}
		return
	}

	var req models.NewsTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newsType.Name = req.Name
	newsType.Icon = req.Icon
	newsType.Color = req.Color
	newsType.Order = req.Order
	newsType.IsActive = req.IsActive

	if err := h.db.Save(&newsType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update type"})
		return
	}

	c.JSON(http.StatusOK, newsType)
}

// DeleteNewsType - Supprimer un type d'article (admin uniquement)
func (h *NewsHandler) DeleteNewsType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var newsType models.NewsType
	if err := h.db.First(&newsType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Type not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch type"})
		}
		return
	}

	// Vérifier si des news utilisent ce type
	var newsCount int64
	if err := h.db.Model(&models.News{}).Where("type = ?", newsType.Slug).Count(&newsCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check associated news"})
		return
	}
	if newsCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Cannot delete type with associated news",
			"news_count": newsCount,
		})
		return
	}

	if err := h.db.Delete(&newsType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Type deleted successfully"})
}
