<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Card with modern styling -->
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8 space-y-8">
        <!-- Header -->
        <div class="text-center space-y-4">
          <div class="flex items-center justify-center">
            <div class="h-16 w-16 bg-gradient-to-br from-primary-400 to-primary-600 rounded-2xl flex items-center justify-center shadow-lg transform hover:scale-105 transition-transform duration-200">
              <Icon icon="lucide:layout-dashboard" class="h-8 w-8 text-white" />
            </div>
          </div>
          <div>
            <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('auth.register.title') }}</h1>
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">{{ t('auth.register.subtitle') }}</p>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleSubmit" class="space-y-5">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="first_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {{ t('auth.fields.firstName') }}
              </label>
              <input
                id="first_name"
                v-model="form.first_name"
                type="text"
                class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
                placeholder="John"
                :disabled="loading"
              />
            </div>

            <div>
              <label for="last_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                {{ t('auth.fields.lastName') }}
              </label>
              <input
                id="last_name"
                v-model="form.last_name"
                type="text"
                class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
                placeholder="Doe"
                :disabled="loading"
              />
            </div>
          </div>

          <div>
            <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('auth.fields.username') }} <span class="text-red-500" aria-hidden="true">*</span>
            </label>
            <input
              id="username"
              v-model="form.username"
              type="text"
              required
              class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
              placeholder="johndoe"
              :disabled="loading"
            />
            <p v-if="errors.username" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.username }}</p>
          </div>

          <div>
            <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('auth.fields.email') }} <span class="text-red-500" aria-hidden="true">*</span>
            </label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              required
              class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
              placeholder="john@example.com"
              :disabled="loading"
            />
            <p v-if="errors.email" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.email }}</p>
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('auth.fields.password') }} <span class="text-red-500" aria-hidden="true">*</span>
            </label>
            <div class="relative">
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                required
                class="w-full px-4 py-3 pe-12 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
                :placeholder="t('auth.fields.passwordPlaceholder')"
                autocomplete="new-password"
                :disabled="loading"
                minlength="8"
              />
              <button
                type="button"
                :aria-label="showPassword ? t('auth.fields.hidePassword') : t('auth.fields.showPassword')"
                :aria-pressed="showPassword"
                @click="showPassword = !showPassword"
                class="absolute end-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors duration-200"
              >
                <Icon :icon="showPassword ? 'lucide:eye-off' : 'lucide:eye'" class="h-5 w-5" />
              </button>
            </div>
            <p v-if="errors.password" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.password }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('auth.register.passwordHint') }}</p>
          </div>

          <div>
            <label for="password_confirmation" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('auth.fields.passwordConfirm') }} <span class="text-red-500" aria-hidden="true">*</span>
            </label>
            <div class="relative">
              <input
                id="password_confirmation"
                v-model="form.password_confirmation"
                :type="showConfirmPassword ? 'text' : 'password'"
                required
                class="w-full px-4 py-3 pe-12 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-transparent transition-all duration-200"
                :placeholder="t('auth.fields.passwordConfirmPlaceholder')"
                autocomplete="new-password"
                :disabled="loading"
              />
              <button
                type="button"
                :aria-label="showConfirmPassword ? t('auth.fields.hidePassword') : t('auth.fields.showPassword')"
                :aria-pressed="showConfirmPassword"
                @click="showConfirmPassword = !showConfirmPassword"
                class="absolute end-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors duration-200"
              >
                <Icon :icon="showConfirmPassword ? 'lucide:eye-off' : 'lucide:eye'" class="h-5 w-5" />
              </button>
            </div>
            <p v-if="errors.password_confirmation" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.password_confirmation }}</p>
          </div>

          <div class="flex items-start pt-2">
            <label class="flex items-start cursor-pointer group">
              <input
                v-model="form.terms"
                type="checkbox"
                required
                class="h-4 w-4 mt-0.5 text-primary-600 focus:ring-primary-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 cursor-pointer"
              />
              <span class="ms-2 text-sm text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-200 transition-colors duration-200">
                <!-- Pas encore de page de conditions ni de confidentialité : texte simple plutôt qu'un lien vers « # ». -->
                <i18n-t keypath="auth.register.terms" tag="span" scope="global">
                  <template #terms>
                    <strong class="text-gray-800 dark:text-gray-200 font-medium">{{ t('auth.register.termsLink') }}</strong>
                  </template>
                  <template #privacy>
                    <strong class="text-gray-800 dark:text-gray-200 font-medium">{{ t('auth.register.privacyLink') }}</strong>
                  </template>
                </i18n-t>
              </span>
            </label>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full flex items-center justify-center px-4 py-3 border border-transparent rounded-xl text-sm font-medium text-white bg-gradient-to-r from-primary-500 to-primary-600 hover:from-primary-600 hover:to-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
          >
            <Icon v-if="loading" icon="lucide:loader-circle" class="animate-spin h-5 w-5 me-2" />
            <span>{{ loading ? t('auth.register.submitting') : t('auth.register.submit') }}</span>
          </button>

          <div class="text-center pt-4">
            <span class="text-sm text-gray-600 dark:text-gray-400">{{ t('auth.register.haveAccount') }} </span>
            <router-link
              to="/auth/login"
              class="ms-1 text-sm text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 font-semibold transition-colors duration-200"
            >
              {{ t('auth.register.signIn') }}
            </router-link>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { authService } from '@/services/api'
