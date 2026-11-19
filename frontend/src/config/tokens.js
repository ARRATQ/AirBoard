// Tokens de la direction A+ dérivés d'une palette (voir plans/refonte-airboard.md, « Direction A+ »).
// Les tokens neutres (surface, line, ok, board…) sont fixes et vivent dans main.css ; seuls
// l'accent et ce qui en dépend sont générés ici, pour chaque palette, palette personnalisée comprise.

import { adjustForContrast, contrastRatio, hexToRgb, isHex, mixHex, toChannels } from './color'

const WHITE = '#ffffff'
const DARK_SURFACE = '#15181b'
const MIN_TEXT_CONTRAST = 4.5

// Variante foncée de l'accent : lisible en blanc dessus et en texte sur fond blanc.
export function inkForLight(accent) {
  if (contrastRatio(accent, WHITE) >= MIN_TEXT_CONTRAST) return accent
  return adjustForContrast(accent, WHITE, MIN_TEXT_CONTRAST, 'darker')
}

// En sombre, l'« encre » est une variante claire, lisible en texte sur la surface sombre.
export function inkForDark(accent) {
  return adjustForContrast(accent, DARK_SURFACE, 8, 'lighter')
}

// Échelle primary-50…950 calquée sur l'accent clair ; 600 = encre (texte blanc lisible).
export function primaryScale(accent, ink) {
  return {
    50: mixHex(accent, WHITE, 0.92),
    100: mixHex(accent, WHITE, 0.84),
    200: mixHex(accent, WHITE, 0.68),
    300: mixHex(accent, WHITE, 0.48),
    400: mixHex(accent, WHITE, 0.24),
    500: accent,
    600: ink,
    700: mixHex(ink, '#000000', 0.18),
    800: mixHex(ink, '#000000', 0.34),
    900: mixHex(ink, '#000000', 0.5),
    950: mixHex(ink, '#000000', 0.66),
  }
}

const declarations = (vars) =>
  Object.entries(vars).map(([name, value]) => `  --${name}: ${value};`).join('\n')

function accentVars(mode, isDark) {
  const acc = isHex(mode.accent) ? mode.accent : '#d97757'
  const ink = isHex(mode.accentInk) ? mode.accentInk : (isDark ? inkForDark(acc) : inkForLight(acc))
  const { r, g, b } = hexToRgb(acc)
  return {
    acc,
    'acc-ink': ink,
    // Variables historiques encore lues par les composants existants.
    accent: acc,
    'accent-light': mode.accentLight || mixHex(acc, WHITE, 0.24),
    'accent-dark': ink,
    'accent-rgb': `${r}, ${g}, ${b}`,
    'shadow-rgb': mode.shadowRgb || (isDark ? '0, 0, 0' : '18, 21, 24'),
    'gradient-from': mode.gradientFrom || ink,
    'gradient-mid': mode.gradientMid || acc,
    'gradient-to': mode.gradientTo || ink,
  }
}

/**
 * Feuille de style des variables d'une palette.
 * @param {{ light: object, dark: object, customBg?: boolean }} palette
 * @returns {string}
 */
export function paletteCSS(palette) {
  const light = accentVars(palette.light, false)
  const dark = accentVars(palette.dark, true)
  const scale = primaryScale(light.acc, light['acc-ink'])
  const primary = Object.fromEntries(
    Object.entries(scale).map(([step, hex]) => [`primary-${step}`, toChannels(hex)]),
  )
  // Fond choisi explicitement pour une palette personnalisée : on le respecte en clair.
  const bg = palette.customBg && isHex(palette.light.bgPage) ? { bg: palette.light.bgPage } : {}
  return `:root {\n${declarations({ ...light, ...primary, ...bg })}\n}\n`
    + `html.dark {\n${declarations(dark)}\n}`
}
