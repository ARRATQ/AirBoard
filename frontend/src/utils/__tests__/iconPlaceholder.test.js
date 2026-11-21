import { describe, it, expect } from 'vitest'
import { placeholderIconSet, PLACEHOLDER_BODY } from '../iconPlaceholder'

describe('placeholderIconSet', () => {
  it('répond à chaque icône demandée avec le pictogramme neutre', () => {
    const set = placeholderIconSet('https://api.iconify.design/fluent-emoji-flat.json?icons=robot%2Cwaving-hand')
    expect(set.prefix).toBe('fluent-emoji-flat')
    expect(Object.keys(set.icons)).toEqual(['robot', 'waving-hand'])
    expect(set.icons.robot.body).toBe(PLACEHOLDER_BODY)
    expect(set.width).toBe(24)
    expect(set.height).toBe(24)
  })

  it('renvoie null pour une adresse qui ne demande pas d\'icônes', () => {
    expect(placeholderIconSet('https://api.iconify.design/collections')).toBeNull()
    expect(placeholderIconSet('pas une url')).toBeNull()
  })
})
