<template>
  <div class="content-area">
    <!-- Header -->
    <PageHeader>
      <template #title>{{ headerTitle }}</template>
      <template #description>{{ headerSubtitle }}</template>
      <template #actions>
        <router-link :to="LIST_PATH" class="btn btn-secondary">
          <Icon icon="mdi:arrow-left" class="rtl-flip h-4 w-4 me-2" />
          {{ t('common.backToList') }}
        </router-link>
      </template>
    </PageHeader>

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
                class="w-full px-3 py-1.5 text-lg font-medium border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white transition-colors"
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
              <div class="border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden">
                <RichTextEditor
                  v-model="form.description"
                  :placeholder="t('events.editor.descriptionPlaceholder')"
                  :as-group-admin="!isGlobal"
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
                    class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
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
                    class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                </div>
              </div>

              <!-- All Day Toggle -->
              <div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-300 dark:border-gray-600">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_all_day"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ms-3 text-sm font-medium text-gray-900 dark:text-white">
                    <Icon icon="mdi:weather-sunny" class="inline h-4 w-4 me-1" />
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
                class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
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
              <div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-300 dark:border-gray-600 mb-4">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_recurring"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ms-3 text-sm font-medium text-gray-900 dark:text-white">
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
                      class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
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
                      class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
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
                    class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
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
                    class="flex-1 px-3 py-1.5 border border-gray-200 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
                  />
                  <input
                    v-model="link.url"
                    type="url"
                    :placeholder="t('events.editor.linkUrlPlaceholder')"
                    class="flex-1 px-3 py-1.5 border border-gray-200 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white"
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
                  class="w-full px-3 py-[5px] border border-dashed border-gray-300 dark:border-gray-600 rounded-lg hover:border-primary-500 dark:hover:border-primary-400 transition-colors flex items-center justify-center gap-2 text-gray-600 dark:text-gray-400 hover:text-primary-600 dark:hover:text-primary-400"
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
              <div class="bg-white dark:bg-gray-800 rounded-lg p-3 border border-gray-300 dark:border-gray-600">
                <label class="flex items-center cursor-pointer">
                  <input
                    v-model="form.is_published"
                    type="checkbox"
                    class="form-checkbox h-5 w-5 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
                  />
                  <span class="ms-3 text-sm font-medium text-gray-900 dark:text-white">
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
                  class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm"
                />
              </div>
            </div>
          </div>

          <!-- Visibilité : public ou groupes ciblés -->
          <EventVisibilityPicker
            v-model:group-ids="form.target_group_ids"
            v-model:is-public="isPublicEvent"
            :groups="availableGroups"
          />

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
                  <select v-model="form.category_id" class="flex-1 px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm">
                    <option :value="null">{{ t('events.editor.noCategory') }}</option>
                    <option v-for="cat in categories" :key="cat.id" :value="cat.id">
                      {{ cat.name }}
                    </option>
                  </select>
                  <!-- Création de catégorie : POST /admin/events/categories, admin global uniquement -->
                  <button
                    v-if="isAdmin"
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
                <select v-model="form.priority" class="w-full px-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm">
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
                  class="w-full h-8 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
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
                class="w-full px-4 py-1.5 bg-primary-600 hover:bg-primary-700 disabled:bg-gray-400 text-white font-semibold rounded-lg transition-all transform hover:scale-[1.02] active:scale-[0.98] disabled:scale-100 shadow-lg hover:shadow-xl flex items-center justify-center gap-2"
              >
                <Icon v-if="isSaving" icon="mdi:loading" class="h-5 w-5 animate-spin" />
                <Icon v-else icon="mdi:content-save" class="h-5 w-5" />
                <span>{{ isSaving ? t('common.saving') : (isEditMode ? t('events.editor.updateEvent') : t('events.editor.createEvent')) }}</span>
              </button>

              <router-link
                :to="LIST_PATH"
                class="w-full px-4 py-1.5 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-800 dark:text-gray-200 font-medium rounded-lg transition-all flex items-center justify-center gap-2"
              >
                <Icon icon="mdi:close" class="h-5 w-5" />
                <span>{{ t('common.cancel') }}</span>
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </form>

    <EventCategoryDialog
      v-if="isAdmin && showAddCategoryDialog"
      :categories="categories"
      @close="showAddCategoryDialog = false"
      @created="handleCategoryCreated"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { toDateTimeInput, fromDateTimeInput, isEndBeforeStart } from '@/utils/dateInput'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import RichTextEditor from '@/components/news/RichTextEditor.vue'
