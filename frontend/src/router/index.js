import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ADMIN, GROUP_ADMIN, RESOURCES, NEWS, EVENTS, getRoles, canAccess } from '@/config/access'

// Import des vues
const Dashboard = () => import('@/views/dashboard/Dashboard.vue')
const Login = () => import('@/views/auth/Login.vue')
const Register = () => import('@/views/auth/Register.vue')
const OAuthCallback = () => import('@/views/auth/OAuthCallback.vue')

// Admin views
const AppGroupsManagement = () => import('@/views/admin/AppGroupsManagement.vue')
const ApplicationsManagement = () => import('@/views/admin/ApplicationsManagement.vue')
const UsersManagement = () => import('@/views/admin/UsersManagement.vue')
const GroupsManagement = () => import('@/views/admin/GroupsManagement.vue')
const RolesManagement = () => import('@/views/admin/RolesManagement.vue')
const Analytics = () => import('@/views/admin/Analytics.vue')
const AnnouncementsManagement = () => import('@/views/admin/AnnouncementsManagement.vue')
const ChatbotsManagement = () => import('@/views/admin/ChatbotsManagement.vue')
const NewsManagement = () => import('@/views/admin/NewsManagement.vue')
const NewsEditor = () => import('@/views/admin/NewsEditor.vue')
const PollsManagement = () => import('@/views/admin/PollsManagement.vue')
const GamificationSettings = () => import('@/views/admin/GamificationSettings.vue')
const SuggestionsManagement = () => import('@/views/admin/SuggestionsManagement.vue')
const MediaManagement = () => import('@/views/admin/MediaManagement.vue')
const NewsCategoriesManagement = () => import('@/views/admin/NewsCategoriesManagement.vue')
const NewsTypesManagement = () => import('@/views/admin/NewsTypesManagement.vue')

// Page parente à onglets (module + taxonomies)
const SectionTabs = () => import('@/components/layout/SectionTabs.vue')

// Group Admin views
const AdminOverview = () => import('@/views/admin/AdminOverview.vue')

// News views (public)
const NewsCenter = () => import('@/views/NewsCenter.vue')
const NewsDetail = () => import('@/views/NewsDetail.vue')

// Polls views (public)
const PollsPage = () => import('@/views/PollsPage.vue')
const PollDetailPage = () => import('@/views/PollDetailPage.vue')
const SuggestionsPage = () => import('@/views/SuggestionsPage.vue')
const SuggestionDetailPage = () => import('@/views/SuggestionDetailPage.vue')
const SuggestionCategoriesManagement = () => import('@/views/admin/SuggestionCategoriesManagement.vue')

// Events views (public)
const EventsCenter = () => import('@/views/EventsCenter.vue')
const EventDetail = () => import('@/views/EventDetail.vue')

// Events admin views
const EventsManagement = () => import('@/views/admin/EventsManagement.vue')
const EventEditor = () => import('@/views/admin/EventEditor.vue')

// Comment moderation
const CommentModeration = () => import('@/views/admin/CommentModeration.vue')

// Trombinoscope
const DirectoryPage = () => import('@/views/DirectoryPage.vue')
const DirectoryManagement = () => import('@/views/admin/DirectoryManagement.vue')

// Search
const SearchPage = () => import('@/views/SearchPage.vue')

// Error views
const NotFound = () => import('@/views/errors/NotFound.vue')
const Unauthorized = () => import('@/views/errors/Unauthorized.vue')

