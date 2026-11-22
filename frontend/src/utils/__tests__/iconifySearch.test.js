import { describe, it, expect, vi } from 'vitest'
import { buildSearchUrl, isIconName, parseSearchResponse, searchIcons, SEARCH_LIMIT } from '../iconifySearch'

describe('isIconName', () => {
  it('accepte collection:nom', () => {
    expect(isIconName('mdi:home')).toBe(true)
    expect(isIconName('material-symbols:home-outline')).toBe(true)
    expect(isIconName('tabler:home_2')).toBe(true)
  })

  it('refuse le reste', () => {
    expect(isIconName('home')).toBe(false)
    expect(isIconName('mdi:')).toBe(false)
    expect(isIconName('mdi:home"><script>')).toBe(false)
    expect(isIconName(42)).toBe(false)
  })
})

describe('buildSearchUrl', () => {
  it('requête vide : pas d’adresse', () => {
    expect(buildSearchUrl('   ')).toBeNull()
  })

  it('encode la requête et fixe la limite', () => {
    const url = new URL(buildSearchUrl('maison & co'))
    expect(url.origin).toBe('https://api.iconify.design')
    expect(url.searchParams.get('query')).toBe('maison & co')
    expect(url.searchParams.get('limit')).toBe(String(SEARCH_LIMIT))
    expect(url.searchParams.has('prefixes')).toBe(false)
  })

  it('limite aux collections du style', () => {
    const url = new URL(buildSearchUrl('home', 'line'))
    expect(url.searchParams.get('prefixes')).toBe('lucide,tabler,iconoir')
  })
})

describe('parseSearchResponse', () => {
  it('garde les noms valides, sans doublon', () => {
    expect(parseSearchResponse({ icons: ['lucide:home', 'bad', 'lucide:home', 'tabler:x'] })).toEqual(['lucide:home', 'tabler:x'])
  })

  it('réponse inattendue', () => {
    expect(parseSearchResponse(null)).toEqual([])
    expect(parseSearchResponse({ icons: 'x' })).toEqual([])
  })
})

describe('searchIcons', () => {
  it('interroge Iconify et renvoie les noms', async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: true, json: () => Promise.resolve({ icons: ['lucide:home'] }) })
    await expect(searchIcons('home', 'all', { fetchImpl })).resolves.toEqual(['lucide:home'])
    expect(fetchImpl).toHaveBeenCalledOnce()
  })

  it('erreur HTTP', async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: false, status: 503 })
    await expect(searchIcons('home', 'all', { fetchImpl })).rejects.toThrow('503')
  })

  it('requête vide : aucun appel', async () => {
    const fetchImpl = vi.fn()
    await expect(searchIcons('', 'all', { fetchImpl })).resolves.toEqual([])
    expect(fetchImpl).not.toHaveBeenCalled()
  })
})
