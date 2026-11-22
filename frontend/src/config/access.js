/**
 * Qui peut ouvrir quelle page d'administration.
 *
 * Se combinent :
 * - admin global (`role === 'admin'`) : tout ;
 * - rôles attribués par l'admin (`user.permissions`, {permission: 'all'|'groups'}) :
 *   chaque permission ouvre son module, sur tout le site ou sur les groupes de la personne ;
 * - ancien admin de groupe (administre au moins un groupe) : actualités, événements,
 *   sondages et ressources de ses groupes, en attendant la fin des admins de groupe.
 *
 * Une page d'administration déclare un niveau d'accès dans `meta.access`
 * (routeur) ou `access` (menu) ; `canAccess` dit si l'utilisateur l'atteint.
 */

/** Admin global uniquement */
export const ADMIN = 'admin'
/** Admin global ou admin de groupe */
export const GROUP_ADMIN = 'groupAdmin'
/** Ressources et catégories : permission resources.manage (ou admin de groupe) */
export const RESOURCES = 'resources'
/** Actualités : permission news.manage (ou admin de groupe) */
export const NEWS = 'news'
/** Événements : permission events.manage (ou admin de groupe) */
export const EVENTS = 'events'
/** Sondages : permission polls.manage (ou admin de groupe) */
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

/** Niveaux que l'ancien admin de groupe atteint sans permission (contenus de ses groupes) */
const GROUP_ADMIN_LEVELS = [RESOURCES, NEWS, EVENTS, POLLS]

/** Portées d'une permission */
export const SCOPE_ALL = 'all'
export const SCOPE_GROUPS = 'groups'

/**
 * @param {object|null|undefined} user utilisateur du store d'authentification
 * @returns {{ isAdmin: boolean, isGroupAdmin: boolean, permissions: Record<string, string> }}
 */
export function getRoles(user) {
  const adminOfGroups = Array.isArray(user?.admin_of_groups) ? user.admin_of_groups : []
  const managedGroupIds = Array.isArray(user?.managed_group_ids) ? user.managed_group_ids : []
  return {
    isAdmin: user?.role === 'admin',
    isGroupAdmin: adminOfGroups.length > 0 || managedGroupIds.length > 0,
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
 * @param {string|undefined} access niveau demandé (ADMIN, GROUP_ADMIN ou un niveau géré par permission) ; vide = libre
 * @param {{ isAdmin: boolean, isGroupAdmin: boolean, permissions?: Record<string, string> }} roles
 * @returns {boolean}
 */
export function canAccess(access, roles) {
  if (!access) return true
  if (roles.isAdmin) return true
  if (access === GROUP_ADMIN) return roles.isGroupAdmin
  // Niveau géré par permission : qui l'a (toute portée), ou un ancien admin de groupe
  const permission = ACCESS_PERMISSION[access]
  if (!permission) return false
  if (roles.isGroupAdmin && GROUP_ADMIN_LEVELS.includes(access)) return true
  return permissionScope(roles, permission) !== ''
}

/**
 * Périmètre d'une page partagée entre admin global et admin de groupe.
 * `global` : l'utilisateur voit et gère tout le module (API admin ou éditeur).
 * `group` : seulement ce qui concerne les groupes qu'il administre (API group-admin).
 *
 * @param {'groupAdmin'|'resources'|'news'|'events'|'polls'} access niveau de la page
 * @param {{ isAdmin: boolean, isGroupAdmin: boolean, permissions?: Record<string, string> }} roles
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
 * groupe où l'on gère ce module (groupes membres si la portée est « ses groupes », groupes
 * administrés pour un ancien admin de groupe). Un contenu sans groupe cible est global.
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
  const ids = new Set([
    ...(Array.isArray(user.managed_group_ids) ? user.managed_group_ids : []),
    ...(Array.isArray(user.admin_of_groups) ? user.admin_of_groups.map((g) => g.id) : []),
    ...(scope === SCOPE_GROUPS && Array.isArray(user.permission_group_ids) ? user.permission_group_ids : [])
  ])
  return targets.some((id) => ids.has(id))
}

/** Modère-t-on les commentaires (permission comments.moderate, admin compris) ? */
export function canModerateComments(user) {
  return permissionScope(getRoles(user), ACCESS_PERMISSION[MODERATION]) === SCOPE_ALL
}
