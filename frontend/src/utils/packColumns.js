// Disposition « maçonnerie » : chaque bloc va dans la colonne la plus courte, dans l'ordre
// d'origine, pour qu'un bloc court ne laisse pas un grand vide à côté d'un bloc long.

/**
 * Nombre de colonnes d'au moins `minWidth` px qui tiennent dans `width` px.
 * @param {number} width largeur disponible
 * @param {number} minWidth largeur minimale d'une colonne
 * @param {number} gap écart entre colonnes
 * @returns {number} au moins 1
 */
export function columnCountFor (width, minWidth, gap) {
  return Math.max(1, Math.floor((width + gap) / (minWidth + gap)))
}

/**
 * Répartit des blocs en colonnes : chacun dans la colonne la moins haute ; à égalité, la
 * première (ordre de lecture). Les hauteurs inconnues passent par `estimate`.
 * @template K
 * @param {K[]} ids blocs dans l'ordre d'affichage voulu
 * @param {Map<K, number>} heights hauteurs mesurées, en px
 * @param {number} count nombre de colonnes
 * @param {number} [gap=0] écart vertical entre deux blocs, en px
 * @param {(id: K) => number} [estimate] hauteur supposée d'un bloc pas encore mesuré
 * @returns {K[][]} une liste d'identifiants par colonne
 */
export function packColumns (ids, heights, count, gap = 0, estimate = () => 0) {
  const columns = Array.from({ length: Math.max(1, count) }, () => [])
  const totals = columns.map(() => 0)
  for (const id of ids) {
    let target = 0
    for (let i = 1; i < totals.length; i++) {
      if (totals[i] < totals[target]) target = i
    }
    const height = heights.has(id) ? heights.get(id) : estimate(id)
    totals[target] += (columns[target].length ? gap : 0) + height
    columns[target].push(id)
  }
  return columns
}
