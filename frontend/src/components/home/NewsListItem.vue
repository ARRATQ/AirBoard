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
      <div class="flex items-center gap-1.5">
        <h4 class="font-semibold text-xs text-gray-900 dark:text-white truncate">
          {{ article.title }}
        </h4>
        <!-- Badge épinglé -->
        <div v-if="article.is_pinned" class="badge-pinned" :title="t('news.detail.pinned')">
          <Icon icon="mdi:pin" class="h-3 w-3" />
        </div>
      </div>
      <p v-if="article.summary" class="text-xs text-gray-600 dark:text-gray-400 truncate mt-0.5">
        {{ article.summary }}
      </p>
      <div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-500 mt-0.5">
        <Icon icon="mdi:calendar" class="h-3 w-3" />
        <span>{{ formatDate(article.published_at || article.created_at) }}</span>

        <!-- Badge groupe cible (si l'article est destiné à un groupe spécifique) -->
        <div v-if="hasTargetGroups" class="badge-group" :title="targetGroupsTooltip">
          <Icon icon="mdi:account-group" class="h-3 w-3" />
          <span>{{ firstTargetGroupName }}</span>
        </div>

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
import { computed } from 'vue'

const { t } = useI18n()

const props = defineProps({
  article: {
    type: Object,
    required: true
  }
})

defineEmits(['click'])

// Computed pour vérifier si l'article a des groupes cibles
const hasTargetGroups = computed(() => {
  return props.article.target_groups && props.article.target_groups.length > 0
})

// Nom du premier groupe cible (pour affichage compact)
const firstTargetGroupName = computed(() => {
  if (!hasTargetGroups.value) return ''
  const groups = props.article.target_groups
  if (groups.length === 1) {
    return groups[0].name
  }
  return `${groups[0].name} +${groups.length - 1}`
})

// Tooltip avec tous les groupes cibles
const targetGroupsTooltip = computed(() => {
  if (!hasTargetGroups.value) return ''
  return props.article.target_groups.map(g => g.name).join(', ')
})

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

/* Badge Pinned (épinglé) */
.badge-pinned {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.125rem;
  background-color: rgb(254 243 199);
  border-radius: 0.25rem;
  color: rgb(180 83 9);
  flex-shrink: 0;
}

.dark .badge-pinned {
  background-color: rgb(120 53 15);
  color: rgb(253 230 138);
}

/* Badge Group (groupe cible) */
.badge-group {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.125rem 0.375rem;
  background-color: rgb(219 234 254);
  border-radius: 0.375rem;
  color: rgb(29 78 216);
  font-weight: 500;
  font-size: 0.625rem;
  max-width: 100px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.badge-group span {
  overflow: hidden;
  text-overflow: ellipsis;
}

.dark .badge-group {
  background-color: rgb(30 58 138);
  color: rgb(147 197 253);
}

.news-list-item:hover .badge-group {
  background-color: rgb(191 219 254);
}

.dark .news-list-item:hover .badge-group {
  background-color: rgb(30 64 175);
}
</style>
