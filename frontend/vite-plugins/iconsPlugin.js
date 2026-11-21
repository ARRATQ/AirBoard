// Plugin Vite : icônes Iconify embarquées, sans appel à api.iconify.design.
//
// - Chaque module de src/ qui écrit une icône en toutes lettres (« mdi:home ») reçoit ces icônes
//   via addCollection : elles suivent le découpage en chunks (une page admin n'alourdit pas l'accueil).
// - `virtual:icons` exporte `shardLoaders` : pour chaque collection, des import() de paquets
//   d'icônes. Ils servent aux icônes saisies par les admins et stockées en base (applications,
//   groupes, catégories…). Un paquet : `virtual:icons-shard/<préfixe>-<n>` (voir src/utils/iconBucket.js).
import { readFileSync } from 'fs'
import { createRequire } from 'module'
import { getIcons } from '@iconify/utils'
import { ICON_BUCKETS, ICON_PREFIXES, extractIconNames, iconBucket } from '../src/utils/iconBucket.js'

const VIRTUAL_MAIN = 'virtual:icons'
const VIRTUAL_SHARD = 'virtual:icons-shard/'
const RESOLVED_PREFIX = '\0'
const SOURCE_FILE = /\/src\/(?!.*__tests__).*\.(vue|js)$/

const require = createRequire(import.meta.url)

function loadIconSet(prefix) {
  return JSON.parse(readFileSync(require.resolve(`@iconify-json/${prefix}/icons.json`), 'utf8'))
}

/** Noms (icônes et alias) de la collection, regroupés par paquet. */
function namesByBucket(iconSet) {
  const count = ICON_BUCKETS[iconSet.prefix]
  const buckets = new Map()
  const names = [...Object.keys(iconSet.icons), ...Object.keys(iconSet.aliases || {})]
  for (const name of names) {
    const bucket = iconBucket(name, count)
    if (!buckets.has(bucket)) buckets.set(bucket, [])
    buckets.get(bucket).push(name)
  }
  return buckets
}

/** Collections réduites aux icônes demandées, plus les noms introuvables. */
function subsetsFor(getIconSet, pairs) {
  const byPrefix = new Map()
  for (const [prefix, name] of pairs) {
    if (!byPrefix.has(prefix)) byPrefix.set(prefix, new Set())
    byPrefix.get(prefix).add(name)
  }
  const sets = []
  const missing = []
  for (const [prefix, names] of byPrefix) {
    const subset = getIcons(getIconSet(prefix), [...names], true)
    if (subset.not_found) missing.push(...subset.not_found.map((name) => `${prefix}:${name}`))
    delete subset.not_found
    if (Object.keys(subset.icons).length) sets.push(subset)
  }
  return { sets, missing }
}

export default function iconsPlugin() {
  const iconSets = new Map()
  const getIconSet = (prefix) => {
    if (!iconSets.has(prefix)) iconSets.set(prefix, loadIconSet(prefix))
    return iconSets.get(prefix)
  }
  const reported = new Set()

  function buildMainModule() {
    const loaders = ICON_PREFIXES.map((prefix) => {
      const entries = [...namesByBucket(getIconSet(prefix)).keys()]
        .sort((a, b) => a - b)
        .map((bucket) => `${bucket}: () => import('${VIRTUAL_SHARD}${prefix}-${bucket}')`)
      return `${JSON.stringify(prefix)}: { ${entries.join(', ')} }`
    })
    return `export const shardLoaders = { ${loaders.join(',\n')} }`
  }

  function buildShardModule(id) {
    const [prefix, bucketText] = id.slice(VIRTUAL_SHARD.length).split('-')
    if (!ICON_PREFIXES.includes(prefix)) throw new Error(`collection d'icônes inconnue : ${prefix}`)
    const iconSet = getIconSet(prefix)
    const names = namesByBucket(iconSet).get(Number(bucketText)) || []
    return `export default ${JSON.stringify(getIcons(iconSet, names) || { prefix, icons: {} })}`
  }

  return {
    name: 'airboard-icons',
    resolveId(id) {
      if (id === VIRTUAL_MAIN || id.startsWith(VIRTUAL_SHARD)) return RESOLVED_PREFIX + id
      return null
    },
    load(id) {
      if (!id.startsWith(RESOLVED_PREFIX)) return null
      const virtualId = id.slice(RESOLVED_PREFIX.length)
      if (virtualId === VIRTUAL_MAIN) return buildMainModule()
      if (virtualId.startsWith(VIRTUAL_SHARD)) return buildShardModule(virtualId)
      return null
    },
    // Passe après le plugin Vue : le code reçu est déjà du JS (template compilé compris).
    transform(code, id) {
      if (!SOURCE_FILE.test(id) || id.includes('/node_modules/')) return null
      const { sets, missing } = subsetsFor(getIconSet, extractIconNames(code))
      missing.filter((name) => !reported.has(name)).forEach((name) => {
        reported.add(name)
        this.warn(`icône introuvable : ${name}`)
      })
      if (!sets.length) return null
      const calls = sets.map((set) => `__airboardAddIcons(${JSON.stringify(set)})`).join('\n')
      return {
        code: `${code}\nimport { addCollection as __airboardAddIcons } from '@iconify/vue'\n${calls}\n`,
        map: null,
      }
    },
  }
}
