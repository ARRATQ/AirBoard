<template>
  <aside :class="['sidebar', { 'sidebar-collapsed': !appStore.sidebarOpen }]">
    <!-- Header -->
    <div class="sidebar-header">
      <!-- Sidebar expanded content -->
      <div v-if="appStore.sidebarOpen" class="flex-1">
        <div class="flex items-center gap-3">
          <div class="h-8 w-8 bg-white dark:bg-gray-800 rounded-lg flex items-center justify-center flex-shrink-0">
            <Icon :icon="appStore.appSettings?.app_icon || 'mdi:view-dashboard'" class="h-5 w-5 text-gray-900 dark:text-white" />
          </div>
          <div class="flex-1 min-w-0">
            <h1 class="sidebar-brand truncate">{{ appStore.appSettings?.app_name || 'Airboard' }}</h1>
            <!-- Version info with update badge -->
            <div class="flex items-center gap-2 mt-0.5">
              <span class="text-xs text-gray-400">
                v{{ versionStore.versionInfo.version }}
              </span>
              <button
                v-if="versionStore.shouldShowUpdateBadge()"
                @click="showUpdateModal = true"
                class="relative inline-flex items-center gap-1 px-2 py-0.5 text-xs font-medium rounded-full bg-red-500 text-white hover:bg-red-600 transition-colors animate-pulse"
                :title="$t('version.newVersionAvailable')"
              >
                <Icon icon="mdi:update" class="h-3 w-3" />
                <span>{{ $t('version.update') }}</span>
              </button>
            </div>
          </div>
          <!-- Notification Bell in Header -->
          <div class="flex-shrink-0">
            <NotificationBell />
          </div>
        </div>
      </div>

      <!-- Sidebar collapsed content - just icon -->
      <div
        v-else
        class="flex items-center justify-center w-full"
      >
        <div class="h-8 w-8 bg-white dark:bg-gray-800 rounded-lg flex items-center justify-center">
          <Icon :icon="appStore.appSettings?.app_icon || 'mdi:view-dashboard'" class="h-5 w-5 text-gray-900 dark:text-white" />
        </div>
      </div>
    </div>

    <!-- Navigation -->
    <nav class="sidebar-nav" :aria-label="$t('nav.mainLabel')">
      <!-- Espace utilisateur : identique pour tous les rôles -->
      <div class="sidebar-section">
        <SidebarLink
          v-for="item in USER_NAV"
          :key="item.to"
          :to="item.to"
          :icon="item.icon"
          :label="$t(item.labelKey)"
        />
      </div>

      <!-- Administration : construite d'après le rôle (config/navigation.js) -->
      <div v-if="adminNav" class="sidebar-section">
        <div class="sidebar-section-title">
          {{ $t(adminNav.titleKey) }}
        </div>

        <SidebarLink
          v-if="adminNav.overview"
          :to="adminNav.overview.to"
          :icon="adminNav.overview.icon"
          :label="$t(adminNav.overview.labelKey)"
          exact
        />

        <!-- Admin global : un groupe repliable par domaine -->
        <template v-if="adminNav.grouped">
          <CollapsibleMenu
            v-for="section in adminNav.sections"
            :key="section.id"
            :title="$t(section.labelKey)"
            :icon="section.icon"
            :storage-key="`sidebar-admin-${section.id}`"
            :compact="!appStore.sidebarOpen"
            :default-open="isSectionActive(section)"
          >
            <SidebarLink
              v-for="item in section.items"
              :key="item.to"
              :to="item.to"
              :icon="item.icon"
              :label="$t(item.labelKey)"
            />
          </CollapsibleMenu>
        </template>

        <!-- Admin de groupe, éditeur : la courte liste des pages permises -->
        <template v-else>
          <SidebarLink
            v-for="item in adminNav.items"
            :key="item.to"
            :to="item.to"
            :icon="item.icon"
            :label="$t(item.labelKey)"
          />
        </template>
      </div>
    </nav>

    <!-- Footer -->
    <div class="mt-auto border-t border-gray-200 dark:border-gray-700 p-4">
      <!-- Sidebar expanded content -->
      <div v-if="appStore.sidebarOpen">
        <!-- Zoom & Sidebar Toggle Row -->
        <div class="flex items-center justify-between mb-3">
          <ZoomControl v-model="appStore.zoomLevel" @update:modelValue="appStore.setZoomLevel" />
          <button
            @click="appStore.toggleSidebar()"
            class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            :title="$t('common.toggleSidebar')"
          >
            <Icon icon="mdi:menu-open" class="h-5 w-5" />
          </button>
        </div>

        <!-- Theme, Help & Language Row -->
        <div class="flex items-center justify-between mb-4">
          <!-- Theme toggle -->
          <button
            @click="appStore.toggleDarkMode()"
            class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            :title="$t('common.darkMode')"
          >
            <Icon
              :icon="appStore.isDarkMode ? 'mdi:weather-night' : 'mdi:weather-sunny'"
              class="h-5 w-5"
            />
          </button>

          <!-- Help button -->
          <button
            @click="showHelpDrawer = true"
            class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
            :title="$t('help.title')"
          >
            <Icon icon="mdi:help-circle-outline" class="h-5 w-5" />
          </button>

          <!-- Language selector - compact -->
          <select
            class="bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm text-gray-900 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500"
            :value="appStore.locale"
            @change="onChangeLocale($event.target.value)"
            :title="$t('common.language')"
          >
            <option value="ar">Ar</option>
            <option value="en">En</option>
            <option value="es">Es</option>
            <option value="fr">Fr</option>
          </select>
        </div>
      </div>

      <!-- Sidebar collapsed content - compact controls -->
      <div v-else class="flex flex-col items-center gap-3 mb-4">
        <!-- Sidebar expand button -->
        <button
          @click="appStore.toggleSidebar()"
          class="hidden lg:block p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
          :title="$t('common.expandSidebar')"
        >
          <Icon icon="mdi:menu" class="h-5 w-5" />
        </button>

        <!-- Theme toggle -->
        <button
          @click="appStore.toggleDarkMode()"
          class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
          :title="$t('common.darkMode')"
        >
          <Icon
            :icon="appStore.isDarkMode ? 'mdi:weather-night' : 'mdi:weather-sunny'"
            class="h-5 w-5"
          />
        </button>

        <!-- Help button -->
        <button
          @click="showHelpDrawer = true"
          class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white"
          :title="$t('help.title')"
        >
          <Icon icon="mdi:help-circle-outline" class="h-5 w-5" />
        </button>

        <!-- Language selector - icon only -->
        <select
          class="bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg px-2 py-1.5 text-xs text-gray-900 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 w-12 text-center"
          :value="appStore.locale"
          @change="onChangeLocale($event.target.value)"
          :title="$t('common.language')"
        >
          <option value="ar">Ar</option>
          <option value="en">En</option>
          <option value="es">Es</option>
          <option value="fr">Fr</option>
        </select>
      </div>

      <!-- User profile with dropdown -->
      <div class="relative" ref="userMenuRef">
        <button
          @click="toggleUserMenu"
          class="w-full flex items-center space-x-3 p-3 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors cursor-pointer"
        >
          <div class="flex-shrink-0">
            <div class="h-8 w-8 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center overflow-hidden">
              <img
                v-if="authStore.user?.avatar_url"
                :src="authStore.user.avatar_url"
                alt="Avatar"
                class="h-full w-full object-cover"
              />
              <span v-else class="text-sm font-medium text-gray-900 dark:text-white">
                {{ authStore.userInitials }}
              </span>
              <!-- Prominent Level Badge -->
              <div 
                class="absolute -bottom-1 -end-1 flex items-center justify-center w-5 h-5 rounded-full bg-gradient-to-br from-indigo-500 via-purple-500 to-pink-500 border-2 border-white dark:border-gray-800 text-[10px] font-black text-white shadow-lg animate-pulse-slow" 
                :title="`Niveau ${authStore.gamificationProfile?.level || 1}`"
              >
                {{ authStore.gamificationProfile?.level || 1 }}
              </div>
            </div>
          </div>
          <div v-if="appStore.sidebarOpen" class="flex-1 min-w-0 text-start">
            <p class="text-sm font-medium text-gray-900 dark:text-white truncate">
              {{ authStore.userDisplayName }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400 truncate">
              {{ authStore.user?.email }}
            </p>
          </div>
          <Icon
            v-if="appStore.sidebarOpen"
            :icon="showUserMenu ? 'mdi:chevron-up' : 'mdi:chevron-down'"
            class="h-4 w-4 text-gray-400 flex-shrink-0"
          />
        </button>

        <!-- Dropdown menu -->
        <Transition
          enter-active-class="transition ease-out duration-100"
          enter-from-class="transform opacity-0 scale-95"
          enter-to-class="transform opacity-100 scale-100"
          leave-active-class="transition ease-in duration-75"
          leave-from-class="transform opacity-100 scale-100"
          leave-to-class="transform opacity-0 scale-95"
        >
          <div
            v-if="showUserMenu"
            class="absolute bottom-full start-0 end-0 mb-2 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 overflow-hidden z-50"
          >
            <router-link
              to="/profile"
              @click="showUserMenu = false"
              class="flex items-center gap-3 px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              <Icon icon="mdi:account-circle-outline" class="h-5 w-5 text-gray-500 dark:text-gray-400" />
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ $t('profile.title') }}</span>
            </router-link>
            <div class="border-t border-gray-200 dark:border-gray-700"></div>
            <button
              @click="handleLogout"
              class="w-full flex items-center gap-3 px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-red-600 dark:text-red-400"
            >
              <Icon icon="mdi:logout" class="h-5 w-5" />
              <span class="text-sm">{{ $t('common.logout') }}</span>
            </button>
          </div>
        </Transition>
      </div>
    </div>

    <!-- Update Modal -->
    <Teleport to="body">
      <div
        v-if="showUpdateModal"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black bg-opacity-50"
        @click.self="showUpdateModal = false"
      >
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
          <!-- Header -->
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-primary-500 to-primary-600">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="h-10 w-10 bg-white rounded-lg flex items-center justify-center">
                  <Icon icon="mdi:update" class="h-6 w-6 text-primary-600" />
                </div>
                <div>
                  <h3 class="text-lg font-semibold text-white">{{ $t('version.newVersionAvailable') }}</h3>
                  <p class="text-sm text-primary-100">{{ $t('version.updateReady') }}</p>
                </div>
              </div>
              <button
                @click="showUpdateModal = false"
                class="text-white hover:text-gray-200 transition-colors"
              >
                <Icon icon="mdi:close" class="h-6 w-6" />
              </button>
            </div>
          </div>

          <!-- Content -->
          <div class="px-6 py-4 overflow-y-auto max-h-[60vh]">
            <div v-if="versionStore.updateDetails" class="space-y-4">
              <!-- Version comparison -->
              <div class="flex items-center justify-center gap-4 py-4">
                <div class="text-center">
                  <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">{{ $t('version.currentVersion') }}</p>
                  <div class="px-4 py-2 bg-gray-100 dark:bg-gray-700 rounded-lg">
                    <p class="text-xl font-bold text-gray-900 dark:text-white">
                      {{ versionStore.updateDetails.currentVersion }}
                    </p>
                  </div>
                </div>
                <Icon icon="mdi:arrow-right" class="rtl-flip h-8 w-8 text-gray-400" />
                <div class="text-center">
                  <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">{{ $t('version.newVersion') }}</p>
                  <div class="px-4 py-2 bg-green-100 dark:bg-green-900 rounded-lg">
                    <p class="text-xl font-bold text-green-600 dark:text-green-400">
                      {{ versionStore.updateDetails.latestVersion }}
                    </p>
                  </div>
                </div>
              </div>

              <!-- Release date -->
              <div v-if="versionStore.updateDetails.releaseDate" class="text-center text-sm text-gray-500 dark:text-gray-400">
                {{ $t('version.releasedOn') }} {{ formatDate(versionStore.updateDetails.releaseDate) }}
              </div>

              <!-- Release notes -->
              <div v-if="versionStore.updateDetails.releaseNotes" class="mt-6">
                <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{{ $t('version.releaseNotes') }}</h4>
                <div class="bg-gray-50 dark:bg-gray-900 rounded-lg p-4 prose prose-sm dark:prose-invert max-w-none">
                  <pre class="whitespace-pre-wrap text-sm text-gray-700 dark:text-gray-300">{{ versionStore.updateDetails.releaseNotes }}</pre>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900 flex items-center justify-end gap-3">
            <button
              @click="showUpdateModal = false"
              class="btn-secondary"
            >
              {{ $t('common.close') }}
            </button>
            <a
              v-if="versionStore.updateDetails?.releaseURL"
              :href="versionStore.updateDetails.releaseURL"
              target="_blank"
              class="btn-primary inline-flex items-center gap-2"
            >
              <Icon icon="mdi:github" class="h-5 w-5" />
              {{ $t('version.viewOnGithub') }}
            </a>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Help Drawer -->
    <HelpDrawer v-if="helpDrawerLoaded" v-model="showHelpDrawer" />
  </aside>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, defineAsyncComponent } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useVersionStore } from '@/stores/version'
