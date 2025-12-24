package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	Upload(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, string, error)
	Delete(ctx context.Context, path string) error
	GetURL(path string) string
	GetType() string
}

// LocalStorage implements file storage on local disk
type LocalStorage struct {
	uploadDir string
	baseURL   string
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(uploadDir, baseURL string) (*LocalStorage, error) {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &LocalStorage{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}, nil
}

// Upload uploads a file to local storage
func (ls *LocalStorage) Upload(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader) (string, string, error) {
	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102-150405"), uuid.New().String()[:8], ext)

	// Create subdirectory based on date (for organization)
	dateDir := time.Now().Format("2006/01")
	targetDir := filepath.Join(ls.uploadDir, dateDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create date directory: %w", err)
	}

	// Full path on disk
	storagePath := filepath.Join(dateDir, filename)
	fullPath := filepath.Join(ls.uploadDir, storagePath)

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	// Generate public URL (use relative path for better compatibility)
	// Always use forward slashes for URLs, even on Windows
	urlPath := filepath.ToSlash(storagePath)
	url := fmt.Sprintf("/uploads/%s", urlPath)

	return storagePath, url, nil
}

// Delete deletes a file from local storage
func (ls *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(ls.uploadDir, path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetURL returns the public URL for a file
func (ls *LocalStorage) GetURL(path string) string {
	// Always use forward slashes for URLs, even on Windows
	urlPath := filepath.ToSlash(path)
	return fmt.Sprintf("/uploads/%s", urlPath)
}

// GetType returns the storage type
func (ls *LocalStorage) GetType() string {
	return "local"
}

// TODO: Implement S3Storage for production use
// type S3Storage struct {
// 	bucket    string
// 	region    string
// 	client    *s3.Client
// 	baseURL   string
// }
//
// func NewS3Storage(bucket, region, endpoint, accessKey, secretKey string) (*S3Storage, error) {
// 	// Implementation for S3/MinIO storage
// 	return &S3Storage{}, nil
// }
