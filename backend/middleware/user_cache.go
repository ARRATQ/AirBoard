package middleware

import (
	"context"
	"errors"
	"sync"
	"time"
)

// userStatusTTL borne le délai pendant lequel un changement de rôle,
// une désactivation ou un changement de rôle ou de groupe peut être ignoré.
const userStatusTTL = time.Minute

// userStatusCacheMaxEntries évite une croissance illimitée du cache.
const userStatusCacheMaxEntries = 10000

// errUserUnavailable signale un compte supprimé ou introuvable.
var errUserUnavailable = errors.New("utilisateur introuvable")

// userStatus est l'état d'autorisation d'un utilisateur, relu en base.
type userStatus struct {
	Role     string
	IsActive bool
	// Permissions effectives {permission: portée} issues des rôles attribués
	Permissions map[string]string
	// PermissionGroups : pour chaque permission de portée « ses groupes », les groupes
	// couverts (voir models.PermissionGroups)
	PermissionGroups map[string][]uint
}

type userStatusLoader func(ctx context.Context, userID uint) (userStatus, error)

type cachedUserStatus struct {
	status    userStatus
	expiresAt time.Time
}

// userStatusCache garde en mémoire, pour une courte durée, le rôle, l'état
// actif, les permissions et les groupes couverts de chaque utilisateur authentifié.
type userStatusCache struct {
	mu      sync.Mutex
	entries map[uint]cachedUserStatus
	ttl     time.Duration
	load    userStatusLoader
	now     func() time.Time
}

func newUserStatusCache(ttl time.Duration, load userStatusLoader) *userStatusCache {
	return &userStatusCache{
		entries: make(map[uint]cachedUserStatus),
		ttl:     ttl,
		load:    load,
		now:     time.Now,
	}
}

// Get renvoie l'état de l'utilisateur, depuis le cache s'il est encore frais.
func (c *userStatusCache) Get(ctx context.Context, userID uint) (userStatus, error) {
	now := c.now()

	c.mu.Lock()
	entry, ok := c.entries[userID]
	c.mu.Unlock()
	if ok && now.Before(entry.expiresAt) {
		return entry.status, nil
	}

	status, err := c.load(ctx, userID)
	if err != nil {
		return userStatus{}, err
	}

	c.mu.Lock()
	if len(c.entries) >= userStatusCacheMaxEntries {
		c.evictExpiredLocked(now)
		if len(c.entries) >= userStatusCacheMaxEntries {
			c.entries = make(map[uint]cachedUserStatus)
		}
	}
	c.entries[userID] = cachedUserStatus{status: status, expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()

	return status, nil
}

func (c *userStatusCache) evictExpiredLocked(now time.Time) {
	for id, entry := range c.entries {
		if !now.Before(entry.expiresAt) {
			delete(c.entries, id)
		}
	}
}

// Clear vide le cache : un rôle modifié ou réattribué s'applique à la requête suivante.
func (c *userStatusCache) Clear() {
	c.mu.Lock()
	c.entries = make(map[uint]cachedUserStatus)
	c.mu.Unlock()
}
