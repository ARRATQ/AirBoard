package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// FileSecurityResult résultat de la validation de sécurité d'un fichier
type FileSecurityResult struct {
	IsValid    bool
	SafeMIME   string
	SafeExt    string
	Reason     string
	ScanNeeded bool
}

// SecureFileValidator validateur sécurisé de fichiers
type SecureFileValidator struct {
	allowedExtensions map[string]string // extension -> MIME type
	allowedMIMETypes  map[string]bool
	maxFileSize       int64
	forbiddenPatterns []*regexp.Regexp
}

// NewSecureFileValidator crée un nouveau validateur sécurisé
func NewSecureFileValidator() *SecureFileValidator {
	return &SecureFileValidator{
		allowedExtensions: map[string]string{
			".jpg":  "image/jpeg",
			".jpeg": "image/jpeg",
			".png":  "image/png",
			".gif":  "image/gif",
			".webp": "image/webp",
			".pdf":  "application/pdf",
		},
		allowedMIMETypes: map[string]bool{
			"image/jpeg":      true,
			"image/png":       true,
			"image/gif":       true,
			"image/webp":      true,
			"application/pdf": true,
		},
		maxFileSize: 10 * 1024 * 1024, // 10MB
		forbiddenPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)\.(php|phtml|php3|php4|php5|pht|phar)$`),
			regexp.MustCompile(`(?i)\.(asp|aspx|jsp|js)$`),
			regexp.MustCompile(`(?i)\.(exe|bat|cmd|com|pif)$`),
			regexp.MustCompile(`(?i)\.(sh|ps1|py|pl|rb)$`),
			regexp.MustCompile(`(?i)\.(html|htm|svg)$`), // SVG contient du JavaScript potentiel
		},
	}
}

// ValidateSecureFile valide un fichier de manière sécurisée
func (v *SecureFileValidator) ValidateSecureFile(file multipart.File, fileHeader *multipart.FileHeader) *FileSecurityResult {
	// 1. Validation de base
	if fileHeader == nil {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  "Fichier invalide",
		}
	}

	// 2. Validation de taille
	if fileHeader.Size > v.maxFileSize {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Fichier trop volumineux (max: %d MB)", v.maxFileSize/(1024*1024)),
		}
	}

	// 3. Validation du nom de fichier
	cleanName := filepath.Base(fileHeader.Filename)
	if !v.isValidFilename(cleanName) {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  "Nom de fichier invalide ou dangereux",
		}
	}

	// 4. Validation de l'extension
	ext := strings.ToLower(filepath.Ext(cleanName))
	if !v.isAllowedExtension(ext) {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Extension %s non autorisée", ext),
		}
	}

	// 5. Validation du MIME type par lecture du contenu
	mimeType, err := v.detectMimeTypeFromContent(file)
	if err != nil {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Impossible de détecter le type MIME: %v", err),
		}
	}

	// 6. Vérification cohérence MIME/Extension
	if !v.isMimeTypeConsistent(mimeType, ext) {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  "Incohérence entre l'extension et le contenu du fichier",
		}
	}

	// 7. Validation du MIME type autorisé
	if !v.allowedMIMETypes[mimeType] {
		return &FileSecurityResult{
			IsValid: false,
			Reason:  fmt.Sprintf("Type MIME %s non autorisé", mimeType),
		}
	}

	// 8. Vérifications spéciales selon le type
	if mimeType == "image/svg+xml" {
		// SVG potentiellement dangereux, nécessite une sanitization
		file.Seek(0, 0)
		if !v.isSafeSVG(file) {
			return &FileSecurityResult{
				IsValid:    false,
				Reason:     "SVG contient du contenu potentiellement dangereux",
				ScanNeeded: true,
			}
		}
	}

	return &FileSecurityResult{
		IsValid:  true,
		SafeMIME: mimeType,
		SafeExt:  ext,
		Reason:   "Fichier valide",
	}
}

// isValidFilename vérifie si le nom de fichier est sûr
func (v *SecureFileValidator) isValidFilename(filename string) bool {
	// Vérifier les patterns dangereux
	for _, pattern := range v.forbiddenPatterns {
		if pattern.MatchString(filename) {
			return false
		}
	}

	// Vérifier les caractères interdits
	forbiddenChars := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range forbiddenChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	// Vérifier la longueur
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}

	return true
}

// isAllowedExtension vérifie si l'extension est autorisée
func (v *SecureFileValidator) isAllowedExtension(ext string) bool {
	_, exists := v.allowedExtensions[strings.ToLower(ext)]
	return exists
}

// detectMimeTypeFromContent détecte le MIME type en lisant le contenu
func (v *SecureFileValidator) detectMimeTypeFromContent(file multipart.File) (string, error) {
	// Lire les premiers 512 bytes pour la détection
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "", err
	}
	header = header[:n]

	// Détecter le MIME type basé sur le contenu
	mimeType := http.DetectContentType(header)

	// Réinitialiser le curseur
	file.Seek(0, 0)

	return mimeType, nil
}

// isMimeTypeConsistent vérifie la cohérence entre extension et MIME type
func (v *SecureFileValidator) isMimeTypeConsistent(mimeType string, ext string) bool {
	expectedMIME, exists := v.allowedExtensions[strings.ToLower(ext)]
	if !exists {
		return false
	}

	// Pour les images, accepter plusieurs variantes
	if strings.HasPrefix(mimeType, "image/") && strings.HasPrefix(expectedMIME, "image/") {
		return true
	}

	return mimeType == expectedMIME
}

// isSafeSVG vérifie si un SVG est sûr (pas de JavaScript, pas de liens externes)
func (v *SecureFileValidator) isSafeSVG(file multipart.File) bool {
	data := make([]byte, 4096) // Lire les premiers 4KB
	n, _ := file.Read(data)
	data = data[:n]

	content := strings.ToLower(string(data))

	// Patterns dangereux dans SVG
	dangerousPatterns := []string{
		"<script",
		"javascript:",
		"onclick=",
		"onload=",
		"onerror=",
		"<foreignobject",
		"<iframe",
		"xlink:href=",
		"style=",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(content, pattern) {
			return false
		}
	}

	return true
}

// SanitizeFilename assainit un nom de fichier
func (v *SecureFileValidator) SanitizeFilename(filename string) string {
	// Extraire le nom sans extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Remplacer les caractères non-alphanumériques
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	cleanName := re.ReplaceAllString(name, "-")

	// Limiter la longueur
	if len(cleanName) > 50 {
		cleanName = cleanName[:50]
	}

	// Ajouter un timestamp pour éviter les conflits
	timestamp := time.Now().Format("20060102150405")
	hash := v.generateShortHash(filename)

	return fmt.Sprintf("%s_%s_%s", cleanName, timestamp, hash)
}

// generateShortHash génère un hash court pour éviter les conflits
func (v *SecureFileValidator) generateShortHash(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])[:8]
}

// GetAllowedFileTypes retourne la liste des types de fichiers autorisés
func (v *SecureFileValidator) GetAllowedFileTypes() []string {
	types := make([]string, 0, len(v.allowedExtensions))
	for ext := range v.allowedExtensions {
		types = append(types, ext)
	}
	return types
}
