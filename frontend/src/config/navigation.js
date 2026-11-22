import { ADMIN, GROUP_ADMIN, RESOURCES, NEWS, EVENTS, canAccess, hasAnyPermission } from './access'

/**
 * Menu latéral, décrit une seule fois pour tous les rôles.
 *
 * Chaque lien porte un niveau d'accès (`access`, voir config/access.js) ; le menu
 * d'un utilisateur ne garde que ce qu'il peut ouvrir. Les libellés sont des clés i18n.
 */

/** Espace utilisateur : visible par tous, jamais filtré */
export const USER_NAV = [
  { to: '/home', icon: 'lucide:house', labelKey: 'common.home' },
  { to: '/resources', icon: 'lucide:app-window', labelKey: 'common.applications' },
  { to: '/news', icon: 'lucide:newspaper', labelKey: 'common.newsHub' },
  { to: '/events', icon: 'lucide:calendar', labelKey: 'common.events' },
  { to: '/polls', icon: 'lucide:chart-column', labelKey: 'common.polls' },
  { to: '/suggestions', icon: 'lucide:lightbulb', labelKey: 'common.suggestions' },
  { to: '/gamification', icon: 'lucide:medal', labelKey: 'common.gamification' },
  { to: '/directory', icon: 'lucide:users', labelKey: 'directory.title' },
  { to: '/search', icon: 'lucide:search', labelKey: 'search.title' }
]

/** Page d'accueil de l'administration */
export const ADMIN_OVERVIEW = {
  to: '/admin',
  icon: 'lucide:layout-dashboard',
  labelKey: 'nav.admin.overview',
  access: GROUP_ADMIN,
  exact: true
}

/** Administration, rangée par domaine */
export const ADMIN_SECTIONS = [
  {
    id: 'content',
    labelKey: 'nav.admin.content',
    icon: 'lucide:file-text',
    items: [
      { to: '/admin/news', icon: 'lucide:newspaper', labelKey: 'nav.admin.news', access: NEWS },
      { to: '/admin/events', icon: 'lucide:calendar-cog', labelKey: 'nav.admin.events', access: EVENTS },
      { to: '/admin/polls', icon: 'lucide:chart-column', labelKey: 'nav.admin.polls', access: GROUP_ADMIN },
      { to: '/admin/announcements', icon: 'lucide:megaphone', labelKey: 'nav.admin.announcements', access: ADMIN },
      { to: '/admin/media', icon: 'lucide:images', labelKey: 'nav.admin.media', access: ADMIN }
    ]
  },
  {
    id: 'community',
    labelKey: 'nav.admin.community',
    icon: 'lucide:user-round',
    items: [
      { to: '/admin/suggestions', icon: 'lucide:lightbulb', labelKey: 'nav.admin.suggestions', access: ADMIN },
      { to: '/admin/comments', icon: 'lucide:message-square-text', labelKey: 'nav.admin.moderation', access: ADMIN },
      { to: '/admin/gamification', icon: 'lucide:medal', labelKey: 'nav.admin.gamification', access: ADMIN },
      { to: '/admin/directory', icon: 'lucide:network', labelKey: 'nav.admin.directory', access: ADMIN }
    ]
  },
  {
    id: 'resources',
    labelKey: 'nav.admin.resources',
    icon: 'lucide:layout-grid',
    items: [
      { to: '/admin/resources', icon: 'lucide:layout-grid', labelKey: 'nav.admin.resources', access: RESOURCES }
    ]
  },
  {
    id: 'users',
    labelKey: 'nav.admin.users',
    icon: 'lucide:users',
    items: [
      { to: '/admin/users', icon: 'lucide:users', labelKey: 'nav.admin.users', access: ADMIN },
      { to: '/admin/groups', icon: 'lucide:users', labelKey: 'nav.admin.groups', access: ADMIN },
      { to: '/admin/roles', icon: 'lucide:shield-check', labelKey: 'nav.admin.roles', access: ADMIN },
      { to: '/admin/oauth', icon: 'lucide:shield-ellipsis', labelKey: 'nav.admin.oauth', access: ADMIN }
    ]
  },
  {
    id: 'system',
    labelKey: 'nav.admin.system',
    icon: 'lucide:settings',
    items: [
      { to: '/admin/settings', icon: 'lucide:settings', labelKey: 'nav.admin.settings', access: ADMIN },
      { to: '/admin/email', icon: 'lucide:mail', labelKey: 'nav.admin.email', access: ADMIN },
      { to: '/admin/analytics', icon: 'lucide:chart-line', labelKey: 'nav.admin.analytics', access: ADMIN },
      { to: '/admin/reports', icon: 'lucide:chart-spline', labelKey: 'nav.admin.reports', access: ADMIN },
      { to: '/admin/chatbots', icon: 'lucide:bot', labelKey: 'nav.admin.chatbots', access: ADMIN }
    ]
  }
]

/**
 * Menu d'administration d'un utilisateur.
 *
 * - Admin global : vue d'ensemble puis domaines repliables (`grouped: true`).
 * - Admin de groupe ou titulaire de permissions : une liste courte des seules pages permises,
 *   sous un seul titre (`grouped: false`).
 * - Sinon : `null`, pas de bloc d'administration.
 *
 * @param {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean }} roles
 * @returns {null | { titleKey: string, grouped: boolean, overview: object|null, sections: Array, items: Array }}
 */
export function buildAdminNav(roles) {
  const allowed = (item) => canAccess(item.access, roles)
  const sections = ADMIN_SECTIONS
    .map((section) => ({ ...section, items: section.items.filter(allowed) }))
    .filter((section) => section.items.length > 0)

  if (sections.length === 0) return null

  const overview = allowed(ADMIN_OVERVIEW) ? ADMIN_OVERVIEW : null

  if (roles.isAdmin) {
    return { titleKey: 'nav.admin.title', grouped: true, overview, sections, items: [] }
  }

  return {
    titleKey: roles.isGroupAdmin ? 'groupAdmin.title' : hasAnyPermission(roles) ? 'nav.admin.title' : 'nav.admin.content',
    grouped: false,
    overview,
    sections: [],
    items: sections.flatMap((section) => section.items)
  }
}

/**
 * Onglets des pages qui regroupent un module et ses taxonomies.
 * Le premier onglet est la page parente ; un onglet invisible pour le rôle disparaît.
 */
export const SECTION_TABS = {
  news: [
    { to: '/admin/news', labelKey: 'nav.tabs.articles', access: NEWS, exact: true },
    { to: '/admin/news/categories', labelKey: 'nav.tabs.categories', access: ADMIN },
    { to: '/admin/news/types', labelKey: 'nav.tabs.types', access: ADMIN }
  ],
  suggestions: [
    { to: '/admin/suggestions', labelKey: 'nav.tabs.suggestions', access: ADMIN, exact: true },
    { to: '/admin/suggestions/categories', labelKey: 'nav.tabs.categories', access: ADMIN }
  ],
  resources: [
    { to: '/admin/resources', labelKey: 'nav.tabs.resources', access: RESOURCES, exact: true },
    { to: '/admin/resources/categories', labelKey: 'nav.tabs.categories', access: RESOURCES }
  ]
}

/**
 * @param {keyof typeof SECTION_TABS} section
 * @param {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean }} roles
 */
export function visibleTabs(section, roles) {
  return (SECTION_TABS[section] || []).filter((tab) => canAccess(tab.access, roles))
}
