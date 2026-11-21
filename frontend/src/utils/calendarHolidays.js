// Jours fériés lus dans la vue calendrier de l'agenda, pour le calculateur de jours ouvrés
// hors de la page Agenda (accueil). La période couvre l'année en cours et la suivante.
import { holidayMap } from './agenda'

/**
 * Période à demander à /events/calendar.
 * @param {Date} [today]
 * @returns {{ start: string, end: string }} dates AAAA-MM-JJ
 */
export function holidayRange (today = new Date()) {
  const year = today.getFullYear()
  return { start: `${year}-01-01`, end: `${year + 1}-12-31` }
}

/**
 * Jours fériés d'une réponse de /events/calendar (clé AAAA-MM-JJ → nom).
 * @param {{ events?: object[], Events?: object[] } | null | undefined} response
 * @returns {Map<string, string>}
 */
export function holidaysFromCalendar (response) {
  const events = response?.events || response?.Events || []
  return holidayMap(events)
}
