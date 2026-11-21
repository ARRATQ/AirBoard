// Icône saisie en base dans une collection non embarquée (autre que mdi, lucide, heroicons) :
// plutôt qu'un vide, on affiche un pictogramme neutre, sans aucune requête vers api.iconify.design.

/** Corps SVG de mdi:shape-outline (24 × 24). */
export const PLACEHOLDER_BODY =
  '<path fill="currentColor" d="M11 13.5v8H3v-8zm-2 2H5v4h4zM12 2l5.5 9h-11zm0 3.86L10.08 9h3.84zM17.5 13c2.5 0 4.5 2 4.5 4.5S20 22 17.5 22S13 20 13 17.5s2-4.5 4.5-4.5m0 2a2.5 2.5 0 0 0-2.5 2.5a2.5 2.5 0 0 0 2.5 2.5a2.5 2.5 0 0 0 2.5-2.5a2.5 2.5 0 0 0-2.5-2.5"/>'

/**
 * Réponse au format de l'API Iconify pour l'adresse que @iconify/vue voulait interroger.
 * @param {string} url ex. https://api.iconify.design/fluent-emoji-flat.json?icons=robot
 * @returns {{prefix: string, icons: Record<string, {body: string}>, width: number, height: number} | null}
 */
export function placeholderIconSet (url) {
  let parsed
  try { parsed = new URL(url) } catch { return null }
  const match = parsed.pathname.match(/\/([a-z0-9-]+)\.json$/)
  const names = (parsed.searchParams.get('icons') || '').split(',').filter(Boolean)
  if (!match || names.length === 0) return null
  return {
    prefix: match[1],
    icons: Object.fromEntries(names.map((name) => [name, { body: PLACEHOLDER_BODY }])),
    width: 24,
    height: 24
  }
}
