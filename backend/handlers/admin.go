package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"airboard/apierror"
	"airboard/config"
	"airboard/middleware"
	"airboard/models"
	"airboard/services"
	"airboard/utils"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdminHandler struct {
	db                  *gorm.DB
	bcryptCost          int
	gamificationService *services.GamificationService
}

func NewAdminHandler(db *gorm.DB, cfg *config.Config, gs *services.GamificationService) *AdminHandler {
	return &AdminHandler{
		db:                  db,
		bcryptCost:          cfg.Security.BcryptCost,
		gamificationService: gs,
	}
}

// ============ GESTION DES GROUPES D'APPLICATIONS ============

// @Summary Lister les groupes d'applications
// @Description Récupère tous les groupes d'applications (admin uniquement)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Numéro de page" default(1)
// @Param limit query int false "Nombre d'éléments par page" default(10)
// @Success 200 {object} models.PaginatedResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/app-groups [get]
func (h *AdminHandler) GetAppGroups(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var appGroups []models.AppGroup
	var total int64

	// Filtrage selon le rôle
	query := h.db.Model(&models.AppGroup{})

	// Gestion limitée à ses groupes : catégories publiques et privées de ses groupes
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	if !middleware.IsUnrestricted(c) && len(managedGroupIDs) > 0 {
		// (même logique que le dashboard pour la cohérence)
		// AppGroups accessibles à ses groupes
		query = query.Where(`
			(
				-- AppGroups publics
				(is_private = false)
			)
			OR
			(
				-- AppGroups privés de ses groupes
				is_private = true AND owner_group_id IN ?
			)
		`, managedGroupIDs)
	}
	// Admin voit tout (pas de filtre)

	query.Count(&total)
	if err := query.Preload("Applications").
		Preload("OwnerGroup").
		Order("\"order\" ASC, name ASC").
		Limit(limit).Offset(offset).
		Find(&appGroups).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des groupes d'applications")
		return
	}

	totalPages := int(total)/limit + 1
	if int(total)%limit == 0 && total > 0 {
		totalPages = int(total) / limit
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       appGroups,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	})
}

// @Summary Créer un groupe d'applications
// @Description Crée un nouveau groupe d'applications (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param appGroup body models.AppGroup true "Données du groupe d'applications"
// @Success 201 {object} models.AppGroup
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Router /admin/app-groups [post]
func (h *AdminHandler) CreateAppGroup(c *gin.Context) {
	var appGroup models.AppGroup
	var flags resourceFlags
	if err := c.ShouldBindBodyWith(&appGroup, binding.JSON); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	_ = c.ShouldBindBodyWith(&flags, binding.JSON)

	// Gestion limitée à ses groupes : catégorie privée, rattachée au groupe propriétaire choisi
	// parmi ses groupes. Sur tout le site, le choix public ou privé est libre.
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	isGroupAdmin := !middleware.IsUnrestricted(c)

	if isGroupAdmin {
		ownerID, ok := resolveOwnerGroup(appGroup.OwnerGroupID, managedGroupIDs)
		if !ok {
			apierror.Respond(c, http.StatusForbidden, "Le groupe propriétaire choisi ne fait pas partie de vos groupes")
			return
		}
		appGroup.IsPrivate = true
		appGroup.OwnerGroupID = &ownerID
		flags.IsPrivate = nil
	}

	// Créer l'AppGroup
	if err := h.db.Create(&appGroup).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création du groupe d'applications")
		return
	}

	// Create ignore les false des champs à valeur par défaut (is_active) : les appliquer ensuite
	if err := h.applyFlags(&appGroup, flags, "is_active"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création du groupe d'applications")
		return
	}

	// Gestion limitée à ses groupes : lier l'AppGroup à ses groupes via group_app_groups
	if isGroupAdmin {
		for _, groupID := range managedGroupIDs {
			h.db.Exec("INSERT INTO group_app_groups (group_id, app_group_id) VALUES (?, ?)", groupID, appGroup.ID)
		}
	}

	c.JSON(http.StatusCreated, appGroup)
}

// @Summary Modifier un groupe d'applications
// @Description Modifie un groupe d'applications existant (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID du groupe d'applications"
// @Param appGroup body models.AppGroup true "Données modifiées"
// @Success 200 {object} models.AppGroup
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/app-groups/{id} [put]
func (h *AdminHandler) UpdateAppGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var appGroup models.AppGroup
	if err := h.db.First(&appGroup, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Groupe d'applications non trouvé")
		return
	}

	// Gestion limitée à ses groupes : seulement un AppGroup privé de ses groupes
	if !middleware.CanManageAppGroupWithDB(c, appGroup.ID, h.db) {
		apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez pas modifier cette catégorie : seules les catégories privées de vos groupes peuvent l'être.")
		return
	}

	var updateData models.AppGroup
	var flags resourceFlags
	if err := c.ShouldBindBodyWith(&updateData, binding.JSON); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	_ = c.ShouldBindBodyWith(&flags, binding.JSON)

	// Gestion limitée à ses groupes : la catégorie reste privée et ne passe qu'à un de ses groupes
	if !middleware.IsUnrestricted(c) {
		updateData.IsPrivate = true
		flags.IsPrivate = nil
		if updateData.OwnerGroupID != nil && !containsUint(middleware.GetManagedGroupIDs(c), *updateData.OwnerGroupID) {
			apierror.Respond(c, http.StatusForbidden, "Le groupe propriétaire choisi ne fait pas partie de vos groupes")
			return
		}
	}

	// Updates(struct) ignore false, 0 et "" : cases à cocher, ordre et description sont appliqués à part
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&appGroup).Updates(updateData).Error; err != nil {
			return err
		}
		if values := flagUpdates(flags, "is_active", "is_private", "order", "description"); len(values) > 0 {
			return tx.Model(&appGroup).Omit(clause.Associations).Updates(values).Error
		}
		return nil
	})
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la modification")
		return
	}

	c.JSON(http.StatusOK, appGroup)
}

