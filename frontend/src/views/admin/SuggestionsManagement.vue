<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ t('suggestions.admin.title') }}</template>
      <template #description>{{ t('suggestions.admin.subtitle') }}</template>
    </PageHeader>

    <DataTable
      :columns="columns"
      :rows="rows"
      :loading="loading"
      :searchable="false"
      :footer="false"
      :page-size="1000"
      :empty-title="t('suggestions.emptyTitle')"
    >
      <template #toolbar>
        <label class="sm-search">
          <Icon icon="mdi:magnify" class="sm-search__icon" aria-hidden="true" />
          <input
            v-model="filters.search"
            type="search"
            class="sm-search__input"
            :placeholder="t('suggestions.filters.search')"
            :aria-label="t('suggestions.filters.search')"
            @input="debounceLoad"
          />
        </label>
        <label>
          <span class="sr-only">{{ t('ui.table.columns.category') }}</span>
          <select v-model="filters.category_id" class="ui-select" @change="applyFilters">
            <option value="">{{ t('suggestions.filters.allCategories') }}</option>
            <option v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.name }}</option>
          </select>
        </label>
        <label>
          <span class="sr-only">{{ t('ui.table.columns.status') }}</span>
          <select v-model="filters.status" class="ui-select" @change="applyFilters">
            <option value="">{{ t('suggestions.filters.allStatuses') }}</option>
            <option v-for="status in statuses" :key="status" :value="status">{{ t(`suggestions.statuses.${status}`) }}</option>
          </select>
        </label>
        <label>
          <span class="sr-only">{{ t('suggestions.admin.archivedAll') }}</span>
          <select v-model="filters.archived" class="ui-select" @change="applyFilters">
            <option value="">{{ t('suggestions.admin.archivedAll') }}</option>
            <option value="0">{{ t('suggestions.admin.archivedNo') }}</option>
            <option value="1">{{ t('suggestions.admin.archivedYes') }}</option>
          </select>
        </label>
      </template>

      <template #cell-title="{ row }">
        <RouterLink :to="`/suggestions/${row.id}`" class="sm-title">{{ row.title }}</RouterLink>
      </template>

      <template #cell-createdAt="{ row }">
        <span class="sm-mono">{{ row.createdAt }}</span>
      </template>

      <template #cell-archived="{ row }">
        <StatusBadge :tone="row.raw.is_archived ? 'neutral' : 'ok'">
          {{ row.raw.is_archived ? t('suggestions.admin.archivedBadge') : t('suggestions.admin.activeBadge') }}
        </StatusBadge>
      </template>

      <template #cell-status="{ row }">
        <label>
          <span class="sr-only">{{ t('ui.table.columns.status') }} : {{ row.title }}</span>
          <select class="ui-select" :value="row.raw.status" @change="updateStatus(row.raw, $event.target.value)">
            <option v-for="status in statuses" :key="status" :value="status">{{ t(`suggestions.statuses.${status}`) }}</option>
          </select>
        </label>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.title" :editable="false" :deletable="false">
          <button
            type="button"
            class="row-actions__btn"
            :aria-label="`${row.raw.is_archived ? t('suggestions.admin.activate') : t('suggestions.admin.archive')} : ${row.title}`"
            :title="row.raw.is_archived ? t('suggestions.admin.activate') : t('suggestions.admin.archive')"
            @click="toggleArchive(row.raw)"
          >
            <Icon :icon="row.raw.is_archived ? 'mdi:archive-arrow-up-outline' : 'mdi:archive-arrow-down-outline'" aria-hidden="true" />
          </button>
        </RowActions>
      </template>
    </DataTable>

    <nav v-if="pagination.total_pages > 1" class="sm-pager" :aria-label="t('ui.table.range', { from: 1, to: pagination.page_size, total: pagination.total })">
      <button type="button" class="sm-pager__btn" :disabled="pagination.page === 1" :aria-label="t('ui.table.prev')" @click="changePage(pagination.page - 1)">
        <Icon icon="mdi:chevron-left" aria-hidden="true" />
      </button>
      <button
        v-for="page in visiblePages"
        :key="page"
        type="button"
        class="sm-pager__btn"
        :class="{ 'sm-pager__btn--on': page === pagination.page }"
        :aria-current="page === pagination.page ? 'page' : undefined"
        @click="changePage(page)"
      >
        {{ page }}
      </button>
      <button type="button" class="sm-pager__btn" :disabled="pagination.page === pagination.total_pages" :aria-label="t('ui.table.next')" @click="changePage(pagination.page + 1)">
        <Icon icon="mdi:chevron-right" aria-hidden="true" />
      </button>
    </nav>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { suggestionsService } from '@/services/api'
import { dateLocaleFor } from '@/utils/homeFormat'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import RowActions from '@/components/ui/RowActions.vue'

const { t, locale } = useI18n()

const loading = ref(false)
const suggestions = ref([])
const categories = ref([])
const statuses = ['new', 'reviewing', 'planned', 'in_progress', 'implemented', 'rejected']

const filters = ref({
  search: '',
  category_id: '',
  status: '',
  archived: ''
})

