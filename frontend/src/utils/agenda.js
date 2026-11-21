// Calculs de l'agenda : dates en heure locale (clé AAAA-MM-JJ), jours fériés, jours ouvrés,
// événements d'un jour. Aucune dépendance à Vue, testés dans __tests__/agenda.test.js.

/** Samedi et dimanche (Date.getDay()). */
export const DEFAULT_WEEKEND = [0, 6]

/** Au-delà, le calcul de jours ouvrés est refusé (10 ans). */
export const MAX_RANGE_DAYS = 3660

const pad = (n) => String(n).padStart(2, '0')

/** Clé AAAA-MM-JJ d'une date, en heure locale (jamais toISOString, qui passe en UTC). */
export const dateKey = (date) => `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`

/** AAAA-MM-JJ → Date locale à minuit, ou null. */
export function parseDateKey(key) {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(key || '')
  if (!m) return null
  const d = new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]))
  return d.getMonth() === Number(m[2]) - 1 ? d : null
}

/**
 * Jours couverts par un événement : [clé de début, clé de fin]. Journée entière et férié :
 * la date du calendrier telle qu'elle est saisie (pas de décalage de fuseau) ; sinon l'heure locale.
 */
export function eventDayRange(event) {
  if (!event?.start_date) return null
  const calendarDate = event.is_all_day || event.is_holiday
  const keyOf = (value) => {
    if (calendarDate && /^\d{4}-\d{2}-\d{2}/.test(value)) return String(value).slice(0, 10)
    const d = new Date(value)
    return Number.isNaN(d.getTime()) ? null : dateKey(d)
  }
  const start = keyOf(event.start_date)
  if (!start) return null
  const end = (event.end_date && keyOf(event.end_date)) || start
  return [start, end < start ? start : end]
}

/** L'événement couvre-t-il ce jour (clé AAAA-MM-JJ) ? */
export function eventCoversDay(event, key) {
  const range = eventDayRange(event)
  return !!range && range[0] <= key && key <= range[1]
}

/** Map clé → nom du férié, pour tous les jours couverts par les événements is_holiday. */
export function holidayMap(events) {
  const map = new Map()
  for (const event of events || []) {
    if (!event?.is_holiday) continue
    const range = eventDayRange(event)
    if (!range) continue
    const day = parseDateKey(range[0])
    for (let i = 0; day && dateKey(day) <= range[1] && i < 366; i++) {
      const key = dateKey(day)
      if (!map.has(key)) map.set(key, event.title)
      day.setDate(day.getDate() + 1)
    }
  }
  return map
}

/**
 * Jours ouvrés entre deux dates incluses (clés AAAA-MM-JJ). Un férié tombant un week-end
 * compte comme week-end. Renvoie null si une date manque, si la fin précède le début ou si
 * la période dépasse MAX_RANGE_DAYS.
 */
export function countWorkingDays({ start, end, weekendDays = DEFAULT_WEEKEND, holidays = new Map(), excludeHolidays = true }) {
  const from = parseDateKey(start)
  const to = parseDateKey(end)
  if (!from || !to || from > to) return null

  const result = { totalDays: 0, workingDays: 0, nonWorkingDays: 0, weekendDays: 0, holidayDays: 0, holidays: [] }
  const day = new Date(from)
  while (day <= to) {
    if (result.totalDays >= MAX_RANGE_DAYS) return null
    result.totalDays++
    const key = dateKey(day)
    if (weekendDays.includes(day.getDay())) {
      result.weekendDays++
    } else if (excludeHolidays && holidays.has(key)) {
      result.holidayDays++
      result.holidays.push({ date: key, title: holidays.get(key) })
    } else {
      result.workingDays++
    }
    day.setDate(day.getDate() + 1)
  }
  result.nonWorkingDays = result.totalDays - result.workingDays
  return result
}

/** Ton d'un statut d'événement : seules les exceptions sont colorées. */
export function eventStatusTone(status) {
  if (status === 'tentative') return 'warn'
  if (status === 'cancelled') return 'bad'
  return ''
}
