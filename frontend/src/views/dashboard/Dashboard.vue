<template>
  <div class="content-area">
    <PageHeader :eyebrow="$t('home.rail.portal')">
      <template #title>{{ pageTitle }}</template>
      <template #description>{{ pageDescription }}</template>
      <template #actions>
        <SegmentedControl v-model="viewMode" :options="viewModes" :label="$t('dashboard.viewModes.label')" icon-only-small />
        <BaseButton
          v-if="dashboard?.app_groups?.length && viewMode !== 'table'"
          size="sm"
          :icon="allGroupsCollapsed ? 'lucide:chevrons-up-down' : 'lucide:chevrons-down-up'"
          @click="toggleAllGroups"
        >
          {{ allGroupsCollapsed ? $t('dashboard.expandAll') : $t('dashboard.collapseAll') }}
        </BaseButton>
      </template>
    </PageHeader>

    <AnnouncementPanel v-if="activeAnnouncements.length" class="portal__announcements" :announcements="activeAnnouncements" />

    <div v-if="isLoading" class="portal__state" role="status">
      <Icon icon="lucide:loader-circle" class="portal__spin" aria-hidden="true" />
      <span class="sr-only">{{ $t('common.loading') }}</span>
    </div>

    <EmptyState
      v-else-if="!dashboard?.app_groups?.length"
      icon="lucide:app-window"
      :title="$t('dashboard.emptyTitle')"
      :description="authStore.isAdmin ? $t('dashboard.emptyAdmin') : $t('dashboard.emptyUser')"
    >
      <BaseButton v-if="authStore.isAdmin" tag="router-link" to="/admin/resources/categories" variant="primary" icon="lucide:plus">
        {{ $t('dashboard.createCategory') }}
      </BaseButton>
    </EmptyState>

    <div v-else class="portal">
      <div class="portal__toolbar">
        <SegmentedControl v-model="typeFilter" :options="typeOptions" :label="$t('dashboard.typeFilter.label')" />
        <p class="portal__count" aria-live="polite">{{ $t('dashboard.count', visibleCount) }}</p>
      </div>

      <ResourceGroup
        v-if="favoriteApps.length && viewMode !== 'table'"
        group-key="favorites"
        :title="$t('common.myFavorites')"
        icon="lucide:star"
        color="#d99a00"
        :count="favoriteApps.length"
        :collapsed="collapsedGroups.has('favorites')"
        @toggle="toggleGroup('favorites')"
      >
        <ResourceTile
          v-for="app in favoriteApps"
          :key="app.id"
          :resource="app"
          favorite
          @toggle-favorite="toggleFavorite"
          @show-details="openAppDetails"
        />
      </ResourceGroup>

      <EmptyState v-if="!visibleCount" compact icon="lucide:funnel-x" :title="$t('dashboard.noneOfType')" />
      <ResourceTable v-else-if="viewMode === 'table'" :app-groups="visibleGroups" @show-details="openAppDetails" />
      <ResourceGroupsView
        v-else
        :app-groups="visibleGroups"
        :collapsed-groups="collapsedGroups"
        :layout="viewMode === 'grid' ? 'columns' : 'wide'"
        @toggle-group="toggleGroup"
        @show-details="openAppDetails"
      />
    </div>

    <ApplicationDetailModal v-if="showDetailsModal && selectedApp" :app="selectedApp" @close="showDetailsModal = false" />
  </div>
</template>

