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
                  <h3 class="modal-title truncate">{{ isEdit ? 'Edit Resource Group' : 'New Resource Group' }}</h3>
                  <p class="modal-subtitle truncate">{{ isEdit ? 'Modify the resource group settings' : 'Create a new group to organize resources' }}</p>
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

              <div class="grid grid-cols-[1fr_7rem] gap-4">
                <div class="form-group mb-0">
                  <label for="name" class="form-label form-label-required">Group Name</label>
                  <input id="name" v-model="form.name" type="text" required class="form-input" placeholder="Development Tools" />
                  <p v-if="errors.name" class="form-error">{{ errors.name }}</p>
                </div>
                <div class="form-group mb-0">
                  <label for="order" class="form-label">Order</label>
                  <input id="order" v-model.number="form.order" type="number" min="0" class="form-input" placeholder="0" />
                </div>
              </div>

              <div class="form-group mb-0">
                <label for="description" class="form-label">Description</label>
                <textarea
                  id="description"
                  v-model="form.description"
                  rows="2"
                  class="form-textarea resize-none"
                  placeholder="Describe this resource group..."
                ></textarea>
              </div>

              <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">Active</span>
                  <span class="toggle-row-desc">Visible to users on their dashboard</span>
                </span>
                <input v-model="form.is_active" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>

              <template v-if="userRole === 'admin'">
                <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">Private Resource Group</span>
                  <span class="toggle-row-desc">Can only be managed by the admins of its owner group</span>
                </span>
                <input v-model="form.is_private" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>

                <div v-if="form.is_private" class="form-group mb-0">
                  <label for="owner_group_id" class="form-label form-label-required">Owner Group</label>
                  <select id="owner_group_id" v-model.number="form.owner_group_id" required class="form-select">
                    <option :value="null" disabled>Select owner group...</option>
                    <option v-for="group in groups" :key="group.id" :value="group.id">{{ group.name }}</option>
                  </select>
                  <p v-if="errors.owner_group_id" class="form-error">{{ errors.owner_group_id }}</p>
                  <p v-else class="text-xs text-gray-400 mt-1">The group admins of this group will manage this resource group</p>
                </div>
              </template>
            </div>

            <!-- Colonne droite -->
            <div class="space-y-4">
              <div class="section-header mb-0">
                <Icon icon="mdi:palette-outline" class="section-icon" />
                <h4 class="section-title">Appearance</h4>
              </div>

              <div class="card p-3">
                <div class="flex items-center gap-3">
                  <Icon :icon="form.icon || 'mdi:folder'" class="h-9 w-9 flex-shrink-0" :style="{ color: form.color }" />
                  <div class="min-w-0 flex-1">
                    <div class="font-medium text-gray-900 dark:text-white truncate">{{ form.name || 'Group Name' }}</div>
                    <div class="text-sm text-gray-500 dark:text-gray-400 truncate">{{ form.description || 'Description...' }}</div>
                  </div>
                  <Icon v-if="form.is_private" icon="mdi:lock" class="h-4 w-4 text-gray-400" />
                </div>
              </div>

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
                  <input v-model="form.color" type="text" class="form-input font-mono" placeholder="#10b981" />
                </div>
              </div>

              <IconInput
                v-model="form.icon"
                label="Icon"
                placeholder="mdi:folder, lucide:folder, heroicons:folder"
                category="group"
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
import { useAuthStore } from '@/stores/auth'
import { adminService } from '@/services/api'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  appGroup: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'submit'])

const authStore = useAuthStore()
const userRole = computed(() => authStore.user?.role)

const loading = ref(false)
const errors = ref({})
const groups = ref([])

const form = reactive({
  name: '',
  description: '',
  color: '#10b981',
  icon: 'mdi:folder',
  order: 0,
  is_active: true,
  is_private: false,
  owner_group_id: null
})

const isEdit = computed(() => !!props.appGroup)

// Load groups for owner selection (admin only)
const loadGroups = async () => {
  if (userRole.value === 'admin') {
    try {
      const data = await adminService.getGroups()
      groups.value = data || []
    } catch (error) {
      console.error('Failed to load groups:', error)
    }
  }
}

// Reset form when modal opens/closes
watch(() => props.show, async (newVal) => {
  if (newVal) {
    await loadGroups() // Load groups first before resetting form
    resetForm()
  }
})

watch(() => props.appGroup, (newVal) => {
  if (newVal && props.show) {
    // Ensure owner_group_id is properly typed (number or null)
    const ownerGroupId = newVal.owner_group_id ? Number(newVal.owner_group_id) : null

    Object.assign(form, {
      name: newVal.name || '',
      description: newVal.description || '',
      color: newVal.color || '#10b981',
      icon: newVal.icon || 'mdi:folder',
      order: newVal.order || 0,
      is_active: newVal.is_active ?? true,
      is_private: newVal.is_private ?? false,
      owner_group_id: ownerGroupId
    })
  }
}, { immediate: true })

const resetForm = () => {
  if (props.appGroup) {
    // Ensure owner_group_id is properly typed (number or null)
    const ownerGroupId = props.appGroup.owner_group_id ? Number(props.appGroup.owner_group_id) : null

    Object.assign(form, {
      name: props.appGroup.name || '',
      description: props.appGroup.description || '',
      color: props.appGroup.color || '#10b981',
      icon: props.appGroup.icon || 'mdi:folder',
      order: props.appGroup.order || 0,
      is_active: props.appGroup.is_active ?? true,
      is_private: props.appGroup.is_private ?? false,
      owner_group_id: ownerGroupId
    })
  } else {
    Object.assign(form, {
      name: '',
      description: '',
      color: '#10b981',
      icon: 'mdi:folder',
      order: 0,
      is_active: true,
      is_private: false,
      owner_group_id: null
    })
  }
  errors.value = {}
}

const validateForm = () => {
  errors.value = {}

  if (!form.name.trim()) {
    errors.value.name = 'Name is required'
  }

  // Validate owner group for private app groups (admin only)
  if (userRole.value === 'admin' && form.is_private && !form.owner_group_id) {
    errors.value.owner_group_id = 'Owner group is required for private app groups'
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