import { describe, it, expect } from 'vitest'
import { pageTitle } from '../pageTitle'

const messages = {
  'pageTitles.AdminNewsEdit': 'Modifier l’article',
  'pageTitles.default': 'Portail',
}
const t = (key) => messages[key] ?? key
const te = (key) => key in messages

describe('pageTitle', () => {
  it('translates the title from the route name', () => {
    expect(pageTitle({ name: 'AdminNewsEdit', meta: { title: 'Edit Article' } }, t, te)).toBe('Modifier l’article - Airboard')
  })

  it('falls back to meta.title when no translation exists', () => {
    expect(pageTitle({ name: 'Unknown', meta: { title: 'Ancien titre' } }, t, te)).toBe('Ancien titre - Airboard')
  })

  it('uses the default title for routes without any title', () => {
    expect(pageTitle({ name: undefined, meta: {} }, t, te)).toBe('Airboard - Portail')
    expect(pageTitle(null, t, te)).toBe('Airboard - Portail')
  })
})
