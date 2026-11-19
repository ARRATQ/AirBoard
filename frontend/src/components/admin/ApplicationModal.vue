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
                  <Icon :icon="form.icon || 'mdi:application'" class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                  <h3 class="modal-title truncate">{{ isEdit ? 'Edit Resource' : 'New Resource' }}</h3>
                  <p class="modal-subtitle truncate">{{ isEdit ? 'Modify resource settings' : 'Add a new resource to the portal' }}</p>
                </div>
              </div>
              <button type="button" class="btn btn-ghost p-2" aria-label="Close" @click="closeModal">
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
                <h4 class="section-title">Basic Information</h4>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div class="form-group mb-0">
                  <label for="name" class="form-label form-label-required">Resource Name</label>
                  <input id="name" v-model="form.name" type="text" required class="form-input" placeholder="GitLab" />
                  <p v-if="errors.name" class="form-error">{{ errors.name }}</p>
                </div>
                <div class="form-group mb-0">
                  <label for="app_group_id" class="form-label form-label-required">Resource Group</label>
                  <select id="app_group_id" v-model="form.app_group_id" required class="form-select">
                    <option value="">Select a group</option>
                    <option v-for="group in props.appGroups" :key="group.id" :value="group.id">{{ group.name }}</option>
                  </select>
                  <p v-if="errors.app_group_id" class="form-error">{{ errors.app_group_id }}</p>
                </div>
              </div>

              <div class="form-group mb-0">
                <label for="url" class="form-label form-label-required">Resource URL</label>
                <div class="relative">
                  <Icon icon="mdi:link-variant" class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
                  <input id="url" v-model="form.url" type="url" required class="form-input !pl-9" placeholder="https://gitlab.example.com" />
                </div>
                <p v-if="errors.url" class="form-error">{{ errors.url }}</p>
                <p v-else class="text-xs text-gray-400 mt-1">Complete URL to the resource (with https://)</p>
              </div>

              <div class="form-group mb-0">
                <label for="description" class="form-label">Description</label>
                <textarea
                  id="description"
                  v-model="form.description"
                  rows="2"
                  class="form-textarea resize-none"
                  placeholder="Resource description..."
                ></textarea>
              </div>

              <div v-if="props.users && props.users.length" class="form-group mb-0">
                <label for="created_by_id" class="form-label">Créateur</label>
                <select id="created_by_id" v-model="form.created_by_id" class="form-select">
                  <option :value="null">— Non attribué —</option>
                  <option v-for="u in props.users" :key="u.id" :value="u.id">
                    {{ u.first_name || u.last_name ? `${u.first_name} ${u.last_name}`.trim() : u.username }}
                    <template v-if="u.first_name || u.last_name"> ({{ u.username }})</template>
                  </option>
                </select>
                <p class="text-xs text-gray-400 mt-1">Définit le créateur pour les rapports d'activité.</p>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">New tab</span>
                  <span class="toggle-row-desc">Open in a new tab</span>
                </span>
                <input v-model="form.open_in_new_tab" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>
                <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">Active</span>
                  <span class="toggle-row-desc">Visible to users</span>
                </span>
                <input v-model="form.is_active" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>
              </div>
            </div>

            <!-- Colonne droite -->
            <div class="space-y-4">
              <div class="section-header mb-0">
                <Icon icon="mdi:palette-outline" class="section-icon" />
                <h4 class="section-title">Appearance</h4>
              </div>

              <div class="card p-3">
                <div class="flex items-center gap-3">
                  <Icon :icon="form.icon || 'mdi:application'" class="h-9 w-9 flex-shrink-0" :style="{ color: form.color }" />
                  <div class="min-w-0 flex-1">
                    <div class="font-medium text-gray-900 dark:text-white truncate">{{ form.name || 'Resource Name' }}</div>
                    <div class="text-sm text-gray-500 dark:text-gray-400 truncate">{{ form.url || 'URL...' }}</div>
                  </div>
                  <Icon v-if="form.open_in_new_tab" icon="mdi:open-in-new" class="h-4 w-4 text-gray-400" />
                </div>
              </div>

              <div class="grid grid-cols-[1fr_7rem] gap-4">
                <div class="form-group mb-0">
                <label for="color" class="form-label">Color</label>
                <div class="flex items-center gap-2">
                  <input
                    id="color"
                    v-model="form.color"
                    type="color"
                    class="h-10 w-12 flex-shrink-0 rounded-lg border cursor-pointer p-1"
                    style="border-color: var(--border); background-color: var(--bg-card)"
                  />
                  <input v-model="form.color" type="text" class="form-input font-mono" placeholder="#6366f1" />
                </div>
              </div>
                <div class="form-group mb-0">
                  <label for="order" class="form-label">Order</label>
                  <input id="order" v-model.number="form.order" type="number" min="0" class="form-input" placeholder="0" />
                </div>
              </div>

              <IconInput
                v-model="form.icon"
                label="Icon"
                placeholder="mdi:application, lucide:app-window, heroicons:squares-2x2"
                category="application"
                :show-suggestions="true"
              />

            </div>
          </div>

          <!-- Footer -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">* Required fields</span>
            <div class="modal-footer-actions">
              <button type="button" class="btn btn-secondary" @click="closeModal">Cancel</button>
              <button type="submit" :disabled="loading" class="btn btn-primary">
                <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
                <Icon v-else :icon="isEdit ? 'mdi:content-save' : 'mdi:plus'" class="h-4 w-4 mr-2" />
                {{ isEdit ? 'Update' : 'Create' }}
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
import IconInput from '@/components/ui/IconInput.vue'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  application: {
    type: Object,
    default: null
  },
  appGroups: {
    type: Array,
    default: () => []
  },
  users: {
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
  url: '',
  app_group_id: '',
  color: '#6366f1',
  icon: 'mdi:application',
  order: 0,
  open_in_new_tab: true,
  is_active: true,
  created_by_id: null
})

