// Titre de l'onglet : traduit depuis pageTitles.<nom de la route>, sinon meta.title.

const SUFFIX = 'Airboard'

/**
 * @param {{ name?: string|symbol, meta?: { title?: string } }|null} route
 * @param {(key: string) => string} t
 * @param {(key: string) => boolean} te
 * @returns {string}
 */
export function pageTitle(route, t, te) {
  const key = route?.name ? `pageTitles.${String(route.name)}` : null
  const label = key && te(key) ? t(key) : route?.meta?.title
  return label ? `${label} - ${SUFFIX}` : `${SUFFIX} - ${t('pageTitles.default')}`
}
