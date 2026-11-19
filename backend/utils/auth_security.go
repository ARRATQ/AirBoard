package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// maxTrackedIdentifiers déclenche un nettoyage des échecs périmés pour borner la mémoire.
const maxTrackedIdentifiers = 10000

// AuthSecurityManager gère la sécurité d'authentification
type AuthSecurityManager struct {
	failedAttempts  map[string]FailedAttempt // identifiant -> FailedAttempt
	mu              sync.RWMutex
	lockoutDuration time.Duration
	maxAttempts     int
	now             func() time.Time
}

// FailedAttempt représente une tentative d'authentification échouée
type FailedAttempt struct {
	Count       int
	FirstAt     time.Time
	LastAt      time.Time
	LockedUntil time.Time
}

// PasswordPolicy définit la politique de mots de passe
type PasswordPolicy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	DisallowCommon bool
	BannedWords    []string
}

// NewAuthSecurityManager crée un gestionnaire : 5 échecs, verrouillage de 30 minutes.
func NewAuthSecurityManager() *AuthSecurityManager {
	return NewAuthSecurityManagerWithLimits(5, 30*time.Minute)
}

// NewAuthSecurityManagerWithLimits crée un gestionnaire avec des seuils personnalisés.
func NewAuthSecurityManagerWithLimits(maxAttempts int, lockoutDuration time.Duration) *AuthSecurityManager {
	return &AuthSecurityManager{
		failedAttempts:  make(map[string]FailedAttempt),
		lockoutDuration: lockoutDuration,
		maxAttempts:     maxAttempts,
		now:             time.Now,
	}
}

// ValidatePassword valide un mot de passe selon la politique
func (asm *AuthSecurityManager) ValidatePassword(password string) error {
	policy := &PasswordPolicy{
		MinLength:      8,
		MaxLength:      128,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
		DisallowCommon: true,
		BannedWords: []string{
			"password", "admin", "airboard", "user", "123456",
			"qwerty", "letmein", "welcome", "monkey", "dragon",
		},
	}

	return asm.validatePasswordWithPolicy(password, policy)
}

// validatePasswordWithPolicy valide un mot de passe selon une politique donnée
func (asm *AuthSecurityManager) validatePasswordWithPolicy(password string, policy *PasswordPolicy) error {
	if len(password) < policy.MinLength {
		return fmt.Errorf("mot de passe trop court (minimum %d caractères)", policy.MinLength)
	}

	if len(password) > policy.MaxLength {
		return fmt.Errorf("mot de passe trop long (maximum %d caractères)", policy.MaxLength)
	}

	// Vérifications de complexité
	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if policy.RequireUpper && !hasUpper {
		return fmt.Errorf("mot de passe doit contenir au moins une majuscule")
	}

	if policy.RequireLower && !hasLower {
		return fmt.Errorf("mot de passe doit contenir au moins une minuscule")
	}

	if policy.RequireDigit && !hasDigit {
		return fmt.Errorf("mot de passe doit contenir au moins un chiffre")
	}

	if policy.RequireSpecial && !hasSpecial {
		return fmt.Errorf("mot de passe doit contenir au moins un caractère spécial")
	}

	// Vérifier les mots interdits
	if policy.DisallowCommon {
		lowerPassword := strings.ToLower(password)
		for _, word := range policy.BannedWords {
			if strings.Contains(lowerPassword, word) {
				return fmt.Errorf("mot de passe contient des mots interdits")
			}
		}
	}

	// Vérifier les patterns dangereux
	dangerousPatterns := []string{
		"(.)\\1{2,}",                            // Caractères répétés 3+ fois
		"(012|123|234|345|456|567|678|789|890)", // Séquences numériques
		"(abc|bcd|cde|def|efg|fgh|ghi|hij|ijk|jkl|klm|lmn|mno|nop|opq|pqr|qrs|rst|stu|tuv|uvw|vwx|wxy|xyz)", // Séquences alphabétiques
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(pattern, password)
		if matched {
			return fmt.Errorf("mot de passe contient des patterns prévisibles")
		}
	}

	return nil
}

// CheckFailedLogin indique si l'identifiant est verrouillé et pour combien de temps.
func (asm *AuthSecurityManager) CheckFailedLogin(identifier string) (bool, time.Duration) {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	attempt, exists := asm.failedAttempts[identifier]
	if !exists || attempt.LockedUntil.IsZero() {
		return false, 0
	}

	now := asm.now()
	if now.Before(attempt.LockedUntil) {
		return true, attempt.LockedUntil.Sub(now)
	}

	// Verrou expiré : on repart de zéro.
	delete(asm.failedAttempts, identifier)
	return false, 0
}

// RecordFailedLogin enregistre un échec. Les échecs sont comptés sur une fenêtre
// égale à la durée de verrouillage ; au-delà de maxAttempts, l'identifiant est verrouillé.
func (asm *AuthSecurityManager) RecordFailedLogin(identifier string) (bool, time.Duration) {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	now := asm.now()
	attempt, exists := asm.failedAttempts[identifier]

	lockExpired := !attempt.LockedUntil.IsZero() && !now.Before(attempt.LockedUntil)
	windowExpired := attempt.LockedUntil.IsZero() && now.Sub(attempt.FirstAt) > asm.lockoutDuration
	if !exists || lockExpired || windowExpired {
		attempt = FailedAttempt{FirstAt: now}
	}
	attempt.Count++
	attempt.LastAt = now

	if attempt.Count >= asm.maxAttempts {
		attempt.LockedUntil = now.Add(asm.lockoutDuration)
	}

	if !exists && len(asm.failedAttempts) >= maxTrackedIdentifiers {
		asm.pruneLocked(now)
	}
	asm.failedAttempts[identifier] = attempt

	if attempt.Count >= asm.maxAttempts {
		return true, asm.lockoutDuration
	}
	return false, 0
}

