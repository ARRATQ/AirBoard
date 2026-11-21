<template>
  <div class="card">
    <div class="card-header flex items-center justify-between">
      <div class="flex items-center gap-3">
        <Icon icon="mdi:shield-lock" class="h-6 w-6 text-green-400" />
        <h3 class="card-title">{{ t('email.oauth.title') }}</h3>
      </div>
      <div class="flex items-center gap-2">
        <span :class="['text-sm font-medium', oauthConfig.is_enabled ? 'text-green-400 dark:text-green-300' : 'text-gray-400 dark:text-gray-500']">
          {{ oauthConfig.is_enabled ? t('email.enabled') : t('email.disabled') }}
        </span>
        <label class="relative inline-flex items-center cursor-pointer">
          <input type="checkbox" v-model="oauthConfig.is_enabled" class="sr-only peer" :aria-label="t('email.oauth.title')">
          <div class="w-11 h-6 bg-gray-600 dark:bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-500"></div>
        </label>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- Info Banner -->
      <div class="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
        <div class="flex gap-3">
          <Icon icon="mdi:information" class="h-5 w-5 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
          <div class="space-y-1 text-sm text-blue-800 dark:text-blue-300">
            <p class="font-medium">{{ t('email.oauth.bannerTitle') }}</p>
            <p class="text-blue-700 dark:text-blue-400">
              {{ t('email.oauth.bannerText') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Provider Selection -->
      <div class="form-group">
        <label class="form-label form-label-required">{{ t('email.oauth.provider') }}</label>
        <select v-model="oauthConfig.provider" class="form-input">
          <option value="microsoft">Microsoft 365</option>
          <option value="google" disabled>{{ t('email.oauth.googleSoon') }}</option>
        </select>
      </div>

      <!-- Microsoft 365 Configuration -->
      <div v-if="oauthConfig.provider === 'microsoft'" class="space-y-4">
        <!-- Tenant ID -->
        <div class="form-group">
          <label class="form-label form-label-required">
            {{ t('email.oauth.tenantId') }}
            <button
              type="button"
              @click="showHelp = !showHelp"
              :aria-label="t('email.oauth.toggleHelp')"
              :aria-expanded="showHelp"
              class="ms-2 text-gray-400 hover:text-gray-300"
            >
              <Icon icon="mdi:help-circle" class="h-4 w-4" />
            </button>
          </label>
          <input
            v-model="oauthConfig.tenant_id"
            type="text"
            class="form-input"
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
          />
          <p class="mt-1 text-xs text-gray-400">
            {{ t('email.oauth.tenantIdHelp') }}
          </p>
        </div>

        <!-- Client ID -->
        <div class="form-group">
          <label class="form-label form-label-required">{{ t('oauth.clientId') }}</label>
          <input
            v-model="oauthConfig.client_id"
            type="text"
            class="form-input"
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
          />
          <p class="mt-1 text-xs text-gray-400">
            {{ t('email.oauth.clientIdHelp') }}
          </p>
        </div>

        <!-- Client Secret -->
        <div class="form-group">
          <label class="form-label" :class="{ 'form-label-required': !oauthConfig.id }">
            {{ t('oauth.clientSecret') }}
          </label>
          <div class="relative">
            <input
              v-model="clientSecret"
              :type="showSecret ? 'text' : 'password'"
              class="form-input pe-10"
              :placeholder="oauthConfig.id ? '••••••••••••••••' : t('email.oauth.clientSecretPlaceholder')"
            />
            <button
              type="button"
              @click="showSecret = !showSecret"
              :aria-label="showSecret ? t('email.oauth.hideSecret') : t('email.oauth.showSecret')"
              class="absolute end-3 top-1/2 transform -translate-y-1/2 text-gray-400 dark:text-gray-500 hover:text-white dark:hover:text-gray-300"
            >
              <Icon :icon="showSecret ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
            </button>
          </div>
          <p class="mt-1 text-xs text-gray-400">
            {{ oauthConfig.id ? t('oauth.clientSecretPlaceholder') : t('email.oauth.clientSecretHelp') }}
          </p>
        </div>

        <!-- Grant Type -->
        <div class="form-group">
          <label class="form-label form-label-required">{{ t('email.oauth.grantType') }}</label>
          <select v-model="oauthConfig.grant_type" class="form-input">
            <option value="client_credentials">{{ t('email.oauth.grantClientCredentials') }}</option>
            <option value="refresh_token">Refresh Token</option>
          </select>
          <p class="mt-1 text-xs text-gray-400">
            {{ t('email.oauth.grantHelp') }}
          </p>
        </div>

        <!-- Scopes -->
        <div class="form-group">
          <label class="form-label">{{ t('oauth.scopes') }}</label>
          <input
            v-model="oauthConfig.scopes"
            type="text"
            class="form-input"
            :placeholder="oauthConfig.grant_type === 'client_credentials' ? 'https://graph.microsoft.com/.default' : 'https://outlook.office365.com/SMTP.Send'"
          />
          <p class="mt-1 text-xs text-gray-400">
            <i18n-t v-if="oauthConfig.grant_type === 'client_credentials'" keypath="email.oauth.scopesClientCredentials" tag="span">
              <template #scope><code class="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded" dir="ltr">https://graph.microsoft.com/.default</code></template>
            </i18n-t>
            <i18n-t v-else keypath="email.oauth.scopesRefreshToken" tag="span">
              <template #scope><code class="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded" dir="ltr">https://outlook.office365.com/SMTP.Send</code></template>
            </i18n-t>
          </p>
        </div>
      </div>

      <!-- Token Status -->
      <div v-if="oauthConfig.id && oauthConfig.is_enabled" class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">{{ t('email.oauth.tokenStatus') }}</h4>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('email.oauth.tokenType') }}</span>
            <span class="font-medium text-gray-900 dark:text-gray-100">{{ oauthConfig.token_type || 'Bearer' }}</span>
          </div>
          <div v-if="oauthConfig.expires_at" class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('email.oauth.tokenExpires') }}</span>
            <span :class="['font-medium', isTokenExpired ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400']">
              {{ formatTokenExpiry(oauthConfig.expires_at) }}
            </span>
          </div>
          <div v-if="oauthConfig.last_token_refresh" class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('email.oauth.lastRefresh') }}</span>
            <span class="font-medium text-gray-900 dark:text-gray-100">{{ formatDate(oauthConfig.last_token_refresh) }}</span>
          </div>
          <div v-if="oauthConfig.last_refresh_error" class="mt-2 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-xs text-red-700 dark:text-red-400">
            <strong>{{ t('email.oauth.errorLabel') }}</strong> {{ oauthConfig.last_refresh_error }}
          </div>
        </div>
      </div>

      <!-- Help Section -->
      <div v-if="showHelp" class="space-y-4">
        <!-- Configuration Steps -->
        <div class="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
          <h4 class="text-sm font-medium text-yellow-800 dark:text-yellow-300 mb-2">
            <Icon icon="mdi:help-circle" class="inline h-4 w-4 me-1" />
            {{ t('email.oauth.azureSetupTitle') }}
          </h4>
          <ol class="text-xs text-yellow-700 dark:text-yellow-400 space-y-1 list-decimal list-inside">
            <li>{{ t('email.oauth.step1') }}</li>
            <i18n-t keypath="email.oauth.step2" tag="li">
              <template #permission><code class="px-1 py-0.5 bg-yellow-100 dark:bg-yellow-900 rounded">Mail.Send</code></template>
            </i18n-t>
            <i18n-t keypath="email.oauth.step3" tag="li">
              <template #consent><strong>{{ t('email.oauth.step3Consent') }}</strong></template>
            </i18n-t>
            <li>{{ t('email.oauth.step4') }}</li>
            <i18n-t keypath="email.oauth.step5" tag="li">
              <template #secretValue><strong>Secret VALUE</strong></template>
            </i18n-t>
          </ol>
          <a
            href="https://portal.azure.com"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-1 text-xs text-yellow-800 dark:text-yellow-300 hover:underline mt-2"
          >
            <Icon icon="mdi:open-in-new" class="h-3 w-3" />
            {{ t('email.oauth.openPortal') }}
          </a>
        </div>

        <!-- Common Issues -->
        <div class="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <h4 class="text-sm font-medium text-red-800 dark:text-red-300 mb-2">
            <Icon icon="mdi:alert-circle" class="inline h-4 w-4 me-1" />
            {{ t('email.oauth.error535Title') }}
          </h4>
          <div class="text-xs text-red-700 dark:text-red-400 space-y-2">
            <p><strong>{{ t('email.oauth.commonCauses') }}</strong></p>
            <ul class="list-disc list-inside space-y-1 ms-2">
              <li>{{ t('email.oauth.cause1') }}</li>
              <li>{{ t('email.oauth.cause2') }}</li>
              <li>{{ t('email.oauth.cause3') }}</li>
              <li>{{ t('email.oauth.cause4') }}</li>
            </ul>
            <p class="mt-2">
              <a href="https://github.com/votre-repo/airboard/blob/main/docs/EMAIL_OAUTH_TROUBLESHOOTING.md" target="_blank" rel="noopener noreferrer" class="underline">
                {{ t('email.oauth.troubleshootingLink') }}
              </a>
            </p>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex flex-wrap gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
        <button @click="saveOAuthConfig" :disabled="saving" class="btn btn-primary">
          <Icon v-if="saving" icon="mdi:loading" class="h-4 w-4 me-2 animate-spin" />
          <Icon v-else icon="mdi:content-save" class="h-4 w-4 me-2" />
          {{ t('common.save') }}
        </button>
        <button
          @click="showTestModal = true"
          :disabled="!oauthConfig.id || !oauthConfig.is_enabled"
          class="btn btn-secondary"
        >
          <Icon icon="mdi:email-fast" class="h-4 w-4 me-2" />
          {{ t('email.oauth.test') }}
        </button>
        <button
          @click="refreshToken"
          :disabled="!oauthConfig.id || !oauthConfig.is_enabled || refreshing"
          class="btn btn-secondary"
        >
          <Icon v-if="refreshing" icon="mdi:loading" class="h-4 w-4 me-2 animate-spin" />
          <Icon v-else icon="mdi:refresh" class="h-4 w-4 me-2" />
          {{ t('email.oauth.refreshToken') }}
        </button>
      </div>
    </div>
  </div>

  <!-- Test Modal -->
  <Teleport to="body">
    <div v-if="showTestModal" class="modal-overlay" @click.self="showTestModal = false">
      <div v-focus-trap class="modal-container max-w-md" role="dialog" aria-modal="true" aria-labelledby="email-oauth-test-modal-title">
        <div class="modal-header">
          <h3 id="email-oauth-test-modal-title" class="modal-title">{{ t('email.oauth.testModalTitle') }}</h3>
          <button @click="showTestModal = false" class="modal-close" :aria-label="t('common.close')">
            <Icon icon="mdi:close" class="h-5 w-5" />
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label form-label-required">{{ t('email.testEmailAddress') }}</label>
            <input
              v-model="testEmail"
              type="email"
              class="form-input"
              placeholder="test@example.com"
              @keyup.enter="testOAuth"
            />
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showTestModal = false" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button @click="testOAuth" :disabled="!testEmail || testing" class="btn btn-primary">
            <Icon v-if="testing" icon="mdi:loading" class="h-4 w-4 me-2 animate-spin" />
            <Icon v-else icon="mdi:send" class="h-4 w-4 me-2" />
            {{ t('email.send') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useEscapeClose } from '@/composables/useEscapeClose'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { emailService } from '@/services/api'
import { useAppStore } from '@/stores/app'
import { dateLocaleFor } from '@/utils/homeFormat'

const appStore = useAppStore()
const { t, locale } = useI18n()

// State
const oauthConfig = ref({
  provider: 'microsoft',
  tenant_id: '',
  client_id: '',
  grant_type: 'client_credentials',
  scopes: '',
  is_enabled: false
})

const clientSecret = ref('')
const showSecret = ref(false)
const showHelp = ref(false)
const saving = ref(false)
const testing = ref(false)
const refreshing = ref(false)
const showTestModal = ref(false)
const testEmail = ref('')

// Computed
const isTokenExpired = computed(() => {
  if (!oauthConfig.value.expires_at) return false
  return new Date(oauthConfig.value.expires_at) < new Date()
})

// Methods
const loadOAuthConfig = async () => {
  try {
    const data = await emailService.getOAuthConfig()
    if (data) {
      oauthConfig.value = {
        ...oauthConfig.value,
        ...data
      }
    }
  } catch (error) {
    console.error('Erreur chargement config OAuth:', error)
  }
}

const saveOAuthConfig = async () => {
  saving.value = true
  try {
    const payload = {
      provider: oauthConfig.value.provider,
      tenant_id: oauthConfig.value.tenant_id,
      client_id: oauthConfig.value.client_id,
      grant_type: oauthConfig.value.grant_type,
      is_enabled: oauthConfig.value.is_enabled
    }

    // Add scopes if provided
    if (oauthConfig.value.scopes) {
      payload.scopes = oauthConfig.value.scopes
    }

    // Add client secret only if modified
    if (clientSecret.value) {
      payload.client_secret = clientSecret.value
    }

    const response = await emailService.updateOAuthConfig(payload)

    if (response.error) {
      appStore.showError(response.message || response.error)
    } else {
      appStore.showSuccess(t('email.oauth.saved'))
      clientSecret.value = '' // Clear password field
      await loadOAuthConfig()
    }
  } catch (error) {
    console.error('Erreur sauvegarde OAuth:', error)
    appStore.showError(error.response?.data?.error || t('common.feedback.saveError'))
  } finally {
    saving.value = false
  }
}

const testOAuth = async () => {
  if (!testEmail.value) return

  testing.value = true
  try {
    await emailService.testOAuthConnection(testEmail.value)
    appStore.showSuccess(t('email.testSuccess'))
    showTestModal.value = false
    testEmail.value = ''
  } catch (error) {
    console.error('Erreur test OAuth:', error)
    appStore.showError(error.response?.data?.error || t('email.oauth.testFailed'))
  } finally {
    testing.value = false
  }
}

const refreshToken = async () => {
  refreshing.value = true
  try {
    await emailService.refreshOAuthToken()
    appStore.showSuccess(t('email.oauth.tokenRefreshed'))
    await loadOAuthConfig()
  } catch (error) {
    console.error('Erreur rafraîchissement token:', error)
    appStore.showError(error.response?.data?.error || t('email.oauth.refreshError'))
  } finally {
    refreshing.value = false
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString(dateLocaleFor(locale.value), {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatTokenExpiry = (dateString) => {
  if (!dateString) return '-'
  const expiry = new Date(dateString)
  const now = new Date()
  const diff = expiry - now

  if (diff < 0) return t('email.oauth.expired')

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return t('email.oauth.inDaysHours', { days, hours: hours % 24 })
  if (hours > 0) return t('email.oauth.inHoursMinutes', { hours, minutes: minutes % 60 })
  return t('email.oauth.inMinutes', { minutes })
}

// Lifecycle
onMounted(() => {
  loadOAuthConfig()
})

// Échap ferme la modale
useEscapeClose(showTestModal, () => { showTestModal.value = false })
</script>
