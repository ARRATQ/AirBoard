import { describe, it, expect } from 'vitest'
import { resourceTypeOf, typeFamilyOf, isNewResource, activeAdminNote, noteUntilToInput, noteUntilFromInput, NEW_RESOURCE_DAYS } from '../resourceType'

const SP = 'https://contoso.sharepoint.com'

describe('resourceTypeOf', () => {
  it.each([
    [`${SP}/:w:/r/sites/rh/Documents/charte.docx?d=1`, 'word'],
    [`${SP}/:x:/g/personal/a/EaBc`, 'excel'],
    [`${SP}/:p:/r/sites/com/deck.pptx`, 'powerpoint'],
    [`${SP}/:b:/s/direction/EfGh`, 'pdf'],
    [`${SP}/:f:/s/rh/Ej9k`, 'folder'],
  ])('reads the SharePoint marker in %s', (url, type) => {
    expect(resourceTypeOf(url)).toBe(type)
  })

  it('recognises an intranet site', () => {
    expect(resourceTypeOf(`${SP}/sites/direction`)).toBe('site')
    expect(resourceTypeOf(`${SP}/sites/direction/SitePages/Accueil.aspx`)).toBe('site')
  })

  it('reads direct file extensions', () => {
    expect(resourceTypeOf('https://example.com/files/organigramme.PDF')).toBe('pdf')
    expect(resourceTypeOf('https://example.com/a/bareme.xlsx')).toBe('excel')
    expect(resourceTypeOf('https://example.com/a/note.doc')).toBe('word')
  })

  it('treats a bare host as an application and a deeper path as a web link', () => {
    expect(resourceTypeOf('https://gitlab.com')).toBe('app')
    expect(resourceTypeOf('https://outlook.office.com/')).toBe('app')
    expect(resourceTypeOf('https://example.com/acces/plan.html')).toBe('link')
  })

  it('falls back to a web link for unreadable URLs', () => {
    expect(resourceTypeOf('')).toBe('link')
    expect(resourceTypeOf('not a url')).toBe('link')
  })
})

describe('typeFamilyOf', () => {
  it('groups types for the filter', () => {
    expect(typeFamilyOf('app')).toBe('apps')
    expect(typeFamilyOf('excel')).toBe('docs')
    expect(typeFamilyOf('folder')).toBe('docs')
    expect(typeFamilyOf('site')).toBe('sites')
    expect(typeFamilyOf('link')).toBe('sites')
  })
})

describe('isNewResource', () => {
  const now = new Date('2026-11-20T12:00:00Z')

  it(`is new for ${NEW_RESOURCE_DAYS} days after creation`, () => {
    expect(isNewResource({ created_at: '2026-11-07T12:00:01Z' }, now)).toBe(true)
    expect(isNewResource({ created_at: '2026-11-06T11:59:59Z' }, now)).toBe(false)
  })

  it('ignores missing or future dates', () => {
    expect(isNewResource({}, now)).toBe(false)
    expect(isNewResource({ created_at: '2026-12-01T00:00:00Z' }, now)).toBe(false)
  })
})

describe('activeAdminNote', () => {
  const now = new Date('2026-11-20T12:00:00Z')

  it('returns the note while its end date is not past', () => {
    expect(activeAdminNote({ admin_note: ' Nouveau barème ', admin_note_until: '2026-11-20T23:59:00Z' }, now)).toBe('Nouveau barème')
    expect(activeAdminNote({ admin_note: 'Sans fin' }, now)).toBe('Sans fin')
  })

  it('hides an expired or empty note', () => {
    expect(activeAdminNote({ admin_note: 'Passée', admin_note_until: '2026-11-19T00:00:00Z' }, now)).toBe('')
    expect(activeAdminNote({ admin_note: '   ' }, now)).toBe('')
  })
})

describe('note end date conversions', () => {
  it('keeps the note until the end of the chosen local day', () => {
    const iso = noteUntilFromInput('2026-12-01')
    const d = new Date(iso)
    expect([d.getFullYear(), d.getMonth(), d.getDate(), d.getHours(), d.getMinutes()]).toEqual([2026, 11, 1, 23, 59])
    expect(noteUntilToInput(iso)).toBe('2026-12-01')
  })

  it('maps empty or invalid values to no date', () => {
    expect(noteUntilFromInput('')).toBeNull()
    expect(noteUntilFromInput('pas une date')).toBeNull()
    expect(noteUntilToInput(null)).toBe('')
    expect(noteUntilToInput('n/a')).toBe('')
  })
})
