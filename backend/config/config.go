package config

import (
	"airboard/utils"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT      JWTConfig
	Server   ServerConfig
	SSO      SSOConfig
	Storage  StorageConfig
	Security SecurityConfig
}

type SecurityConfig struct {
	BcryptCost int // Coût de hashage bcrypt (recommandé: 12 ou plus)
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret                string
	TokenExpirationHours  int
	RefreshExpirationDays int
}

type ServerConfig struct {
	Port          string
	Mode          string // debug, release
	Origins       []string
	PublicURL     string // URL publique de l'application (ex: https://tools.marocpme.gov.ma)
	SignupEnabled bool   // Activer/désactiver l'inscription
}

type SSOConfig struct {
	Enabled       bool
	AutoProvision bool
	DefaultRole   string
	DefaultGroup  string
	GroupMapping  map[string]string // map[AuthentikGroup]AirboardGroup
	AdminGroups   []string          // Groupes Authentik qui ont le rôle admin
}

type StorageConfig struct {
	Type      string // local, s3, minio
	UploadDir string // For local storage
	BaseURL   string // Base URL for serving files
	// S3/MinIO config (for future use)
	S3Bucket    string
	S3Region    string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
}

func LoadConfig() *Config {
	// Charger le fichier .env si présent
	if err := godotenv.Load(); err != nil {
		log.Println("Aucun fichier .env trouvé, utilisation des variables d'environnement")
	}

	// Configuration base de données
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		dbPort = 5432
	}

	// Configuration JWT
	tokenExp, err := strconv.Atoi(getEnv("JWT_TOKEN_EXPIRATION_HOURS", "2"))
	if err != nil {
		tokenExp = 2
	}

	refreshExp, err := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRATION_DAYS", "1"))
	if err != nil {
		refreshExp = 1
	}

	// Configuration sécurité d'authentification
	authSecurity := utils.NewAuthSecurityManager()

	// Générer ou valider le secret JWT
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		// Générer un secret sécurisé par défaut
		generatedSecret, err := authSecurity.GenerateSecureSecret()
		if err != nil {
			log.Printf("⚠️ Erreur lors de la génération du secret JWT: %v", err)
			jwtSecret = "airboard-super-secret-key-2024-unsafe-default"
		} else {
			jwtSecret = generatedSecret
			log.Printf("✅ Secret JWT sécurisé généré automatiquement (512 bits)")
		}
	} else {
		// Valider le secret JWT fourni
		if err := authSecurity.ValidateJWTSecret(jwtSecret); err != nil {
			log.Printf("⚠️ Secret JWT faible détecté: %v", err)
			log.Printf("⚠️ Veuillez utiliser JWT_SECRET avec au moins 32 caractères et haute entropie")
			log.Printf("🔄 Génération d'un secret sécurisé pour remplacer le secret faible...")

			// Générer un nouveau secret sécurisé pour remplacer le faible
			secureSecret, genErr := authSecurity.GenerateSecureSecret()
			if genErr != nil {
				log.Printf("❌ Erreur lors de la génération du secret sécurisé: %v", genErr)
			} else {
				jwtSecret = secureSecret
				log.Printf("✅ Secret JWT vulnérable remplacé par un secret sécurisé")
			}
		}
	}

	// Configuration SSO
	ssoEnabled := getEnv("SSO_ENABLED", "false") == "true"
	ssoAutoProvision := getEnv("SSO_AUTO_PROVISION", "true") == "true"

	// Configuration Signup
	signupEnabled := getEnv("SIGNUP_ENABLED", "true") == "true"

	// Parse admin groups (comma-separated)
	adminGroups := []string{}
	if adminGroupsStr := getEnv("SSO_ADMIN_GROUPS", "airboard-admins"); adminGroupsStr != "" {
		for _, group := range splitAndTrim(adminGroupsStr, ",") {
			adminGroups = append(adminGroups, group)
		}
	}

	// Configuration Security - Bcrypt cost
	bcryptCost, err := strconv.Atoi(getEnv("BCRYPT_COST", "12"))
	if err != nil || bcryptCost < 10 || bcryptCost > 31 {
		bcryptCost = 12 // Valeur par défaut sécurisée (OWASP recommandation 2025)
		log.Printf("⚠️ BCRYPT_COST invalide, utilisation de la valeur par défaut: 12")
	}
	if bcryptCost < 12 {
		log.Printf("⚠️ BCRYPT_COST=%d est faible. Recommandation OWASP 2025: minimum 12", bcryptCost)
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "airboard"),
			Password: getEnv("DB_PASSWORD", "airboard123"),
			DBName:   getEnv("DB_NAME", "airboard"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:                jwtSecret,
			TokenExpirationHours:  tokenExp,
			RefreshExpirationDays: refreshExp,
		},
		Server: ServerConfig{
			Port:          getEnv("PORT", "8080"),
			Mode:          getEnv("GIN_MODE", "debug"),
			PublicURL:     getEnv("PUBLIC_URL", "http://localhost:80"),
			SignupEnabled: signupEnabled,
			Origins: []string{
				getEnv("FRONTEND_URL", "http://localhost:3000"),
				"http://localhost:3001", // Vite dev server (fallback)
				"http://localhost:5173", // Vite dev server
				"http://localhost:8080", // Swagger
			},
		},
		SSO: SSOConfig{
			Enabled:       ssoEnabled,
			AutoProvision: ssoAutoProvision,
			DefaultRole:   getEnv("SSO_DEFAULT_ROLE", "user"),
			DefaultGroup:  getEnv("SSO_DEFAULT_GROUP", "Common"),
			GroupMapping:  make(map[string]string), // Sera peuplé par les groupes Authentik
			AdminGroups:   adminGroups,
		},
		Storage: StorageConfig{
			Type:        getEnv("STORAGE_TYPE", "local"),
			UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
			BaseURL:     getEnv("PUBLIC_URL", "http://localhost:80"),
			S3Bucket:    getEnv("S3_BUCKET", ""),
			S3Region:    getEnv("S3_REGION", ""),
			S3Endpoint:  getEnv("S3_ENDPOINT", ""),
			S3AccessKey: getEnv("S3_ACCESS_KEY", ""),
			S3SecretKey: getEnv("S3_SECRET_KEY", ""),
		},
		Security: SecurityConfig{
			BcryptCost: bcryptCost,
		},
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range strings.Split(s, sep) {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
