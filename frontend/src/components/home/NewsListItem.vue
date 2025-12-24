<template>
  <div class="news-list-item" @click="$emit('click', article)">
    <!-- News Icon -->
    <div
      class="h-8 w-8 rounded-md flex items-center justify-center flex-shrink-0 shadow-sm"
      :style="{ backgroundColor: article.category?.color || '#f97316' }"
    >
      <Icon :icon="article.category?.icon || 'mdi:newspaper'" class="h-3.5 w-3.5 text-white" />
    </div>

    <!-- News Info -->
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-1">
        <h4 class="font-semibold text-xs text-gray-900 dark:text-white truncate">
          {{ article.title }}
        </h4>
      </div>
      <p v-if="article.summary" class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">
        {{ article.summary }}
      </p>
      <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-500 mt-0.5">
        <Icon icon="mdi:calendar" class="h-3 w-3" />
        <span>{{ formatDate(article.published_at || article.created_at) }}</span>

        <!-- Badges pour commentaires et likes -->
        <div class="flex items-center gap-1.5 ml-auto">
          <div v-if="article.comment_count > 0" class="badge-stat">
            <Icon icon="mdi:comment" class="h-3 w-3" />
            <span>{{ article.comment_count }}</span>
          </div>
          <div v-if="article.reaction_count > 0" class="badge-stat">
            <Icon icon="mdi:heart" class="h-3 w-3" />
            <span>{{ article.reaction_count }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  article: {
    type: Object,
    required: true
  }
})

defineEmits(['click'])

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const now = new Date()
  const diffInMs = now - date
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))

  if (diffInDays === 0) return t('common.today')
  if (diffInDays === 1) return t('common.yesterday')
  if (diffInDays < 7) return `${t('common.daysAgo', { count: diffInDays })}`

  return date.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })
}
</script>

<style scoped>
/* List Item - Same style as NewAppsWidget */
.news-list-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.5rem;
  background-color: rgb(255 255 255);
  border: 1px solid rgb(229 231 235);
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.15s ease;
}

.news-list-item:hover {
  background-color: rgb(249 250 251);
  border-color: rgb(209 213 219);
  box-shadow: 0 2px 4px 0 rgb(0 0 0 / 0.05);
  transform: translateX(2px);
}

.dark .news-list-item {
  background-color: rgb(31 41 55);
  border-color: rgb(55 65 81);
}

.dark .news-list-item:hover {
  background-color: rgb(55 65 81);
  border-color: rgb(75 85 99);
}

/* Badge Stats */
.badge-stat {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.375rem;
  background-color: rgb(243 244 246);
  border-radius: 0.375rem;
  color: rgb(107 114 128);
  font-weight: 500;
  transition: all 0.2s ease;
}

.dark .badge-stat {
  background-color: rgb(55 65 81);
  color: rgb(156 163 175);
}

.news-list-item:hover .badge-stat {
  background-color: rgb(229 231 235);
}

.dark .news-list-item:hover .badge-stat {
  background-color: rgb(75 85 99);
}
</style>
