// Conversions pour les champs <input type="datetime-local">, qui affichent et
// renvoient une heure LOCALE sans fuseau (« AAAA-MM-JJTHH:mm »).
// toISOString() donnerait l'heure UTC : relue comme locale, la date reculerait
// du décalage horaire à chaque enregistrement.

const pad = (n) => String(n).padStart(2, '0')

/**
 * Instant stocké (ISO, Date…) → valeur d'un champ datetime-local en heure locale.
 * @param {string|number|Date|null|undefined} value
 * @param {string|null} [empty=''] valeur renvoyée si la date est absente ou invalide
 * @returns {string|null}
 */
export function toDateTimeInput(value, empty = '') {
  if (!value) return empty
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return empty
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

/**
 * Valeur d'un champ datetime-local (heure locale) → chaîne ISO UTC pour l'API.
 * @param {string|null|undefined} value
 * @returns {string|null}
 */
export function fromDateTimeInput(value) {
  if (!value) return null
  const d = new Date(value)
  return Number.isNaN(d.getTime()) ? null : d.toISOString()
}

/**
 * La fin d'une période précède son début (bornes au format des champs date ou datetime-local).
 * Une borne absente ou invalide n'est pas une erreur ; une fin égale au début est acceptée.
 * @param {string|Date|null|undefined} start
 * @param {string|Date|null|undefined} end
 * @returns {boolean}
 */
export function isEndBeforeStart(start, end) {
  if (!start || !end) return false
  const s = new Date(start).getTime()
  const e = new Date(end).getTime()
  return !Number.isNaN(s) && !Number.isNaN(e) && e < s
}
