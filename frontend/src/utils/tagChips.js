// Étiquettes d'actualités : combien en montrer sur une carte et dans la barre de filtres,
// pour qu'elles restent discrètes au lieu d'occuper plusieurs lignes.

/**
 * Sépare les étiquettes affichées de celles résumées par « +N ».
 * @template T
 * @param {T[] | undefined | null} tags
 * @param {number} max nombre d'étiquettes affichées
 * @returns {{ shown: T[], rest: T[] }}
 */
export function splitTags (tags, max) {
  const list = Array.isArray(tags) ? tags : []
  return { shown: list.slice(0, max), rest: list.slice(max) }
}

/**
 * Étiquettes de la barre de filtres. Repliée : les `limit` premières, plus celles déjà
 * choisies (un filtre actif ne doit jamais disparaître). Dépliée : toutes.
 * @template {{ id: number }} T
 * @param {T[]} tags
 * @param {number[]} selectedIds
 * @param {number} limit
 * @param {boolean} expanded
 * @returns {T[]}
 */
export function filterTagsToShow (tags, selectedIds, limit, expanded) {
  if (expanded) return tags
  return tags.filter((tag, index) => index < limit || selectedIds.includes(tag.id))
}
