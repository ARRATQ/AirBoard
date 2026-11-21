<template>
  <div class="app-layout">
    <!-- Sidebar (non zoomée) -->
    <Sidebar v-if="isAuthenticated && !isAuthPage" />

    <!-- Overlay mobile - s'affiche quand la sidebar est ouverte sur mobile/tablette -->
    <div
      v-if="isAuthenticated && !isAuthPage && appStore.sidebarOpen"
      aria-hidden="true"
      @click="appStore.setSidebarOpen(false)"
      class="fixed inset-0 z-40 bg-black bg-opacity-50 lg:hidden"
    />

    <!-- Main content -->
    <div ref="mainContent" :class="mainContentClasses">
      <!-- Barre mobile : le bouton menu occupe sa propre bande au lieu de flotter sur les titres -->
      <div v-if="isAuthenticated && !isAuthPage" class="mobile-bar lg:hidden">
        <button
          type="button"
          @click="appStore.setSidebarOpen(true)"
          class="mobile-bar__btn"
          :title="$t('common.expandSidebar')"
          :aria-label="$t('common.expandSidebar')"
          :aria-expanded="appStore.sidebarOpen"
        >
          <Icon icon="mdi:menu" class="h-6 w-6" />
        </button>
      </div>

      <!-- Zoom wrapper (contenu zoomé) -->
      <div v-if="isAuthenticated && !isAuthPage" class="zoom-wrapper" :style="zoomWrapperStyle">
        <!-- Loading global -->
        <LoadingOverlay v-if="appStore.isLoading" />

        <!-- Router view -->
        <router-view v-slot="{ Component, route }">
          <transition name="page">
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
          <transition name="page">
            <component :is="Component" :key="route.path" />
          </transition>
        </router-view>
      </template>
    </div>

    <!-- Notifications -->
    <NotificationContainer />

    <!-- Confirmation (await confirm(…) via composables/useConfirm) -->
    <ConfirmDialog />

    <!-- Speed Dial FAB (chat + chatbots fusionnés en un seul bouton flottant) -->
    <FloatingSpeedDial v-if="isAuthenticated && !isAuthPage" />
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { pageTitle } from '@/utils/pageTitle'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

// Components
import Sidebar from '@/components/layout/Sidebar.vue'
import LoadingOverlay from '@/components/ui/LoadingOverlay.vue'
import NotificationContainer from '@/components/ui/NotificationContainer.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import FloatingSpeedDial from '@/components/layout/FloatingSpeedDial.vue'

// Stores
const authStore = useAuthStore()
const appStore = useAppStore()
const route = useRoute()
const router = useRouter()
const { t, te } = useI18n()

// Titre de l'onglet traduit, mis à jour à chaque navigation et changement de langue
watchEffect(() => {
  document.title = pageTitle(route, t, te)
})

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

// Raccourci clavier Ctrl+K pour la recherche
const handleKeydown = (e) => {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    if (authStore.isAuthenticated) {
      router.push('/search')
    }
  }
}

// Palette de couleurs publique (pour tous les utilisateurs authentifiés)
const loadPalette = async () => {
  try {
    const { settingsService } = await import('@/services/api')
    const publicSettings = await settingsService.getPublicSettings()
    if (publicSettings.custom_palettes) {
      try {
        const cp = typeof publicSettings.custom_palettes === 'string'
          ? JSON.parse(publicSettings.custom_palettes)
          : publicSettings.custom_palettes
        appStore.setCustomPalettes(Array.isArray(cp) ? cp : [])
      } catch (e) {
        console.warn('Palettes personnalisées illisibles:', e)
      }
    }
    if (publicSettings.color_palette) {
      const customHex = publicSettings.color_palette === 'custom' ? publicSettings.custom_accent_color : null
      appStore.applyPalette(publicSettings.color_palette, customHex)
    }
  } catch (error) {
    // Silencieux — la palette localStorage reste active comme fallback
  }
}

const loadAdminSettings = async () => {
  if (!authStore.isAdmin) return
  try {
    await appStore.ensureAppSettings()
  } catch (error) {
    console.error('Failed to load app settings on startup:', error)
  }
}