<script setup>
// Portail des ressources : liens vers des applis en ligne, documents et sites Office 365, liens web.
// Le type de chaque ressource est déduit de son URL (utils/resourceType.js) ; le filtre par type est
// gardé dans l'URL (?type=docs). Vues : grille, colonnes par catégorie, tableau.
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useFavoritesStore } from '@/stores/favorites'
import { dashboardService, announcementsService } from '@/services/api'
import { TYPE_FAMILIES, resourceTypeOf, typeFamilyOf } from '@/utils/resourceType'
import PageHeader from '@/components/ui/PageHeader.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import SegmentedControl from '@/components/ui/SegmentedControl.vue'
import AnnouncementPanel from '@/components/ui/AnnouncementPanel.vue'
import ResourceGroup from '@/components/dashboard/ResourceGroup.vue'
import ResourceTile from '@/components/dashboard/ResourceTile.vue'
import ResourceGroupsView from '@/components/dashboard/ResourceGroupsView.vue'
import ResourceTable from '@/components/dashboard/ResourceTable.vue'
import ApplicationDetailModal from '@/components/dashboard/ApplicationDetailModal.vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const favoritesStore = useFavoritesStore()

const dashboard = ref(null)
const appSettings = ref({})
const isLoading = ref(false)
const activeAnnouncements = ref([])

// Valeurs livrées par défaut côté serveur (handlers/settings.go) : tant qu'un admin ne les a pas
// personnalisées, la page garde le vocabulaire commun.
const DEFAULT_TITLE = 'Dashboard'
const DEFAULT_WELCOME = 'Welcome to your application portal'
const customized = (value, fallback) => (value && value !== fallback ? value : null)
const pageTitle = computed(() => customized(appSettings.value.dashboard_title, DEFAULT_TITLE) || t('common.applications'))
const pageDescription = computed(() => customized(appSettings.value.welcome_message, DEFAULT_WELCOME) || t('dashboard.subtitle'))

// ── Vues ─────────────────────────────────────────────────────────────
// Valeurs conservées pour le choix déjà mémorisé (localStorage) : list = une catégorie par ligne,
// grid = catégories en colonnes.
const VIEW_KEY = 'dashboard-view-mode'
const readView = () => {
  try {
    return localStorage.getItem(VIEW_KEY) || 'list'
  } catch {
    return 'list'
  }
}
const viewMode = ref(readView())
watch(viewMode, (mode) => {
  try {
    localStorage.setItem(VIEW_KEY, mode)
  } catch {
    /* préférence non mémorisée */
  }
})
const viewModes = computed(() => [
  { value: 'list', label: t('dashboard.viewModes.grid'), icon: 'lucide:layout-grid' },
  { value: 'grid', label: t('dashboard.viewModes.list'), icon: 'lucide:columns-3' },
  { value: 'table', label: t('dashboard.viewModes.table'), icon: 'lucide:table' },
])

// ── Filtre par type (dans l'URL) ─────────────────────────────────────
const typeFilter = computed({
  get: () => (TYPE_FAMILIES.includes(route.query.type) ? route.query.type : 'all'),
  set: (value) => router.replace({ query: { ...route.query, type: value === 'all' ? undefined : value } }),
})

const allApps = computed(() => (dashboard.value?.app_groups || []).flatMap((g) => g.applications || []))
const familyOf = (app) => typeFamilyOf(resourceTypeOf(app.url))
const matchesType = (app) => typeFilter.value === 'all' || familyOf(app) === typeFilter.value

const typeOptions = computed(() => [
  { value: 'all', label: t('dashboard.typeFilter.all'), count: allApps.value.length },
  ...TYPE_FAMILIES.map((family) => ({
    value: family,
    label: t(`dashboard.typeFilter.${family}`),
    count: allApps.value.filter((app) => familyOf(app) === family).length,
  })),
])

const visibleGroups = computed(() => {
  const groups = dashboard.value?.app_groups || []
  if (typeFilter.value === 'all') return groups
  return groups
    .map((g) => ({ ...g, applications: (g.applications || []).filter(matchesType) }))
    .filter((g) => g.applications.length)
})
const visibleCount = computed(() => visibleGroups.value.reduce((n, g) => n + (g.applications?.length || 0), 0))

const favoriteApps = computed(() => allApps.value.filter((app) => favoritesStore.isFavorite(app.id) && matchesType(app)))

// ── Catégories repliées (mémorisées) ─────────────────────────────────
const COLLAPSED_KEY = 'airboard_collapsed_groups'
const collapsedGroups = ref(new Set())

