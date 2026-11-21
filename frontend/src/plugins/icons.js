// Icônes Iconify servies par l'application elle-même (voir vite-plugins/iconsPlugin.js) : celles écrites
// dans le code arrivent avec leur module ; les autres (saisies en base) via des paquets à la demande.
import { _api, setCustomIconsLoader } from '@iconify/vue'
import { ICON_BUCKETS, ICON_PREFIXES, iconBucket } from '@/utils/iconBucket'

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

export function setupIcons() {
  // Plus aucune requête vers api.iconify.design : une collection non embarquée reste sans icône.
  _api.setFetch(() => Promise.resolve(new Response(null, { status: 404 })))

  ICON_PREFIXES.forEach((prefix) => {
    setCustomIconsLoader((names) => loadFromShards(names, prefix), prefix)
  })
}
