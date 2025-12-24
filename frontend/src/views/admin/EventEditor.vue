<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ isEditMode ? t('events.editor.editEvent') : t('events.editor.newEvent') }}</h1>
          <p class="page-subtitle">{{ isEditMode ? t('events.editor.editSubtitle') : t('events.editor.newSubtitle') }}</p>
        </div>
        <router-link to="/admin/events" class="btn btn-secondary">
          <Icon icon="mdi:arrow-left" class="h-4 w-4 mr-2" />
          {{ t('common.backToList') }}
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
                {{ t('events.editor.title') }} *
              </label>
              <input
                v-model="form.title"
                type="text"
                required
                :placeholder="t('events.editor.titlePlaceholder')"
                class="w-full px-4 py-3 text-lg font-medium border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white transition-colors"
              />
            </div>
          </div>

          <!-- Description -->
          <div class="card">
            <div class="mb-3">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-3 flex items-center gap-2">
                <Icon icon="mdi:text-box" class="h-5 w-5 text-primary-500" />
                {{ t('events.editor.description') }} *
              </label>
              <div class="border-2 border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
                <RichTextEditor
                  v-model="form.description"
                  :placeholder="t('events.editor.descriptionPlaceholder')"
                />
              </div>
            </div>
          </div>

          <!-- Date & Time -->
          <div class="card">
            <div class="mb-4">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
                <Icon icon="mdi:calendar-clock" class="h-5 w-5 text-primary-500" />
                {{ t('events.editor.dateTime') }}
              </label>

              <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                <!-- Start Date & Time -->
                <div>
                  <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    {{ t('events.editor.startDate') }} *
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
                    {{ t('events.editor.endDate') }}
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
                    {{ t('events.editor.allDay') }}
                  </span>
                </label>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  {{ t('events.editor.allDayHelp') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Location -->
          <div class="card">
            <div class="mb-3">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-2 flex items-center gap-2">
                <Icon icon="mdi:map-marker" class="h-5 w-5 text-primary-500" />
                {{ t('events.editor.location') }}
              </label>
              <input
                v-model="form.location"
                type="text"
                :placeholder="t('events.editor.locationPlaceholder')"
                class="w-full px-4 py-3 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
              />
            </div>
          </div>

          <!-- Recurrence -->
          <div class="card">
            <div class="mb-4">
              <label class="block text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
                <Icon icon="mdi:repeat" class="h-5 w-5 text-primary-500" />
                {{ t('events.editor.recurrence') }}
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
                    {{ t('events.editor.isRecurring') }}
                  </span>
                </label>
              </div>

              <!-- Recurrence Pattern -->
              <div v-if="form.is_recurring" class="space-y-4 p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <!-- Recurrence Type -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      {{ t('events.editor.recurrenceType') }}
                    </label>
                    <select
                      v-model="recurrencePattern.type"
                      class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                    >
                      <option value="daily">{{ t('events.recurrence.daily') }}</option>
                      <option value="weekly">{{ t('events.recurrence.weekly') }}</option>
                      <option value="monthly">{{ t('events.recurrence.monthly') }}</option>
                      <option value="yearly">{{ t('events.recurrence.yearly') }}</option>
                    </select>
                  </div>

                  <!-- Interval -->
                  <div>
                    <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                      {{ t('events.editor.interval') }}
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
                    {{ t('events.editor.recurrenceEnd') }}
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
                {{ t('events.editor.externalLinks') }}
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
                    :placeholder="t('events.editor.linkTitlePlaceholder')"
                    class="flex-1 px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                  <input
                    v-model="link.url"
                    type="url"
                    :placeholder="t('events.editor.linkUrlPlaceholder')"
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
                  {{ t('events.editor.addLink') }}
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
              {{ t('events.editor.publication') }}
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
                    {{ t('events.editor.publishImmediately') }}
                  </span>
                </label>
              </div>

              <!-- Published date -->
              <div v-if="form.is_published">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:calendar-clock" class="h-4 w-4" />
                  {{ t('events.editor.publishDate') }}
                </label>
                <input
                  v-model="form.published_at"
                  type="datetime-local"
                  class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                />
              </div>
            </div>
          </div>

          <!-- Category & Classification -->
          <div class="card">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="mdi:tag-multiple" class="h-5 w-5 text-primary-500" />
              {{ t('events.editor.classification') }}
            </h3>

            <div class="space-y-4">
              <!-- Category -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:folder" class="h-4 w-4" />
                  {{ t('events.editor.category') }}
                </label>
                <div class="flex gap-2">
                  <select v-model="form.category_id" class="flex-1 px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm">
                    <option :value="null">{{ t('events.editor.noCategory') }}</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                      {{ cat.name }}
                    </option>
                  </select>
                  <button
                    type="button"
                    @click="showAddCategoryDialog = true"
                    class="flex-shrink-0 px-3 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors flex items-center gap-1"
                    :title="t('events.editor.addCategory')"
                  >
                    <Icon icon="mdi:plus" class="h-4 w-4" />
                  </button>
                </div>
              </div>

              <!-- Priority -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:flag" class="h-4 w-4" />
                  {{ t('events.editor.priority') }}
                </label>
                <select v-model="form.priority" class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm">
                  <option value="normal">⚪ {{ t('events.priority.normal') }}</option>
                  <option value="important">🟠 {{ t('events.priority.important') }}</option>
                  <option value="urgent">🔴 {{ t('events.priority.urgent') }}</option>
                </select>
              </div>

              <!-- Color -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2 flex items-center gap-1">
                  <Icon icon="mdi:palette" class="h-4 w-4" />
                  {{ t('events.editor.color') }}
                </label>
                <input
                  v-model="form.color"
                  type="color"
                  class="w-full h-10 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                />
              </div>
            </div>
          </div>

          <!-- Target Groups -->
          <div class="card">
            <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
              <Icon icon="mdi:account-group" class="h-5 w-5 text-primary-500" />
              {{ t('events.editor.visibility') }}
            </h3>

            <div class="space-y-3">
              <div class="bg-white dark:bg-gray-800 rounded-lg p-3 border-2 border-gray-200 dark:border-gray-700">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="isPublic"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ml-3 text-sm font-medium text-gray-900 dark:text-white">
                    <Icon icon="mdi:earth" class="inline h-4 w-4 mr-1" />
                    {{ t('events.visibility.public') }}
                  </span>
                </label>
              </div>

              <div v-if="!isPublic" class="space-y-2">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('events.editor.targetGroups') }}
                </label>
                <div class="max-h-32 overflow-y-auto space-y-2">
                  <label
                    v-for="group in availableGroups"
                    :key="group.id"
                    class="flex items-center cursor-pointer"
                  >
                    <input
                      v-model="form.target_group_ids"
                      type="checkbox"
                      :value="group.id"
                      class="form-checkbox h-4 w-4 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                    />
                    <span class="ml-2 text-sm text-gray-900 dark:text-white">
                      {{ group.name }}
                    </span>
                  </label>
                </div>
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
                <span>{{ isSaving ? t('common.saving') : (isEditMode ? t('events.editor.updateEvent') : t('events.editor.createEvent')) }}</span>
              </button>

              <router-link
                to="/admin/events"
                class="w-full px-6 py-3 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 font-medium rounded-lg transition-all flex items-center justify-center gap-2"
              >
                <Icon icon="mdi:close" class="h-5 w-5" />
                <span>{{ t('common.cancel') }}</span>
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </form>

    <!-- Add Category Dialog -->
    <div v-if="showAddCategoryDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[9999]" @click.self="showAddCategoryDialog = false">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4 transform transition-all">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('events.editor.addCategory') }}</h3>
          <button
            type="button"
            @click="showAddCategoryDialog = false"
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
          >
            <Icon icon="mdi:close" class="h-5 w-5" />
          </button>
        </div>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('events.editor.categoryName') }} *</label>
            <input
              v-model="newCategory.name"
              type="text"
              required
              class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white transition-colors"
              :placeholder="t('events.editor.categoryNamePlaceholder')"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('events.editor.categoryIcon') }}</label>
            <input
              v-model="newCategory.icon"
              type="text"
              class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white transition-colors"
              placeholder="mdi:folder"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">{{ t('events.editor.categoryColor') }}</label>
            <input
              v-model="newCategory.color"
              type="color"
              class="w-full h-10 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 cursor-pointer"
            />
          </div>
        </div>
        <div class="flex justify-end gap-3 mt-6">
          <button
            type="button"
            @click="showAddCategoryDialog = false"
            class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors font-medium"
          >
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            @click="createCategory"
            :disabled="!newCategory.name.trim()"
            class="px-4 py-2 bg-primary-500 hover:bg-primary-600 disabled:bg-gray-400 text-white rounded-lg transition-colors font-medium flex items-center gap-2"
          >
            <Icon icon="mdi:plus" class="h-4 w-4" />
            {{ t('common.create') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useEventsStore } from '@/stores/events'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import RichTextEditor from '@/components/news/RichTextEditor.vue'

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const eventsStore = useEventsStore()
const appStore = useAppStore()
const authStore = useAuthStore()

const isEditMode = computed(() => route.params.slug && route.params.slug !== 'new')
const isSaving = ref(false)
const categories = ref([])
const availableGroups = ref([])

// Dialogs
const showAddCategoryDialog = ref(false)

// New category form
const newCategory = ref({
  name: '',
  icon: 'mdi:calendar',
  color: '#6366f1'
})

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
  timezone: 'UTC',
  location: '',
  is_recurring: false,
  recurrence_rule: '',
  recurrence_end: '',
  external_links: '',
  color: '#6366f1',
  priority: 'normal',
  is_published: true,
  published_at: null,
  category_id: null,
  target_group_ids: []
})

