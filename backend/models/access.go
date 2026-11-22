package models

import (
	"time"

	"gorm.io/gorm"
)

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
	// PermPollsManage : créer, modifier, clore et supprimer les sondages.
	PermPollsManage = "polls.manage"
	// PermCommentsModerate : voir les commentaires en attente, les approuver, les modifier
	// ou les supprimer. Sans notion de groupe : sur tout le site seulement.
	PermCommentsModerate = "comments.moderate"
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
	{Key: PermPollsManage, Module: "polls", Scopes: []string{ScopeAll, ScopeGroups}},
	{Key: PermCommentsModerate, Module: "comments", Scopes: []string{ScopeAll}},
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

// ContentPermissions : permissions qui font d'une personne une contributrice (elle rédige des
// actualités, des événements ou des sondages). Remplace l'ancien rôle « éditeur ».
var ContentPermissions = []string{PermNewsManage, PermEventsManage, PermPollsManage}

// UsersWithAnyPermission renvoie, parmi les comptes donnés (tous si aucun), ceux qu'un de leurs
// rôles dote d'au moins une des permissions.
func UsersWithAnyPermission(db *gorm.DB, permissions []string, userIDs ...uint) (map[uint]bool, error) {
	query := db.Table("user_access_roles").
		Select("DISTINCT user_access_roles.user_id").
		Joins("JOIN access_role_permissions ON access_role_permissions.role_id = user_access_roles.access_role_id").
		Where("access_role_permissions.permission IN ?", permissions)
	if len(userIDs) > 0 {
		query = query.Where("user_access_roles.user_id IN ?", userIDs)
	}
	var ids []uint
	if err := query.Pluck("user_access_roles.user_id", &ids).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// IsContributor : admin, ou compte qu'un rôle dote d'une permission de rédaction.
// (Les anciens admins de groupe sont vérifiés à part, par la table group_admins.)
func IsContributor(db *gorm.DB, user User) bool {
	if user.Role == "admin" {
		return true
	}
	ids, err := UsersWithAnyPermission(db, ContentPermissions, user.ID)
	return err == nil && ids[user.ID]
}

// EveryoneGroupID renvoie le groupe par défaut, auquel chaque compte appartient : celui des
// paramètres de l'application, sinon le groupe « Common » (même règle qu'à la création des
// comptes). 0 s'il n'y en a pas.
func EveryoneGroupID(db *gorm.DB) (uint, error) {
	var settings AppSettings
	err := db.Select("default_group_id").Order("id").Limit(1).Find(&settings).Error
	if err != nil {
		return 0, err
	}
	if settings.DefaultGroupID != nil && *settings.DefaultGroupID != 0 {
		return *settings.DefaultGroupID, nil
	}
	var ids []uint
	if err := db.Model(&Group{}).Where("LOWER(name) = ?", "common").Limit(1).Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	return ids[0], nil
}

// PermissionGroupIDs renvoie les groupes couverts par une permission de portée « ses
// groupes » : ceux dont le compte est membre, hors groupe par défaut. Ce groupe réunit tout
// le monde : le viser revient à publier pour tout le site, ce que seule la portée « tout le
// site » permet.
func PermissionGroupIDs(db *gorm.DB, userID uint) ([]uint, error) {
	everyone, err := EveryoneGroupID(db)
	if err != nil {
		return nil, err
	}
	ids := []uint{}
	query := db.Table("user_groups").Where("user_id = ?", userID)
	if everyone != 0 {
		query = query.Where("group_id <> ?", everyone)
	}
	if err := query.Pluck("group_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