const routes = [
  {
    path: '/',
    redirect: '/home'
  },
  {
    path: '/home',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: {
      requiresAuth: true,
      title: 'Accueil'
    }
  },
  {
    path: '/resources',
    name: 'Resources',
    component: Dashboard,
    meta: {
      requiresAuth: true,
      title: 'Ressources'
    }
  },
  { path: '/dashboard', redirect: '/resources' },
  {
    path: '/directory',
    name: 'Directory',
    component: DirectoryPage,
    meta: {
      requiresAuth: true,
      title: 'Trombinoscope'
    }
  },
  {
    path: '/admin/directory',
    name: 'DirectoryManagement',
    component: DirectoryManagement,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Gestion du trombinoscope'
    }
  },
  {
    path: '/search',
    name: 'Search',
    component: SearchPage,
    meta: {
      requiresAuth: true,
      title: 'Recherche'
    }
  },
  {
    path: '/news',
    name: 'NewsCenter',
    component: NewsCenter,
    meta: {
      requiresAuth: true,
      title: 'News Hub'
    }
  },
  {
    path: '/news/:slug',
    name: 'NewsDetail',
    component: NewsDetail,
    meta: {
      requiresAuth: true,
      title: 'Article'
    }
  },
  {
    path: '/polls',
    name: 'Polls',
    component: PollsPage,
    meta: {
      requiresAuth: true,
      title: 'Sondages'
    }
  },
  {
    path: '/polls/:id',
    name: 'PollDetail',
    component: PollDetailPage,
    meta: {
      requiresAuth: true,
      title: 'Sondage'
    }
  },
  {
    path: '/suggestions',
    name: 'Suggestions',
    component: SuggestionsPage,
    meta: {
      requiresAuth: true,
      title: 'Boite a suggestions'
    }
  },
  {
    path: '/suggestions/:id',
    name: 'SuggestionDetail',
    component: SuggestionDetailPage,
    meta: {
      requiresAuth: true,
      title: 'Suggestion'
    }
  },
  {
    path: '/events',
    name: 'EventsCenter',
    component: EventsCenter,
    meta: {
      requiresAuth: true,
      title: 'Agenda'
    }
  },
  {
    path: '/profile',
    name: 'Profile',
    component: () => import('@/views/ProfileView.vue'),
    meta: {
      requiresAuth: true,
      title: 'Profil'
    }
  },
  {
    path: '/gamification',
    name: 'GamificationHub',
    component: () => import('@/views/GamificationHub.vue'),
    meta: {
      requiresAuth: true,
      title: 'Centre de Gamification'
    }
  },
  {
    path: '/notifications',
    name: 'Notifications',
    component: () => import('@/views/Notifications.vue'),
    meta: {
      requiresAuth: true,
      title: 'Notifications'
    }
  },
  {
    path: '/events/:slug',
    name: 'EventDetail',
    component: EventDetail,
    meta: {
      requiresAuth: true,
      title: 'Événement'
    }
  },
  {
    path: '/auth/login',
    name: 'Login',
    component: Login,
    meta: { 
      requiresGuest: true,
      title: 'Connexion'
    }
  },
  {
    path: '/auth/register',
    name: 'Register',
    component: Register,
    meta: {
      requiresGuest: true,
      title: 'Inscription'
    }
  },
  {
    path: '/auth/oauth/:provider/callback',
    name: 'OAuthCallback',
    component: OAuthCallback,
    meta: {
      title: 'Authentification OAuth'
    }
  },
  {
    // Vue d'ensemble, commune à l'admin et à l'admin de groupe
    path: '/admin',
    name: 'AdminOverview',
    component: AdminOverview,
    meta: {
      requiresAuth: true,
      access: GROUP_ADMIN,
      title: "Vue d'ensemble"
    }
  },
  // Modules et leurs taxonomies, rangés en onglets (config/navigation.js : SECTION_TABS)
  {
    path: '/admin/news',
    component: SectionTabs,
    props: { section: 'news' },
    meta: { requiresAuth: true, access: NEWS },
    children: [
      { path: '', name: 'AdminNews', component: NewsManagement, meta: { title: 'Actualités' } },
      { path: 'categories', name: 'AdminNewsCategories', component: NewsCategoriesManagement, meta: { access: ADMIN, title: 'Catégories d\'articles' } },
      { path: 'types', name: 'AdminNewsTypes', component: NewsTypesManagement, meta: { access: ADMIN, title: 'Types d\'articles' } }
    ]
  },
  {
    path: '/admin/suggestions',
    component: SectionTabs,
    props: { section: 'suggestions' },
    meta: { requiresAuth: true, access: ADMIN },
    children: [
      { path: '', name: 'AdminSuggestions', component: SuggestionsManagement, meta: { title: 'Gestion suggestions' } },
      { path: 'categories', name: 'AdminSuggestionCategories', component: SuggestionCategoriesManagement, meta: { title: 'Catégories de suggestions' } }
    ]
  },
  {
    path: '/admin/resources',
    component: SectionTabs,
    props: { section: 'resources' },
    meta: { requiresAuth: true, access: RESOURCES },
    children: [
      { path: '', name: 'AdminApplications', component: ApplicationsManagement, meta: { title: 'Ressources' } },
      { path: 'categories', name: 'AdminAppGroups', component: AppGroupsManagement, meta: { title: 'Catégories de ressources' } }
    ]
  },
  // Anciennes adresses des taxonomies et des ressources
  { path: '/admin/news-categories', redirect: '/admin/news/categories' },
  { path: '/admin/news-types', redirect: '/admin/news/types' },
  { path: '/admin/suggestion-categories', redirect: '/admin/suggestions/categories' },
  { path: '/admin/app-groups', redirect: '/admin/resources/categories' },
  { path: '/admin/applications', redirect: '/admin/resources' },
  { path: '/group-admin/directory', redirect: '/admin/directory' },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: UsersManagement,
    meta: { 
      requiresAuth: true,
      access: ADMIN,
      title: 'Utilisateurs'
    }
  },
  {
    path: '/admin/groups',
    name: 'AdminGroups',
    component: GroupsManagement,
    meta: { 
      requiresAuth: true,
      access: ADMIN,
      title: 'Groupes'
    }
  },
  {
    path: '/admin/roles',
    name: 'AdminRoles',
    component: RolesManagement,
    meta: { requiresAuth: true, access: ADMIN, title: 'Rôles' }
  },
  {
    path: '/admin/settings',
    name: 'AdminSettings',
    component: () => import('@/views/admin/SettingsManagement.vue'),
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Paramètres'
    }
  },
  {
    path: '/admin/oauth',
    name: 'AdminOAuth',
    component: () => import('@/views/admin/OAuthSettings.vue'),
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'OAuth Configuration'
    }
  },
  {
    path: '/admin/email',
    name: 'AdminEmail',
    component: () => import('@/views/admin/EmailSettings.vue'),
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Notifications Email'
    }
  },
  {
    path: '/admin/analytics',
    name: 'AdminAnalytics',
    component: Analytics,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Analytics'
    }
  },
  {
    path: '/admin/reports',
    name: 'AdminReports',
    component: () => import('@/views/admin/Reports.vue'),
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Reports'
    }
  },
  {
    path: '/admin/announcements',
    name: 'AdminAnnouncements',
    component: AnnouncementsManagement,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Announcements'
    }
  },
  {
    path: '/admin/chatbots',
    name: 'AdminChatbots',
    component: ChatbotsManagement,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Chatbots IA'
    }
  },
  {
    path: '/admin/polls',
    name: 'AdminPolls',
    component: PollsManagement,
    meta: {
      requiresAuth: true,
      access: GROUP_ADMIN,
      title: 'Sondages'
    }
  },
  {
    path: '/admin/gamification',
    name: 'AdminGamification',
    component: GamificationSettings,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Gamification'
    }
  },
  {
    path: '/admin/media',
    name: 'AdminMedia',
    component: MediaManagement,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Médias'
    }
  },
  {
    path: '/admin/news/:slug/edit',
    name: 'AdminNewsEdit',
    component: NewsEditor,
    meta: {
      requiresAuth: true,
      access: NEWS,
      title: 'Modifier l\'article'
    }
  },
  {
    path: '/admin/news/new',
    name: 'AdminNewsNew',
    component: NewsEditor,
    meta: {
      requiresAuth: true,
      access: NEWS,
      title: 'Nouvel Article'
    }
  },
  {
    path: '/admin/events',
    name: 'AdminEvents',
    component: EventsManagement,
    meta: {
      requiresAuth: true,
      access: EVENTS,
      title: 'Gestion des Événements'
    }
  },
  {
    path: '/admin/events/:slug/edit',
    name: 'AdminEventEdit',
    component: EventEditor,
    meta: {
      requiresAuth: true,
      access: EVENTS,
      title: 'Modifier l\'Événement'
    }
  },
  {
    path: '/admin/events/new',
    name: 'AdminEventNew',
    component: EventEditor,
    meta: {
      requiresAuth: true,
      access: EVENTS,
      title: 'Nouvel Événement'
    }
  },
  {
    path: '/admin/comments',
    name: 'CommentModeration',
    component: CommentModeration,
    meta: {
      requiresAuth: true,
      access: ADMIN,
      title: 'Modération des Commentaires'
    }
  },
  // Anciennes adresses de l'administration de groupe : mêmes pages que l'admin,
  // limitées aux groupes administrés (voir useManagementScope)
  { path: '/group-admin', redirect: '/admin' },
  { path: '/group-admin/app-groups', redirect: '/admin/resources/categories' },
  { path: '/group-admin/applications', redirect: '/admin/resources' },
  { path: '/group-admin/news', redirect: '/admin/news' },
  { path: '/group-admin/news/new', redirect: '/admin/news/new' },
  { path: '/group-admin/news/:slug/edit', redirect: (to) => `/admin/news/${to.params.slug}/edit` },
  { path: '/group-admin/events', redirect: '/admin/events' },
  { path: '/group-admin/events/new', redirect: '/admin/events/new' },
  { path: '/group-admin/events/:slug/edit', redirect: (to) => `/admin/events/${to.params.slug}/edit` },
  { path: '/group-admin/polls', redirect: '/admin/polls' },
  {
    path: '/unauthorized',
    name: 'Unauthorized',
    component: Unauthorized,
    meta: {
      title: 'Accès non autorisé'
    }
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: NotFound,
    meta: {
      title: 'Page non trouvée'
    }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

/**
 * L'utilisateur atteint-il le niveau d'accès demandé ?
 * Un admin de groupe tout juste nommé n'a pas encore ses groupes dans le profil
 * en mémoire : on relit le profil une fois avant de refuser.
 */
async function hasAccess(authStore, access) {
  if (canAccess(access, getRoles(authStore.user))) return true
  if (access === ADMIN || authStore.user?.role === 'admin') return false
  try {
    await authStore.updateProfile()
  } catch (error) {
    console.error('Erreur lors de la mise à jour du profil:', error)
    return false
  }
  return canAccess(access, getRoles(authStore.user))
}

// Guards de navigation
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Charger l'état d'authentification depuis le localStorage SEULEMENT au premier chargement
  if (!authStore.isAuthenticated && !authStore.isLoading && !authStore.token) {
    authStore.loadFromStorage()
  }

  // Le titre de l'onglet (traduit) est posé par App.vue

  // Vérifier l'authentification
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({
      name: 'Login',
      query: { redirect: to.fullPath }
    })
    return
  }

  // Vérifier le niveau d'accès de la page (voir config/access.js)
  if (to.meta.access && !(await hasAccess(authStore, to.meta.access))) {
    next({ name: 'Unauthorized' })
    return
  }

  // Rediriger les utilisateurs connectés loin des pages d'auth
  // Déjà connecté : même page d'arrivée qu'après la connexion (Login.vue)
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    next({ name: 'Home' })
    return
  }

  next()
})

// Gérer les erreurs de navigation
router.onError((error) => {
  console.error('Erreur de navigation:', error)
  
  // Rediriger vers la page d'erreur pour les erreurs de chargement de composant
  if (error.message.includes('Failed to fetch dynamically imported module')) {
    window.location.reload()
  }
})

export default router
