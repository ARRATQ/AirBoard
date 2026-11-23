package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"airboard/apierror"
	"airboard/middleware"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Groupes d'utilisateurs gérés par la permission groups.manage (page « Mes groupes »).
//
// Portée « ses groupes » : la personne crée des sous-groupes sous les groupes que la permission
// couvre (et sous leurs sous-groupes), les modifie et les supprime, et n'y met que des membres
// des groupes couverts : un sous-groupe donne accès à ce qui cible les groupes au-dessus, il ne
// doit pas ouvrir ces contenus à quelqu'un d'extérieur. Les groupes couverts eux-mêmes restent à
// l'admin.
// Portée « tout le site » : tous les groupes, avec n'importe quel compte.
// Dans les deux cas, le groupe par défaut et les groupes liés au trombinoscope restent à l'admin.

const (
	msgGroupNotManageable   = "Ce groupe n'est pas modifiable avec vos droits"
	msgGroupParentForbidden = "Vous ne pouvez créer un groupe que sous l'un de vos groupes"
	msgGroupMembersOutside  = "Vous ne pouvez ajouter que des membres de vos groupes"
	msgGroupNameTaken       = "Un groupe porte déjà ce nom"
	msgGroupNameRequired    = "Le nom du groupe est obligatoire"
	maxGroupNameLength      = 100
	maxCandidateResults     = 50
)

// groupScope : ce que la requête peut faire sur les groupes.
type groupScope struct {
	all      bool          // portée « tout le site » (admin compris)
	roots    map[uint]bool // groupes couverts directement (portée « ses groupes »)
	covered  map[uint]bool // groupes couverts, sous-groupes compris
	linked   map[uint]bool // groupes liés à une entité du trombinoscope
	everyone uint          // groupe par défaut
}

func (h *GroupAdminHandler) loadGroupScope(c *gin.Context) (*groupScope, error) {
	s := &groupScope{roots: map[uint]bool{}, covered: map[uint]bool{}, linked: map[uint]bool{}}
	var err error
	if s.everyone, err = models.EveryoneGroupID(h.db); err != nil {
		return nil, err
	}
	var linked []uint
	if err := h.db.Model(&models.OrgUnit{}).Where("group_id IS NOT NULL").Pluck("group_id", &linked).Error; err != nil {
		return nil, err
	}
	for _, id := range linked {
		s.linked[id] = true
	}
	if middleware.HasPermissionEverywhere(c, models.PermGroupsManage) {
		s.all = true
		return s, nil
	}
	roots, err := models.PermissionRootGroups(h.db, c.GetUint("user_id"), models.PermGroupsManage)
	if err != nil {
		return nil, err
	}
	covered, err := models.GroupsWithDescendants(h.db, roots)
	if err != nil {
		return nil, err
	}
	// Groupes couverts « en tête » seulement : un groupe sous un autre groupe couvert (un
	// sous-groupe dont la personne est aussi membre, par exemple) reste modifiable.
	for _, id := range roots {
		s.roots[id] = true
	}
	for _, id := range roots {
		above, err := models.GroupsWithAncestors(h.db, []uint{id})
		if err != nil {
			return nil, err
		}
		for _, a := range above {
			if a != id && s.roots[a] {
				delete(s.roots, id)
				break
			}
		}
	}
	for _, id := range covered {
		if id != s.everyone {
			s.covered[id] = true
		}
	}
	return s, nil
}

func (s *groupScope) visible(id uint) bool { return s.all || s.covered[id] }

// editable : la requête peut modifier ou supprimer ce groupe.
func (s *groupScope) editable(id uint) bool {
	if id == s.everyone || s.linked[id] {
		return false
	}
	return s.all || (s.covered[id] && !s.roots[id])
}

// parentAllowed : la requête peut placer un groupe sous parentID (nil : groupe principal).
func (s *groupScope) parentAllowed(parentID *uint) bool {
	if parentID == nil {
		return s.all
	}
	if *parentID == s.everyone {
		return false
	}
	return s.all || s.covered[*parentID]
}

