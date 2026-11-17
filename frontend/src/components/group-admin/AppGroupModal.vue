<template>
  <div v-if="show" class="modal-overlay">
    <div class="modal-container">
      <div class="modal-panel sm:max-w-5xl sm:w-full">
        <form class="modal-form" @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="modal-title">
                  {{ isEdit ? $t('groupAdmin.appGroupModal.title.edit') : $t('groupAdmin.appGroupModal.title.new') }}
                </h3>
                <p class="modal-subtitle">
                  {{ isEdit ? $t('groupAdmin.appGroupModal.subtitle.edit') : $t('groupAdmin.appGroupModal.subtitle.new') }}
                </p>
              </div>
              <button
                type="button"
                @click="closeModal"
                class="btn-ghost p-2"
              >
                <Icon icon="mdi:close" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Content : 2 colonnes (infos | apparence) pour un modal large et compact -->
          <div class="modal-content grid grid-cols-1 lg:grid-cols-2 gap-x-8 gap-y-5">
            <!-- Colonne gauche : informations -->
            <div class="space-y-4">
              <div class="section-header">
                <Icon icon="mdi:information-outline" class="section-icon" />
                <h4 class="section-title">{{ $t('groupAdmin.appGroupModal.basicInfo') }}</h4>
              </div>

              <div class="form-group">
                <label for="owner_group_id" class="form-label form-label-required">
                  {{ $t('groupAdmin.appGroupModal.ownerGroup') }}
                </label>
                <select
                  id="owner_group_id"
                  v-model="form.owner_group_id"
                  required
                  class="form-select"
                  :disabled="isEdit"
                >
                  <option value="">{{ $t('groupAdmin.appGroupModal.selectGroup') }}</option>
                  <option
                    v-for="group in managedGroups"
                    :key="group.id"
                    :value="group.id"
                  >
                    {{ group.name }}
                  </option>
                </select>
                <p class="text-xs text-gray-400 mt-1">
                  {{ $t('groupAdmin.appGroupModal.ownerGroupHelp') }}
                  {{ isEdit ? ' ' + $t('groupAdmin.appGroupModal.ownerGroupEditHelp') : '' }}
                </p>
                <p v-if="errors.owner_group_id" class="form-error">{{ errors.owner_group_id }}</p>
              </div>

              <div class="grid grid-cols-[1fr_7rem] gap-4">
                <div class="form-group">
                  <label for="name" class="form-label form-label-required">
                    {{ $t('groupAdmin.appGroupModal.groupName') }}
                  </label>
                  <input
                    id="name"
                    v-model="form.name"
                    type="text"
                    required
                    class="form-input"
                    :placeholder="$t('groupAdmin.appGroupModal.groupName')"
                  />
                  <p v-if="errors.name" class="form-error">{{ errors.name }}</p>
                </div>

                <div class="form-group">
                  <label for="order" class="form-label">
                    {{ $t('groupAdmin.appGroupModal.displayOrder') }}
                  </label>
                  <input
                    id="order"
                    v-model.number="form.order"
                    type="number"
                    min="0"
                    class="form-input"
                    placeholder="0"
                  />
                </div>
              </div>

              <div class="form-group">
                <label for="description" class="form-label">
                  {{ $t('groupAdmin.appGroupModal.description') }}
                </label>
                <textarea
                  id="description"
                  v-model="form.description"
                  rows="2"
                  class="form-textarea"
                  :placeholder="$t('groupAdmin.appGroupModal.descriptionPlaceholder')"
                ></textarea>
              </div>

              <label class="flex items-center space-x-3 cursor-pointer">
                <input
                  v-model="form.is_active"
                  type="checkbox"
                  class="h-4 w-4 text-green-600 focus:ring-green-500 border-gray-600 rounded bg-gray-700"
                />
                <span class="text-sm text-gray-300">
                  {{ $t('groupAdmin.appGroupModal.activeAndVisible') }}
                </span>
              </label>
            </div>

            <!-- Colonne droite : apparence -->
            <div class="space-y-4">
              <div class="section-header">
                <Icon icon="mdi:palette-outline" class="section-icon" />
                <h4 class="section-title">{{ $t('groupAdmin.appGroupModal.appearance') }}</h4>
              </div>

              <div class="card p-3">
                <div class="flex items-center space-x-3">
                  <Icon
                    :icon="form.icon || 'mdi:folder'"
                    class="h-9 w-9 flex-shrink-0"
                    :style="{ color: form.color }"
                  />
                  <div class="min-w-0">
                    <div class="font-medium text-white truncate">{{ form.name || $t('groupAdmin.appGroupModal.previewName') }}</div>
                    <div class="text-sm text-gray-400 truncate">{{ form.description || $t('groupAdmin.appGroupModal.previewDescription') }}</div>
                  </div>
                </div>
              </div>

              <div class="form-group">
                <label for="color" class="form-label">{{ $t('groupAdmin.appGroupModal.color') }}</label>
                <div class="flex items-center space-x-3">
                  <input
                    id="color"
                    v-model="form.color"
                    type="color"
                    class="h-10 w-16 rounded-lg border border-gray-600 cursor-pointer bg-gray-800"
                  />
                  <input
                    v-model="form.color"
                    type="text"
                    class="form-input flex-1"
                    placeholder="#10b981"
                  />
                </div>
              </div>

              <IconInput
                v-model="form.icon"
                :label="$t('groupAdmin.appGroupModal.icon')"
                placeholder="mdi:folder, lucide:folder, heroicons:folder"
                category="group"
                :show-suggestions="true"
              />
            </div>
          </div>

          <!-- Footer -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">{{ $t('common.requiredFieldsHint') }}</span>
            <div class="modal-footer-actions">
              <button type="button" @click="closeModal" class="btn-secondary">
                {{ $t('groupAdmin.appGroupModal.cancel') }}
              </button>
              <button type="submit" :disabled="loading" class="btn-primary">
                <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
                <Icon v-else :icon="isEdit ? 'mdi:content-save' : 'mdi:plus'" class="h-4 w-4 mr-2" />
                {{ isEdit ? $t('groupAdmin.appGroupModal.update') : $t('groupAdmin.appGroupModal.create') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import IconInput from '@/components/ui/IconInput.vue'

const { t } = useI18n()

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  appGroup: {
    type: Object,
    default: null
  },
  managedGroups: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'submit'])

