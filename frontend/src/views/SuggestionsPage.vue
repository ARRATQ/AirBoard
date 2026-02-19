<template>
  <div class="w-full p-6">
    <div class="mb-8">
      <div class="flex items-start justify-between gap-4 flex-wrap mb-4">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
          <Icon icon="mdi:lightbulb-on-outline" class="inline mr-3 text-3xl text-amber-500" />
          {{ t('suggestions.title') }}
        </h1>
        <p class="text-gray-600 dark:text-gray-400">{{ t('suggestions.subtitle') }}</p>
      </div>

        <div class="flex items-center gap-3">
          <div class="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
            <button
              type="button"
              class="flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-all duration-200"
              :class="viewMode === 'grid' ? 'bg-white dark:bg-gray-700 text-primary-600 dark:text-primary-400 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'"
              @click="viewMode = 'grid'"
            >
              <Icon icon="mdi:view-grid" class="h-4 w-4" />
              <span class="hidden sm:inline">Grille</span>
            </button>
            <button
              type="button"
              class="flex items-center gap-2 px-3 py-2 rounded-md text-sm font-medium transition-all duration-200"
              :class="viewMode === 'list' ? 'bg-white dark:bg-gray-700 text-primary-600 dark:text-primary-400 shadow-sm' : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'"
              @click="viewMode = 'list'"
            >
              <Icon icon="mdi:view-list" class="h-4 w-4" />
              <span class="hidden sm:inline">Liste</span>
            </button>
          </div>

          <button class="suggestion-cta" type="button" @click="openCreateModal">
            <Icon icon="mdi:plus-circle" class="h-5 w-5" />
            {{ t('suggestions.newSuggestion') }}
          </button>
        </div>
      </div>
    </div>

    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-3">
        <div class="md:col-span-2">
          <div class="relative">
            <Icon icon="mdi:magnify" class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
            <input
              v-model="filters.search"
              type="text"
              class="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
              :placeholder="t('suggestions.filters.search')"
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
          <option v-for="category in categoryOptions" :key="category.id" :value="String(category.id)">
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
      </div>
    </div>

    <div v-if="loading" class="loading-state">
      <Icon icon="mdi:loading" class="animate-spin text-2xl text-amber-500" />
      <p>{{ t('suggestions.loading') }}</p>
    </div>

    <div v-else-if="suggestions.length === 0" class="empty-state">
      <Icon icon="mdi:lightbulb-outline" class="text-5xl text-gray-400" />
      <h3>{{ t('suggestions.emptyTitle') }}</h3>
      <p>{{ t('suggestions.emptyDescription') }}</p>
    </div>

    <transition-group
      v-else-if="viewMode === 'list'"
      name="fade-slide"
      tag="div"
      class="space-y-4"
    >
      <article
        v-for="item in suggestions"
        :key="item.id"
        class="card p-5 border border-gray-200 dark:border-gray-700"
      >
        <div class="flex items-start justify-between gap-4 flex-wrap">
          <div class="flex-1 min-w-0">
            <RouterLink :to="`/suggestions/${item.id}`" class="group">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-amber-600 dark:group-hover:text-amber-400 transition-colors">
                {{ item.title }}
              </h3>
            </RouterLink>

            <div class="mt-2 flex items-center gap-2 flex-wrap text-xs">
              <span class="px-2 py-1 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
                {{ item.category_ref?.name || item.category || t('suggestions.noCategory') }}
              </span>
              <span class="px-2 py-1 rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
                {{ t(`suggestions.statuses.${item.status}`) }}
              </span>
              <span class="text-gray-500 dark:text-gray-400">
                {{ formatAuthor(item) }} · {{ formatDate(item.created_at) }}
              </span>
            </div>

            <p class="mt-3 text-sm text-gray-700 dark:text-gray-300 whitespace-pre-line line-clamp-3">{{ item.description }}</p>

            <div v-if="item.poll" class="mt-3 text-xs text-gray-500 dark:text-gray-400 flex items-center gap-2">
              <Icon icon="mdi:poll" class="h-4 w-4" />
              <span>{{ t('suggestions.linkedPoll') }}: {{ item.poll.title }}</span>
            </div>
          </div>

          <div class="flex items-center gap-3 text-sm text-gray-600 dark:text-gray-300">
            <span class="vote-chip vote-chip-up">
              <Icon icon="mdi:thumb-up-outline" class="h-4 w-4" />
              {{ item.upvotes || 0 }}
            </span>
            <span class="vote-chip vote-chip-down">
              <Icon icon="mdi:thumb-down-outline" class="h-4 w-4" />
              {{ item.downvotes || 0 }}
            </span>
            <RouterLink :to="`/suggestions/${item.id}`" class="suggestion-detail-btn">
              {{ t('suggestions.viewDetails') }}
              <Icon icon="mdi:arrow-right" class="h-4 w-4" />
            </RouterLink>
          </div>
        </div>
      </article>
    </transition-group>

    <transition-group
      v-else
      name="fade-scale"
      tag="div"
      class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6"
    >
      <article
        v-for="item in suggestions"
        :key="item.id"
        class="card p-5 border border-gray-200 dark:border-gray-700 h-full flex flex-col"
      >
        <RouterLink :to="`/suggestions/${item.id}`" class="group">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-amber-600 dark:group-hover:text-amber-400 transition-colors line-clamp-2">
            {{ item.title }}
          </h3>
        </RouterLink>

        <div class="mt-2 flex items-center gap-2 flex-wrap text-xs">
          <span class="px-2 py-1 rounded-full bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300">
            {{ item.category_ref?.name || item.category || t('suggestions.noCategory') }}
          </span>
          <span class="px-2 py-1 rounded-full bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
            {{ t(`suggestions.statuses.${item.status}`) }}
          </span>
        </div>

        <p class="mt-3 text-sm text-gray-700 dark:text-gray-300 whitespace-pre-line line-clamp-4 flex-1">{{ item.description }}</p>

        <div class="mt-4 flex items-center justify-between gap-2 text-sm text-gray-600 dark:text-gray-300">
          <span class="vote-chip vote-chip-up">
            <Icon icon="mdi:thumb-up-outline" class="h-4 w-4" />
            {{ item.upvotes || 0 }}
          </span>
          <span class="vote-chip vote-chip-down">
            <Icon icon="mdi:thumb-down-outline" class="h-4 w-4" />
            {{ item.downvotes || 0 }}
          </span>
          <RouterLink :to="`/suggestions/${item.id}`" class="suggestion-detail-btn">
            {{ t('suggestions.viewDetails') }}
            <Icon icon="mdi:arrow-right" class="h-4 w-4" />
          </RouterLink>
        </div>
      </article>
    </transition-group>

    <div v-if="pagination.total_pages > 1" class="flex justify-center items-center gap-2">
      <button class="btn-secondary btn-sm" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
        <Icon icon="mdi:chevron-left" />
        {{ t('polls.pagination.previous') }}
      </button>
      <span class="text-sm text-gray-600 dark:text-gray-400">{{ pagination.page }} / {{ pagination.total_pages }}</span>
      <button class="btn-secondary btn-sm" :disabled="pagination.page >= pagination.total_pages" @click="changePage(pagination.page + 1)">
        {{ t('polls.pagination.next') }}
        <Icon icon="mdi:chevron-right" />
      </button>
    </div>

    <Teleport to="body">
      <div
        v-if="showCreateModal"
        class="fixed inset-0 z-50 bg-black/50 flex items-center justify-center p-4"
        @click.self="closeCreateModal"
      >
        <div class="w-full max-w-2xl bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('suggestions.form.title') }}</h2>
            <button class="btn-ghost p-2" @click="closeCreateModal" type="button">
              <Icon icon="mdi:close" />
            </button>
          </div>

          <div class="p-5">
            <div
              v-if="pollContext.active"
              class="mb-4 rounded-lg border border-amber-200 bg-amber-50 dark:bg-amber-900/20 dark:border-amber-800 p-3"
            >
              <div class="text-sm text-amber-900 dark:text-amber-300">
                {{ t('suggestions.form.fromPoll') }}: <strong>{{ pollContext.pollTitle }}</strong>
              </div>
              <div v-if="pollContext.optionText" class="text-xs text-amber-800 dark:text-amber-400 mt-1">
                {{ t('suggestions.form.selectedOption') }}: {{ pollContext.optionText }}
              </div>
            </div>

            <form class="space-y-4" @submit.prevent="submitSuggestion">
              <div>
                <label class="form-label form-label-required" for="suggestion-title">{{ t('suggestions.form.subject') }}</label>
                <input
                  id="suggestion-title"
                  v-model="form.title"
                  type="text"
                  class="form-input"
                  :placeholder="t('suggestions.form.subjectPlaceholder')"
                  required
                />
              </div>

              <div>
                <label class="form-label form-label-required" for="suggestion-category">{{ t('suggestions.form.category') }}</label>
                <select id="suggestion-category" v-model="form.category_id" class="form-select" required>
                  <option v-if="categoryOptions.length === 0" :value="null">{{ t('suggestions.noCategory') }}</option>
                  <option v-for="category in categoryOptions" :key="category.id" :value="category.id">
                    {{ category.name }}
                  </option>
                </select>
              </div>

              <div>
                <label class="form-label form-label-required" for="suggestion-description">{{ t('suggestions.form.description') }}</label>
                <textarea
                  id="suggestion-description"
                  v-model="form.description"
                  rows="4"
                  class="form-textarea"
                  :placeholder="t('suggestions.form.descriptionPlaceholder')"
                  required
                ></textarea>
              </div>

              <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300 cursor-pointer">
                <input v-model="form.is_anonymous" type="checkbox" class="form-checkbox" />
                {{ t('suggestions.form.anonymous') }}
              </label>

              <div class="flex justify-end gap-2">
                <button class="suggestion-cancel" type="button" @click="closeCreateModal">{{ t('common.cancel') }}</button>
                <button class="suggestion-submit" type="submit" :disabled="submitting || !form.category_id">
                  <Icon v-if="submitting" icon="mdi:loading" class="animate-spin" />
                  <Icon v-else icon="mdi:send" />
                  {{ t('suggestions.form.submit') }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { Icon } from '@iconify/vue'
import { suggestionsService } from '@/services/api'

const { t, locale } = useI18n()
const route = useRoute()

const statuses = ['new', 'reviewing', 'planned', 'in_progress', 'implemented', 'rejected']

const showCreateModal = ref(false)
const viewMode = ref(localStorage.getItem('suggestions-view-mode') || 'list')
const loading = ref(false)
const submitting = ref(false)
const suggestions = ref([])
const categoryOptions = ref([])
const form = ref({
  title: '',
  description: '',
  category_id: null,
  is_anonymous: false
})

const filters = ref({
  search: '',
  category_id: '',
  status: ''
})

const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 1
})

