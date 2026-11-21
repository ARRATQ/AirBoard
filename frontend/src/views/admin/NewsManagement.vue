<template>
  <div class="content-area">
    <!-- En-tête : libellés propres au périmètre (toute l'organisation ou groupes administrés) -->
    <PageHeader>
      <template #title>{{ isGlobal ? $t('news.center.title') : $t('groupAdmin.newsManagement.title') }}</template>
      <template #description>{{ isGlobal ? $t('adminShared.news.subtitle') : $t('groupAdmin.newsManagement.subtitle') }}</template>
      <template #actions>
        <router-link to="/admin/news/new" class="btn btn-primary">
          <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
          {{ $t('groupAdmin.newsManagement.newArticle') }}
        </router-link>
      </template>
    </PageHeader>

    <!-- Rappel du périmètre pour l'admin de groupe -->
    <div v-if="!isGlobal" class="mb-6 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-300 dark:border-blue-600 rounded-lg">
      <div class="flex items-start">
        <Icon icon="mdi:information" class="h-5 w-5 text-blue-600 dark:text-blue-500 mr-3 mt-0.5" />
        <div>
          <h3 class="text-blue-800 dark:text-blue-500 font-medium">{{ $t('groupAdmin.newsManagement.infoTitle') }}</h3>
          <p class="text-sm text-blue-700 dark:text-blue-200 mt-1">
            {{ $t('groupAdmin.newsManagement.infoDescription') }}
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
            :placeholder="$t('groupAdmin.newsManagement.searchPlaceholder')"
            class="input"
          />
        </div>

        <select v-model="filters.category" class="input w-full sm:w-48">
          <option value="">{{ $t('groupAdmin.newsManagement.categoryFilter.all') }}</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>

        <select v-model="filters.status" class="input w-full sm:w-40">
          <option value="">{{ $t('groupAdmin.newsManagement.statusFilter.all') }}</option>
          <option value="published">{{ $t('groupAdmin.newsManagement.statusFilter.published') }}</option>
          <option value="draft">{{ $t('groupAdmin.newsManagement.statusFilter.draft') }}</option>
        </select>

        <select v-model="filters.priority" class="input w-full sm:w-40">
          <option value="">{{ $t('groupAdmin.newsManagement.priorityFilter.all') }}</option>
          <option value="urgent">{{ $t('groupAdmin.newsManagement.priorityFilter.urgent') }}</option>
          <option value="important">{{ $t('groupAdmin.newsManagement.priorityFilter.important') }}</option>
          <option value="normal">{{ $t('groupAdmin.newsManagement.priorityFilter.normal') }}</option>
        </select>
      </div>
    </div>

    <!-- Chargement -->
    <div v-if="isLoading" class="flex justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-gray-400" />
    </div>

    <!-- Liste -->
    <div v-else-if="newsList.length > 0" class="space-y-4">
      <div
        v-for="news in newsList"
        :key="news.id"
        class="card hover:shadow-lg transition-shadow"
      >
        <div class="flex items-start gap-4">
          <!-- Pastille de statut -->
          <div class="flex-shrink-0 pt-1">
            <div
              :class="news.is_published ? 'bg-green-500' : 'bg-gray-400'"
              class="h-3 w-3 rounded-full"
              :title="news.is_published ? $t('groupAdmin.newsManagement.status.published') : $t('groupAdmin.newsManagement.status.draft')"
            ></div>
          </div>

          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <Icon
                    v-if="news.is_pinned"
                    icon="mdi:pin"
                    class="h-4 w-4 text-yellow-500"
                  />

                  <h3
                    class="text-lg font-semibold text-gray-900 dark:text-white cursor-pointer hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
                    @click="viewNews(news)"
                  >
                    {{ news.title }}
                  </h3>

                  <span
                    v-if="news.priority !== 'normal'"
                    :class="{
                      'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200': news.priority === 'urgent',
                      'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200': news.priority === 'important'
                    }"
                    class="px-2 py-0.5 rounded text-xs font-medium"
                  >
                    {{ $t(`groupAdmin.newsManagement.priority.${news.priority}`) }}
                  </span>
                </div>

                <p v-if="news.summary" class="text-sm text-gray-600 dark:text-gray-400 mb-2">
                  {{ news.summary }}
                </p>

                <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                  <div v-if="news.category" class="flex items-center gap-1">
                    <Icon :icon="news.category.icon" class="h-4 w-4" />
                    <span>{{ news.category.name }}</span>
                  </div>

                  <div class="flex items-center gap-1">
                    <Icon icon="mdi:account" class="h-4 w-4" />
                    <span>{{ news.author?.username || $t('groupAdmin.newsManagement.author.unknown') }}</span>
                  </div>

                  <div class="flex items-center gap-1">
                    <Icon icon="mdi:calendar" class="h-4 w-4" />
                    <span>{{ formatDate(news.published_at || news.created_at) }}</span>
                  </div>

                  <div class="flex items-center gap-1">
                    <Icon icon="mdi:eye" class="h-4 w-4" />
                    <span>{{ news.view_count }}</span>
                  </div>

                  <!-- Groupes cibles (renvoyés par l'API quel que soit le périmètre) -->
                  <div
                    v-if="news.target_groups && news.target_groups.length > 0"
                    class="flex items-center gap-1"
                    :title="$t('groupAdmin.newsManagement.targetGroups')"
                  >
                    <Icon icon="mdi:account-group" class="h-4 w-4" />
                    <div class="flex gap-1">
                      <span
                        v-for="group in news.target_groups.slice(0, 2)"
                        :key="group.id"
                        class="px-1.5 py-0.5 bg-primary-100 dark:bg-primary-900 rounded text-xs"
                      >
                        {{ group.name }}
                      </span>
                      <span v-if="news.target_groups.length > 2" class="text-xs">
                        +{{ news.target_groups.length - 2 }}
                      </span>
                    </div>
                  </div>

                  <div v-if="news.tags && news.tags.length > 0" class="flex items-center gap-1">
                    <Icon icon="mdi:tag" class="h-4 w-4" />
                    <div class="flex gap-1">
                      <span
                        v-for="tag in news.tags.slice(0, 2)"
                        :key="tag.id"
                        class="px-1.5 py-0.5 bg-gray-100 dark:bg-gray-700 rounded text-xs"
                      >
                        {{ tag.name }}
                      </span>
                      <span v-if="news.tags.length > 2" class="text-xs">
                        +{{ news.tags.length - 2 }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Actions -->
              <div class="flex items-center gap-2">
                <router-link
                  :to="`/admin/news/${news.slug}/edit`"
                  class="btn btn-secondary btn-sm"
                  :title="$t('groupAdmin.newsManagement.actions.edit')"
                  :aria-label="$t('groupAdmin.newsManagement.actions.edit')"
                >
                  <Icon icon="mdi:pencil" class="h-4 w-4" />
                </router-link>

                <!-- Épingler : endpoint réservé à l'admin global -->
                <button
                  v-if="isAdmin"
                  type="button"
                  class="btn btn-secondary btn-sm"
                  :title="news.is_pinned ? $t('adminShared.news.unpin') : $t('adminShared.news.pin')"
                  :aria-label="news.is_pinned ? $t('adminShared.news.unpin') : $t('adminShared.news.pin')"
                  @click="togglePin(news)"
                >
                  <Icon :icon="news.is_pinned ? 'mdi:pin-off' : 'mdi:pin'" class="h-4 w-4" />
                </button>

                <button
                  type="button"
                  class="btn btn-danger btn-sm"
                  :title="$t('groupAdmin.newsManagement.actions.delete')"
                  :aria-label="$t('groupAdmin.newsManagement.actions.delete')"
                  @click="confirmDelete(news)"
                >
                  <Icon icon="mdi:delete" class="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="pagination.total_pages > 1" class="flex justify-center gap-2 mt-6">
        <button
          type="button"
          :disabled="pagination.page === 1"
          class="btn btn-secondary btn-sm"
          @click="changePage(pagination.page - 1)"
        >
          <Icon icon="mdi:chevron-left" class="h-4 w-4" />
        </button>

        <span class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300">
          {{ $t('groupAdmin.newsManagement.pagination.page', { page: pagination.page, pages: pagination.total_pages }) }}
        </span>

        <button
          type="button"
          :disabled="pagination.page === pagination.total_pages"
          class="btn btn-secondary btn-sm"
          @click="changePage(pagination.page + 1)"
        >
          <Icon icon="mdi:chevron-right" class="h-4 w-4" />
        </button>
      </div>
    </div>

    <!-- État vide -->
    <div v-else class="empty-state">
      <Icon icon="mdi:newspaper-variant" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('groupAdmin.newsManagement.empty.title') }}</h3>
      <p class="empty-state-description">
        {{ emptyMessage }}
      </p>
      <router-link v-if="!hasActiveFilters" to="/admin/news/new" class="btn btn-primary mt-4">
        <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
        {{ $t('groupAdmin.newsManagement.empty.createButton') }}
      </router-link>
    </div>
  </div>
</template>

<script setup>
import PageHeader from '@/components/ui/PageHeader.vue'
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
    // Le filtre de statut n'a pas d'équivalent côté API : il n'est pas transmis
    const params = {
      page: pagination.value.page,
      page_size: PAGE_SIZE,
      search: filters.value.search || undefined,
      category_id: filters.value.category || undefined,
      priority: filters.value.priority || undefined
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
  return new Date(dateString).toLocaleDateString(locale.value, {
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