// @Summary Supprimer un groupe d'applications
// @Description Supprime un groupe d'applications (admin uniquement)
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "ID du groupe d'applications"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /admin/app-groups/{id} [delete]
func (h *AdminHandler) DeleteAppGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	// Gestion limitée à ses groupes : seulement un AppGroup privé de ses groupes
	if !middleware.CanManageAppGroupWithDB(c, uint(id), h.db) {
		apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez pas supprimer cette catégorie : seules les catégories privées de vos groupes peuvent l'être.")
		return
	}

	// Vérifier si le groupe existe
	var appGroup models.AppGroup
	if err := h.db.First(&appGroup, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Groupe d'applications non trouvé")
		return
	}

	// Suppression définitive (le nom redevient libre), d'un seul tenant
	if err := h.db.Transaction(func(tx *gorm.DB) error { return purgeAppGroup(tx, uint(id)) }); err != nil {
		log.Printf("[ADMIN] Suppression de la catégorie de ressources %d : %v", id, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression de la catégorie")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Groupe d'applications supprimé avec succès",
	})
}

// purgeAppGroup efface une catégorie de ressources et ses ressources (corbeille comprise),
// avec ce qui s'y rattache : clics, favoris et commentaires. Les clics et les favoris ont une
// clé étrangère vers la ressource : sans eux, la base refuserait la suppression.
func purgeAppGroup(tx *gorm.DB, appGroupID uint) error {
	resources := tx.Unscoped().Model(&models.Application{}).Select("id").Where("app_group_id = ?", appGroupID)
	steps := []struct {
		what string
		run  func() error
	}{
		{"clics", func() error {
			return tx.Where("application_id IN (?)", resources).Delete(&models.ApplicationClick{}).Error
		}},
		{"favoris", func() error {
			return tx.Exec("DELETE FROM user_favorites WHERE application_id IN (?)", resources).Error
		}},
		{"commentaires", func() error {
			return tx.Unscoped().Where("entity_type = ? AND entity_id IN (?)", "application", resources).Delete(&models.Comment{}).Error
		}},
		{"ressources", func() error {
			return tx.Unscoped().Where("app_group_id = ?", appGroupID).Delete(&models.Application{}).Error
		}},
		{"accès des groupes", func() error {
			return tx.Exec("DELETE FROM group_app_groups WHERE app_group_id = ?", appGroupID).Error
		}},
		{"catégorie", func() error {
			return tx.Unscoped().Delete(&models.AppGroup{}, appGroupID).Error
		}},
	}
	for _, step := range steps {
		if err := step.run(); err != nil {
			return fmt.Errorf("%s : %w", step.what, err)
		}
	}
	return nil
}

// ============ GESTION DES APPLICATIONS ============

// @Summary Lister les applications
// @Description Récupère toutes les applications (admin uniquement)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Numéro de page" default(1)
// @Param limit query int false "Nombre d'éléments par page" default(10)
// @Param app_group_id query int false "Filtrer par groupe d'applications"
// @Success 200 {object} models.PaginatedResponse
// @Router /admin/applications [get]
func (h *AdminHandler) GetApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	appGroupID := c.Query("app_group_id")
	offset := (page - 1) * limit

	var applications []models.Application
	var total int64

	query := h.db.Model(&models.Application{})
	if appGroupID != "" {
		query = query.Where("app_group_id = ?", appGroupID)
	}

	// Gestion limitée à ses groupes
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	if !middleware.IsUnrestricted(c) && len(managedGroupIDs) > 0 {
		// (même logique que le dashboard pour la cohérence)
		if len(managedGroupIDs) > 0 {
			// Applications des AppGroups accessibles à ses groupes
			query = query.Joins("JOIN app_groups ON applications.app_group_id = app_groups.id").
				Where(`
					(
						-- Applications dans des AppGroups publics
						(app_groups.is_private = false)
					)
					OR
					(
						-- Applications des AppGroups privés de ses groupes
						app_groups.is_private = true AND app_groups.owner_group_id IN ?
					)
				`, managedGroupIDs)
		} else {
			// Aucun groupe : aucune application
			query = query.Where("1 = 0")
		}
	}

	query.Count(&total)
	if err := query.Preload("AppGroup").Preload("CreatedBy").
		Order("\"order\" ASC, name ASC").
		Limit(limit).Offset(offset).
		Find(&applications).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des applications")
		return
	}

	totalPages := int(total)/limit + 1
	if int(total)%limit == 0 && total > 0 {
		totalPages = int(total) / limit
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Data:       applications,
		Total:      total,
		Page:       page,
		PageSize:   limit,
		TotalPages: totalPages,
	})
}