// Computed
const isPublic = computed({
  get: () => form.value.target_group_ids.length === 0,
  set: (value) => {
    if (value) {
      form.value.target_group_ids = []
    }
  }
})

// Methods
const addLink = () => {
  externalLinks.value.push({ title: '', url: '' })
}

const removeLink = (index) => {
  externalLinks.value.splice(index, 1)
}

// Create new category
const createCategory = async () => {
  try {
    // Check if category name already exists
    const existingCategory = categories.value.find(cat => 
      cat.name.toLowerCase() === newCategory.value.name.trim().toLowerCase()
    )
    
    if (existingCategory) {
      appStore.showError(t('events.editor.categoryNameExists'))
      return
    }
    
    const category = await eventsStore.createCategory(newCategory.value)
    categories.value.push(category)
    form.value.category_id = category.id
    showAddCategoryDialog.value = false
    // Reset form
    newCategory.value = {
      name: '',
      icon: 'mdi:calendar',
      color: '#6366f1'
    }
    appStore.showSuccess(t('events.editor.categoryCreateSuccess'))
  } catch (error) {
    console.error('Error creating category:', error)
    if (error.response?.status === 409) {
      appStore.showError(t('events.editor.categoryNameExists'))
    } else {
      appStore.showError(t('events.editor.categoryCreateError'))
    }
  }
}

