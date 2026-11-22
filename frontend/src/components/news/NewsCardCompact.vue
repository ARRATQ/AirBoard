<template>
  <router-link
    :to="{ name: 'NewsDetail', params: { slug: news.slug } }"
    class="group block cursor-pointer p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-primary-300 dark:hover:border-primary-700 hover:shadow-md transition-all duration-200 bg-white dark:bg-gray-800"
  >
    <div class="flex items-center gap-3">
      <!-- Priority & Pin Indicator -->
      <div class="flex-shrink-0 flex flex-col items-center gap-1">
        <span
          v-if="news.priority === 'urgent'"
          class="block w-2 h-2 rounded-full bg-red-500 animate-pulse"
          role="img"
          :title="t('news.priority.urgent')"
          :aria-label="t('news.priority.urgent')"
        ></span>
        <span
          v-else-if="news.priority === 'important'"
          class="block w-2 h-2 rounded-full bg-orange-500"
          role="img"
          :title="t('news.priority.important')"
          :aria-label="t('news.priority.important')"
        ></span>
        <Icon
          v-if="news.is_pinned"
          icon="lucide:pin"
          class="h-3 w-3 text-yellow-500"
          role="img"
          :title="t('news.detail.pinned')"
          :aria-label="t('news.detail.pinned')"
        />
      </div>

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <div class="flex items-start justify-between gap-2 mb-1">
          <h3 class="font-medium text-sm text-gray-900 dark:text-white line-clamp-1 flex-1">
            {{ news.title }}
          </h3>
          <span
            :class="[
              'tag',
              getTypeBadgeClass(news.type)
            ]"
          >
            {{ $t(`news.types.${news.type}`) }}
          </span>
        </div>

        <div class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
          <span class="flex items-center gap-1">
            <Icon icon="lucide:calendar" class="h-3 w-3" />
            {{ formatDate(news.published_at || news.created_at) }}
          </span>

          <span v-if="news.category" class="flex items-center gap-1">
            <Icon :icon="news.category.icon || 'lucide:folder'" class="h-3 w-3" :style="{ color: news.category.color }" />
            {{ news.category.name }}
          </span>

          <span class="flex items-center gap-1 ms-auto">
            <Icon icon="lucide:eye" class="h-3 w-3" />
            {{ news.views || 0 }}
          </span>

          <span v-if="news.reactions && news.reactions.length > 0" class="flex items-center gap-1">
            <Icon icon="lucide:heart" class="h-3 w-3" />
            {{ news.reactions.length }}
          </span>
        </div>
      </div>
    </div>
  </router-link>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { dateLocaleFor } from '@/utils/homeFormat'

const { t, locale } = useI18n()

defineProps({
  news: {
    type: Object,
    required: true
  }
})

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
  const dateLocale = dateLocaleFor(locale.value)

  // Format court dans la langue de l'interface (« hier », « il y a 3 j »…)
  if (diffInDays < 7) {
    return new Intl.RelativeTimeFormat(dateLocale, { numeric: 'auto', style: 'short' }).format(-diffInDays, 'day')
  }

  return date.toLocaleDateString(dateLocale, { day: 'numeric', month: 'short' })
}
</script>
