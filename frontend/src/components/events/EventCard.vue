<template>
  <div
    class="bg-white dark:bg-gray-800 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden hover:shadow-md transition-shadow cursor-pointer"
    @click="$emit('click', event)"
  >
    <!-- Event Image -->
    <div v-if="event.cover_image" class="aspect-video overflow-hidden">
      <img
        :src="event.cover_image"
        :alt="event.title"
        class="w-full h-full object-cover"
      />
    </div>
    
    <!-- Event Header -->
    <div class="p-4">
      <!-- Category Badge -->
      <div v-if="event.category" class="flex items-center justify-between mb-2">
        <span
          class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
          :class="getCategoryClass(event.category)"
        >
          {{ event.category.name }}
        </span>
        
        <!-- Event Status Badges -->
        <div class="flex gap-1">
          <span
            v-if="event.is_all_day"
            class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200"
          >
            {{ t('events.badge.allDay') }}
          </span>
          <span
            v-if="event.is_recurring"
            class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
          >
            {{ t('events.badge.recurring') }}
          </span>
          <span
            v-if="isPastEvent"
            class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300"
          >
            {{ t('events.badge.past') }}
          </span>
        </div>
      </div>

      <!-- Event Title -->
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2 line-clamp-2">
        {{ event.title }}
      </h3>

      <!-- Event Description -->
      <p v-if="event.description" class="text-gray-600 dark:text-gray-400 text-sm mb-3 line-clamp-2">
        {{ extractTextFromTiptap(event.description) }}
      </p>

      <!-- Event Date & Time -->
      <div class="flex items-center text-sm text-gray-500 dark:text-gray-400 mb-2">
        <Icon icon="mdi:calendar" class="h-4 w-4 mr-1" />
        <span>{{ formatDateRange(event.start_date, event.end_date) }}</span>
      </div>

      <!-- Event Time -->
      <div v-if="!event.is_all_day" class="flex items-center text-sm text-gray-500 dark:text-gray-400 mb-2">
        <Icon icon="mdi:clock" class="h-4 w-4 mr-1" />
        <span>{{ formatTimeRange(event.start_date, event.end_date) }}</span>
      </div>

      <!-- Event Location -->
      <div v-if="event.location" class="flex items-center text-sm text-gray-500 dark:text-gray-400 mb-3">
        <Icon icon="mdi:map-marker" class="h-4 w-4 mr-1" />
        <span class="line-clamp-1">{{ event.location }}</span>
      </div>

      <!-- Event Footer -->
      <div class="flex items-center justify-between pt-3 border-t border-gray-200 dark:border-gray-700">
        <div class="flex items-center text-xs text-gray-500 dark:text-gray-400">
          <Icon icon="mdi:account" class="h-3 w-3 mr-1" />
          <span>{{ event.author?.first_name }} {{ event.author?.last_name }}</span>
        </div>
        
        <div class="flex items-center gap-2">
          <!-- View Count -->
          <div v-if="event.views" class="flex items-center text-xs text-gray-500 dark:text-gray-400">
            <Icon icon="mdi:eye" class="h-3 w-3 mr-1" />
            <span>{{ event.views }}</span>
          </div>
          
          <!-- View Icon (Clickable) -->
          <button
            @click.stop="viewEvent"
            class="flex items-center text-xs text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300 transition-colors"
            :title="$t('events.actions.view')"
          >
            <Icon icon="mdi:eye" class="h-3 w-3 mr-1" />
            <span class="hidden sm:inline">{{ $t('events.actions.view') }}</span>
          </button>
          
          <!-- Published Status -->
          <span
            v-if="!event.is_published"
            class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200"
          >
            {{ t('events.status.draft') }}
          </span>
          <span
            v-else-if="event.published_at"
            class="text-xs text-gray-500 dark:text-gray-400"
          >
            {{ t('events.status.published') }} {{ formatRelativeTime(event.published_at) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'

const { t, locale } = useI18n()
const router = useRouter()

const props = defineProps({
  event: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['click'])

// Computed
const isPastEvent = computed(() => {
  const eventDate = new Date(props.event.start_date)
  const now = new Date()
  return eventDate < now
})

// Methods
const formatDateRange = (startDate, endDate) => {
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  
  const options = { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  }
  
  const currentLocale = getCurrentLocale()
  
  if (end && start.toDateString() !== end.toDateString()) {
    return `${start.toLocaleDateString(currentLocale, options)} - ${end.toLocaleDateString(currentLocale, options)}`
  }
  
  return start.toLocaleDateString(currentLocale, options)
}

const formatTimeRange = (startDate, endDate) => {
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  
  const timeOptions = { 
    hour: '2-digit', 
    minute: '2-digit' 
  }
  
  const currentLocale = getCurrentLocale()
  
  if (end) {
    return `${start.toLocaleTimeString(currentLocale, timeOptions)} - ${end.toLocaleTimeString(currentLocale, timeOptions)}`
  }
  
  return start.toLocaleTimeString(currentLocale, timeOptions)
}

const formatRelativeTime = (dateString) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffInSeconds = Math.floor((now - date) / 1000)
  
  if (diffInSeconds < 60) return t('dashboard.time.justNow')
  if (diffInSeconds < 3600) return t('dashboard.time.minutesAgo', { count: Math.floor(diffInSeconds / 60) })
  if (diffInSeconds < 86400) return t('dashboard.time.hoursAgo', { count: Math.floor(diffInSeconds / 3600) })
  if (diffInSeconds < 2592000) return t('dashboard.time.daysAgo', { count: Math.floor(diffInSeconds / 86400) })
  
  const currentLocale = getCurrentLocale()
  return date.toLocaleDateString(currentLocale, { year: 'numeric', month: 'short', day: 'numeric' })
}

const getCurrentLocale = () => {
  // Map locale codes to locale strings for date formatting
  const localeMap = {
    'fr': 'fr-FR',
    'en': 'en-US',
    'ar': 'ar-SA',
    'es': 'es-ES'
  }
  
  return localeMap[locale.value] || 'fr-FR'
}

// Utility function to extract text from Tiptap JSON format
const extractTextFromTiptap = (content) => {
  if (!content) return ''
  
  let contentObj = content
  
  // If content is a string, try to parse it as JSON
  if (typeof content === 'string') {
    try {
      contentObj = JSON.parse(content)
    } catch (e) {
      // If parsing fails, try to strip HTML tags
      return stripHtml(content)
    }
  }
  
  // Extract text from Tiptap JSON structure
  const extractText = (node) => {
    if (!node) return ''
    
    let text = ''
    
    if (node.type === 'text' && node.text) {
      text += node.text
    }
    
    if (node.content && Array.isArray(node.content)) {
      for (const child of node.content) {
        text += extractText(child)
      }
    }
    
    return text
  }
  
  return extractText(contentObj)
}

const stripHtml = (html) => {
  // Use DOMParser instead of innerHTML for security
  const parser = new DOMParser()
  const doc = parser.parseFromString(html, 'text/html')
  return doc.body.textContent || doc.body.innerText || ''
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

// Navigation method
const viewEvent = () => {
  if (!props.event.slug) {
    console.error('Événement sans slug:', props.event)
    return
  }

  router.push(`/events/${props.event.slug}`)
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

.line-clamp-2 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}
</style>