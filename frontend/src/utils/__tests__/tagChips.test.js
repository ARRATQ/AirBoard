import { describe, it, expect } from 'vitest'
import { splitTags, filterTagsToShow } from '../tagChips'

const tags = ['a', 'b', 'c', 'd', 'e'].map((name, i) => ({ id: i + 1, name }))

describe('splitTags', () => {
  it('garde les premières étiquettes et met le reste de côté', () => {
    const { shown, rest } = splitTags(tags, 3)
    expect(shown.map((t) => t.name)).toEqual(['a', 'b', 'c'])
    expect(rest.map((t) => t.name)).toEqual(['d', 'e'])
  })

  it('ne met rien de côté quand tout tient', () => {
    expect(splitTags(tags.slice(0, 2), 3)).toEqual({ shown: tags.slice(0, 2), rest: [] })
  })

  it('accepte une liste absente', () => {
    expect(splitTags(undefined, 3)).toEqual({ shown: [], rest: [] })
  })
})

describe('filterTagsToShow', () => {
  it('replié : les premières étiquettes seulement', () => {
    expect(filterTagsToShow(tags, [], 2, false).map((t) => t.id)).toEqual([1, 2])
  })

  it('replié : une étiquette choisie plus loin reste visible', () => {
    expect(filterTagsToShow(tags, [5], 2, false).map((t) => t.id)).toEqual([1, 2, 5])
  })

  it('déplié : toutes les étiquettes, dans l’ordre', () => {
    expect(filterTagsToShow(tags, [5], 2, true)).toEqual(tags)
  })
})
