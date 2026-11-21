package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"airboard/apierror"
	"airboard/models"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch types")
		return
	}

	c.JSON(http.StatusOK, types)
}

// GetNewsType - Récupérer un type d'article par ID
func (h *NewsHandler) GetNewsType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var newsType models.NewsType
	if err := h.db.First(&newsType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Type not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch type")
		}
		return
	}

	c.JSON(http.StatusOK, newsType)
}

// CreateNewsType - Créer un type d'article (admin uniquement)
func (h *NewsHandler) CreateNewsType(c *gin.Context) {
	var req models.NewsTypeRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	newsType := models.NewsType{
		Name:     req.Name,
		Icon:     req.Icon,
		Color:    req.Color,
		Order:    req.Order,
		IsActive: req.IsActive,
	}

	if err := createKeepingFalse(h.db, c, &newsType, "is_active"); err != nil {
		if isUniqueViolation(err) {
			apierror.Respond(c, http.StatusConflict, "A type with this name already exists")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to create type")
		}
		return
	}

	c.JSON(http.StatusCreated, newsType)
}

// UpdateNewsType - Modifier un type d'article (admin uniquement)
func (h *NewsHandler) UpdateNewsType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var newsType models.NewsType
	if err := h.db.First(&newsType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Type not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch type")
		}
		return
	}

	var req models.NewsTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}

	newsType.Name = req.Name
	newsType.Icon = req.Icon
	newsType.Color = req.Color
	newsType.Order = req.Order
	newsType.IsActive = req.IsActive

	if err := h.db.Save(&newsType).Error; err != nil {
		if isUniqueViolation(err) {
			apierror.Respond(c, http.StatusConflict, "A type with this name already exists")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Failed to update type")
		return
	}

	c.JSON(http.StatusOK, newsType)
}

// DeleteNewsType - Supprimer un type d'article (admin uniquement)
func (h *NewsHandler) DeleteNewsType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var newsType models.NewsType
	if err := h.db.First(&newsType, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Type not found")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch type")
		}
		return
	}

	// Vérifier si des news utilisent ce type
	var newsCount int64
	if err := h.db.Model(&models.News{}).Where("type = ?", newsType.Slug).Count(&newsCount).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to check associated news")
		return
	}
	if newsCount > 0 {
		apierror.Respond(c, http.StatusBadRequest,
			fmt.Sprintf("Cannot delete type with associated news (%d)", newsCount))
		return
	}

	if err := h.db.Delete(&newsType).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to delete type")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Type deleted successfully"})
}
