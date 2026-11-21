import { describe, it, expect } from 'vitest'
import { dateKey, parseDateKey, eventDayRange, eventCoversDay, holidayMap, countWorkingDays, eventStatusTone, MAX_RANGE_DAYS } from '../agenda'

describe('dateKey / parseDateKey', () => {
  it('uses local time, not UTC', () => {
    expect(dateKey(new Date(2026, 6, 14, 0, 30))).toBe('2026-07-14')
    expect(dateKey(parseDateKey('2026-03-29'))).toBe('2026-03-29')
  })

  it('rejects malformed or impossible dates', () => {
    expect(parseDateKey('')).toBeNull()
    expect(parseDateKey('2026-02-30')).toBeNull()
    expect(parseDateKey('14/07/2026')).toBeNull()
  })
})

describe('eventDayRange / eventCoversDay', () => {
  it('keeps the calendar date of all-day events and holidays', () => {
    expect(eventDayRange({ start_date: '2026-12-25T00:00:00Z', is_holiday: true })).toEqual(['2026-12-25', '2026-12-25'])
    expect(eventDayRange({ start_date: '2026-08-03T00:00:00Z', end_date: '2026-08-07T00:00:00Z', is_all_day: true })).toEqual(['2026-08-03', '2026-08-07'])
  })

  it('covers every day of a multi-day event', () => {
    const event = { start_date: new Date(2026, 10, 2, 9).toISOString(), end_date: new Date(2026, 10, 4, 17).toISOString() }
    expect(['2026-11-01', '2026-11-02', '2026-11-03', '2026-11-04', '2026-11-05'].map((k) => eventCoversDay(event, k))).toEqual([false, true, true, true, false])
  })

  it('ignores events without a valid start', () => {
    expect(eventDayRange({ start_date: 'n/a' })).toBeNull()
    expect(eventCoversDay({}, '2026-11-02')).toBe(false)
  })

  it('never ends before it starts', () => {
    expect(eventDayRange({ start_date: '2026-05-10T00:00:00Z', end_date: '2026-05-01T00:00:00Z', is_all_day: true })).toEqual(['2026-05-10', '2026-05-10'])
  })
})

describe('holidayMap', () => {
  it('maps each covered day to the holiday name and ignores other events', () => {
    const map = holidayMap([
      { title: 'Aïd', start_date: '2026-05-27T00:00:00Z', end_date: '2026-05-28T00:00:00Z', is_holiday: true },
      { title: 'Réunion', start_date: '2026-05-27T09:00:00Z' },
    ])
    expect([...map.entries()]).toEqual([['2026-05-27', 'Aïd'], ['2026-05-28', 'Aïd']])
  })
})

describe('countWorkingDays', () => {
  const holidays = new Map([['2026-12-25', 'Noël'], ['2026-12-26', 'Lendemain']])

  it('splits a period into working, weekend and holiday days', () => {
    // Lun 21 → Dim 27 déc. 2026 ; Noël vendredi, lendemain samedi (compte en week-end).
    const r = countWorkingDays({ start: '2026-12-21', end: '2026-12-27', holidays })
    expect(r).toMatchObject({ totalDays: 7, workingDays: 4, weekendDays: 2, holidayDays: 1, nonWorkingDays: 3 })
    expect(r.holidays).toEqual([{ date: '2026-12-25', title: 'Noël' }])
  })

  it('counts holidays as working days when they are not excluded', () => {
    expect(countWorkingDays({ start: '2026-12-21', end: '2026-12-27', holidays, excludeHolidays: false }).workingDays).toBe(5)
  })

  it('honours a custom weekend (Friday and Saturday)', () => {
    expect(countWorkingDays({ start: '2026-12-21', end: '2026-12-27', weekendDays: [5, 6] }).workingDays).toBe(5)
  })

  it('stays correct across a daylight-saving change', () => {
    // 29 mars 2026 : passage à l'heure d'été en Europe ; lundi 30 doit rester un jour ouvré.
    expect(countWorkingDays({ start: '2026-03-28', end: '2026-03-31' }).workingDays).toBe(2)
  })

  it('returns null for missing, reversed or too long periods', () => {
    expect(countWorkingDays({ start: '', end: '2026-12-31' })).toBeNull()
    expect(countWorkingDays({ start: '2026-12-31', end: '2026-12-01' })).toBeNull()
    expect(countWorkingDays({ start: '2000-01-01', end: '2026-01-01' })).toBeNull()
    expect(MAX_RANGE_DAYS).toBeGreaterThan(365)
  })
})

describe('eventStatusTone', () => {
  it('colours only exceptions', () => {
    expect(eventStatusTone('tentative')).toBe('warn')
    expect(eventStatusTone('cancelled')).toBe('bad')
    expect(eventStatusTone('confirmed')).toBe('')
    expect(eventStatusTone(undefined)).toBe('')
  })
})
