<template>
  <div class="w-full p-6">
    <div class="mb-8">
      <div class="flex items-start justify-between gap-4 flex-wrap mb-4">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">{{ t('suggestions.admin.title') }}</h1>
          <p class="text-gray-600 dark:text-gray-400">{{ t('suggestions.admin.subtitle') }}</p>
        </div>
      </div>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
        <div class="md:col-span-2">
          <div class="relative">
            <Icon icon="mdi:magnify" class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
            <input
              v-model="filters.search"
              type="text"
              :placeholder="t('suggestions.filters.search')"
              class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
              @input="debounceLoad"
            />
          </div>
        </div>

        <select
          v-model="filters.category_id"
          class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
          @change="loadSuggestions"
        >
          <option value="">{{ t('suggestions.filters.allCategories') }}</option>
          <option v-for="category in categories" :key="category.id" :value="String(category.id)">
            {{ category.name }}
          </option>
        </select>

        <select
          v-model="filters.status"
          class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
          @change="loadSuggestions"
        >
          <option value="">{{ t('suggestions.filters.allStatuses') }}</option>
          <option v-for="status in statuses" :key="status" :value="status">
            {{ t(`suggestions.statuses.${status}`) }}
          </option>
        </select>

        <select
          v-model="filters.archived"
          class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
          @change="loadSuggestions"
        >
          <option value="">{{ t('suggestions.admin.archivedAll') }}</option>
          <option value="0">{{ t('suggestions.admin.archivedNo') }}</option>
          <option value="1">{{ t('suggestions.admin.archivedYes') }}</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-primary-500" />
    </div>

    <div v-else-if="suggestions.length === 0" class="text-center py-12">
      <Icon icon="mdi:lightbulb-off-outline" class="h-16 w-16 mx-auto text-gray-400 mb-4" />
      <p class="text-gray-600 dark:text-gray-400 text-lg">{{ t('suggestions.emptyTitle') }}</p>
    </div>

    <div v-else class="space-y-4">
      <article
        v-for="item in suggestions"
        :key="item.id"
        class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 border border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-start justify-between gap-4 flex-wrap">
          <div class="min-w-0 flex-1">
            <RouterLink :to="`/suggestions/${item.id}`" class="font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400">
              {{ item.title }}
            </RouterLink>
            <div class="mt-2 text-xs text-gray-500 dark:text-gray-400 flex items-center gap-2 flex-wrap">
              <span class="px-2 py-1 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
                {{ item.category_ref?.name || item.category || t('suggestions.noCategory') }}
              </span>
              <span class="px-2 py-1 rounded-full" :class="item.is_archived ? 'bg-gray-200 text-gray-700 dark:bg-gray-700 dark:text-gray-300' : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'">
                {{ item.is_archived ? t('suggestions.admin.archivedBadge') : t('suggestions.admin.activeBadge') }}
              </span>
              <span>{{ formatAuthor(item) }} · {{ formatDate(item.created_at) }}</span>
            </div>
          </div>

          <div class="flex items-center gap-2 flex-wrap">
            <select
              class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm dark:bg-gray-700 dark:text-white"
              :value="item.status"
              @change="updateStatus(item, $event.target.value)"
            >
              <option v-for="status in statuses" :key="status" :value="status">
                {{ t(`suggestions.statuses.${status}`) }}
              </option>
            </select>

            <button
              class="px-3 py-2 rounded-lg text-sm font-semibold border transition-colors"
              :class="item.is_archived
                ? 'border-emerald-300 text-emerald-700 bg-emerald-50 hover:bg-emerald-100 dark:border-emerald-700 dark:text-emerald-300 dark:bg-emerald-900/20'
                : 'border-amber-300 text-amber-700 bg-amber-50 hover:bg-amber-100 dark:border-amber-700 dark:text-amber-300 dark:bg-amber-900/20'"
              @click="toggleArchive(item)"
            >
              {{ item.is_archived ? t('suggestions.admin.activate') : t('suggestions.admin.archive') }}
            </button>
          </div>
        </div>
      </article>
    </div>

    <div v-if="pagination.total_pages > 1" class="mt-8 flex justify-center">
      <nav class="flex items-center gap-2">
        <button
          @click="changePage(pagination.page - 1)"
          :disabled="pagination.page === 1"
          class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <Icon icon="mdi:chevron-left" class="h-5 w-5" />
        </button>

        <button
          v-for="page in visiblePages"
          :key="page"
          @click="changePage(page)"
          :class="[
            'px-4 py-2 rounded-lg border',
            page === pagination.page
              ? 'bg-primary-500 text-white border-primary-500'
              : 'border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'
          ]"
        >
          {{ page }}
        </button>

        <button
          @click="changePage(pagination.page + 1)"
          :disabled="pagination.page === pagination.total_pages"
          class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <Icon icon="mdi:chevron-right" class="h-5 w-5" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { suggestionsService } from '@/services/api'

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

const changePage = (page) => {
  if (page < 1 || page > pagination.value.total_pages) return
  pagination.value.page = page
  loadSuggestions()
}

const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString(locale.value)
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
