// Catégorie affichée d'une suggestion : la catégorie liée, sinon l'ancien code
// (« functionality », « general »…) traduit quand il est connu.

/**
 * @param {{ category_ref?: { name?: string }|null, category?: string }|null} suggestion
 * @param {(key: string) => string} t
 * @param {(key: string) => boolean} te
 * @returns {string}
 */
export function suggestionCategoryLabel(suggestion, t, te) {
  if (suggestion?.category_ref?.name) return suggestion.category_ref.name
  const code = suggestion?.category
  if (!code) return t('suggestions.noCategory')
  const key = `suggestions.categories.${code}`
  return te(key) ? t(key) : code
}
