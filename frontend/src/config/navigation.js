import { ADMIN, GROUP_ADMIN, CONTENT, canAccess } from './access'

/**
 * Menu latéral, décrit une seule fois pour tous les rôles.
 *
 * Chaque lien porte un niveau d'accès (`access`, voir config/access.js) ; le menu
 * d'un utilisateur ne garde que ce qu'il peut ouvrir. Les libellés sont des clés i18n.
 */

/** Espace utilisateur : visible par tous, jamais filtré */
export const USER_NAV = [
  { to: '/home', icon: 'mdi:home', labelKey: 'common.home' },
  { to: '/resources', icon: 'mdi:application', labelKey: 'common.applications' },
  { to: '/news', icon: 'mdi:newspaper', labelKey: 'common.newsHub' },
  { to: '/events', icon: 'mdi:calendar', labelKey: 'common.events' },
  { to: '/polls', icon: 'mdi:poll', labelKey: 'common.polls' },
  { to: '/suggestions', icon: 'mdi:lightbulb-on-outline', labelKey: 'common.suggestions' },
  { to: '/gamification', icon: 'mdi:medal', labelKey: 'common.gamification' },
  { to: '/directory', icon: 'mdi:account-group-outline', labelKey: 'directory.title' },
  { to: '/search', icon: 'mdi:magnify', labelKey: 'search.title' }
]

/** Page d'accueil de l'administration */
export const ADMIN_OVERVIEW = {
  to: '/admin',
  icon: 'mdi:view-dashboard-outline',
  labelKey: 'nav.admin.overview',
  access: GROUP_ADMIN,
  exact: true
}

/** Administration, rangée par domaine */
export const ADMIN_SECTIONS = [
  {
    id: 'content',
    labelKey: 'nav.admin.content',
    icon: 'mdi:file-document-multiple',
    items: [
      { to: '/admin/news', icon: 'mdi:newspaper', labelKey: 'nav.admin.news', access: CONTENT },
      { to: '/admin/events', icon: 'mdi:calendar-edit', labelKey: 'nav.admin.events', access: CONTENT },
      { to: '/admin/polls', icon: 'mdi:poll', labelKey: 'nav.admin.polls', access: GROUP_ADMIN },
      { to: '/admin/announcements', icon: 'mdi:bullhorn', labelKey: 'nav.admin.announcements', access: ADMIN },
      { to: '/admin/media', icon: 'mdi:image-multiple', labelKey: 'nav.admin.media', access: ADMIN }
    ]
  },
  {
    id: 'community',
    labelKey: 'nav.admin.community',
    icon: 'mdi:account-voice',
    items: [
      { to: '/admin/suggestions', icon: 'mdi:lightbulb-on-outline', labelKey: 'nav.admin.suggestions', access: ADMIN },
      { to: '/admin/comments', icon: 'mdi:comment-text-outline', labelKey: 'nav.admin.moderation', access: ADMIN },
      { to: '/admin/gamification', icon: 'mdi:medal-outline', labelKey: 'nav.admin.gamification', access: ADMIN },
      { to: '/admin/directory', icon: 'mdi:sitemap-outline', labelKey: 'nav.admin.directory', access: ADMIN }
    ]
  },
  {
    id: 'resources',
    labelKey: 'nav.admin.resources',
    icon: 'mdi:apps-box',
    items: [
      { to: '/admin/resources', icon: 'mdi:apps', labelKey: 'nav.admin.resources', access: GROUP_ADMIN }
    ]
  },
  {
    id: 'users',
    labelKey: 'nav.admin.users',
    icon: 'mdi:account-multiple',
    items: [
      { to: '/admin/users', icon: 'mdi:account-multiple', labelKey: 'nav.admin.users', access: ADMIN },
      { to: '/admin/groups', icon: 'mdi:account-group', labelKey: 'nav.admin.groups', access: ADMIN },
      { to: '/admin/oauth', icon: 'mdi:shield-key', labelKey: 'nav.admin.oauth', access: ADMIN }
    ]
  },
  {
    id: 'system',
    labelKey: 'nav.admin.system',
    icon: 'mdi:cog',
    items: [
      { to: '/admin/settings', icon: 'mdi:cog', labelKey: 'nav.admin.settings', access: ADMIN },
      { to: '/admin/email', icon: 'mdi:email-outline', labelKey: 'nav.admin.email', access: ADMIN },
      { to: '/admin/analytics', icon: 'mdi:chart-line', labelKey: 'nav.admin.analytics', access: ADMIN },
      { to: '/admin/reports', icon: 'mdi:chart-timeline-variant', labelKey: 'nav.admin.reports', access: ADMIN },
      { to: '/admin/chatbots', icon: 'mdi:robot-outline', labelKey: 'nav.admin.chatbots', access: ADMIN }
    ]
  }
]

/**
 * Menu d'administration d'un utilisateur.
 *
 * - Admin global : vue d'ensemble puis domaines repliables (`grouped: true`).
 * - Admin de groupe ou éditeur : une liste courte des seules pages permises,
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
    titleKey: roles.isGroupAdmin ? 'groupAdmin.title' : 'nav.admin.content',
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
    { to: '/admin/news', labelKey: 'nav.tabs.articles', access: CONTENT, exact: true },
    { to: '/admin/news/categories', labelKey: 'nav.tabs.categories', access: ADMIN },
    { to: '/admin/news/types', labelKey: 'nav.tabs.types', access: ADMIN }
  ],
  suggestions: [
    { to: '/admin/suggestions', labelKey: 'nav.tabs.suggestions', access: ADMIN, exact: true },
    { to: '/admin/suggestions/categories', labelKey: 'nav.tabs.categories', access: ADMIN }
  ],
  resources: [
    { to: '/admin/resources', labelKey: 'nav.tabs.resources', access: GROUP_ADMIN, exact: true },
    { to: '/admin/resources/categories', labelKey: 'nav.tabs.categories', access: GROUP_ADMIN }
  ]
}

/**
 * @param {keyof typeof SECTION_TABS} section
 * @param {{ isAdmin: boolean, isEditor: boolean, isGroupAdmin: boolean }} roles
 */
export function visibleTabs(section, roles) {
  return (SECTION_TABS[section] || []).filter((tab) => canAccess(tab.access, roles))
}