const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 1
})

let searchTimeout = null

const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, pagination.value.page - 2)
  const end = Math.min(pagination.value.total_pages, pagination.value.page + 2)
  for (let i = start; i <= end; i += 1) pages.push(i)
  return pages
})

const columns = computed(() => [
  { key: 'title', label: t('ui.table.columns.title') },
  { key: 'category', label: t('ui.table.columns.category'), hideSm: true },
  { key: 'author', label: t('ui.table.columns.author'), hideSm: true },
  { key: 'createdAt', label: t('ui.table.columns.date'), hideSm: true },
  { key: 'archived', label: '', hideSm: true },
  { key: 'status', label: t('ui.table.columns.status') },
  { key: 'actions', label: '', align: 'end', width: '60px' },
])

// Le filtrage et la pagination se font côté serveur ; le tableau n'affiche que la page reçue.
const rows = computed(() =>
  suggestions.value.map(item => ({
    id: item.id,
    title: item.title,
    category: item.category_ref?.name || item.category || t('suggestions.noCategory'),
    author: formatAuthor(item),
    createdAt: formatDate(item.created_at),
    raw: item,
  })),
)

const loadCategories = async () => {
  const response = await suggestionsService.getAdminCategories({ include_inactive: 1 })
  categories.value = response.categories || []
}

const loadSuggestions = async () => {
  try {
    loading.value = true
    const response = await suggestionsService.getAdminSuggestions({
      search: filters.value.search || undefined,
      category_id: filters.value.category_id || undefined,
      status: filters.value.status || undefined,
      archived: filters.value.archived || undefined,
      page: pagination.value.page,
      page_size: pagination.value.page_size
    })

    suggestions.value = response.suggestions || []
    pagination.value = {
      page: response.page || 1,
      page_size: response.page_size || 20,
      total: response.total || 0,
      total_pages: response.total_pages || 1
    }
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.loadError'))
    }
  } finally {
    loading.value = false
  }
}

const updateStatus = async (item, status) => {
  try {
    const updated = await suggestionsService.updateSuggestionStatus(item.id, status)
    item.status = updated.status
    if (window.$toast) {
      window.$toast.success(t('suggestions.statusUpdated'))
    }
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.statusUpdateError'))
    }
  }
}

const toggleArchive = async (item) => {
  try {
    const updated = await suggestionsService.updateSuggestionArchive(item.id, !item.is_archived)
    item.is_archived = updated.is_archived
    if (window.$toast) {
      window.$toast.success(item.is_archived ? t('suggestions.admin.archivedDone') : t('suggestions.admin.activatedDone'))
    }
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.statusUpdateError'))
    }
  }
}

const debounceLoad = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.value.page = 1
    loadSuggestions()
  }, 300)
}

const applyFilters = () => {
  pagination.value.page = 1
  loadSuggestions()
}

const changePage = (page) => {
  if (page < 1 || page > pagination.value.total_pages) return
  pagination.value.page = page
  loadSuggestions()
}

const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString(dateLocaleFor(locale.value), { day: '2-digit', month: 'short', year: 'numeric' })
}

const formatAuthor = (item) => {
  if (item.is_anonymous) return t('suggestions.anonymous')
  const first = item.user?.first_name || ''
  const last = item.user?.last_name || ''
  const fullName = `${first} ${last}`.trim()
  return fullName || item.user?.username || t('suggestions.anonymous')
}

onMounted(async () => {
  await loadCategories()
  await loadSuggestions()
})
</script>

<style scoped>
.sm-title { font-weight: 600; color: var(--text); }
.sm-title:hover { color: var(--acc-ink); text-decoration: underline; }

.sm-mono {
  font-family: var(--f-num);
  font-size: 11.5px;
  letter-spacing: 0.04em;
  color: var(--muted);
  white-space: nowrap;
}

.sm-search { position: relative; flex: 1 1 200px; min-width: 160px; }

.sm-search__icon {
  position: absolute;
  inset-inline-start: 9px;
  inset-block-start: 50%;
  width: 16px;
  height: 16px;
  transform: translateY(-50%);
  color: var(--muted);
  pointer-events: none;
}

.sm-search__input {
  width: 100%;
  padding-block: 7px;
  padding-inline: 30px 10px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  font-size: 13px;
}

.sm-search__input:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.sm-pager { display: flex; justify-content: center; gap: 4px; margin-block-start: 14px; }

.sm-pager__btn {
  display: grid;
  place-items: center;
  min-width: 30px;
  height: 30px;
  padding-inline: 8px;
  border: 1px solid var(--line);
  border-radius: 3px;
  background: var(--surface);
  color: var(--text);
  font-family: var(--f-num);
  font-size: 12px;
  cursor: pointer;
}

.sm-pager__btn:hover:not(:disabled) { background: var(--sunken); }
.sm-pager__btn--on { border-color: var(--acc); color: var(--acc-ink); font-weight: 600; }
.sm-pager__btn:disabled { opacity: 0.4; cursor: not-allowed; }
.sm-pager__btn:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

[dir='rtl'] .sm-pager__btn .iconify { transform: scaleX(-1); }
</style>