const pollContext = computed(() => ({
  active: route.query.from_poll === '1',
  pollId: route.query.poll_id ? Number(route.query.poll_id) : null,
  pollOptionId: route.query.poll_option_id ? Number(route.query.poll_option_id) : null,
  pollTitle: route.query.poll_title || '',
  optionText: route.query.option_text || ''
}))

let searchTimeout = null

const loadCategories = async () => {
  try {
    const response = await suggestionsService.getCategories()
    categoryOptions.value = response.categories || []
    if (!form.value.category_id && categoryOptions.value.length > 0) {
      form.value.category_id = categoryOptions.value[0].id
    }
  } catch (error) {
    console.error('Failed to load suggestion categories:', error)
    if (window.$toast) {
      window.$toast.error(t('suggestions.categoryLoadError'))
    }
  }
}

const loadSuggestions = async () => {
  try {
    loading.value = true
    const params = {
      search: filters.value.search,
      category_id: filters.value.category_id || undefined,
      status: filters.value.status || undefined,
      page: pagination.value.page,
      page_size: pagination.value.page_size
    }

    const response = await suggestionsService.getSuggestions(params)

    suggestions.value = response.suggestions || []
    pagination.value = {
      page: response.page || 1,
      page_size: response.page_size || 20,
      total: response.total || 0,
      total_pages: response.total_pages || 1
    }
  } catch (error) {
    console.error('Failed to load suggestions:', error)
    if (window.$toast) {
      window.$toast.error(t('suggestions.loadError'))
    }
  } finally {
    loading.value = false
  }
}

