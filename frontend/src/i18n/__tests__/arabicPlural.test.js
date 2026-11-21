import { describe, it, expect } from 'vitest'
import { arabicPluralRule } from '../arabicPlural'

const fallback = (choice, length) => (length === 2 ? (choice === 1 ? 0 : 1) : Math.min(Math.abs(choice), 2))

describe('arabicPluralRule', () => {
  it.each([
    [0, 0], // zéro
    [1, 1], // un
    [2, 2], // duel
    [3, 3], [10, 3], [103, 3], // 3 à 10 : pluriel
    [11, 4], [99, 4], [111, 4], // 11 à 99 : singulier
    [100, 5], [101, 5], [102, 5], [1000, 5] // 100, 101, 102… : autre
  ])('%i → forme %i sur 6', (count, form) => {
    expect(arabicPluralRule(count, 6, fallback)).toBe(form)
  })

  it('laisse la règle par défaut pour les messages à 2 ou 3 formes', () => {
    expect(arabicPluralRule(5, 3, fallback)).toBe(2)
    expect(arabicPluralRule(1, 2, fallback)).toBe(0)
  })
})

describe('vue-i18n avec ar.json', () => {
  it('accorde chat.unreadCount selon les six formes', async () => {
    const { createI18n } = await import('vue-i18n')
    const ar = (await import('@/locales/ar.json')).default
    const i18n = createI18n({ legacy: false, locale: 'ar', messages: { ar }, pluralRules: { ar: arabicPluralRule } })
    const t = (n) => i18n.global.t('chat.unreadCount', { count: n })
    expect(t(0)).toBe('لا توجد رسائل غير مقروءة')
    expect(t(1)).toBe('رسالة واحدة غير مقروءة')
    expect(t(2)).toBe('رسالتان غير مقروءتين')
    expect(t(5)).toBe('5 رسائل غير مقروءة')
    expect(t(11)).toBe('11 رسالة غير مقروءة')
    expect(t(100)).toBe('100 رسالة غير مقروءة')
  })
})
