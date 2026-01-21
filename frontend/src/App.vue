<template>
  <div class="app-layout">
    <!-- Sidebar (non zoomée) -->
    <Sidebar v-if="isAuthenticated && !isAuthPage" />

    <!-- Overlay mobile - s'affiche quand la sidebar est ouverte sur mobile/tablette -->
    <div
      v-if="isAuthenticated && !isAuthPage && appStore.sidebarOpen"
      @click="appStore.setSidebarOpen(false)"
      class="fixed inset-0 z-40 bg-black bg-opacity-50 lg:hidden"
    />

    <!-- Bouton hamburger mobile - visible seulement sur mobile/tablette quand sidebar fermée -->
    <button
      v-if="isAuthenticated && !isAuthPage && !appStore.sidebarOpen"
      @click="appStore.setSidebarOpen(true)"
      class="fixed top-4 left-4 z-50 p-2.5 bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 text-white rounded-lg shadow-lg transition-all duration-200 lg:hidden"
      :title="$t('common.expandSidebar')"
    >
      <Icon icon="mdi:menu" class="h-6 w-6" />
    </button>

    <!-- Main content -->
    <div :class="mainContentClasses">
      <!-- Zoom wrapper (contenu zoomé) -->
      <div v-if="isAuthenticated && !isAuthPage" class="zoom-wrapper" :style="zoomWrapperStyle">
        <!-- Loading global -->
        <LoadingOverlay v-if="appStore.isLoading" />

        <!-- Router view -->
        <router-view v-slot="{ Component, route }">
          <transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </transition>
        </router-view>
      </div>

      <!-- Auth pages (non zoomées) -->
      <template v-else>
        <!-- Loading global -->
        <LoadingOverlay v-if="appStore.isLoading" />

        <!-- Router view -->
        <router-view v-slot="{ Component, route }">
          <transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </transition>
        </router-view>
      </template>
    </div>

    <!-- Notifications -->
    <NotificationContainer />
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

// Components
import Sidebar from '@/components/layout/Sidebar.vue'
import LoadingOverlay from '@/components/ui/LoadingOverlay.vue'
import NotificationContainer from '@/components/ui/NotificationContainer.vue'

// Stores
const authStore = useAuthStore()
const appStore = useAppStore()
const route = useRoute()
const router = useRouter()

// Computed
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAuthPage = computed(() => route.path.startsWith('/auth'))

const mainContentClasses = computed(() => {
  if (isAuthenticated.value && !isAuthPage.value) {
    return ['main-content', { 'sidebar-collapsed': !appStore.sidebarOpen }]
  } else {
    return 'w-full'
  }
})

const zoomWrapperStyle = computed(() => {
  // Utiliser la propriété zoom CSS au lieu de transform scale
  // Cela préserve les coordonnées pour Vue DevTools
  return {
    zoom: appStore.zoomScale
  }
})

// Lifecycle
onMounted(async () => {
  // Charger les préférences depuis le localStorage
  const hasStoredAuth = authStore.loadFromStorage()
  appStore.loadFromStorage()

  // Si authentification stockée, rafraîchir le profil pour avoir les données à jour
  // (notamment admin_of_groups qui peut changer sans reconnexion)
  if (hasStoredAuth) {
    try {
      await authStore.updateProfile()
    } catch (error) {
      // Si erreur (token expiré, etc.), continuer normalement
      // Le refresh token sera utilisé si nécessaire
      console.log('Profil non rafraîchi:', error.message)
    }
  }

  // Si pas d'authentification stockée, tenter SSO auto-login
  // SAUF si on est sur un callback OAuth (Microsoft/Google)
  const isOAuthCallback = route.path.includes('/auth/oauth/') && route.query.code

  if (!hasStoredAuth && !isOAuthCallback) {
    try {
      const ssoResult = await authStore.autoLoginSSO()
      if (ssoResult) {
        // Rediriger vers la page d'accueil après SSO réussi
        // Toujours rediriger si on est sur une page d'auth
        if (route.path.startsWith('/auth')) {
          const redirectPath = route.query.redirect || '/home'
          await router.push(redirectPath)
        } else if (route.path === '/') {
          // Si on est à la racine, rediriger vers home
          await router.push('/home')
        }
      }
    } catch (error) {
      // Échec SSO silencieux (mode classique)
    }
  }

  // Initialiser le watcher du thème système
  const cleanupThemeWatcher = appStore.initSystemThemeWatcher()

  // Initialiser le watcher de redimensionnement pour la sidebar
  const cleanupResizeWatcher = appStore.initResizeWatcher()

  // Charger les settings de l'application (admin uniquement)
  if (authStore.isAuthenticated && authStore.isAdmin) {
    try {
      await appStore.refreshAppSettings()
    } catch (error) {
      console.error('Failed to load app settings on startup:', error)
    }
  }

  // Nettoyer au démontage
  onUnmounted(() => {
    if (cleanupThemeWatcher) {
      cleanupThemeWatcher()
    }
    if (cleanupResizeWatcher) {
      cleanupResizeWatcher()
    }
  })
})

</script>

<style>
/* Zoom wrapper */
.zoom-wrapper {
  /* Avec la propriété zoom CSS, pas besoin de compensation de taille */
  width: 100%;
  min-height: 100vh;
}

/* Transitions globales */
.page-enter-active,
.page-leave-active {
  transition: all 0.3s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateX(10px);
}

.page-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}

/* Loading overlay */
.loading-overlay {
  backdrop-filter: blur(2px);
}
</style>