import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

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
const Analytics = () => import('@/views/admin/Analytics.vue')
const AnnouncementsManagement = () => import('@/views/admin/AnnouncementsManagement.vue')
const NewsManagement = () => import('@/views/admin/NewsManagement.vue')
const NewsEditor = () => import('@/views/admin/NewsEditor.vue')
const PollsManagement = () => import('@/views/admin/PollsManagement.vue')
const GamificationSettings = () => import('@/views/admin/GamificationSettings.vue')
const SuggestionsManagement = () => import('@/views/admin/SuggestionsManagement.vue')
const MediaManagement = () => import('@/views/admin/MediaManagement.vue')
const NewsCategoriesManagement = () => import('@/views/admin/NewsCategoriesManagement.vue')
const NewsTypesManagement = () => import('@/views/admin/NewsTypesManagement.vue')

// Group Admin views
const GroupAdminDashboard = () => import('@/views/group-admin/GroupAdminDashboard.vue')
const GroupAdminAppGroups = () => import('@/views/group-admin/AppGroupsManagement.vue')
const GroupAdminApplications = () => import('@/views/group-admin/ApplicationsManagement.vue')
const GroupAdminNews = () => import('@/views/group-admin/NewsManagement.vue')
const GroupAdminNewsEditor = () => import('@/views/group-admin/NewsEditor.vue')
const GroupAdminPolls = () => import('@/views/group-admin/PollsManagement.vue')

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

// Events group admin views
const GroupAdminEvents = () => import('@/views/group-admin/EventsManagement.vue')
const GroupAdminEventEditor = () => import('@/views/group-admin/EventEditor.vue')

// Comment moderation
const CommentModeration = () => import('@/views/admin/CommentModeration.vue')

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
    path: '/dashboard',
    name: 'Dashboard',
    component: Dashboard,
    meta: {
      requiresAuth: true,
      title: 'Applications'
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
    path: '/admin',
    redirect: '/admin/settings'
  },
  {
    path: '/admin/app-groups',
    name: 'AdminAppGroups',
    component: AppGroupsManagement,
    meta: { 
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Groupes d\'applications'
    }
  },
  {
    path: '/admin/applications',
    name: 'AdminApplications',
    component: ApplicationsManagement,
    meta: { 
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Applications'
    }
  },
  {
    path: '/admin/users',
    name: 'AdminUsers',
    component: UsersManagement,
    meta: { 
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Utilisateurs'
    }
  },
  {
    path: '/admin/groups',
    name: 'AdminGroups',
    component: GroupsManagement,
    meta: { 
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Groupes'
    }
  },
  {
    path: '/admin/settings',
    name: 'AdminSettings',
    component: () => import('@/views/admin/SettingsManagement.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Paramètres'
    }
  },
  {
    path: '/admin/oauth',
    name: 'AdminOAuth',
    component: () => import('@/views/admin/OAuthSettings.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'OAuth Configuration'
    }
  },
  {
    path: '/admin/email',
    name: 'AdminEmail',
    component: () => import('@/views/admin/EmailSettings.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Notifications Email'
    }
  },
  {
    path: '/admin/analytics',
    name: 'AdminAnalytics',
    component: Analytics,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Analytics'
    }
  },
  {
    path: '/admin/announcements',
    name: 'AdminAnnouncements',
    component: AnnouncementsManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Announcements'
    }
  },
  {
    path: '/admin/news',
    name: 'AdminNews',
    component: NewsManagement,
    meta: {
      requiresAuth: true,
      requiresEditor: true,
      title: 'News Hub'
    }
  },
  {
    path: '/admin/news-categories',
    name: 'AdminNewsCategories',
    component: NewsCategoriesManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Catégories d\'articles'
    }
  },
  {
    path: '/admin/news-types',
    name: 'AdminNewsTypes',
    component: NewsTypesManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Types d\'articles'
    }
  },
  {
    path: '/admin/polls',
    name: 'AdminPolls',
    component: PollsManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Sondages'
    }
  },
  {
    path: '/admin/gamification',
    name: 'AdminGamification',
    component: GamificationSettings,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Gamification'
    }
  },
  {
    path: '/admin/suggestions',
    name: 'AdminSuggestions',
    component: SuggestionsManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Gestion suggestions'
    }
  },
  {
    path: '/admin/suggestion-categories',
    name: 'AdminSuggestionCategories',
    component: SuggestionCategoriesManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Categories suggestions'
    }
  },
  {
    path: '/admin/media',
    name: 'AdminMedia',
    component: MediaManagement,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Médias'
    }
  },
  {
    path: '/admin/news/:slug/edit',
    name: 'AdminNewsEdit',
    component: NewsEditor,
    meta: {
      requiresAuth: true,
      requiresEditor: true,
      title: 'Edit Article'
    }
  },
  {
    path: '/admin/news/new',
    name: 'AdminNewsNew',
    component: NewsEditor,
    meta: {
      requiresAuth: true,
      requiresEditor: true,
      title: 'Nouvel Article'
    }
  },
  {
    path: '/admin/events',
    name: 'AdminEvents',
    component: EventsManagement,
    meta: {
      requiresAuth: true,
      requiresEditor: true,
      title: 'Gestion des Événements'
    }
  },
  {
    path: '/admin/events/:slug/edit',
    name: 'AdminEventEdit',
    component: EventEditor,
    meta: {
      requiresAuth: true,
      requiresEditor: true,
      title: 'Modifier l\'Événement'
    }
  },
  {
    path: '/admin/events/new',
    name: 'AdminEventNew',
    component: EventEditor,
    meta: {
      requiresAuth: true,
      requiresEditor: true,
      title: 'Nouvel Événement'
    }
  },
  {
    path: '/admin/comments',
    name: 'CommentModeration',
    component: CommentModeration,
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Modération des Commentaires'
    }
  },
  // Routes Group Admin
  {
    path: '/group-admin',
    name: 'GroupAdminDashboard',
    component: GroupAdminDashboard,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Administration de Groupe'
    }
  },
  {
    path: '/group-admin/app-groups',
    name: 'GroupAdminAppGroups',
    component: GroupAdminAppGroups,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Mes Groupes d\'Applications'
    }
  },
  {
    path: '/group-admin/applications',
    name: 'GroupAdminApplications',
    component: GroupAdminApplications,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Mes Applications'
    }
  },
  {
    path: '/group-admin/news',
    name: 'GroupAdminNews',
    component: GroupAdminNews,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'News de Groupe'
    }
  },
  {
    path: '/group-admin/news/new',
    name: 'GroupAdminNewsNew',
    component: GroupAdminNewsEditor,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Nouvel Article de Groupe'
    }
  },
  {
    path: '/group-admin/news/:slug/edit',
    name: 'GroupAdminNewsEdit',
    component: GroupAdminNewsEditor,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Modifier l\'Article'
    }
  },
  {
    path: '/group-admin/events',
    name: 'GroupAdminEvents',
    component: GroupAdminEvents,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Événements de Groupe'
    }
  },
  {
    path: '/group-admin/events/new',
    name: 'GroupAdminEventNew',
    component: GroupAdminEventEditor,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Nouvel Événement de Groupe'
    }
  },
  {
    path: '/group-admin/events/:slug/edit',
    name: 'GroupAdminEventEdit',
    component: GroupAdminEventEditor,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Modifier l\'Événement'
    }
  },
  {
    path: '/group-admin/polls',
    name: 'GroupAdminPolls',
    component: GroupAdminPolls,
    meta: {
      requiresAuth: true,
      requiresGroupAdmin: true,
      title: 'Mes Sondages'
    }
  },
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

