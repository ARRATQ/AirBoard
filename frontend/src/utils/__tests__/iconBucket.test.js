import { describe, expect, it } from 'vitest'
import { ICON_BUCKETS, extractIconNames, iconBucket } from '../iconBucket'

describe('iconBucket', () => {
  it.each(Object.entries(ICON_BUCKETS))('is stable and stays within range for %s', (_, count) => {
    expect(iconBucket('home', count)).toBe(iconBucket('home', count))
    for (const name of ['home', 'account-circle', 'a', '', 'x'.repeat(200)]) {
      const bucket = iconBucket(name, count)
      expect(bucket).toBeGreaterThanOrEqual(0)
      expect(bucket).toBeLessThan(count)
    }
  })

  it('spreads names over most buckets', () => {
    const names = Array.from({ length: 5000 }, (_, i) => `icon-${i}`)
    const used = new Set(names.map((name) => iconBucket(name, 256)))
    expect(used.size).toBeGreaterThan(200)
  })
})

describe('extractIconNames', () => {
  it('finds icons of the embedded collections in templates and scripts', () => {
    const source = `
      <Icon icon="mdi:home" />
      const a = 'lucide:layout-dashboard'
      const b = { icon: "heroicons:squares-2x2" }
      placeholder="mdi:application, lucide:app-window"
    `
    expect(extractIconNames(source)).toEqual([
      ['mdi', 'home'],
      ['lucide', 'layout-dashboard'],
      ['heroicons', 'squares-2x2'],
      ['mdi', 'application'],
      ['lucide', 'app-window'],
    ])
  })

  it('ignores other collections, dynamic names and CSS classes', () => {
    const source = 'icon="carbon:home" `mdi:${name}` class="mdi-home" xmdi:home'
    expect(extractIconNames(source)).toEqual([])
  })
})
