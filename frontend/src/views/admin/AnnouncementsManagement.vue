<template>
  <div class="content-area">
    <PageHeader>
      <template #title>{{ $t('announcements.title') }}</template>
      <template #description>{{ $t('announcements.subtitle') }}</template>
      <template #actions>
        <BaseButton variant="primary" icon="lucide:plus" @click="openCreateModal">
          {{ $t('announcements.new') }}
        </BaseButton>
      </template>
    </PageHeader>

    <EmptyState
      v-if="!isLoading && announcements.length === 0"
      icon="lucide:megaphone"
      :title="$t('announcements.emptyTitle')"
      :description="$t('announcements.emptyText')"
    >
      <BaseButton variant="primary" size="sm" icon="lucide:plus" @click="openCreateModal">
        {{ $t('announcements.create') }}
      </BaseButton>
    </EmptyState>

    <DataTable
      v-else
      :columns="columns"
      :rows="rows"
      :loading="isLoading"
      :search-keys="['title', 'content', 'typeLabel']"
      :page-size="25"
    >
      <template #toolbar>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.type') }}</span>
          <select v-model="typeFilter" class="ui-select">
            <option value="">{{ $t('ui.table.columns.type') }}</option>
            <option v-for="type in TYPES" :key="type" :value="type">{{ $t(`announcements.type_${type}`) }}</option>
          </select>
        </label>
        <label>
          <span class="sr-only">{{ $t('ui.table.columns.status') }}</span>
          <select v-model="statusFilter" class="ui-select">
            <option value="">{{ $t('common.status_all') }}</option>
            <option value="active">{{ $t('common.status_active') }}</option>
            <option value="inactive">{{ $t('common.status_inactive') }}</option>
          </select>
        </label>
      </template>

      <template #cell-title="{ row }">
        <span class="an-title">
          <span class="an-title__text">{{ row.title }}</span>
          <span v-if="row.content" class="an-title__content">{{ row.content }}</span>
        </span>
      </template>

      <template #cell-typeLabel="{ row }">
        <StatusBadge :tone="TYPE_TONES[row.raw.type] || 'info'">{{ row.typeLabel }}</StatusBadge>
      </template>

      <template #cell-status="{ row }">
        <StatusBadge :tone="row.raw.is_active ? 'ok' : 'neutral'">{{ row.status }}</StatusBadge>
      </template>

      <template #cell-priority="{ row }">
        <span class="an-mono">{{ row.priority }}</span>
      </template>

      <template #cell-period="{ row }">
        <span class="an-mono">{{ row.period }}</span>
      </template>

      <template #cell-actions="{ row }">
        <RowActions :label="row.title" @edit="openEditModal(row.raw)" @delete="confirmDelete(row.raw)" />
      </template>
    </DataTable>

      <!-- Create/Edit Modal -->
      <div
        v-if="showModal"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
        @click.self="closeModal"
      >
        <div v-focus-trap class="bg-white dark:bg-gray-800 rounded-lg max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto" role="dialog" aria-modal="true" aria-labelledby="announcement-modal-title">
          <div class="flex items-center justify-between mb-6">
            <h2 id="announcement-modal-title" class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ isEditMode ? $t('announcements.edit') : $t('announcements.create') }}
            </h2>
            <button
              type="button"
              @click="closeModal"
              :aria-label="$t('common.close')"
              class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            >
              <Icon icon="lucide:x" class="h-5 w-5" />
            </button>
          </div>

          <form @submit.prevent="saveAnnouncement" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {{ $t('announcements.titleField') }} *
              </label>
              <input
                v-model="formData.title"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {{ $t('announcements.contentField') }}
              </label>
              <textarea
                v-model="formData.content"
                rows="4"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
              ></textarea>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {{ $t('announcements.typeField') }} *
                </label>
                <select
                  v-model="formData.type"
                  required
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
                >
                  <option value="info">{{ $t('announcements.type_info') }}</option>
                  <option value="warning">{{ $t('announcements.type_warning') }}</option>
                  <option value="success">{{ $t('announcements.type_success') }}</option>
                  <option value="error">{{ $t('announcements.type_error') }}</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {{ $t('announcements.priorityField') }}
                </label>
                <input
                  v-model.number="formData.priority"
                  type="number"
                  min="0"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
                />
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {{ $t('announcements.startDateField') }}
                </label>
                <input
                  v-model="formData.start_date"
                  type="datetime-local"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  {{ $t('announcements.endDateField') }}
                </label>
                <input
                  v-model="formData.end_date"
                  type="datetime-local"
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-primary-500"
                />
              </div>
            </div>

            <div class="flex items-center gap-2">
              <input
                v-model="formData.is_active"
                type="checkbox"
                id="is_active"
                class="w-4 h-4 text-primary-600 rounded focus:ring-2 focus:ring-primary-500"
              />
              <label for="is_active" class="text-sm font-medium text-gray-700 dark:text-gray-300">
                {{ $t('announcements.isActive') }}
              </label>
            </div>

            <div class="flex gap-3 pt-4">
              <button
                type="button"
                @click="closeModal"
                class="flex-1 px-4 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg font-medium hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
              >
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="isSaving"
                class="flex-1 px-4 py-2 bg-primary-600 hover:bg-primary-700 disabled:bg-primary-400 text-white rounded-lg font-medium transition-colors"
              >
                {{ isSaving ? $t('announcements.saving') : $t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
  </div>
</template>

<script setup>
import StatusBadge from '@/components/ui/StatusBadge.vue'
import PageHeader from '@/components/ui/PageHeader.vue'
import DataTable from '@/components/ui/DataTable.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import RowActions from '@/components/ui/RowActions.vue'
import { useConfirm } from '@/composables/useConfirm'
import { dateLocaleFor } from '@/utils/homeFormat'
import { ref, computed, onMounted } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { announcementsService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { isEndBeforeStart } from '@/utils/dateInput'

const { t, locale } = useI18n()
const appStore = useAppStore()
const { confirm: askConfirm } = useConfirm()

const TYPES = ['info', 'warning', 'success', 'error']
const TYPE_TONES = { info: 'info', warning: 'warn', success: 'ok', error: 'bad' }

const announcements = ref([])
const typeFilter = ref('')
const statusFilter = ref('')
const isLoading = ref(false)
const isSaving = ref(false)
const showModal = ref(false)
const isEditMode = ref(false)

const formData = ref({
  title: '',
  content: '',
  type: 'info',
  priority: 0,
  is_active: true,
  start_date: null,
  end_date: null
})

const loadAnnouncements = async () => {
  isLoading.value = true
  try {
    announcements.value = await announcementsService.getAllAnnouncements()
  } catch (error) {
    console.error('Error loading announcements:', error)
    appStore.showError(t('common.feedback.loadError'))
  } finally {
    isLoading.value = false
  }
}

const openCreateModal = () => {
  isEditMode.value = false
  formData.value = {
    title: '',
    content: '',
    type: 'info',
    priority: 0,
    is_active: true,
    start_date: null,
    end_date: null
  }
  showModal.value = true
}

const openEditModal = (announcement) => {
  isEditMode.value = true
  formData.value = {
    id: announcement.id,
    title: announcement.title,
    content: announcement.content,
    type: announcement.type,
    priority: announcement.priority,
    is_active: announcement.is_active,
    start_date: announcement.start_date ? formatDateTimeLocal(announcement.start_date) : null,
    end_date: announcement.end_date ? formatDateTimeLocal(announcement.end_date) : null
  }
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const saveAnnouncement = async () => {
  if (isEndBeforeStart(formData.value.start_date, formData.value.end_date)) {
    appStore.showError(t('common.validation.endBeforeStart'))
    return
  }
  isSaving.value = true
  try {
    const data = {
      ...formData.value,
      start_date: formData.value.start_date ? new Date(formData.value.start_date).toISOString() : null,
      end_date: formData.value.end_date ? new Date(formData.value.end_date).toISOString() : null
    }
    delete data.id

    if (isEditMode.value) {
      await announcementsService.updateAnnouncement(formData.value.id, data)
    } else {
      await announcementsService.createAnnouncement(data)
    }

    appStore.showSuccess(t(isEditMode.value ? 'common.feedback.saved' : 'common.feedback.created'))
    closeModal()
    await loadAnnouncements()
  } catch (error) {
    console.error('Error saving announcement:', error)
    appStore.showError(t('common.feedback.saveError'))
  } finally {
    isSaving.value = false
  }
}

const confirmDelete = async (announcement) => {
  const confirmed = await askConfirm({
    title: t('announcements.confirmDeleteTitle'),
    message: t('announcements.confirmDeleteText', { title: announcement.title }),
    confirmLabel: t('common.delete'),
    tone: 'danger'
  })
  if (!confirmed) return

  try {
    await announcementsService.deleteAnnouncement(announcement.id)
    appStore.showSuccess(t('common.feedback.deleted'))
    await loadAnnouncements()
  } catch (error) {
    console.error('Error deleting announcement:', error)
    appStore.showError(t('common.feedback.deleteError'))
  }
}

const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString(dateLocaleFor(locale.value), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatDateTimeLocal = (date) => {
  if (!date) return null
  const d = new Date(date)
  const offset = d.getTimezoneOffset()
  d.setMinutes(d.getMinutes() - offset)
  return d.toISOString().slice(0, 16)
}

const columns = computed(() => [
  { key: 'title', label: t('ui.table.columns.title'), sortable: true },
  { key: 'typeLabel', label: t('ui.table.columns.type'), sortable: true, hideSm: true },
  { key: 'status', label: t('ui.table.columns.status'), sortable: true },
  { key: 'priority', label: t('announcements.priority'), sortable: true, align: 'end', hideSm: true },
  { key: 'period', label: t('ui.table.columns.period'), hideSm: true },
  { key: 'actions', label: '', align: 'end', width: '80px' },
])

// Lignes du tableau : filtres type et statut ; recherche et tri dans DataTable.
const rows = computed(() =>
  announcements.value
    .filter(a => !typeFilter.value || a.type === typeFilter.value)
    .filter(a => {
      if (statusFilter.value === 'active') return a.is_active
      if (statusFilter.value === 'inactive') return !a.is_active
      return true
    })
    .map(a => ({
      id: a.id,
      title: a.title,
      content: a.content || '',
      typeLabel: t(`announcements.type_${a.type}`),
      status: a.is_active ? t('common.active') : t('common.inactive'),
      priority: a.priority || 0,
      period: (a.start_date || a.end_date)
        ? `${a.start_date ? formatDate(a.start_date) : '…'} → ${a.end_date ? formatDate(a.end_date) : '…'}`
        : '—',
      raw: a,
    })),
)

onMounted(() => {
  loadAnnouncements()
})

// Échap ferme la modale
useEscapeClose(showModal, () => closeModal())
</script>

<style scoped>
.an-title { display: flex; flex-direction: column; min-width: 0; }
.an-title__text { font-weight: 600; color: var(--text); }
.an-title__content {
  max-width: 52ch;
  overflow: hidden;
  font-size: 12px;
  color: var(--muted);
  text-overflow: ellipsis;
  white-space: nowrap;
}
.an-mono { font-family: var(--f-num); font-size: 11.5px; color: var(--muted); white-space: nowrap; }
</style>