// allowedMembers : comptes que l'on peut mettre dans un groupe placé sous parentID, c'est-à-dire
// les membres (sous-groupes compris) des groupes couverts au-dessus de ce parent. nil : tous
// (portée « tout le site »).
func (h *GroupAdminHandler) allowedMembers(s *groupScope, parentID *uint) (map[uint]bool, error) {
	if s.all {
		return nil, nil
	}
	out := map[uint]bool{}
	if parentID == nil {
		return out, nil
	}
	above, err := models.GroupsWithAncestors(h.db, []uint{*parentID})
	if err != nil {
		return nil, err
	}
	roots := []uint{}
	for _, id := range above {
		if s.roots[id] {
			roots = append(roots, id)
		}
	}
	if len(roots) == 0 {
		return out, nil
	}
	var ids []uint
	if err := h.db.Table("user_effective_groups").Where("group_id IN ?", roots).Distinct().Pluck("user_id", &ids).Error; err != nil {
		return nil, err
	}
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// managedGroupView : un groupe tel que la page « Mes groupes » l'affiche.
type managedGroupView struct {
	models.Group
	Members  []groupMemberView `json:"members"`
	Editable bool              `json:"editable"`
}

type groupMemberView struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// GetUserGroups liste les groupes visibles avec cette permission et dit lesquels sont modifiables.
func (h *GroupAdminHandler) GetUserGroups(c *gin.Context) {
	s, err := h.loadGroupScope(c)
	if err != nil {
		respondGroupScopeError(c, err)
		return
	}
	groups := []models.Group{}
	query := h.db.Order("name ASC")
	if !s.all {
		query = query.Where("id IN ?", keysOf(s.covered))
	}
	if s.all || len(s.covered) > 0 {
		if err := query.Find(&groups).Error; err != nil {
			apierror.Respond(c, http.StatusInternalServerError, "Lecture des groupes impossible")
			return
		}
	}
	attachOrgUnits(h.db, groups)
	members, err := h.membersByGroup(groupIDsOf(groups))
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Lecture des membres impossible")
		return
	}
	out := make([]managedGroupView, 0, len(groups))
	for _, g := range groups {
		g.IsDefault = g.ID == s.everyone
		list := members[g.ID]
		if list == nil {
			list = []groupMemberView{}
		}
		out = append(out, managedGroupView{Group: g, Members: list, Editable: s.editable(g.ID)})
	}
	c.JSON(http.StatusOK, gin.H{"groups": out, "can_create_root": s.all})
}

func (h *GroupAdminHandler) membersByGroup(ids []uint) (map[uint][]groupMemberView, error) {
	out := map[uint][]groupMemberView{}
	if len(ids) == 0 {
		return out, nil
	}
	var rows []struct {
		GroupID   uint
		ID        uint
		Username  string
		FirstName string
		LastName  string
		Email     string
	}
	err := h.db.Table("user_groups ug").
		Select("ug.group_id, u.id, u.username, u.first_name, u.last_name, u.email").
		Joins("JOIN users u ON u.id = ug.user_id AND u.deleted_at IS NULL").
		Where("ug.group_id IN ?", ids).
		Order("u.last_name, u.first_name, u.username").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.GroupID] = append(out[r.GroupID], groupMemberView{ID: r.ID, Username: r.Username, FirstName: r.FirstName, LastName: r.LastName, Email: r.Email})
	}
	return out, nil
}

// userGroupRequest : corps de création et de modification. ParentID présent mais null : groupe
// principal ; absent (modification) : inchangé. UserIDs absent (modification) : membres inchangés.
type userGroupRequest struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Color       *string         `json:"color"`
	IsActive    *bool           `json:"is_active"`
	ParentID    json.RawMessage `json:"parent_id"`
	UserIDs     *[]uint         `json:"user_ids"`
}

// parentFrom lit parent_id : (présent, valeur). Erreur si ce n'est ni null ni un identifiant.
func parentFrom(raw json.RawMessage) (bool, *uint, error) {
	if len(raw) == 0 {
		return false, nil, nil
	}
	if string(raw) == "null" {
		return true, nil, nil
	}
	var id uint
	if err := json.Unmarshal(raw, &id); err != nil || id == 0 {
		return true, nil, errors.New("parent invalide")
	}
	return true, &id, nil
}

