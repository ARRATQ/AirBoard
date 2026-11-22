/**
 * Qui peut ouvrir quelle page d'administration.
 *
 * Se combinent :
 * - admin global (`role === 'admin'`) : tout ;
 * - rôles attribués par l'admin (`user.permissions`, {permission: 'all'|'groups'}) :
 *   chaque permission ouvre son module, sur tout le site ou sur les groupes de la personne ;
 * - éditeur (`role === 'editor'`) : actualités et événements de toute l'organisation ;
 * - admin de groupe (administre au moins un groupe, quel que soit son rôle) :
 *   actualités, événements, sondages et ressources de ses groupes.
 *   Ces deux derniers disparaîtront quand chaque module sera passé aux permissions.
 *
 * Une page d'administration déclare un niveau d'accès dans `meta.access`
 * (routeur) ou `access` (menu) ; `canAccess` dit si l'utilisateur l'atteint.
 */

/** Admin global uniquement */
export const ADMIN = 'admin'
/** Admin global ou admin de groupe */
export const GROUP_ADMIN = 'groupAdmin'
/** Admin global, éditeur ou admin de groupe */
export const CONTENT = 'content'
/** Ressources et catégories : permission resources.manage (ou admin de groupe) */
export const RESOURCES = 'resources'

/** Permission du serveur qui ouvre chaque niveau d'accès géré par permission */
export const ACCESS_PERMISSION = {
  [RESOURCES]: 'resources.manage'
}

/** Portées d'une permission */
export const SCOPE_ALL = 'all'
export const SCOPE_GROUPS = 'groups'

/**
 * @param {object|null|undefined} user utilisateur du store d'authentification
 * @returns {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean, permissions: Record<string, string> }}
 */
export function getRoles(user) {
  const adminOfGroups = Array.isArray(user?.admin_of_groups) ? user.admin_of_groups : []
  const managedGroupIds = Array.isArray(user?.managed_group_ids) ? user.managed_group_ids : []
  return {
    isAdmin: user?.role === 'admin',
    isEditor: user?.role === 'editor',
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
 * @param {string|undefined} access niveau demandé (ADMIN, GROUP_ADMIN, CONTENT) ; vide = libre
 * @param {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean }} roles
 * @returns {boolean}
 */
export function canAccess(access, roles) {
  if (!access) return true
  if (roles.isAdmin) return true
  switch (access) {
    case GROUP_ADMIN:
      return roles.isGroupAdmin
    case CONTENT:
      return roles.isEditor || roles.isGroupAdmin
    case RESOURCES:
      return roles.isGroupAdmin || permissionScope(roles, ACCESS_PERMISSION[RESOURCES]) !== ''
    default:
      return false
  }
}

/**
 * Périmètre d'une page partagée entre admin global et admin de groupe.
 * `global` : l'utilisateur voit et gère tout le module (API admin ou éditeur).
 * `group` : seulement ce qui concerne les groupes qu'il administre (API group-admin).
 *
 * @param {'content'|'groupAdmin'|'resources'} access niveau de la page
 * @param {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean, permissions?: Record<string, string> }} roles
 * @returns {'global'|'group'}
 */
export function getScope(access, roles) {
  if (roles.isAdmin) return 'global'
  if (ACCESS_PERMISSION[access] && permissionScope(roles, ACCESS_PERMISSION[access]) === SCOPE_ALL) return 'global'
  if (access === CONTENT && roles.isEditor) return 'global'
  return 'group'
}
