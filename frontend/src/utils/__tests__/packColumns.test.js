import { describe, it, expect } from 'vitest'
import { packColumns, columnCountFor } from '../packColumns'

describe('packColumns', () => {
  it('place chaque bloc dans la colonne la plus courte, dans l’ordre', () => {
    // A long, B et C courts : C passe sous B plutôt que sous A
    const heights = new Map([['A', 500], ['B', 100], ['C', 100], ['D', 100]])
    expect(packColumns(['A', 'B', 'C', 'D'], heights, 2)).toEqual([['A'], ['B', 'C', 'D']])
  })

  it('à hauteur égale, garde l’ordre de lecture (colonne de début d’abord)', () => {
    const heights = new Map([['A', 100], ['B', 100], ['C', 100], ['D', 100]])
    expect(packColumns(['A', 'B', 'C', 'D'], heights, 2)).toEqual([['A', 'C'], ['B', 'D']])
  })

  it('compte l’écart entre les blocs', () => {
    // A + C : 110 sans écart (< 120, D va sous C) ; 170 avec 60 d'écart (> 120, D va sous B)
    const heights = new Map([['A', 100], ['B', 120], ['C', 10], ['D', 10]])
    expect(packColumns(['A', 'B', 'C', 'D'], heights, 2, 0)).toEqual([['A', 'C', 'D'], ['B']])
    expect(packColumns(['A', 'B', 'C', 'D'], heights, 2, 60)).toEqual([['A', 'C'], ['B', 'D']])
  })

  it('utilise une estimation quand la hauteur n’est pas encore mesurée', () => {
    const estimate = (id) => (id === 'A' ? 400 : 50)
    expect(packColumns(['A', 'B', 'C'], new Map(), 2, 0, estimate)).toEqual([['A'], ['B', 'C']])
  })

  it('une seule colonne, ou aucune entrée', () => {
    expect(packColumns(['A', 'B'], new Map(), 1)).toEqual([['A', 'B']])
    expect(packColumns([], new Map(), 3)).toEqual([[], [], []])
  })
})

describe('columnCountFor', () => {
  it.each([
    [320, 1], [599, 1], [616, 2], [1000, 3], [1136, 3], [1400, 4]
  ])('%ipx → %i colonne(s) de 300px minimum avec 16px d’écart', (width, count) => {
    expect(columnCountFor(width, 300, 16)).toBe(count)
  })

  it('au moins une colonne même à largeur nulle', () => {
    expect(columnCountFor(0, 300, 16)).toBe(1)
  })
})