import { apiErrorMessage } from '@/services/apiError'
import { passwordProblem, PASSWORD_MIN_LENGTH, PASSWORD_MAX_LENGTH } from '@/utils/passwordPolicy'

const { t } = useI18n()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const loading = ref(false)
const showPassword = ref(false)
const showConfirmPassword = ref(false)
const errors = ref({})

// Vérifier si l'inscription est activée au chargement
onMounted(async () => {
  try {
    const data = await authService.getSignupStatus()
    if (!data.signup_enabled) {
      appStore.showError(t('auth.register.disabled'))
      router.push('/auth/login')
    }
  } catch (error) {
    console.error('Error checking signup status:', error)
  }
})

const form = reactive({
  first_name: '',
  last_name: '',
  username: '',
  email: '',
  password: '',
  password_confirmation: '',
  terms: false
})

const validateForm = () => {
  errors.value = {}
  
  if (!form.username.trim()) {
    errors.value.username = t('auth.validation.usernameRequired')
  } else if (form.username.length < 3) {
    errors.value.username = t('auth.validation.usernameMin', { min: 3 })
  }
  
  if (!form.email.trim()) {
    errors.value.email = t('auth.validation.emailRequired')
  } else if (!isValidEmail(form.email)) {
    errors.value.email = t('auth.validation.emailInvalid')
  }
  
  if (!form.password.trim()) {
    errors.value.password = t('auth.validation.passwordRequired')
  } else {
    const problem = passwordProblem(form.password)
    if (problem) errors.value.password = passwordMessage(problem)
  }
  
  if (form.password !== form.password_confirmation) {
    errors.value.password_confirmation = t('auth.validation.passwordMismatch')
  }
  
  return Object.keys(errors.value).length === 0
}

const PASSWORD_MESSAGES = {
  tooShort: () => t('auth.validation.passwordMin', { min: PASSWORD_MIN_LENGTH }),
  tooLong: () => t('auth.validation.passwordTooLong', { max: PASSWORD_MAX_LENGTH }),
  upper: () => t('auth.validation.passwordUpper'),
  lower: () => t('auth.validation.passwordLower'),
  digit: () => t('auth.validation.passwordDigit'),
  special: () => t('auth.validation.passwordSpecial'),
  common: () => t('auth.validation.passwordCommon')
}
const passwordMessage = (problem) => PASSWORD_MESSAGES[problem]()

const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  loading.value = true
  
  try {
    await authStore.register(form)
    appStore.showSuccess(t('auth.register.success'))
    router.push('/auth/login')
  } catch (error) {
    console.error('Registration error:', error)
    
    if (error.response?.status === 422) {
      const validationErrors = error.response.data.errors
      if (validationErrors) {
        errors.value = validationErrors
      } else {
        appStore.showError(t('auth.invalidInput'))
      }
    } else if (error.response?.data?.code === 'weak_password') {
      // Règle du backend plus stricte que la vérification locale : afficher son motif
      errors.value = { password: apiErrorMessage(error, t('auth.register.failed')) }
    } else if (error.response?.status === 409) {
      appStore.showError(t('auth.register.alreadyExists'))
    } else {
      appStore.showError(t('auth.register.failed'))
    }
  } finally {
    loading.value = false
  }
}
</script>