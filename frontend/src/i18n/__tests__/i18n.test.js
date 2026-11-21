import { beforeEach, describe, expect, it } from 'vitest'
import { i18n, initI18n, loadLocaleMessages, normalizeLocale, setLocale } from '../index'

describe('i18n à la demande', () => {
  beforeEach(() => localStorage.clear())

  it('ramène une langue inconnue à l’anglais', () => {
    expect(normalizeLocale('fr')).toBe('fr')
    expect(normalizeLocale('de')).toBe('en')
    expect(normalizeLocale(undefined)).toBe('en')
  })

  it('ne charge que la langue demandée, puis l’anglais de secours', async () => {
    await setLocale('es')
    expect(i18n.global.locale.value).toBe('es')
    expect(i18n.global.availableLocales).toContain('es')
    expect(i18n.global.availableLocales).not.toContain('ar')
    await loadLocaleMessages('en')
    expect(i18n.global.t('common.enabled')).toBeTruthy()
    expect(localStorage.getItem('airboard_locale')).toBe('es')
  })

  it('partage un chargement en cours', () => {
    expect(loadLocaleMessages('fr')).toBe(loadLocaleMessages('fr'))
  })

  it('applique la dernière langue demandée et sa direction', async () => {
    await Promise.all([setLocale('ar'), setLocale('fr')])
    expect(i18n.global.locale.value).toBe('fr')
    expect(document.documentElement.getAttribute('dir')).toBe('ltr')
    expect(document.documentElement.getAttribute('lang')).toBe('fr')
  })

  it('passe en RTL et charge la police arabe pour l’arabe', async () => {
    await setLocale('ar')
    expect(document.documentElement.getAttribute('dir')).toBe('rtl')
    expect(document.getElementById('font-arabic')?.getAttribute('href')).toContain('IBM+Plex+Sans+Arabic')
  })

  it('initI18n ne mémorise pas la langue détectée', async () => {
    await initI18n()
    expect(localStorage.getItem('airboard_locale')).toBeNull()
  })
})
