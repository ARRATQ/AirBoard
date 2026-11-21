// Miroir de la politique de mots de passe du backend
// (backend/utils/auth_security.go, ValidatePassword), pour prévenir avant l'envoi.

export const PASSWORD_MIN_LENGTH = 8
export const PASSWORD_MAX_LENGTH = 128

const SPECIAL_CHARS = '!@#$%^&*()_+-=[]{}|;:,.<>?'
const BANNED_WORDS = [
  'password', 'admin', 'airboard', 'user', '123456',
  'qwerty', 'letmein', 'welcome', 'monkey', 'dragon'
]

/**
 * Premier défaut du mot de passe, dans l'ordre où le backend le vérifie.
 * La longueur est comptée en octets UTF-8, comme `len()` en Go.
 * @param {string} password
 * @returns {null | 'tooShort' | 'tooLong' | 'upper' | 'lower' | 'digit' | 'special' | 'common'}
 */
export function passwordProblem (password) {
  const value = password || ''
  const bytes = new TextEncoder().encode(value).length
  if (bytes < PASSWORD_MIN_LENGTH) return 'tooShort'
  if (bytes > PASSWORD_MAX_LENGTH) return 'tooLong'
  if (!/[A-Z]/.test(value)) return 'upper'
  if (!/[a-z]/.test(value)) return 'lower'
  if (!/[0-9]/.test(value)) return 'digit'
  if (![...value].some((c) => SPECIAL_CHARS.includes(c))) return 'special'
  const lower = value.toLowerCase()
  if (BANNED_WORDS.some((w) => lower.includes(w))) return 'common'
  return null
}
