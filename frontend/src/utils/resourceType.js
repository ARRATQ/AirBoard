// Type d'une ressource du portail, déduit de son URL (aucun champ en base) :
// - liens SharePoint/OneDrive : marqueur /:w:/ Word, /:x:/ Excel, /:p:/ PowerPoint, /:b:/ PDF, /:f:/ dossier ;
// - lien direct vers un fichier : extension .doc(x), .xls(x), .ppt(x), .pdf ;
// - chemin contenant /sites/ : site intranet ;
// - sinon : appli si l'URL s'arrête au nom d'hôte, lien web si elle pointe plus loin.

export const NEW_RESOURCE_DAYS = 14
const DAY_MS = 86_400_000

const SHAREPOINT_MARKERS = { w: 'word', x: 'excel', p: 'powerpoint', b: 'pdf', f: 'folder' }
const EXTENSIONS = [
  [/\.docx?$/i, 'word'],
  [/\.xlsx?$/i, 'excel'],
  [/\.pptx?$/i, 'powerpoint'],
  [/\.pdf$/i, 'pdf'],
]

/** @typedef {'app'|'word'|'excel'|'powerpoint'|'pdf'|'folder'|'site'|'link'} ResourceType */

/**
 * @param {string} url
 * @returns {ResourceType}
 */
export function resourceTypeOf(url) {
  let parsed
  try {
    parsed = new URL(url)
  } catch {
    return 'link'
  }
  const path = decodeURIComponent(parsed.pathname)

  const marker = path.match(/\/:([wxpbf]):\//i)
  if (marker) return SHAREPOINT_MARKERS[marker[1].toLowerCase()]

  const byExtension = EXTENSIONS.find(([pattern]) => pattern.test(path))
  if (byExtension) return byExtension[1]

  if (/\/sites\//i.test(path) || /\/sites$/i.test(path)) return 'site'
  if (path === '' || path === '/') return 'app'
  return 'link'
}

const FAMILIES = {
  app: 'apps',
  word: 'docs',
  excel: 'docs',
  powerpoint: 'docs',
  pdf: 'docs',
  folder: 'docs',
  site: 'sites',
  link: 'sites',
}

/** Famille utilisée par le filtre du portail : apps | docs | sites. */
export function typeFamilyOf(type) {
  return FAMILIES[type] || 'sites'
}

export const TYPE_FAMILIES = ['apps', 'docs', 'sites']

// Repère visuel de chaque type : un sigle (documents Office et PDF) ou une icône.
// Les applis gardent l'icône et la couleur choisies par l'admin.
export const TYPE_META = {
  app: { icon: null },
  word: { glyph: 'W', color: '#2b579a' },
  excel: { glyph: 'X', color: '#217346' },
  powerpoint: { glyph: 'P', color: '#c43e1c' },
  pdf: { glyph: 'PDF', color: '#b42318' },
  folder: { icon: 'lucide:folder', color: '#a16207' },
  site: { icon: 'lucide:globe', color: '#0369a1' },
  link: { icon: 'lucide:link', color: '#475569' },
}

/** « Nouveau » pendant NEW_RESOURCE_DAYS jours après la création. */
export function isNewResource(resource, now = new Date()) {
  if (!resource?.created_at) return false
  const age = now - new Date(resource.created_at)
  return age >= 0 && age < NEW_RESOURCE_DAYS * DAY_MS
}

/** Note de l'admin à afficher sur la tuile, vide une fois sa date de fin passée. */
export function activeAdminNote(resource, now = new Date()) {
  const note = (resource?.admin_note || '').trim()
  if (!note) return ''
  if (resource.admin_note_until && new Date(resource.admin_note_until) < now) return ''
  return note
}

/** Date de fin de la note (ISO) vers la valeur d'un champ date (AAAA-MM-JJ, heure locale). */
export function noteUntilToInput(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

/** Champ date (AAAA-MM-JJ) vers ISO : la note reste affichée jusqu'à la fin de ce jour. */
export function noteUntilFromInput(value) {
  if (!value) return null
  const d = new Date(`${value}T23:59:59`)
  return Number.isNaN(d.getTime()) ? null : d.toISOString()
}
