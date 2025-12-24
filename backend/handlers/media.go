package handlers

import (
	"airboard/models"
	"airboard/services"
	"airboard/utils"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MediaHandler struct {
	db             *gorm.DB
	storageService services.StorageService
	fileValidator  *utils.SecureFileValidator
}

func NewMediaHandler(db *gorm.DB, storageService services.StorageService) *MediaHandler {
	return &MediaHandler{
		db:             db,
		storageService: storageService,
		fileValidator:  utils.NewSecureFileValidator(),
	}
}

// Constants and variables replaced by SecureFileValidator
// File size limits, MIME types, and validation logic now handled by utils.SecureFileValidator

// UploadMedia handles file upload with enhanced security
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// Parse multipart form
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_file",
			Message: "No file provided or invalid file",
			Code:    http.StatusBadRequest,
		})
		return
	}
	defer file.Close()

	// Enhanced security validation
	validationResult := h.fileValidator.ValidateSecureFile(file, fileHeader)
	if !validationResult.IsValid {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "file_security_violation",
			Message: validationResult.Reason,
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Sanitize filename
	safeFilename := h.fileValidator.SanitizeFilename(fileHeader.Filename)

	// Update fileHeader with safe filename
	fileHeader.Filename = safeFilename

	// Upload file to storage
	storagePath, url, err := h.storageService.Upload(c.Request.Context(), file, fileHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "upload_failed",
			Message: fmt.Sprintf("Failed to upload file: %v", err),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Get image dimensions if it's an image (only for safe image types)
	var width, height *int
	if strings.HasPrefix(validationResult.SafeMIME, "image/") &&
		validationResult.SafeMIME != "image/svg+xml" &&
		validationResult.SafeMIME != "image/gif" {
		// Reset file reader
		file.Seek(0, 0)
		if img, _, err := image.DecodeConfig(file); err == nil {
			w, h := img.Width, img.Height
			width, height = &w, &h
		}
	}

	// Create media record in database
	media := models.Media{
		Filename:    safeFilename,
		StoragePath: storagePath,
		URL:         url,
		MimeType:    validationResult.SafeMIME,
		FileSize:    fileHeader.Size,
		Width:       width,
		Height:      height,
		StorageType: h.storageService.GetType(),
		UploadedBy:  userID.(uint),
	}

	if err := h.db.Create(&media).Error; err != nil {
		// Try to delete uploaded file on database error
		h.storageService.Delete(c.Request.Context(), storagePath)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to save media record",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"media":   media,
	})
}

// GetMediaList returns paginated list of uploaded media
func (h *MediaHandler) GetMediaList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	mediaType := c.Query("type") // filter by type: image, video, document, pdf, etc.

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := h.db.Model(&models.Media{}).Preload("Uploader")

	// Filter by media type if specified
	if mediaType != "" {
		switch mediaType {
		case "image":
			query = query.Where("mime_type LIKE ?", "image/%")
		case "video":
			query = query.Where("mime_type LIKE ?", "video/%")
		case "pdf":
			query = query.Where("mime_type = ?", "application/pdf")
		case "document":
			query = query.Where("mime_type LIKE ? OR mime_type LIKE ?", "%word%", "%document%")
		case "spreadsheet":
			query = query.Where("mime_type LIKE ? OR mime_type LIKE ?", "%sheet%", "%excel%")
		case "presentation":
			query = query.Where("mime_type LIKE ? OR mime_type LIKE ?", "%presentation%", "%powerpoint%")
		}
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get paginated results
	var mediaList []models.Media
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&mediaList).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch media list",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       mediaList,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetMedia returns a single media item by ID
func (h *MediaHandler) GetMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid media ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var media models.Media
	if err := h.db.Preload("Uploader").First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Media not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch media",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, media)
}

// UpdateMedia updates media metadata (admin only)
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid media ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	var media models.Media
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Media not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch media",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	var input struct {
		Filename string `json:"filename"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_input",
			Message: "Invalid input data",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// Update only filename (metadata)
	if input.Filename != "" {
		media.Filename = input.Filename
	}

	if err := h.db.Save(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update media",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Media updated successfully",
		"media":   media,
	})
}

// DeleteMedia deletes a media file
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid media ID",
			Code:    http.StatusBadRequest,
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var media models.Media
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "not_found",
				Message: "Media not found",
				Code:    http.StatusNotFound,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to fetch media",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// Check permissions: only uploader or admin can delete
	if media.UploadedBy != userID.(uint) && role.(string) != "admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "forbidden",
			Message: "You don't have permission to delete this media",
			Code:    http.StatusForbidden,
		})
		return
	}

	// Delete from storage
	if err := h.storageService.Delete(c.Request.Context(), media.StoragePath); err != nil {
		// Log error but continue with database deletion
		fmt.Printf("Warning: Failed to delete file from storage: %v\n", err)
	}

	// Delete from database
	if err := h.db.Delete(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database_error",
			Message: "Failed to delete media record",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Media deleted successfully",
	})
}

// validateFileSize function removed - validation now handled by SecureFileValidator
