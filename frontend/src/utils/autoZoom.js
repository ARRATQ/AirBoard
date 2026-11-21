// Taille d'affichage automatique : l'échelle du contenu suit la largeur réellement disponible
// à droite de la barre latérale. Les pages sont dessinées pour environ 1 200 px de contenu
// (écran de 1 440 px, barre ouverte) : en dessous on réduit, au-dessus on agrandit, dans des
// bornes raisonnables. Tablettes et mobiles gardent 100 % : leur mise en page s'adapte déjà.

const REFERENCE_WIDTH = 1200
const MIN_SCALE = 0.85
const MAX_SCALE = 1.15
// En dessous, la barre latérale est masquée et les pages passent en mise en page étroite.
const DESKTOP_MIN_VIEWPORT = 1025
const DESKTOP_MIN_CONTENT = 1000
const STEP = 0.05

/**
 * Échelle automatique pour une zone de contenu donnée.
 * @param {number} contentWidth largeur de la zone de contenu, en px
 * @param {number} [viewportWidth] largeur de la fenêtre, en px (si connue)
 * @returns {number} échelle, par pas de 0,05
 */
export function autoZoomScale (contentWidth, viewportWidth = Infinity) {
  if (!Number.isFinite(contentWidth) || contentWidth < DESKTOP_MIN_CONTENT) return 1
  if (viewportWidth < DESKTOP_MIN_VIEWPORT) return 1
  const raw = Math.min(MAX_SCALE, Math.max(MIN_SCALE, contentWidth / REFERENCE_WIDTH))
  return Number((Math.round(raw / STEP) * STEP).toFixed(2))
}

/**
 * Échelle pour un réglage de taille : « auto » ou un pourcentage (« 85 », « 100 »…).
 * @param {string} level
 * @param {number} contentWidth
 * @param {number} [viewportWidth]
 * @returns {number}
 */
export function zoomScaleFor (level, contentWidth, viewportWidth) {
  if (level === 'auto') return autoZoomScale(contentWidth, viewportWidth)
  const percent = Number.parseInt(level, 10)
  return Number.isFinite(percent) && percent > 0 ? percent / 100 : 1
}
