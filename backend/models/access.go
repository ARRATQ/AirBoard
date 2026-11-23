package models

import (
	"sort"
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
	// PermGroupsManage : créer, modifier et supprimer des groupes d'utilisateurs et choisir
	// leurs membres. Portée « ses groupes » : des sous-groupes sous les groupes couverts, avec
	// des membres de ces groupes ; les groupes couverts eux-mêmes restent à l'admin. Portée
	// « tout le site » : tous les groupes. Dans les deux cas, le groupe par défaut et les
	// groupes liés au trombinoscope restent réservés à l'admin.
	PermGroupsManage = "groups.manage"
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
	{Key: PermGroupsManage, Module: "groups", Scopes: []string{ScopeAll, ScopeGroups}},
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
	// GroupIDs : groupes choisis pour ce rôle quand il est lu pour un compte (vide : ses groupes)
	GroupIDs  []uint    `json:"group_ids,omitempty" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

// PermissionGroupIDs renvoie les groupes dont le compte est membre, hors groupe par défaut :
// ceux que couvre la portée « ses groupes » d'un rôle attribué sans groupes choisis. Le groupe
// par défaut réunit tout le monde : le viser revient à publier pour tout le site, ce que seule
// la portée « tout le site » permet.
func PermissionGroupIDs(db *gorm.DB, userID uint) ([]uint, error) {
	everyone, err := EveryoneGroupID(db)
	if err != nil {
		return nil, err
	}
	members, err := memberGroupsExcept(db, userID, everyone)
	if err != nil {
		return nil, err
	}
	return withSubgroupsExcept(db, members, everyone)
}

// withSubgroupsExcept : les groupes et leurs sous-groupes (un rôle limité à un groupe couvre
// aussi ses sous-groupes), sans le groupe par défaut.
func withSubgroupsExcept(db *gorm.DB, ids []uint, everyone uint) ([]uint, error) {
	all, err := GroupsWithDescendants(db, ids)
	if err != nil {
		return nil, err
	}
	out := make([]uint, 0, len(all))
	for _, id := range all {
		if id != everyone {
			out = append(out, id)
		}
	}
	return out, nil
}

func memberGroupsExcept(db *gorm.DB, userID, everyone uint) ([]uint, error) {
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

// UserAccessRoleGroup : groupe choisi pour un rôle attribué à un compte. Un rôle attribué
// sans groupe choisi couvre les groupes dont le compte est membre (PermissionGroupIDs).
type UserAccessRoleGroup struct {
	UserID       uint `gorm:"primaryKey"`
	AccessRoleID uint `gorm:"primaryKey"`
	GroupID      uint `gorm:"primaryKey"`
}

// TableName : nom de table fixé, indépendant des règles de pluriel de GORM.
func (UserAccessRoleGroup) TableName() string { return "user_access_role_groups" }

// AssignmentGroups renvoie, pour chaque rôle attribué au compte, les groupes choisis
// (groupes supprimés et groupe par défaut exclus) ; un rôle absent n'a pas de groupe choisi.
func AssignmentGroups(db *gorm.DB, userID uint) (map[uint][]uint, error) {
	everyone, err := EveryoneGroupID(db)
	if err != nil {
		return nil, err
	}
	return assignmentGroupsExcept(db, userID, everyone)
}

func assignmentGroupsExcept(db *gorm.DB, userID, everyone uint) (map[uint][]uint, error) {
	var rows []UserAccessRoleGroup
	query := db.Table("user_access_role_groups uarg").
		Select("uarg.user_id, uarg.access_role_id, uarg.group_id").
		Joins("JOIN groups g ON g.id = uarg.group_id AND g.deleted_at IS NULL").
		Where("uarg.user_id = ?", userID)
	if everyone != 0 {
		query = query.Where("uarg.group_id <> ?", everyone)
	}
	if err := query.Order("uarg.group_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[uint][]uint)
	for _, r := range rows {
		out[r.AccessRoleID] = append(out[r.AccessRoleID], r.GroupID)
	}
	return out, nil
}

// PermissionGroups renvoie, pour chaque permission que le compte exerce avec la portée « ses
// groupes », les groupes couverts : l'union, sur les rôles qui la donnent, des groupes choisis
// à l'attribution, ou des groupes dont il est membre si aucun n'a été choisi, sous-groupes
// compris.
func PermissionGroups(db *gorm.DB, userID uint) (map[string][]uint, error) {
	roots, everyone, err := permissionRootSets(db, userID)
	if err != nil {
		return nil, err
	}
	out := map[string][]uint{}
	for perm, ids := range roots {
		// Sous-groupes compris (résultat trié)
		expanded, err := withSubgroupsExcept(db, ids, everyone)
		if err != nil {
			return nil, err
		}
		out[perm] = expanded
	}
	return out, nil
}

// PermissionRootGroups renvoie les groupes que la permission couvre directement (portée « ses
// groupes »), sans leurs sous-groupes : groupes choisis à l'attribution, ou groupes dont le
// compte est membre. Triés ; vide si le compte n'exerce pas la permission avec cette portée.
func PermissionRootGroups(db *gorm.DB, userID uint, permission string) ([]uint, error) {
	roots, _, err := permissionRootSets(db, userID)
	if err != nil {
		return nil, err
	}
	if roots[permission] == nil {
		return []uint{}, nil
	}
	return roots[permission], nil
}

// permissionRootSets : {permission: groupes couverts directement (triés)} pour la portée
// « ses groupes », et l'identifiant du groupe par défaut.
func permissionRootSets(db *gorm.DB, userID uint) (map[string][]uint, uint, error) {
	var rows []struct {
		Permission   string
		AccessRoleID uint
	}
	if err := db.Table("access_role_permissions arp").
		Select("arp.permission, uar.access_role_id").
		Joins("JOIN user_access_roles uar ON uar.access_role_id = arp.role_id").
		Where("uar.user_id = ? AND arp.scope = ?", userID, ScopeGroups).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	out := map[string][]uint{}
	if len(rows) == 0 {
		return out, 0, nil
	}
	everyone, err := EveryoneGroupID(db)
	if err != nil {
		return nil, 0, err
	}
	chosen, err := assignmentGroupsExcept(db, userID, everyone)
	if err != nil {
		return nil, 0, err
	}
	var members []uint
	membersLoaded := false
	sets := map[string]map[uint]bool{}
	for _, r := range rows {
		groups, ok := chosen[r.AccessRoleID]
		if !ok {
			if !membersLoaded {
				if members, err = memberGroupsExcept(db, userID, everyone); err != nil {
					return nil, 0, err
				}
				membersLoaded = true
			}
			groups = members
		}
		if sets[r.Permission] == nil {
			sets[r.Permission] = map[uint]bool{}
		}
		for _, id := range groups {
			sets[r.Permission][id] = true
		}
	}
	for perm, set := range sets {
		ids := make([]uint, 0, len(set))
		for id := range set {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		out[perm] = ids
	}
	return out, everyone, nil
}