function loadCollapsedGroups() {
  try {
    const saved = localStorage.getItem(COLLAPSED_KEY)
    if (saved) collapsedGroups.value = new Set(JSON.parse(saved))
  } catch (error) {
    console.error('Error loading collapsed groups:', error)
  }
}

function saveCollapsedGroups(next) {
  collapsedGroups.value = next
  try {
    localStorage.setItem(COLLAPSED_KEY, JSON.stringify([...next]))
  } catch (error) {
    console.error('Error saving collapsed groups:', error)
  }
}

function toggleGroup(id) {
  const next = new Set(collapsedGroups.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  saveCollapsedGroups(next)
}

const allGroupsCollapsed = computed(() => {
  const groups = dashboard.value?.app_groups || []
  return groups.length > 0 && groups.every((g) => collapsedGroups.value.has(g.id))
})

function toggleAllGroups() {
  if (allGroupsCollapsed.value) saveCollapsedGroups(new Set())
  else saveCollapsedGroups(new Set([...(dashboard.value?.app_groups || []).map((g) => g.id), 'favorites']))
}

// ── Détails et favoris ───────────────────────────────────────────────
const showDetailsModal = ref(false)
const selectedApp = ref(null)

function openAppDetails(app) {
  selectedApp.value = app
  showDetailsModal.value = true
}

async function toggleFavorite(app) {
  try {
    await favoritesStore.toggleFavorite(app.id)
  } catch (error) {
    console.error('Error toggling favorite:', error)
    appStore.showError(t('ui.toast.error'))
  }
}

// ── Chargement ───────────────────────────────────────────────────────
const uniqueById = (items) => {
  if (!Array.isArray(items)) return []
  const seen = new Set()
  return items.filter((item) => {
    const id = item?.id ?? JSON.stringify(item)
    if (seen.has(id)) return false
    seen.add(id)
    return true
  })
}

async function loadDashboard() {
  isLoading.value = true
  try {
    const data = await dashboardService.getDashboard()
    // Déduplication côté client par sécurité
    if (data?.app_groups) {
      data.app_groups = uniqueById(data.app_groups).map((g) => ({ ...g, applications: uniqueById(g.applications || []) }))
    }
    dashboard.value = data
  } catch (error) {
    console.error('Error loading dashboard:', error)
    appStore.showError(t('dashboard.loadError'))
  } finally {
    isLoading.value = false
  }
}

async function loadAppSettings() {
  try {
    if (appStore.appSettings) {
      appSettings.value = appStore.appSettings
      return
    }
    // Ne pas appeler l'API si l'utilisateur n'est pas admin (évite 403)
    if (!authStore.isAdmin) {
      appSettings.value = {}
      return
    }
    appSettings.value = await appStore.ensureAppSettings()
  } catch (error) {
    console.error('Error loading app settings:', error)
    appSettings.value = {}
  }
}

async function loadAnnouncements() {
  try {
    activeAnnouncements.value = (await announcementsService.getActiveAnnouncements()) || []
  } catch (error) {
    console.error('Error loading announcements:', error)
  }
}

watch(() => appStore.settingsLastUpdated, loadAppSettings)

onMounted(async () => {
  loadCollapsedGroups()
  await Promise.all([loadDashboard(), loadAppSettings(), favoritesStore.loadFavorites(), loadAnnouncements()])
})
</script>

<style scoped>
.portal {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.portal__announcements { margin-block-end: 18px; }

.portal__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 16px;
}

.portal__count {
  margin: 0;
  font-family: var(--f-num);
  font-size: 12px;
  color: var(--muted);
}

.portal__state {
  display: flex;
  justify-content: center;
  padding-block: 48px;
  color: var(--muted);
}

.portal__spin {
  width: 30px;
  height: 30px;
  animation: portal-spin 1s linear infinite;
}

@keyframes portal-spin {
  to { transform: rotate(360deg); }
}

@media (prefers-reduced-motion: reduce) {
  .portal__spin { animation: none; }
}
</style>
