/**
 * Qui peut ouvrir quelle page d'administration.
 *
 * Se combinent :
 * - admin (`role === 'admin'`) : tout ;
 * - rôles attribués par l'admin (`user.permissions`, {permission: 'all'|'groups'}) :
 *   chaque permission ouvre son module, sur tout le site ou sur les groupes de la personne.
 *
 * Une page d'administration déclare un niveau d'accès dans `meta.access`
 * (routeur) ou `access` (menu) ; `canAccess` dit si l'utilisateur l'atteint.
 */

/** Admin global uniquement */
export const ADMIN = 'admin'
/** Administration en général (vue d'ensemble) : admin, ou au moins une permission */
export const MANAGEMENT = 'management'
/** Ressources et catégories : permission resources.manage */
export const RESOURCES = 'resources'
/** Actualités : permission news.manage */
export const NEWS = 'news'
/** Événements : permission events.manage */
export const EVENTS = 'events'
/** Sondages : permission polls.manage */
export const POLLS = 'polls'
/** Modération des commentaires : permission comments.moderate (sur tout le site) */
export const MODERATION = 'moderation'

/** Permission du serveur qui ouvre chaque niveau d'accès géré par permission */
export const ACCESS_PERMISSION = {
  [RESOURCES]: 'resources.manage',
  [NEWS]: 'news.manage',
  [EVENTS]: 'events.manage',
  [POLLS]: 'polls.manage',
  [MODERATION]: 'comments.moderate'
}

/** Portées d'une permission */
export const SCOPE_ALL = 'all'
export const SCOPE_GROUPS = 'groups'

/**
 * @param {object|null|undefined} user utilisateur du store d'authentification
 * @returns {{ isAdmin: boolean, permissions: Record<string, string> }}
 */
export function getRoles(user) {
  return {
    isAdmin: user?.role === 'admin',
    permissions: user?.permissions && typeof user.permissions === 'object' ? user.permissions : {}
  }
}

/**
 * Portée d'une permission pour l'utilisateur : 'all', 'groups' ou '' s'il ne l'a pas.
 * @param {{ isAdmin: boolean, permissions?: Record<string, string> }} roles
 * @param {string} permission
 */
export function permissionScope(roles, permission) {
  if (roles.isAdmin) return SCOPE_ALL
  return roles.permissions?.[permission] || ''
}

/** L'utilisateur a-t-il au moins une permission attribuée par un rôle ? */
export function hasAnyPermission(roles) {
  return Object.keys(roles.permissions || {}).length > 0
}

/**
 * @param {string|undefined} access niveau demandé (ADMIN, MANAGEMENT ou un niveau géré par permission) ; vide = libre
 * @param {{ isAdmin: boolean, permissions?: Record<string, string> }} roles
 * @returns {boolean}
 */
export function canAccess(access, roles) {
  if (!access) return true
  if (roles.isAdmin) return true
  if (access === MANAGEMENT) return hasAnyPermission(roles)
  // Niveau géré par permission : qui l'a, quelle qu'en soit la portée
  const permission = ACCESS_PERMISSION[access]
  if (!permission) return false
  return permissionScope(roles, permission) !== ''
}

/**
 * Périmètre d'une page d'administration d'un module.
 * `global` : l'utilisateur voit et gère tout le module (admin, permission « tout le site »).
 * `group` : seulement ce qui concerne ses groupes (API group-admin).
 *
 * @param {string} access niveau de la page
 * @param {{ isAdmin: boolean, permissions?: Record<string, string> }} roles
 * @returns {'global'|'group'}
 */
export function getScope(access, roles) {
  if (roles.isAdmin) return 'global'
  if (ACCESS_PERMISSION[access] && permissionScope(roles, ACCESS_PERMISSION[access]) === SCOPE_ALL) return 'global'
  return 'group'
}

/**
 * Peut-on modifier ou supprimer ce contenu (actualité, événement, sondage) ? Même règle que
 * le serveur : permission sur tout le site (admin compris), auteur, ou contenu ciblant un
 * groupe où l'on gère ce module (portée « ses groupes »). Un contenu sans groupe cible est
 * global.
 * Les groupes de la portée « ses groupes » viennent du serveur (`permission_group_ids`) :
 * ceux dont la personne est membre, hors groupe par défaut (Common), qui réunit tout le monde.
 *
 * @param {object|null|undefined} user utilisateur du store d'authentification
 * @param {string} permission permission du module (ACCESS_PERMISSION[...])
 * @param {{ author_id?: number, target_groups?: Array<{ id: number }> }|null|undefined} item
 * @returns {boolean}
 */
export function canManageItem(user, permission, item) {
  if (!user) return false
  const roles = getRoles(user)
  const scope = permissionScope(roles, permission)
  if (scope === SCOPE_ALL) return true
  if (!item) return false
  if (item.author_id && item.author_id === user.id) return true
  const targets = Array.isArray(item.target_groups) ? item.target_groups.map((g) => g.id) : []
  if (targets.length === 0) return false
  if (scope !== SCOPE_GROUPS || !Array.isArray(user.permission_group_ids)) return false
  const ids = new Set(user.permission_group_ids)
  return targets.some((id) => ids.has(id))
}

/** Modère-t-on les commentaires (permission comments.moderate, admin compris) ? */
export function canModerateComments(user) {
  return permissionScope(getRoles(user), ACCESS_PERMISSION[MODERATION]) === SCOPE_ALL
}
