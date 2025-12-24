package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuthStateManager gère les états OAuth pour la protection CSRF
type OAuthStateManager struct {
	states map[string]OAuthState // state -> OAuthState
	mu     sync.RWMutex
}

// OAuthState représente un état OAuth avec son expiration
type OAuthState struct {
	Provider  string    `json:"provider"`
	ClientID  string    `json:"client_id"`
	Nonce     string    `json:"nonce"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
}

// NewOAuthStateManager crée un nouveau gestionnaire d'états OAuth
func NewOAuthStateManager() *OAuthStateManager {
	return &OAuthStateManager{
		states: make(map[string]OAuthState),
	}
}

// GenerateState génère un nouvel état OAuth sécurisé
func (osm *OAuthStateManager) GenerateState(provider, clientID string) (string, string, error) {
	// Générer un state aléatoire
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random state: %w", err)
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Générer un nonce pour une protection supplémentaire
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := base64.URLEncoding.EncodeToString(nonceBytes)

	// Créer l'état avec expiration (10 minutes)
	now := time.Now()
	oauthState := OAuthState{
		Provider:  provider,
		ClientID:  clientID,
		Nonce:     nonce,
		CreatedAt: now,
		ExpiresAt: now.Add(10 * time.Minute),
		Used:      false,
	}

	// Stocker l'état
	osm.mu.Lock()
	osm.states[state] = oauthState
	osm.mu.Unlock()

	return state, nonce, nil
}

// ValidateState valide un état OAuth reçu
func (osm *OAuthStateManager) ValidateState(state, expectedProvider, expectedClientID string) (string, error) {
	// Validation de base
	if state == "" {
		return "", fmt.Errorf("missing OAuth state")
	}

	osm.mu.RLock()
	oauthState, exists := osm.states[state]
	osm.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("invalid or expired state")
	}

	// Vérifier l'expiration
	if time.Now().After(oauthState.ExpiresAt) {
		osm.mu.Lock()
		delete(osm.states, state)
		osm.mu.Unlock()
		return "", fmt.Errorf("state expired")
	}

	// Vérifier si déjà utilisé (protection contre replay)
	if oauthState.Used {
		return "", fmt.Errorf("state already used")
	}

	// Valider le provider
	if oauthState.Provider != expectedProvider {
		return "", fmt.Errorf("state provider mismatch")
	}

	// Valider le client ID
	if oauthState.ClientID != expectedClientID {
		return "", fmt.Errorf("state client ID mismatch")
	}

	// Marquer comme utilisé
	osm.mu.Lock()
	if existingState, ok := osm.states[state]; ok {
		existingState.Used = true
		osm.states[state] = existingState
	}
	osm.mu.Unlock()

	// Retourner le nonce pour validation supplémentaire
	return oauthState.Nonce, nil
}

// ValidateNonce valide un nonce OAuth
func (osm *OAuthStateManager) ValidateNonce(state, receivedNonce string) error {
	if receivedNonce == "" {
		return fmt.Errorf("missing nonce")
	}

	osm.mu.RLock()
	oauthState, exists := osm.states[state]
	osm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("invalid state for nonce validation")
	}

	if oauthState.Nonce != receivedNonce {
		return fmt.Errorf("invalid nonce")
	}

	return nil
}

// CleanupExpiredStates nettoie les états expirés
func (osm *OAuthStateManager) CleanupExpiredStates() {
	now := time.Now()

	osm.mu.Lock()
	defer osm.mu.Unlock()

	for state, oauthState := range osm.states {
		if now.After(oauthState.ExpiresAt) || oauthState.Used {
			delete(osm.states, state)
		}
	}
}

// GetStats retourne des statistiques sur les états gérés
func (osm *OAuthStateManager) GetStats() (total, active, expired, used int) {
	osm.mu.RLock()
	defer osm.mu.RUnlock()

	now := time.Now()
	for _, oauthState := range osm.states {
		total++
		if oauthState.Used {
			used++
		} else if now.After(oauthState.ExpiresAt) {
			expired++
		} else {
			active++
		}
	}
	return
}

// SecureOAuthURL construit une URL OAuth sécurisée avec state et nonce
func SecureOAuthURL(baseURL string, params map[string]string, state, nonce string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid OAuth URL: %w", err)
	}

	// Ajouter tous les paramètres
	q := u.Query()
	for key, value := range params {
		q.Set(key, value)
	}

	// Ajouter state et nonce
	q.Set("state", state)
	q.Set("nonce", nonce)

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ValidateOAuthCallbackURL valide une URL de callback OAuth
func ValidateOAuthCallbackURL(callbackURL, expectedDomain string) error {
	u, err := url.Parse(callbackURL)
	if err != nil {
		return fmt.Errorf("invalid callback URL format: %w", err)
	}

	// Vérifier le scheme
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("invalid callback URL scheme")
	}

	// Vérifier le domain attendu
	if !strings.Contains(u.Host, expectedDomain) {
		return fmt.Errorf("callback URL domain mismatch")
	}

	// Vérifier qu'il n'y a pas de paramètres malveillants
	for key := range u.Query() {
		if strings.Contains(strings.ToLower(key), "javascript") {
			return fmt.Errorf("malicious parameter in callback URL")
		}
	}

	return nil
}
