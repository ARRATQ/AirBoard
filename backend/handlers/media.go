package handlers

import (
	"airboard/apierror"
	"airboard/models"
	"airboard/services"
	"airboard/utils"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
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
		apierror.Respond(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse multipart form
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "No file provided or invalid file")
		return
	}
	defer file.Close()

	// Enhanced security validation
	validationResult := h.fileValidator.ValidateSecureFile(file, fileHeader)
	if !validationResult.IsValid {
		apierror.RespondCode(c, http.StatusBadRequest, apierror.CodeFileSecurity, validationResult.Reason)
		return
	}

	// Nom affiché et téléchargé : celui d'origine ; le stockage utilise le nom assaini
	displayName := utils.DisplayFilename(fileHeader.Filename)

	// Sanitize filename
	safeFilename := h.fileValidator.SanitizeFilename(fileHeader.Filename)
	if displayName == "" {
		displayName = safeFilename
	}

	// Update fileHeader with safe filename
	fileHeader.Filename = safeFilename

	// Upload file to storage
	storagePath, url, err := h.storageService.Upload(c.Request.Context(), file, fileHeader)
	if err != nil {
		log.Printf("[Media] Upload impossible : %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to upload file")
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
		Filename:    displayName,
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
		apierror.Respond(c, http.StatusInternalServerError, "Failed to save media record")
		return
	}

	h.loadUploader(&media)

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
	search := strings.TrimSpace(c.Query("search"))

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

	// Recherche sur le nom, sur toute la médiathèque (pas seulement la page affichée)
	if search != "" {
		query = query.Where("filename ILIKE ?", "%"+escapeLikePattern(search)+"%")
	}

	// Count total
	var total int64
	query.Count(&total)

	// Get paginated results
	var mediaList []models.Media
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&mediaList).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch media list")
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
		apierror.Respond(c, http.StatusBadRequest, "Invalid media ID")
		return
	}

	var media models.Media
	if err := h.db.Preload("Uploader").First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Media not found")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch media")
		return
	}

	c.JSON(http.StatusOK, media)
}

// UpdateMedia updates media metadata (admin only)
func (h *MediaHandler) UpdateMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid media ID")
		return
	}

	var media models.Media
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Media not found")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch media")
		return
	}

	var input struct {
		Filename string `json:"filename"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid input data")
		return
	}

	// Update only filename (metadata)
	if name := utils.DisplayFilename(input.Filename); name != "" {
		media.Filename = name
	}

	if err := h.db.Save(&media).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to update media")
		return
	}

	h.loadUploader(&media)

	c.JSON(http.StatusOK, gin.H{
		"message": "Media updated successfully",
		"media":   media,
	})
}

// DeleteMedia deletes a media file
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Invalid media ID")
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var media models.Media
	if err := h.db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			apierror.Respond(c, http.StatusNotFound, "Media not found")
			return
		}
		apierror.Respond(c, http.StatusInternalServerError, "Failed to fetch media")
		return
	}

	// Check permissions: only uploader or admin can delete
	if media.UploadedBy != userID.(uint) && role.(string) != "admin" {
		apierror.Respond(c, http.StatusForbidden, "You don't have permission to delete this media")
		return
	}

	// Delete from storage
	if err := h.storageService.Delete(c.Request.Context(), media.StoragePath); err != nil {
		// Log error but continue with database deletion
		fmt.Printf("Warning: Failed to delete file from storage: %v\n", err)
	}

	// Delete from database
	if err := h.db.Delete(&media).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Failed to delete media record")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Media deleted successfully",
	})
}

// validateFileSize function removed - validation now handled by SecureFileValidator

// loadUploader charge l'auteur de l'envoi, pour que la réponse s'affiche sans rechargement.
func (h *MediaHandler) loadUploader(media *models.Media) {
	if media.UploadedBy == 0 {
		return
	}
	if err := h.db.First(&media.Uploader, media.UploadedBy).Error; err != nil {
		log.Printf("[media] auteur %d du média %d : %v", media.UploadedBy, media.ID, err)
	}
}

// mediaStats résume toute la médiathèque, indépendamment des filtres et de la pagination.
type mediaStats struct {
	TotalFiles     int64 `json:"total_files"`
	TotalImages    int64 `json:"total_images"`
	TotalDocuments int64 `json:"total_documents"`
	TotalSize      int64 `json:"total_size"`
}

// GetMediaStats renvoie les totaux de la médiathèque (une seule requête d'agrégat).
func (h *MediaHandler) GetMediaStats(c *gin.Context) {
	var stats mediaStats
	err := h.db.Model(&models.Media{}).Select(`
		COUNT(*) AS total_files,
		COUNT(*) FILTER (WHERE mime_type LIKE 'image/%') AS total_images,
		COUNT(*) FILTER (WHERE mime_type NOT LIKE 'image/%' AND mime_type NOT LIKE 'video/%') AS total_documents,
		COALESCE(SUM(file_size), 0) AS total_size`).Scan(&stats).Error
	if err != nil {
		log.Printf("[media] statistiques : %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Failed to compute media stats")
		return
	}
	c.JSON(http.StatusOK, stats)
}
