<template>
  <div class="content-area">
    <!-- En-tête -->
    <PageHeader>
      <template #title>{{ isGlobal ? $t('events.admin.title') : $t('groupAdmin.eventsManagement.title') }}</template>
      <template #description>{{ isGlobal ? $t('events.admin.subtitle') : $t('groupAdmin.eventsManagement.subtitle') }}</template>
      <template #actions>
        <!-- Import des jours fériés : routes /admin/events/holidays/* réservées à l'admin global -->
        <BaseButton v-if="isAdmin" variant="secondary" icon="mdi:calendar-star" @click="showImportHolidaysModal = true">
          {{ $t('events.holidays.importButton') }}
        </BaseButton>
        <BaseButton tag="router-link" to="/admin/events/new" variant="primary" icon="mdi:plus">
          {{ $t('events.admin.newEvent') }}
        </BaseButton>
      </template>
    </PageHeader>

    <!-- Rappel du périmètre pour l'admin de groupe -->
    <div v-if="!isGlobal" class="ev-callout" role="note">
      <Icon icon="mdi:information-outline" class="ev-callout__icon" aria-hidden="true" />
      <div>
        <p class="ev-callout__title">{{ $t('groupAdmin.eventsManagement.infoTitle') }}</p>
        <p class="ev-callout__text">{{ $t('groupAdmin.eventsManagement.infoDescription') }}</p>
      </div>
    </div>

    <!-- Aucun événement et aucun filtre : invitation à en créer un -->
    <EmptyState
      v-if="!isLoading && eventsList.length === 0 && !hasActiveFilters"
      icon="mdi:calendar"
      :title="$t('events.admin.noEvents')"
      :description="emptyDescription"
    >
      <BaseButton tag="router-link" to="/admin/events/new" variant="primary" size="sm" icon="mdi:plus">
        {{ $t('events.admin.createEvent') }}
      </BaseButton>
    </EmptyState>

    <template v-else>
      <!-- Recherche, filtres et pagination sont gérés côté serveur : seule la page courante est passée au tableau -->
      <DataTable
        :columns="columns"
        :rows="rows"
        :loading="isLoading"
        :searchable="false"
      :footer="false"
        :page-size="100"
        :empty-title="$t('events.admin.noEvents')"
      >
        <template #toolbar>
          <label class="ev-search">
            <Icon icon="mdi:magnify" class="ev-search__icon" aria-hidden="true" />
            <input
              v-model="filters.search"
              type="search"
              class="ev-search__input"
              :placeholder="$t('events.admin.searchPlaceholder')"
              :aria-label="$t('events.admin.searchPlaceholder')"
            />
          </label>
          <label>
            <span class="sr-only">{{ $t('ui.table.columns.category') }}</span>
            <select v-model="filters.category_id" class="ui-select">
              <option value="">{{ $t('events.admin.allCategories') }}</option>
              <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
            </select>
          </label>
          <label>
            <span class="sr-only">{{ $t('ui.table.columns.status') }}</span>
            <select v-model="filters.status" class="ui-select">
              <option value="">{{ $t('events.admin.allStatus') }}</option>
              <option value="published">{{ $t('events.admin.published') }}</option>
              <option value="draft">{{ $t('events.admin.draft') }}</option>
            </select>
          </label>
          <label>
            <span class="sr-only">{{ $t('ui.table.columns.type') }}</span>
            <select v-model="filters.event_type" class="ui-select">
              <option value="">{{ $t('events.admin.allTypes') }}</option>
              <option value="recurring">{{ $t('events.admin.recurring') }}</option>
              <option value="all_day">{{ $t('events.admin.allDay') }}</option>
              <option value="upcoming">{{ $t('events.admin.upcoming') }}</option>
              <option value="past">{{ $t('events.admin.past') }}</option>
            </select>
          </label>
        </template>

        <template #cell-title="{ row }">
          <span class="ev-title">
            <span class="ev-title__line">
              <Icon v-if="row.raw.is_recurring" icon="mdi:repeat" class="ev-title__flag" :title="$t('groupAdmin.eventsManagement.indicators.recurring')" />
              <Icon v-if="row.raw.is_all_day" icon="mdi:weather-sunny" class="ev-title__flag" :title="$t('groupAdmin.eventsManagement.indicators.allDay')" />
              <router-link
                v-if="row.raw.slug"
                :to="{ name: 'EventDetail', params: { slug: row.raw.slug } }"
                class="ev-title__link"
              >{{ row.title }}</router-link>
              <span v-else class="ev-title__text">{{ row.title }}</span>
              <StatusBadge v-if="row.priorityLabel" :tone="row.raw.priority === 'urgent' ? 'bad' : 'warn'">{{ row.priorityLabel }}</StatusBadge>
            </span>
            <span v-if="row.raw.location" class="ev-title__sub">{{ row.raw.location }}</span>
          </span>
        </template>

        <template #cell-date="{ row }">
          <span class="ev-mono">{{ row.dateLabel }}</span>
        </template>

        <template #cell-status="{ row }">
          <StatusBadge :tone="row.raw.is_published ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
        </template>

        <template #cell-actions="{ row }">
          <RowActions :label="row.title" :editable="false" @delete="confirmDelete(row.raw)">
            <router-link
              :to="`/admin/events/${row.raw.slug}/edit`"
              class="row-actions__btn"
              :title="$t('events.admin.edit')"
              :aria-label="`${$t('events.admin.edit')} : ${row.title}`"
            >
              <Icon icon="mdi:pencil-outline" aria-hidden="true" />
            </router-link>
            <button
              type="button"
              class="row-actions__btn"
              :title="row.raw.is_published ? $t('events.admin.unpublish') : $t('events.admin.publish')"
              :aria-label="`${row.raw.is_published ? $t('events.admin.unpublish') : $t('events.admin.publish')} : ${row.title}`"
              @click="togglePublish(row.raw)"
            >
              <Icon :icon="row.raw.is_published ? 'mdi:eye-off-outline' : 'mdi:eye-outline'" aria-hidden="true" />
            </button>
          </RowActions>
        </template>
      </DataTable>

      <!-- Pagination serveur -->
      <nav v-if="pagination.pages > 1" class="ev-pager" :aria-label="$t('events.admin.pageInfo', { page: pagination.page, pages: pagination.pages })">
        <BaseButton variant="secondary" size="sm" icon="mdi:chevron-left" :disabled="pagination.page === 1" :aria-label="$t('ui.table.prev')" @click="changePage(pagination.page - 1)" />
        <span class="ev-mono">{{ $t('events.admin.pageInfo', { page: pagination.page, pages: pagination.pages }) }}</span>
        <BaseButton variant="secondary" size="sm" icon="mdi:chevron-right" :disabled="pagination.page === pagination.pages" :aria-label="$t('ui.table.next')" @click="changePage(pagination.page + 1)" />
      </nav>
    </template>

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
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { dateLocaleFor } from '@/utils/homeFormat'
import ImportHolidaysModal from '@/components/admin/ImportHolidaysModal.vue'
import { getEventListFetcher, getEventWriteService } from '@/components/admin/events/eventServices'
import { eventsService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useConfirm } from '@/composables/useConfirm'
import { useManagementScope } from '@/composables/useManagementScope'
import { CONTENT } from '@/config/access'

// Taille de page : /admin/events (paramètre `limit`) et listes filtrées par rôle (`page_size`)
const ADMIN_PAGE_SIZE = 10
const SCOPED_PAGE_SIZE = 20

const { t, locale } = useI18n()
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

const columns = computed(() => [
  { key: 'title', label: t('ui.table.columns.title'), sortable: true },
  { key: 'categoryName', label: t('ui.table.columns.category'), sortable: true, hideSm: true },
  { key: 'date', label: t('ui.table.columns.date'), sortable: true, hideSm: true },
  { key: 'author', label: t('ui.table.columns.author'), sortable: true, hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '120px' },
])

const DATE_OPTIONS = { year: 'numeric', month: 'short', day: 'numeric' }

const formatDateRange = (startDate, endDate) => {
  if (!startDate) return '—'
  const lang = dateLocaleFor(locale.value)
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  if (end && start.toDateString() !== end.toDateString()) {
    return `${start.toLocaleDateString(lang, DATE_OPTIONS)} – ${end.toLocaleDateString(lang, DATE_OPTIONS)}`
  }
  return start.toLocaleDateString(lang, DATE_OPTIONS)
}

const authorName = (author) => {
  const first = author?.first_name || author?.username || t('groupAdmin.eventsManagement.author.unknown')
  return `${first} ${author?.last_name || ''}`.trim()
}

const rows = computed(() =>
  eventsList.value.map(event => ({
    id: event.id,
    title: event.title,
    categoryName: event.category?.name || '',
    date: event.start_date || '',
    dateLabel: formatDateRange(event.start_date, event.end_date),
    author: authorName(event.author),
    status: event.is_published ? t('events.admin.published') : t('events.admin.draft'),
    priorityLabel: event.priority && event.priority !== 'normal' ? t(`events.priority.${event.priority}`) : '',
    raw: event,
  })),
)

onMounted(() => {
  loadCategories()
  loadEvents()
})
</script>

<style scoped>
.ev-callout {
  display: flex;
  gap: 10px;
  margin-block-end: 16px;
  padding: 12px 14px;
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--acc);
  border-radius: 4px;
  background: var(--surface);
}
.ev-callout__icon { flex: none; width: 18px; height: 18px; margin-block-start: 1px; color: var(--acc-ink); }
.ev-callout__title { margin: 0; font-size: 13.5px; font-weight: 600; color: var(--text); }
.ev-callout__text { margin: 2px 0 0; font-size: 12.5px; color: var(--muted); }

.ev-search { position: relative; display: flex; align-items: center; flex: 1 1 220px; max-width: 340px; }
.ev-search__icon { position: absolute; inset-inline-start: 9px; width: 16px; height: 16px; color: var(--muted); pointer-events: none; }
.ev-search__input {
  width: 100%;
  height: 32px;
  padding-inline: 30px 10px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  font: inherit;
  font-size: 13px;
  color: var(--text);
}
.ev-search__input:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.ev-title { display: flex; flex-direction: column; min-width: 0; }
.ev-title__line { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }
.ev-title__flag { flex: none; width: 15px; height: 15px; color: var(--muted); }
.ev-title__link, .ev-title__text { font-weight: 600; color: var(--text); }
.ev-title__link:hover { color: var(--acc-ink); text-decoration: underline; }
.ev-title__sub { font-size: 12px; color: var(--muted); }

.ev-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); white-space: nowrap; }

.ev-pager { display: flex; align-items: center; justify-content: center; gap: 12px; margin-block-start: 12px; }
</style>
