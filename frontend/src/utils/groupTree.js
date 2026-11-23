// Hiérarchie des groupes d'utilisateurs (champ parent_id renvoyé par /admin/groups).

const byName = (a, b) => (a.name || '').localeCompare(b.name || '')

/**
 * Ordonne les groupes en arbre : chaque parent suivi de ses sous-groupes, triés par nom.
 * Un groupe dont le parent est absent de la liste est traité comme une racine.
 * @returns {Array<{ group: object, depth: number, path: string[] }>} path : noms des groupes au-dessus
 */
export function orderGroupTree(groups) {
  const list = Array.isArray(groups) ? groups : []
  const ids = new Set(list.map(g => g.id))
  const children = new Map()
  const roots = []
  for (const g of list) {
    if (g.parent_id != null && ids.has(g.parent_id) && g.parent_id !== g.id) {
      if (!children.has(g.parent_id)) children.set(g.parent_id, [])
      children.get(g.parent_id).push(g)
    } else {
      roots.push(g)
    }
  }
  const out = []
  const seen = new Set()
  const walk = (g, depth, path) => {
    if (seen.has(g.id)) return
    seen.add(g.id)
    out.push({ group: g, depth, path })
    for (const child of [...(children.get(g.id) || [])].sort(byName)) {
      walk(child, depth + 1, [...path, g.name])
    }
  }
  ;[...roots].sort(byName).forEach(g => walk(g, 0, []))
  // Groupes pris dans une boucle (ne devrait pas arriver) : affichés à la racine
  list.filter(g => !seen.has(g.id)).sort(byName).forEach(g => walk(g, 0, []))
  return out
}

/** Identifiants du groupe et de tous ses sous-groupes. */
export function descendantIds(groups, id) {
  const list = Array.isArray(groups) ? groups : []
  const out = new Set([id])
  let grew = true
  while (grew) {
    grew = false
    for (const g of list) {
      if (g.parent_id != null && out.has(g.parent_id) && !out.has(g.id)) {
        out.add(g.id)
        grew = true
      }
    }
  }
  return out
}

/**
 * Groupes pouvant servir de parent à `groupId` (null pour un nouveau groupe) : ni lui-même, ni
 * ses sous-groupes, ni le groupe par défaut. Ordonnés en arbre.
 */
export function parentCandidates(groups, groupId) {
  const excluded = groupId == null ? new Set() : descendantIds(groups, groupId)
  return orderGroupTree(groups).filter(({ group }) => !excluded.has(group.id) && !group.is_default)
}

/** Libellé indenté pour une liste déroulante (les <option> n'acceptent pas de mise en forme). */
export function indentedLabel({ group, depth }) {
  return `${'   '.repeat(depth)}${depth > 0 ? '└ ' : ''}${group.name}`
}