const loading = ref(false)
const errors = ref({})

const form = reactive({
  name: '',
  description: '',
  color: '#10b981',
  icon: 'mdi:folder',
  order: 0,
  is_active: true,
  owner_group_id: '',
  is_private: true // Toujours privé pour les group admins
})

const isEdit = computed(() => !!props.appGroup)

// Reset form when modal opens/closes
watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
  }
})

watch(() => props.appGroup, (newVal) => {
  if (newVal) {
    Object.assign(form, {
      name: newVal.name || '',
      description: newVal.description || '',
      color: newVal.color || '#10b981',
      icon: newVal.icon || 'mdi:folder',
      order: newVal.order || 0,
      is_active: newVal.is_active ?? true,
      owner_group_id: newVal.owner_group_id || '',
      is_private: true
    })
  }
}, { immediate: true })

const resetForm = () => {
  if (props.appGroup) {
    Object.assign(form, {
      name: props.appGroup.name || '',
      description: props.appGroup.description || '',
      color: props.appGroup.color || '#10b981',
      icon: props.appGroup.icon || 'mdi:folder',
      order: props.appGroup.order || 0,
      is_active: props.appGroup.is_active ?? true,
      owner_group_id: props.appGroup.owner_group_id || '',
      is_private: true
    })
  } else {
    Object.assign(form, {
      name: '',
      description: '',
      color: '#10b981',
      icon: 'mdi:folder',
      order: 0,
      is_active: true,
      owner_group_id: props.managedGroups.length === 1 ? props.managedGroups[0].id : '',
      is_private: true
    })
  }
  errors.value = {}
}

const validateForm = () => {
  errors.value = {}

  if (!form.name.trim()) {
    errors.value.name = t('groupAdmin.appGroupModal.errors.nameRequired')
  }

  if (!form.owner_group_id) {
    errors.value.owner_group_id = t('groupAdmin.appGroupModal.errors.ownerGroupRequired')
  }

  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) return

  loading.value = true
  try {
    const formData = { ...form }
    if (isEdit.value) {
      formData.id = props.appGroup.id
    }

    emit('submit', formData)
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  resetForm()
  emit('close')
}
</script>
