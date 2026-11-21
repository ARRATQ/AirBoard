import { describe, it, expect } from 'vitest'
import { toDateTimeInput, fromDateTimeInput, isEndBeforeStart } from '../dateInput'

// Les tests tournent en Europe/Paris (vite.config.js) : sans fuseau non UTC,
// le décalage à chaque enregistrement ne se verrait pas.

describe('toDateTimeInput', () => {
  it('shows the stored instant in local time', () => {
    expect(toDateTimeInput('2026-12-10T13:00:00Z')).toBe('2026-12-10T14:00') // hiver, UTC+1
    expect(toDateTimeInput('2026-07-14T08:30:00Z')).toBe('2026-07-14T10:30') // été, UTC+2
  })

  it('crosses midnight in local time', () => {
    expect(toDateTimeInput('2026-12-31T23:30:00Z')).toBe('2027-01-01T00:30')
  })

  it('returns the empty value for missing or invalid input', () => {
    expect(toDateTimeInput(null)).toBe('')
    expect(toDateTimeInput('')).toBe('')
    expect(toDateTimeInput('pas une date')).toBe('')
    expect(toDateTimeInput(null, null)).toBeNull()
  })
})

describe('fromDateTimeInput', () => {
  it('reads the input as local time', () => {
    expect(fromDateTimeInput('2026-12-10T14:00')).toBe('2026-12-10T13:00:00.000Z')
  })

  it('returns null for missing or invalid input', () => {
    expect(fromDateTimeInput('')).toBeNull()
    expect(fromDateTimeInput(null)).toBeNull()
    expect(fromDateTimeInput('2026-13-45T99:00')).toBeNull()
  })
})

describe('round trip', () => {
  it('does not move a date that is saved without changes', () => {
    for (const iso of ['2026-12-10T13:00:00.000Z', '2026-07-14T08:30:00.000Z', '2026-03-29T00:15:00.000Z']) {
      let value = iso
      for (let i = 0; i < 3; i++) value = fromDateTimeInput(toDateTimeInput(value))
      expect(value).toBe(iso)
    }
  })
})

describe('isEndBeforeStart', () => {
  it('flags an end before the start', () => {
    expect(isEndBeforeStart('2026-12-10T14:00', '2026-12-10T13:59')).toBe(true)
    expect(isEndBeforeStart('2026-12-10T14:00', '2026-12-09')).toBe(true)
  })

  it('accepts an end equal to or after the start, or a missing bound', () => {
    expect(isEndBeforeStart('2026-12-10T14:00', '2026-12-10T14:00')).toBe(false)
    expect(isEndBeforeStart('2026-12-10T14:00', '2026-12-11T09:00')).toBe(false)
    expect(isEndBeforeStart('2026-12-10T14:00', '')).toBe(false)
    expect(isEndBeforeStart(null, '2026-12-10T14:00')).toBe(false)
  })
})
