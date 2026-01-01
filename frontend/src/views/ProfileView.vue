<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ $t('profile.title') }}</h1>
          <p class="page-subtitle">{{ $t('profile.subtitle') }}</p>
        </div>
        <div class="flex gap-3">
          <button @click="showPasswordModal = true" class="btn btn-secondary">
            <Icon icon="mdi:key" class="h-4 w-4 mr-2" />
            {{ $t('profile.changePassword') }}
          </button>
        </div>
      </div>
    </div>

    <div class="max-w-4xl">
      <!-- Avatar Section -->
      <div class="card mb-6">
        <div class="section-header mb-4">
          <Icon icon="mdi:account-circle" class="section-icon" />
          <h4 class="section-title">{{ $t('profile.avatar') }}</h4>
        </div>

        <div class="flex items-center gap-6">
          <!-- Avatar Preview -->
          <div class="relative group">
            <div
              class="h-24 w-24 rounded-full flex items-center justify-center overflow-hidden border-4 border-gray-600 bg-gray-700 cursor-pointer"
              @click="triggerAvatarUpload"
            >
              <img
                v-if="avatarPreview || authStore.user?.avatar_url"
                :src="avatarPreview || getAvatarUrl(authStore.user?.avatar_url)"
                alt="Avatar"
                class="h-full w-full object-cover"
              />
              <span v-else class="text-3xl font-bold text-white">
                {{ authStore.userInitials }}
              </span>

              <!-- Overlay on hover -->
              <div class="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity rounded-full">
                <Icon icon="mdi:camera" class="h-8 w-8 text-white" />
              </div>
            </div>

            <!-- Delete button if avatar exists -->
            <button
              v-if="authStore.user?.avatar_url && !avatarPreview"
              @click.stop="confirmDeleteAvatar"
              class="absolute -top-1 -right-1 h-6 w-6 bg-red-500 hover:bg-red-600 rounded-full flex items-center justify-center text-white transition-colors"
              :title="$t('profile.deleteAvatar')"
            >
              <Icon icon="mdi:close" class="h-4 w-4" />
            </button>
          </div>

          <div class="flex-1">
            <input
              type="file"
              ref="avatarInput"
              @change="handleAvatarChange"
              accept="image/jpeg,image/png,image/gif,image/webp"
              class="hidden"
            />

            <div class="space-y-2">
              <button @click="triggerAvatarUpload" class="btn btn-secondary">
                <Icon icon="mdi:upload" class="h-4 w-4 mr-2" />
                {{ $t('profile.uploadAvatar') }}
              </button>

              <p class="text-sm text-gray-400">
                {{ $t('profile.avatarHelp') }}
              </p>
            </div>

            <!-- Upload preview actions -->
            <div v-if="avatarPreview" class="flex gap-2 mt-3">
              <button @click="uploadAvatar" :disabled="isUploading" class="btn btn-primary">
                <Icon v-if="isUploading" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
                <Icon v-else icon="mdi:check" class="h-4 w-4 mr-2" />
                {{ isUploading ? $t('profile.uploading') : $t('common.save') }}
              </button>
              <button @click="cancelAvatarUpload" class="btn btn-secondary">
                {{ $t('common.cancel') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Profile Form -->
      <div class="card">
        <form @submit.prevent="handleSubmit" class="space-y-6">
          <!-- Account Information (Read-only) -->
          <div>
            <div class="section-header mb-4">
              <Icon icon="mdi:account-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('profile.accountInfo') }}</h4>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="form-group">
                <label class="form-label">{{ $t('profile.username') }}</label>
                <input
                  :value="authStore.user?.username"
                  type="text"
                  disabled
                  class="form-input bg-gray-700 cursor-not-allowed"
                />
                <p class="form-help">{{ $t('profile.usernameReadonly') }}</p>
              </div>

              <div class="form-group">
                <label class="form-label">{{ $t('profile.email') }}</label>
                <input
                  :value="authStore.user?.email"
                  type="email"
                  disabled
                  class="form-input bg-gray-700 cursor-not-allowed"
                />
                <p class="form-help">{{ $t('profile.emailReadonly') }}</p>
              </div>

              <div class="form-group">
                <label class="form-label">{{ $t('profile.role') }}</label>
                <input
                  :value="getRoleLabel(authStore.user?.role)"
                  type="text"
                  disabled
                  class="form-input bg-gray-700 cursor-not-allowed"
                />
              </div>

              <div class="form-group">
                <label class="form-label">{{ $t('profile.memberSince') }}</label>
                <input
                  :value="formatDate(authStore.user?.created_at)"
                  type="text"
                  disabled
                  class="form-input bg-gray-700 cursor-not-allowed"
                />
              </div>
            </div>
          </div>

          <!-- Personal Information (Editable) -->
          <div>
            <div class="section-header mb-4">
              <Icon icon="mdi:card-account-details-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('profile.personalInfo') }}</h4>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="form-group">
                <label for="first_name" class="form-label">{{ $t('profile.firstName') }}</label>
                <input
                  id="first_name"
                  v-model="form.first_name"
                  type="text"
                  class="form-input"
                  :placeholder="$t('profile.firstNamePlaceholder')"
                  maxlength="50"
                />
              </div>

              <div class="form-group">
                <label for="last_name" class="form-label">{{ $t('profile.lastName') }}</label>
                <input
                  id="last_name"
                  v-model="form.last_name"
                  type="text"
                  class="form-input"
                  :placeholder="$t('profile.lastNamePlaceholder')"
                  maxlength="50"
                />
              </div>

              <div class="form-group">
                <label for="phone" class="form-label">{{ $t('profile.phone') }}</label>
                <input
                  id="phone"
                  v-model="form.phone"
                  type="tel"
                  class="form-input"
                  :placeholder="$t('profile.phonePlaceholder')"
                  maxlength="20"
                />
              </div>

              <div class="form-group">
                <label for="location" class="form-label">{{ $t('profile.location') }}</label>
                <input
                  id="location"
                  v-model="form.location"
                  type="text"
                  class="form-input"
                  :placeholder="$t('profile.locationPlaceholder')"
                  maxlength="100"
                />
              </div>
            </div>
          </div>

          <!-- Professional Information -->
          <div>
            <div class="section-header mb-4">
              <Icon icon="mdi:briefcase-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('profile.professionalInfo') }}</h4>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="form-group">
                <label for="job_title" class="form-label">{{ $t('profile.jobTitle') }}</label>
                <input
                  id="job_title"
                  v-model="form.job_title"
                  type="text"
                  class="form-input"
                  :placeholder="$t('profile.jobTitlePlaceholder')"
                  maxlength="100"
                />
              </div>

              <div class="form-group">
                <label for="department" class="form-label">{{ $t('profile.department') }}</label>
                <input
                  id="department"
                  v-model="form.department"
                  type="text"
                  class="form-input"
                  :placeholder="$t('profile.departmentPlaceholder')"
                  maxlength="100"
                />
              </div>
            </div>
          </div>

          <!-- Groups -->
          <div v-if="authStore.user?.groups?.length > 0">
            <div class="section-header mb-4">
              <Icon icon="mdi:account-group-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('profile.groups') }}</h4>
            </div>

            <div class="flex flex-wrap gap-2">
              <span
                v-for="group in authStore.user.groups"
                :key="group.id"
                class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium"
                :style="{ backgroundColor: group.color + '20', color: group.color }"
              >
                {{ group.name }}
              </span>
            </div>
          </div>

          <!-- Submit Button -->
          <div class="flex justify-end pt-4 border-t border-gray-700">
            <button
              type="submit"
              :disabled="isSaving"
              class="btn btn-primary"
            >
              <Icon v-if="isSaving" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              <Icon v-else icon="mdi:content-save" class="h-4 w-4 mr-2" />
              {{ isSaving ? $t('common.saving') : $t('common.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Password Change Modal -->
    <Teleport to="body">
      <div v-if="showPasswordModal" class="modal-overlay" @click.self="showPasswordModal = false">
        <div class="modal-container max-w-md">
          <div class="modal-header">
            <h3 class="modal-title">{{ $t('profile.changePassword') }}</h3>
            <button @click="showPasswordModal = false" class="modal-close">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <form @submit.prevent="handlePasswordChange" class="modal-body space-y-4">
            <div class="form-group">
              <label for="current_password" class="form-label form-label-required">
                {{ $t('profile.currentPassword') }}
              </label>
              <div class="relative">
                <input
                  id="current_password"
                  v-model="passwordForm.current_password"
                  :type="showCurrentPassword ? 'text' : 'password'"
                  required
                  class="form-input pr-10"
                  :placeholder="$t('profile.currentPasswordPlaceholder')"
                />
                <button
                  type="button"
                  @click="showCurrentPassword = !showCurrentPassword"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white"
                >
                  <Icon :icon="showCurrentPassword ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
                </button>
              </div>
            </div>

            <div class="form-group">
              <label for="new_password" class="form-label form-label-required">
                {{ $t('profile.newPassword') }}
              </label>
              <div class="relative">
                <input
                  id="new_password"
                  v-model="passwordForm.new_password"
                  :type="showNewPassword ? 'text' : 'password'"
                  required
                  minlength="6"
                  class="form-input pr-10"
                  :placeholder="$t('profile.newPasswordPlaceholder')"
                />
                <button
                  type="button"
                  @click="showNewPassword = !showNewPassword"
                  class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-white"
                >
                  <Icon :icon="showNewPassword ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
                </button>
              </div>
              <p class="form-help">{{ $t('profile.passwordMinLength') }}</p>
            </div>

            <div class="form-group">
              <label for="confirm_password" class="form-label form-label-required">
                {{ $t('profile.confirmPassword') }}
              </label>
              <input
                id="confirm_password"
                v-model="passwordForm.confirm_password"
                type="password"
                required
                class="form-input"
                :placeholder="$t('profile.confirmPasswordPlaceholder')"
              />
              <p v-if="passwordForm.new_password && passwordForm.confirm_password && passwordForm.new_password !== passwordForm.confirm_password" class="form-error">
                {{ $t('profile.passwordMismatch') }}
              </p>
            </div>

            <div class="modal-footer">
              <button type="button" @click="showPasswordModal = false" class="btn btn-secondary">
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="isChangingPassword || !isPasswordFormValid"
                class="btn btn-primary"
              >
                <Icon v-if="isChangingPassword" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
                {{ isChangingPassword ? $t('common.saving') : $t('profile.changePassword') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Delete Avatar Confirmation Modal -->
    <Teleport to="body">
      <div v-if="showDeleteAvatarModal" class="modal-overlay" @click.self="showDeleteAvatarModal = false">
        <div class="modal-container max-w-md">
          <div class="modal-header">
            <h3 class="modal-title">{{ $t('profile.deleteAvatarTitle') }}</h3>
            <button @click="showDeleteAvatarModal = false" class="modal-close">
              <Icon icon="mdi:close" class="h-5 w-5" />
            </button>
          </div>

          <div class="modal-body">
            <p class="text-gray-300">{{ $t('profile.deleteAvatarConfirm') }}</p>
          </div>

          <div class="modal-footer">
            <button @click="showDeleteAvatarModal = false" class="btn btn-secondary">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="deleteAvatar"
              :disabled="isDeletingAvatar"
              class="btn btn-danger"
            >
              <Icon v-if="isDeletingAvatar" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
              {{ isDeletingAvatar ? $t('common.deleting') : $t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { authService } from '@/services/api'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const authStore = useAuthStore()

// Form state
const form = reactive({
  first_name: '',
  last_name: '',
  phone: '',
  department: '',
  job_title: '',
  location: ''
})

// Password form
const passwordForm = reactive({
  current_password: '',
  new_password: '',
  confirm_password: ''
})

// UI state
const isSaving = ref(false)
const isUploading = ref(false)
const isDeletingAvatar = ref(false)
const isChangingPassword = ref(false)
const showPasswordModal = ref(false)
const showDeleteAvatarModal = ref(false)
const showCurrentPassword = ref(false)
const showNewPassword = ref(false)
const avatarInput = ref(null)
const avatarPreview = ref(null)
const selectedAvatarFile = ref(null)

// Computed
const isPasswordFormValid = computed(() => {
  return passwordForm.current_password &&
         passwordForm.new_password &&
         passwordForm.new_password.length >= 6 &&
         passwordForm.new_password === passwordForm.confirm_password
})

// Methods
const loadProfile = () => {
  if (authStore.user) {
    form.first_name = authStore.user.first_name || ''
    form.last_name = authStore.user.last_name || ''
    form.phone = authStore.user.phone || ''
    form.department = authStore.user.department || ''
    form.job_title = authStore.user.job_title || ''
    form.location = authStore.user.location || ''
  }
}

const getAvatarUrl = (url) => {
  if (!url) return null
  if (url.startsWith('http')) return url
  return url
}

const getRoleLabel = (role) => {
  const labels = {
    admin: t('users.role_admin'),
    group_admin: t('profile.roleGroupAdmin'),
    editor: t('users.role_editor'),
    user: t('users.role_user')
  }
  return labels[role] || role
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString()
}

const triggerAvatarUpload = () => {
  avatarInput.value?.click()
}

const handleAvatarChange = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // Validate file type
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!allowedTypes.includes(file.type)) {
    alert(t('profile.avatarTypeError'))
    return
  }

  // Validate file size (max 5MB)
  if (file.size > 5 * 1024 * 1024) {
    alert(t('profile.avatarSizeError'))
    return
  }

  selectedAvatarFile.value = file

  // Create preview
  const reader = new FileReader()
  reader.onload = (e) => {
    avatarPreview.value = e.target.result
  }
  reader.readAsDataURL(file)
}

const cancelAvatarUpload = () => {
  avatarPreview.value = null
  selectedAvatarFile.value = null
  if (avatarInput.value) {
    avatarInput.value.value = ''
  }
}

const uploadAvatar = async () => {
  if (!selectedAvatarFile.value) return

  try {
    isUploading.value = true
    await authStore.uploadAvatar(selectedAvatarFile.value)
    avatarPreview.value = null
    selectedAvatarFile.value = null
    if (avatarInput.value) {
      avatarInput.value.value = ''
    }
  } catch (error) {
    console.error('Error uploading avatar:', error)
    alert(t('profile.avatarUploadError'))
  } finally {
    isUploading.value = false
  }
}

const confirmDeleteAvatar = () => {
  showDeleteAvatarModal.value = true
}

const deleteAvatar = async () => {
  try {
    isDeletingAvatar.value = true
    await authStore.deleteAvatar()
    showDeleteAvatarModal.value = false
  } catch (error) {
    console.error('Error deleting avatar:', error)
    alert(t('profile.avatarDeleteError'))
  } finally {
    isDeletingAvatar.value = false
  }
}

const handleSubmit = async () => {
  try {
    isSaving.value = true
    await authStore.saveProfile({
      first_name: form.first_name,
      last_name: form.last_name,
      phone: form.phone,
      department: form.department,
      job_title: form.job_title,
      location: form.location
    })
  } catch (error) {
    console.error('Error saving profile:', error)
    alert(t('profile.saveError'))
  } finally {
    isSaving.value = false
  }
}

const handlePasswordChange = async () => {
  if (!isPasswordFormValid.value) return

  try {
    isChangingPassword.value = true
    await authService.changePassword(
      passwordForm.current_password,
      passwordForm.new_password
    )
    showPasswordModal.value = false
    passwordForm.current_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    alert(t('profile.passwordChanged'))
  } catch (error) {
    console.error('Error changing password:', error)
    if (error.response?.status === 401) {
      alert(t('profile.currentPasswordIncorrect'))
    } else {
      alert(t('profile.passwordChangeError'))
    }
  } finally {
    isChangingPassword.value = false
  }
}

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.section-header {
  @apply flex items-center gap-2 mb-4;
}

.section-icon {
  @apply h-5 w-5 text-blue-400;
}

.section-title {
  @apply text-lg font-semibold text-white;
}
</style>
