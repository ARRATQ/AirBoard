package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"airboard/apierror"
	"airboard/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// accessRoleNameMax borne la longueur du nom d'un rôle (en caractères).
const accessRoleNameMax = 80

// accessRoleDescriptionMax borne la longueur de sa description.
const accessRoleDescriptionMax = 300

// AccessRolesHandler gère les rôles créés par l'administrateur et leurs membres.
type AccessRolesHandler struct {
	db *gorm.DB
	// onChange vide le cache des droits : un rôle modifié s'applique à la requête suivante.
	onChange func()
}

func NewAccessRolesHandler(db *gorm.DB, onChange func()) *AccessRolesHandler {
	if onChange == nil {
		onChange = func() {}
	}
	return &AccessRolesHandler{db: db, onChange: onChange}
}

// accessRoleRequest est le corps de création ou de modification d'un rôle.
type accessRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Permissions []struct {
		Permission string `json:"permission"`
		Scope      string `json:"scope"`
	} `json:"permissions"`
}

// roleMember est un membre d'un rôle, sans données sensibles.
type roleMember struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetPermissionCatalog renvoie les permissions que l'on peut confier à un rôle.
func (h *AccessRolesHandler) GetPermissionCatalog(c *gin.Context) {
	c.JSON(http.StatusOK, models.PermissionCatalog)
}

// GetRoles renvoie les rôles, leurs permissions et leur nombre de membres.
func (h *AccessRolesHandler) GetRoles(c *gin.Context) {
	var roles []models.AccessRole
	if err := h.db.Preload("Permissions").Order("lower(name) ASC").Find(&roles).Error; err != nil {
		log.Printf("access-roles: lecture impossible : %v", err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des rôles")
		return
	}

	var counts []struct {
		AccessRoleID uint
		Total        int64
	}
	if err := h.db.Table("user_access_roles").
		Select("user_access_roles.access_role_id, COUNT(*) AS total").
		Joins("JOIN users ON users.id = user_access_roles.user_id AND users.deleted_at IS NULL").
		Group("user_access_roles.access_role_id").
		Scan(&counts).Error; err != nil {
		log.Printf("access-roles: comptage des membres impossible : %v", err)
	}
	byRole := make(map[uint]int64, len(counts))
	for _, row := range counts {
		byRole[row.AccessRoleID] = row.Total
	}
	for i := range roles {
		roles[i].UserCount = byRole[roles[i].ID]
		if roles[i].Permissions == nil {
			roles[i].Permissions = []models.AccessRolePermission{}
		}
	}
	c.JSON(http.StatusOK, roles)
}

// CreateRole crée un rôle avec ses permissions.
func (h *AccessRolesHandler) CreateRole(c *gin.Context) {
	role, perms, ok := h.bindRole(c)
	if !ok {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Permissions").Create(&role).Error; err != nil {
			return err
		}
		return replaceRolePermissions(tx, role.ID, perms)
	})
	if !h.respondWriteError(c, err) {
		return
	}
	h.onChange()
	role.Permissions = perms
	c.JSON(http.StatusCreated, role)
}

// UpdateRole renomme un rôle et remplace ses permissions.
func (h *AccessRolesHandler) UpdateRole(c *gin.Context) {
	existing, ok := h.findRole(c)
	if !ok {
		return
	}
	role, perms, ok := h.bindRole(c)
	if !ok {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"name":        role.Name,
			"description": role.Description,
		}).Error; err != nil {
			return err
		}
		return replaceRolePermissions(tx, existing.ID, perms)
	})
	if !h.respondWriteError(c, err) {
		return
	}
	h.onChange()
	existing.Name, existing.Description, existing.Permissions = role.Name, role.Description, perms
	c.JSON(http.StatusOK, existing)
}