// @Summary Créer une application
// @Description Crée une nouvelle application (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param application body models.Application true "Données de l'application"
// @Success 201 {object} models.Application
// @Router /admin/applications [post]
func (h *AdminHandler) CreateApplication(c *gin.Context) {
	var application models.Application
	var flags resourceFlags
	if err := c.ShouldBindBodyWith(&application, binding.JSON); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	_ = c.ShouldBindBodyWith(&flags, binding.JSON)

	note, noteUntil, err := models.NormalizeAdminNote(application.AdminNote, application.AdminNoteUntil)
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, err.Error())
		return
	}
	application.AdminNote, application.AdminNoteUntil = note, noteUntil

	// Auto-renseigner le créateur si non fourni
	if application.CreatedByID == nil {
		if userIDVal, exists := c.Get("user_id"); exists {
			uid := userIDVal.(uint)
			application.CreatedByID = &uid
		}
	}

	// Gestion limitée à ses groupes : seulement dans un AppGroup privé de ses groupes
	if !middleware.CanManageAppGroupWithDB(c, application.AppGroupID, h.db) {
		apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez pas créer de ressource dans cette catégorie : seules les catégories privées de vos groupes sont permises.")
		return
	}

	if err := h.db.Create(&application).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création de l'application")
		return
	}

	// Create ignore les false des champs à valeur par défaut : une application créée inactive le reste
	if err := h.applyFlags(&application, flags, "is_active", "open_in_new_tab"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création de l'application")
		return
	}

	// Charger la relation AppGroup
	h.db.Preload("AppGroup").First(&application, application.ID)

	// Envoyer une notification email si l'application est active
	if application.IsActive {
		utils.SafeGo("email-application", func() {
			emailService := services.NewEmailService(h.db, config.LoadConfig())
			// Récupérer les groupes liés à l'AppGroup
			var targetGroupIDs []uint
			h.db.Table("group_app_groups").Where("app_group_id = ?", application.AppGroupID).Pluck("group_id", &targetGroupIDs)
			if err := emailService.SendNotification("application", application.ID, targetGroupIDs); err != nil {
				log.Printf("[Email] Échec de l'envoi de la notification application: %v", err)
			}
		})
	}

	c.JSON(http.StatusCreated, application)
}

// @Summary Modifier une application
// @Description Modifie une application existante (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de l'application"
// @Param application body models.Application true "Données modifiées"
// @Success 200 {object} models.Application
// @Router /admin/applications/{id} [put]
func (h *AdminHandler) UpdateApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var application models.Application
	if err := h.db.Preload("AppGroup").First(&application, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Application non trouvée")
		return
	}

	// Gestion limitée à ses groupes : seulement une application d'un AppGroup privé de ses groupes
	if application.AppGroup == nil {
		apierror.Respond(c, http.StatusInternalServerError, "AppGroup introuvable pour cette application")
		return
	}

	if !middleware.CanManageAppGroupWithDB(c, application.AppGroup.ID, h.db) {
		apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez pas modifier cette application. Seules les applications des AppGroups privés appartenant à votre groupe sont autorisées.")
		return
	}

	var updateData models.Application
	var flags resourceFlags
	if err := c.ShouldBindBodyWith(&updateData, binding.JSON); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	_ = c.ShouldBindBodyWith(&flags, binding.JSON)

	// Updates() ignore les valeurs vides : la note est écrite à part, seulement si la clé
	// admin_note est présente, pour pouvoir l'effacer sans toucher aux clients qui l'omettent.
	var notePresence struct {
		AdminNote *string `json:"admin_note"`
	}
	_ = c.ShouldBindBodyWith(&notePresence, binding.JSON)
	var noteFields map[string]interface{}
	if notePresence.AdminNote != nil {
		note, noteUntil, err := models.NormalizeAdminNote(*notePresence.AdminNote, updateData.AdminNoteUntil)
		if err != nil {
			apierror.Respond(c, http.StatusBadRequest, err.Error())
			return
		}
		noteFields = map[string]interface{}{"admin_note": note, "admin_note_until": noteUntil}
	}
	updateData.AdminNote, updateData.AdminNoteUntil = "", nil

	// Déplacer l'application exige aussi le droit de gérer la catégorie d'arrivée
	if updateData.AppGroupID != 0 && updateData.AppGroupID != application.AppGroupID &&
		!middleware.CanManageAppGroupWithDB(c, updateData.AppGroupID, h.db) {
		apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez pas déplacer cette application dans cette catégorie.")
		return
	}

	// Updates(struct) ignore false, 0 et "" : cases à cocher, ordre et note sont appliqués à part.
	// Omit(clause.Associations) : sinon GORM réenregistre l'AppGroup préchargé et remet
	// app_group_id sur l'ancienne catégorie, ce qui annule un déplacement.
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&application).Omit(clause.Associations).Updates(updateData).Error; err != nil {
			return err
		}
		if values := flagUpdates(flags, "is_active", "open_in_new_tab", "order", "description"); len(values) > 0 {
			if err := tx.Model(&application).Omit(clause.Associations).Updates(values).Error; err != nil {
				return err
			}
		}
		if noteFields != nil {
			return tx.Model(&application).Omit(clause.Associations).Updates(noteFields).Error
		}
		return nil
	})
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la modification")
		return
	}

	// Mettre à jour CreatedByID séparément (Updates() ignore les pointeurs nil)
	if updateData.CreatedByID != nil {
		h.db.Model(&application).Update("created_by_id", updateData.CreatedByID)
	}

	// Charger les relations
	h.db.Preload("AppGroup").Preload("CreatedBy").First(&application, application.ID)

	c.JSON(http.StatusOK, application)
}

