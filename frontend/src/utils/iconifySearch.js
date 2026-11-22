// Recherche d'icônes dans le catalogue Iconify (api.iconify.design), pour le sélecteur d'icônes
// de l'administration. Seule la requête de recherche et le chargement des icônes choisies
// sortent vers Iconify ; aucune donnée d'utilisateur n'y part.

export const ICONIFY_API = 'https://api.iconify.design'

/** Nombre de résultats demandés (l'API accepte 32 à 999). */
export const SEARCH_LIMIT = 64

/**
 * Familles de style proposées au filtre. « line » : trait fin, comme les étiquettes ;
 * « solid » : pictogrammes pleins. Vide : toutes les collections.
 */
export const ICON_STYLES = {
  all: [],
  line: ['lucide', 'tabler', 'iconoir'],
  solid: ['mdi', 'material-symbols', 'ic'],
}

const ICON_NAME_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*:[a-z0-9]+(?:[-_][a-z0-9]+)*$/

/**
 * Nom d'icône Iconify valide, au format « collection:nom ».
 * @param {unknown} value
 * @returns {boolean}
 */
export function isIconName(value) {
  return typeof value === 'string' && ICON_NAME_RE.test(value)
}

/**
 * Adresse de recherche pour une requête et un style.
 * @param {string} query texte saisi
 * @param {keyof typeof ICON_STYLES} [style]
 * @returns {string | null} null si la requête est vide
 */
export function buildSearchUrl(query, style = 'all') {
  const q = (query || '').trim()
  if (!q) return null
  const params = new URLSearchParams({ query: q.slice(0, 64), limit: String(SEARCH_LIMIT) })
  const prefixes = ICON_STYLES[style] || []
  if (prefixes.length) params.set('prefixes', prefixes.join(','))
  return `${ICONIFY_API}/search?${params}`
}

/**
 * Noms d'icônes valides et uniques d'une réponse de recherche Iconify.
 * @param {unknown} data réponse JSON
 * @returns {string[]}
 */
export function parseSearchResponse(data) {
  const icons = data && Array.isArray(data.icons) ? data.icons : []
  return [...new Set(icons.filter(isIconName))]
}

/**
 * Cherche des icônes dans Iconify.
 * @param {string} query
 * @param {keyof typeof ICON_STYLES} [style]
 * @param {{ signal?: AbortSignal, fetchImpl?: typeof fetch }} [options]
 * @returns {Promise<string[]>}
 */
export async function searchIcons(query, style = 'all', { signal, fetchImpl = fetch } = {}) {
  const url = buildSearchUrl(query, style)
  if (!url) return []
  const response = await fetchImpl(url, { signal })
  if (!response.ok) throw new Error(`Iconify search failed: ${response.status}`)
  return parseSearchResponse(await response.json())
}
