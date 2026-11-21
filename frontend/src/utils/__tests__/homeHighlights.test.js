import { describe, it, expect } from 'vitest'
import { latestNews, agendaAhead, newsTypeOf } from '../homeHighlights'

describe('latestNews', () => {
  const news = [
    { id: 1, is_pinned: true, published_at: '2026-01-10T10:00:00Z' },
    { id: 2, published_at: '2026-11-20T10:00:00Z' },
    { id: 3, published_at: null, created_at: '2026-11-21T08:00:00Z' },
    { id: 4, published_at: '2026-11-01T10:00:00Z' },
  ]

  it('les plus récemment publiés d’abord, épinglés ou non', () => {
    expect(latestNews(news, 2).map((n) => n.id)).toEqual([3, 2])
  })

  it('ne modifie pas la liste reçue', () => {
    latestNews(news, 2)
    expect(news.map((n) => n.id)).toEqual([1, 2, 3, 4])
  })

  it('liste absente', () => {
    expect(latestNews(undefined, 2)).toEqual([])
  })
})

describe('agendaAhead', () => {
  const today = [{ id: 7, start_date: '2026-11-21T09:00:00Z' }]
  const upcoming = [
    { id: 9, start_date: '2026-11-25T09:00:00Z' },
    { id: 8, start_date: '2026-11-22T14:00:00Z' },
    { id: 7, start_date: '2026-11-21T09:00:00Z' },
  ]

  it('événements du jour puis à venir, par date, sans doublon', () => {
    expect(agendaAhead(today, upcoming, 5).map((e) => e.id)).toEqual([7, 8, 9])
  })

  it('limite le nombre', () => {
    expect(agendaAhead(today, upcoming, 2).map((e) => e.id)).toEqual([7, 8])
  })

  it('marque ceux du jour', () => {
    expect(agendaAhead(today, upcoming, 5).map((e) => e.isToday)).toEqual([true, false, false])
  })

  it('listes absentes', () => {
    expect(agendaAhead(undefined, null, 4)).toEqual([])
  })
})

describe('newsTypeOf', () => {
  const types = [{ id: 1, slug: 'article', name: 'Article' }, null, { id: 2, slug: 'faq', name: 'FAQ' }]

  it('par identifiant d’abord', () => {
    expect(newsTypeOf({ news_type_id: 2, type: 'article' }, types).name).toBe('FAQ')
  })

  it('sinon par le slug historique', () => {
    expect(newsTypeOf({ type: 'article' }, types).name).toBe('Article')
  })

  it('inconnu', () => {
    expect(newsTypeOf({ news_type_id: 9, type: 'x' }, types)).toBeNull()
    expect(newsTypeOf({ type: 'article' }, undefined)).toBeNull()
  })
})
