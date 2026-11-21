<template>
  <div class="content-area">
    <!-- En-tête : libellés propres au périmètre (toute l'organisation ou groupes administrés) -->
    <PageHeader>
      <template #title>{{ isGlobal ? $t('news.center.title') : $t('groupAdmin.newsManagement.title') }}</template>
      <template #description>{{ isGlobal ? $t('adminShared.news.subtitle') : $t('groupAdmin.newsManagement.subtitle') }}</template>
      <template #actions>
        <BaseButton tag="router-link" to="/admin/news/new" variant="primary" icon="mdi:plus">
          {{ $t('groupAdmin.newsManagement.newArticle') }}
        </BaseButton>
      </template>
    </PageHeader>

    <!-- Rappel du périmètre pour l'admin de groupe -->
    <div v-if="!isGlobal" class="nw-callout" role="note">
      <Icon icon="mdi:information-outline" class="nw-callout__icon" aria-hidden="true" />
      <div>
        <p class="nw-callout__title">{{ $t('groupAdmin.newsManagement.infoTitle') }}</p>
        <p class="nw-callout__text">{{ $t('groupAdmin.newsManagement.infoDescription') }}</p>
      </div>
    </div>

    <!-- Aucune actualité et aucun filtre : invitation à en créer une -->
    <EmptyState
      v-if="!isLoading && newsList.length === 0 && !hasActiveFilters"
      icon="mdi:newspaper-variant"
      :title="$t('groupAdmin.newsManagement.empty.title')"
      :description="emptyMessage"
    >
      <BaseButton tag="router-link" to="/admin/news/new" variant="primary" size="sm" icon="mdi:plus">
        {{ $t('groupAdmin.newsManagement.empty.createButton') }}
      </BaseButton>
    </EmptyState>

    <template v-else>
      <!-- Recherche, catégorie et priorité sont gérées côté serveur : seule la page courante est passée au tableau -->
      <DataTable
        :columns="columns"
        :rows="rows"
        :loading="isLoading"
        :searchable="false"
      :footer="false"
        :page-size="100"
        :empty-title="$t('groupAdmin.newsManagement.empty.title')"
      >
        <template #toolbar>
          <label class="nw-search">
            <Icon icon="mdi:magnify" class="nw-search__icon" aria-hidden="true" />
            <input
              v-model="filters.search"
              type="search"
              class="nw-search__input"
              :placeholder="$t('groupAdmin.newsManagement.searchPlaceholder')"
              :aria-label="$t('groupAdmin.newsManagement.searchPlaceholder')"
            />
          </label>
          <label>
            <span class="sr-only">{{ $t('ui.table.columns.category') }}</span>
            <select v-model="filters.category" class="ui-select">
              <option value="">{{ $t('groupAdmin.newsManagement.categoryFilter.all') }}</option>
              <option v-for="cat in categories" :key="cat.id" :value="cat.id">{{ cat.name }}</option>
            </select>
          </label>
          <label>
            <span class="sr-only">{{ $t('ui.table.columns.status') }}</span>
            <select v-model="filters.status" class="ui-select">
              <option value="">{{ $t('groupAdmin.newsManagement.statusFilter.all') }}</option>
              <option value="published">{{ $t('groupAdmin.newsManagement.statusFilter.published') }}</option>
              <option value="draft">{{ $t('groupAdmin.newsManagement.statusFilter.draft') }}</option>
            </select>
          </label>
          <label>
            <span class="sr-only">{{ $t('groupAdmin.newsManagement.priorityFilter.all') }}</span>
            <select v-model="filters.priority" class="ui-select">
              <option value="">{{ $t('groupAdmin.newsManagement.priorityFilter.all') }}</option>
              <option value="urgent">{{ $t('groupAdmin.newsManagement.priorityFilter.urgent') }}</option>
              <option value="important">{{ $t('groupAdmin.newsManagement.priorityFilter.important') }}</option>
              <option value="normal">{{ $t('groupAdmin.newsManagement.priorityFilter.normal') }}</option>
            </select>
          </label>
        </template>

        <template #cell-title="{ row }">
          <span class="nw-title">
            <span class="nw-title__line">
              <Icon v-if="row.raw.is_pinned" icon="mdi:pin" class="nw-title__pin" :title="$t('adminShared.news.pin')" />
              <button type="button" class="nw-title__link" @click="viewNews(row.raw)">{{ row.title }}</button>
              <StatusBadge v-if="row.priorityLabel" :tone="row.raw.priority === 'urgent' ? 'bad' : 'warn'">{{ row.priorityLabel }}</StatusBadge>
            </span>
            <span v-if="row.raw.summary" class="nw-title__sub">{{ row.raw.summary }}</span>
          </span>
        </template>

        <template #cell-categoryName="{ row }">
          <span v-if="row.raw.category" class="nw-cat">
            <Icon v-if="row.raw.category.icon" :icon="row.raw.category.icon" class="nw-cat__icon" aria-hidden="true" />
            {{ row.categoryName }}
          </span>
        </template>

        <template #cell-groupNames="{ row }">
          <span v-if="row.groupNames" class="nw-groups" :title="$t('groupAdmin.newsManagement.targetGroups')">{{ row.groupNames }}</span>
        </template>

        <template #cell-date="{ row }">
          <span class="nw-mono">{{ row.dateLabel }}</span>
        </template>

        <template #cell-views="{ row }">
          <span class="nw-mono">{{ row.views }}</span>
        </template>

        <template #cell-status="{ row }">
          <StatusBadge :tone="row.raw.is_published ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
        </template>

        <template #cell-actions="{ row }">
          <RowActions :label="row.title" :editable="false" @delete="confirmDelete(row.raw)">
            <router-link
              :to="`/admin/news/${row.raw.slug}/edit`"
              class="row-actions__btn"
              :title="$t('groupAdmin.newsManagement.actions.edit')"
              :aria-label="`${$t('groupAdmin.newsManagement.actions.edit')} : ${row.title}`"
            >
              <Icon icon="mdi:pencil-outline" aria-hidden="true" />
            </router-link>
            <!-- Épingler : endpoint réservé à l'admin global -->
            <button
              v-if="isAdmin"
              type="button"
              class="row-actions__btn"
              :title="row.raw.is_pinned ? $t('adminShared.news.unpin') : $t('adminShared.news.pin')"
              :aria-label="`${row.raw.is_pinned ? $t('adminShared.news.unpin') : $t('adminShared.news.pin')} : ${row.title}`"
              @click="togglePin(row.raw)"
            >
              <Icon :icon="row.raw.is_pinned ? 'mdi:pin-off-outline' : 'mdi:pin-outline'" aria-hidden="true" />
            </button>
          </RowActions>
        </template>
      </DataTable>

      <!-- Pagination serveur -->
      <nav v-if="pagination.total_pages > 1" class="nw-pager" :aria-label="$t('groupAdmin.newsManagement.pagination.page', { page: pagination.page, pages: pagination.total_pages })">
        <BaseButton variant="secondary" size="sm" icon="mdi:chevron-left" :disabled="pagination.page === 1" :aria-label="$t('ui.table.prev')" @click="changePage(pagination.page - 1)" />
        <span class="nw-mono">{{ $t('groupAdmin.newsManagement.pagination.page', { page: pagination.page, pages: pagination.total_pages }) }}</span>
        <BaseButton variant="secondary" size="sm" icon="mdi:chevron-right" :disabled="pagination.page === pagination.total_pages" :aria-label="$t('ui.table.next')" @click="changePage(pagination.page + 1)" />
      </nav>
    </template>
  </div>
</template>

<script setup>
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { dateLocaleFor } from '@/utils/homeFormat'
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { newsService, groupAdminService } from '@/services/api'
import { CONTENT } from '@/config/access'
import { useManagementScope } from '@/composables/useManagementScope'
import { useConfirm } from '@/composables/useConfirm'

const PAGE_SIZE = 20

const router = useRouter()
const { t, locale } = useI18n()
const appStore = useAppStore()
const { confirm: askConfirm } = useConfirm()

// Admin et éditeur : toutes les actualités ; admin de groupe seul : celles de ses groupes
const { isGlobal, isAdmin } = useManagementScope(CONTENT)

const isLoading = ref(false)
const newsList = ref([])
const categories = ref([])

const filters = ref({
  search: '',
  category: '',
  status: '',
  priority: ''
})

const pagination = ref({
  page: 1,
  total: 0,
  total_pages: 0
})

const hasActiveFilters = computed(() =>
  Boolean(filters.value.search || filters.value.category || filters.value.status || filters.value.priority)
)

const emptyMessage = computed(() => {
  if (hasActiveFilters.value) return t('groupAdmin.newsManagement.empty.noResults')
  return isGlobal.value ? t('adminShared.news.createFirst') : t('groupAdmin.newsManagement.empty.createFirst')
})

// Service selon le périmètre : API éditeur/admin ou API group-admin (filtrée côté serveur)
const listNews = (params) => (isGlobal.value ? newsService.getNews(params) : groupAdminService.getNews(params))
const removeNews = (id) => (isGlobal.value ? newsService.deleteNews(id) : groupAdminService.deleteNews(id))

const loadNews = async () => {
  try {
    isLoading.value = true
    const params = {
      page: pagination.value.page,
      page_size: PAGE_SIZE,
      search: filters.value.search || undefined,
      category_id: filters.value.category || undefined,
      priority: filters.value.priority || undefined,
      status: filters.value.status || undefined
    }
    const response = await listNews(params)
    newsList.value = response.news || []
    pagination.value = {
      page: response.page || 1,
      total: response.total || 0,
      total_pages: response.total_pages || 1
    }
  } catch (error) {
    console.error('Erreur au chargement des actualités :', error)
    newsList.value = []
    appStore.showError(t('adminShared.news.loadError'))
  } finally {
    isLoading.value = false
  }
}

const loadCategories = async () => {
  try {
    const data = await newsService.getCategories()
    categories.value = Array.isArray(data) ? data : (data?.data || [])
  } catch (error) {
    console.error('Erreur au chargement des catégories :', error)
    categories.value = []
  }
}

const togglePin = async (news) => {
  try {
    await newsService.togglePin(news.id)
    appStore.showSuccess(news.is_pinned ? t('adminShared.news.unpinned') : t('adminShared.news.pinned'))
    await loadNews()
  } catch (error) {
    console.error('Erreur lors de l\'épinglage :', error)
    appStore.showError(t('adminShared.news.pinError'))
  }
}

const confirmDelete = async (news) => {
  const confirmed = await askConfirm({
    title: t('groupAdmin.newsManagement.deleteConfirm.title'),
    message: t('groupAdmin.newsManagement.deleteConfirm.message', { title: news.title }),
    confirmLabel: t('common.delete'),
    tone: 'danger'
  })
  if (!confirmed) return

  try {
    await removeNews(news.id)
    appStore.showSuccess(t('adminShared.news.deleteSuccess'))
    await loadNews()
  } catch (error) {
    console.error('Erreur lors de la suppression :', error)
    appStore.showError(t('groupAdmin.newsManagement.deleteError'))
  }
}

const changePage = (page) => {
  if (page >= 1 && page <= pagination.value.total_pages) {
    pagination.value = { ...pagination.value, page }
    loadNews()
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '—'
  return new Date(dateString).toLocaleDateString(dateLocaleFor(locale.value), {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}

const viewNews = (news) => {
  if (news?.slug) {
    router.push({ name: 'NewsDetail', params: { slug: news.slug } })
  }
}

const columns = computed(() => [
  { key: 'title', label: t('ui.table.columns.title'), sortable: true },
  { key: 'categoryName', label: t('ui.table.columns.category'), sortable: true, hideSm: true },
  { key: 'groupNames', label: t('ui.table.columns.groups'), hideSm: true },
  { key: 'author', label: t('ui.table.columns.author'), sortable: true, hideSm: true },
  { key: 'date', label: t('ui.table.columns.date'), sortable: true, hideSm: true },
  { key: 'views', label: t('reports.views'), sortable: true, align: 'end', hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'actions', label: '', align: 'end', width: '120px' },
])

// Le statut est filtré par l'API, sur toute la liste.
const rows = computed(() =>
  newsList.value
    .map(news => ({
      id: news.id,
      title: news.title,
      categoryName: news.category?.name || '',
      groupNames: (news.target_groups || []).map(g => g.name).join(', '),
      author: news.author?.username || t('groupAdmin.newsManagement.author.unknown'),
      date: news.published_at || news.created_at || '',
      dateLabel: formatDate(news.published_at || news.created_at),
      views: news.view_count ?? 0,
      status: news.is_published ? t('groupAdmin.newsManagement.status.published') : t('groupAdmin.newsManagement.status.draft'),
      priorityLabel: news.priority && news.priority !== 'normal' ? t(`groupAdmin.newsManagement.priority.${news.priority}`) : '',
      raw: news,
    })),
)

// Un changement de filtre ramène à la première page
watch(
  () => [filters.value.search, filters.value.category, filters.value.status, filters.value.priority],
  () => {
    pagination.value = { ...pagination.value, page: 1 }
    loadNews()
  }
)

onMounted(() => {
  loadCategories()
  loadNews()
})
</script>

<style scoped>
.nw-callout {
  display: flex;
  gap: 10px;
  margin-block-end: 16px;
  padding: 12px 14px;
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--acc);
  border-radius: 4px;
  background: var(--surface);
}
.nw-callout__icon { flex: none; width: 18px; height: 18px; margin-block-start: 1px; color: var(--acc-ink); }
.nw-callout__title { margin: 0; font-size: 13.5px; font-weight: 600; color: var(--text); }
.nw-callout__text { margin: 2px 0 0; font-size: 12.5px; color: var(--muted); }

.nw-search { position: relative; display: flex; align-items: center; flex: 1 1 220px; max-width: 340px; }
.nw-search__icon { position: absolute; inset-inline-start: 9px; width: 16px; height: 16px; color: var(--muted); pointer-events: none; }
.nw-search__input {
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
.nw-search__input:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.nw-title { display: flex; flex-direction: column; min-width: 0; }
.nw-title__line { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }
.nw-title__pin { flex: none; width: 15px; height: 15px; color: var(--warn); }
.nw-title__link {
  padding: 0;
  border: 0;
  background: none;
  font: inherit;
  font-weight: 600;
  color: var(--text);
  text-align: start;
  cursor: pointer;
}
.nw-title__link:hover { color: var(--acc-ink); text-decoration: underline; }
.nw-title__link:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; border-radius: 2px; }
.nw-title__sub {
  max-width: 52ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nw-cat { display: inline-flex; align-items: center; gap: 5px; font-size: 12.5px; color: var(--text); }
.nw-cat__icon { width: 15px; height: 15px; color: var(--muted); }
.nw-groups { font-size: 12px; color: var(--muted); }
.nw-mono { font-family: var(--f-num); font-size: 11.5px; letter-spacing: 0.04em; color: var(--muted); white-space: nowrap; }
.nw-pager { display: flex; align-items: center; justify-content: center; gap: 12px; margin-block-start: 12px; }
</style>
