import { describe, it, expect } from 'vitest'
import { ADMIN, GROUP_ADMIN, RESOURCES, NEWS, EVENTS, POLLS, MODERATION, getRoles, canAccess, getScope, permissionScope, canManageItem, canModerateComments } from '../access'

const admin = getRoles({ role: 'admin' })
// Ancien compte « editor » resté sans rôle : traité comme un utilisateur
const editor = getRoles({ role: 'editor' })
// Éditeur migré : rôle « Éditeur » (actualités et événements partout, ressources de ses groupes)
const migratedEditor = getRoles({ role: 'editor', permissions: { 'news.manage': 'all', 'events.manage': 'all', 'resources.manage': 'groups' } })
const groupAdmin = getRoles({ role: 'user', managed_group_ids: [3] })
const editorAndGroupAdmin = getRoles({ role: 'editor', admin_of_groups: [{ id: 3 }], permissions: { 'news.manage': 'all' } })
const user = getRoles({ role: 'user', managed_group_ids: [] })

describe('getRoles', () => {
  it('reconnaît un admin de groupe par admin_of_groups ou managed_group_ids', () => {
    expect(getRoles({ role: 'user', admin_of_groups: [{ id: 1 }] }).isGroupAdmin).toBe(true)
    expect(groupAdmin.isGroupAdmin).toBe(true)
    expect(user.isGroupAdmin).toBe(false)
  })

  it('accepte un utilisateur absent', () => {
    expect(getRoles(null)).toEqual({ isAdmin: false, isGroupAdmin: false, permissions: {} })
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

  it("ouvre NEWS et EVENTS à qui a la permission, à l'admin et à l'admin de groupe", () => {
    for (const level of [NEWS, EVENTS]) {
      expect(canAccess(level, admin)).toBe(true)
      expect(canAccess(level, migratedEditor)).toBe(true)
      expect(canAccess(level, groupAdmin)).toBe(true)
      expect(canAccess(level, user)).toBe(false)
    }
  })

  it("ne donne plus rien à l'ancien rôle éditeur sans permission", () => {
    expect(canAccess(NEWS, editor)).toBe(false)
    expect(canAccess(EVENTS, editor)).toBe(false)
  })

  it('ouvre chaque module selon sa propre permission', () => {
    const newsOnly = getRoles({ role: 'user', permissions: { 'news.manage': 'groups' } })
    expect(canAccess(NEWS, newsOnly)).toBe(true)
    expect(canAccess(EVENTS, newsOnly)).toBe(false)
    expect(canAccess(RESOURCES, newsOnly)).toBe(false)
  })

  it('refuse un niveau inconnu', () => {
    expect(canAccess('superuser', groupAdmin)).toBe(false)
  })
})

describe('getScope', () => {
  it("donne le périmètre global à l'admin", () => {
    expect(getScope(GROUP_ADMIN, admin)).toBe('global')
    expect(getScope(NEWS, admin)).toBe('global')
  })

  it('donne le périmètre global selon la portée de chaque permission', () => {
    expect(getScope(NEWS, migratedEditor)).toBe('global')
    expect(getScope(EVENTS, migratedEditor)).toBe('global')
    expect(getScope(RESOURCES, migratedEditor)).toBe('group')
    expect(getScope(NEWS, editorAndGroupAdmin)).toBe('global')
    expect(getScope(GROUP_ADMIN, editorAndGroupAdmin)).toBe('group')
  })

  it("limite l'admin de groupe à ses groupes", () => {
    expect(getScope(NEWS, groupAdmin)).toBe('group')
    expect(getScope(GROUP_ADMIN, groupAdmin)).toBe('group')
  })
})

describe('permissions des rôles', () => {
  const inGroups = getRoles({ role: 'user', permissions: { 'resources.manage': 'groups' } })
  const everywhere = getRoles({ role: 'editor', permissions: { 'resources.manage': 'all' } })
  const other = getRoles({ role: 'user', permissions: { 'news.manage': 'all' } })

  it('lit la portée de la permission ; admin : partout', () => {
    expect(permissionScope(inGroups, 'resources.manage')).toBe('groups')
    expect(permissionScope(everywhere, 'resources.manage')).toBe('all')
    expect(permissionScope(user, 'resources.manage')).toBe('')
    expect(permissionScope(admin, 'resources.manage')).toBe('all')
  })

  it('ouvre les ressources à qui a la permission, ou à un ancien admin de groupe', () => {
    expect(canAccess(RESOURCES, inGroups)).toBe(true)
    expect(canAccess(RESOURCES, everywhere)).toBe(true)
    expect(canAccess(RESOURCES, groupAdmin)).toBe(true)
    expect(canAccess(RESOURCES, other)).toBe(false)
    expect(canAccess(RESOURCES, editor)).toBe(false)
    expect(canAccess(RESOURCES, user)).toBe(false)
  })

  it('périmètre global seulement avec la portée « tout le site »', () => {
    expect(getScope(RESOURCES, everywhere)).toBe('global')
    expect(getScope(RESOURCES, inGroups)).toBe('group')
    expect(getScope(RESOURCES, groupAdmin)).toBe('group')
    expect(getScope(RESOURCES, admin)).toBe('global')
  })

  it('ignore des permissions mal formées', () => {
    expect(getRoles({ role: 'user', permissions: 'resources.manage' }).permissions).toEqual({})
  })
})

describe('sondages et modération', () => {
  const pollsInGroups = getRoles({ role: 'user', permissions: { 'polls.manage': 'groups' } })
  const moderator = getRoles({ role: 'user', permissions: { 'comments.moderate': 'all' } })

  it('ouvre les sondages à la permission, à l\'admin et à l\'ancien admin de groupe', () => {
    expect(canAccess(POLLS, pollsInGroups)).toBe(true)
    expect(canAccess(POLLS, admin)).toBe(true)
    expect(canAccess(POLLS, groupAdmin)).toBe(true)
    expect(canAccess(POLLS, moderator)).toBe(false)
    expect(canAccess(POLLS, user)).toBe(false)
  })

  it('réserve la modération à la permission (pas à l\'admin de groupe)', () => {
    expect(canAccess(MODERATION, moderator)).toBe(true)
    expect(canAccess(MODERATION, admin)).toBe(true)
    expect(canAccess(MODERATION, groupAdmin)).toBe(false)
    expect(canAccess(MODERATION, pollsInGroups)).toBe(false)
    expect(canModerateComments({ role: 'user', permissions: { 'comments.moderate': 'all' } })).toBe(true)
    expect(canModerateComments({ role: 'admin' })).toBe(true)
    expect(canModerateComments({ role: 'user' })).toBe(false)
    expect(canModerateComments(null)).toBe(false)
  })
})

describe('canManageItem', () => {
  const g1 = { id: 1 }
  const g2 = { id: 2 }
  const common = { id: 3 }
  // Membre de g1 et de Common ; le serveur ne compte que g1 (hors groupe par défaut)
  const member = { id: 7, role: 'user', groups: [g1, common], permission_group_ids: [1], permissions: { 'polls.manage': 'groups' } }

  it('tout le site ou admin : tout', () => {
    expect(canManageItem({ id: 1, role: 'admin' }, 'polls.manage', { target_groups: [] })).toBe(true)
    expect(canManageItem({ id: 2, role: 'user', permissions: { 'polls.manage': 'all' } }, 'polls.manage', {})).toBe(true)
  })

  it('ses groupes : contenus ciblant un de ses groupes, pas les contenus globaux', () => {
    expect(canManageItem(member, 'polls.manage', { target_groups: [g1] })).toBe(true)
    expect(canManageItem(member, 'polls.manage', { target_groups: [g2] })).toBe(false)
    expect(canManageItem(member, 'polls.manage', { target_groups: [common] })).toBe(false)
    expect(canManageItem(member, 'polls.manage', { target_groups: [] })).toBe(false)
    expect(canManageItem(member, 'news.manage', { target_groups: [g1] })).toBe(false)
  })

  it('auteur : toujours ; ancien admin de groupe : ses groupes administrés', () => {
    expect(canManageItem({ id: 9, role: 'user' }, 'news.manage', { author_id: 9 })).toBe(true)
    expect(canManageItem({ id: 9, role: 'user', managed_group_ids: [2] }, 'news.manage', { target_groups: [g2] })).toBe(true)
    expect(canManageItem(null, 'news.manage', { author_id: 9 })).toBe(false)
  })
})
