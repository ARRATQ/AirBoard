import { describe, it, expect } from 'vitest'
import { ADMIN, GROUP_ADMIN, CONTENT, getRoles, canAccess, getScope } from '../access'

const admin = getRoles({ role: 'admin' })
const editor = getRoles({ role: 'editor' })
const groupAdmin = getRoles({ role: 'user', managed_group_ids: [3] })
const editorAndGroupAdmin = getRoles({ role: 'editor', admin_of_groups: [{ id: 3 }] })
const user = getRoles({ role: 'user', managed_group_ids: [] })

describe('getRoles', () => {
  it('reconnaît un admin de groupe par admin_of_groups ou managed_group_ids', () => {
    expect(getRoles({ role: 'user', admin_of_groups: [{ id: 1 }] }).isGroupAdmin).toBe(true)
    expect(groupAdmin.isGroupAdmin).toBe(true)
    expect(user.isGroupAdmin).toBe(false)
  })

  it('accepte un utilisateur absent', () => {
    expect(getRoles(null)).toEqual({ isAdmin: false, isEditor: false, isGroupAdmin: false })
  })
})

describe('canAccess', () => {
  it('laisse passer une page sans niveau', () => {
    expect(canAccess(undefined, user)).toBe(true)
  })

  it("réserve ADMIN à l'admin global", () => {
    expect(canAccess(ADMIN, admin)).toBe(true)
    expect(canAccess(ADMIN, editor)).toBe(false)
    expect(canAccess(ADMIN, groupAdmin)).toBe(false)
  })

  it("ouvre GROUP_ADMIN à l'admin global et à l'admin de groupe", () => {
    expect(canAccess(GROUP_ADMIN, admin)).toBe(true)
    expect(canAccess(GROUP_ADMIN, groupAdmin)).toBe(true)
    expect(canAccess(GROUP_ADMIN, editor)).toBe(false)
    expect(canAccess(GROUP_ADMIN, user)).toBe(false)
  })

  it("ouvre CONTENT à l'admin, l'éditeur et l'admin de groupe", () => {
    expect(canAccess(CONTENT, admin)).toBe(true)
    expect(canAccess(CONTENT, editor)).toBe(true)
    expect(canAccess(CONTENT, groupAdmin)).toBe(true)
    expect(canAccess(CONTENT, user)).toBe(false)
  })

  it('refuse un niveau inconnu', () => {
    expect(canAccess('superuser', groupAdmin)).toBe(false)
  })
})

describe('getScope', () => {
  it("donne le périmètre global à l'admin", () => {
    expect(getScope(GROUP_ADMIN, admin)).toBe('global')
    expect(getScope(CONTENT, admin)).toBe('global')
  })

  it("donne le périmètre global à l'éditeur pour le contenu seulement", () => {
    expect(getScope(CONTENT, editor)).toBe('global')
    expect(getScope(CONTENT, editorAndGroupAdmin)).toBe('global')
    expect(getScope(GROUP_ADMIN, editorAndGroupAdmin)).toBe('group')
  })

  it("limite l'admin de groupe à ses groupes", () => {
    expect(getScope(CONTENT, groupAdmin)).toBe('group')
    expect(getScope(GROUP_ADMIN, groupAdmin)).toBe('group')
  })
})
