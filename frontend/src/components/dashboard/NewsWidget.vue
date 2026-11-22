<template>
  <div class="card">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white flex items-center gap-2">
        <Icon icon="mdi:newspaper" class="h-5 w-5 text-primary-500" />
        {{ $t('news.widget.title') }}
        <span v-if="unreadCount > 0" class="ms-2 px-2 py-1 bg-red-500 text-white text-xs rounded-full">
          <span aria-hidden="true">{{ unreadCount }}</span>
          <span class="sr-only">{{ $t('news.widget.unreadCount', { count: unreadCount }, unreadCount) }}</span>
        </span>
      </h2>
      <router-link
        to="/news"
        class="text-sm text-primary-500 hover:text-primary-600 font-medium flex items-center gap-1"
      >
        {{ $t('news.widget.viewAll') }}
        <Icon icon="mdi:arrow-right" class="rtl-flip h-4 w-4" />
      </router-link>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center py-8" role="status">
      <Icon icon="mdi:loading" class="h-6 w-6 animate-spin text-primary-500" aria-hidden="true" />
      <span class="sr-only">{{ $t('common.loading') }}</span>
    </div>

    <!-- News List -->
    <div v-else-if="newsList.length > 0" class="space-y-3">
      <router-link
        v-for="news in newsList"
        :key="news.id"
        :to="{ name: 'NewsDetail', params: { slug: news.slug } }"
        class="group block p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-700 hover:shadow-sm transition-all"
      >
        <!-- Header -->
        <div class="flex items-start justify-between mb-2">
          <div class="flex items-center gap-2 flex-1 min-w-0">
            <!-- Priority Indicator -->
            <span
              v-if="news.priority === 'urgent'"
              class="flex-shrink-0 w-2 h-2 rounded-full bg-red-500 animate-pulse"
              role="img"
              :aria-label="$t('news.priority.urgent')"
            ></span>
            <span
              v-else-if="news.priority === 'important'"
              class="flex-shrink-0 w-2 h-2 rounded-full bg-orange-500"
              role="img"
              :aria-label="$t('news.priority.important')"
            ></span>

            <!-- Title -->
            <h3 class="font-medium text-gray-900 dark:text-white text-sm line-clamp-1 flex-1">
              {{ news.title }}
            </h3>

            <!-- Pinned Icon -->
            <Icon
              v-if="news.is_pinned"
              icon="mdi:pin"
              class="h-4 w-4 text-yellow-500 flex-shrink-0"
            />
          </div>

          <!-- Type Badge -->
          <span
            :class="[
              'tag ms-2',
              getTypeBadgeClass(news.type)
            ]"
          >
            {{ $t(`news.types.${news.type}`) }}
          </span>
        </div>

        <!-- Summary -->
        <p class="text-sm text-gray-600 dark:text-gray-400 line-clamp-2 mb-2">
          {{ news.summary }}
        </p>

        <!-- Footer -->
        <div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-500">
          <div class="flex items-center gap-3">
            <span class="flex items-center gap-1">
              <Icon icon="mdi:calendar" class="h-3 w-3" />
              {{ formatDate(news.published_at || news.created_at) }}
            </span>
            <span v-if="news.category" class="flex items-center gap-1">
              <Icon :icon="news.category.icon || 'mdi:folder'" class="h-3 w-3" :style="{ color: news.category.color }" />
              {{ news.category.name }}
            </span>
          </div>

          <!-- Reactions count -->
          <div v-if="news.reactions && news.reactions.length > 0" class="flex items-center gap-1">
            <Icon icon="mdi:heart" class="h-3 w-3" aria-hidden="true" />
            <span aria-hidden="true">{{ news.reactions.length }}</span>
            <span class="sr-only">{{ $t('news.widget.reactionsCount', { count: news.reactions.length }, news.reactions.length) }}</span>
          </div>
        </div>
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-8">
      <Icon icon="mdi:newspaper-variant-outline" class="h-12 w-12 mx-auto text-gray-400 mb-2" />
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ $t('news.widget.noNews') }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { newsService } from '@/services/api'
import { dateLocaleFor } from '@/utils/homeFormat'

const { t, locale } = useI18n()

// Data
const loading = ref(false)
const newsList = ref([])
const unreadCount = ref(0)

// Methods
const fetchTopNews = async () => {
  loading.value = true
  try {
    const response = await newsService.getNews({
      limit: 3,
      is_pinned: true // Get pinned news first, or latest
    })
    newsList.value = response.news || []
  } catch (error) {
    console.error('Error fetching top news:', error)
  } finally {
    loading.value = false
  }
}

const fetchUnreadCount = async () => {
  try {
    const response = await newsService.getUnreadCount()
    unreadCount.value = response.count || 0
  } catch (error) {
    console.error('Error fetching unread count:', error)
  }
}

// Ton de l'étiquette (.tag, main.css) selon le type historique de l'actualité
const getTypeBadgeClass = (type) => {
  const tones = { article: 'tag--info', tutorial: 'tag--ok', announcement: 'tag--acc', faq: 'tag--warn' }
  return tones[type] || tones.article
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const now = new Date()
  const diffInDays = Math.floor((now - date) / (1000 * 60 * 60 * 24))

  if (diffInDays === 0) return t('common.today')
  if (diffInDays === 1) return t('common.yesterday')
  if (diffInDays < 7) return t('common.daysAgo', { count: diffInDays })

  return date.toLocaleDateString(dateLocaleFor(locale.value), { day: 'numeric', month: 'short' })
}

// Lifecycle
onMounted(() => {
  fetchTopNews()
  fetchUnreadCount()
})
</script>
