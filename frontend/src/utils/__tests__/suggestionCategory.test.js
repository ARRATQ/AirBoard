import { describe, it, expect } from 'vitest'
import { suggestionCategoryLabel } from '../suggestionCategory'

const messages = {
  'suggestions.categories.functionality': 'Fonctionnalité',
  'suggestions.noCategory': 'Aucune catégorie',
}
const t = (key) => messages[key] ?? key
const te = (key) => key in messages

describe('suggestionCategoryLabel', () => {
  it('prefers the linked category name', () => {
    expect(suggestionCategoryLabel({ category_ref: { name: 'RH' }, category: 'functionality' }, t, te)).toBe('RH')
  })

  it('translates a legacy category code', () => {
    expect(suggestionCategoryLabel({ category_ref: null, category: 'functionality' }, t, te)).toBe('Fonctionnalité')
  })

  it('keeps an unknown legacy value as is', () => {
    expect(suggestionCategoryLabel({ category: 'Logistique' }, t, te)).toBe('Logistique')
  })

  it('falls back to "no category"', () => {
    expect(suggestionCategoryLabel({ category: '' }, t, te)).toBe('Aucune catégorie')
    expect(suggestionCategoryLabel(null, t, te)).toBe('Aucune catégorie')
  })
})