// @Summary Supprimer une application
// @Description Supprime une application (admin uniquement)
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "ID de l'application"
// @Success 200 {object} models.SuccessResponse
// @Router /admin/applications/{id} [delete]
func (h *AdminHandler) DeleteApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	// Gestion limitée à ses groupes
	managedGroupIDs := middleware.GetManagedGroupIDs(c)
	if !middleware.IsUnrestricted(c) && len(managedGroupIDs) > 0 {
		var application models.Application
		if err := h.db.Preload("AppGroup").First(&application, id).Error; err != nil {
			apierror.Respond(c, http.StatusNotFound, "Application non trouvée")
			return
		}

		// L'application doit appartenir à un AppGroup privé de ses groupes
		if application.AppGroup == nil {
			apierror.Respond(c, http.StatusInternalServerError, "AppGroup introuvable pour cette application")
			return
		}

		if !middleware.CanManageAppGroupWithDB(c, application.AppGroup.ID, h.db) {
			apierror.Respond(c, http.StatusForbidden, "Vous ne pouvez pas supprimer cette application. Seules les applications des AppGroups privés appartenant à votre groupe sont autorisées.")
			return
		}
	}

	if err := h.db.Delete(&models.Application{}, id).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Application supprimée avec succès",
	})
}

// ============ GESTION DES UTILISATEURS ============

// @Summary Lister les utilisateurs
// @Description Liste paginée des utilisateurs, avec recherche, filtres et tri (admin uniquement)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Numéro de page" default(1)
// @Param page_size query int false "Taille de page (max 500)" default(25)
// @Param search query string false "Nom d'utilisateur, e-mail, prénom ou nom"
// @Param role query string false "admin, editor ou user"
// @Param status query string false "active ou inactive"
// @Param sort query string false "username, email, role, last_login, is_active ou created_at"
// @Param order query string false "asc ou desc"
// @Success 200 {object} models.PaginatedResponse
// @Router /admin/users [get]
func (h *AdminHandler) GetUsers(c *gin.Context) {
	page := parsePageParams(c, userPageLimits)
	query := filterUsers(h.db.Model(&models.User{}), c.Query("search"), c.Query("role"), c.Query("status"))

	var total int64
	if err := query.Count(&total).Error; err != nil {
		respondUsersError(c)
		return
	}

	var users []models.User
	if err := query.Preload("Groups").Preload("AccessRoles").
		Order(userSortClause(c.Query("sort"), c.Query("order"))).
		Limit(page.Size).Offset(page.Offset()).
		Find(&users).Error; err != nil {
		respondUsersError(c)
		return
	}

	// Masquer les mots de passe
	for i := range users {
		users[i].Password = ""
	}
	attachDirectoryGroups(h.db, users)

	c.JSON(http.StatusOK, page.Response(users, total))
}

// attachDirectoryGroups renseigne les groupes où chaque compte est membre par le trombinoscope
// (l'interface ne propose pas de les retirer à la main).
func attachDirectoryGroups(db *gorm.DB, users []models.User) {
	if len(users) == 0 {
		return
	}
	ids := make([]uint, len(users))
	index := make(map[uint]int, len(users))
	for i, u := range users {
		ids[i] = u.ID
		index[u.ID] = i
	}
	var rows []struct{ UserID, GroupID uint }
	if err := db.Table("user_groups").Select("user_id, group_id").
		Where("user_id IN ? AND source = ?", ids, "directory").Order("group_id").Scan(&rows).Error; err != nil {
		log.Printf("utilisateurs : groupes du trombinoscope illisibles : %v", err)
		return
	}
	for _, r := range rows {
		i := index[r.UserID]
		users[i].DirectoryGroupIDs = append(users[i].DirectoryGroupIDs, r.GroupID)
	}
}

var userPageLimits = pageLimits{Default: 25, Max: 500}

// Colonnes triables de la liste des utilisateurs (liste blanche : la valeur part dans ORDER BY).
var userSortColumns = map[string]string{
	"username":   "username",
	"email":      "email",
	"role":       "role",
	"last_login": "last_login",
	"is_active":  "is_active",
	"created_at": "created_at",
}

// userSortClause construit un ORDER BY sûr ; l'id départage pour une pagination stable.
func userSortClause(sort, order string) string {
	column, ok := userSortColumns[sort]
	if !ok {
		column = "username"
	}
	direction := "ASC"
	if strings.EqualFold(order, "desc") {
		direction = "DESC"
	}
	nulls := ""
	if column == "last_login" {
		nulls = " NULLS LAST"
	}
	return column + " " + direction + nulls + ", id ASC"
}

// filterUsers applique la recherche texte, le rôle et le statut.
func filterUsers(query *gorm.DB, search, role, status string) *gorm.DB {
	if term := strings.TrimSpace(search); term != "" {
		like := "%" + escapeLikePattern(strings.ToLower(term)) + "%"
		query = query.Where(
			"LOWER(username) LIKE ? OR LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			like, like, like, like,
		)
	}
	switch role {
	case "admin", "user":
		query = query.Where("role = ?", role)
	}
	switch status {
	case "active":
		query = query.Where("is_active = ?", true)
	case "inactive":
		query = query.Where("is_active = ?", false)
	}
	return query
}

