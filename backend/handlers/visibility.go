package handlers

import (
	"log"
	"sort"
	"time"

	"airboard/middleware"
	"airboard/models"

	"gorm.io/gorm"
)

// Visibilité par groupes cibles.
//
// Une news, un événement ou un sondage sans groupe cible est global ; sinon il
// n'est visible que des membres (ou administrateurs) d'au moins un de ses
// groupes cibles. Ces règles étaient recopiées dans chaque handler ; elles sont
// réunies ici sous forme de scopes GORM.

// targetGroups décrit la table de liaison « groupes cibles » d'une entité.
type targetGroups struct {
	table  string // table de liaison, ex. news_target_groups
	fk     string // colonne vers l'entité, ex. news_id
	parent string // table de l'entité, ex. news
}

var (
	newsTargets  = targetGroups{table: "news_target_groups", fk: "news_id", parent: "news"}
	eventTargets = targetGroups{table: "event_target_groups", fk: "event_id", parent: "events"}
	pollTargets  = targetGroups{table: "poll_target_groups", fk: "poll_id", parent: "polls"}
)

// untargeted : l'entité n'a aucun groupe cible (contenu global).
func (t targetGroups) untargeted() string {
	return "NOT EXISTS (SELECT 1 FROM " + t.table + " WHERE " + t.table + "." + t.fk + " = " + t.parent + ".id)"
}

// targetsAnyOf : l'entité cible au moins un des groupes passés en paramètre (IN (?)).
func (t targetGroups) targetsAnyOf() string {
	return "EXISTS (SELECT 1 FROM " + t.table + " WHERE " + t.table + "." + t.fk + " = " + t.parent + ".id AND " +
		t.table + ".group_id IN (?))"
}

// visibleCondition : contenu global, ou ciblant un des groupes donnés. La
// condition est entre parenthèses pour rester correcte une fois combinée par
// AND avec d'autres filtres.
func (t targetGroups) visibleCondition(groupIDs []uint) (string, []interface{}) {
	if len(groupIDs) == 0 {
		return t.untargeted(), nil
	}
	return "(" + t.untargeted() + " OR " + t.targetsAnyOf() + ")", []interface{}{groupIDs}
}

// visibleTo filtre les entités visibles par un utilisateur membre (ou
// administrateur) des groupes donnés.
func (t targetGroups) visibleTo(groupIDs []uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		cond, args := t.visibleCondition(groupIDs)
		return db.Where(cond, args...)
	}
}

// publishedVisibleTo : publiées (date de publication atteinte) et visibles par
// les groupes donnés.
func (t targetGroups) publishedVisibleTo(groupIDs []uint, now time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		cond, args := t.visibleCondition(groupIDs)
		return db.Where("("+publishedCondition+" AND "+cond+")", append([]interface{}{true, now}, args...)...)
	}
}

// ownOrPublishedVisibleTo : les contenus de l'auteur (brouillons compris), plus
// les contenus publiés visibles par les groupes donnés.
func (t targetGroups) ownOrPublishedVisibleTo(authorID uint, groupIDs []uint, now time.Time) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		cond, args := t.visibleCondition(groupIDs)
		return db.Where("(author_id = ? OR ("+publishedCondition+" AND "+cond+"))",
			append([]interface{}{authorID, true, now}, args...)...)
	}
}

// manageableBy : ce qu'un administrateur de groupe gère dans son interface,
// soit ses propres contenus et ceux qui ciblent un de ses groupes.
func (t targetGroups) manageableBy(authorID uint, managedGroupIDs []uint) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(managedGroupIDs) == 0 {
			return db.Where("author_id = ?", authorID)
		}
		return db.Where("(author_id = ? OR "+t.targetsAnyOf()+")", authorID, managedGroupIDs)
	}
}

// publishedCondition : publié et date de publication atteinte (arguments : true, now).
const publishedCondition = "is_published = ? AND (published_at IS NULL OR published_at <= ?)"

// userGroupIDs renvoie les groupes dont l'utilisateur est membre.
func userGroupIDs(db *gorm.DB, userID uint) []uint {
	var ids []uint
	if err := db.Table("user_groups").Where("user_id = ?", userID).Pluck("group_id", &ids).Error; err != nil {
		log.Printf("[Visibility] groupes de l'utilisateur %d : %v", userID, err)
	}
	return ids
}

