import { describe, it, expect } from 'vitest'
import { getRoles } from '../access'
import { ADMIN_SECTIONS, buildAdminNav, visibleTabs } from '../navigation'

const admin = getRoles({ role: 'admin' })
// Éditeur migré : rôle « Éditeur » (actualités et événements partout)
const editor = getRoles({ role: 'editor', permissions: { 'news.manage': 'all', 'events.manage': 'all' } })
const groupAdmin = getRoles({ role: 'user', managed_group_ids: [1] })
const user = getRoles({ role: 'user' })

const allLinks = (nav) => [
  ...(nav.overview ? [nav.overview.to] : []),
  ...nav.sections.flatMap((s) => s.items.map((i) => i.to)),
  ...nav.items.map((i) => i.to)
]

describe('buildAdminNav', () => {
  it("n'affiche rien à un utilisateur", () => {
    expect(buildAdminNav(user)).toBeNull()
  })

  it("range le menu de l'admin par domaine, sans doublon", () => {
    const nav = buildAdminNav(admin)
    expect(nav.grouped).toBe(true)
    expect(nav.sections.map((s) => s.id)).toEqual(['content', 'community', 'resources', 'users', 'system'])
    const links = allLinks(nav)
    expect(new Set(links).size).toBe(links.length)
    expect(links).toContain('/admin')
    expect(links.some((to) => to.startsWith('/group-admin'))).toBe(false)
  })

  it("donne à l'admin chaque page déclarée", () => {
    const declared = ADMIN_SECTIONS.flatMap((s) => s.items.map((i) => i.to))
    expect(allLinks(buildAdminNav(admin))).toEqual(['/admin', ...declared])
  })

  it("donne à l'admin de groupe une liste courte de ses pages", () => {
    const nav = buildAdminNav(groupAdmin)
    expect(nav.grouped).toBe(false)
    expect(nav.titleKey).toBe('groupAdmin.title')
    expect(allLinks(nav)).toEqual(['/admin', '/admin/news', '/admin/events', '/admin/polls', '/admin/resources'])
  })

  it("donne à l'éditeur les actualités et les événements", () => {
    const nav = buildAdminNav(editor)
    expect(nav.overview).toBeNull()
    expect(allLinks(nav)).toEqual(['/admin/news', '/admin/events'])
  })

  it('ajoute les ressources à qui a la permission resources.manage', () => {
    const editorWithRole = getRoles({ role: 'editor', permissions: { 'news.manage': 'all', 'events.manage': 'all', 'resources.manage': 'groups' } })
    expect(allLinks(buildAdminNav(editorWithRole))).toEqual(['/admin/news', '/admin/events', '/admin/resources'])
    const userWithRole = getRoles({ role: 'user', permissions: { 'resources.manage': 'all' } })
    const nav = buildAdminNav(userWithRole)
    expect(allLinks(nav)).toEqual(['/admin/resources'])
    expect(nav.titleKey).toBe('nav.admin.title')
  })

  it("n'affiche rien à l'ancien rôle éditeur sans permission", () => {
    expect(buildAdminNav(getRoles({ role: 'editor' }))).toBeNull()
  })

  it('donne les actualités seules à qui ne gère que les actualités', () => {
    const newsOnly = getRoles({ role: 'user', permissions: { 'news.manage': 'groups' } })
    expect(allLinks(buildAdminNav(newsOnly))).toEqual(['/admin/news'])
  })

  it("réserve la page Rôles à l'admin", () => {
    expect(allLinks(buildAdminNav(admin))).toContain('/admin/roles')
    expect(allLinks(buildAdminNav(groupAdmin))).not.toContain('/admin/roles')
  })
})

describe('visibleTabs', () => {
  it("montre les taxonomies d'actualités à l'admin seulement", () => {
    expect(visibleTabs('news', admin).map((t) => t.to)).toEqual([
      '/admin/news', '/admin/news/categories', '/admin/news/types'
    ])
    expect(visibleTabs('news', groupAdmin).map((t) => t.to)).toEqual(['/admin/news'])
    expect(visibleTabs('news', editor).map((t) => t.to)).toEqual(['/admin/news'])
  })

  it("montre les catégories de ressources à l'admin de groupe", () => {
    expect(visibleTabs('resources', groupAdmin)).toHaveLength(2)
  })

  it('renvoie une liste vide pour une section inconnue', () => {
    expect(visibleTabs('inconnue', admin)).toEqual([])
  })
})