func respondUsersError(c *gin.Context) {
	apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs")
}

// @Summary Créer un utilisateur
// @Description Crée un nouvel utilisateur (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body models.RegisterRequest true "Données de l'utilisateur"
// @Success 201 {object} models.User
// @Router /admin/users [post]
func (h *AdminHandler) CreateUser(c *gin.Context) {
	var createData struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
		IsActive  bool   `json:"is_active"`
		GroupIDs  []uint `json:"group_ids"`
	}

	if err := c.ShouldBindBodyWith(&createData, binding.JSON); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	if createData.Role == "" {
		createData.Role = "user"
	}
	if !validUserRole(createData.Role) {
		apierror.Respond(c, http.StatusBadRequest, msgInvalidUserRole)
		return
	}

	// Vérifier si l'utilisateur existe déjà
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", createData.Username, createData.Email).First(&existingUser).Error; err == nil {
		apierror.Respond(c, http.StatusConflict, "Nom d'utilisateur ou email déjà utilisé")
		return
	}

	// Hasher le mot de passe avec coût sécurisé (12 minimum - OWASP 2025)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createData.Password), h.bcryptCost)
	if err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors du hashage du mot de passe")
		return
	}

	user := models.User{
		Username:  createData.Username,
		Email:     createData.Email,
		Password:  string(hashedPassword),
		FirstName: createData.FirstName,
		LastName:  createData.LastName,
		Role:      createData.Role,
		IsActive:  createData.IsActive,
	}

	if err := createKeepingFalse(h.db, c, &user, "is_active"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création de l'utilisateur")
		return
	}

	// Associer les groupes si fournis
	if len(createData.GroupIDs) > 0 {
		var groups []models.Group
		if err := h.db.Where("id IN ?", createData.GroupIDs).Find(&groups).Error; err == nil {
			h.db.Model(&user).Association("Groups").Replace(groups)
		}
	}
	// Une fiche du trombinoscope au même e-mail est liée au compte, qui rejoint ses groupes
	services.SyncDirectoryGroupsLogged(h.db, "compte créé")

	user.Password = ""
	h.db.Preload("Groups").First(&user, user.ID)

	c.JSON(http.StatusCreated, user)
}

