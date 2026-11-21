package handlers

import (
	"strconv"
	"unicode/utf8"
)

// maxListSearchLength est la longueur maximale, en caractères, d'une recherche dans une liste.
const maxListSearchLength = 100

// resourceFlags lit dans le corps JSON les champs que Updates(struct) de GORM ignore
// quand ils valent false ou 0. Un pointeur nil signifie « champ absent de la requête ».
type resourceFlags struct {
	IsActive     *bool `json:"is_active"`
	IsPrivate    *bool `json:"is_private"`
	OpenInNewTab *bool `json:"open_in_new_tab"`
	Order        *int  `json:"order"`
}

// flagUpdates renvoie, pour les colonnes demandées, les valeurs présentes dans la requête.
func flagUpdates(f resourceFlags, columns ...string) map[string]interface{} {
	values := map[string]interface{}{}
	for _, column := range columns {
		switch column {
		case "is_active":
			if f.IsActive != nil {
				values[column] = *f.IsActive
			}
		case "is_private":
			if f.IsPrivate != nil {
				values[column] = *f.IsPrivate
			}
		case "open_in_new_tab":
			if f.OpenInNewTab != nil {
				values[column] = *f.OpenInNewTab
			}
		case "order":
			if f.Order != nil {
				values[column] = *f.Order
			}
		}
	}
	return values
}

// applyFlags enregistre les valeurs présentes dans la requête pour les colonnes données,
// y compris false et 0, que Create et Updates(struct) laissent de côté.
func (h *AdminHandler) applyFlags(model interface{}, f resourceFlags, columns ...string) error {
	values := flagUpdates(f, columns...)
	if len(values) == 0 {
		return nil
	}
	return h.db.Model(model).Updates(values).Error
}

// resolveOwnerGroup choisit le groupe propriétaire d'une catégorie créée par un admin de groupe :
// le groupe demandé s'il l'administre, sinon son premier groupe quand aucun n'est demandé.
// Renvoie false si le groupe demandé ne fait pas partie de ses groupes administrés.
func resolveOwnerGroup(requested *uint, managed []uint) (uint, bool) {
	if len(managed) == 0 {
		return 0, false
	}
	if requested == nil || *requested == 0 {
		return managed[0], true
	}
	if containsUint(managed, *requested) {
		return *requested, true
	}
	return 0, false
}

func containsUint(values []uint, target uint) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// listSearchTooLong indique si une recherche de liste dépasse la longueur permise,
// comptée en caractères et non en octets (l'arabe compte deux octets par lettre).
func listSearchTooLong(search string) bool {
	return utf8.RuneCountInString(search) > maxListSearchLength
}

// parseBoolFilter lit un filtre booléen de l'URL ("true" ou "false") ; ok vaut false s'il est absent ou invalide.
func parseBoolFilter(raw string) (value bool, ok bool) {
	if raw == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(raw)
	if err != nil {
		return false, false
	}
	return parsed, true
}
