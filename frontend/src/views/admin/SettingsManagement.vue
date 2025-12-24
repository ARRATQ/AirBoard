<template>
  <div class="content-area">
    <!-- Header -->
    <div class="page-header">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="page-title">{{ $t('settings.title') }}</h1>
          <p class="page-subtitle">{{ $t('settings.subtitle') }}</p>
        </div>
        <div class="flex gap-3">
          <button @click="showPasswordModal = true" class="btn btn-secondary">
            <Icon icon="mdi:key" class="h-4 w-4 mr-2" />
            {{ $t('settings.changePassword') }}
          </button>
          <button @click="resetSettings" class="btn btn-secondary">
            <Icon icon="mdi:restore" class="h-4 w-4 mr-2" />
            {{ $t('settings.resetToDefault') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Settings Form -->
    <div class="max-w-4xl">
      <div class="card">
        <form @submit.prevent="handleSubmit" class="space-y-8">
          <!-- Preview Section -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:eye-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.livePreview') }}</h4>
            </div>
            
            <div class="bg-gray-700 rounded-lg p-6 border-2 border-dashed border-gray-600">
              <!-- Header Preview -->
              <div class="flex items-center gap-3 mb-4">
                <div class="h-10 w-10 bg-white rounded-lg flex items-center justify-center">
                  <Icon :icon="form.app_icon || 'mdi:view-dashboard'" class="h-6 w-6 text-black" />
                </div>
                <h1 class="text-xl font-bold text-white">{{ form.app_name || 'Airboard' }}</h1>
              </div>
              
              <!-- Dashboard Preview -->
              <div class="bg-gray-800 rounded-lg p-4">
                <h2 class="text-lg font-semibold text-white mb-2">{{ form.dashboard_title || $t('settings.previewDashboardTitle') }}</h2>
                <p class="text-gray-300 text-sm mb-2">{{ form.welcome_message || $t('settings.previewWelcome') }}</p>
                <p class="text-xs text-gray-400 italic">▲ Message Dashboard (page Applications)</p>
                <p class="text-gray-300 text-sm mt-3">{{ form.home_page_message || 'Message de la page d\'accueil' }}</p>
                <p class="text-xs text-gray-400 italic">▲ Message Home (page d'accueil)</p>
                
                <!-- Sample App Cards -->
                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  <div class="bg-gray-700 rounded-lg p-4 hover:bg-gray-600 transition-colors">
              <div class="flex items-center gap-3 mb-2">
                      <div class="h-8 w-8 bg-blue-500 rounded-lg flex items-center justify-center">
                        <Icon icon="mdi:git" class="h-4 w-4 text-white" />
                      </div>
                      <div>
                        <h3 class="font-medium text-white text-sm">GitLab</h3>
                      </div>
                    </div>
                    <p class="text-xs text-gray-400">Git repository management</p>
                  </div>
                  
                  <div class="bg-gray-700 rounded-lg p-4 hover:bg-gray-600 transition-colors">
                    <div class="flex items-center gap-3 mb-2">
                      <div class="h-8 w-8 bg-green-500 rounded-lg flex items-center justify-center">
                        <Icon icon="mdi:email" class="h-4 w-4 text-white" />
                      </div>
                      <div>
                        <h3 class="font-medium text-white text-sm">Mail</h3>
                      </div>
                    </div>
                    <p class="text-xs text-gray-400">Email management</p>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- App Information -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:information-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.appInformation') }}</h4>
            </div>
            
            <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div class="form-group">
                <label for="app_name" class="form-label form-label-required">{{ $t('settings.appName') }}</label>
                <input
                  id="app_name"
                  v-model="form.app_name"
                  type="text"
                  required
                  class="form-input"
                  :placeholder="$t('settings.appNamePlaceholder')"
                />
                <p v-if="errors.app_name" class="form-error">{{ errors.app_name }}</p>
                <p class="form-help">This name appears in the header and browser title</p>
              </div>

              <IconInput
                v-model="form.app_icon"
                :label="$t('settings.appIcon')"
                placeholder="mdi:view-dashboard, lucide:layout-dashboard"
                category="general"
                :show-suggestions="true"
                required
              />
            </div>
          </div>

          <!-- Dashboard Customization -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:view-dashboard-outline" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.homePageCustomization') }}</h4>
            </div>

            <div class="space-y-4">
              <div class="form-group">
                <label for="dashboard_title" class="form-label form-label-required">{{ $t('settings.dashboardTitle') }}</label>
                <input
                  id="dashboard_title"
                  v-model="form.dashboard_title"
                  type="text"
                  required
                  class="form-input"
                  :placeholder="$t('settings.dashboardTitlePlaceholder')"
                />
                <p v-if="errors.dashboard_title" class="form-error">{{ errors.dashboard_title }}</p>
                <p class="form-help">Main title displayed on the dashboard page</p>
              </div>

              <div class="form-group">
                <label for="welcome_message" class="form-label form-label-required">{{ $t('settings.welcomeMessage') }}</label>
                <textarea
                  id="welcome_message"
                  v-model="form.welcome_message"
                  rows="3"
                  required
                  class="form-textarea"
                  :placeholder="$t('settings.welcomeMessagePlaceholder')"
                ></textarea>
                <p v-if="errors.welcome_message" class="form-error">{{ errors.welcome_message }}</p>
                <p class="form-help">Friendly message displayed on the applications dashboard page</p>
              </div>

              <div class="form-group">
                <label for="home_page_message" class="form-label form-label-required">{{ $t('settings.homePageMessage') }}</label>
                <textarea
                  id="home_page_message"
                  v-model="form.home_page_message"
                  rows="3"
                  required
                  class="form-textarea"
                  :placeholder="$t('settings.homePageMessagePlaceholder')"
                ></textarea>
                <p v-if="errors.home_page_message" class="form-error">{{ errors.home_page_message }}</p>
                <p class="form-help">Message displayed in the hero section of the home page</p>
              </div>
            </div>
          </div>

          <!-- Authentication Configuration -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:shield-account" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.authConfiguration') }}</h4>
            </div>

            <div class="space-y-4">
              <div class="form-group">
                <div class="flex items-center justify-between">
                  <div>
                    <label class="form-label">{{ $t('settings.signupEnabled') }}</label>
                    <p class="form-help">{{ $t('settings.signupEnabledHelp') }}</p>
                  </div>
                  <label class="relative inline-flex items-center cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="form.signup_enabled"
                      class="sr-only peer"
                    />
                    <div class="w-11 h-6 bg-gray-700 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-green-800 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-600"></div>
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Default Group Configuration -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:account-group" class="section-icon" />
              <h4 class="section-title">{{ $t('settings.defaultGroupConfiguration') }}</h4>
            </div>

            <div class="form-group">
              <label for="default_group" class="form-label">{{ $t('settings.defaultGroup') }}</label>
              <select
                id="default_group"
                v-model="form.default_group_id"
                class="form-input"
              >
                <option :value="null">{{ $t('settings.noDefaultGroup') }}</option>
                <option
                  v-for="group in groups"
                  :key="group.id"
                  :value="group.id"
                >
                  {{ group.name }}
                </option>
              </select>
              <p class="form-help">{{ $t('settings.defaultGroupHelp') }}</p>
            </div>
          </div>

          <!-- Danger Zone -->
          <div>
            <div class="section-header">
              <Icon icon="mdi:alert-circle" class="section-icon text-red-500" />
              <h4 class="section-title text-red-600 dark:text-red-400">{{ $t('admin.dangerZone') }}</h4>
            </div>

            <div class="bg-red-50 dark:bg-red-900/20 border-2 border-red-200 dark:border-red-800 rounded-lg p-6">
              <div class="flex items-start gap-4">
                <div class="flex-shrink-0">
                  <div class="h-12 w-12 bg-red-100 dark:bg-red-900 rounded-lg flex items-center justify-center">
                    <Icon icon="mdi:database-refresh" class="h-6 w-6 text-red-600 dark:text-red-400" />
                  </div>
                </div>
                <div class="flex-1">
                  <h5 class="text-lg font-semibold text-red-900 dark:text-red-100 mb-2">
                    {{ $t('admin.resetDatabaseTitle') }}
                  </h5>
                  <p class="text-sm text-gray-700 dark:text-gray-300 mb-4">
                    {{ $t('admin.resetDatabaseTitle') }}
                    <strong class="text-red-600 dark:text-red-400">{{ $t('admin.resetDatabaseWarning') }}</strong>
                  </p>
                  <button
                    type="button"
                    @click="showResetDatabaseModal = true"
                    :disabled="isResetting"
                    class="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-red-400 text-white rounded-lg font-medium transition-colors flex items-center gap-2"
                  >
                    <Icon icon="mdi:database-refresh" class="h-5 w-5" />
                    <span v-if="!isResetting">{{ $t('admin.resetDatabaseButton') }}</span>
                    <span v-else>{{ $t('admin.resetDatabaseInProgress') }}</span>
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex justify-end space-x-4 pt-6 border-t border-gray-700">
            <button
              type="button"
              @click="loadSettings"
              class="btn btn-secondary"
            >
              <Icon icon="mdi:refresh" class="h-4 w-4 mr-2" />
              {{ $t('settings.reload') }}
            </button>
            <button
              type="submit"
              :disabled="loading"
              class="btn btn-primary"
            >
              <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ $t('settings.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Password Change Modal -->
    <div v-if="showPasswordModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-md sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:key" class="h-6 w-6 text-green-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('settings.changePassword') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('settings.changePasswordSubtitle') }}
                </p>
              </div>
            </div>
          </div>
          
          <form @submit.prevent="handlePasswordChange" class="modal-content">
            <div class="space-y-4">
              <div class="form-group">
                <label for="old_password" class="form-label form-label-required">{{ $t('settings.currentPassword') }}</label>
                <input
                  id="old_password"
                  v-model="passwordForm.oldPassword"
                  type="password"
                  required
                  class="form-input"
                  :placeholder="$t('settings.currentPasswordPlaceholder')"
                />
                <p v-if="passwordErrors.oldPassword" class="form-error">{{ passwordErrors.oldPassword }}</p>
              </div>

              <div class="form-group">
                <label for="new_password" class="form-label form-label-required">{{ $t('settings.newPassword') }}</label>
                <input
                  id="new_password"
                  v-model="passwordForm.newPassword"
                  type="password"
                  required
                  class="form-input"
                  :placeholder="$t('settings.newPasswordPlaceholder')"
                  minlength="6"
                />
                <p v-if="passwordErrors.newPassword" class="form-error">{{ passwordErrors.newPassword }}</p>
                <p class="form-help">Minimum 6 characters</p>
              </div>

              <div class="form-group">
                <label for="confirm_password" class="form-label form-label-required">{{ $t('settings.confirmNewPassword') }}</label>
                <input
                  id="confirm_password"
                  v-model="passwordForm.confirmPassword"
                  type="password"
                  required
                  class="form-input"
                  :placeholder="$t('settings.confirmNewPasswordPlaceholder')"
                  minlength="6"
                />
                <p v-if="passwordErrors.confirmPassword" class="form-error">{{ passwordErrors.confirmPassword }}</p>
              </div>
            </div>

            <div class="modal-footer">
              <button
                type="button"
                @click="closePasswordModal"
                class="btn btn-secondary w-full sm:w-auto"
              >
                {{ $t('common.cancel') }}
              </button>
              <button
                type="submit"
                :disabled="passwordLoading"
                class="btn btn-primary w-full sm:w-auto"
              >
                <Icon v-if="passwordLoading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
                {{ $t('settings.changePassword') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Reset Confirmation Modal -->
    <div v-if="showResetModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-lg sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <Icon icon="mdi:restore" class="h-6 w-6 text-yellow-500" />
              </div>
              <div>
                <h3 class="modal-title">{{ $t('settings.resetTitle') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('settings.resetSubtitle') }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeResetModal" class="btn btn-secondary w-full sm:w-auto">
              {{ $t('common.cancel') }}
            </button>
            <button
              @click="confirmReset"
              :disabled="resetLoading"
              class="btn btn-danger w-full sm:w-auto"
            >
              <Icon v-if="resetLoading" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ $t('settings.resetButton') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Reset Database Confirmation Modal -->
    <div v-if="showResetDatabaseModal" class="modal-overlay">
      <div class="modal-container">
        <div class="modal-panel sm:max-w-md sm:w-full">
          <div class="modal-header">
            <div class="flex items-start gap-3">
              <div class="flex-shrink-0">
                <div class="h-12 w-12 bg-red-100 dark:bg-red-900 rounded-full flex items-center justify-center">
                  <Icon icon="mdi:alert" class="h-6 w-6 text-red-600 dark:text-red-400" />
                </div>
              </div>
              <div>
                <h3 class="modal-title">{{ $t('admin.confirmResetTitle') }}</h3>
                <p class="modal-subtitle">
                  {{ $t('admin.confirmResetMessage') }}
                </p>
              </div>
            </div>
          </div>

          <div class="modal-content">
            <p class="text-gray-700 dark:text-gray-300">
              <strong class="text-red-600 dark:text-red-400 block mt-2">
                {{ $t('admin.confirmResetWarning') }}
              </strong>
            </p>
          </div>

          <div class="modal-footer">
            <button
              @click="showResetDatabaseModal = false"
              class="btn btn-secondary w-full sm:w-auto"
            >
              {{ $t('admin.confirmResetCancel') }}
            </button>
            <button
              @click="handleResetDatabase"
              :disabled="isResetting"
              class="btn btn-danger w-full sm:w-auto"
            >
              <Icon v-if="isResetting" icon="mdi:loading" class="animate-spin h-4 w-4 mr-2" />
              {{ isResetting ? $t('admin.resetInProgress') : $t('admin.confirmResetButton') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { adminService } from '@/services/api'
import { authService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import IconInput from '@/components/ui/IconInput.vue'

const appStore = useAppStore()
const router = useRouter()

// State
const loading = ref(false)
const resetLoading = ref(false)
const showResetModal = ref(false)
const showPasswordModal = ref(false)
const passwordLoading = ref(false)
const passwordErrors = ref({})
const errors = ref({})
const showResetDatabaseModal = ref(false)
const isResetting = ref(false)

const form = reactive({
  app_name: '',
  app_icon: '',
  dashboard_title: '',
  welcome_message: '',
  home_page_message: '',
  signup_enabled: true,
  default_group_id: null
})

const groups = ref([])

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// Methods
const loadGroups = async () => {
  try {
    const data = await adminService.getGroups()
    groups.value = data
  } catch (error) {
    // Ne pas afficher d'erreur si l'utilisateur n'est pas authentifié (lors de la déconnexion)
    if (authStore.isAuthenticated) {
      console.error('Error loading groups:', error)
      appStore.showError('Failed to load groups')
    }
  }
}

const loadSettings = async () => {
  try {
    appStore.setLoading(true)
    const data = await adminService.getAppSettings()

    Object.assign(form, {
      app_name: data.app_name || 'Airboard',
      app_icon: data.app_icon || 'mdi:view-dashboard',
      dashboard_title: data.dashboard_title || 'Dashboard',
      welcome_message: data.welcome_message || 'Welcome to your application portal',
      home_page_message: data.home_page_message || 'Discover your personalized workspace',
      signup_enabled: data.signup_enabled !== undefined ? data.signup_enabled : true,
      default_group_id: data.default_group_id || null
    })
  } catch (error) {
    // Ne pas afficher d'erreur si l'utilisateur n'est pas authentifié (lors de la déconnexion)
    if (authStore.isAuthenticated) {
      console.error('Error loading settings:', error)
      appStore.showError('Failed to load settings')
    }
  } finally {
    appStore.setLoading(false)
  }
}

const validateForm = () => {
  errors.value = {}
  
  if (!form.app_name.trim()) {
    errors.value.app_name = 'Application name is required'
  }
  
  if (!form.app_icon.trim()) {
    errors.value.app_icon = 'Application icon is required'
  }
  
  if (!form.dashboard_title.trim()) {
    errors.value.dashboard_title = 'Dashboard title is required'
  }
  
  if (!form.welcome_message.trim()) {
    errors.value.welcome_message = 'Welcome message is required'
  }
  
  if (!form.home_page_message.trim()) {
    errors.value.home_page_message = 'Home page message is required'
  }
  
  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  loading.value = true
  try {
    await adminService.updateAppSettings(form)
    appStore.showSuccess('Settings updated successfully')
    // 🔧 FIX: Recharger les settings et mettre à jour le store global
    await loadSettings()
    appStore.invalidateSettings()
    // Optionnel: rafraîchir les settings dans le store global
    try {
      await appStore.refreshAppSettings()
    } catch (e) {
      console.warn('Could not refresh global settings:', e)
    }
  } catch (error) {
    console.error('Error updating settings:', error)
    appStore.showError('Failed to update settings')
  } finally {
    loading.value = false
  }
}

const resetSettings = () => {
  showResetModal.value = true
}

const closeResetModal = () => {
  showResetModal.value = false
}

const confirmReset = async () => {
  resetLoading.value = true
  try {
    await adminService.resetAppSettings()
    // 🔧 FIX: Recharger les settings et mettre à jour le store global
    await loadSettings()
    appStore.invalidateSettings()
    // Optionnel: rafraîchir les settings dans le store global
    try {
      await appStore.refreshAppSettings()
    } catch (e) {
      console.warn('Could not refresh global settings:', e)
    }
    appStore.showSuccess('Settings reset to defaults successfully')
    closeResetModal()
  } catch (error) {
    console.error('Error resetting settings:', error)
    appStore.showError('Failed to reset settings')
  } finally {
    resetLoading.value = false
  }
}

const validatePasswordForm = () => {
  passwordErrors.value = {}
  
  if (!passwordForm.oldPassword.trim()) {
    passwordErrors.value.oldPassword = 'Current password is required'
  }
  
  if (!passwordForm.newPassword.trim()) {
    passwordErrors.value.newPassword = 'New password is required'
  } else if (passwordForm.newPassword.length < 6) {
    passwordErrors.value.newPassword = 'Password must be at least 6 characters'
  }
  
  if (!passwordForm.confirmPassword.trim()) {
    passwordErrors.value.confirmPassword = 'Please confirm your password'
  } else if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    passwordErrors.value.confirmPassword = 'Passwords do not match'
  }
  
  return Object.keys(passwordErrors.value).length === 0
}

const handlePasswordChange = async () => {
  if (!validatePasswordForm()) return
  
  passwordLoading.value = true
  try {
    await authService.changePassword(passwordForm.oldPassword, passwordForm.newPassword)
    appStore.showSuccess('Password changed successfully')
    closePasswordModal()
  } catch (error) {
    console.error('Error changing password:', error)
    if (error.response?.status === 401) {
      passwordErrors.value.oldPassword = 'Incorrect current password'
    } else {
      appStore.showError('Failed to change password')
    }
  } finally {
    passwordLoading.value = false
  }
}

const closePasswordModal = () => {
  showPasswordModal.value = false
  Object.assign(passwordForm, {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  passwordErrors.value = {}
}

const handleResetDatabase = async () => {
  isResetting.value = true
  try {
    const response = await adminService.resetDatabase()

    // Afficher un message de succès
    alert('Base de données réinitialisée avec succès!\n\n' + response.message + '\n\nVous allez être déconnecté.')

    // Fermer le modal
    showResetDatabaseModal.value = false

    // Déconnecter l'utilisateur et rediriger vers la page de connexion
    localStorage.removeItem('airboard_token')
    localStorage.removeItem('airboard_refresh_token')
    localStorage.removeItem('airboard_user')

    // Rediriger vers la page de connexion après 1 seconde
    setTimeout(() => {
      router.push('/auth/login')
      // Recharger la page pour que le serveur recrée les données initiales
      window.location.reload()
    }, 1000)
  } catch (error) {
    console.error('Erreur lors de la réinitialisation:', error)
    alert('Erreur lors de la réinitialisation de la base de données: ' + (error.response?.data?.message || error.message))
  } finally {
    isResetting.value = false
  }
}

// Lifecycle
onMounted(async () => {
  await Promise.all([loadSettings(), loadGroups()])
})
</script>