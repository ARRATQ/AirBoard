<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ $t('events.admin.title') }}</h1>
          <p class="page-subtitle">{{ $t('events.admin.subtitle') }}</p>
        </div>
        <router-link to="/admin/events/new" class="btn btn-primary">
          <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
          {{ $t('events.admin.newEvent') }}
        </router-link>
      </div>
    </div>

    <!-- Filters -->
    <div class="card mb-6">
      <div class="flex flex-col sm:flex-row gap-4">
        <!-- Search -->
        <div class="flex-1 min-w-[200px]">
          <input
            v-model="filters.search"
            type="text"
            :placeholder="$t('events.admin.searchPlaceholder')"
            class="input"
          />
        </div>

        <!-- Category filter -->
        <select v-model="filters.category_id" class="input w-full sm:w-48">
          <option value="">{{ $t('events.admin.allCategories') }}</option>
          <option v-for="cat in categories" :key="cat.id" :value="cat.id">
            {{ cat.name }}
          </option>
        </select>

        <!-- Status filter -->
        <select v-model="filters.status" class="input w-full sm:w-40">
          <option value="">{{ $t('events.admin.allStatus') }}</option>
          <option value="published">{{ $t('events.admin.published') }}</option>
          <option value="draft">{{ $t('events.admin.draft') }}</option>
        </select>

        <!-- Event Type filter -->
        <select v-model="filters.event_type" class="input w-full sm:w-40">
          <option value="">{{ $t('events.admin.allTypes') }}</option>
          <option value="recurring">{{ $t('events.admin.recurring') }}</option>
          <option value="all_day">{{ $t('events.admin.allDay') }}</option>
          <option value="upcoming">{{ $t('events.admin.upcoming') }}</option>
          <option value="past">{{ $t('events.admin.past') }}</option>
        </select>
      </div>
    </div>

    <!-- Loading -->
    <div v-if="eventsStore.isLoading" class="flex justify-center py-12">
      <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-gray-400" />
    </div>

    <!-- Events List -->
    <div v-else-if="eventsList.length > 0" class="space-y-4">
      <template v-for="event in eventsList" :key="event.id">
        <div
          v-if="event && event.id"
          class="card hover:shadow-lg transition-shadow"
        >
        <div class="flex items-start gap-4">
          <!-- Status indicator -->
          <div class="flex-shrink-0 pt-1">
            <div
              :class="event.is_published ? 'bg-green-500' : 'bg-gray-400'"
              class="h-3 w-3 rounded-full"
              :title="event.is_published ? $t('events.admin.published') : $t('events.admin.draft')"
            ></div>
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1">
                <div class="flex items-center gap-2 mb-1">
                  <!-- Recurring indicator -->
                  <Icon
                    v-if="event.is_recurring"
                    icon="mdi:repeat"
                    class="h-4 w-4 text-green-500"
                  />

                  <!-- All day indicator -->
                  <Icon
                    v-if="event.is_all_day"
                    icon="mdi:weather-sunny"
                    class="h-4 w-4 text-blue-500"
                  />

                  <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    {{ event.title }}
                  </h3>

                  <!-- Priority badge -->
                  <span
                    v-if="event.priority !== 'normal'"
                    :class="{
                      'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200': event.priority === 'urgent',
                      'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200': event.priority === 'important'
                    }"
                    class="px-2 py-0.5 rounded text-xs font-medium"
                  >
                    {{ $t(`events.priority.${event.priority}`) }}
                  </span>
                </div>

                <p v-if="event.description" class="text-sm text-gray-600 dark:text-gray-400 mb-2 line-clamp-2">
                  {{ extractTextFromTiptap(event.description) }}
                </p>

                <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                  <!-- Category -->
                  <div v-if="event.category" class="flex items-center gap-1">
                    <Icon :icon="event.category.icon || 'mdi:folder'" class="h-4 w-4" />
                    <span>{{ event.category.name }}</span>
                  </div>

                  <!-- Author -->
                  <div class="flex items-center gap-1">
                    <Icon icon="mdi:account" class="h-4 w-4" />
                    <span>{{ event.author?.first_name || '' }} {{ event.author?.last_name || '' }}</span>
                  </div>

                  <!-- Date Range -->
                  <div class="flex items-center gap-1">
                    <Icon icon="mdi:calendar" class="h-4 w-4" />
                    <span>{{ formatDateRange(event.start_date, event.end_date) }}</span>
                  </div>

                  <!-- Time -->
                  <div v-if="!event.is_all_day" class="flex items-center gap-1">
                    <Icon icon="mdi:clock" class="h-4 w-4" />
                    <span>{{ formatTimeRange(event.start_date, event.end_date) }}</span>
                  </div>

                  <!-- Location -->
                  <div v-if="event.location" class="flex items-center gap-1">
                    <Icon icon="mdi:map-marker" class="h-4 w-4" />
                    <span class="line-clamp-1">{{ event.location }}</span>
                  </div>

                  <!-- Views -->
                  <div v-if="event.views" class="flex items-center gap-1">
                    <Icon icon="mdi:eye" class="h-4 w-4" />
                    <span>{{ event.views }}</span>
                  </div>

                  <!-- Visibility -->
                  <div class="flex items-center gap-1">
                    <Icon :icon="getVisibilityIcon(event.target_groups)" class="h-4 w-4" />
                    <span>{{ getVisibilityLabel(event.target_groups) }}</span>
                  </div>
                </div>
              </div>

              <!-- Actions -->
              <div class="flex items-center gap-2">
                <router-link
                  :to="`/admin/events/${event.slug}/edit`"
                  class="btn btn-secondary btn-sm"
                  :title="$t('events.admin.edit')"
                >
                  <Icon icon="mdi:pencil" class="h-4 w-4" />
                </router-link>

                <button
                  @click="togglePublish(event)"
                  class="btn btn-secondary btn-sm"
                  :title="event.is_published ? $t('events.admin.unpublish') : $t('events.admin.publish')"
                >
                  <Icon :icon="event.is_published ? 'mdi:eye-off' : 'mdi:eye'" class="h-4 w-4" />
                </button>

                <button
                  @click="confirmDelete(event)"
                  class="btn btn-danger btn-sm"
                  :title="$t('events.admin.delete')"
                >
                  <Icon icon="mdi:delete" class="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
        </div>
      </template>

      <!-- Pagination -->
      <div v-if="pagination.pages > 1" class="flex justify-center gap-2 mt-6">
        <button
          @click="changePage(pagination.page - 1)"
          :disabled="pagination.page === 1"
          class="btn btn-secondary btn-sm"
        >
          <Icon icon="mdi:chevron-left" class="h-4 w-4" />
        </button>

        <span class="px-4 py-2 text-sm text-gray-700 dark:text-gray-300">
          {{ $t('events.admin.pageInfo', { page: pagination.page, pages: pagination.pages }) }}
        </span>

        <button
          @click="changePage(pagination.page + 1)"
          :disabled="pagination.page === pagination.pages"
          class="btn btn-secondary btn-sm"
        >
          <Icon icon="mdi:chevron-right" class="h-4 w-4" />
        </button>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="empty-state">
      <Icon icon="mdi:calendar" class="empty-state-icon" />
      <h3 class="empty-state-title">{{ $t('events.admin.noEvents') }}</h3>
      <p class="empty-state-description">
        {{ $t('events.admin.noEventsHelp') }}
      </p>
      <router-link to="/admin/events/new" class="btn btn-primary mt-4">
        <Icon icon="mdi:plus" class="h-4 w-4 mr-2" />
        {{ $t('events.admin.createEvent') }}
      </router-link>
    </div>

    <!-- Delete Confirmation Modal -->
    <div
      v-if="eventToDelete"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="eventToDelete = null"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full mx-4">
        <h3 class="text-lg font-semibold mb-4 text-gray-900 dark:text-white">
          {{ $t('events.admin.deleteConfirmTitle') }}
        </h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {{ $t('events.admin.deleteConfirmMessage', { title: eventToDelete.title }) }}
        </p>
        <div class="flex justify-end gap-3">
          <button @click="eventToDelete = null" class="btn btn-secondary">
            {{ $t('common.cancel') }}
          </button>
          <button @click="deleteEvent" class="btn btn-danger">
            {{ $t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, getCurrentInstance } from 'vue'
import { Icon } from '@iconify/vue'
import { useEventsStore } from '@/stores/events'
import { useAppStore } from '@/stores/app'

const { proxy } = getCurrentInstance()
const eventsStore = useEventsStore()
const appStore = useAppStore()

const eventToDelete = ref(null)

const filters = ref({
  search: '',
  category_id: '',
  status: '',
  event_type: '',
  page: 1,
  limit: 10
})

// Computed
const eventsList = computed(() => eventsStore.events?.filter(event => event && event.id) || [])
const pagination = computed(() => eventsStore.pagination || { page: 1, pages: 1 })
const categories = computed(() => eventsStore.categories || [])

// Methods
const loadEvents = async () => {
  try {
    const params = {
      page: filters.value.page,
      limit: filters.value.limit,
      ...filters.value
    }

    // Remove empty filters
    Object.keys(params).forEach(key => {
      if (!params[key]) {
        delete params[key]
      }
    })

    await eventsStore.fetchAdminEvents(params)
  } catch (error) {
    console.error('Error loading events:', error)
    appStore.showError(proxy.$t('events.admin.loadError'))
  }
}

const loadCategories = async () => {
  try {
    await eventsStore.fetchCategories()
  } catch (error) {
    console.error('Error loading categories:', error)
  }
}

const togglePublish = async (event) => {
  try {
    const updatedEvent = { ...event, is_published: !event.is_published }
    await eventsStore.updateEvent(event.slug, updatedEvent)
    appStore.showSuccess(
      event.is_published 
        ? proxy.$t('events.admin.unpublishedSuccess') 
        : proxy.$t('events.admin.publishedSuccess')
    )
  } catch (error) {
    console.error('Error toggling publish status:', error)
    appStore.showError(proxy.$t('events.admin.updateError'))
  }
}

const confirmDelete = (event) => {
  eventToDelete.value = event
}

const deleteEvent = async () => {
  try {
    await eventsStore.deleteEvent(eventToDelete.value.id)
    appStore.showSuccess(proxy.$t('events.admin.deleteSuccess'))
    eventToDelete.value = null
    loadEvents()
  } catch (error) {
    console.error('Error deleting event:', error)
    appStore.showError(proxy.$t('events.admin.deleteError'))
  }
}

const changePage = (page) => {
  if (page >= 1 && page <= pagination.value.pages) {
    filters.value.page = page
  }
}

const formatDateRange = (startDate, endDate) => {
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  
  const options = { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric' 
  }
  
  if (end && start.toDateString() !== end.toDateString()) {
    return `${start.toLocaleDateString('fr-FR', options)} - ${end.toLocaleDateString('fr-FR', options)}`
  }
  
  return start.toLocaleDateString('fr-FR', options)
}

const formatTimeRange = (startDate, endDate) => {
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  
  const timeOptions = { 
    hour: '2-digit', 
    minute: '2-digit' 
  }
  
  if (end) {
    return `${start.toLocaleTimeString('fr-FR', timeOptions)} - ${end.toLocaleTimeString('fr-FR', timeOptions)}`
  }
  
  return start.toLocaleTimeString('fr-FR', timeOptions)
}

const extractTextFromTiptap = (content) => {
  if (!content) return ''

  let contentObj = content

  // If content is a string, try to parse it as JSON
  if (typeof content === 'string') {
    try {
      contentObj = JSON.parse(content)
    } catch (e) {
      // If parsing fails, return the content as is (plain text)
      return content
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
        // Add space between paragraphs
        if (child.type === 'paragraph') {
          text += ' '
        }
      }
    }

    return text
  }

  return extractText(contentObj).trim()
}

const getVisibilityIcon = (targetGroups) => {
  if (!targetGroups || targetGroups.length === 0) {
    return 'mdi:earth'
  }
  return 'mdi:lock'
}

const getVisibilityLabel = (targetGroups) => {
  if (!targetGroups || targetGroups.length === 0) {
    return 'Public'
  }
  if (targetGroups.length === 1) {
    return targetGroups[0].name
  }
  return `${targetGroups.length} groups`
}

// Watch filters
watch(
  () => [filters.value.search, filters.value.category_id, filters.value.status, filters.value.event_type, filters.value.page],
  () => {
    loadEvents()
  }
)

// Lifecycle
onMounted(() => {
  loadEvents()
  loadCategories()
})
</script>

<style scoped>
.line-clamp-2 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.line-clamp-1 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  line-clamp: 1;
}
</style>