// CreateUserGroup crée un groupe (portée « ses groupes » : un sous-groupe d'un de ses groupes).
func (h *GroupAdminHandler) CreateUserGroup(c *gin.Context) {
	s, err := h.loadGroupScope(c)
	if err != nil {
		respondGroupScopeError(c, err)
		return
	}
	var req userGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	_, parentID, err := parentFrom(req.ParentID)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Groupe parent invalide")
		return
	}
	if !s.parentAllowed(parentID) || (parentID != nil && !h.groupExists(*parentID)) {
		apierror.Respond(c, http.StatusForbidden, msgGroupParentForbidden)
		return
	}
	name, msg := cleanGroupName(req.Name)
	if msg != "" {
		apierror.Respond(c, http.StatusBadRequest, msg)
		return
	}
	if h.groupNameTaken(name, 0) {
		apierror.Respond(c, http.StatusConflict, msgGroupNameTaken)
		return
	}
	userIDs := []uint{}
	if req.UserIDs != nil {
		userIDs = *req.UserIDs
	}
	if msg, status := h.checkNewMembers(s, parentID, userIDs, nil); msg != "" {
		apierror.Respond(c, status, msg)
		return
	}

	group := models.Group{
		Name:        name,
		Description: strings.TrimSpace(derefString(req.Description)),
		Color:       defaultColor(derefString(req.Color)),
		IsActive:    req.IsActive == nil || *req.IsActive,
		ParentID:    parentID,
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&group).Error; err != nil {
			return err
		}
		if !group.IsActive {
			if err := tx.Model(&group).Update("is_active", false).Error; err != nil {
				return err
			}
		}
		return replaceManualMembers(tx, group.ID, userIDs)
	})
	if err != nil {
		log.Printf("mes groupes : création : %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de créer le groupe")
		return
	}
	c.JSON(http.StatusCreated, group)
}

// UpdateUserGroup modifie un groupe modifiable avec cette permission.
func (h *GroupAdminHandler) UpdateUserGroup(c *gin.Context) {
	s, group, ok := h.editableGroup(c)
	if !ok {
		return
	}
	var req userGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.InvalidBody(c, err)
		return
	}
	updates := map[string]interface{}{}
	if req.Name != nil {
		name, msg := cleanGroupName(req.Name)
		if msg != "" {
			apierror.Respond(c, http.StatusBadRequest, msg)
			return
		}
		if h.groupNameTaken(name, group.ID) {
			apierror.Respond(c, http.StatusConflict, msgGroupNameTaken)
			return
		}
		updates["name"] = name
	}
	if req.Description != nil {
		updates["description"] = strings.TrimSpace(*req.Description)
	}
	if req.Color != nil {
		updates["color"] = defaultColor(*req.Color)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	parentID := group.ParentID
	present, newParent, err := parentFrom(req.ParentID)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Groupe parent invalide")
		return
	}
	if present && !sameParent(group.ParentID, newParent) {
		if msg, status := h.checkMove(s, group.ID, newParent); msg != "" {
			apierror.Respond(c, status, msg)
			return
		}
		parentID = newParent
		updates["parent_id"] = newParent
	}

	if req.UserIDs != nil {
		current, err := h.memberIDs(group.ID)
		if err != nil {
			apierror.Respond(c, http.StatusInternalServerError, "Lecture des membres impossible")
			return
		}
		if msg, status := h.checkNewMembers(s, parentID, *req.UserIDs, current); msg != "" {
			apierror.Respond(c, status, msg)
			return
		}
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&models.Group{}).Where("id = ?", group.ID).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.UserIDs != nil {
			return replaceManualMembers(tx, group.ID, *req.UserIDs)
		}
		return nil
	})
	if err != nil {
		log.Printf("mes groupes : modification du groupe %d : %v", group.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de modifier le groupe")
		return
	}
	h.db.First(&group, group.ID)
	c.JSON(http.StatusOK, group)
}

