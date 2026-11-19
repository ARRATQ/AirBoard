package utils

import "strings"

// SafeOrderClause builds an ORDER BY clause from untrusted query parameters.
// The column must be in allowed, otherwise defaultSort is used; the direction
// is normalised to ASC or DESC. Never pass raw user input to GORM's Order().
func SafeOrderClause(sortBy, order string, allowed []string, defaultSort, defaultOrder string) string {
	column := defaultSort
	for _, a := range allowed {
		if sortBy == a {
			column = a
			break
		}
	}

	direction := normaliseDirection(order)
	if direction == "" {
		direction = normaliseDirection(defaultOrder)
	}
	if direction == "" {
		direction = "ASC"
	}

	return column + " " + direction
}

func normaliseDirection(order string) string {
	switch strings.ToLower(strings.TrimSpace(order)) {
	case "asc":
		return "ASC"
	case "desc":
		return "DESC"
	default:
		return ""
	}
}
