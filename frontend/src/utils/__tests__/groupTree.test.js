import { describe, it, expect } from 'vitest'
import { orderGroupTree, descendantIds, parentCandidates, indentedLabel } from '../groupTree'

const groups = [
  { id: 4, name: 'Zeta', parent_id: null },
  { id: 1, name: 'Direction', parent_id: null },
  { id: 3, name: 'Service B', parent_id: 1 },
  { id: 2, name: 'Service A', parent_id: 1 },
  { id: 5, name: 'Équipe', parent_id: 2 },
  { id: 6, name: 'Tous', parent_id: null, is_default: true },
]

describe('orderGroupTree', () => {
  it('place chaque parent avant ses sous-groupes, triés par nom', () => {
    const rows = orderGroupTree(groups)
    expect(rows.map(r => r.group.id)).toEqual([1, 2, 5, 3, 6, 4])
    expect(rows.find(r => r.group.id === 5)).toMatchObject({ depth: 2, path: ['Direction', 'Service A'] })
  })

  it('traite un parent absent de la liste comme une racine', () => {
    const rows = orderGroupTree([{ id: 9, name: 'Orphelin', parent_id: 42 }])
    expect(rows).toEqual([{ group: { id: 9, name: 'Orphelin', parent_id: 42 }, depth: 0, path: [] }])
  })

  it('affiche quand même des groupes pris dans une boucle', () => {
    const rows = orderGroupTree([{ id: 1, name: 'A', parent_id: 2 }, { id: 2, name: 'B', parent_id: 1 }])
    expect(rows.map(r => r.group.id).sort()).toEqual([1, 2])
  })

  it('accepte une liste absente et ne modifie pas la liste reçue', () => {
    expect(orderGroupTree(null)).toEqual([])
    const copy = groups.map(g => g.id)
    orderGroupTree(groups)
    expect(groups.map(g => g.id)).toEqual(copy)
  })
})

describe('descendantIds / parentCandidates / indentedLabel', () => {
  it('renvoie le groupe et tous ses sous-groupes', () => {
    expect([...descendantIds(groups, 1)].sort()).toEqual([1, 2, 3, 5])
  })

  it('exclut le groupe, ses sous-groupes et le groupe par défaut des parents possibles', () => {
    expect(parentCandidates(groups, 2).map(r => r.group.id)).toEqual([1, 3, 4])
    expect(parentCandidates(groups, null).map(r => r.group.id)).toEqual([1, 2, 5, 3, 4])
  })

  it('indente le libellé selon la profondeur', () => {
    expect(indentedLabel({ group: { name: 'A' }, depth: 0 })).toBe('A')
    expect(indentedLabel({ group: { name: 'B' }, depth: 2 })).toBe(' '.repeat(6) + '└ B')
  })
})
