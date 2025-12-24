<template>
  <div class="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800 flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <!-- Card with modern styling -->
      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8 space-y-8">
        <!-- Header -->
        <div class="text-center space-y-4">
          <div class="flex items-center justify-center">
            <div class="h-16 w-16 bg-gradient-to-br from-green-400 to-green-600 rounded-2xl flex items-center justify-center shadow-lg transform hover:scale-105 transition-transform duration-200">
              <Icon icon="mdi:view-dashboard" class="h-8 w-8 text-white" />
            </div>
          </div>
          <div>
            <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Create Account</h1>
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">Join Airboard today</p>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleSubmit" class="space-y-5">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="first_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                First Name
              </label>
              <input
                id="first_name"
                v-model="form.first_name"
                type="text"
                class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
                placeholder="John"
                :disabled="loading"
              />
            </div>

            <div>
              <label for="last_name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Last Name
              </label>
              <input
                id="last_name"
                v-model="form.last_name"
                type="text"
                class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
                placeholder="Doe"
                :disabled="loading"
              />
            </div>
          </div>

          <div>
            <label for="username" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Username <span class="text-red-500">*</span>
            </label>
            <input
              id="username"
              v-model="form.username"
              type="text"
              required
              class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
              placeholder="johndoe"
              :disabled="loading"
            />
            <p v-if="errors.username" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.username }}</p>
          </div>

          <div>
            <label for="email" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Email <span class="text-red-500">*</span>
            </label>
            <input
              id="email"
              v-model="form.email"
              type="email"
              required
              class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
              placeholder="john@example.com"
              :disabled="loading"
            />
            <p v-if="errors.email" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.email }}</p>
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Password <span class="text-red-500">*</span>
            </label>
            <div class="relative">
              <input
                id="password"
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'"
                required
                class="w-full px-4 py-3 pr-12 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
                placeholder="Enter your password"
                :disabled="loading"
                minlength="6"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors duration-200"
              >
                <Icon :icon="showPassword ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
              </button>
            </div>
            <p v-if="errors.password" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.password }}</p>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">At least 6 characters</p>
          </div>

          <div>
            <label for="password_confirmation" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Confirm Password <span class="text-red-500">*</span>
            </label>
            <div class="relative">
              <input
                id="password_confirmation"
                v-model="form.password_confirmation"
                :type="showConfirmPassword ? 'text' : 'password'"
                required
                class="w-full px-4 py-3 pr-12 border border-gray-300 dark:border-gray-600 rounded-xl text-gray-900 dark:text-white bg-white dark:bg-gray-700 placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-transparent transition-all duration-200"
                placeholder="Confirm your password"
                :disabled="loading"
              />
              <button
                type="button"
                @click="showConfirmPassword = !showConfirmPassword"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors duration-200"
              >
                <Icon :icon="showConfirmPassword ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
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
                class="h-4 w-4 mt-0.5 text-green-600 focus:ring-green-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 cursor-pointer"
              />
              <span class="ml-2 text-sm text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-200 transition-colors duration-200">
                I agree to the
                <a href="#" class="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-medium">Terms of Service</a>
                and
                <a href="#" class="text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-medium">Privacy Policy</a>
              </span>
            </label>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full flex items-center justify-center px-4 py-3 border border-transparent rounded-xl text-sm font-medium text-white bg-gradient-to-r from-green-500 to-green-600 hover:from-green-600 hover:to-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
          >
            <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-5 w-5 mr-2" />
            <span>{{ loading ? 'Creating account...' : 'Create Account' }}</span>
          </button>

          <div class="text-center pt-4">
            <span class="text-sm text-gray-600 dark:text-gray-400">Already have an account? </span>
            <router-link
              to="/auth/login"
              class="text-sm text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold transition-colors duration-200"
            >
              Sign in
            </router-link>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { authService } from '@/services/api'

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
      appStore.showError('L\'inscription est actuellement désactivée')
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
    errors.value.username = 'Username is required'
  } else if (form.username.length < 3) {
    errors.value.username = 'Username must be at least 3 characters'
  }
  
  if (!form.email.trim()) {
    errors.value.email = 'Email is required'
  } else if (!isValidEmail(form.email)) {
    errors.value.email = 'Invalid email format'
  }
  
  if (!form.password.trim()) {
    errors.value.password = 'Password is required'
  } else if (form.password.length < 6) {
    errors.value.password = 'Password must be at least 6 characters'
  }
  
  if (form.password !== form.password_confirmation) {
    errors.value.password_confirmation = 'Passwords do not match'
  }
  
  return Object.keys(errors.value).length === 0
}

const isValidEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const handleSubmit = async () => {
  if (!validateForm()) return
  
  loading.value = true
  
  try {
    await authStore.register(form)
    appStore.showSuccess('Account created successfully! Please sign in.')
    router.push('/auth/login')
  } catch (error) {
    console.error('Registration error:', error)
    
    if (error.response?.status === 422) {
      const validationErrors = error.response.data.errors
      if (validationErrors) {
        errors.value = validationErrors
      } else {
        appStore.showError('Invalid input data')
      }
    } else if (error.response?.status === 409) {
      appStore.showError('Username or email already exists')
    } else {
      appStore.showError('Registration failed. Please try again.')
    }
  } finally {
    loading.value = false
  }
}
</script>