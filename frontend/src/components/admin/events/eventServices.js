import {
  adminEventsService,
  editorEventsService,
  eventsService,
  groupAdminEventsService
} from '@/services/api'

/**
 * Service d'écriture des événements selon le rôle (droits vérifiés dans backend/main.go) :
 * - admin global : /admin/events ;
 * - permission sur tout le site : /editor/events (create, update, delete) ;
 * - permission limitée à ses groupes : /group-admin/events.
 *
 * @param {boolean} isAdmin
 * @param {boolean} isGlobal
 * @returns {{ createEvent: Function, updateEvent: Function, deleteEvent: Function }}
 */
export function getEventWriteService(isAdmin, isGlobal) {
  if (isAdmin) return adminEventsService
  if (isGlobal) return editorEventsService
  return groupAdminEventsService
}

/**
 * Liste paginée des événements gérables selon le rôle.
 * L'éditeur n'a pas accès à GET /admin/events : il passe par la liste commune /events,
 * qui lui renvoie ses brouillons et les événements publiés.
 *
 * @param {boolean} isAdmin
 * @param {boolean} isGlobal
 * @returns {(params: object) => Promise<object>}
 */
export function getEventListFetcher(isAdmin, isGlobal) {
  if (isAdmin) return (params) => adminEventsService.getEvents(params)
  if (isGlobal) return (params) => eventsService.getEvents(params)
  return (params) => groupAdminEventsService.getEvents(params)
}