// Guards de navigation
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  // Charger l'état d'authentification depuis le localStorage SEULEMENT au premier chargement
  if (!authStore.isAuthenticated && !authStore.isLoading && !authStore.token) {
    authStore.loadFromStorage()
  }

  // Définir le titre de la page
  if (to.meta.title) {
    document.title = `${to.meta.title} - Airboard`
  } else {
    document.title = 'Airboard - Portail Applicatif'
  }

  // Vérifier l'authentification
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({
      name: 'Login',
      query: { redirect: to.fullPath }
    })
    return
  }

  // Vérifier les droits admin
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next({ name: 'Unauthorized' })
    return
  }

  // Vérifier les droits editor (admin ou editor)
  if (to.meta.requiresEditor) {
    const userRole = authStore.user?.role
    if (userRole !== 'admin' && userRole !== 'editor') {
      next({ name: 'Unauthorized' })
      return
    }
  }

  // Vérifier les droits group admin (admin ou utilisateur admin d'au moins un groupe)
  if (to.meta.requiresGroupAdmin) {
    const userRole = authStore.user?.role
    // Vérifier admin_of_groups (tableau de groupes) OU managed_group_ids (tableau d'IDs)
    const adminOfGroups = authStore.user?.admin_of_groups
    const managedGroupIds = authStore.user?.managed_group_ids
    const isAdminOfGroups = (adminOfGroups && Array.isArray(adminOfGroups) && adminOfGroups.length > 0) ||
                            (managedGroupIds && Array.isArray(managedGroupIds) && managedGroupIds.length > 0)

    console.log('🔐 Guard requiresGroupAdmin:', {
      route: to.path,
      userRole,
      adminOfGroups,
      managedGroupIds,
      isAdminOfGroups
    })

    // Si l'utilisateur n'est pas admin et n'a pas de groupes administrés localement,
    // rafraîchir le profil pour vérifier les permissions à jour
    if (userRole !== 'admin' && !isAdminOfGroups) {
      try {
        await authStore.updateProfile()

        // Re-vérifier les permissions avec les données fraîches
        const updatedUserRole = authStore.user?.role
        const updatedAdminOfGroups = authStore.user?.admin_of_groups
        const updatedManagedGroupIds = authStore.user?.managed_group_ids
        const updatedIsAdminOfGroups = (updatedAdminOfGroups && Array.isArray(updatedAdminOfGroups) && updatedAdminOfGroups.length > 0) ||
                                       (updatedManagedGroupIds && Array.isArray(updatedManagedGroupIds) && updatedManagedGroupIds.length > 0)

        if (updatedUserRole !== 'admin' && !updatedIsAdminOfGroups) {
          next({ name: 'Unauthorized' })
          return
        }
      } catch (error) {
        console.error('Erreur lors de la mise à jour du profil:', error)
        next({ name: 'Unauthorized' })
        return
      }
    }
  }

  // Rediriger les utilisateurs connectés loin des pages d'auth
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    next({ name: 'Dashboard' })
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
