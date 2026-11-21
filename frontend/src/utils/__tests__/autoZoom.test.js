import { describe, it, expect } from 'vitest'
import { autoZoomScale, zoomScaleFor } from '../autoZoom'

describe('autoZoomScale', () => {
  it.each([
    [1184, 1], // 1440 px avec la barre latérale ouverte : taille de référence
    [1024, 0.85], // portable 1280 px : réduit
    [1110, 0.95], // portable 1366 px
    [1280, 1.05], // 1536 px
    [1664, 1.15], // 1920 px : agrandi, plafonné
    [2304, 1.15], // 2560 px : plafond
  ])('zone de %ipx → %s', (width, scale) => {
    expect(autoZoomScale(width)).toBe(scale)
  })

  it('ne touche pas aux tablettes et mobiles (mise en page déjà adaptée)', () => {
    expect(autoZoomScale(375)).toBe(1)
    expect(autoZoomScale(960)).toBe(1)
    // Tablette 1024 px : barre latérale masquée, contenu pleine largeur, mais 100 %
    expect(autoZoomScale(1024, 1024)).toBe(1)
  })

  it('largeur inconnue : 100 %', () => {
    expect(autoZoomScale(0)).toBe(1)
    expect(autoZoomScale(Number.NaN)).toBe(1)
  })
})

describe('zoomScaleFor', () => {
  it('niveau manuel : le pourcentage choisi', () => {
    expect(zoomScaleFor('85', 1664)).toBe(0.85)
    expect(zoomScaleFor('115', 1024)).toBe(1.15)
  })

  it('auto : suit la largeur', () => {
    expect(zoomScaleFor('auto', 1024)).toBe(0.85)
  })

  it('valeur inconnue : 100 %', () => {
    expect(zoomScaleFor('abc', 1024)).toBe(1)
  })
})
