/**
 * Dit si un lien de menu doit être marqué actif pour le chemin courant.
 *
 * Un lien est actif sur son propre chemin et sur ses sous-pages
 * (`/news` reste actif sur `/news/mon-article`), mais pas sur un chemin
 * qui commence seulement par les mêmes lettres (`/admin/news` n'est pas
 * actif sur `/admin/news-types`).
 *
 * @param {string} currentPath chemin de la route courante
 * @param {string} linkPath chemin du lien
 * @param {{ exact?: boolean }} [options] `exact` : actif seulement sur le chemin lui-même
 * @returns {boolean}
 */
export function isPathActive(currentPath, linkPath, { exact = false } = {}) {
  if (!currentPath || !linkPath) return false
  const current = trimTrailingSlash(currentPath)
  const link = trimTrailingSlash(linkPath)
  if (current === link) return true
  if (exact) return false
  return current.startsWith(`${link}/`)
}

function trimTrailingSlash(path) {
  return path.length > 1 && path.endsWith('/') ? path.slice(0, -1) : path
}
