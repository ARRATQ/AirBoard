<template>
  <div v-if="show" class="modal-overlay" @click.self="closeModal">
    <div class="modal-container" @click.self="closeModal">
      <div v-focus-trap class="modal-panel w-full sm:max-w-3xl" role="dialog" aria-modal="true" aria-labelledby="user-modal-title">
        <form class="modal-form" novalidate @submit.prevent="handleSubmit">
          <!-- Header -->
          <div class="modal-header">
            <div class="flex items-center justify-between gap-4">
              <div class="flex items-center gap-3 min-w-0">
                <div
                  class="h-10 w-10 flex-shrink-0 rounded-full flex items-center justify-center text-sm font-semibold text-white shadow-sm"
                  style="background-color: var(--accent)"
                >
                  <span v-if="initials">{{ initials }}</span>
                  <Icon v-else icon="lucide:user" class="h-5 w-5" />
                </div>
                <div class="min-w-0">
                  <h3 id="user-modal-title" class="modal-title truncate">
                    {{ isEdit ? t('userModal.titleEdit') : t('users.new') }}
                  </h3>
                  <p class="modal-subtitle truncate">
                    {{ isEdit ? t('userModal.subtitleEdit') : t('userModal.subtitleCreate') }}
                  </p>
                </div>
              </div>
              <button type="button" class="btn btn-ghost p-2" :aria-label="t('common.close')" @click="closeModal">
                <Icon icon="lucide:x" class="h-5 w-5" />
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
              <span v-if="tab.hasError" class="modal-tab-error" :aria-label="t('userModal.tabError')"></span>
            </button>
          </div>

          <!-- Content -->
          <div class="modal-content modal-body-tabbed">
            <!-- Onglet : Compte -->
            <div v-show="activeTab === 'account'" class="space-y-5">
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="form-group mb-0">
                  <label for="username" class="form-label form-label-required">{{ t('userModal.username') }}</label>
                  <input
                    id="username"
                    v-model="form.username"
                    type="text"
                    required
                    :disabled="isEdit"
                    class="form-input"
                    :class="{ 'opacity-60 cursor-not-allowed': isEdit }"
                    placeholder="johndoe"
                  />
                  <p v-if="errors.username" class="form-error">{{ errors.username }}</p>
                  <p v-else-if="isEdit" class="text-xs text-gray-400 mt-1">{{ t('userModal.usernameLocked') }}</p>
                </div>

                <div class="form-group mb-0">
                  <label for="email" class="form-label form-label-required">{{ t('userModal.email') }}</label>
                  <input
                    id="email"
                    v-model="form.email"
                    type="email"
                    required
                    class="form-input"
                    placeholder="john.doe@example.com"
                  />
                  <p v-if="errors.email" class="form-error">{{ errors.email }}</p>
                </div>

                <div class="form-group mb-0">
                  <label for="first_name" class="form-label">{{ t('userModal.firstName') }}</label>
                  <input id="first_name" v-model="form.first_name" type="text" class="form-input" placeholder="John" />
                </div>

                <div class="form-group mb-0">
                  <label for="last_name" class="form-label">{{ t('userModal.lastName') }}</label>
                  <input id="last_name" v-model="form.last_name" type="text" class="form-input" placeholder="Doe" />
                </div>

                <div v-if="!isEdit" class="form-group mb-0 md:col-span-2">
                  <label for="password" class="form-label form-label-required">{{ t('userModal.password') }}</label>
                  <div class="relative">
                    <input
                      id="password"
                      v-model="form.password"
                      :type="showPassword ? 'text' : 'password'"
                      required
                      class="form-input !pe-10"
                      placeholder="••••••••"
                      minlength="6"
                      autocomplete="new-password"
                    />
                    <button
                      type="button"
                      class="absolute inset-y-0 end-0 px-3 flex items-center text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
                      :aria-label="showPassword ? t('userModal.hidePassword') : t('userModal.showPassword')"
                      @click="showPassword = !showPassword"
                    >
                      <Icon :icon="showPassword ? 'lucide:eye-off' : 'lucide:eye'" class="h-4 w-4" />
                    </button>
                  </div>
                  <p v-if="errors.password" class="form-error">{{ errors.password }}</p>
                  <p v-else class="text-xs text-gray-400 mt-1">{{ t('userModal.passwordHint', { min: PASSWORD_MIN_LENGTH }) }}</p>
                </div>
              </div>

              <label class="toggle-row">
                <span>
                  <span class="toggle-row-title">{{ t('userModal.activeAccount') }}</span>
                  <span class="toggle-row-desc">{{ t('userModal.activeAccountDesc') }}</span>
                </span>
                <input v-model="form.is_active" type="checkbox" class="switch-input" />
                <span class="switch-track" aria-hidden="true"></span>
              </label>
            </div>

            <!-- Onglet : Rôle & groupes -->
            <div v-show="activeTab === 'access'" class="space-y-5">
              <div>
                <span class="form-label form-label-required">{{ t('userModal.role') }}</span>
                <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
                  <label
                    v-for="role in roles"
                    :key="role.value"
                    class="choice-card items-start"
                    :class="{ 'choice-card-active': form.role === role.value }"
                  >
                    <input v-model="form.role" type="radio" name="role" :value="role.value" class="mt-0.5" />
                    <span class="min-w-0">
                      <span class="flex items-center gap-1.5 text-sm font-medium text-gray-900 dark:text-white">
                        <Icon :icon="role.icon" class="h-4 w-4 text-gray-500" />
                        {{ role.label }}
                      </span>
                      <span class="block text-xs text-gray-500 dark:text-gray-400 mt-0.5">{{ role.desc }}</span>
                    </span>
                  </label>
                </div>
                <p v-if="errors.role" class="form-error">{{ errors.role }}</p>
                <p v-if="form.role === 'user'" class="text-xs text-gray-400 mt-2">
                  {{ t('userModal.groupAdminHint') }}
                </p>
              </div>

              <div>
                <div class="flex items-center justify-between mb-2">
                  <span class="form-label mb-0">{{ t('userModal.groups') }}</span>
                  <span class="text-xs text-gray-400">{{ t('userModal.groupsHint') }}</span>
                </div>
                <div v-if="groups.length === 0" class="text-sm text-gray-500 text-center py-6">
                  {{ t('userModal.noGroups') }}
                </div>
                <div v-else class="grid grid-cols-2 sm:grid-cols-3 gap-2 max-h-52 overflow-y-auto pe-1">
                  <label
                    v-for="group in groups"
                    :key="group.id"
                    class="choice-card py-2"
                    :class="{ 'choice-card-active': form.group_ids.includes(group.id) }"
                  >
                    <input v-model="form.group_ids" :value="group.id" type="checkbox" />
                    <span class="h-2.5 w-2.5 flex-shrink-0 rounded-full" :style="{ backgroundColor: group.color || '#6B7280' }"></span>
                    <span class="text-sm text-gray-700 dark:text-gray-200 truncate">{{ group.name }}</span>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Footer -->
          <div class="modal-footer-bar">
            <span class="modal-footer-hint">{{ t('common.requiredFieldsHint') }}</span>
            <div class="modal-footer-actions">
              <button type="button" class="btn btn-secondary" @click="closeModal">{{ t('common.cancel') }}</button>
              <button type="submit" :disabled="loading" class="btn btn-primary">
                <Icon v-if="loading" icon="lucide:loader-circle" class="animate-spin h-4 w-4 me-2" />
                <Icon v-else :icon="isEdit ? 'lucide:save' : 'lucide:user-plus'" class="h-4 w-4 me-2" />
                {{ isEdit ? t('common.save') : t('userModal.submitCreate') }}
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
  user: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'submit'])

