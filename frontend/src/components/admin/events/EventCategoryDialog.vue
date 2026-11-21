<template>
  <!-- Création d'une catégorie d'événements : POST /admin/events/categories, admin global uniquement -->
  <div
    class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[9999]"
    @click.self="emit('close')"
  >
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 max-w-md w-full mx-4 transform transition-all">
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('events.editor.addCategory') }}</h3>
        <button
          type="button"
          class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
          :aria-label="t('common.close')"
          @click="emit('close')"
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
          class="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors font-medium"
          @click="emit('close')"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          :disabled="!newCategory.name.trim() || isCreating"
          class="px-4 py-2 bg-primary-500 hover:bg-primary-600 disabled:bg-gray-400 text-white rounded-lg transition-colors font-medium flex items-center gap-2"
          @click="createCategory"
        >
          <Icon :icon="isCreating ? 'mdi:loading' : 'mdi:plus'" class="h-4 w-4" :class="{ 'animate-spin': isCreating }" />
          {{ t('common.create') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { adminEventsService } from '@/services/api'
import { useAppStore } from '@/stores/app'

const DEFAULT_CATEGORY = Object.freeze({ name: '', icon: 'mdi:calendar', color: '#6366f1' })
const HTTP_CONFLICT = 409

const props = defineProps({
  // Catégories existantes, pour refuser un doublon avant l'appel
  categories: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'created'])

const { t } = useI18n()
const appStore = useAppStore()

const newCategory = ref({ ...DEFAULT_CATEGORY })
const isCreating = ref(false)

const createCategory = async () => {
  const name = newCategory.value.name.trim()
  const exists = props.categories.some((cat) => cat.name?.toLowerCase() === name.toLowerCase())
  if (exists) {
    appStore.showError(t('events.editor.categoryNameExists'))
    return
  }

  isCreating.value = true
  try {
    const category = await adminEventsService.createCategory({ ...newCategory.value, name })
    newCategory.value = { ...DEFAULT_CATEGORY }
    appStore.showSuccess(t('events.editor.categoryCreateSuccess'))
    emit('created', category)
  } catch (error) {
    console.error('Erreur à la création de la catégorie :', error)
    appStore.showError(
      error.response?.status === HTTP_CONFLICT
        ? t('events.editor.categoryNameExists')
        : t('events.editor.categoryCreateError')
    )
  } finally {
    isCreating.value = false
  }
}
</script>
