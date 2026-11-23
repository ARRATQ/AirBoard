import { NEWS, EVENTS, POLLS, RESOURCES, MODERATION, USER_GROUPS, ADMIN, ACCESS_PERMISSION, canAccess, permissionScope, SCOPE_ALL, SCOPE_GROUPS } from '@/config/access'

/**
 * Aide intégrée : un guide par sujet, rangé par public.
 *
 * - `everyone` : ce que tout le monde utilise (aucun niveau d'accès) ;
 * - `contributor` : un guide par permission de rôle, visible par qui l'a (et l'admin) ;
 * - `admin` : l'administration réservée à l'admin.
 *
 * `routes` : pages pour lesquelles le guide s'ouvre directement (préfixe de chemin).
 * Le texte de chaque guide est dans `content/<langue>/<id>.md`.
 */
export const HELP_AUDIENCES = ['everyone', 'contributor', 'admin']

export const HELP_TOPICS = [
  { id: 'start', audience: 'everyone', icon: 'lucide:compass', routes: ['/home', '/search'] },
  { id: 'resources', audience: 'everyone', icon: 'lucide:app-window', routes: ['/resources'] },
  { id: 'news', audience: 'everyone', icon: 'lucide:newspaper', routes: ['/news'] },
  { id: 'agenda', audience: 'everyone', icon: 'lucide:calendar', routes: ['/events'] },
  { id: 'polls', audience: 'everyone', icon: 'lucide:chart-column', routes: ['/polls'] },
  { id: 'suggestions', audience: 'everyone', icon: 'lucide:lightbulb', routes: ['/suggestions'] },
  { id: 'directory', audience: 'everyone', icon: 'lucide:users', routes: ['/directory'] },
  { id: 'messaging', audience: 'everyone', icon: 'lucide:message-circle', routes: [] },
  { id: 'xp', audience: 'everyone', icon: 'lucide:medal', routes: ['/gamification'] },
  { id: 'profile', audience: 'everyone', icon: 'lucide:circle-user', routes: ['/profile', '/notifications'] },

  { id: 'manage-news', audience: 'contributor', access: NEWS, icon: 'lucide:newspaper', routes: ['/admin/news'] },
  { id: 'manage-events', audience: 'contributor', access: EVENTS, icon: 'lucide:calendar-cog', routes: ['/admin/events'] },
  { id: 'manage-polls', audience: 'contributor', access: POLLS, icon: 'lucide:chart-column', routes: ['/admin/polls'] },
  { id: 'manage-resources', audience: 'contributor', access: RESOURCES, icon: 'lucide:layout-grid', routes: ['/admin/resources'] },
  { id: 'moderation', audience: 'contributor', access: MODERATION, icon: 'lucide:message-square-text', routes: ['/admin/comments'] },
  // L'admin gère les groupes depuis « Groupes » (guide admin-groups)
  { id: 'sub-groups', audience: 'contributor', access: USER_GROUPS, hideForAdmin: true, icon: 'lucide:users-round', routes: ['/admin/my-groups'] },

  { id: 'admin-users', audience: 'admin', access: ADMIN, icon: 'lucide:users', routes: ['/admin/users'] },
  { id: 'admin-groups', audience: 'admin', access: ADMIN, icon: 'lucide:users-round', routes: ['/admin/groups'] },
  { id: 'admin-roles', audience: 'admin', access: ADMIN, icon: 'lucide:shield-check', routes: ['/admin/roles'] },
  { id: 'admin-directory', audience: 'admin', access: ADMIN, icon: 'lucide:network', routes: ['/admin/directory'] },
  { id: 'admin-content', audience: 'admin', access: ADMIN, icon: 'lucide:megaphone', routes: ['/admin/announcements', '/admin/media', '/admin/suggestions', '/admin/gamification'] },
  { id: 'admin-system', audience: 'admin', access: ADMIN, icon: 'lucide:settings', routes: ['/admin', '/admin/settings', '/admin/email', '/admin/oauth', '/admin/analytics', '/admin/reports', '/admin/chatbots'] }
]

/**
 * Guides que l'utilisateur peut lire, dans l'ordre du catalogue.
 * @param {{ isAdmin: boolean, permissions?: Record<string, string> }} roles
 * @param {Array} [topics]
 */
export function visibleTopics(roles, topics = HELP_TOPICS) {
  return topics.filter((topic) => canAccess(topic.access, roles) && !(roles.isAdmin && topic.hideForAdmin))
}

/**
 * Portée du guide pour l'utilisateur : 'all', 'groups', ou '' (guide sans permission).
 * @param {{ access?: string }} topic
 * @param {{ isAdmin: boolean, permissions?: Record<string, string> }} roles
 */
export function topicScope(topic, roles) {
  const permission = ACCESS_PERMISSION[topic?.access]
  if (!permission) return ''
  return permissionScope(roles, permission)
}

/**
 * Guide de la page affichée : celui dont un chemin est le plus long préfixe de `path`.
 * @param {string} path chemin de la route (sans requête)
 * @param {Array} topics guides visibles
 * @returns {object|null}
 */
export function topicForRoute(path, topics) {
  let best = null
  let bestLength = -1
  for (const topic of topics) {
    for (const route of topic.routes || []) {
      const matches = path === route || path.startsWith(`${route}/`)
      if (matches && route.length > bestLength) {
        best = topic
        bestLength = route.length
      }
    }
  }
  return best
}

const SCOPE_BLOCK = /^:::(all|groups)[ \t]*\r?\n([\s\S]*?)^:::[ \t]*$\r?\n?/gm

/**
 * Garde, dans un guide, les blocs écrits pour la portée de l'utilisateur.
 * `:::all` … `:::` : permission sur tout le site ; `:::groups` … `:::` : sur ses groupes.
 * Sans portée (guide pour tous), les deux blocs sont retirés.
 * @param {string} markdown
 * @param {string} scope 'all', 'groups' ou ''
 */
export function applyScope(markdown, scope) {
  return markdown.replace(SCOPE_BLOCK, (_, blockScope, body) =>
    (blockScope === SCOPE_ALL && scope === SCOPE_ALL) || (blockScope === SCOPE_GROUPS && scope === SCOPE_GROUPS) ? body : ''
  )
}