// @Summary Modifier un utilisateur
// @Description Modifie un utilisateur existant (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de l'utilisateur"
// @Param user body models.User true "Données modifiées"
// @Success 200 {object} models.User
// @Router /admin/users/{id} [put]
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	var updateData struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
		IsActive  *bool  `json:"is_active"`
		Password  string `json:"password,omitempty"`
		GroupIDs  []uint `json:"group_ids"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	if updateData.Role != "" && !validUserRole(updateData.Role) {
		apierror.Respond(c, http.StatusBadRequest, msgInvalidUserRole)
		return
	}

	// Mise à jour des champs
	if updateData.Username != "" {
		user.Username = updateData.Username
	}
	if updateData.Email != "" {
		user.Email = updateData.Email
	}
	if updateData.FirstName != "" {
		user.FirstName = updateData.FirstName
	}
	if updateData.LastName != "" {
		user.LastName = updateData.LastName
	}
	if updateData.Role != "" {
		user.Role = updateData.Role
	}
	if updateData.IsActive != nil {
		user.IsActive = *updateData.IsActive
	}

	// Hash du nouveau mot de passe si fourni avec coût sécurisé (12 minimum - OWASP 2025)
	if updateData.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updateData.Password), h.bcryptCost)
		if err != nil {
			apierror.Respond(c, http.StatusInternalServerError, "Erreur lors du hashage du mot de passe")
			return
		}
		user.Password = string(hashedPassword)
	}

	if err := h.db.Save(&user).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la modification")
		return
	}

	// Mise à jour des groupes si fournis ; les groupes venus du trombinoscope restent (la
	// synchronisation les rétablit tant que la personne est dans l'entité)
	if updateData.GroupIDs != nil {
		var groups []models.Group
		if err := h.db.Where("id IN ?", updateData.GroupIDs).Find(&groups).Error; err == nil {
			h.db.Model(&user).Association("Groups").Replace(groups)
		}
	}
	services.SyncDirectoryGroupsLogged(h.db, "compte modifié")

	// Masquer le mot de passe
	user.Password = ""

	// Charger les relations
	h.db.Preload("Groups").First(&user, user.ID)

	c.JSON(http.StatusOK, user)
}

// @Summary Supprimer un utilisateur
// @Description Supprime un utilisateur (admin uniquement)
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {object} models.SuccessResponse
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	// Vérifier que l'utilisateur existe
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Supprimer les associations avec les groupes
	h.db.Model(&user).Association("Groups").Clear()

	// Supprimer l'utilisateur
	if err := h.db.Delete(&user).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Utilisateur supprimé avec succès",
	})
}

// GetDeletedUsers récupère tous les utilisateurs supprimés (soft deleted)
func (h *AdminHandler) GetDeletedUsers(c *gin.Context) {
	var users []models.User

	// Utiliser Unscoped() pour voir les enregistrements soft deleted
	if err := h.db.Unscoped().Where("deleted_at IS NOT NULL").Preload("Groups").Find(&users).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des utilisateurs supprimés")
		return
	}

	// Masquer les mots de passe
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

// RestoreUser restaure un utilisateur supprimé
func (h *AdminHandler) RestoreUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	// Trouver l'utilisateur supprimé
	var user models.User
	if err := h.db.Unscoped().First(&user, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier qu'il est bien supprimé
	if user.DeletedAt.Time.IsZero() {
		apierror.Respond(c, http.StatusBadRequest, "Cet utilisateur n'est pas supprimé")
		return
	}

	// Restaurer l'utilisateur
	user.DeletedAt = gorm.DeletedAt{}
	if err := h.db.Unscoped().Save(&user).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la restauration")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Utilisateur restauré avec succès",
		Data:    user,
	})
}

// PermanentlyDeleteUser supprime définitivement un utilisateur
func (h *AdminHandler) PermanentlyDeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	// Trouver l'utilisateur supprimé
	var user models.User
	if err := h.db.Unscoped().First(&user, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	// Vérifier qu'il est bien supprimé
	if user.DeletedAt.Time.IsZero() {
		apierror.Respond(c, http.StatusBadRequest, "Cet utilisateur doit d'abord être supprimé")
		return
	}

	// Supprimer définitivement dans une transaction
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Supprimer les associations many-to-many directement dans les tables de jointure :
		// Association(...).Clear() en mode Unscoped produit un SQL invalide sur user_favorites.
		txUnscoped := tx.Unscoped()
		for _, joinTable := range []string{"user_groups", "user_favorites", "user_access_roles"} {
			// joinTable vient de la liste constante ci-dessus, jamais d'une entrée utilisateur
			if err := tx.Exec("DELETE FROM "+joinTable+" WHERE user_id = ?", user.ID).Error; err != nil {
				return err
			}
		}

		// 2. Supprimer les enregistrements liés à l'utilisateur
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.NewsReaction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.NewsRead{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.ApplicationClick{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.Notification{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("sender_id = ? OR recipient_id = ?", user.ID, user.ID).Delete(&models.ChatMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.Comment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Feedback{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.PollVote{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.GamificationProfile{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.UserAchievement{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.XPTransaction{}).Error; err != nil {
			return err
		}

		// 3. Nullifier les références d'auteur sur le contenu (préserver les articles/sondages)
		if err := tx.Model(&models.News{}).Where("author_id = ?", user.ID).Update("author_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Poll{}).Where("author_id = ?", user.ID).Update("author_id", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("author_id = ?", user.ID).Delete(&models.Event{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.Media{}).Where("uploaded_by = ?", user.ID).Update("uploaded_by", nil).Error; err != nil {
			return err
		}
		// Nullifier moderated_by dans les commentaires
		if err := tx.Model(&models.Comment{}).Where("moderated_by = ?", user.ID).Update("moderated_by", nil).Error; err != nil {
			return err
		}

		// 4. Supprimer définitivement l'utilisateur
		if err := txUnscoped.Delete(&user).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		log.Printf("[PermanentlyDeleteUser] utilisateur %d : %v", user.ID, err)
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression définitive")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Utilisateur supprimé définitivement",
	})
}

// @Summary Assigner un utilisateur à des groupes
// @Description Assigne un utilisateur à des groupes d'utilisateurs (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID de l'utilisateur"
// @Param groups body []int true "IDs des groupes"
// @Success 200 {object} models.User
// @Router /admin/users/{id}/groups [put]
func (h *AdminHandler) AssignUserToGroups(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Utilisateur non trouvé")
		return
	}

	var groupIDs []uint
	if err := c.ShouldBindJSON(&groupIDs); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "IDs de groupes invalides")
		return
	}

	// Récupérer les groupes
	var groups []models.Group
	if err := h.db.Where("id IN ?", groupIDs).Find(&groups).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des groupes")
		return
	}

	// Remplacer les associations (les groupes venus du trombinoscope sont rétablis ensuite)
	if err := h.db.Model(&user).Association("Groups").Replace(groups); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de l'assignation")
		return
	}
	services.SyncDirectoryGroupsLogged(h.db, "groupes du compte")

	// Recharger avec les relations
	h.db.Preload("Groups").First(&user, user.ID)
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// @Summary Lister les groupes
// @Description Récupère tous les groupes d'utilisateurs (admin uniquement)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Group
// @Router /admin/groups [get]
func (h *AdminHandler) GetGroups(c *gin.Context) {
	var groups []models.Group
	if err := h.db.Preload("Users").Preload("AppGroups").Find(&groups).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des groupes")
		return
	}
	// Le groupe par défaut ne se choisit pas pour un rôle limité à des groupes
	if everyone, err := models.EveryoneGroupID(h.db); err != nil {
		log.Printf("groupes : groupe par défaut illisible : %v", err)
	} else {
		for i := range groups {
			groups[i].IsDefault = groups[i].ID == everyone
		}
	}
	attachOrgUnits(h.db, groups)

	c.JSON(http.StatusOK, groups)
}

// @Summary Créer un groupe
// @Description Crée un nouveau groupe d'utilisateurs (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param group body models.Group true "Données du groupe"
// @Success 201 {object} models.Group
// @Router /admin/groups [post]
func (h *AdminHandler) CreateGroup(c *gin.Context) {
	var createData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		IsActive    bool   `json:"is_active"`
		ParentID    *uint  `json:"parent_id"`
		AppGroupIDs []uint `json:"app_group_ids"`
	}

	if err := c.ShouldBindBodyWith(&createData, binding.JSON); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	if createData.ParentID != nil && !h.groupExists(*createData.ParentID) {
		apierror.Respond(c, http.StatusBadRequest, "Groupe parent introuvable")
		return
	}

	group := models.Group{
		Name:        createData.Name,
		Description: createData.Description,
		Color:       createData.Color,
		IsActive:    createData.IsActive,
		ParentID:    createData.ParentID,
	}

	if err := createKeepingFalse(h.db, c, &group, "is_active"); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la création du groupe")
		return
	}

	// Associer les groupes d'applications si fournis
	if len(createData.AppGroupIDs) > 0 {
		var appGroups []models.AppGroup
		if err := h.db.Where("id IN ?", createData.AppGroupIDs).Find(&appGroups).Error; err == nil {
			h.db.Model(&group).Association("AppGroups").Replace(appGroups)
		}
	}

	// Charger les relations
	h.db.Preload("Users").Preload("AppGroups").First(&group, group.ID)

	c.JSON(http.StatusCreated, group)
}

// @Summary Modifier un groupe
// @Description Modifie un groupe d'utilisateurs existant (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID du groupe"
// @Param group body models.Group true "Données modifiées"
// @Success 200 {object} models.Group
// @Router /admin/groups/{id} [put]
func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var group models.Group
	if err := h.db.First(&group, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Groupe non trouvé")
		return
	}

	var updateData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		IsActive    *bool  `json:"is_active"`
		// Présent : nouveau parent (null = groupe racine) ; absent : inchangé
		ParentID    json.RawMessage `json:"parent_id"`
		AppGroupIDs []uint          `json:"app_group_ids"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "Données invalides")
		return
	}
	if len(updateData.ParentID) > 0 {
		parentID, msg := h.validParent(group.ID, updateData.ParentID)
		if msg != "" {
			apierror.Respond(c, http.StatusBadRequest, msg)
			return
		}
		group.ParentID = parentID
	}

	// Mise à jour des champs
	if updateData.Name != "" {
		group.Name = updateData.Name
	}
	if updateData.Description != "" {
		group.Description = updateData.Description
	}
	if updateData.Color != "" {
		group.Color = updateData.Color
	}
	if updateData.IsActive != nil {
		group.IsActive = *updateData.IsActive
	}

	if err := h.db.Save(&group).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la modification")
		return
	}

	// Mise à jour des groupes d'applications si fournis
	if updateData.AppGroupIDs != nil {
		var appGroups []models.AppGroup
		// Si le tableau n'est pas vide, chercher les groupes
		if len(updateData.AppGroupIDs) > 0 {
			if err := h.db.Where("id IN ?", updateData.AppGroupIDs).Find(&appGroups).Error; err == nil {
				h.db.Model(&group).Association("AppGroups").Replace(appGroups)
			}
		} else {
			// Tableau vide signifie supprimer toutes les associations
			h.db.Model(&group).Association("AppGroups").Clear()
		}
	}

	// Charger les relations
	h.db.Preload("Users").Preload("AppGroups").First(&group, group.ID)

	c.JSON(http.StatusOK, group)
}

