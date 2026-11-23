import { describe, it, expect } from 'vitest'
import { readFileSync, readdirSync } from 'node:fs'
import { resolve } from 'node:path'
import { getRoles } from '@/config/access'
import { HELP_TOPICS, HELP_AUDIENCES, visibleTopics, topicScope, topicForRoute, applyScope } from '../catalog'
import { HELP_LOCALES } from '../content'

const admin = getRoles({ role: 'admin' })
const user = getRoles({ role: 'user' })
const groupNews = getRoles({ role: 'user', permissions: { 'news.manage': 'groups' } })
const subGroups = getRoles({ role: 'user', permissions: { 'groups.manage': 'groups' } })
const ids = (topics) => topics.map((t) => t.id)
const everyoneIds = HELP_TOPICS.filter((t) => t.audience === 'everyone').map((t) => t.id)

describe('visibleTopics', () => {
  it('donne à un utilisateur les seuls guides pour tous', () => {
    expect(ids(visibleTopics(user))).toEqual(everyoneIds)
  })

  it('ajoute le guide de chaque permission, et seulement celui-là', () => {
    expect(ids(visibleTopics(groupNews))).toEqual([...everyoneIds, 'manage-news'])
    expect(ids(visibleTopics(subGroups))).toEqual([...everyoneIds, 'sub-groups'])
  })

  it("donne tout à l'admin, sauf « Mes groupes » qu'il gère depuis « Groupes »", () => {
    const visible = ids(visibleTopics(admin))
    expect(visible).toContain('admin-roles')
    expect(visible).toContain('manage-news')
    expect(visible).not.toContain('sub-groups')
    expect(visible).toContain('admin-groups')
  })

  it("ne montre jamais l'administration à qui n'est pas admin", () => {
    expect(visibleTopics(groupNews).some((t) => t.audience === 'admin')).toBe(false)
  })
})

describe('topicScope', () => {
  const news = HELP_TOPICS.find((t) => t.id === 'manage-news')

  it('suit la portée de la permission', () => {
    expect(topicScope(news, groupNews)).toBe('groups')
    expect(topicScope(news, getRoles({ role: 'user', permissions: { 'news.manage': 'all' } }))).toBe('all')
    expect(topicScope(news, admin)).toBe('all')
  })

  it("est vide pour un guide sans permission ou réservé à l'admin", () => {
    expect(topicScope(HELP_TOPICS[0], user)).toBe('')
    expect(topicScope(HELP_TOPICS.find((t) => t.id === 'admin-roles'), admin)).toBe('')
  })
})

describe('topicForRoute', () => {
  const all = visibleTopics(admin)

  it('ouvre le guide de la page affichée, sous-pages comprises', () => {
    expect(topicForRoute('/resources', all)?.id).toBe('resources')
    expect(topicForRoute('/news/mon-article', all)?.id).toBe('news')
    expect(topicForRoute('/admin/news/categories', all)?.id).toBe('manage-news')
  })

  it('préfère le chemin le plus précis', () => {
    expect(topicForRoute('/admin', all)?.id).toBe('admin-system')
    expect(topicForRoute('/admin/roles', all)?.id).toBe('admin-roles')
  })

  it('ne confond pas deux chemins qui se ressemblent', () => {
    expect(topicForRoute('/newsletter', all)).toBeNull()
  })

  it("ne propose pas un guide que l'utilisateur ne voit pas", () => {
    expect(topicForRoute('/admin/roles', visibleTopics(groupNews))).toBeNull()
  })
})

describe('applyScope', () => {
  const md = 'Avant\n:::all\nTout le site\n:::\n:::groups\nVos groupes\n:::\nAprès\n'

  it('garde le bloc de la portée et retire l’autre', () => {
    expect(applyScope(md, 'all')).toBe('Avant\nTout le site\nAprès\n')
    expect(applyScope(md, 'groups')).toBe('Avant\nVos groupes\nAprès\n')
  })

  it('retire les deux blocs sans portée', () => {
    expect(applyScope(md, '')).toBe('Avant\nAprès\n')
  })
})

describe('textes des guides', () => {
  const root = resolve(__dirname, '../content')

  it('existent dans chaque langue pour chaque guide, sans fichier en trop', () => {
    const expected = HELP_TOPICS.map((t) => `${t.id}.md`).sort()
    for (const locale of HELP_LOCALES) {
      expect(readdirSync(resolve(root, locale)).sort(), locale).toEqual(expected)
    }
  })

  it('ont un titre et des blocs de portée bien fermés', () => {
    for (const locale of HELP_LOCALES) {
      for (const topic of HELP_TOPICS) {
        const text = readFileSync(resolve(root, locale, `${topic.id}.md`), 'utf8')
        expect(text.startsWith('# '), `${locale}/${topic.id}`).toBe(true)
        const opens = text.match(/^:::(all|groups)\s*$/gm) || []
        const closes = text.match(/^:::\s*$/gm) || []
        expect(closes.length, `${locale}/${topic.id}`).toBe(opens.length)
        // Une fois les blocs appliqués, aucun marqueur ne reste
        for (const scope of ['all', 'groups', '']) {
          expect(applyScope(text, scope), `${locale}/${topic.id}`).not.toMatch(/^:::/m)
        }
      }
    }
  })

  it('ne lient qu’à des pages de l’application', () => {
    for (const locale of HELP_LOCALES) {
      for (const topic of HELP_TOPICS) {
        const text = readFileSync(resolve(root, locale, `${topic.id}.md`), 'utf8')
        for (const [, href] of text.matchAll(/\]\(([^)]+)\)/g)) {
          expect(href.startsWith('/') && !href.startsWith('//'), `${locale}/${topic.id} → ${href}`).toBe(true)
        }
      }
    }
  })

  it('rangent chaque guide dans un public connu', () => {
    for (const topic of HELP_TOPICS) expect(HELP_AUDIENCES).toContain(topic.audience)
  })
})
