<template>
  <div v-if="show" class="modal-overlay" @click.self="closeModal">
    <div class="modal-container" @click.self="closeModal">
      <div v-focus-trap class="modal-panel w-full sm:max-w-3xl" role="dialog" aria-modal="true" aria-labelledby="group-modal-title">
        <form class="modal-form" novalidate @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="h-10 w-10 flex-shrink-0 rounded-lg flex items-center justify-center text-white shadow-sm"
                  :style="{ backgroundColor: form.color || '#6B7280' }"
                >
                  <Icon icon="mdi:account-group" class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                  <h3 id="group-modal-title" class="modal-title truncate">
                    {{ isEdit ? t('groupModal.titleEdit') : t('groupModal.titleCreate') }}
                  </h3>
                  <p class="modal-subtitle truncate">
                    {{ isEdit ? t('groupModal.subtitleEdit') : t('groupModal.subtitleCreate') }}
                  </p>
                </div>
              </div>
              <button type="button" class="btn btn-ghost p-2" :aria-label="t('common.close')" @click="closeModal">
                <Icon icon="mdi:close" class="h-5 w-5" />
              </button>
            </div>
          </div>

          <!-- Tabs -->
          <div class="modal-tabs" role="tablist">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              type="button"
              role="tab"
              :aria-selected="activeTab === tab.id"
              class="modal-tab"
              :class="{ 'modal-tab-active': activeTab === tab.id }"
              @click="activeTab = tab.id"
            >
              <Icon :icon="tab.icon" class="h-4 w-4" />
              {{ tab.label }}
              <span v-if="tab.count !== undefined" class="modal-tab-count">{{ tab.count }}</span>
              <span v-if="tab.hasError" class="modal-tab-error" :aria-label="t('groupModal.tabError')"></span>
            </button>
          </div>

          <!-- Content -->
          <div class="modal-content modal-body-tabbed">
            <!-- Onglet : Général -->
            <div v-show="activeTab === 'general'" class="space-y-5">
              <div class="grid grid-cols-1 md:grid-cols-[1fr_15rem] gap-4">
                <div class="form-group mb-0">
                  <label for="name" class="form-label form-label-required">{{ t('groupModal.name') }}</label>
                  <input
                    id="name"
                    v-model="form.name"
                    type="text"
                    required
                    class="form-input"
                    :placeholder="t('groupModal.namePlaceholder')"
                  />
                  <p v-if="errors.name" class="form-error">{{ errors.name }}</p>
                </div>

                <div class="form-group mb-0">
                  <label for="color" class="form-label">{{ t('groupModal.color') }}</label>
                  <div class="flex items-center gap-2">
                    <input
                      id="color"
                      v-model="form.color"
                      type="color"
                      class="h-10 w-12 flex-shrink-0 rounded-lg border cursor-pointer p-1"
                      style="border-color: var(--border); background-color: var(--bg-card)"
                    />
                    <input v-model="form.color" type="text" class="form-input font-mono" placeholder="#6B7280" />
                  </div>
                </div>
              </div>

              <div class="flex flex-wrap gap-2">
                <button
                  v-for="c in presetColors"
                  :key="c"
                  type="button"
                  class="h-6 w-6 rounded-full border-2 transition-transform hover:scale-110 focus:outline-none focus-visible:ring-2"
                  :class="form.color?.toLowerCase() === c.toLowerCase() ? 'border-gray-900 dark:border-white scale-110' : 'border-transparent'"
                  :style="{ backgroundColor: c }"
                  :aria-label="t('groupModal.colorSwatch', { color: c })"
                  @click="form.color = c"
                ></button>
              </div>

              <div class="form-group mb-0">
                <label for="description" class="form-label">{{ t('common.description') }}</label>
                <textarea
                  id="description"
                  v-model="form.description"
                  rows="3"
                  class="form-textarea resize-none"
                  :placeholder="t('groupModal.descriptionPlaceholder')"
                ></textarea>
              </div>

              <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">{{ t('groupModal.activeGroup') }}</span>
                  <span class="toggle-row-desc">{{ t('groupModal.activeGroupDesc') }}</span>
                </span>
                <input v-model="form.is_active" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>
            </div>

            <!-- Onglet : Accès -->
            <div v-show="activeTab === 'access'" class="flex flex-col h-full gap-3">
              <div class="flex flex-wrap items-center gap-3">
                <div class="search-container flex-1 min-w-[12rem]">
                  <Icon icon="mdi:magnify" class="search-icon" />
                  <input
                    v-model="appGroupSearch"
                    type="search"
                    class="form-input search-input"
                    :placeholder="t('groupModal.searchAppGroups')"
                  />
                </div>
                <div class="flex items-center gap-3">
                  <button type="button" class="btn-link" @click="selectAllAppGroups">{{ t('groupModal.selectAll') }}</button>
                  <span class="text-gray-300 dark:text-gray-600">|</span>
                  <button type="button" class="btn-link" @click="form.app_group_ids = []">{{ t('groupModal.selectNone') }}</button>
                </div>
              </div>

              <div v-if="appGroups.length === 0" class="empty-state flex-1 flex flex-col items-center justify-center text-center">
                <Icon icon="mdi:folder-open-outline" class="h-12 w-12 text-gray-400 mb-2" />
                <p class="text-sm text-gray-500">{{ t('groupModal.noAppGroups') }}</p>
                <p class="text-xs text-gray-400">{{ t('groupModal.noAppGroupsHint') }}</p>
              </div>

              <div v-else class="flex-1 min-h-0 overflow-y-auto pe-1">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <label
                    v-for="appGroup in filteredAppGroups"
                    :key="appGroup.id"
                    class="choice-card"
                    :class="{ 'choice-card-active': form.app_group_ids.includes(appGroup.id) }"
                  >
                    <input v-model="form.app_group_ids" :value="appGroup.id" type="checkbox" />
                    <span
                      class="h-8 w-8 flex-shrink-0 rounded-lg flex items-center justify-center"
                      :style="{ backgroundColor: appGroup.color || '#6B7280' }"
                    >
                      <Icon :icon="appGroup.icon || 'mdi:folder'" class="h-4 w-4 text-white" />
                    </span>
                    <span class="min-w-0 flex-1">
                      <span class="flex items-center gap-1.5 text-sm font-medium text-gray-900 dark:text-white">
                        <span class="truncate">{{ appGroup.name }}</span>
                        <Icon v-if="appGroup.is_private" icon="mdi:lock" class="h-3.5 w-3.5 flex-shrink-0 text-gray-400" />
                      </span>
                      <span class="block text-xs text-gray-500 dark:text-gray-400">
                        {{ t('groupModal.resourcesCount', appGroup.applications?.length || 0) }}
                      </span>
                    </span>
                  </label>
                </div>
                <p v-if="filteredAppGroups.length === 0" class="text-sm text-gray-500 text-center py-8">
                  {{ t('groupModal.noResults', { query: appGroupSearch }) }}
                </p>
              </div>
            </div>

            <!-- Onglet : Membres (édition uniquement) -->
            <div v-if="isEdit" v-show="activeTab === 'members'" class="h-full">
              <div v-if="!form.users?.length" class="h-full flex flex-col items-center justify-center text-center">
                <Icon icon="mdi:account-off-outline" class="h-12 w-12 text-gray-400 mb-2" />
                <p class="text-sm text-gray-500">{{ t('groupModal.noMembers') }}</p>
                <p class="text-xs text-gray-400">{{ t('groupModal.noMembersHint') }}</p>
              </div>
              <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <div
                  v-for="user in form.users"
                  :key="user.id"
                  class="flex items-center gap-3 p-2.5 rounded-lg border"
                  style="border-color: var(--border); background-color: var(--bg-surface)"
                >
                  <div
                    class="h-8 w-8 flex-shrink-0 rounded-full flex items-center justify-center text-xs font-semibold text-white"
                    :style="{ backgroundColor: form.color || '#6B7280' }"
                  >
                    {{ getUserInitials(user) }}
                  </div>
                  <div class="min-w-0">
                    <div class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ getUserDisplayName(user) }}</div>
                    <div class="text-xs text-gray-500 truncate">{{ user.email }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">
              {{ t('common.requiredFieldsHint') }} · {{ t('groupModal.selectedCount', form.app_group_ids.length) }}
            </span>
            <div class="modal-footer-actions">
              <button type="button" class="btn btn-secondary" @click="closeModal">{{ t('common.cancel') }}</button>
              <button type="submit" :disabled="loading" class="btn btn-primary">
                <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-4 w-4 me-2" />
                <Icon v-else :icon="isEdit ? 'mdi:content-save' : 'mdi:plus'" class="h-4 w-4 me-2" />
                {{ isEdit ? t('common.save') : t('groupModal.submitCreate') }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { adminService } from '@/services/api'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  group: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'submit'])

