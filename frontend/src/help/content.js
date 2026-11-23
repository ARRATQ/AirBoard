// Textes de l'aide, livrés avec l'application (un fichier par langue et par guide, chargé
// à la demande). Une langue sans le guide retombe sur le français.
const FILES = import.meta.glob('./content/*/*.md', { query: '?raw', import: 'default' })

export const HELP_LOCALES = ['fr', 'en', 'es', 'ar']
const FALLBACK_LOCALE = 'fr'

const pathFor = (locale, id) => `./content/${locale}/${id}.md`

/**
 * Texte Markdown d'un guide.
 * @param {string} locale langue de l'interface
 * @param {string} id identifiant du guide (catalog.js)
 * @returns {Promise<string>}
 */
export async function loadTopic(locale, id) {
  const lang = HELP_LOCALES.includes(locale) ? locale : FALLBACK_LOCALE
  const loader = FILES[pathFor(lang, id)] || FILES[pathFor(FALLBACK_LOCALE, id)]
  if (!loader) throw new Error(`Guide d'aide introuvable : ${id}`)
  return loader()
}
