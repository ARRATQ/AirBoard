<template>
  <div class="w-full p-6">
    <!-- Header -->
    <div class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">
        {{ $t('search.title') }}
      </h1>
      <p class="text-gray-600 dark:text-gray-400">
        {{ $t('search.subtitle') }}
      </p>
    </div>

    <!-- Search Bar -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 mb-6">
      <div class="relative">
        <Icon icon="mdi:magnify" class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
        <input
          ref="searchInput"
          v-model="searchQuery"
          type="text"
          :placeholder="$t('search.placeholder')"
          class="w-full pl-10 pr-10 py-3 text-lg border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
          @input="debouncedSearch"
        />
        <button
          v-if="searchQuery"
          @click="clearSearch"
          class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
        >
          <Icon icon="mdi:close" class="h-5 w-5" />
        </button>
      </div>

      <!-- Type Filter Tabs -->
      <div class="mt-4 flex flex-wrap gap-2">
        <button
          v-for="tab in tabs"
          :key="tab.value"
          @click="setTypeFilter(tab.value)"
          :class="[
            'px-3 py-1.5 rounded-full text-sm font-medium transition-colors',
            activeType === tab.value
              ? 'bg-primary-600 text-white'
              : 'bg-gray-100 text-gray-700 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
          ]"
        >
          <Icon :icon="tab.icon" class="h-4 w-4 inline mr-1" />
          {{ tab.label }}
          <span v-if="tab.value === '' && results" class="ml-1 text-xs opacity-75">({{ results.total_count }})</span>
          <span v-else-if="results && getCountForType(tab.value) > 0" class="ml-1 text-xs opacity-75">({{ getCountForType(tab.value) }})</span>
        </button>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 text-primary-500 animate-spin" />
    </div>

    <!-- Results -->
    <div v-else-if="results && results.total_count > 0" class="space-y-6">
      <!-- Applications -->
      <div v-if="results.applications.length > 0 && (activeType === '' || activeType === 'app')">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
          <Icon icon="mdi:application" class="h-5 w-5 text-indigo-500" />
          {{ $t('search.types.applications') }}
          <span class="text-sm font-normal text-gray-500">({{ results.applications.length }})</span>
        </h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          <a
            v-for="app in results.applications"
            :key="'app-' + app.id"
            :href="app.url"
            target="_blank"
            rel="noopener noreferrer"
            class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 hover:shadow-md transition-shadow flex items-start gap-3 group"
          >
            <div
              class="w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0"
              :style="{ backgroundColor: app.color + '20' }"
            >
              <Icon :icon="app.icon || 'mdi:application'" class="h-5 w-5" :style="{ color: app.color }" />
            </div>
            <div class="min-w-0">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white group-hover:text-primary-600 dark:group-hover:text-primary-400 truncate">
                {{ app.name }}
              </h3>
              <p v-if="app.description" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-0.5">
                {{ app.description }}
              </p>
              <span v-if="app.app_group_name" class="text-[10px] text-gray-400 dark:text-gray-500 mt-1 inline-block">
                {{ app.app_group_name }}
              </span>
            </div>
          </a>
        </div>
      </div>

      <!-- News -->
      <div v-if="results.news.length > 0 && (activeType === '' || activeType === 'news')">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
          <Icon icon="mdi:newspaper" class="h-5 w-5 text-blue-500" />
          {{ $t('search.types.news') }}
          <span class="text-sm font-normal text-gray-500">({{ results.news.length }})</span>
        </h2>
        <div class="space-y-2">
          <router-link
            v-for="item in results.news"
            :key="'news-' + item.id"
            :to="{ name: 'NewsDetail', params: { slug: item.slug } }"
            class="block bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 hover:shadow-md transition-shadow"
          >
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400">
              {{ item.title }}
            </h3>
            <p v-if="item.summary" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-1">
              {{ item.summary }}
            </p>
            <div class="flex items-center gap-3 mt-2 text-[10px] text-gray-400 dark:text-gray-500">
              <span v-if="item.category_name">{{ item.category_name }}</span>
              <span v-if="item.author_name">{{ item.author_name }}</span>
              <span>{{ formatDate(item.created_at) }}</span>
            </div>
          </router-link>
        </div>
      </div>

      <!-- Events -->
      <div v-if="results.events.length > 0 && (activeType === '' || activeType === 'event')">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
          <Icon icon="mdi:calendar" class="h-5 w-5 text-green-500" />
          {{ $t('search.types.events') }}
          <span class="text-sm font-normal text-gray-500">({{ results.events.length }})</span>
        </h2>
        <div class="space-y-2">
          <router-link
            v-for="item in results.events"
            :key="'event-' + item.id"
            :to="{ name: 'EventDetail', params: { slug: item.slug } }"
            class="block bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 hover:shadow-md transition-shadow"
          >
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400">
              {{ item.title }}
            </h3>
            <div class="flex items-center gap-3 mt-2 text-xs text-gray-500 dark:text-gray-400">
              <span class="flex items-center gap-1">
                <Icon icon="mdi:calendar" class="h-3.5 w-3.5" />
                {{ formatDate(item.start_date) }}
              </span>
              <span v-if="item.location" class="flex items-center gap-1">
                <Icon icon="mdi:map-marker" class="h-3.5 w-3.5" />
                {{ item.location }}
              </span>
            </div>
          </router-link>
        </div>
      </div>

      <!-- Polls -->
      <div v-if="results.polls.length > 0 && (activeType === '' || activeType === 'poll')">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
          <Icon icon="mdi:poll" class="h-5 w-5 text-purple-500" />
          {{ $t('search.types.polls') }}
          <span class="text-sm font-normal text-gray-500">({{ results.polls.length }})</span>
        </h2>
        <div class="space-y-2">
          <router-link
            v-for="item in results.polls"
            :key="'poll-' + item.id"
            :to="{ name: 'PollDetail', params: { id: item.id } }"
            class="block bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 hover:shadow-md transition-shadow"
          >
            <div class="flex items-center gap-2">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white hover:text-primary-600 dark:hover:text-primary-400">
                {{ item.title }}
              </h3>
              <span
                :class="item.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'"
                class="px-2 py-0.5 text-[10px] font-medium rounded-full"
              >
                {{ item.is_active ? $t('search.active') : $t('search.closed') }}
              </span>
            </div>
            <p v-if="item.description" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-1">
              {{ item.description }}
            </p>
          </router-link>
        </div>
      </div>

      <!-- Announcements -->
      <div v-if="results.announcements.length > 0 && (activeType === '' || activeType === 'announcement')">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
          <Icon icon="mdi:bullhorn" class="h-5 w-5 text-orange-500" />
          {{ $t('search.types.announcements') }}
          <span class="text-sm font-normal text-gray-500">({{ results.announcements.length }})</span>
        </h2>
        <div class="space-y-2">
          <div
            v-for="item in results.announcements"
            :key="'ann-' + item.id"
            class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4"
          >
            <div class="flex items-center gap-2">
              <span
                :class="{
                  'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200': item.type === 'info',
                  'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200': item.type === 'warning',
                  'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200': item.type === 'success',
                  'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200': item.type === 'error',
                }"
                class="px-2 py-0.5 text-[10px] font-medium rounded-full"
              >
                {{ item.type }}
              </span>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
                {{ item.title }}
              </h3>
            </div>
            <p v-if="item.content" class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mt-1">
              {{ item.content }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- No Results -->
    <div v-else-if="results && results.total_count === 0 && searchQuery.length >= 2" class="text-center py-16">
      <Icon icon="mdi:magnify-close" class="h-16 w-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-1">
        {{ $t('search.noResults') }}
      </h3>
      <p class="text-sm text-gray-500 dark:text-gray-400">
        {{ $t('search.noResultsHint') }}
      </p>
    </div>

    <!-- Initial State -->
    <div v-else-if="!loading && searchQuery.length < 2" class="text-center py-16">
      <Icon icon="mdi:magnify" class="h-16 w-16 text-gray-300 dark:text-gray-600 mx-auto mb-4" />
      <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-1">
        {{ $t('search.initialHint') }}
      </h3>
      <p class="text-sm text-gray-500 dark:text-gray-400">
        {{ $t('search.initialHintSub') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { globalSearch } from '@/services/api'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const searchInput = ref(null)
const searchQuery = ref('')
const activeType = ref('')
const loading = ref(false)
const results = ref(null)
let debounceTimer = null

const tabs = computed(() => [
  { value: '', label: t('search.tabs.all'), icon: 'mdi:magnify' },
  { value: 'app', label: t('search.tabs.applications'), icon: 'mdi:application' },
  { value: 'news', label: t('search.tabs.news'), icon: 'mdi:newspaper' },
  { value: 'event', label: t('search.tabs.events'), icon: 'mdi:calendar' },
  { value: 'poll', label: t('search.tabs.polls'), icon: 'mdi:poll' },
  { value: 'announcement', label: t('search.tabs.announcements'), icon: 'mdi:bullhorn' },
])

const getCountForType = (type) => {
  if (!results.value) return 0
  const map = {
    app: 'applications',
    news: 'news',
    event: 'events',
    poll: 'polls',
    announcement: 'announcements',
  }
  return results.value[map[type]]?.length || 0
}

const debouncedSearch = () => {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    performSearch()
  }, 300)
}

const performSearch = async () => {
  if (searchQuery.value.length < 2) {
    results.value = null
    updateURL()
    return
  }

  loading.value = true
  try {
    const response = await globalSearch(searchQuery.value, activeType.value)
    results.value = response.data
  } catch (error) {
    console.error('Search error:', error)
    results.value = null
  } finally {
    loading.value = false
  }

  updateURL()
}

const setTypeFilter = (type) => {
  activeType.value = type
  if (searchQuery.value.length >= 2) {
    performSearch()
  }
}

const clearSearch = () => {
  searchQuery.value = ''
  results.value = null
  updateURL()
  searchInput.value?.focus()
}

const updateURL = () => {
  const query = {}
  if (searchQuery.value) query.q = searchQuery.value
  if (activeType.value) query.type = activeType.value
  router.replace({ query })
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  return new Date(dateString).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

// Init from URL params
onMounted(() => {
  if (route.query.q) {
    searchQuery.value = route.query.q
  }
  if (route.query.type) {
    activeType.value = route.query.type
  }
  if (searchQuery.value.length >= 2) {
    performSearch()
  }
  searchInput.value?.focus()
})

onUnmounted(() => {
  clearTimeout(debounceTimer)
})
</script>
