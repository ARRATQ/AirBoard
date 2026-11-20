import { createI18n } from 'vue-i18n'
import fr from '@/locales/fr.json'

// i18n réel (messages fr) pour que les libellés testés soient ceux de l'application.
export const makeI18n = () =>
  createI18n({ legacy: false, locale: 'fr', fallbackLocale: 'fr', messages: { fr } })

// Headless UI rend le Dialog dans un portail : on interroge le document.
export const $ = (selector) => document.body.querySelector(selector)
export const $$ = (selector) => [...document.body.querySelectorAll(selector)]

export const flush = () => new Promise((resolve) => setTimeout(resolve, 0))
