package middleware

import (
	"context"
	"errors"
	"testing"
	"time"

	"airboard/config"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-with-enough-length-0123456789"

func newTestAuthMiddleware() *AuthMiddleware {
	cfg := &config.Config{}
	cfg.JWT.Secret = testSecret
	return &AuthMiddleware{config: cfg}
}

func signTestToken(t *testing.T, secret string, claims jwt.MapClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signature du jeton de test: %v", err)
	}
	return token
}

// baseClaims reproduit les claims émis par GenerateToken / GenerateRefreshToken.
func baseClaims(tokenType string) jwt.MapClaims {
	claims := jwt.MapClaims{
		"user_id":  float64(42),
		"username": "alice",
		"role":     "admin",
		"email":    "alice@example.com",
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	if tokenType != "" {
		claims["type"] = tokenType
	}
	return claims
}

func TestVerifyToken(t *testing.T) {
	am := newTestAuthMiddleware()

	expired := baseClaims(tokenTypeAccess)
	expired["exp"] = time.Now().Add(-time.Minute).Unix()

	noExp := baseClaims(tokenTypeAccess)
	delete(noExp, "exp")

	noUserID := baseClaims(tokenTypeAccess)
	delete(noUserID, "user_id")

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"access token is accepted", signTestToken(t, testSecret, baseClaims(tokenTypeAccess)), false},
		{"legacy access token without type is accepted", signTestToken(t, testSecret, baseClaims("")), false},
		{"refresh token is rejected", signTestToken(t, testSecret, baseClaims(tokenTypeRefresh)), true},
		{"unknown type is rejected", signTestToken(t, testSecret, baseClaims("magic")), true},
		{"expired token is rejected", signTestToken(t, testSecret, expired), true},
		{"token without exp is rejected", signTestToken(t, testSecret, noExp), true},
		{"token without user_id is rejected", signTestToken(t, testSecret, noUserID), true},
		{"wrong signature is rejected", signTestToken(t, "another-secret-of-sufficient-length-xx", baseClaims(tokenTypeAccess)), true},
		{"garbage is rejected", "not-a-jwt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := am.verifyToken(tt.token)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("verifyToken() accepted the token, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("verifyToken() error = %v", err)
			}
			if claims.UserID != 42 || claims.Username != "alice" {
				t.Errorf("verifyToken() claims = %+v", claims)
			}
		})
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	am := newTestAuthMiddleware()

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{"refresh token is accepted", signTestToken(t, testSecret, baseClaims(tokenTypeRefresh)), false},
		{"access token is rejected", signTestToken(t, testSecret, baseClaims(tokenTypeAccess)), true},
		{"token without type is rejected", signTestToken(t, testSecret, baseClaims("")), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := am.VerifyRefreshToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Fatalf("VerifyRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStatusCache(t *testing.T) {
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	calls := 0
	current := userStatus{Role: "admin", IsActive: true, MemberGroupIDs: []uint{3}}

	cache := newUserStatusCache(time.Minute, func(_ context.Context, userID uint) (userStatus, error) {
		calls++
		if userID == 99 {
			return userStatus{}, errUserUnavailable
		}
		return current, nil
	})
	cache.now = func() time.Time { return now }

	got, err := cache.Get(context.Background(), 1)
	if err != nil || got.Role != "admin" || calls != 1 {
		t.Fatalf("first Get = %+v, %v (calls %d)", got, err, calls)
	}

	// Rétrogradation en base : ignorée tant que l'entrée est fraîche.
	current = userStatus{Role: "user", IsActive: false}
	now = now.Add(30 * time.Second)
	if got, _ := cache.Get(context.Background(), 1); got.Role != "admin" || calls != 1 {
		t.Fatalf("Get within TTL = %+v (calls %d), want cached admin", got, calls)
	}

	// Après le TTL, le nouvel état est relu.
	now = now.Add(31 * time.Second)
	got, _ = cache.Get(context.Background(), 1)
	if got.Role != "user" || got.IsActive || calls != 2 {
		t.Fatalf("Get after TTL = %+v (calls %d), want refreshed inactive user", got, calls)
	}

	// Un utilisateur introuvable n'est pas mis en cache.
	if _, err := cache.Get(context.Background(), 99); !errors.Is(err, errUserUnavailable) {
		t.Fatalf("Get(99) error = %v, want errUserUnavailable", err)
	}
	if _, ok := cache.entries[99]; ok {
		t.Error("an unavailable user must not be cached")
	}
}