import ZoomControl from '@/components/dashboard/ZoomControl.vue'
import NotificationBell from '@/components/notifications/NotificationBell.vue'
// Chargé à la première ouverture (embarque marked et les guides)
const HelpDrawer = defineAsyncComponent(() => import('@/components/help/HelpDrawer.vue'))
import CollapsibleMenu from '@/components/layout/CollapsibleMenu.vue'
import SidebarLink from '@/components/layout/SidebarLink.vue'
import { USER_NAV, buildAdminNav } from '@/config/navigation'
import { getRoles } from '@/config/access'
import { isPathActive } from '@/utils/navigation'

const authStore = useAuthStore()
const appStore = useAppStore()
const versionStore = useVersionStore()
const route = useRoute()
const router = useRouter()

// State for update modal
const showUpdateModal = ref(false)
const showHelpDrawer = ref(false)
const helpDrawerLoaded = ref(false)
watch(showHelpDrawer, (open) => {
  if (open) helpDrawerLoaded.value = true
})

// State for user menu
const showUserMenu = ref(false)
const userMenuRef = ref(null)

// Menu d'administration du rôle courant ; un domaine s'ouvre s'il contient la page affichée
const adminNav = computed(() => buildAdminNav(getRoles(authStore.user)))
const isSectionActive = (section) => section.items.some((item) => isPathActive(route.path, item.to))