// DeleteUserGroup supprime un groupe modifiable ; ses sous-groupes remontent d'un niveau.
func (h *GroupAdminHandler) DeleteUserGroup(c *gin.Context) {
	_, group, ok := h.editableGroup(c)
	if !ok {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Group{}).Where("parent_id = ?", group.ID).Update("parent_id", group.ParentID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM user_groups WHERE group_id = ?", group.ID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM group_app_groups WHERE group_id = ?", group.ID).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Group{}, group.ID).Error
	})
	if err != nil {
		log.Printf("mes groupes : suppression du groupe %d : %v", group.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Impossible de supprimer le groupe")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Groupe supprimé"})
}

// GetUserGroupCandidates : comptes que l'on peut mettre dans un groupe placé sous parent_id
// (?parent_id=12&q=texte), 50 au plus.
func (h *GroupAdminHandler) GetUserGroupCandidates(c *gin.Context) {
	s, err := h.loadGroupScope(c)
	if err != nil {
		respondGroupScopeError(c, err)
		return
	}
	var parentID *uint
	if raw := c.Query("parent_id"); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || id == 0 {
			apierror.Respond(c, http.StatusBadRequest, "Groupe parent invalide")
			return
		}
		p := uint(id)
		parentID = &p
	}
	if !s.parentAllowed(parentID) {
		apierror.Respond(c, http.StatusForbidden, msgGroupParentForbidden)
		return
	}
	allowed, err := h.allowedMembers(s, parentID)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Lecture des membres impossible")
		return
	}
	users := []groupMemberView{}
	if allowed != nil && len(allowed) == 0 {
		c.JSON(http.StatusOK, users)
		return
	}
	query := h.db.Table("users").Select("id, username, first_name, last_name, email").
		Where("deleted_at IS NULL AND is_active = ?", true)
	if allowed != nil {
		query = query.Where("id IN ?", keysOf(allowed))
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("LOWER(username) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name || ' ' || last_name) LIKE ?", like, like, like, like, like)
	}
	if err := query.Order("last_name, first_name, username").Limit(maxCandidateResults).Scan(&users).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Recherche impossible")
		return
	}
	c.JSON(http.StatusOK, users)
}

// editableGroup lit le groupe :id et vérifie qu'il est modifiable ; répond sinon.
func (h *GroupAdminHandler) editableGroup(c *gin.Context) (*groupScope, models.Group, bool) {
	var group models.Group
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		apierror.Respond(c, http.StatusBadRequest, "Identifiant invalide")
		return nil, group, false
	}
	s, err := h.loadGroupScope(c)
	if err != nil {
		respondGroupScopeError(c, err)
		return nil, group, false
	}
	if err := h.db.First(&group, id).Error; err != nil || !s.visible(group.ID) {
		apierror.Respond(c, http.StatusNotFound, "Groupe introuvable")
		return nil, group, false
	}
	if !s.editable(group.ID) {
		apierror.Respond(c, http.StatusForbidden, msgGroupNotManageable)
		return nil, group, false
	}
	return s, group, true
}

// checkMove : placer groupID sous newParent est permis (pas de boucle, parent autorisé, et en
// portée « ses groupes », les membres du groupe et de ses sous-groupes restent des membres des
// groupes couverts au-dessus du nouveau parent).
func (h *GroupAdminHandler) checkMove(s *groupScope, groupID uint, newParent *uint) (string, int) {
	if !s.parentAllowed(newParent) || (newParent != nil && !h.groupExists(*newParent)) {
		return msgGroupParentForbidden, http.StatusForbidden
	}
	if newParent != nil {
		cycle, err := models.WouldCreateCycle(h.db, groupID, *newParent)
		if err != nil {
			return "Groupe parent invalide", http.StatusInternalServerError
		}
		if cycle {
			return "Un groupe ne peut pas être placé sous lui-même ni sous l'un de ses sous-groupes", http.StatusBadRequest
		}
	}
	allowed, err := h.allowedMembers(s, newParent)
	if err != nil {
		return "Lecture des membres impossible", http.StatusInternalServerError
	}
	if allowed == nil {
		return "", 0
	}
	subtree, err := models.GroupsWithDescendants(h.db, []uint{groupID})
	if err != nil {
		return "Lecture des groupes impossible", http.StatusInternalServerError
	}
	var members []uint
	if err := h.db.Table("user_groups").Where("group_id IN ?", subtree).Distinct().Pluck("user_id", &members).Error; err != nil {
		return "Lecture des membres impossible", http.StatusInternalServerError
	}
	for _, id := range members {
		if !allowed[id] {
			return "Ce groupe a des membres extérieurs aux groupes du nouveau parent", http.StatusForbidden
		}
	}
	return "", 0
}

