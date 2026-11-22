<template>
  <!-- Création rapide d'une catégorie d'actualités : endpoint réservé à l'admin global -->
  <BaseModal :open="open" :title="$t('events.editor.addCategory')" size="sm" @close="close">
    <form :id="formId" class="space-y-4" @submit.prevent="submit">
      <div>
        <label :for="`${formId}-name`" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ $t('events.editor.categoryName') }} *
        </label>
        <input
          :id="`${formId}-name`"
          v-model="form.name"
          type="text"
          required
          :placeholder="$t('events.editor.categoryNamePlaceholder')"
          class="input"
        />
      </div>
      <div>
        <label :for="`${formId}-description`" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ $t('events.editor.description') }}
        </label>
        <textarea
          :id="`${formId}-description`"
          v-model="form.description"
          rows="2"
          class="input resize-none"
        ></textarea>
      </div>
      <div class="grid grid-cols-2 gap-3">
        <div class="col-span-2">
          <IconInput :id="`${formId}-icon`" v-model="form.icon" :label="$t('events.editor.categoryIcon')" placeholder="mdi:folder" />
        </div>
        <div>
          <label :for="`${formId}-color`" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
            {{ $t('events.editor.categoryColor') }}
          </label>
          <input
            :id="`${formId}-color`"
            v-model="form.color"
            type="color"
            class="input h-8"
          />
        </div>
      </div>
    </form>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="close">
        {{ $t('common.cancel') }}
      </button>
      <button
        type="submit"
        :form="formId"
        :disabled="!form.name.trim() || isSaving"
        class="btn btn-primary"
      >
        <Icon v-if="isSaving" icon="mdi:loading" class="h-4 w-4 me-2 animate-spin" />
        {{ isSaving ? $t('groupAdmin.newsEditor.actions.saving') : $t('common.create') }}
      </button>
    </template>
  </BaseModal>
</template>

<script setup>
import IconInput from '@/components/ui/IconInput.vue'
import { ref } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import BaseModal from '@/components/ui/BaseModal.vue'
import { useAppStore } from '@/stores/app'
import { newsService } from '@/services/api'

defineProps({
  open: { type: Boolean, default: false }
})

const emit = defineEmits(['close', 'created'])

const { t } = useI18n()
const appStore = useAppStore()

const formId = 'news-category-form'

// is_active et order sont envoyés explicitement : le backend lit is_active tel quel (false si absent)
const emptyForm = () => ({
  name: '',
  description: '',
  icon: 'mdi:folder',
  color: '#3B82F6',
  order: 0,
  is_active: true
})

const form = ref(emptyForm())
const isSaving = ref(false)

const close = () => {
  form.value = emptyForm()
  emit('close')
}

const submit = async () => {
  if (!form.value.name.trim() || isSaving.value) return

  isSaving.value = true
  try {
    const category = await newsService.createCategory({ ...form.value, name: form.value.name.trim() })
    appStore.showSuccess(t('events.editor.categoryCreateSuccess'))
    emit('created', category)
    close()
  } catch (error) {
    console.error('Erreur à la création de la catégorie :', error)
    appStore.showError(error.response?.data?.error || t('admin.categorySaveError'))
  } finally {
    isSaving.value = false
  }
}
</script>
