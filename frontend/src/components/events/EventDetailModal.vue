<template>
  <div v-if="event" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="$emit('close')">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto" @click.stop>
      <!-- Header -->
      <div class="flex items-start justify-between p-6 border-b border-gray-200 dark:border-gray-700">
        <div class="flex-1">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-2">
            {{ event.title }}
          </h2>
          <div class="flex items-center gap-4 text-sm text-gray-600 dark:text-gray-400">
            <!-- Date and time -->
            <div class="flex items-center gap-1">
              <Icon icon="mdi:calendar" class="h-4 w-4" />
              <span>{{ formatDateRange(event.start_date, event.end_date) }}</span>
            </div>
            
            <!-- Location -->
            <div v-if="event.location" class="flex items-center gap-1">
              <Icon icon="mdi:map-marker" class="h-4 w-4" />
              <span>{{ event.location }}</span>
            </div>
          </div>
        </div>
        
        <!-- Status badges -->
        <div class="flex items-center gap-2 ml-4">
          <span 
            v-if="event.is_all_day"
            class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200 rounded-full"
          >
            {{ $t('events.types.allDay') }}
          </span>
          <span 
            :class="getPriorityClass(event.priority)"
            class="px-2 py-1 text-xs font-medium rounded-full"
          >
            {{ $t(`events.priority.${event.priority}`) }}
          </span>
          <span 
            :class="getStatusClass(event.status)"
            class="px-2 py-1 text-xs font-medium rounded-full"
          >
            {{ $t(`events.status.${event.status}`) }}
          </span>
        </div>
      </div>

      <!-- Content -->
      <div class="p-6">
        <!-- Description -->
        <div v-if="event.description" class="mb-6">
          <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">
            {{ $t('events.detail.description') }}
          </h3>
          <div class="prose prose-sm max-w-none dark:prose-invert">
            <TiptapRenderer :content="event.description" />
          </div>
        </div>

        <!-- Category -->
        <div v-if="event.category" class="mb-6">
          <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">
            {{ $t('events.detail.category') }}
          </h3>
          <div class="flex items-center gap-2">
            <Icon :icon="event.category.icon || 'mdi:calendar'" class="h-4 w-4" :style="{ color: event.category.color }" />
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ event.category.name }}
            </span>
          </div>
        </div>

        <!-- External links -->
        <div v-if="event.external_links" class="mb-6">
          <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">
            {{ $t('events.detail.links') }}
          </h3>
          <div class="space-y-2">
            <a 
              v-for="(link, index) in parseExternalLinks(event.external_links)" 
              :key="index"
              :href="link.url"
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center gap-2 text-sm text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
            >
              <Icon :icon="link.icon || 'mdi:link'" class="h-4 w-4" />
              <span>{{ link.title }}</span>
              <Icon icon="mdi:external-link" class="h-3 w-3" />
            </a>
          </div>
        </div>

        <!-- Author info -->
        <div v-if="event.author" class="mb-6">
          <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">
            {{ $t('events.detail.organizer') }}
          </h3>
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 bg-gray-300 dark:bg-gray-600 rounded-full flex items-center justify-center">
              <Icon icon="mdi:account" class="h-4 w-4 text-gray-600 dark:text-gray-300" />
            </div>
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white">
                {{ event.author.first_name }} {{ event.author.last_name }}
              </p>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                {{ event.author.email }}
              </p>
            </div>
          </div>
        </div>

        <!-- Recurring info -->
        <div v-if="event.is_recurring" class="mb-6">
          <h3 class="text-sm font-medium text-gray-900 dark:text-white mb-2">
            {{ $t('events.detail.recurring') }}
          </h3>
          <div class="text-sm text-gray-700 dark:text-gray-300">
            <div class="flex items-center gap-1 mb-1">
              <Icon icon="mdi:repeat" class="h-4 w-4" />
              <span>{{ formatRecurrenceRule(event.recurrence_rule) }}</span>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ $t('events.detail.recurringDescription') }}
            </p>
          </div>
        </div>

        <!-- Publication info -->
        <div v-if="!event.is_published" class="mb-6">
          <div class="bg-yellow-50 dark:bg-yellow-900 border border-yellow-200 dark:border-yellow-700 rounded-lg p-3">
            <div class="flex items-center gap-2">
              <Icon icon="mdi:eye-off" class="h-4 w-4 text-yellow-600 dark:text-yellow-400" />
              <span class="text-sm font-medium text-yellow-800 dark:text-yellow-200">
                {{ $t('events.detail.draft') }}
              </span>
            </div>
            <p class="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
              {{ $t('events.detail.draftDescription') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between px-6 py-4 bg-gray-50 dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700">
        <div class="flex items-center gap-2">
          <!-- Color indicator -->
          <div 
            class="w-4 h-4 rounded-full" 
            :style="{ backgroundColor: event.color || '#3B82F6' }"
          ></div>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            {{ $t('events.detail.created') }} {{ formatDate(event.created_at) }}
          </span>
        </div>
        
        <div class="flex items-center gap-2">
          <!-- Edit button for authorized users -->
          <button
            v-if="canEdit"
            @click="editEvent"
            class="px-3 py-1 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
          >
            <Icon icon="mdi:pencil" class="h-4 w-4 mr-1 inline" />
            {{ $t('common.edit') }}
          </button>
          
          <!-- Close button -->
          <button
            @click="$emit('close')"
            class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-600 dark:hover:bg-gray-700"
          >
            {{ $t('common.close') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import TiptapRenderer from '@/components/news/TiptapRenderer.vue'

const props = defineProps({
  event: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'updated'])

const router = useRouter()
const authStore = useAuthStore()

// Computed
const canEdit = computed(() => {
  const userRole = authStore.user?.role
  const userId = authStore.user?.id
  return userRole === 'admin' || 
         userRole === 'editor' || 
         (userRole === 'group_admin' && props.event.author_id === userId)
})

// Methods
const formatDateRange = (startDate, endDate) => {
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  
  const formatOptions = {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }
  
  if (end && start.toDateString() === end.toDateString()) {
    return `${start.toLocaleDateString('fr-FR', { ...formatOptions, hour: undefined, minute: undefined })} - ${end.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}`
  } else if (end) {
    return `${start.toLocaleDateString('fr-FR', formatOptions)} - ${end.toLocaleDateString('fr-FR', formatOptions)}`
  } else {
    return start.toLocaleDateString('fr-FR', formatOptions)
  }
}

const formatDate = (date) => {
  return new Date(date).toLocaleDateString('fr-FR', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const parseExternalLinks = (linksJson) => {
  try {
    const links = JSON.parse(linksJson)
    return Array.isArray(links) ? links : []
  } catch (e) {
    return []
  }
}

const formatRecurrenceRule = (ruleJson) => {
  try {
    const rule = JSON.parse(ruleJson)
    switch (rule.type) {
      case 'daily':
        return 'Tous les jours'
      case 'weekly':
        const days = ['Dimanche', 'Lundi', 'Mardi', 'Mercredi', 'Jeudi', 'Vendredi', 'Samedi']
        const dayNames = rule.days_of_week?.map(day => days[day]).join(', ') || 'Semanalement'
        return `Hebdomadaire (${dayNames})`
      case 'monthly':
        return 'Mensuel'
      case 'yearly':
        return 'Annuel'
      default:
        return 'Récurrent'
    }
  } catch (e) {
    return 'Récurrent'
  }
}

const getPriorityClass = (priority) => {
  const classes = {
    'urgent': 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
    'high': 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
    'normal': 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
    'low': 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  }
  return classes[priority] || classes.normal
}

const getStatusClass = (status) => {
  const classes = {
    'confirmed': 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
    'tentative': 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
    'cancelled': 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
  }
  return classes[status] || classes.confirmed
}

const editEvent = () => {
  const userRole = authStore.user?.role
  const editRoute = userRole === 'admin' || userRole === 'editor' ? 'AdminEventEdit' : 'GroupAdminEventEdit'

  router.push({
    name: editRoute,
    params: { slug: props.event.slug }
  })

  emit('close')
}
</script>

<style scoped>
.prose {
  max-width: none;
}

.prose p {
  margin-bottom: 0.5rem;
}
</style>