// DeleteRole supprime un rôle ; ses membres perdent les permissions qu'il donnait.
func (h *AccessRolesHandler) DeleteRole(c *gin.Context) {
	role, ok := h.findRole(c)
	if !ok {
		return
	}
	if err := h.db.Delete(&role).Error; err != nil {
		log.Printf("access-roles: suppression de %d impossible : %v", role.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression du rôle")
		return
	}
	h.onChange()
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Rôle supprimé"})
}

// GetRoleMembers renvoie les comptes (non supprimés) qui ont ce rôle.
func (h *AccessRolesHandler) GetRoleMembers(c *gin.Context) {
	role, ok := h.findRole(c)
	if !ok {
		return
	}
	members := []roleMember{}
	if err := h.db.Table("users").
		Select("users.id, users.username, users.email, users.first_name, users.last_name").
		Joins("JOIN user_access_roles ON user_access_roles.user_id = users.id").
		Where("user_access_roles.access_role_id = ? AND users.deleted_at IS NULL", role.ID).
		Order("lower(users.last_name), lower(users.first_name), users.id").
		Scan(&members).Error; err != nil {
		log.Printf("access-roles: membres de %d illisibles : %v", role.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des membres")
		return
	}
	c.JSON(http.StatusOK, members)
}

// SetRoleMembers remplace la liste des membres d'un rôle.
func (h *AccessRolesHandler) SetRoleMembers(c *gin.Context) {
	role, ok := h.findRole(c)
	if !ok {
		return
	}
	var req struct {
		UserIDs []uint `json:"user_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Format de requête invalide")
		return
	}
	ids := uniqueUint(req.UserIDs)
	if len(ids) > 0 {
		var found int64
		if err := h.db.Model(&models.User{}).Where("id IN ?", ids).Count(&found).Error; err != nil {
			apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la vérification des utilisateurs")
			return
		}
		if found != int64(len(ids)) {
			apierror.Respond(c, http.StatusBadRequest, "Certains utilisateurs n'existent pas")
			return
		}
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM user_access_roles WHERE access_role_id = ?", role.ID).Error; err != nil {
			return err
		}
		for _, id := range ids {
			if err := tx.Exec("INSERT INTO user_access_roles (user_id, access_role_id) VALUES (?, ?)", id, role.ID).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("access-roles: membres de %d non enregistrés : %v", role.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de l'enregistrement des membres")
		return
	}
	h.onChange()
	c.JSON(http.StatusOK, models.SuccessResponse{Message: "Membres du rôle enregistrés"})
}

// findRole charge le rôle désigné par :id, ou répond 400/404.
func (h *AccessRolesHandler) findRole(c *gin.Context) (models.AccessRole, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return models.AccessRole{}, false
	}
	var role models.AccessRole
	if err := h.db.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.Respond(c, http.StatusNotFound, "Rôle introuvable")
		} else {
			apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la lecture du rôle")
		}
		return models.AccessRole{}, false
	}
	return role, true
}

// bindRole lit et valide le corps : nom, description, permissions du catalogue.
func (h *AccessRolesHandler) bindRole(c *gin.Context) (models.AccessRole, []models.AccessRolePermission, bool) {
	var req accessRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Format de requête invalide")
		return models.AccessRole{}, nil, false
	}
	role, perms, msg := validateAccessRole(req)
	if msg != "" {
		apierror.Respond(c, http.StatusBadRequest, msg)
		return models.AccessRole{}, nil, false
	}
	return role, perms, true
}

// validateAccessRole nettoie la demande ; renvoie un message d'erreur lisible si elle est invalide.
func validateAccessRole(req accessRoleRequest) (models.AccessRole, []models.AccessRolePermission, string) {
	name := strings.TrimSpace(req.Name)
	description := strings.TrimSpace(req.Description)
	switch {
	case name == "":
		return models.AccessRole{}, nil, "Le nom du rôle est obligatoire"
	case utf8.RuneCountInString(name) > accessRoleNameMax:
		return models.AccessRole{}, nil, "Le nom du rôle ne doit pas dépasser 80 caractères"
	case utf8.RuneCountInString(description) > accessRoleDescriptionMax:
		return models.AccessRole{}, nil, "La description ne doit pas dépasser 300 caractères"
	case strings.EqualFold(name, "admin") || strings.EqualFold(name, "administrateur"):
		return models.AccessRole{}, nil, "Ce nom est réservé au rôle administrateur"
	}

	perms := make([]models.AccessRolePermission, 0, len(req.Permissions))
	seen := map[string]bool{}
	for _, p := range req.Permissions {
		def, ok := models.LookupPermission(p.Permission)
		if !ok {
			return models.AccessRole{}, nil, "Permission inconnue : " + p.Permission
		}
		if !def.AcceptsScope(p.Scope) {
			return models.AccessRole{}, nil, "Portée invalide pour la permission " + p.Permission
		}
		if seen[p.Permission] {
			return models.AccessRole{}, nil, "Permission en double : " + p.Permission
		}
		seen[p.Permission] = true
		perms = append(perms, models.AccessRolePermission{Permission: p.Permission, Scope: p.Scope})
	}
	return models.AccessRole{Name: name, Description: description}, perms, ""
}

// replaceRolePermissions remplace les permissions d'un rôle dans la transaction.
func replaceRolePermissions(tx *gorm.DB, roleID uint, perms []models.AccessRolePermission) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&models.AccessRolePermission{}).Error; err != nil {
		return err
	}
	for i := range perms {
		perms[i].RoleID = roleID
		if err := tx.Create(&perms[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

// respondWriteError répond à une erreur d'écriture ; renvoie true s'il n'y en a pas.
func (h *AccessRolesHandler) respondWriteError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}
	if isUniqueViolation(err) {
		apierror.Respond(c, http.StatusConflict, "Un rôle porte déjà ce nom")
		return false
	}
	log.Printf("access-roles: enregistrement impossible : %v", err)
	apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de l'enregistrement du rôle")
	return false
}

// uniqueUint retire les doublons et les zéros, en gardant l'ordre.
func uniqueUint(ids []uint) []uint {
	seen := make(map[uint]bool, len(ids))
	out := make([]uint, 0, len(ids))
	for _, id := range ids {
		if id == 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
