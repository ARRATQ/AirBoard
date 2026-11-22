/**
 * Recherche des listes à saisie semi-automatique (ComboSelect) : insensible à la casse et aux
 * accents, sur l'étiquette et l'étiquette secondaire (e-mail, nombre de membres…).
 */

/** Texte sans accents ni majuscules. */
export const normalizeText = (value) =>
  String(value ?? '')
    .normalize('NFD')
    .replace(/[̀-ͯ]/g, '')
    .toLowerCase()

/**
 * L'option correspond-elle au texte saisi ? Chaque mot saisi doit apparaître.
 * @param {{ label?: string, sublabel?: string }} option
 * @param {string} query
 * @returns {boolean}
 */
export function matchesQuery(option, query) {
  const words = normalizeText(query).split(/\s+/).filter(Boolean)
  if (words.length === 0) return true
  const haystack = normalizeText(`${option?.label ?? ''} ${option?.sublabel ?? ''}`)
  return words.every((word) => haystack.includes(word))
}

/**
 * Option de personne : nom affiché (prénom nom, sinon identifiant) et e-mail.
 * @param {{ id: number, username?: string, email?: string, first_name?: string, last_name?: string }} user
 */
export const personOption = (user) => ({
  id: user.id,
  label: [user.first_name, user.last_name].filter(Boolean).join(' ') || user.username || `#${user.id}`,
  sublabel: user.email || user.username || '',
})

/**
 * Option de groupe : nom et couleur.
 * @param {{ id: number, name: string, color?: string }} group
 * @param {string} [sublabel]
 */
export const groupOption = (group, sublabel = '') => ({
  id: group.id,
  label: group.name,
  color: group.color || undefined,
  sublabel,
})