// @Summary Supprimer un groupe
// @Description Supprime un groupe d'utilisateurs (admin uniquement)
// @Tags Admin
// @Security BearerAuth
// @Param id path int true "ID du groupe"
// @Success 200 {object} models.SuccessResponse
// @Router /admin/groups/{id} [delete]
func (h *AdminHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var group models.Group
	if err := h.db.First(&group, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Groupe non trouvé")
		return
	}

	// Supprimer les associations ; les sous-groupes remontent d'un niveau et l'entité du
	// trombinoscope liée n'a plus de groupe
	h.db.Model(&group).Association("Users").Clear()
	h.db.Model(&group).Association("AppGroups").Clear()
	if err := h.db.Model(&models.Group{}).Where("parent_id = ?", group.ID).Update("parent_id", group.ParentID).Error; err != nil {
		log.Printf("groupes : sous-groupes de %d non rattachés : %v", group.ID, err)
	}
	if err := h.db.Model(&models.OrgUnit{}).Where("group_id = ?", group.ID).Update("group_id", nil).Error; err != nil {
		log.Printf("groupes : entité liée à %d non déliée : %v", group.ID, err)
	}

	// Supprimer le groupe
	if err := h.db.Delete(&group).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la suppression")
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Groupe supprimé avec succès",
	})
}