const { t } = useI18n()

const loading = ref(false)
const errors = ref({})
const appGroups = ref([])

const form = reactive({
  name: '',
  description: '',
  color: '#6B7280',
  app_group_ids: [],
  users: [],
  is_active: true
})

const isEdit = computed(() => !!props.group)

const activeTab = ref('general')
const appGroupSearch = ref('')

const presetColors = ['#6B7280', '#d97757', '#ef4444', '#f59e0b', '#10b981', '#06b6d4', '#3b82f6', '#8b5cf6', '#ec4899']

const tabs = computed(() => {
  const list = [
    { id: 'general', label: t('groupModal.tabGeneral'), icon: 'mdi:information-outline', hasError: !!errors.value.name },
    { id: 'access', label: t('groupModal.tabAccess'), icon: 'mdi:shield-key-outline', count: form.app_group_ids.length }
  ]
  if (isEdit.value) {
    list.push({ id: 'members', label: t('groupModal.tabMembers'), icon: 'mdi:account-multiple-outline', count: form.users?.length || 0 })
  }
  return list
})

const filteredAppGroups = computed(() => {
  const q = appGroupSearch.value.trim().toLowerCase()
  if (!q) return appGroups.value
  return appGroups.value.filter(ag => ag.name?.toLowerCase().includes(q))
})

