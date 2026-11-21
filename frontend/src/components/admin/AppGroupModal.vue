<template>
  <div v-if="show" class="modal-overlay" @click.self="closeModal">
    <div class="modal-container" @click.self="closeModal">
      <div class="modal-panel w-full sm:max-w-5xl">
        <form class="modal-form" @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="h-10 w-10 flex-shrink-0 rounded-lg flex items-center justify-center text-white shadow-sm"
                  :style="{ backgroundColor: form.color }"
                >
                  <Icon :icon="form.icon || 'mdi:folder'" class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                  <h3 class="modal-title truncate">
                    {{ isEdit ? $t('groupAdmin.appGroupModal.title.edit') : $t('groupAdmin.appGroupModal.title.new') }}
                  </h3>
                  <p class="modal-subtitle truncate">
                    {{ isEdit ? $t('groupAdmin.appGroupModal.subtitle.edit') : $t('groupAdmin.appGroupModal.subtitle.new') }}
                  </p>
                </div>
              </div>
              <button type="button" class="btn btn-ghost p-2" :aria-label="$t('common.close')" @click="closeModal">
                <Icon icon="mdi:close" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Content : infos | apparence -->
          <div class="modal-content grid grid-cols-1 lg:grid-cols-2 gap-x-8 gap-y-6">
            <!-- Colonne gauche -->
            <div class="space-y-4">
              <div class="section-header mb-0">
                <Icon icon="mdi:information-outline" class="section-icon" />
                <h4 class="section-title">{{ $t('groupAdmin.appGroupModal.basicInfo') }}</h4>
              </div>

              <div class="grid grid-cols-[1fr_7rem] gap-4">
                <div class="form-group mb-0">
                  <label for="name" class="form-label form-label-required">{{ $t('groupAdmin.appGroupModal.groupName') }}</label>
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
                <div class="form-group mb-0">
                  <label for="order" class="form-label">{{ $t('groupAdmin.appGroupModal.displayOrder') }}</label>
                  <input id="order" v-model.number="form.order" type="number" min="0" class="form-input" placeholder="0" />
                </div>
              </div>

              <div class="form-group mb-0">
                <label for="description" class="form-label">{{ $t('groupAdmin.appGroupModal.description') }}</label>
                <textarea
                  id="description"
                  v-model="form.description"
                  rows="2"
                  class="form-textarea resize-none"
                  :placeholder="$t('groupAdmin.appGroupModal.descriptionPlaceholder')"
                ></textarea>
              </div>

              <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">{{ $t('common.active') }}</span>
                  <span class="toggle-row-desc">{{ $t('groupAdmin.appGroupModal.activeAndVisible') }}</span>
                </span>
                <input v-model="form.is_active" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>

              <!-- Admin global : groupe public ou privé au choix -->
              <label v-if="isGlobal" class="toggle-row">
                <span>
                  <span class="toggle-row-title">{{ $t('adminShared.appGroups.privateToggle') }}</span>
                  <span class="toggle-row-desc">{{ $t('adminShared.appGroups.privateToggleHelp') }}</span>
                </span>
                <input v-model="form.is_private" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>

              <!-- Groupe propriétaire : si privé (admin global), toujours (admin de groupe) -->
              <div v-if="showOwnerSelect" class="form-group mb-0">
                <label for="owner_group_id" class="form-label form-label-required">{{ $t('groupAdmin.appGroupModal.ownerGroup') }}</label>
                <select
                  id="owner_group_id"
                  v-model.number="form.owner_group_id"
                  required
                  class="form-select"
                  :disabled="isOwnerLocked"
                >
                  <option :value="null" disabled>{{ $t('groupAdmin.appGroupModal.selectGroup') }}</option>
                  <option v-for="group in ownerOptions" :key="group.id" :value="group.id">{{ group.name }}</option>
                </select>
                <p v-if="errors.owner_group_id" class="form-error">{{ errors.owner_group_id }}</p>
                <p v-else class="text-xs text-gray-400 mt-1">{{ ownerHelp }}</p>
              </div>
            </div>

            <!-- Colonne droite -->
            <div class="space-y-4">
              <div class="section-header mb-0">
                <Icon icon="mdi:palette-outline" class="section-icon" />
                <h4 class="section-title">{{ $t('groupAdmin.appGroupModal.appearance') }}</h4>
              </div>

              <div class="card p-3">
                <div class="flex items-center gap-3">
                  <Icon :icon="form.icon || 'mdi:folder'" class="h-9 w-9 flex-shrink-0" :style="{ color: form.color }" />
                  <div class="min-w-0 flex-1">
                    <div class="font-medium text-gray-900 dark:text-white truncate">
                      {{ form.name || $t('groupAdmin.appGroupModal.previewName') }}
                    </div>
                    <div class="text-sm text-gray-500 dark:text-gray-400 truncate">
                      {{ form.description || $t('groupAdmin.appGroupModal.previewDescription') }}
                    </div>
                  </div>
                  <Icon v-if="form.is_private" icon="mdi:lock" class="h-4 w-4 text-gray-400" />
                </div>
              </div>

              <div class="form-group mb-0">
                <label for="color" class="form-label">{{ $t('groupAdmin.appGroupModal.color') }}</label>
                <div class="flex items-center gap-2">
                  <input
                    id="color"
                    v-model="form.color"
                    type="color"
                    class="h-10 w-12 flex-shrink-0 rounded-lg border cursor-pointer p-1"
                    style="border-color: var(--border); background-color: var(--bg-card)"
                  />
                  <input v-model="form.color" type="text" class="form-input font-mono" placeholder="#10b981" />
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
              <button type="button" class="btn btn-secondary" @click="closeModal">{{ $t('common.cancel') }}</button>
              <button type="submit" :disabled="loading" class="btn btn-primary">
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
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import IconInput from '@/components/ui/IconInput.vue'
import { adminService } from '@/services/api'
import { useManagementScope } from '@/composables/useManagementScope'
import { GROUP_ADMIN } from '@/config/access'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  appGroup: {
    type: Object,
    default: null
  },
  // Groupes administrés (admin de groupe uniquement) : choix du groupe propriétaire
  managedGroups: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'submit'])

