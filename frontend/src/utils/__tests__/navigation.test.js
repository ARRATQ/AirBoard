import { describe, it, expect } from 'vitest'
import { isPathActive } from '../navigation'

describe('isPathActive', () => {
  it('est actif sur le chemin du lien', () => {
    expect(isPathActive('/admin/news', '/admin/news')).toBe(true)
  })

  it('est actif sur une sous-page', () => {
    expect(isPathActive('/news/mon-article', '/news')).toBe(true)
    expect(isPathActive('/admin/news/new', '/admin/news')).toBe(true)
  })

  it("n'est pas actif sur un chemin qui partage seulement le début", () => {
    expect(isPathActive('/admin/news-types', '/admin/news')).toBe(false)
    expect(isPathActive('/admin/news-categories', '/admin/news')).toBe(false)
    expect(isPathActive('/admin/suggestion-categories', '/admin/suggestions')).toBe(false)
  })

  it('ignore la barre oblique finale', () => {
    expect(isPathActive('/admin/news/', '/admin/news')).toBe(true)
  })

  it("en mode exact, n'est pas actif sur une sous-page", () => {
    expect(isPathActive('/admin/users', '/admin', { exact: true })).toBe(false)
    expect(isPathActive('/admin', '/admin', { exact: true })).toBe(true)
  })

  it('renvoie false sans chemin', () => {
    expect(isPathActive('', '/admin')).toBe(false)
    expect(isPathActive('/admin', undefined)).toBe(false)
  })
})