const { t } = useI18n()

const USERNAME_MIN_LENGTH = 3
const PASSWORD_MIN_LENGTH = 6

const loading = ref(false)
const errors = ref({})
const groups = ref([])

const form = reactive({
  username: '',
  email: '',
  password: '',
  first_name: '',
  last_name: '',
  role: 'user',
  group_ids: [],
  is_active: true
})

const isEdit = computed(() => !!props.user)

const activeTab = ref('account')
const showPassword = ref(false)

const roles = computed(() => [
  { value: 'user', label: t('users.role_user'), icon: 'lucide:user', desc: t('userModal.roleUserDesc') },
  { value: 'editor', label: t('users.role_editor'), icon: 'lucide:pencil', desc: t('userModal.roleEditorDesc') },
  { value: 'admin', label: t('users.role_admin'), icon: 'lucide:shield-check', desc: t('userModal.roleAdminDesc') }
])

const initials = computed(() => {
  const f = form.first_name?.trim()
  const l = form.last_name?.trim()
  if (f || l) return `${f?.charAt(0) || ''}${l?.charAt(0) || ''}`.toUpperCase()
  return form.username?.trim().charAt(0).toUpperCase() || ''
})

const tabs = computed(() => [
  {
    id: 'account',
    label: t('userModal.tabAccount'),
    icon: 'lucide:id-card',
    hasError: !!(errors.value.username || errors.value.email || errors.value.password)
  },
  {
    id: 'access',
    label: t('userModal.tabAccess'),
    icon: 'lucide:shield-user',
    count: form.group_ids.length,
    hasError: !!errors.value.role
  }
])

