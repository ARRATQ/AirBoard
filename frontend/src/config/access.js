/**
 * Qui peut ouvrir quelle page d'administration.
 *
 * Trois rôles se combinent :
 * - admin global (`role === 'admin'`) : tout ;
 * - éditeur (`role === 'editor'`) : actualités et événements de toute l'organisation ;
 * - admin de groupe (administre au moins un groupe, quel que soit son rôle) :
 *   actualités, événements, sondages et ressources de ses groupes.
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

/**
 * @param {object|null|undefined} user utilisateur du store d'authentification
 * @returns {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean }}
 */
export function getRoles(user) {
  const adminOfGroups = Array.isArray(user?.admin_of_groups) ? user.admin_of_groups : []
  const managedGroupIds = Array.isArray(user?.managed_group_ids) ? user.managed_group_ids : []
  return {
    isAdmin: user?.role === 'admin',
    isEditor: user?.role === 'editor',
    isGroupAdmin: adminOfGroups.length > 0 || managedGroupIds.length > 0
  }
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
    default:
      return false
  }
}

/**
 * Périmètre d'une page partagée entre admin global et admin de groupe.
 * `global` : l'utilisateur voit et gère tout le module (API admin ou éditeur).
 * `group` : seulement ce qui concerne les groupes qu'il administre (API group-admin).
 *
 * @param {'content'|'groupAdmin'} access niveau de la page
 * @param {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean }} roles
 * @returns {'global'|'group'}
 */
export function getScope(access, roles) {
  if (roles.isAdmin) return 'global'
  if (access === CONTENT && roles.isEditor) return 'global'
  return 'group'
}
