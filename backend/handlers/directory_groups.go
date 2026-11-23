package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"airboard/apierror"
	"airboard/models"
	"airboard/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Groupes d'utilisateurs liés aux entités du trombinoscope (admin). Les membres venus du
// trombinoscope sont tenus à jour par services.SyncDirectoryGroups.

var errDirectoryGroupConflict = errors.New("groupe déjà lié")

// LinkUnitGroup lie une entité à un groupe existant ({"group_id": 12}) ou retire le lien
// ({"group_id": null}).
func (h *DirectoryHandler) LinkUnitGroup(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var unit models.OrgUnit
	if err := h.db.First(&unit, id).Error; err != nil {
		respondNotFoundOr500(c, err, "Entité introuvable")
		return
	}
	var req struct {
		GroupID *uint `json:"group_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	if req.GroupID != nil {
		if msg, status := h.checkLinkableGroup(*req.GroupID, unit.ID); msg != "" {
			apierror.Respond(c, status, msg)
			return
		}
	}
	if err := h.db.Model(&unit).Update("group_id", req.GroupID).Error; err != nil {
		log.Printf("trombinoscope : lien de l'entité %d au groupe : %v", unit.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de lier le groupe")
		return
	}
	result, err := services.SyncDirectoryGroups(h.db)
	if err != nil {
		log.Printf("trombinoscope : synchronisation après lien : %v", err)
	}
	h.db.First(&unit, id)
	c.JSON(http.StatusOK, gin.H{"unit": unit, "sync": result})
}

// checkLinkableGroup renvoie un message d'erreur si le groupe ne peut pas être lié à l'entité.
func (h *DirectoryHandler) checkLinkableGroup(groupID, unitID uint) (string, int) {
	var group models.Group
	if err := h.db.First(&group, groupID).Error; err != nil {
		return "Groupe introuvable", http.StatusBadRequest
	}
	if everyone, err := models.EveryoneGroupID(h.db); err == nil && everyone == groupID {
		return "Le groupe par défaut réunit tout le monde : il ne se lie pas à une entité", http.StatusBadRequest
	}
	var other models.OrgUnit
	err := h.db.Where("group_id = ? AND id <> ?", groupID, unitID).First(&other).Error
	if err == nil {
		return fmt.Sprintf("Ce groupe est déjà lié à l'entité « %s »", other.Name), http.StatusConflict
	}
	return "", 0
}

// unitGroupsResult résume la création des groupes d'une branche du trombinoscope.
type unitGroupsResult struct {
	Created int                            `json:"created"` // groupes créés
	Linked  int                            `json:"linked"`  // groupes existants du même nom liés
	Already int                            `json:"already"` // entités déjà liées, laissées telles quelles
	Sync    services.DirectoryGroupsResult `json:"sync"`
}

// CreateUnitGroups crée les groupes d'une entité et de toutes ses sous-entités, avec la même
// hiérarchie. Une entité déjà liée garde son groupe ; un groupe existant du même nom, libre,
// est lié au lieu d'être recréé.
func (h *DirectoryHandler) CreateUnitGroups(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var units []models.OrgUnit
	if err := h.db.Order("\"order\" ASC, id ASC").Find(&units).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Lecture du trombinoscope impossible")
		return
	}
	branch := unitBranch(units, id)
	if len(branch) == 0 {
		apierror.Respond(c, http.StatusNotFound, "Entité introuvable")
		return
	}
	byID := make(map[uint]models.OrgUnit, len(units))
	for _, u := range units {
		byID[u.ID] = u
	}

	var res unitGroupsResult
	err := h.db.Transaction(func(tx *gorm.DB) error {
		everyone, err := models.EveryoneGroupID(tx)
		if err != nil {
			return err
		}
		groupOf := map[uint]uint{} // entité → groupe
		for _, unit := range branch {
			var parentGroup *uint
			if unit.ParentID != nil {
				if g, ok := groupOf[*unit.ParentID]; ok {
					parentGroup = &g
				} else if p, ok := byID[*unit.ParentID]; ok && p.GroupID != nil && groupAlive(tx, *p.GroupID) {
					parentGroup = p.GroupID
				}
			}
			if unit.GroupID != nil && groupAlive(tx, *unit.GroupID) {
				groupOf[unit.ID] = *unit.GroupID
				res.Already++
				continue
			}
			groupID, created, err := h.groupForUnit(tx, unit, byID, parentGroup, everyone)
			if err != nil {
				return err
			}
			if err := tx.Model(&models.OrgUnit{}).Where("id = ?", unit.ID).Update("group_id", groupID).Error; err != nil {
				return err
			}
			groupOf[unit.ID] = groupID
			if created {
				res.Created++
			} else {
				res.Linked++
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("trombinoscope : création des groupes de l'entité %d : %v", id, err)
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer les groupes")
		return
	}
	if res.Sync, err = services.SyncDirectoryGroups(h.db); err != nil {
		log.Printf("trombinoscope : synchronisation après création des groupes : %v", err)
	}
	c.JSON(http.StatusOK, res)
}

// groupForUnit trouve un groupe libre du même nom que l'entité, ou en crée un sous parentGroup.
func (h *DirectoryHandler) groupForUnit(tx *gorm.DB, unit models.OrgUnit, byID map[uint]models.OrgUnit, parentGroup *uint, everyone uint) (uint, bool, error) {
	name := strings.TrimSpace(unit.Name)
	var existing models.Group
	err := tx.Where("LOWER(name) = LOWER(?) AND id <> ?", name, everyone).
		Where("NOT EXISTS (SELECT 1 FROM org_units u WHERE u.group_id = groups.id AND u.deleted_at IS NULL)").
		First(&existing).Error
	if err == nil {
		if existing.ParentID == nil && parentGroup != nil {
			if cycle, cerr := models.WouldCreateCycle(tx, existing.ID, *parentGroup); cerr == nil && !cycle {
				if err := tx.Model(&existing).Update("parent_id", *parentGroup).Error; err != nil {
					return 0, false, err
				}
			}
		}
		return existing.ID, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, err
	}
	free, err := freeGroupName(tx, name, unit, byID)
	if err != nil {
		return 0, false, err
	}
	group := models.Group{Name: free, Description: unit.Description, Color: defaultColor(unit.Color), IsActive: true, ParentID: parentGroup}
	if err := tx.Create(&group).Error; err != nil {
		return 0, false, err
	}
	return group.ID, true, nil
}

// freeGroupName : le nom de l'entité, sinon « Nom (entité parente) », sinon numéroté. Les noms
// de groupes sont uniques, groupes supprimés compris.
func freeGroupName(tx *gorm.DB, name string, unit models.OrgUnit, byID map[uint]models.OrgUnit) (string, error) {
	candidates := []string{name}
	if unit.ParentID != nil {
		if p, ok := byID[*unit.ParentID]; ok {
			label := strings.TrimSpace(p.Code)
			if label == "" {
				label = strings.TrimSpace(p.Name)
			}
			candidates = append(candidates, fmt.Sprintf("%s (%s)", name, label))
		}
	}
	for i := 2; i <= 50; i++ {
		candidates = append(candidates, fmt.Sprintf("%s %d", name, i))
	}
	for _, candidate := range candidates {
		var n int64
		if err := tx.Unscoped().Model(&models.Group{}).Where("LOWER(name) = LOWER(?)", candidate).Count(&n).Error; err != nil {
			return "", err
		}
		if n == 0 {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("aucun nom libre pour %q : %w", name, errDirectoryGroupConflict)
}

// unitBranch renvoie l'entité et ses sous-entités, chaque parent avant ses enfants.
func unitBranch(units []models.OrgUnit, rootID uint) []models.OrgUnit {
	children := map[uint][]models.OrgUnit{}
	var root *models.OrgUnit
	for i := range units {
		u := units[i]
		if u.ID == rootID {
			root = &units[i]
		}
		if u.ParentID != nil {
			children[*u.ParentID] = append(children[*u.ParentID], u)
		}
	}
	if root == nil {
		return nil
	}
	out := []models.OrgUnit{*root}
	seen := map[uint]bool{root.ID: true}
	for i := 0; i < len(out); i++ {
		for _, child := range children[out[i].ID] {
			if !seen[child.ID] {
				seen[child.ID] = true
				out = append(out, child)
			}
		}
	}
	return out
}

func groupAlive(tx *gorm.DB, id uint) bool {
	var n int64
	tx.Model(&models.Group{}).Where("id = ?", id).Count(&n)
	return n > 0
}

// SyncGroups relance la synchronisation des groupes liés au trombinoscope.
func (h *DirectoryHandler) SyncGroups(c *gin.Context) {
	result, err := services.SyncDirectoryGroups(h.db)
	if err != nil {
		log.Printf("trombinoscope : synchronisation des groupes : %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Synchronisation impossible")
		return
	}
	c.JSON(http.StatusOK, result)
}
