// Icônes Iconify : mdi, lucide et heroicons sont servies par l'application elle-même (voir
// vite-plugins/iconsPlugin.js) ; les autres collections, choisies dans le sélecteur d'icônes,
// viennent de l'API Iconify, et seulement d'elle (pas des hôtes de secours de la bibliothèque).
import { _api, addAPIProvider, setCustomIconsLoader } from '@iconify/vue'
import { ICONIFY_API } from '@/utils/iconifySearch'
import { ICON_BUCKETS, ICON_PREFIXES, iconBucket } from '@/utils/iconBucket'
import { placeholderIconSet } from '@/utils/iconPlaceholder'

// La table des paquets (~640 entrées) n'est chargée qu'à la première icône venue de la base.
let shardLoadersPromise = null
const getShardLoaders = () => (shardLoadersPromise ||= import('virtual:icons').then((m) => m.shardLoaders))

/**
 * Charge les paquets qui contiennent ces icônes et les fusionne en une collection.
 * @param {string[]} names icônes demandées
 * @param {string} prefix collection
 */
async function loadFromShards(names, prefix) {
  const loaders = (await getShardLoaders())[prefix]
  const count = ICON_BUCKETS[prefix]
  const buckets = [...new Set(names.map((name) => iconBucket(name, count)))].filter((bucket) => loaders[bucket])
  const shards = await Promise.all(buckets.map((bucket) => loaders[bucket]().then((m) => m.default)))
  return shards.reduce(
    (merged, shard) => ({
      ...shard,
      icons: { ...merged.icons, ...shard.icons },
      aliases: { ...merged.aliases, ...shard.aliases },
    }),
    { prefix, icons: {}, aliases: {} },
  )
}

/**
 * Réponse locale : le pictogramme neutre pour chaque icône demandée.
 * @param {string} url adresse que la bibliothèque voulait interroger
 * @returns {Response}
 */
function placeholderResponse(url) {
  const set = placeholderIconSet(url)
  return set
    ? new Response(JSON.stringify(set), { status: 200, headers: { 'Content-Type': 'application/json' } })
    : new Response(null, { status: 404 })
}

export function setupIcons() {
  addAPIProvider('', { resources: [ICONIFY_API] })

  // Iconify hors d'atteinte (intranet coupé d'internet) : pictogramme neutre plutôt qu'un vide.
  _api.setFetch(async (url, init) => {
    try {
      const response = await fetch(url, init)
      if (response.ok) return response
    } catch {
      // réseau indisponible : repli ci-dessous
    }
    return placeholderResponse(String(url))
  })

  ICON_PREFIXES.forEach((prefix) => {
    setCustomIconsLoader((names) => loadFromShards(names, prefix), prefix)
  })
}
