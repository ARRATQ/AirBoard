<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-4">
    <div class="flex items-start gap-3">
      <!-- Event Icon/Color -->
      <div 
        class="flex-shrink-0 w-10 h-10 rounded-lg flex items-center justify-center"
        :class="getCategoryColor(event.category)"
      >
        <Icon icon="mdi:calendar" class="h-5 w-5 text-white" />
      </div>

      <!-- Event Content -->
      <div class="flex-1 min-w-0">
        <div class="flex items-start justify-between gap-2">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white line-clamp-1">
            {{ event.title }}
          </h3>
          
          <!-- Status Badges -->
          <div class="flex gap-1 flex-shrink-0">
            <span
              v-if="event.is_all_day"
              class="tag tag--info"
            >
              {{ t('events.badge.allDayShort') }}
            </span>
            <span
              v-if="event.is_recurring"
              class="tag tag--ok"
            >
              {{ t('events.badge.recurringShort') }}
            </span>
            <span
              v-if="isPastEvent"
              class="tag"
            >
              {{ t('events.badge.pastShort') }}
            </span>
          </div>
        </div>

        <!-- Date and Time -->
        <div class="flex items-center gap-4 mt-1 text-xs text-gray-500 dark:text-gray-400">
          <div class="flex items-center gap-1">
            <Icon icon="mdi:calendar" class="h-3 w-3" />
            <span>{{ formatDate(event.start_date) }}</span>
          </div>
          
          <div v-if="!event.is_all_day" class="flex items-center gap-1">
            <Icon icon="mdi:clock" class="h-3 w-3" />
            <span>{{ formatTime(event.start_date) }}</span>
          </div>
          
          <div v-if="event.location" class="flex items-center gap-1">
            <Icon icon="mdi:map-marker" class="h-3 w-3" />
            <span class="line-clamp-1">{{ event.location }}</span>
          </div>
        </div>

        <!-- Category -->
        <div v-if="event.category" class="mt-2">
          <span
            class="tag tag--custom"
            :style="{ '--tag-c': getCategoryTagColor(event.category) }"
          >
            {{ event.category.name }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'

const { t } = useI18n()

const props = defineProps({
  event: {
    type: Object,
    required: true
  }
})

// Computed
const isPastEvent = computed(() => {
  const eventDate = new Date(props.event.start_date)
  const now = new Date()
  return eventDate < now
})

// Methods
const formatDate = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric'
  })
}

const formatTime = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getCategoryColor = (category) => {
  const colors = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    yellow: 'bg-yellow-500',
    red: 'bg-red-500',
    purple: 'bg-purple-500',
    indigo: 'bg-indigo-500',
    pink: 'bg-pink-500',
    gray: 'bg-gray-500'
  }
  
  return colors[category.color] || colors.gray
}

// Couleur de l'étiquette de catégorie (.tag--custom) : nom de couleur historique ou hex
const CATEGORY_COLORS = {
  blue: '#2563eb', green: '#16a34a', yellow: '#ca8a04', red: '#dc2626',
  purple: '#9333ea', indigo: '#4f46e5', pink: '#db2777', gray: 'var(--muted)',
}
const getCategoryTagColor = (category) =>
  CATEGORY_COLORS[category.color] || (/^#[0-9a-f]{3,8}$/i.test(category.color || '') ? category.color : CATEGORY_COLORS.gray)
</script>

<style scoped>
.line-clamp-1 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  line-clamp: 1;
}
</style>