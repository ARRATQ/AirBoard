<template>
  <div class="content-area">
    <!-- En-tête -->
    <PageHeader>
      <template #title>{{ isGlobal ? $t('events.admin.title') : $t('groupAdmin.eventsManagement.title') }}</template>
      <template #description>{{ isGlobal ? $t('events.admin.subtitle') : $t('groupAdmin.eventsManagement.subtitle') }}</template>
      <template #actions>
        <div class="flex items-center gap-3">
          <!-- Import des jours fériés : routes /admin/events/holidays/* réservées à l'admin global -->
          <button v-if="isAdmin" class="btn btn-secondary" @click="showImportHolidaysModal = true">
            <Icon icon="mdi:calendar-star" class="h-4 w-4 mr-2" />
            {{ $t('events.holidays.importButton') }}
          </button>
          <router-link to="/admin/events/new" class="btn btn-primary">
            <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
            {{ $t('events.admin.newEvent') }}
          </router-link>
        </div>
      </template>
    </PageHeader>

    <!-- Rappel du périmètre pour l'admin de groupe -->
    <div
      v-if="!isGlobal"
      class="mb-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-300 dark:border-blue-600 rounded-lg"
    >
      <div class="flex items-start">
        <Icon icon="mdi:information" class="h-5 w-5 text-blue-600 dark:text-blue-500 mr-3 mt-0.5" />
        <div>
          <h3 class="text-blue-800 dark:text-blue-500 font-medium">{{ $t('groupAdmin.eventsManagement.infoTitle') }}</h3>
          <p class="text-sm text-blue-700 dark:text-blue-200 mt-1">
            {{ $t('groupAdmin.eventsManagement.infoDescription') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Filtres -->
    <div class="card mb-6">
      <div class="flex flex-col sm:flex-row gap-4">
        <div class="flex-1 min-w-[200px]">
          <input
            v-model="filters.search"
            type="text"
            :placeholder="$t('events.admin.searchPlaceholder')"
            class="input"
          />
        </div>

        <select v-model="filters.category_id" class="input w-full sm:w-48">
          <option value="">{{ $t('events.admin.allCategories') }}</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>

        <select v-model="filters.status" class="input w-full sm:w-40">
          <option value="">{{ $t('events.admin.allStatus') }}</option>
          <option value="published">{{ $t('events.admin.published') }}</option>
          <option value="draft">{{ $t('events.admin.draft') }}</option>
        </select>

        <select v-model="filters.event_type" class="input w-full sm:w-40">
          <option value="">{{ $t('events.admin.allTypes') }}</option>
          <option value="recurring">{{ $t('events.admin.recurring') }}</option>
          <option value="all_day">{{ $t('events.admin.allDay') }}</option>
          <option value="upcoming">{{ $t('events.admin.upcoming') }}</option>
          <option value="past">{{ $t('events.admin.past') }}</option>
        </select>
      </div>
    </div>

    <!-- Chargement -->
    <div v-if="isLoading" class="flex justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-gray-400" />
    </div>

    <!-- Liste -->
    <div v-else-if="eventsList.length > 0" class="space-y-4">
      <EventListItem
        v-for="event in eventsList"
        :key="event.id"
        :event="event"
        @toggle-publish="togglePublish"
        @delete="confirmDelete"
      />

      <!-- Pagination -->
      <div v-if="pagination.pages > 1" class="flex justify-center gap-2 mt-6">
        <button
          :disabled="pagination.page === 1"
          class="btn btn-secondary btn-sm"
          @click="changePage(pagination.page - 1)"
        >
          <Icon icon="mdi:chevron-left" class="h-4 w-4" />
        </button>

        <span class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300">
          {{ $t('events.admin.pageInfo', { page: pagination.page, pages: pagination.pages }) }}
        </span>

        <button
          :disabled="pagination.page === pagination.pages"
          class="btn btn-secondary btn-sm"
          @click="changePage(pagination.page + 1)"
        >
          <Icon icon="mdi:chevron-right" class="h-4 w-4" />
        </button>
      </div>
    </div>

    <!-- État vide -->
    <div v-else class="empty-state">
      <Icon icon="mdi:calendar" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('events.admin.noEvents') }}</h3>
      <p class="empty-state-description">{{ emptyDescription }}</p>
      <router-link v-if="!hasActiveFilters" to="/admin/events/new" class="btn btn-primary mt-4">
        <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
        {{ $t('events.admin.createEvent') }}
      </router-link>
    </div>

    <!-- Import des jours fériés (admin global) -->
    <ImportHolidaysModal
      v-if="isAdmin && showImportHolidaysModal"
      :categories="categories"
      @close="showImportHolidaysModal = false"
      @imported="handleHolidaysImported"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import ImportHolidaysModal from '@/components/admin/ImportHolidaysModal.vue'
import EventListItem from '@/components/admin/events/EventListItem.vue'
import { getEventListFetcher, getEventWriteService } from '@/components/admin/events/eventServices'
import { eventsService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useConfirm } from '@/composables/useConfirm'
import { useManagementScope } from '@/composables/useManagementScope'
import { CONTENT } from '@/config/access'

// Taille de page : /admin/events (paramètre `limit`) et listes filtrées par rôle (`page_size`)
const ADMIN_PAGE_SIZE = 10
const SCOPED_PAGE_SIZE = 20

const { t } = useI18n()
const appStore = useAppStore()
const { confirm: askConfirm } = useConfirm()
// Admin et éditeur : tous les événements ; admin de groupe : ceux de ses groupes
const { isGlobal, isAdmin } = useManagementScope(CONTENT)

const eventsList = ref([])
const categories = ref([])
const isLoading = ref(false)
const showImportHolidaysModal = ref(false)
const page = ref(1)
const pagination = ref({ page: 1, pages: 1 })

const filters = ref({
  search: '',
  category_id: '',
  status: '',
  event_type: ''
})

const hasActiveFilters = computed(() => {
  const { search, category_id: categoryId, status, event_type: eventType } = filters.value
  return Boolean(search || categoryId || status || eventType)
})

const emptyDescription = computed(() => {
  if (hasActiveFilters.value) return t('groupAdmin.eventsManagement.empty.noResults')
  return isGlobal.value ? t('events.admin.noEventsHelp') : t('groupAdmin.eventsManagement.empty.createFirst')
})

// Paramètres de GET /admin/events (filtres status / event_type gérés côté serveur)
const buildAdminParams = () => {
  const params = { page: page.value, limit: ADMIN_PAGE_SIZE, ...filters.value }
  return Object.fromEntries(Object.entries(params).filter(([, value]) => Boolean(value)))
}

// Paramètres de GET /events et /group-admin/events (même gestionnaire GetEvents)
const EVENT_TYPE_PARAMS = {
  recurring: { is_recurring: true },
  all_day: { is_all_day: true },
  upcoming: { upcoming: true },
  past: { past: true }
}
const STATUS_PARAMS = {
  published: { is_published: true },
  draft: { is_published: false }
}

const buildScopedParams = () => ({
  page: page.value,
  page_size: SCOPED_PAGE_SIZE,
  search: filters.value.search || undefined,
  category_id: filters.value.category_id || undefined,
  ...(STATUS_PARAMS[filters.value.status] || {}),
  ...(EVENT_TYPE_PARAMS[filters.value.event_type] || {})
})

const loadEvents = async () => {
  isLoading.value = true
  try {
    const fetchEvents = getEventListFetcher(isAdmin.value, isGlobal.value)
    const params = isAdmin.value ? buildAdminParams() : buildScopedParams()
    const response = await fetchEvents(params)
    eventsList.value = (response?.events || response?.data || []).filter((event) => event && event.id)
    pagination.value = {
      page: response?.page || 1,
      pages: response?.total_pages || 1
    }
  } catch (error) {
    console.error('Erreur au chargement des événements :', error)
    eventsList.value = []
    appStore.showError(t('events.admin.loadError'))
  } finally {
    isLoading.value = false
  }
}

const loadCategories = async () => {
  try {
    const data = await eventsService.getCategories()
    categories.value = Array.isArray(data) ? data : (data?.data || [])
  } catch (error) {
    console.error('Erreur au chargement des catégories :', error)
    categories.value = []
  }
}

const togglePublish = async (event) => {
  const nextPublished = !event.is_published
  try {
    const service = getEventWriteService(isAdmin.value, isGlobal.value)
    await service.updateEvent(event.slug, { ...event, is_published: nextPublished })
    eventsList.value = eventsList.value.map((item) =>
      item.id === event.id ? { ...item, is_published: nextPublished } : item
    )
    appStore.showSuccess(
      nextPublished ? t('events.admin.publishedSuccess') : t('events.admin.unpublishedSuccess')
    )
  } catch (error) {
    console.error('Erreur au changement de statut de publication :', error)
    appStore.showError(t('events.admin.updateError'))
  }
}

const confirmDelete = async (event) => {
  const confirmed = await askConfirm({
    title: t('events.admin.deleteConfirmTitle'),
    message: t('events.admin.deleteConfirmMessage', { title: event.title }),
    confirmLabel: t('common.delete'),
    tone: 'danger'
  })
  if (!confirmed) return

  try {
    const service = getEventWriteService(isAdmin.value, isGlobal.value)
    await service.deleteEvent(event.id)
    appStore.showSuccess(t('events.admin.deleteSuccess'))
    await loadEvents()
  } catch (error) {
    console.error('Erreur à la suppression de l\'événement :', error)
    appStore.showError(t('events.admin.deleteError'))
  }
}

const changePage = (target) => {
  if (target < 1 || target > pagination.value.pages) return
  page.value = target
  loadEvents()
}

const handleHolidaysImported = (result) => {
  showImportHolidaysModal.value = false
  appStore.showSuccess(t('events.holidays.importSuccess', { count: result.imported }))
  loadEvents()
}

// Tout changement de filtre repart de la première page
watch(
  () => [filters.value.search, filters.value.category_id, filters.value.status, filters.value.event_type],
  () => {
    page.value = 1
    loadEvents()
  }
)

onMounted(() => {
  loadCategories()
  loadEvents()
})
</script>
