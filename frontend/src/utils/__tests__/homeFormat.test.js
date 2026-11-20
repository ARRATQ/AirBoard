import { describe, it, expect } from 'vitest'
import {
  keyFigures,
  pollPercent,
  isLeadingOption,
  daysUntil,
  relativeDay,
  dateLocaleFor,
} from '../homeFormat'

describe('keyFigures', () => {
  const stats = {
    total_users: 12,
    total_groups: 3,
    total_app_groups: 4,
    total_apps: 20,
    total_news: 7,
    total_events: 5,
    total_polls: 2,
    total_accessible_apps: 9,
    total_accessible_events: 6,
    managed_groups_count: 2,
    total_members_count: 30,
    managed_app_groups_count: 1,
    managed_apps_count: 8,
    managed_news_count: 4,
    managed_polls_count: 1,
  }

  it('returns nothing without stats', () => {
    expect(keyFigures(undefined, 'user', [])).toEqual([])
  })

  it('gives the admin the seven platform figures', () => {
    const figures = keyFigures(stats, 'admin', [])
    expect(figures.map((f) => f.value)).toEqual([12, 3, 4, 20, 7, 5, 2])
    expect(figures[0]).toMatchObject({ labelKey: 'home.stats.users', to: '/admin/users' })
  })

  it('gives a group admin the managed figures', () => {
    const figures = keyFigures(stats, 'user', [1])
    expect(figures.map((f) => f.value)).toEqual([2, 30, 1, 8, 4, 1, 9])
    expect(figures.at(-1)).toMatchObject({ labelKey: 'home.stats.accessibleApps', to: '/resources' })
  })

  it('gives a user four figures and replaces missing values with 0', () => {
    const figures = keyFigures({ total_news: 3 }, 'user', [])
    expect(figures.map((f) => [f.labelKey, f.value])).toEqual([
      ['home.stats.apps', 0],
      ['home.stats.news', 3],
      ['home.stats.polls', 0],
      ['home.stats.events', 0],
    ])
  })
})

describe('poll helpers', () => {
  const poll = { options: [{ votes: 3 }, { votes: 1 }, { votes: 0 }] }

  it('computes percentages from the options when total_votes is missing', () => {
    expect(pollPercent(poll.options[0], poll)).toBe(75)
    expect(pollPercent(poll.options[2], poll)).toBe(0)
  })

  it('prefers total_votes', () => {
    expect(pollPercent({ votes: 1 }, { total_votes: 4, options: [] })).toBe(25)
  })

  it('does not divide by zero', () => {
    expect(pollPercent({ votes: 0 }, { options: [{ votes: 0 }] })).toBe(0)
  })

  it('marks the leading option only when someone voted', () => {
    expect(isLeadingOption(poll.options[0], poll)).toBe(true)
    expect(isLeadingOption(poll.options[1], poll)).toBe(false)
    expect(isLeadingOption({ votes: 0 }, { options: [{ votes: 0 }] })).toBe(false)
  })
})

describe('daysUntil', () => {
  const now = new Date('2026-11-20T10:00:00Z')

  it('returns null without an end date', () => {
    expect(daysUntil(null, now)).toBeNull()
  })

  it('rounds up the remaining days', () => {
    expect(daysUntil('2026-11-22T09:00:00Z', now)).toBe(2)
  })

  it('returns 0 or less once the date is past', () => {
    expect(daysUntil('2026-11-19T10:00:00Z', now)).toBeLessThanOrEqual(0)
  })
})

describe('relativeDay', () => {
  const now = new Date('2026-11-20T15:00:00')

  it('says today and yesterday in the requested language', () => {
    expect(relativeDay('2026-11-20T08:00:00', 'fr-FR', now)).toBe('aujourd’hui')
    expect(relativeDay('2026-11-19T08:00:00', 'en-US', now)).toBe('yesterday')
  })

  it('counts days within the week', () => {
    expect(relativeDay('2026-11-17T08:00:00', 'en-US', now)).toBe('3 days ago')
  })

  it('falls back to a short date after a week', () => {
    expect(relativeDay('2026-10-01T08:00:00', 'en-US', now)).toBe('Oct 1')
  })

  it('returns an empty string without a date', () => {
    expect(relativeDay('', 'fr-FR', now)).toBe('')
  })
})

describe('dateLocaleFor', () => {
  it('maps app languages to date locales', () => {
    expect(dateLocaleFor('ar')).toBe('ar-MA')
    expect(dateLocaleFor('fr')).toBe('fr-FR')
    expect(dateLocaleFor('de')).toBe('de')
  })
})
