/**
 * Message lisible d'une erreur d'API.
 *
 * Le backend répond toujours `{ error: "Message lisible", code: "not_found" }`.
 * `message` reste lu pour les réponses d'avant ce format (et les erreurs réseau
 * n'ont pas de corps) : on retombe alors sur `fallback`.
 *
 * @param {unknown} error  erreur levée par axios
 * @param {string} fallback  message affiché à défaut
 * @returns {string}
 */
export function apiErrorMessage(error, fallback) {
  const data = error?.response?.data
  for (const candidate of [data?.error, data?.message]) {
    if (typeof candidate === 'string' && candidate.trim() !== '') return candidate
  }
  return fallback
}
