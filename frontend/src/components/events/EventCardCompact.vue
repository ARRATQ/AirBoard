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
              class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
            >
              {{ t('events.badge.allDayShort') }}
            </span>
            <span
              v-if="event.is_recurring"
              class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
            >
              {{ t('events.badge.recurringShort') }}
            </span>
            <span
              v-if="isPastEvent"
              class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
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
            class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium"
            :class="getCategoryClass(event.category)"
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

const getCategoryClass = (category) => {
  const colors = {
    blue: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    green: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    yellow: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    red: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    purple: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200',
    indigo: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200',
    pink: 'bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-200',
    gray: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'
  }
  
  return colors[category.color] || colors.gray
}
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