const isEdit = computed(() => !!props.application)

watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
  }
})

watch(() => props.application, (newVal) => {
  if (newVal) {
    Object.assign(form, {
      name: newVal.name || '',
      description: newVal.description || '',
      url: newVal.url || '',
      app_group_id: newVal.app_group_id || '',
      color: newVal.color || '#6366f1',
      icon: newVal.icon || 'mdi:application',
      order: newVal.order || 0,
      open_in_new_tab: newVal.open_in_new_tab ?? true,
      is_active: newVal.is_active ?? true,
      created_by_id: newVal.created_by_id || null
    })
  }
}, { immediate: true })

const resetForm = () => {
  if (props.application) {
    Object.assign(form, {
      name: props.application.name || '',
      description: props.application.description || '',
      url: props.application.url || '',
      app_group_id: props.application.app_group_id || '',
      color: props.application.color || '#6366f1',
      icon: props.application.icon || 'mdi:application',
      order: props.application.order || 0,
      open_in_new_tab: props.application.open_in_new_tab ?? true,
      is_active: props.application.is_active ?? true,
      created_by_id: props.application.created_by_id || null
    })
  } else {
    Object.assign(form, {
      name: '',
      description: '',
      url: '',
      app_group_id: '',
      color: '#6366f1',
      icon: 'mdi:application',
      order: 0,
      open_in_new_tab: true,
      is_active: true,
      created_by_id: null
    })
  }
  errors.value = {}
}

const validateForm = () => {
  errors.value = {}

  if (!form.name.trim()) {
    errors.value.name = 'Name is required'
  }

  if (!form.url.trim()) {
    errors.value.url = 'URL is required'
  } else if (!isValidUrl(form.url)) {
    errors.value.url = 'Invalid URL format'
  }

  if (!form.app_group_id) {
    errors.value.app_group_id = 'Resource group is required'
  }

  return Object.keys(errors.value).length === 0
}

const isValidUrl = (url) => {
  try {
    new URL(url)
    return true
  } catch {
    return false
  }
}

const handleSubmit = async () => {
  if (!validateForm()) return

  loading.value = true
  try {
    const formData = { ...form }
    if (isEdit.value) {
      formData.id = props.application.id
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