let cleanupThemeWatcher = null
let cleanupResizeWatcher = null

// Taille d'affichage automatique : largeur de la zone de contenu (hors barre latérale, non
// zoomée) et de la fenêtre, transmises au store qui en déduit l'échelle.
const mainContent = ref(null)
let contentObserver = null
function observeContentWidth () {
  if (!mainContent.value || typeof ResizeObserver === 'undefined') return
  contentObserver = new ResizeObserver(([entry]) => {
    appStore.setLayoutWidths(Math.round(entry.contentRect.width), window.innerWidth)
  })
  contentObserver.observe(mainContent.value)
}

// Enregistré avant tout await : un onUnmounted appelé après un await n'est pas rattaché au composant.
onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  if (cleanupThemeWatcher) cleanupThemeWatcher()
  if (cleanupResizeWatcher) cleanupResizeWatcher()
  contentObserver?.disconnect()
})

// Lifecycle
onMounted(async () => {
  // Charger les préférences depuis le localStorage
  const hasStoredAuth = authStore.loadFromStorage()
  appStore.loadFromStorage()

  window.addEventListener('keydown', handleKeydown)
  cleanupThemeWatcher = appStore.initSystemThemeWatcher()
  cleanupResizeWatcher = appStore.initResizeWatcher()
  observeContentWidth()

  if (hasStoredAuth) {
    // Profil (admin_of_groups peut changer sans reconnexion), palette et réglages admin
    // partent en parallèle : aucun ne dépend des autres.
    await Promise.all([
      authStore.updateProfile().catch((error) => {
        // Token expiré, etc. : le refresh token prendra le relais si possible
        console.warn('Profil non rafraîchi:', error.message)
      }),
      loadPalette(),
      loadAdminSettings()
    ])
    return
  }

  // Pas d'authentification stockée : tenter le SSO auto-login,
  // SAUF sur un callback OAuth (Microsoft/Google)
  const isOAuthCallback = route.path.includes('/auth/oauth/') && route.query.code
  if (isOAuthCallback) return

  try {
    const ssoResult = await authStore.autoLoginSSO()
    if (ssoResult) {
      if (route.path.startsWith('/auth')) {
        const redirectPath = route.query.redirect || '/home'
        await router.push(redirectPath)
      } else if (route.path === '/') {
        await router.push('/home')
      }
      await Promise.all([loadPalette(), loadAdminSettings()])
    }
  } catch (error) {
    // Échec SSO silencieux (mode classique)
  }
})

</script>

<style>
/* Barre mobile (< lg) : bouton menu dans le flux, collant en haut */
.mobile-bar {
  position: sticky;
  inset-block-start: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  padding-block: 8px;
  padding-inline: 12px;
  background: var(--chrome);
  border-block-end: 1px solid var(--line);
}

.mobile-bar__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  inline-size: 40px;
  block-size: 40px;
  border-radius: 8px;
  color: var(--text);
  transition: background-color 150ms ease;
}

.mobile-bar__btn:hover {
  background: color-mix(in srgb, var(--text) 8%, transparent);
}

.mobile-bar__btn:focus-visible {
  outline: 2px solid rgb(var(--primary-500));
  outline-offset: 2px;
}

/* Zoom wrapper */
.zoom-wrapper {
  /* Avec la propriété zoom CSS, pas besoin de compensation de taille */
  width: 100%;
  min-height: 100vh;
}

/* Transition de page : l'ancienne page disparaît tout de suite,
   la nouvelle apparaît en fondu court (plus de mode out-in qui attend la sortie). */
.page-enter-active {
  transition: opacity 120ms ease-out;
}

.page-enter-from {
  opacity: 0;
}

/* !important : la racine de certaines pages fixe son propre display (styles scopés) */
.page-leave-active {
  display: none !important;
}

@media (prefers-reduced-motion: reduce) {
  .page-enter-active {
    transition: none;
  }
}

/* Loading overlay */
.loading-overlay {
  backdrop-filter: blur(2px);
}
</style>