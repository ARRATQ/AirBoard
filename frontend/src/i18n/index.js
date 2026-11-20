import { createI18n } from 'vue-i18n'

const DEFAULT_LOCALE_KEY = 'airboard_locale'
const FALLBACK_LOCALE = 'en'

// Chaque langue est un chunk à part, chargé à la demande (et non plus les quatre au démarrage).
const LOCALE_LOADERS = {
  fr: () => import('@/locales/fr.json'),
  en: () => import('@/locales/en.json'),
  es: () => import('@/locales/es.json'),
  ar: () => import('@/locales/ar.json'),
}

export const SUPPORTED_LOCALES = Object.keys(LOCALE_LOADERS)

/** Ramène une valeur quelconque (localStorage, profil) à une langue prise en charge. */
export function normalizeLocale(locale) {
  return SUPPORTED_LOCALES.includes(locale) ? locale : FALLBACK_LOCALE
}

export function resolveInitialLocale() {
  const saved = localStorage.getItem(DEFAULT_LOCALE_KEY)
  if (saved) return normalizeLocale(saved)
  const browser = navigator.language?.toLowerCase() || 'en'
  if (browser.startsWith('fr')) return 'fr'
  if (browser.startsWith('es')) return 'es'
  if (browser.startsWith('ar')) return 'ar'
  return 'en'
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: resolveInitialLocale(),
  fallbackLocale: FALLBACK_LOCALE,
  messages: {},
})

const pendingLoads = new Map()

/**
 * Charge les messages d'une langue une seule fois (les appels concurrents partagent la promesse).
 * @param {string} locale langue prise en charge
 * @returns {Promise<void>}
 */
export function loadLocaleMessages(locale) {
  if (!pendingLoads.has(locale)) {
    const load = LOCALE_LOADERS[locale]()
      .then((module) => i18n.global.setLocaleMessage(locale, module.default))
      .catch((error) => {
        pendingLoads.delete(locale) // nouvel essai possible au prochain changement de langue
        throw error
      })
    pendingLoads.set(locale, load)
  }
  return pendingLoads.get(locale)
}

/** La police arabe n'est demandée à Google Fonts que pour l'interface en arabe. */
function loadArabicFont() {
  if (document.getElementById('font-arabic')) return
  const link = document.createElement('link')
  link.id = 'font-arabic'
  link.rel = 'stylesheet'
  link.href = 'https://fonts.googleapis.com/css2?family=IBM+Plex+Sans+Arabic:wght@400;500;600&display=swap'
  document.head.appendChild(link)
}

export function applyDirectionForLocale(locale) {
  const html = document.documentElement
  html.setAttribute('dir', locale === 'ar' ? 'rtl' : 'ltr')
  html.setAttribute('lang', locale)
  if (locale === 'ar') loadArabicFont()
}

let requestedLocale = null

/** Active une langue une fois ses messages chargés ; l'anglais de secours suit en arrière-plan. */
async function activateLocale(locale) {
  requestedLocale = locale
  try {
    await loadLocaleMessages(locale)
  } catch (error) {
    console.error(`Langue « ${locale} » non chargée :`, error)
    return
  }
  // Deux changements rapprochés : seule la dernière langue demandée s'applique.
  if (requestedLocale !== locale) return
  i18n.global.locale.value = locale
  applyDirectionForLocale(locale)
  if (locale !== FALLBACK_LOCALE) {
    // Clés absentes de certaines langues (réglages OAuth) : pas besoin d'attendre pour afficher.
    loadLocaleMessages(FALLBACK_LOCALE).catch((error) => {
      console.warn('Langue de secours non chargée :', error)
    })
  }
}

/**
 * Change de langue et la mémorise.
 * @param {string} locale langue demandée
 * @returns {Promise<void>}
 */
export function setLocale(locale) {
  const target = normalizeLocale(locale)
  localStorage.setItem(DEFAULT_LOCALE_KEY, target)
  return activateLocale(target)
}

/**
 * À attendre avant le montage : charge la langue initiale.
 * @returns {Promise<void>}
 */
export function initI18n() {
  return activateLocale(i18n.global.locale.value)
}

// Direction et police dès le chargement du module, avant le premier rendu
applyDirectionForLocale(i18n.global.locale.value)

export default i18n
