package models

import "time"

// Portée d'une permission accordée par un rôle.
const (
	// ScopeAll : tout le site.
	ScopeAll = "all"
	// ScopeGroups : seulement les groupes dont la personne est membre.
	ScopeGroups = "groups"
)

// Permissions du catalogue. Une permission n'existe que si le serveur sait la vérifier :
// on ne l'ajoute ici qu'avec la route qui la contrôle.
const (
	// PermResourcesManage : créer, modifier et supprimer les ressources et leurs catégories.
	PermResourcesManage = "resources.manage"
	// PermNewsManage : rédiger, modifier et supprimer les actualités (et créer des tags).
	PermNewsManage = "news.manage"
	// PermEventsManage : créer, modifier et supprimer les événements.
	PermEventsManage = "events.manage"
)

// PermissionDef décrit une permission du catalogue et les portées qu'elle accepte.
type PermissionDef struct {
	Key    string   `json:"key"`
	Module string   `json:"module"`
	Scopes []string `json:"scopes"`
}

// PermissionCatalog est la liste fermée des permissions que l'administrateur peut confier.
var PermissionCatalog = []PermissionDef{
	{Key: PermResourcesManage, Module: "resources", Scopes: []string{ScopeAll, ScopeGroups}},
	{Key: PermNewsManage, Module: "news", Scopes: []string{ScopeAll, ScopeGroups}},
	{Key: PermEventsManage, Module: "events", Scopes: []string{ScopeAll, ScopeGroups}},
}

// LookupPermission renvoie la définition d'une permission du catalogue.
func LookupPermission(key string) (PermissionDef, bool) {
	for _, def := range PermissionCatalog {
		if def.Key == key {
			return def, true
		}
	}
	return PermissionDef{}, false
}

// AcceptsScope indique si la permission peut être accordée avec cette portée.
func (d PermissionDef) AcceptsScope(scope string) bool {
	for _, s := range d.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// AccessRole est un rôle créé par l'administrateur. Le rôle administrateur, lui, n'est pas
// une ligne de cette table : il est porté par users.role et a toujours tous les droits.
type AccessRole struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	Name        string                 `json:"name" gorm:"not null"`
	Description string                 `json:"description"`
	Permissions []AccessRolePermission `json:"permissions" gorm:"foreignKey:RoleID"`
	UserCount   int64                  `json:"user_count" gorm:"-"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AccessRolePermission est une permission accordée par un rôle, avec sa portée.
type AccessRolePermission struct {
	RoleID     uint   `json:"-" gorm:"primaryKey"`
	Permission string `json:"permission" gorm:"primaryKey"`
	Scope      string `json:"scope"`
}

// TableName : nom de table fixé, indépendant des règles de pluriel de GORM.
func (AccessRolePermission) TableName() string { return "access_role_permissions" }

// MergePermission ajoute une permission à un ensemble {permission: portée} :
// la portée « tout le site » l'emporte sur « ses groupes ».
func MergePermission(perms map[string]string, permission, scope string) {
	if perms[permission] == ScopeAll {
		return
	}
	perms[permission] = scope
}