const submitSuggestion = async () => {
  const payload = {
    title: form.value.title.trim(),
    description: form.value.description.trim(),
    category_id: form.value.category_id,
    is_anonymous: form.value.is_anonymous
  }

  if (!payload.title || !payload.description || !payload.category_id) {
    if (window.$toast) {
      window.$toast.error(t('suggestions.form.validation'))
    }
    return
  }

  if (payload.description.length < 3) {
    if (window.$toast) {
      window.$toast.error(t('suggestions.form.validationDescriptionMin'))
    }
    return
  }

  if (pollContext.value.active && pollContext.value.pollId) {
    payload.poll_id = pollContext.value.pollId
  }
  if (pollContext.value.active && pollContext.value.pollOptionId) {
    payload.poll_option_id = pollContext.value.pollOptionId
  }

  try {
    submitting.value = true
    await suggestionsService.createSuggestion(payload)

    form.value = {
      title: '',
      description: '',
      category_id: categoryOptions.value[0]?.id || null,
      is_anonymous: false
    }

    if (window.$toast) {
      window.$toast.success(t('suggestions.form.success'))
    }

    closeCreateModal()
    pagination.value.page = 1
    loadSuggestions()
  } catch (error) {
    console.error('Failed to create suggestion:', error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('suggestions.form.error'))
    }
  } finally {
    submitting.value = false
  }
}