// pruneLocked retire les entrées qui ne verrouillent plus rien et dont la fenêtre est close.
func (asm *AuthSecurityManager) pruneLocked(now time.Time) {
	for id, attempt := range asm.failedAttempts {
		stillLocked := now.Before(attempt.LockedUntil)
		inWindow := now.Sub(attempt.FirstAt) <= asm.lockoutDuration
		if !stillLocked && !inWindow {
			delete(asm.failedAttempts, id)
		}
	}
}

// HasRecentFailures indique si l'identifiant a des échecs encore pris en compte
// (dans la fenêtre de comptage ou sous verrou).
func (asm *AuthSecurityManager) HasRecentFailures(identifier string) bool {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	attempt, exists := asm.failedAttempts[identifier]
	if !exists {
		return false
	}
	now := asm.now()
	return now.Before(attempt.LockedUntil) || now.Sub(attempt.FirstAt) <= asm.lockoutDuration
}

// RecordSuccessfulLogin enregistre une connexion réussie et nettoie les tentatives échouées
func (asm *AuthSecurityManager) RecordSuccessfulLogin(identifier string) {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	delete(asm.failedAttempts, identifier)
}

// GenerateSecureSecret génère un secret JWT sécurisé
func (asm *AuthSecurityManager) GenerateSecureSecret() (string, error) {
	bytes := make([]byte, 64) // 512 bits pour une sécurité renforcée
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("échec de génération du secret: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateJWTSecret valide la force d'un secret JWT
func (asm *AuthSecurityManager) ValidateJWTSecret(secret string) error {
	if len(secret) < 32 {
		return fmt.Errorf("secret JWT trop court (minimum 32 caractères)")
	}

	// Vérifier l'entropie
	entropy := asm.calculateEntropy(secret)
	if entropy < 3.0 { // Entropie minimale recommandée
		return fmt.Errorf("secret JWT a une entropie trop faible")
	}

	// Vérifier les patterns prévisibles
	commonPatterns := []string{
		"password", "secret", "key", "token", "jwt",
		"admin", "user", "airboard", "2024", "2025",
		"123456", "abcdef", "qwerty",
	}

	lowerSecret := strings.ToLower(secret)
	for _, pattern := range commonPatterns {
		if strings.Contains(lowerSecret, pattern) {
			return fmt.Errorf("secret JWT contient des patterns prévisibles")
		}
	}

	return nil
}

// calculateEntropy calcule l'entropie d'une chaîne
func (asm *AuthSecurityManager) calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}

	// Compter la fréquence des caractères
	freq := make(map[rune]int)
	for _, char := range s {
		freq[char]++
	}

	// Calculer l'entropie
	entropy := 0.0
	length := float64(len(s))
	for _, count := range freq {
		probability := float64(count) / length
		entropy -= probability * (probability)
	}

	return entropy
}

// HashPassword hache un mot de passe avec salt
func (asm *AuthSecurityManager) HashPassword(password string) (string, error) {
	// Générer un salt aléatoire
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("échec de génération du salt: %w", err)
	}

	// Hacher le mot de passe avec le salt
	hash := sha256.Sum256(append([]byte(password), salt...))

	// Retourner salt + hash encodés en base64
	result := base64.URLEncoding.EncodeToString(salt) + "." + base64.URLEncoding.EncodeToString(hash[:])
	return result, nil
}

// VerifyPassword vérifie un mot de passe contre son hash
func (asm *AuthSecurityManager) VerifyPassword(password, hashedPassword string) error {
	parts := strings.Split(hashedPassword, ".")
	if len(parts) != 2 {
		return fmt.Errorf("format de hash invalide")
	}

	salt, err := base64.URLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("salt invalide: %w", err)
	}

	storedHash, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("hash invalide: %w", err)
	}

	// Recalculer le hash
	computedHash := sha256.Sum256(append([]byte(password), salt...))

	// Comparer en temps constant
	if !asm.constantTimeCompare(computedHash[:], storedHash) {
		return fmt.Errorf("mot de passe incorrect")
	}

	return nil
}

// constantTimeCompare compare deux slices en temps constant
func (asm *AuthSecurityManager) constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// GenerateSecureToken génère un token cryptographiquement sûr
func (asm *AuthSecurityManager) GenerateSecureToken(length int) (string, error) {
	if length < 32 {
		return "", fmt.Errorf("longueur de token insuffisante (minimum 32)")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("échec de génération du token: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// CleanupFailedAttempts nettoie les tentatives échouées expirées
func (asm *AuthSecurityManager) CleanupFailedAttempts() {
	asm.mu.Lock()
	defer asm.mu.Unlock()

	now := time.Now()
	for identifier, attempt := range asm.failedAttempts {
		// Supprimer les tentatives de plus de 24h
		if now.Sub(attempt.FirstAt) > 24*time.Hour {
			delete(asm.failedAttempts, identifier)
		}
	}
}

// GetAuthStats retourne des statistiques d'authentification
func (asm *AuthSecurityManager) GetAuthStats() (total, locked, attempts int) {
	asm.mu.RLock()
	defer asm.mu.RUnlock()

	now := time.Now()
	for _, attempt := range asm.failedAttempts {
		total++
		if now.Before(attempt.LockedUntil) {
			locked++
		}
		attempts += attempt.Count
	}
	return
}
