import { describe, it, expect } from 'vitest'
import { holidayRange, holidaysFromCalendar } from '../calendarHolidays'

describe('holidayRange', () => {
  it('couvre l’année en cours et la suivante', () => {
    expect(holidayRange(new Date(2026, 10, 21))).toEqual({ start: '2026-01-01', end: '2027-12-31' })
  })
})

describe('holidaysFromCalendar', () => {
  it('ne garde que les jours fériés, jour par jour', () => {
    const map = holidaysFromCalendar({
      events: [
        { title: 'Réunion', start_date: '2026-11-18T09:00:00Z', is_holiday: false },
        { title: 'Fête', start_date: '2026-11-18', end_date: '2026-11-19', is_holiday: true },
      ],
    })
    expect([...map.keys()]).toEqual(['2026-11-18', '2026-11-19'])
    expect(map.get('2026-11-19')).toBe('Fête')
  })

  it('accepte la forme Events de l’API et une réponse vide', () => {
    expect(holidaysFromCalendar({ Events: [{ title: 'Fête', start_date: '2026-01-01', is_holiday: true }] }).size).toBe(1)
    expect(holidaysFromCalendar(null).size).toBe(0)
  })
})
