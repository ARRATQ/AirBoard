package models

import (
	"sort"

	"gorm.io/gorm"
)

// Hiérarchie des groupes (groups.parent_id, migration 00010).
//   - Visibilité : un membre d'un sous-groupe compte aussi comme membre des groupes au-dessus
//     (GroupsWithAncestors, vue user_effective_groups).
//   - Portée « ses groupes » d'un rôle et ciblage : un groupe couvre aussi ses sous-groupes
//     (GroupsWithDescendants).
// Les groupes supprimés sont ignorés ; UNION arrête la récursion même en cas de boucle.

const descendantsSQL = `
WITH RECURSIVE down (id) AS (
    SELECT id FROM groups WHERE id IN ? AND deleted_at IS NULL
    UNION
    SELECT g.id FROM groups g JOIN down ON g.parent_id = down.id WHERE g.deleted_at IS NULL
)
SELECT id FROM down`

const ancestorsSQL = `
WITH RECURSIVE up (id) AS (
    SELECT id FROM groups WHERE id IN ? AND deleted_at IS NULL
    UNION
    SELECT p.id FROM groups g JOIN up ON g.id = up.id JOIN groups p ON p.id = g.parent_id
    WHERE p.deleted_at IS NULL
)
SELECT id FROM up`

// GroupsWithDescendants renvoie les groupes donnés et tous leurs sous-groupes (triés).
func GroupsWithDescendants(db *gorm.DB, ids []uint) ([]uint, error) {
	return walkGroups(db, descendantsSQL, ids)
}

// GroupsWithAncestors renvoie les groupes donnés et tous les groupes au-dessus (triés).
func GroupsWithAncestors(db *gorm.DB, ids []uint) ([]uint, error) {
	return walkGroups(db, ancestorsSQL, ids)
}

func walkGroups(db *gorm.DB, query string, ids []uint) ([]uint, error) {
	out := []uint{}
	if len(ids) == 0 {
		return out, nil
	}
	if err := db.Raw(query, ids).Scan(&out).Error; err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out, nil
}

// WouldCreateCycle : donner parentID pour parent à groupID formerait une boucle (le parent est
// le groupe lui-même ou l'un de ses sous-groupes).
func WouldCreateCycle(db *gorm.DB, groupID, parentID uint) (bool, error) {
	if groupID == parentID {
		return true, nil
	}
	below, err := GroupsWithDescendants(db, []uint{groupID})
	if err != nil {
		return false, err
	}
	for _, id := range below {
		if id == parentID {
			return true, nil
		}
	}
	return false, nil
}