const changePage = (page) => {
  if (page < 1 || page > pagination.value.total_pages) return
  pagination.value.page = page
  loadSuggestions()
}

const debounceLoad = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.value.page = 1
    loadSuggestions()
  }, 350)
}

const openCreateModal = () => {
  showCreateModal.value = true
}

const closeCreateModal = () => {
  showCreateModal.value = false
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

watch(viewMode, (value) => {
  localStorage.setItem('suggestions-view-mode', value)
})

onMounted(async () => {
  await loadCategories()
  await loadSuggestions()
  if (pollContext.value.active) {
    showCreateModal.value = true
  }
})
</script>

<style scoped>
.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: all 0.3s ease;
}

.fade-scale-enter-from,
.fade-scale-leave-to {
  opacity: 0;
  transform: scale(0.96);
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: all 0.25s ease;
}

.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(8px);
}

.vote-chip {
  @apply inline-flex items-center gap-1 px-2.5 py-1 rounded-full border text-xs font-semibold;
}

.vote-chip-up {
  @apply border-green-200 text-green-700 bg-green-50 dark:border-green-800 dark:text-green-300 dark:bg-green-900/20;
}

.vote-chip-down {
  @apply border-red-200 text-red-700 bg-red-50 dark:border-red-800 dark:text-red-300 dark:bg-red-900/20;
}

.suggestion-cta {
  @apply inline-flex items-center gap-2 px-4 py-2.5 rounded-lg font-semibold text-white;
  @apply bg-gradient-to-r from-amber-500 to-orange-500 shadow-sm;
  @apply hover:from-amber-600 hover:to-orange-600 hover:shadow-md;
  @apply focus:outline-none focus:ring-2 focus:ring-amber-400 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-900 transition-all duration-200;
}

.suggestion-submit {
  @apply inline-flex items-center gap-2 px-4 py-2 rounded-lg font-semibold text-white disabled:opacity-60 disabled:cursor-not-allowed;
  @apply bg-gradient-to-r from-amber-500 to-orange-500 shadow-sm;
  @apply hover:from-amber-600 hover:to-orange-600 hover:shadow-md;
  @apply focus:outline-none focus:ring-2 focus:ring-amber-400 focus:ring-offset-2;
  @apply dark:focus:ring-offset-gray-800 transition-all duration-200;
}

.suggestion-cancel {
  @apply inline-flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-semibold;
  @apply border border-gray-300 text-gray-700 bg-white hover:bg-gray-50;
  @apply dark:border-gray-600 dark:text-gray-200 dark:bg-gray-800 dark:hover:bg-gray-700;
  @apply focus:outline-none focus:ring-2 focus:ring-gray-300 focus:ring-offset-2 dark:focus:ring-offset-gray-800;
  @apply transition-colors duration-200;
}

.suggestion-detail-btn {
  @apply inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-semibold;
  @apply border border-amber-200 text-amber-700 bg-amber-50;
  @apply hover:bg-amber-100 hover:border-amber-300;
  @apply dark:border-amber-800 dark:text-amber-300 dark:bg-amber-900/20 dark:hover:bg-amber-900/35;
  @apply transition-colors duration-200;
}
</style>
