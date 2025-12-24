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
            <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Welcome back</h1>
            <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">Sign in to access your dashboard</p>
          </div>
        </div>

        <!-- OAuth Buttons -->
        <div v-if="oauthProviders.length > 0" class="space-y-3">
          <button
            v-for="provider in oauthProviders"
            :key="provider.id"
            type="button"
            @click="handleOAuthLogin(provider)"
            class="w-full flex items-center justify-center gap-3 px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors duration-200 shadow-sm hover:shadow-md"
          >
            <Icon :icon="provider.icon" class="h-5 w-5" />
            <span>Continue with {{ provider.display_name }}</span>
          </button>

          <!-- Divider -->
          <div class="relative my-6">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-gray-300 dark:border-gray-600"></div>
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-4 bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400 font-medium">Or continue with</span>
            </div>
          </div>
        </div>

        <!-- Form -->
        <form @submit.prevent="handleSubmit" class="space-y-5">
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
              placeholder="Enter your username"
              :disabled="loading"
            />
            <p v-if="errors.username" class="mt-2 text-sm text-red-600 dark:text-red-400">{{ errors.username }}</p>
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
          </div>

          <div class="flex items-center">
            <label class="flex items-center cursor-pointer group">
              <input
                v-model="form.remember"
                type="checkbox"
                class="h-4 w-4 text-green-600 focus:ring-green-500 border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 cursor-pointer"
              />
              <span class="ml-2 text-sm text-gray-600 dark:text-gray-400 group-hover:text-gray-900 dark:group-hover:text-gray-200 transition-colors duration-200">Remember me</span>
            </label>
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full flex items-center justify-center px-4 py-3 border border-transparent rounded-xl text-sm font-medium text-white bg-gradient-to-r from-green-500 to-green-600 hover:from-green-600 hover:to-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 shadow-lg hover:shadow-xl transform hover:-translate-y-0.5 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
          >
            <Icon v-if="loading" icon="mdi:loading" class="animate-spin h-5 w-5 mr-2" />
            <span>{{ loading ? 'Signing in...' : 'Sign in' }}</span>
          </button>

          <div v-if="signupEnabled" class="text-center pt-4">
            <span class="text-sm text-gray-600 dark:text-gray-400">Don't have an account? </span>
            <router-link
              to="/auth/register"
              class="text-sm text-green-600 dark:text-green-400 hover:text-green-700 dark:hover:text-green-300 font-semibold transition-colors duration-200"
            >
              Sign up
            </router-link>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, nextTick, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { oauthService, authService } from '@/services/api'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const loading = ref(false)
const showPassword = ref(false)
const errors = ref({})
const oauthProviders = ref([])
const signupEnabled = ref(true)

const form = reactive({
  username: '',
  password: '',
  remember: false
})

const validateForm = () => {
  errors.value = {}
  
  if (!form.username.trim()) {
    errors.value.username = 'Username is required'
  }
  
  if (!form.password.trim()) {
    errors.value.password = 'Password is required'
  }
  
  return Object.keys(errors.value).length === 0
}

const handleSubmit = async () => {
  if (!validateForm()) return

  loading.value = true

  try {
    await authStore.login(form)

    appStore.showSuccess('Welcome back!')

    // Force navigation with nextTick to ensure state is updated
    await nextTick()

    // Check for redirect parameter, default to home page
    const redirectPath = router.currentRoute.value.query.redirect || '/home'

    // Force replace instead of push to avoid back button issues
    await router.replace(redirectPath)

  } catch (error) {
    console.error('Login error:', error)

    if (error.response?.status === 401) {
      appStore.showError('Invalid username or password')
    } else if (error.response?.status === 422) {
      const validationErrors = error.response.data.errors
      if (validationErrors) {
        errors.value = validationErrors
      } else {
        appStore.showError('Invalid input data')
      }
    } else {
      appStore.showError('Login failed. Please try again.')
    }
  } finally {
    loading.value = false
  }
}

const loadOAuthProviders = async () => {
  try {
    const data = await oauthService.getEnabledProviders()
    oauthProviders.value = data.providers || []
  } catch (error) {
    console.error('Error loading OAuth providers:', error)
  }
}

const loadSignupStatus = async () => {
  try {
    const data = await authService.getSignupStatus()
    signupEnabled.value = data.signup_enabled || false
  } catch (error) {
    console.error('Error loading signup status:', error)
    signupEnabled.value = true // Par défaut, on active le signup en cas d'erreur
  }
}

const handleOAuthLogin = async (provider) => {
  try {
    loading.value = true
    const { auth_url, state, nonce } = await oauthService.initiateOAuth(provider.provider_name)

    // Sauvegarder le state et nonce dans sessionStorage pour vérification ultérieure
    sessionStorage.setItem('oauth_state', state)
    sessionStorage.setItem('oauth_nonce', nonce)
    sessionStorage.setItem('oauth_provider', provider.provider_name)

    // Rediriger vers la page OAuth du provider
    window.location.href = auth_url
  } catch (error) {
    console.error('OAuth initiation error:', error)
    appStore.showError('Failed to initiate OAuth login')
    loading.value = false
  }
}

// Lifecycle
onMounted(() => {
  loadOAuthProviders()
  loadSignupStatus()

  // Vérifier si on revient d'un callback OAuth
  const code = route.query.code
  const state = route.query.state
  const nonce = route.query.nonce

  if (code && state) {
    handleOAuthCallback(code, state, nonce)
  }
})

const handleOAuthCallback = async (code, state, nonce) => {
  const savedState = sessionStorage.getItem('oauth_state')
  const savedNonce = sessionStorage.getItem('oauth_nonce')
  const provider = sessionStorage.getItem('oauth_provider')

  if (!savedState || savedState !== state) {
    appStore.showError('Invalid OAuth state. Please try again.')
    router.replace('/auth/login')
    return
  }

  try {
    loading.value = true
    const response = await oauthService.handleCallback(provider, code, state, nonce)

    // Sauvegarder les tokens
    authStore.setToken(response.token)
    authStore.setRefreshToken(response.refresh_token)
    authStore.setUser(response.user)

    // Nettoyer le sessionStorage
    sessionStorage.removeItem('oauth_state')
    sessionStorage.removeItem('oauth_nonce')
    sessionStorage.removeItem('oauth_provider')

    appStore.showSuccess('Welcome back!')
    await router.replace('/home')
  } catch (error) {
    console.error('OAuth callback error:', error)
    appStore.showError('OAuth authentication failed')
    router.replace('/auth/login')
  } finally {
    loading.value = false
  }
}
</script>