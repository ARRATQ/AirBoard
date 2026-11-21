package handlers

import (
	"strconv"

	"airboard/models"

	"github.com/gin-gonic/gin"
)

// pageLimits fixe la taille de page par défaut et le plafond d'une liste.
type pageLimits struct {
	Default int
	Max     int
}

// pageParams est une page demandée, déjà bornée.
type pageParams struct {
	Page int
	Size int
}

// parsePageParams lit `page` et `page_size` (ou `limit`, nom utilisé par les listes admin).
// Les valeurs absentes ou invalides prennent la valeur par défaut ; la taille est plafonnée.
func parsePageParams(c *gin.Context, limits pageLimits) pageParams {
	page := positiveInt(c.Query("page"), 1)
	sizeText := c.Query("page_size")
	if sizeText == "" {
		sizeText = c.Query("limit")
	}
	size := positiveInt(sizeText, limits.Default)
	if size > limits.Max {
		size = limits.Max
	}
	return pageParams{Page: page, Size: size}
}

func positiveInt(text string, fallback int) int {
	n, err := strconv.Atoi(text)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}

// Offset est le nombre de lignes à sauter.
func (p pageParams) Offset() int {
	return (p.Page - 1) * p.Size
}

// TotalPages arrondit au supérieur ; 0 quand la liste est vide.
func (p pageParams) TotalPages(total int64) int {
	return int((total + int64(p.Size) - 1) / int64(p.Size))
}

// Response enveloppe une page au format commun des listes paginées.
func (p pageParams) Response(data interface{}, total int64) models.PaginatedResponse {
	return models.PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.Size,
		TotalPages: p.TotalPages(total),
	}
}