import EventCategoryDialog from '@/components/admin/events/EventCategoryDialog.vue'
import EventVisibilityPicker from '@/components/admin/events/EventVisibilityPicker.vue'
import { getEventWriteService } from '@/components/admin/events/eventServices'
import { adminService, eventsService, groupAdminService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { useManagementScope } from '@/composables/useManagementScope'
import { CONTENT } from '@/config/access'

const LIST_PATH = '/admin/events'
const DEFAULT_COLOR = '#6366f1'
const DEFAULT_RECURRENCE = Object.freeze({ type: 'weekly', interval: 1 })

const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const appStore = useAppStore()
// Admin et éditeur : API globale ; admin de groupe : API /group-admin, groupes administrés
const { isGlobal, isAdmin } = useManagementScope(CONTENT)

const isEditMode = computed(() => route.params.slug && route.params.slug !== 'new')
const isSaving = ref(false)
const categories = ref([])
const availableGroups = ref([])
const showAddCategoryDialog = ref(false)
const isPublicEvent = ref(true)
const recurrencePattern = ref({ ...DEFAULT_RECURRENCE })
const externalLinks = ref([])

const headerTitle = computed(() => {
  if (isGlobal.value) return isEditMode.value ? t('events.editor.editEvent') : t('events.editor.newEvent')
  return isEditMode.value ? t('groupAdmin.eventEditor.title.edit') : t('groupAdmin.eventEditor.title.new')
})

const headerSubtitle = computed(() => {
  if (isGlobal.value) return isEditMode.value ? t('events.editor.editSubtitle') : t('events.editor.newSubtitle')
  return isEditMode.value ? t('groupAdmin.eventEditor.subtitle.edit') : t('groupAdmin.eventEditor.subtitle.new')
})

// Valeurs par défaut reprises des deux anciennes pages :
// global → fuseau UTC, publié d'emblée ; groupe → Europe/Paris, brouillon
const defaultTimezone = () => (isGlobal.value ? 'UTC' : 'Europe/Paris')

const form = ref({
  title: '',
  description: '',
  start_date: '',
  end_date: '',
  is_all_day: false,
  timezone: defaultTimezone(),
  location: '',
  is_recurring: false,
  recurrence_rule: '',
  recurrence_end: '',
  external_links: '',
  color: DEFAULT_COLOR,
  priority: 'normal',
  is_published: isGlobal.value,
  published_at: null,
  category_id: null,
  target_group_ids: []
})

const addLink = () => {
  externalLinks.value = [...externalLinks.value, { title: '', url: '' }]
}

const removeLink = (index) => {
  externalLinks.value = externalLinks.value.filter((_, i) => i !== index)
}

const handleCategoryCreated = (category) => {
  categories.value = [...categories.value, category]
  form.value = { ...form.value, category_id: category.id }
  showAddCategoryDialog.value = false
}

// JSON stocké en chaîne ; en cas d'échec on garde la valeur de repli
const parseJson = (value, fallback) => {
  if (!value) return fallback
  try {
    return JSON.parse(value)
  } catch {
    return fallback
  }
}

// Date seule (fin de récurrence) : écrite et relue en UTC, donc stable
const toDateInput = (value) => (value ? new Date(value).toISOString().slice(0, 10) : '')

const loadEvent = async () => {
  if (!isEditMode.value) return

  try {
    const event = await eventsService.getEventBySlug(route.params.slug)

    // Description Tiptap stockée en JSON (sinon texte brut)
    const rawDescription = event.description || ''
    const description = typeof rawDescription === 'string' && rawDescription.trim().startsWith('{')
      ? parseJson(rawDescription, rawDescription)
      : rawDescription

    externalLinks.value = parseJson(event.external_links, [])
    recurrencePattern.value = parseJson(event.recurrence_rule, { ...DEFAULT_RECURRENCE })

    const targetGroups = event.target_groups || []
    isPublicEvent.value = targetGroups.length === 0

    form.value = {
      title: event.title,
      description,
      start_date: toDateTimeInput(event.start_date),
      end_date: toDateTimeInput(event.end_date),
      is_all_day: event.is_all_day,
      timezone: event.timezone || defaultTimezone(),
      location: event.location || '',
      is_recurring: event.is_recurring,
      recurrence_rule: event.recurrence_rule || '',
      recurrence_end: toDateInput(event.recurrence_end),
      external_links: event.external_links || '',
      color: event.color || DEFAULT_COLOR,
      priority: event.priority || 'normal',
      is_published: event.is_published,
      published_at: event.published_at ? toDateTimeInput(event.published_at) : null,
      category_id: event.category_id,
      target_group_ids: targetGroups.map((g) => g.id)
    }
  } catch (error) {
    console.error('Erreur au chargement de l\'événement :', error)
    appStore.showError(t('events.editor.loadError'))
    router.push(LIST_PATH)
  }
}

const loadCategories = async () => {
  try {
    const data = await eventsService.getCategories()
    categories.value = Array.isArray(data) ? data : (data?.data || [])
  } catch (error) {
    console.error('Erreur au chargement des catégories :', error)
    categories.value = []
  }
}

// Groupes proposés comme cibles : tous pour l'admin, ceux qu'il administre pour l'admin de groupe.
// L'éditeur n'a pas d'API listant les groupes (GET /admin/groups est réservé à l'admin).
const fetchTargetGroups = async () => {
  if (!isGlobal.value) {
    const data = await groupAdminService.getManagedGroups()
    return Array.isArray(data) ? data : (data?.groups || data?.data || [])
  }
  if (isAdmin.value) {
    const data = await adminService.getGroups()
    return Array.isArray(data) ? data : (data?.data || data?.groups || [])
  }
  return []
}

const loadGroups = async () => {
  try {
    availableGroups.value = await fetchTargetGroups()
  } catch (error) {
    console.error('Erreur au chargement des groupes :', error)
    availableGroups.value = []
  }
}

const toIsoDateTime = fromDateTimeInput
const toIsoDate = (value) => (value ? new Date(value).toISOString().split('T')[0] : null)

const buildPayload = () => ({
  ...form.value,
  // Description Tiptap (objet) envoyée en chaîne JSON
  description: typeof form.value.description === 'object'
    ? JSON.stringify(form.value.description)
    : form.value.description,
  recurrence_rule: form.value.is_recurring ? JSON.stringify(recurrencePattern.value) : '',
  external_links: JSON.stringify(externalLinks.value.filter((link) => link.title && link.url)),
  start_date: toIsoDateTime(form.value.start_date),
  end_date: toIsoDateTime(form.value.end_date),
  published_at: toIsoDateTime(form.value.published_at),
  recurrence_end: toIsoDate(form.value.recurrence_end),
  // Événement public : aucun groupe ciblé
  target_group_ids: isPublicEvent.value ? [] : form.value.target_group_ids
})

const saveEvent = async () => {
  if (isEndBeforeStart(form.value.start_date, form.value.end_date)) {
    appStore.showError(t('common.validation.endBeforeStart'))
    return
  }
  isSaving.value = true
  try {
    const service = getEventWriteService(isAdmin.value, isGlobal.value)
    const data = buildPayload()

    if (isEditMode.value) {
      await service.updateEvent(route.params.slug, data)
      appStore.showSuccess(t('events.editor.updateSuccess'))
    } else {
      await service.createEvent(data)
      appStore.showSuccess(t('events.editor.createSuccess'))
    }

    router.push(LIST_PATH)
  } catch (error) {
    console.error('Erreur à l\'enregistrement de l\'événement :', error)
    appStore.showError(t('events.editor.saveError'))
  } finally {
    isSaving.value = false
  }
}

onMounted(() => {
  loadCategories()
  loadGroups()
  loadEvent()
})
</script>