// checkNewMembers : les comptes ajoutés (absents de current) existent et, en portée « ses
// groupes », sont membres des groupes couverts au-dessus de parentID. Retirer reste libre.
func (h *GroupAdminHandler) checkNewMembers(s *groupScope, parentID *uint, userIDs []uint, current map[uint]bool) (string, int) {
	added := []uint{}
	for _, id := range uniqueIDs(userIDs) {
		if !current[id] {
			added = append(added, id)
		}
	}
	if len(added) == 0 {
		return "", 0
	}
	var n int64
	if err := h.db.Model(&models.User{}).Where("id IN ?", added).Count(&n).Error; err != nil {
		return "Lecture des comptes impossible", http.StatusInternalServerError
	}
	if int(n) != len(added) {
		return "Compte introuvable", http.StatusBadRequest
	}
	allowed, err := h.allowedMembers(s, parentID)
	if err != nil {
		return "Lecture des membres impossible", http.StatusInternalServerError
	}
	if allowed == nil {
		return "", 0
	}
	for _, id := range added {
		if !allowed[id] {
			return msgGroupMembersOutside, http.StatusForbidden
		}
	}
	return "", 0
}

// replaceManualMembers fixe les membres ajoutés à la main ; les membres venus du trombinoscope
// ne sont pas touchés.
func replaceManualMembers(tx *gorm.DB, groupID uint, userIDs []uint) error {
	ids := uniqueIDs(userIDs)
	if len(ids) == 0 {
		return tx.Exec("DELETE FROM user_groups WHERE group_id = ? AND source = 'manual'", groupID).Error
	}
	if err := tx.Exec("DELETE FROM user_groups WHERE group_id = ? AND source = 'manual' AND user_id NOT IN ?", groupID, ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		if err := tx.Exec("INSERT INTO user_groups (group_id, user_id, source) VALUES (?, ?, 'manual') ON CONFLICT DO NOTHING", groupID, id).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *GroupAdminHandler) memberIDs(groupID uint) (map[uint]bool, error) {
	var ids []uint
	if err := h.db.Table("user_groups").Where("group_id = ?", groupID).Pluck("user_id", &ids).Error; err != nil {
		return nil, err
	}
	out := make(map[uint]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

func (h *GroupAdminHandler) groupExists(id uint) bool {
	var n int64
	h.db.Model(&models.Group{}).Where("id = ?", id).Count(&n)
	return n > 0
}

// groupNameTaken : les noms de groupes sont uniques, groupes supprimés compris.
func (h *GroupAdminHandler) groupNameTaken(name string, exceptID uint) bool {
	var n int64
	h.db.Unscoped().Model(&models.Group{}).Where("name = ? AND id <> ?", name, exceptID).Count(&n)
	return n > 0
}

func cleanGroupName(raw *string) (string, string) {
	name := strings.TrimSpace(derefString(raw))
	if name == "" {
		return "", msgGroupNameRequired
	}
	if len([]rune(name)) > maxGroupNameLength {
		return "", "Le nom du groupe est trop long"
	}
	return name, ""
}

func respondGroupScopeError(c *gin.Context, err error) {
	log.Printf("mes groupes : droits : %v", err)
	apierror.Respond(c, http.StatusInternalServerError, "Lecture des droits impossible")
}

func sameParent(a, b *uint) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func keysOf(m map[uint]bool) []uint {
	out := make([]uint, 0, len(m))
	for id := range m {
		out = append(out, id)
	}
	return out
}

func uniqueIDs(ids []uint) []uint {
	seen := map[uint]bool{}
	out := []uint{}
	for _, id := range ids {
		if id != 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}
