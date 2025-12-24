<template>
  <div class="event-item" @click="$emit('click', event)">
    <div class="flex items-start gap-3">
      <div class="event-date-badge">
        <div class="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase">
          {{ formatDay(event.start_date) }}
        </div>
        <div class="text-lg font-bold text-gray-900 dark:text-white">
          {{ formatDate(event.start_date) }}
        </div>
      </div>
      <div class="flex-1 min-w-0">
        <h4 class="text-sm font-semibold text-gray-900 dark:text-white truncate mb-1">
          {{ event.title }}
        </h4>
        <div class="flex items-center gap-2 text-xs text-gray-600 dark:text-gray-400 flex-wrap">
          <div class="flex items-center gap-1">
            <Icon v-if="!event.is_all_day" icon="mdi:clock" class="h-3 w-3" />
            <span>{{ event.is_all_day ? $t('events.allDay') : formatTime(event.start_date) }}</span>
          </div>
          <div v-if="event.location" class="flex items-center gap-1">
            <Icon icon="mdi:map-marker" class="h-3 w-3" />
            <span class="truncate">{{ event.location }}</span>
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
  event: {
    type: Object,
    required: true
  }
})

defineEmits(['click'])

const formatDay = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const days = ['SUN', 'MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT']
  return days[date.getDay()]
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.getDate()
}

const formatTime = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.event-item {
  @apply p-2 rounded-md hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors;
}

.event-date-badge {
  @apply flex flex-col items-center justify-center w-10 h-10 bg-blue-50 dark:bg-blue-900/20 rounded-md flex-shrink-0;
}
</style>