// Load groups when component mounts
onMounted(async () => {
  try {
    const data = await adminService.getGroups()
    groups.value = Array.isArray(data) ? data : []
  } catch (error) {
    console.error('Erreur lors du chargement des groupes:', error)
    groups.value = []
  }
})

// Reset form when modal opens/closes
watch(() => props.show, (newVal) => {
  if (newVal) {
    resetForm()
  }
})

watch(() => props.user, (newVal) => {
  if (newVal) {
    Object.assign(form, {
      username: newVal.username || '',
      email: newVal.email || '',
      password: '',
      first_name: newVal.first_name || '',
      last_name: newVal.last_name || '',
      role: newVal.role || 'user',
      group_ids: newVal.groups?.map(g => g.id) || [],
      is_active: newVal.is_active ?? true
    })
  }
}, { immediate: true })

const resetForm = () => {
  if (props.user) {
    Object.assign(form, {
      username: props.user.username || '',
      email: props.user.email || '',
      password: '',
      first_name: props.user.first_name || '',
      last_name: props.user.last_name || '',
      role: props.user.role || 'user',
      group_ids: props.user.groups?.map(g => g.id) || [],
      is_active: props.user.is_active ?? true
    })
  } else {
    Object.assign(form, {
      username: '',
      email: '',
      password: '',
      first_name: '',
      last_name: '',
      role: 'user',
      group_ids: [],
      is_active: true
    })
  }
  errors.value = {}
  activeTab.value = 'account'
  showPassword.value = false
}

const validateForm = () => {
  errors.value = {}
  
  if (!form.username.trim()) {
    errors.value.username = t('userModal.errors.usernameRequired')
  } else if (form.username.length < USERNAME_MIN_LENGTH) {
    errors.value.username = t('userModal.errors.usernameTooShort', { min: USERNAME_MIN_LENGTH })
  }
  
  if (!form.email.trim()) {
    errors.value.email = t('userModal.errors.emailRequired')
  } else if (!isValidEmail(form.email)) {
    errors.value.email = t('userModal.errors.emailInvalid')
  }
  
  if (!isEdit.value) {
    if (!form.password.trim()) {
      errors.value.password = t('userModal.errors.passwordRequired')
    } else if (form.password.length < PASSWORD_MIN_LENGTH) {
      errors.value.password = t('userModal.errors.passwordTooShort', { min: PASSWORD_MIN_LENGTH })
    }
  }
  
  if (!form.role) {
    errors.value.role = t('userModal.errors.roleRequired')
  }
  
  return Object.keys(errors.value).length === 0
}

const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const handleSubmit = async () => {
  if (!validateForm()) {
    const e = errors.value
    activeTab.value = (e.username || e.email || e.password) ? 'account' : 'access'
    return
  }
  
  loading.value = true
  try {
    const formData = { ...form }
    if (isEdit.value) {
      formData.id = props.user.id
      // Don't send password if it's empty in edit mode
      if (!formData.password) {
        delete formData.password
      }
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

// Échap ferme la modale
useEscapeClose(() => props.show, () => closeModal())
</script>