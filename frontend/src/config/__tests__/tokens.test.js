import { describe, expect, it } from 'vitest'
import { contrastRatio, mixHex } from '../color'
import { inkForDark, inkForLight, paletteCSS, primaryScale } from '../tokens'
import { PALETTES, generateCustomPalette, generatePaletteFromColors } from '../palettes'

describe('color', () => {
  it('computes WCAG contrast', () => {
    expect(contrastRatio('#000000', '#ffffff')).toBeCloseTo(21, 0)
    expect(contrastRatio('#ffffff', '#ffffff')).toBeCloseTo(1, 5)
  })

  it('mixes two colors', () => {
    expect(mixHex('#000000', '#ffffff', 0.5)).toBe('#808080')
    expect(mixHex('#d97757', '#ffffff', 0)).toBe('#d97757')
  })
})

describe('tokens', () => {
  it.each(['#d97757', '#ffd400', '#10b981', '#e84f7a', '#5c6ef8'])(
    'derives readable inks from accent %s',
    (accent) => {
      expect(contrastRatio(inkForLight(accent), '#ffffff')).toBeGreaterThanOrEqual(4.5)
      expect(contrastRatio(inkForDark(accent), '#15181b')).toBeGreaterThanOrEqual(4.5)
    },
  )

  it('keeps the hand-picked inks of the validated palettes', () => {
    const css = paletteCSS(PALETTES.find(p => p.slug === 'indigo'))
    expect(css).toContain('--acc-ink: #3d4ed4;')
    expect(css).toContain('--acc-ink: #a9b3fd;')
  })

  it('builds a primary scale where 500 is the accent and 600 the ink', () => {
    const scale = primaryScale('#d97757', '#a8462f')
    expect(scale[500]).toBe('#d97757')
    expect(scale[600]).toBe('#a8462f')
    expect(Object.keys(scale)).toHaveLength(11)
  })

  it('emits primary channels for Tailwind <alpha-value>', () => {
    expect(paletteCSS(PALETTES[0])).toContain('--primary-500: 217 119 87;')
  })

  it('respects a background chosen for a custom palette, not a derived one', () => {
    expect(paletteCSS(generatePaletteFromColors('#5c6ef8', '#f0f1ff'))).toContain('--bg: #f0f1ff;')
    expect(paletteCSS(generateCustomPalette('#5c6ef8'))).not.toContain('--bg:')
  })
})