// Load event for editing
const loadEvent = async () => {
  if (!isEditMode.value) return

  try {
    const event = await eventsStore.fetchEventBySlug(route.params.slug)

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

    form.value = {
      title: event.title,
      description: description,
      start_date: event.start_date ? new Date(event.start_date).toISOString().slice(0, 16) : '',
      end_date: event.end_date ? new Date(event.end_date).toISOString().slice(0, 16) : '',
      is_all_day: event.is_all_day,
      timezone: event.timezone || 'UTC',
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
      target_group_ids: event.target_groups ? event.target_groups.map(g => g.id) : []
    }
  } catch (error) {
    console.error('Error loading event:', error)
    appStore.showError(t('events.editor.loadError'))
    router.push('/admin/events')
  }
}

// Load categories and groups
const loadMetadata = async () => {
  try {
    await eventsStore.fetchCategories()
    categories.value = eventsStore.categories

    // Load groups (this would need to be implemented in the store)
    // For now, we'll use a placeholder
    availableGroups.value = []
  } catch (error) {
    console.error('Error loading metadata:', error)
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
      // Convert to ISO 8601 format with UTC timezone
      return date.toISOString()
    }

    const formatDate = (dateStr) => {
      if (!dateStr) return null
      const date = new Date(dateStr)
      // Format as YYYY-MM-DD in UTC
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
      recurrence_end: formatDate(form.value.recurrence_end)
    }

    if (isEditMode.value) {
      await eventsStore.updateEvent(route.params.slug, data)
      appStore.showSuccess(t('events.editor.updateSuccess'))
    } else {
      await eventsStore.createEvent(data)
      appStore.showSuccess(t('events.editor.createSuccess'))
    }

    router.push('/admin/events')
  } catch (error) {
    console.error('Error saving event:', error)
    appStore.showError(t('events.editor.saveError'))
  } finally {
    isSaving.value = false
  }
}

// Lifecycle
onMounted(() => {
  loadMetadata()
  loadEvent()
})
</script>