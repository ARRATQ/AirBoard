// Répartition des icônes en paquets, partagée par le plugin Vite (build) et le chargeur (navigateur).
// Une icône absente du sous-ensemble embarqué est cherchée dans son paquet, chargé à la demande.

/**
 * Collections embarquées (celles que l'admin peut saisir, voir IconInput) et nombre de paquets
 * de chacune : quelques dizaines de noms par paquet, 2 à 4 kB compressés en moyenne.
 */
export const ICON_BUCKETS = { mdi: 512, lucide: 64, heroicons: 64 }

export const ICON_PREFIXES = Object.keys(ICON_BUCKETS)

/**
 * Hash djb2 du nom, ramené au nombre de paquets.
 * @param {string} name nom de l'icône sans préfixe (ex. "home")
 * @param {number} buckets nombre de paquets de la collection
 * @returns {number} index du paquet, entre 0 et buckets - 1
 */
export function iconBucket(name, buckets) {
  let hash = 5381
  for (let i = 0; i < name.length; i++) {
    hash = ((hash << 5) + hash + name.charCodeAt(i)) | 0
  }
  return Math.abs(hash) % buckets
}

const ICON_NAME_RE = new RegExp(`\\b(${ICON_PREFIXES.join('|')}):([a-z0-9]+(?:-[a-z0-9]+)*)`, 'g')

/**
 * Relève les icônes écrites en toutes lettres dans un fichier source.
 * @param {string} source contenu du fichier
 * @returns {Array<[string, string]>} paires [préfixe, nom]
 */
export function extractIconNames(source) {
  const found = []
  for (const match of source.matchAll(ICON_NAME_RE)) {
    found.push([match[1], match[2]])
  }
  return found
}
