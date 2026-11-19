// Utilitaires de couleur purs (hex #rrggbb), partagés par les palettes et les tokens.

const HEX_RE = /^#[0-9A-Fa-f]{6}$/

export const isHex = (value) => typeof value === 'string' && HEX_RE.test(value)

export function hexToRgb(hex) {
  return {
    r: parseInt(hex.slice(1, 3), 16),
    g: parseInt(hex.slice(3, 5), 16),
    b: parseInt(hex.slice(5, 7), 16),
  }
}

const clampByte = (v) => Math.max(0, Math.min(255, Math.round(v)))

export function rgbToHex(r, g, b) {
  return '#' + [r, g, b].map(v => clampByte(v).toString(16).padStart(2, '0')).join('')
}

export function hexToHsl(hex) {
  if (!isHex(hex)) return { h: 18, s: 67, l: 59 }
  const { r: r255, g: g255, b: b255 } = hexToRgb(hex)
  const r = r255 / 255
  const g = g255 / 255
  const b = b255 / 255
  const max = Math.max(r, g, b)
  const min = Math.min(r, g, b)
  const l = (max + min) / 2
  if (max === min) return { h: 0, s: 0, l: l * 100 }
  const d = max - min
  const s = l > 0.5 ? d / (2 - max - min) : d / (max + min)
  let h
  if (max === r) h = ((g - b) / d + (g < b ? 6 : 0)) / 6
  else if (max === g) h = ((b - r) / d + 2) / 6
  else h = ((r - g) / d + 4) / 6
  return { h: h * 360, s: s * 100, l: l * 100 }
}

export function hslToHex(h, s, l) {
  const hue = ((h % 360) + 360) % 360
  const sat = Math.max(0, Math.min(100, s)) / 100
  const lig = Math.max(0, Math.min(100, l)) / 100
  const c = (1 - Math.abs(2 * lig - 1)) * sat
  const x = c * (1 - Math.abs(((hue / 60) % 2) - 1))
  const m = lig - c / 2
  const [r, g, b] =
    hue < 60 ? [c, x, 0]
      : hue < 120 ? [x, c, 0]
        : hue < 180 ? [0, c, x]
          : hue < 240 ? [0, x, c]
            : hue < 300 ? [x, 0, c]
              : [c, 0, x]
  return rgbToHex((r + m) * 255, (g + m) * 255, (b + m) * 255)
}

// Mélange linéaire en sRGB : weight = part de `b` (0 → a, 1 → b).
export function mixHex(a, b, weight) {
  const ca = hexToRgb(a)
  const cb = hexToRgb(b)
  return rgbToHex(
    ca.r + (cb.r - ca.r) * weight,
    ca.g + (cb.g - ca.g) * weight,
    ca.b + (cb.b - ca.b) * weight,
  )
}

export function lighten(hex, amount) {
  const { r, g, b } = hexToRgb(hex)
  return rgbToHex(r + amount, g + amount, b + amount)
}

export function darken(hex, amount) {
  return lighten(hex, -amount)
}

function relativeLuminance(hex) {
  const lin = (v) => {
    const c = v / 255
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
  }
  const { r, g, b } = hexToRgb(hex)
  return 0.2126 * lin(r) + 0.7152 * lin(g) + 0.0722 * lin(b)
}

// Rapport de contraste WCAG entre deux couleurs (1 à 21).
export function contrastRatio(a, b) {
  const la = relativeLuminance(a)
  const lb = relativeLuminance(b)
  const [hi, lo] = la > lb ? [la, lb] : [lb, la]
  return (hi + 0.05) / (lo + 0.05)
}

// Fait varier la luminosité HSL de `hex` (pas de 1 %) jusqu'à atteindre `ratio` contre `against`.
export function adjustForContrast(hex, against, ratio, direction) {
  const { h, s, l } = hexToHsl(hex)
  const step = direction === 'darker' ? -1 : 1
  for (let lig = l; lig >= 0 && lig <= 100; lig += step) {
    const candidate = hslToHex(h, s, lig)
    if (contrastRatio(candidate, against) >= ratio) return candidate
  }
  return direction === 'darker' ? '#000000' : '#ffffff'
}

export const toChannels = (hex) => {
  const { r, g, b } = hexToRgb(hex)
  return `${r} ${g} ${b}`
}