const toggleUserMenu = () => {
  showUserMenu.value = !showUserMenu.value
}

// Close menu when clicking outside
const handleClickOutside = (event) => {
  if (userMenuRef.value && !userMenuRef.value.contains(event.target)) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  versionStore.stopPeriodicUpdateCheck()
})

// Fermer la sidebar sur mobile lors de la navigation
watch(() => route.path, () => {
  // Vérifier si on est sur mobile (< 1024px)
  if (window.innerWidth < 1024 && appStore.sidebarOpen) {
    appStore.setSidebarOpen(false)
  }
})

// Load app settings and version when component mounts
onMounted(async () => {
  if (authStore.isAuthenticated && authStore.isAdmin) {
    try {
      await appStore.ensureAppSettings()
    } catch (error) {
      console.error('Failed to load app settings in sidebar:', error)
    }
  }

  // Load version info
  try {
    await versionStore.fetchVersion()
    // Initialize from cache
    versionStore.initializeFromCache()
    // Start periodic update checks
    versionStore.startPeriodicUpdateCheck()
  } catch (error) {
    console.error('Failed to load version info:', error)
  }

  if (authStore.isAuthenticated) {
    authStore.fetchGamificationProfile()
  }
})

// Watch for settings changes
watch(() => appStore.settingsLastUpdated, async () => {
  if (authStore.isAuthenticated && authStore.isAdmin && !appStore.appSettings) {
    try {
      await appStore.refreshAppSettings()
    } catch (error) {
      console.error('Failed to reload app settings:', error)
    }
  }
})

const handleLogout = async () => {
  try {
    authStore.logout()
    appStore.showInfo('')
    router.push('/auth/login')
  } catch (error) {
    console.error('Logout error:', error)
    appStore.showError('')
  }
}

const onChangeLocale = (loc) => {
  appStore.setAppLocale(loc)
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

// Échap ferme la modale
useEscapeClose(showUpdateModal, () => { showUpdateModal.value = false })
</script>