const selectAllAppGroups = () => {
  const ids = new Set([...form.app_group_ids, ...filteredAppGroups.value.map(ag => ag.id)])
  form.app_group_ids = [...ids]
}

// Load app groups when component mounts
onMounted(async () => {
  try {
    const data = await adminService.getAppGroups()
    appGroups.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Erreur lors du chargement des groupes d\'applications:', error)
    appGroups.value = []
  }
})

// Reset form when modal opens/closes
watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
  }
})

watch(() => props.group, (newVal) => {
  if (newVal) {
    Object.assign(form, {
      name: newVal.name || '',
      description: newVal.description || '',
      color: newVal.color || '#6B7280',
      app_group_ids: newVal.app_groups?.map(ag => ag.id) || [],
      users: newVal.users || [],
      is_active: newVal.is_active ?? true
    })
  }
}, { immediate: true })

const resetForm = () => {
  if (props.group) {
    Object.assign(form, {
      name: props.group.name || '',
      description: props.group.description || '',
      color: props.group.color || '#6B7280',
      app_group_ids: props.group.app_groups?.map(ag => ag.id) || [],
      users: props.group.users || [],
      is_active: props.group.is_active ?? true
    })
  } else {
    Object.assign(form, {
      name: '',
      description: '',
      color: '#6B7280',
      app_group_ids: [],
      users: [],
      is_active: true
    })
  }
  errors.value = {}
  activeTab.value = 'general'
  appGroupSearch.value = ''
}

const getUserInitials = (user) => {
  if (user.first_name && user.last_name) {
    return `${user.first_name.charAt(0)}${user.last_name.charAt(0)}`.toUpperCase()
  }
  return user.username.charAt(0).toUpperCase()
}

const getUserDisplayName = (user) => {
  if (user.first_name && user.last_name) {
    return `${user.first_name} ${user.last_name}`
  }
  return user.username
}

const validateForm = () => {
  errors.value = {}
  
  if (!form.name.trim()) {
    errors.value.name = t('groupModal.errors.nameRequired')
  }
  
  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) {
    activeTab.value = 'general'
    return
  }
  
  loading.value = true
  try {
    const formData = { ...form }
    if (isEdit.value) {
      formData.id = props.group.id
    }
    
    // Remove users from form data as it's read-only
    delete formData.users
    
    emit('submit', formData)
  } finally {
    loading.value = false
  }
}

const closeModal = () => {
  resetForm()
  emit('close')
}

// Échap ferme la modale
useEscapeClose(() => props.show, () => closeModal())
</script>