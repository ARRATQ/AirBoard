<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ isEditMode ? $t('groupAdmin.eventEditor.title.edit') : $t('groupAdmin.eventEditor.title.new') }}</h1>
          <p class="page-subtitle">{{ isEditMode ? $t('groupAdmin.eventEditor.subtitle.edit') : $t('groupAdmin.eventEditor.subtitle.new') }}</p>
        </div>
        <router-link to="/group-admin/events" class="btn btn-secondary">
          <Icon icon="mdi:arrow-left" class="h-4 w-4 mr-2" />
          {{ $t('groupAdmin.eventEditor.backToList') }}
        </router-link>
      </div>
    </div>

    <!-- Form -->
    <form @submit.prevent="saveEvent" class="space-y-6">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Main content -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Title -->
          <div class="card">
            <div class="mb-3">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-2 flex items-center gap-2">
                <Icon icon="mdi:format-title" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.eventEditor.titleLabel') }} *
              </label>
              <input
                v-model="form.title"
                type="text"
                required
                :placeholder="$t('groupAdmin.eventEditor.titlePlaceholder')"
                class="w-full px-4 py-3 text-lg font-medium border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white transition-colors"
              />
            </div>
          </div>

          <!-- Description -->
          <div class="card">
            <div class="mb-3">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                <Icon icon="mdi:text-box" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.eventEditor.descriptionLabel') }}
              </label>
              <div class="border-2 border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
                <RichTextEditor
                  v-model="form.description"
                  :placeholder="$t('groupAdmin.eventEditor.descriptionPlaceholder')"
                  :as-group-admin="true"
                />
              </div>
            </div>
          </div>

          <!-- Date & Time -->
          <div class="card">
            <div class="mb-4">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
                <Icon icon="mdi:calendar-clock" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.eventEditor.dateTime') }}
              </label>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                <!-- Start Date & Time -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    {{ $t('groupAdmin.eventEditor.startDate') }} *
                  </label>
                  <input
                    v-model="form.start_date"
                    type="datetime-local"
                    required
                    class="w-full px-4 py-3 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                </div>

                <!-- End Date & Time -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    {{ $t('groupAdmin.eventEditor.endDate') }}
                  </label>
                  <input
                    v-model="form.end_date"
                    type="datetime-local"
                    class="w-full px-4 py-3 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                </div>
              </div>

              <!-- All Day Toggle -->
              <div class="bg-white dark:bg-gray-800 rounded-lg p-4 border-2 border-gray-200 dark:border-gray-700">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_all_day"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ml-3 text-sm font-medium text-gray-900 dark:text-white">
                    <Icon icon="mdi:weather-sunny" class="inline h-4 w-4 mr-1" />
                    {{ $t('groupAdmin.eventEditor.allDay') }}
                  </span>
                </label>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {{ $t('groupAdmin.eventEditor.allDayHelp') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Location -->
          <div class="card">
            <div class="mb-3">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-2 flex items-center gap-2">
                <Icon icon="mdi:map-marker" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.eventEditor.location') }}
              </label>
              <input
                v-model="form.location"
                type="text"
                :placeholder="$t('groupAdmin.eventEditor.locationPlaceholder')"
                class="w-full px-4 py-3 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
              />
            </div>
          </div>

          <!-- Recurrence -->
          <div class="card">
            <div class="mb-4">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
                <Icon icon="mdi:repeat" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.eventEditor.recurrence') }}
              </label>

              <!-- Recurring Toggle -->
              <div class="bg-white dark:bg-gray-800 rounded-lg p-4 border-2 border-gray-200 dark:border-gray-700 mb-4">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_recurring"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ml-3 text-sm font-medium text-gray-900 dark:text-white">
                    {{ $t('groupAdmin.eventEditor.isRecurring') }}
                  </span>
                </label>
              </div>

              <!-- Recurrence Pattern -->
              <div v-if="form.is_recurring" class="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <!-- Recurrence Type -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      {{ $t('groupAdmin.eventEditor.recurrenceType') }}
                    </label>
                    <select
                      v-model="recurrencePattern.type"
                      class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                    >
                      <option value="daily">{{ $t('groupAdmin.eventEditor.recurrenceTypes.daily') }}</option>
                      <option value="weekly">{{ $t('groupAdmin.eventEditor.recurrenceTypes.weekly') }}</option>
                      <option value="monthly">{{ $t('groupAdmin.eventEditor.recurrenceTypes.monthly') }}</option>
                      <option value="yearly">{{ $t('groupAdmin.eventEditor.recurrenceTypes.yearly') }}</option>
                    </select>
                  </div>

                  <!-- Interval -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      {{ $t('groupAdmin.eventEditor.interval') }}
                    </label>
                    <input
                      v-model.number="recurrencePattern.interval"
                      type="number"
                      min="1"
                      class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                    />
                  </div>
                </div>

                <!-- End Date -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    {{ $t('groupAdmin.eventEditor.recurrenceEnd') }}
                  </label>
                  <input
                    v-model="form.recurrence_end"
                    type="date"
                    class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- External Links -->
          <div class="card">
            <div class="mb-3">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
                <Icon icon="mdi:link" class="h-5 w-5 text-primary-500" />
                {{ $t('groupAdmin.eventEditor.externalLinks') }}
              </label>

              <div class="space-y-3">
                <div
                  v-for="(link, index) in externalLinks"
                  :key="index"
                  class="flex gap-3 p-3 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700"
                >
                  <input
                    v-model="link.title"
                    type="text"
                    :placeholder="$t('groupAdmin.eventEditor.linkTitlePlaceholder')"
                    class="flex-1 px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                  <input
                    v-model="link.url"
                    type="url"
                    :placeholder="$t('groupAdmin.eventEditor.linkUrlPlaceholder')"
                    class="flex-1 px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                  <button
                    type="button"
                    @click="removeLink(index)"
                    class="px-3 py-2 bg-red-500 text-white rounded-lg hover:bg-red-600 transition-colors"
                  >
                    <Icon icon="mdi:delete" class="h-4 w-4" />
                  </button>
                </div>

                <button
                  type="button"
                  @click="addLink"
                  class="w-full px-4 py-3 border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-primary-500 dark:hover:border-primary-400 transition-colors flex items-center justify-center gap-2 text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400"
                >
                  <Icon icon="mdi:plus" class="h-5 w-5" />
                  {{ $t('groupAdmin.eventEditor.addLink') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Sidebar -->
        <div class="space-y-6">
          <!-- Publish settings -->
          <div class="card bg-gradient-to-br from-white to-gray-50 dark:from-gray-800 dark:to-gray-900">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="mdi:publish" class="h-5 w-5 text-primary-500" />
              {{ $t('groupAdmin.eventEditor.publishSection') }}
            </h3>

            <div class="space-y-4">
              <!-- Status -->
              <div class="bg-white dark:bg-gray-800 rounded-lg p-3 border-2 border-gray-200 dark:border-gray-700">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_published"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ml-3 text-sm font-medium text-gray-900 dark:text-white">
                    {{ $t('groupAdmin.eventEditor.publishImmediately') }}
                  </span>
                </label>
              </div>

              <!-- Published date -->
              <div v-if="form.is_published">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:calendar-clock" class="h-4 w-4" />
                  {{ $t('groupAdmin.eventEditor.publishDate') }}
                </label>
                <input
                  v-model="form.published_at"
                  type="datetime-local"
                  class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                />
              </div>
            </div>
          </div>

          <!-- Target Groups -->
          <div class="card bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20 border-2 border-blue-200 dark:border-blue-800">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="mdi:account-group" class="h-5 w-5 text-blue-500" />
              {{ $t('groupAdmin.eventEditor.targetGroupsSection') }}
            </h3>

            <div class="space-y-3">
              <p class="text-xs text-gray-600 dark:text-gray-400">
                {{ $t('groupAdmin.eventEditor.targetGroupsHelp') }}
              </p>

              <!-- Public event toggle -->
              <div class="bg-white dark:bg-gray-800 rounded-lg p-3 border border-blue-200 dark:border-blue-700">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="isPublicEvent"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-blue-600 rounded focus:ring-2 focus:ring-blue-500"
                  />
                  <span class="ml-3 text-sm font-medium text-gray-900 dark:text-white">
                    <Icon icon="mdi:earth" class="inline h-4 w-4 mr-1" />
                    {{ $t('groupAdmin.eventEditor.publicEvent') }}
                  </span>
                </label>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ml-8">
                  {{ $t('groupAdmin.eventEditor.publicEventHelp') }}
                </p>
              </div>

              <!-- Selected groups (only if not public) -->
              <div v-if="!isPublicEvent">
                <div v-if="form.target_group_ids.length > 0" class="flex flex-wrap gap-2 p-3 bg-white dark:bg-gray-800 rounded-lg border border-blue-200 dark:border-blue-700">
                  <span
                    v-for="groupId in form.target_group_ids"
                    :key="groupId"
                    class="inline-flex items-center gap-1 px-3 py-1.5 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 rounded-full text-sm font-medium"
                  >
                    <Icon icon="mdi:account-group" class="h-3 w-3" />
                    {{ getGroupName(groupId) }}
                    <button
                      type="button"
                      @click="removeGroup(groupId)"
                      class="ml-1 hover:text-red-600 dark:hover:text-red-400 transition-colors"
                    >
                      <Icon icon="mdi:close-circle" class="h-4 w-4" />
                    </button>
                  </span>
                </div>
                <div v-else class="text-sm text-orange-500 dark:text-orange-400 italic p-3 bg-white dark:bg-gray-800 rounded-lg border border-orange-200 dark:border-orange-700">
                  {{ $t('groupAdmin.eventEditor.noGroupSelected') }}
                </div>

                <!-- Add group -->
                <select
                  @change="addGroup"
                  class="w-full px-3 py-2 border-2 border-blue-200 dark:border-blue-700 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-white text-sm"
                >
                  <option value="">{{ $t('groupAdmin.eventEditor.addGroup') }}</option>
                  <option
                    v-for="group in availableGroups"
                    :key="group.id"
                    :value="group.id"
                  >
                    {{ group.name }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <!-- Category & Priority -->
          <div class="card">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="mdi:tag-multiple" class="h-5 w-5 text-primary-500" />
              {{ $t('groupAdmin.eventEditor.classification') }}
            </h3>

            <div class="space-y-4">
              <!-- Category -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:folder" class="h-4 w-4" />
                  {{ $t('groupAdmin.eventEditor.category') }}
                </label>
                <select v-model="form.category_id" class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm">
                  <option :value="null">{{ $t('groupAdmin.eventEditor.noCategory') }}</option>
                  <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                    {{ cat.name }}
                  </option>
                </select>
              </div>

              <!-- Priority -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:flag" class="h-4 w-4" />
                  {{ $t('groupAdmin.eventEditor.priority') }}
                </label>
                <select v-model="form.priority" class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm">
                  <option value="normal">{{ $t('groupAdmin.eventEditor.priorities.normal') }}</option>
                  <option value="important">{{ $t('groupAdmin.eventEditor.priorities.important') }}</option>
                  <option value="urgent">{{ $t('groupAdmin.eventEditor.priorities.urgent') }}</option>
                </select>
              </div>

              <!-- Color -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:palette" class="h-4 w-4" />
                  {{ $t('groupAdmin.eventEditor.color') }}
                </label>
                <input
                  v-model="form.color"
                  type="color"
                  class="w-full h-10 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                />
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="card bg-gradient-to-br from-primary-50 to-primary-100 dark:from-primary-900/20 dark:to-primary-800/20 border-2 border-primary-200 dark:border-primary-800">
            <div class="space-y-3">
              <button
                type="submit"
                :disabled="isSaving"
                class="w-full px-6 py-3 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:scale-100 shadow-lg hover:shadow-xl flex items-center justify-center gap-2"
              >
                <Icon v-if="isSaving" icon="mdi:loading" class="h-5 w-5 animate-spin" />
                <Icon v-else icon="mdi:content-save" class="h-5 w-5" />
                <span>{{ isSaving ? $t('common.saving') : (isEditMode ? $t('groupAdmin.eventEditor.updateEvent') : $t('groupAdmin.eventEditor.createEvent')) }}</span>
              </button>

              <router-link
                to="/group-admin/events"
                class="w-full px-6 py-3 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 font-medium rounded-lg transition-all flex items-center justify-center gap-2"
              >
                <Icon icon="mdi:close" class="h-5 w-5" />
                <span>{{ $t('common.cancel') }}</span>
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, getCurrentInstance } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAppStore } from '@/stores/app'
import { groupAdminEventsService, eventsService, groupAdminService } from '@/services/api'
import RichTextEditor from '@/components/news/RichTextEditor.vue'

const { proxy } = getCurrentInstance()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()

const isEditMode = computed(() => route.params.slug && route.params.slug !== 'new')
const isSaving = ref(false)
const categories = ref([])
const managedGroups = ref([])

// Recurrence pattern
const recurrencePattern = ref({
  type: 'weekly',
  interval: 1
})

// External links
const externalLinks = ref([])

// Form
const form = ref({
  title: '',
  description: '',
  start_date: '',
  end_date: '',
  is_all_day: false,
  timezone: 'Europe/Paris',
  location: '',
  is_recurring: false,
  recurrence_rule: '',
  recurrence_end: '',
  external_links: '',
  color: '#6366f1',
  priority: 'normal',
  is_published: false,
  published_at: null,
  category_id: null,
  target_group_ids: []
})

// Public event toggle
const isPublicEvent = ref(true)

// Available groups (excluding already selected)
const availableGroups = computed(() => {
  return managedGroups.value.filter(group => !form.value.target_group_ids.includes(group.id))
})

// Get group name by ID
const getGroupName = (groupId) => {
  const group = managedGroups.value.find(g => g.id === groupId)
  return group ? group.name : ''
}

// Add group
const addGroup = (event) => {
  const groupId = parseInt(event.target.value)
  if (groupId && !form.value.target_group_ids.includes(groupId)) {
    form.value.target_group_ids.push(groupId)
  }
  event.target.value = ''
}

// Remove group
const removeGroup = (groupId) => {
  form.value.target_group_ids = form.value.target_group_ids.filter(id => id !== groupId)
}

// Add link
const addLink = () => {
  externalLinks.value.push({ title: '', url: '' })
}

// Remove link
const removeLink = (index) => {
  externalLinks.value.splice(index, 1)
}

// Load managed groups
const loadManagedGroups = async () => {
  try {
    const data = await groupAdminService.getManagedGroups()
    managedGroups.value = Array.isArray(data) ? data : (data.groups || data.data || [])
  } catch (error) {
    console.error('Error loading managed groups:', error)
    managedGroups.value = []
  }
}

// Load categories
const loadCategories = async () => {
  try {
    const data = await eventsService.getCategories()
    categories.value = Array.isArray(data) ? data : (data.data || [])
  } catch (error) {
    console.error('Error loading categories:', error)
    categories.value = []
  }
}

// Load event for editing
const loadEvent = async () => {
  if (!isEditMode.value) return

  try {
    const event = await eventsService.getEventBySlug(route.params.slug)

    // Parse description content if it's a JSON string
    let description = event.description || ''
    if (typeof description === 'string' && description.trim().startsWith('{')) {
      try {
        description = JSON.parse(description)
      } catch (e) {
        console.warn('Failed to parse description as JSON, using as-is')
      }
    }

    // Parse external links
    let links = []
    if (event.external_links) {
      try {
        links = JSON.parse(event.external_links)
      } catch (e) {
        console.warn('Failed to parse external links')
      }
    }
    externalLinks.value = links

    // Parse recurrence rule
    let recurrenceRule = { type: 'weekly', interval: 1 }
    if (event.recurrence_rule) {
      try {
        recurrenceRule = JSON.parse(event.recurrence_rule)
      } catch (e) {
        console.warn('Failed to parse recurrence rule')
      }
    }
    recurrencePattern.value = recurrenceRule

    // Determine if event is public
    const targetGroups = event.target_groups || []
    isPublicEvent.value = targetGroups.length === 0

    form.value = {
      title: event.title,
      description: description,
      start_date: event.start_date ? new Date(event.start_date).toISOString().slice(0, 16) : '',
      end_date: event.end_date ? new Date(event.end_date).toISOString().slice(0, 16) : '',
      is_all_day: event.is_all_day,
      timezone: event.timezone || 'Europe/Paris',
      location: event.location || '',
      is_recurring: event.is_recurring,
      recurrence_rule: event.recurrence_rule || '',
      recurrence_end: event.recurrence_end ? new Date(event.recurrence_end).toISOString().slice(0, 10) : '',
      external_links: event.external_links || '',
      color: event.color || '#6366f1',
      priority: event.priority || 'normal',
      is_published: event.is_published,
      published_at: event.published_at ? new Date(event.published_at).toISOString().slice(0, 16) : null,
      category_id: event.category_id,
      target_group_ids: targetGroups.map(g => g.id)
    }
  } catch (error) {
    console.error('Error loading event:', error)
    appStore.showError(proxy.$t('groupAdmin.eventEditor.loadError'))
    router.push('/group-admin/events')
  }
}

// Save event
const saveEvent = async () => {
  try {
    isSaving.value = true

    // Prepare recurrence rule
    let recurrenceRule = ''
    if (form.value.is_recurring) {
      recurrenceRule = JSON.stringify(recurrencePattern.value)
    }

    // Prepare external links
    const externalLinksJson = JSON.stringify(externalLinks.value.filter(link => link.title && link.url))

    // Format dates properly with ISO 8601 format (UTC)
    const formatDateTime = (dateStr) => {
      if (!dateStr) return null
      const date = new Date(dateStr)
      return date.toISOString()
    }

    const formatDate = (dateStr) => {
      if (!dateStr) return null
      const date = new Date(dateStr)
      return date.toISOString().split('T')[0]
    }

    const data = {
      ...form.value,
      // Convert Tiptap JSON description to string
      description: typeof form.value.description === 'object'
        ? JSON.stringify(form.value.description)
        : form.value.description,
      recurrence_rule: recurrenceRule,
      external_links: externalLinksJson,
      start_date: formatDateTime(form.value.start_date),
      end_date: form.value.end_date ? formatDateTime(form.value.end_date) : null,
      published_at: formatDateTime(form.value.published_at),
      recurrence_end: formatDate(form.value.recurrence_end),
      // Clear target groups if public event
      target_group_ids: isPublicEvent.value ? [] : form.value.target_group_ids
    }

    if (isEditMode.value) {
      await groupAdminEventsService.updateEvent(route.params.slug, data)
      appStore.showSuccess(proxy.$t('groupAdmin.eventEditor.updateSuccess'))
    } else {
      await groupAdminEventsService.createEvent(data)
      appStore.showSuccess(proxy.$t('groupAdmin.eventEditor.createSuccess'))
    }

    router.push('/group-admin/events')
  } catch (error) {
    console.error('Error saving event:', error)
    appStore.showError(proxy.$t('groupAdmin.eventEditor.saveError'))
  } finally {
    isSaving.value = false
  }
}

// Lifecycle
onMounted(async () => {
  await Promise.all([
    loadCategories(),
    loadManagedGroups(),
    loadEvent()
  ])
})
</script>