const { t } = useI18n()
// Admin global : public/privé et n'importe quel groupe propriétaire ;
// admin de groupe : toujours privé, propriétaire parmi ses groupes administrés
const { isGlobal } = useManagementScope(GROUP_ADMIN)

const DEFAULT_COLOR = '#10b981'
const DEFAULT_ICON = 'mdi:folder'

const loading = ref(false)
const errors = ref({})
const groups = ref([])

const form = reactive({
  name: '',
  description: '',
  color: DEFAULT_COLOR,
  icon: DEFAULT_ICON,
  order: 0,
  is_active: true,
  is_private: false,
  owner_group_id: null
})

const isEdit = computed(() => !!props.appGroup)

const showOwnerSelect = computed(() => !isGlobal.value || form.is_private)
const ownerOptions = computed(() => (isGlobal.value ? groups.value : props.managedGroups))
// L'admin de groupe ne peut pas changer le propriétaire d'un groupe existant
const isOwnerLocked = computed(() => !isGlobal.value && isEdit.value)
const ownerHelp = computed(() => {
  if (isGlobal.value) return t('adminShared.appGroups.ownerGroupManageHelp')
  const help = t('groupAdmin.appGroupModal.ownerGroupHelp')
  return isEdit.value ? `${help} ${t('groupAdmin.appGroupModal.ownerGroupEditHelp')}` : help
})

// Groupes d'utilisateurs pour le choix du propriétaire (admin global uniquement)
const loadGroups = async () => {
  if (!isGlobal.value) return
  try {
    const data = await adminService.getGroups()
    groups.value = data || []
  } catch (error) {
    console.error('Failed to load groups:', error)
  }
}

// Valeurs du formulaire pour un groupe de ressources existant ou nouveau
const buildFormValues = (appGroup) => {
  if (appGroup) {
    return {
      name: appGroup.name || '',
      description: appGroup.description || '',
      color: appGroup.color || DEFAULT_COLOR,
      icon: appGroup.icon || DEFAULT_ICON,
      order: appGroup.order || 0,
      is_active: appGroup.is_active ?? true,
      is_private: isGlobal.value ? (appGroup.is_private ?? false) : true,
      owner_group_id: appGroup.owner_group_id ? Number(appGroup.owner_group_id) : null
    }
  }
  const singleManagedGroup = !isGlobal.value && props.managedGroups.length === 1
  return {
    name: '',
    description: '',
    color: DEFAULT_COLOR,
    icon: DEFAULT_ICON,
    order: 0,
    is_active: true,
    is_private: !isGlobal.value,
    owner_group_id: singleManagedGroup ? props.managedGroups[0].id : null
  }
}

const resetForm = () => {
  Object.assign(form, buildFormValues(props.appGroup))
  errors.value = {}
}

// Réinitialise le formulaire à l'ouverture
watch(() => props.show, async (newVal) => {
  if (newVal) {
    await loadGroups() // Charger les groupes avant de remplir le formulaire
    resetForm()
  }
})

watch(() => props.appGroup, (newVal) => {
  if (newVal && props.show) {
    Object.assign(form, buildFormValues(newVal))
  }
}, { immediate: true })

const validateForm = () => {
  const nextErrors = {}

  if (!form.name.trim()) {
    nextErrors.name = t('groupAdmin.appGroupModal.errors.nameRequired')
  }

  // Un groupe privé doit avoir un groupe propriétaire
  if (form.is_private && !form.owner_group_id) {
    nextErrors.owner_group_id = t('groupAdmin.appGroupModal.errors.ownerGroupRequired')
  }

  errors.value = nextErrors
  return Object.keys(nextErrors).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) return

  loading.value = true
  try {
    const formData = isEdit.value ? { ...form, id: props.appGroup.id } : { ...form }
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
