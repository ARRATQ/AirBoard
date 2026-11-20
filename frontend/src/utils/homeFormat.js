// Calculs et formats de l'accueil, sans dépendance à Vue : testés dans __tests__/homeFormat.test.js.

const DAY_MS = 86_400_000
const DATE_LOCALES = { fr: 'fr-FR', en: 'en-US', es: 'es-ES', ar: 'ar-MA' }

/** Langue de l'appli → locale des dates (ar → ar-MA, comme le reste de l'appli). */
export function dateLocaleFor(lang) {
  return DATE_LOCALES[lang] || lang
}

const figure = (key, labelKey, value, to) => ({ key, labelKey, value: value ?? 0, to })

/**
 * Chiffres clés du bandeau, selon le rôle.
 * @param {object|undefined} stats  homeData.stats
 * @param {string} role             homeData.user_role
 * @param {number[]} managedGroupIds homeData.managed_group_ids
 * @returns {{ key: string, labelKey: string, value: number, to: string }[]}
 */
export function keyFigures(stats, role, managedGroupIds) {
  if (!stats) return []

  if (role === 'admin') {
    return [
      figure('users', 'home.stats.users', stats.total_users, '/admin/users'),
      figure('groups', 'home.stats.groups', stats.total_groups, '/admin/groups'),
      figure('appGroups', 'home.stats.appGroups', stats.total_app_groups, '/admin/resources/categories'),
      figure('apps', 'home.stats.apps', stats.total_apps, '/admin/resources'),
      figure('news', 'home.stats.news', stats.total_news, '/admin/news'),
      figure('events', 'home.stats.events', stats.total_events, '/admin/events'),
      figure('polls', 'home.stats.polls', stats.total_polls, '/admin/polls'),
    ]
  }

  if (managedGroupIds?.length) {
    return [
      figure('managedGroups', 'home.stats.managedGroups', stats.managed_groups_count, '/admin'),
      figure('members', 'home.stats.totalMembers', stats.total_members_count, '/admin'),
      figure('appGroups', 'home.stats.managedAppGroups', stats.managed_app_groups_count, '/admin/resources/categories'),
      figure('apps', 'home.stats.managedApps', stats.managed_apps_count, '/admin/resources'),
      figure('news', 'home.stats.managedNews', stats.managed_news_count, '/admin/news'),
      figure('polls', 'home.stats.managedPolls', stats.managed_polls_count, '/admin/polls'),
      figure('accessibleApps', 'home.stats.accessibleApps', stats.total_accessible_apps, '/resources'),
    ]
  }

  return [
    figure('apps', 'home.stats.apps', stats.total_accessible_apps, '/resources'),
    figure('news', 'home.stats.news', stats.total_news, '/news'),
    figure('polls', 'home.stats.polls', stats.total_polls, '/polls'),
    figure('events', 'home.stats.events', stats.total_accessible_events, '/events'),
  ]
}

const optionVotes = (opt) => opt?.votes || 0

/** Part d'une option, en pourcentage entier (0 si personne n'a voté). */
export function pollPercent(opt, poll) {
  const total = poll.total_votes || (poll.options || []).reduce((sum, o) => sum + optionVotes(o), 0)
  if (!total) return 0
  return Math.round((optionVotes(opt) / total) * 100)
}

/** Option en tête (au moins un vote). */
export function isLeadingOption(opt, poll) {
  const max = Math.max(0, ...(poll.options || []).map(optionVotes))
  return max > 0 && optionVotes(opt) === max
}

/** Jours restants avant une date de fin (arrondi supérieur) ; null sans date. */
export function daysUntil(endDate, now = new Date()) {
  if (!endDate) return null
  return Math.ceil((new Date(endDate) - now) / DAY_MS)
}

const startOfDay = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate())

/**
 * « aujourd'hui », « hier », « il y a 3 jours », puis date courte au-delà d'une semaine.
 * @param {string} dateStr
 * @param {string} locale  locale de date (fr-FR, ar-MA…)
 */
export function relativeDay(dateStr, locale, now = new Date()) {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  if (Number.isNaN(date.getTime())) return ''
  const days = Math.round((startOfDay(now) - startOfDay(date)) / DAY_MS)
  if (days >= 0 && days < 7) {
    return new Intl.RelativeTimeFormat(locale, { numeric: 'auto' }).format(-days, 'day')
  }
  return date.toLocaleDateString(locale, { day: 'numeric', month: 'short' })
}