// adminGroupIDs renvoie les groupes administrés par l'utilisateur, lus en base
// (hors requête HTTP ; sinon middleware.GetManagedGroupIDs).
func adminGroupIDs(db *gorm.DB, userID uint) []uint {
	var ids []uint
	if err := db.Table("group_admins").Where("user_id = ?", userID).Pluck("group_id", &ids).Error; err != nil {
		log.Printf("[Visibility] groupes administrés par %d : %v", userID, err)
	}
	return ids
}

// attachManagedGroups renseigne les groupes administrés d'un utilisateur renvoyé
// au frontend : ManagedGroupIDs, et AdminOfGroups s'il n'a pas été préchargé.
func attachManagedGroups(db *gorm.DB, user *models.User) {
	user.ManagedGroupIDs = adminGroupIDs(db, user.ID)
	// Permissions des rôles attribués : l'interface affiche les menus correspondants
	// (le serveur les vérifie à chaque requête, voir middleware.RequirePermission).
	if perms, err := middleware.LoadPermissions(db, user.ID); err != nil {
		log.Printf("[Visibility] permissions de %d : %v", user.ID, err)
	} else if len(perms) > 0 {
		user.Permissions = perms
		if ids, err := models.PermissionGroupIDs(db, user.ID); err != nil {
			log.Printf("[Visibility] groupes couverts par les permissions de %d : %v", user.ID, err)
		} else {
			user.PermissionGroupIDs = ids
		}
	}
	if len(user.AdminOfGroups) > 0 || len(user.ManagedGroupIDs) == 0 {
		return
	}
	var groups []models.Group
	if err := db.Where("id IN ?", user.ManagedGroupIDs).Find(&groups).Error; err != nil {
		log.Printf("[Visibility] groupes administrés de %d : %v", user.ID, err)
	}
	user.AdminOfGroups = groups
}

// memberOrManagedGroupIDs : groupes dont l'utilisateur est membre, plus ceux
// qu'il administre.
func memberOrManagedGroupIDs(db *gorm.DB, userID uint, managedGroupIDs []uint) []uint {
	return mergeGroupIDs(userGroupIDs(db, userID), managedGroupIDs)
}

// mergeGroupIDs renvoie l'union triée et dédoublonnée des listes.
func mergeGroupIDs(lists ...[]uint) []uint {
	seen := make(map[uint]struct{})
	var out []uint
	for _, list := range lists {
		for _, id := range list {
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				out = append(out, id)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// groupMemberIDs renvoie les membres (distincts) des groupes donnés.
func groupMemberIDs(db *gorm.DB, groupIDs []uint) []uint {
	var ids []uint
	if len(groupIDs) == 0 {
		return ids
	}
	if err := db.Table("user_groups").Where("group_id IN ?", groupIDs).
		Distinct("user_id").Pluck("user_id", &ids).Error; err != nil {
		log.Printf("[Visibility] membres des groupes %v : %v", groupIDs, err)
	}
	return ids
}

// groupIDsOf renvoie les identifiants d'une liste de groupes.
func groupIDsOf(groups []models.Group) []uint {
	ids := make([]uint, 0, len(groups))
	for _, g := range groups {
		ids = append(ids, g.ID)
	}
	return ids
}

// canSeeTargeted : contenu global, ou ciblant un groupe dont l'utilisateur est
// membre ou administrateur (groupes administrés lus en base). Aucune requête
// pour un contenu global.
func canSeeTargeted(db *gorm.DB, userID uint, targets []models.Group) bool {
	if len(targets) == 0 {
		return true
	}
	return targetsAnyGroup(targets, mergeGroupIDs(userGroupIDs(db, userID), adminGroupIDs(db, userID)))
}

// targetsAnyGroup indique si un des groupes cibles figure dans groupIDs.
// Sans groupe cible, le contenu est global : true.
func targetsAnyGroup(targets []models.Group, groupIDs []uint) bool {
	if len(targets) == 0 {
		return true
	}
	for _, target := range targets {
		if containsUint(groupIDs, target.ID) {
			return true
		}
	}
	return false
}
