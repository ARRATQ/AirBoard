import { describe, expect, it } from 'vitest'
import { eventStatusTone, suggestionStatusTone } from '../status'

describe('status tones', () => {
  it('maps event statuses', () => {
    expect(eventStatusTone('confirmed')).toBe('ok')
    expect(eventStatusTone('tentative')).toBe('warn')
    expect(eventStatusTone('cancelled')).toBe('bad')
  })

  it('maps suggestion statuses', () => {
    expect(suggestionStatusTone('implemented')).toBe('ok')
    expect(suggestionStatusTone('rejected')).toBe('bad')
    expect(suggestionStatusTone('in_progress')).toBe('warn')
  })

  it('falls back to neutral for an unknown status', () => {
    expect(eventStatusTone(undefined)).toBe('neutral')
    expect(suggestionStatusTone('archived')).toBe('neutral')
  })
})