// @Summary Assigner des groupes d'applications à un groupe d'utilisateurs
// @Description Définit les permissions d'un groupe d'utilisateurs (admin uniquement)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID du groupe d'utilisateurs"
// @Param app_groups body []int true "IDs des groupes d'applications"
// @Success 200 {object} models.Group
// @Router /admin/groups/{id}/app-groups [put]
func (h *AdminHandler) AssignGroupToAppGroups(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		apierror.Respond(c, http.StatusBadRequest, "ID invalide")
		return
	}

	var group models.Group
	if err := h.db.First(&group, id).Error; err != nil {
		apierror.Respond(c, http.StatusNotFound, "Groupe non trouvé")
		return
	}

	var appGroupIDs []uint
	if err := c.ShouldBindJSON(&appGroupIDs); err != nil {
		apierror.Respond(c, http.StatusBadRequest, "IDs de groupes d'applications invalides")
		return
	}

	// Récupérer les groupes d'applications
	var appGroups []models.AppGroup
	if err := h.db.Where("id IN ?", appGroupIDs).Find(&appGroups).Error; err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de la récupération des groupes d'applications")
		return
	}

	// Remplacer les associations
	if err := h.db.Model(&group).Association("AppGroups").Replace(appGroups); err != nil {
		apierror.Respond(c, http.StatusInternalServerError, "Erreur lors de l'assignation")
		return
	}

	// Recharger avec les relations
	h.db.Preload("Users").Preload("AppGroups").First(&group, group.ID)

	c.JSON(http.StatusOK, group)
}

// ============ GESTION DE LA BASE DE DONNÉES ============

// @Summary Réinitialiser la base de données
// @Description Supprime toutes les données et recrée les données initiales (admin uniquement - ATTENTION: opération destructive)
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} models.SuccessResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /admin/database/reset [post]
func (h *AdminHandler) ResetDatabase(c *gin.Context) {
	// Liste de toutes les tables à réinitialiser (dans l'ordre de dépendances)
	tables := []string{
		// Tables de jointure many-to-many
		"user_favorites",
		"user_groups",
		"group_app_groups",
		"news_tags",

		// Tables dépendantes (analytics, reactions, etc.)
		"application_clicks",
		"news_reactions",
		"news_reads",
		"poll_votes",
		"comments",
		"feedbacks",
		"notifications",
		"suggestions",
		"suggestion_votes",
		"email_notification_logs",

		// Tables avec relations
		"poll_options",
		"polls",
		"suggestion_categories",
		"events",
		"event_categories",
		"news",
		"news_categories",
		"tags",
		"announcements",
		"applications",
		"app_groups",
		"media",
		"comment_settings",
		"email_templates",
		"smtp_configs",
		"email_oauth_configs",
		"oauth_providers",

		// Tables principales
		"users",
		"groups",
		"app_settings",
	}

	// Utiliser TRUNCATE CASCADE pour supprimer toutes les données et gérer automatiquement les dépendances
	// TRUNCATE est plus rapide que DELETE et réinitialise automatiquement les séquences
	for _, table := range tables {
		query := "TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE"
		if err := h.db.Exec(query).Error; err != nil {
			// Ignorer les erreurs si la table n'existe pas
			// (peut arriver si certaines fonctionnalités ne sont pas encore migrées)
			continue
		}
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Base de données réinitialisée avec succès. Veuillez redémarrer le serveur pour recréer les données initiales.",
		Data: map[string]interface{}{
			"note": "Les données de démonstration seront recréées au prochain démarrage du serveur",
		},
	})
}

// msgInvalidUserRole : le rôle d'un compte est « admin » ou « user » ; les autres droits
// passent par les rôles créés par l'administrateur (permissions).
const msgInvalidUserRole = "Rôle invalide : admin ou user"

// validUserRole : rôle de compte accepté.
func validUserRole(role string) bool {
	return role == "admin" || role == "user"
}

// groupExists : groupe existant et non supprimé.
func (h *AdminHandler) groupExists(id uint) bool {
	var n int64
	h.db.Model(&models.Group{}).Where("id = ?", id).Count(&n)
	return n > 0
}

// validParent lit le parent demandé pour groupID (null : groupe racine) et refuse un parent
// inconnu ou une boucle. Renvoie un message d'erreur si la demande est refusée.
func (h *AdminHandler) validParent(groupID uint, raw json.RawMessage) (*uint, string) {
	if string(raw) == "null" {
		return nil, ""
	}
	var parentID uint
	if err := json.Unmarshal(raw, &parentID); err != nil || parentID == 0 {
		return nil, "Groupe parent invalide"
	}
	if !h.groupExists(parentID) {
		return nil, "Groupe parent introuvable"
	}
	cycle, err := models.WouldCreateCycle(h.db, groupID, parentID)
	if err != nil {
		log.Printf("groupes : contrôle de boucle %d → %d : %v", groupID, parentID, err)
		return nil, "Groupe parent invalide"
	}
	if cycle {
		return nil, "Un groupe ne peut pas être placé sous lui-même ni sous l'un de ses sous-groupes"
	}
	return &parentID, ""
}

// attachOrgUnits renseigne l'entité du trombinoscope liée à chaque groupe.
func attachOrgUnits(db *gorm.DB, groups []models.Group) {
	if len(groups) == 0 {
		return
	}
	var units []models.OrgUnit
	if err := db.Select("id, name, group_id").Where("group_id IS NOT NULL").Find(&units).Error; err != nil {
		log.Printf("groupes : entités liées illisibles : %v", err)
		return
	}
	byGroup := make(map[uint]models.OrgUnit, len(units))
	for _, u := range units {
		byGroup[*u.GroupID] = u
	}
	for i := range groups {
		if u, ok := byGroup[groups[i].ID]; ok {
			id := u.ID
			groups[i].OrgUnitID = &id
			groups[i].OrgUnitName = u.Name
		}
	}
}
