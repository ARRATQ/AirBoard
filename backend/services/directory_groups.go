package services

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Synchronisation trombinoscope → groupes d'utilisateurs.
//
// Une entité (org_units) peut être liée à un groupe (org_units.group_id). Les personnes
// affectées à un poste de l'entité (intérims compris), actives et liées à un compte, sont
// membres de ce groupe ; les groupes au-dessus leur sont acquis par la hiérarchie des groupes.
// La synchronisation marque ses ajouts (user_groups.source = 'directory') et ne retire qu'eux :
// les membres ajoutés à la main restent.
//
// Elle lie aussi une fiche sans compte au compte qui a exactement le même e-mail (aux
// différences de casse près), si ce compte n'est lié à aucune autre fiche.

// DirectoryGroupsResult résume une synchronisation.
type DirectoryGroupsResult struct {
	LinkedPeople int64 `json:"linked_people"` // fiches liées à un compte par l'e-mail
	Added        int64 `json:"added"`         // appartenances ajoutées
	Removed      int64 `json:"removed"`       // appartenances retirées
}

// Verrou de transaction PostgreSQL : deux synchronisations ne se croisent pas.
const directoryGroupsLock = 727001

const linkPeopleByEmailSQL = `
UPDATE directory_people p SET user_id = m.user_id, updated_at = NOW()
FROM (
    SELECT DISTINCT ON (u.id) u.id AS user_id, p2.id AS person_id
    FROM directory_people p2
    JOIN users u ON LOWER(u.email) = LOWER(p2.email) AND u.deleted_at IS NULL
    WHERE p2.user_id IS NULL AND p2.deleted_at IS NULL AND TRIM(p2.email) <> ''
      AND NOT EXISTS (
          SELECT 1 FROM directory_people q WHERE q.user_id = u.id AND q.deleted_at IS NULL
      )
    ORDER BY u.id, p2.id
) m
WHERE p.id = m.person_id`

// wantedMembershipsSQL : appartenances que le trombinoscope demande (user_id, group_id).
const wantedMembershipsSQL = `
SELECT DISTINCT p.user_id, u.group_id
FROM position_assignments a
JOIN org_positions pos ON pos.id = a.position_id AND pos.deleted_at IS NULL
JOIN org_units u ON u.id = pos.unit_id AND u.deleted_at IS NULL AND u.group_id IS NOT NULL
JOIN groups g ON g.id = u.group_id AND g.deleted_at IS NULL
JOIN directory_people p ON p.id = a.person_id AND p.deleted_at IS NULL AND p.is_active
    AND p.user_id IS NOT NULL
JOIN users us ON us.id = p.user_id AND us.deleted_at IS NULL
WHERE a.deleted_at IS NULL`

// SyncDirectoryGroups recalcule les appartenances venues du trombinoscope. Elle est
// idempotente : la relancer sans changement ne modifie rien.
func SyncDirectoryGroups(db *gorm.DB) (DirectoryGroupsResult, error) {
	var res DirectoryGroupsResult
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", directoryGroupsLock).Error; err != nil {
			return fmt.Errorf("verrou : %w", err)
		}
		linked := tx.Exec(linkPeopleByEmailSQL)
		if linked.Error != nil {
			return fmt.Errorf("liaison par e-mail : %w", linked.Error)
		}
		res.LinkedPeople = linked.RowsAffected

		removed := tx.Exec(`DELETE FROM user_groups ug
			WHERE ug.source = 'directory'
			  AND NOT EXISTS (SELECT 1 FROM (` + wantedMembershipsSQL + `) w
			                  WHERE w.user_id = ug.user_id AND w.group_id = ug.group_id)`)
		if removed.Error != nil {
			return fmt.Errorf("retraits : %w", removed.Error)
		}
		res.Removed = removed.RowsAffected

		added := tx.Exec(`INSERT INTO user_groups (user_id, group_id, source)
			SELECT user_id, group_id, 'directory' FROM (` + wantedMembershipsSQL + `) w
			ON CONFLICT DO NOTHING`)
		if added.Error != nil {
			return fmt.Errorf("ajouts : %w", added.Error)
		}
		res.Added = added.RowsAffected
		return nil
	})
	return res, err
}

// SyncDirectoryGroupsLogged lance la synchronisation et journalise l'échec : à appeler après
// une modification du trombinoscope ou des comptes, sans faire échouer la requête.
func SyncDirectoryGroupsLogged(db *gorm.DB, why string) {
	res, err := SyncDirectoryGroups(db)
	if err != nil {
		log.Printf("[Trombinoscope] synchronisation des groupes (%s) : %v", why, err)
		return
	}
	if res.LinkedPeople+res.Added+res.Removed > 0 {
		log.Printf("[Trombinoscope] groupes synchronisés (%s) : %d fiche(s) liée(s), +%d / -%d appartenance(s)",
			why, res.LinkedPeople, res.Added, res.Removed)
	}
}
