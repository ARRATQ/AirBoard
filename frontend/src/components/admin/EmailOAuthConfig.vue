<template>
  <div class="card">
    <div class="card-header flex items-center justify-between">
      <div class="flex items-center gap-3">
        <Icon icon="mdi:shield-lock" class="h-6 w-6 text-green-400" />
        <h3 class="card-title">Configuration OAuth 2.0</h3>
      </div>
      <div class="flex items-center gap-2">
        <span :class="['text-sm font-medium', oauthConfig.is_enabled ? 'text-green-400 dark:text-green-300' : 'text-gray-400 dark:text-gray-500']">
          {{ oauthConfig.is_enabled ? 'Activé' : 'Désactivé' }}
        </span>
        <label class="relative inline-flex items-center cursor-pointer">
          <input type="checkbox" v-model="oauthConfig.is_enabled" class="sr-only peer">
          <div class="w-11 h-6 bg-gray-600 dark:bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-green-500"></div>
        </label>
      </div>
    </div>

    <div class="p-6 space-y-6">
      <!-- Info Banner -->
      <div class="p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
        <div class="flex gap-3">
          <Icon icon="mdi:information" class="h-5 w-5 text-blue-600 dark:text-blue-400 flex-shrink-0 mt-0.5" />
          <div class="space-y-1 text-sm text-blue-800 dark:text-blue-300">
            <p class="font-medium">Authentification moderne OAuth 2.0</p>
            <p class="text-blue-700 dark:text-blue-400">
              Remplace l'authentification SMTP obsolète (username/password).
              Requiert une App Registration dans Azure AD.
            </p>
          </div>
        </div>
      </div>

      <!-- Provider Selection -->
      <div class="form-group">
        <label class="form-label form-label-required">Provider</label>
        <select v-model="oauthConfig.provider" class="form-input">
          <option value="microsoft">Microsoft 365</option>
          <option value="google" disabled>Google Workspace (bientôt)</option>
        </select>
      </div>

      <!-- Microsoft 365 Configuration -->
      <div v-if="oauthConfig.provider === 'microsoft'" class="space-y-4">
        <!-- Tenant ID -->
        <div class="form-group">
          <label class="form-label form-label-required">
            Tenant ID
            <button
              type="button"
              @click="showHelp = !showHelp"
              class="ml-2 text-gray-400 hover:text-gray-300"
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
            Trouvé dans Azure AD > App Registration > Overview
          </p>
        </div>

        <!-- Client ID -->
        <div class="form-group">
          <label class="form-label form-label-required">Client ID</label>
          <input
            v-model="oauthConfig.client_id"
            type="text"
            class="form-input"
            placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
          />
          <p class="mt-1 text-xs text-gray-400">
            Application (client) ID depuis Azure AD
          </p>
        </div>

        <!-- Client Secret -->
        <div class="form-group">
          <label class="form-label" :class="{ 'form-label-required': !oauthConfig.id }">
            Client Secret
          </label>
          <div class="relative">
            <input
              v-model="clientSecret"
              :type="showSecret ? 'text' : 'password'"
              class="form-input pr-10"
              :placeholder="oauthConfig.id ? '••••••••••••••••' : 'Secret value depuis Azure AD'"
            />
            <button
              type="button"
              @click="showSecret = !showSecret"
              class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 dark:text-gray-500 hover:text-white dark:hover:text-gray-300"
            >
              <Icon :icon="showSecret ? 'mdi:eye-off' : 'mdi:eye'" class="h-5 w-5" />
            </button>
          </div>
          <p class="mt-1 text-xs text-gray-400">
            {{ oauthConfig.id ? 'Laisser vide pour conserver le secret actuel' : 'Créé dans Certificates & secrets' }}
          </p>
        </div>

        <!-- Grant Type -->
        <div class="form-group">
          <label class="form-label form-label-required">Grant Type</label>
          <select v-model="oauthConfig.grant_type" class="form-input">
            <option value="client_credentials">Client Credentials (recommandé)</option>
            <option value="refresh_token">Refresh Token</option>
          </select>
          <p class="mt-1 text-xs text-gray-400">
            Client Credentials = Service-to-service automatique
          </p>
        </div>

        <!-- Scopes -->
        <div class="form-group">
          <label class="form-label">Scopes OAuth</label>
          <input
            v-model="oauthConfig.scopes"
            type="text"
            class="form-input"
            :placeholder="oauthConfig.grant_type === 'client_credentials' ? 'https://graph.microsoft.com/.default' : 'https://outlook.office365.com/SMTP.Send'"
          />
          <p class="mt-1 text-xs text-gray-400">
            <span v-if="oauthConfig.grant_type === 'client_credentials'">
              Client Credentials utilise <code class="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">https://graph.microsoft.com/.default</code> (Microsoft Graph API)
            </span>
            <span v-else>
              Refresh Token utilise <code class="px-1 py-0.5 bg-gray-100 dark:bg-gray-700 rounded">https://outlook.office365.com/SMTP.Send</code> (SMTP)
            </span>
          </p>
        </div>
      </div>

      <!-- Token Status -->
      <div v-if="oauthConfig.id && oauthConfig.is_enabled" class="p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Statut du Token</h4>
        <div class="space-y-2">
          <div class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">Type:</span>
            <span class="font-medium text-gray-900 dark:text-gray-100">{{ oauthConfig.token_type || 'Bearer' }}</span>
          </div>
          <div v-if="oauthConfig.expires_at" class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">Expire:</span>
            <span :class="['font-medium', isTokenExpired ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400']">
              {{ formatTokenExpiry(oauthConfig.expires_at) }}
            </span>
          </div>
          <div v-if="oauthConfig.last_token_refresh" class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">Dernier refresh:</span>
            <span class="font-medium text-gray-900 dark:text-gray-100">{{ formatDate(oauthConfig.last_token_refresh) }}</span>
          </div>
          <div v-if="oauthConfig.last_refresh_error" class="mt-2 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-xs text-red-700 dark:text-red-400">
            <strong>Erreur:</strong> {{ oauthConfig.last_refresh_error }}
          </div>
        </div>
      </div>

      <!-- Help Section -->
      <div v-if="showHelp" class="space-y-4">
        <!-- Configuration Steps -->
        <div class="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
          <h4 class="text-sm font-medium text-yellow-800 dark:text-yellow-300 mb-2">
            <Icon icon="mdi:help-circle" class="inline h-4 w-4 mr-1" />
            Configuration Azure AD
          </h4>
          <ol class="text-xs text-yellow-700 dark:text-yellow-400 space-y-1 list-decimal list-inside">
            <li>Créer une App Registration dans Azure Portal</li>
            <li>Ajouter permission API: <code class="px-1 py-0.5 bg-yellow-100 dark:bg-yellow-900 rounded">Mail.Send</code> (Application) sous Microsoft Graph</li>
            <li><strong>Accorder le consentement administrateur</strong> (coche verte requise)</li>
            <li>Créer un client secret (expiration 24 mois)</li>
            <li>Copier: Tenant ID, Client ID, <strong>Secret VALUE</strong> (pas Secret ID)</li>
          </ol>
          <a
            href="https://portal.azure.com"
            target="_blank"
            class="inline-flex items-center gap-1 text-xs text-yellow-800 dark:text-yellow-300 hover:underline mt-2"
          >
            <Icon icon="mdi:open-in-new" class="h-3 w-3" />
            Ouvrir Azure Portal
          </a>
        </div>

        <!-- Common Issues -->
        <div class="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <h4 class="text-sm font-medium text-red-800 dark:text-red-300 mb-2">
            <Icon icon="mdi:alert-circle" class="inline h-4 w-4 mr-1" />
            Erreur 535 Authentication unsuccessful ?
          </h4>
          <div class="text-xs text-red-700 dark:text-red-400 space-y-2">
            <p><strong>Causes fréquentes :</strong></p>
            <ul class="list-disc list-inside space-y-1 ml-2">
              <li>Consentement administrateur non accordé (pas de coche verte)</li>
              <li>Mauvais secret copié (utilisez VALUE, pas Secret ID)</li>
              <li>Adresse "From Email" n'existe pas dans votre tenant M365</li>
              <li>Tenant ID incorrect (doit être un GUID, pas un nom de domaine)</li>
            </ul>
            <p class="mt-2">
              <a href="https://github.com/votre-repo/airboard/blob/main/docs/EMAIL_OAUTH_TROUBLESHOOTING.md" target="_blank" class="underline">
                Voir le guide de dépannage complet →
              </a>
            </p>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex flex-wrap gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
        <button @click="saveOAuthConfig" :disabled="saving" class="btn btn-primary">
          <Icon v-if="saving" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
          <Icon v-else icon="mdi:content-save" class="h-4 w-4 mr-2" />
          Sauvegarder
        </button>
        <button
          @click="showTestModal = true"
          :disabled="!oauthConfig.id || !oauthConfig.is_enabled"
          class="btn btn-secondary"
        >
          <Icon icon="mdi:email-fast" class="h-4 w-4 mr-2" />
          Tester OAuth
        </button>
        <button
          @click="refreshToken"
          :disabled="!oauthConfig.id || !oauthConfig.is_enabled || refreshing"
          class="btn btn-secondary"
        >
          <Icon v-if="refreshing" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
          <Icon v-else icon="mdi:refresh" class="h-4 w-4 mr-2" />
          Rafraîchir Token
        </button>
      </div>
    </div>
  </div>

  <!-- Test Modal -->
  <Teleport to="body">
    <div v-if="showTestModal" class="modal-overlay" @click.self="showTestModal = false">
      <div class="modal-container max-w-md">
        <div class="modal-header">
          <h3 class="modal-title">Tester la connexion OAuth</h3>
          <button @click="showTestModal = false" class="modal-close">
            <Icon icon="mdi:close" class="h-5 w-5" />
          </button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label class="form-label form-label-required">Email de test</label>
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
            Annuler
          </button>
          <button @click="testOAuth" :disabled="!testEmail || testing" class="btn btn-primary">
            <Icon v-if="testing" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
            <Icon v-else icon="mdi:send" class="h-4 w-4 mr-2" />
            Envoyer
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { emailService } from '@/services/api'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()

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
      appStore.showError('Erreur', response.message || response.error)
    } else {
      appStore.showSuccess('Succès', response.message || 'Configuration OAuth sauvegardée')
      clientSecret.value = '' // Clear password field
      await loadOAuthConfig()
    }
  } catch (error) {
    console.error('Erreur sauvegarde OAuth:', error)
    appStore.showError('Erreur', error.response?.data?.error || 'Échec de la sauvegarde')
  } finally {
    saving.value = false
  }
}

const testOAuth = async () => {
  if (!testEmail.value) return

  testing.value = true
  try {
    const response = await emailService.testOAuthConnection(testEmail.value)
    appStore.showSuccess('Test réussi', response.message || 'Email de test envoyé avec succès')
    showTestModal.value = false
    testEmail.value = ''
  } catch (error) {
    console.error('Erreur test OAuth:', error)
    appStore.showError('Test échoué', error.response?.data?.error || 'Échec du test de connexion')
  } finally {
    testing.value = false
  }
}

const refreshToken = async () => {
  refreshing.value = true
  try {
    const response = await emailService.refreshOAuthToken()
    appStore.showSuccess('Token rafraîchi', response.message || 'Token rafraîchi avec succès')
    await loadOAuthConfig()
  } catch (error) {
    console.error('Erreur rafraîchissement token:', error)
    appStore.showError('Échec', error.response?.data?.error || 'Échec du rafraîchissement')
  } finally {
    refreshing.value = false
  }
}

const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('fr-FR', {
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

  if (diff < 0) return 'Expiré'

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)

  if (days > 0) return `Dans ${days}j ${hours % 24}h`
  if (hours > 0) return `Dans ${hours}h ${minutes % 60}m`
  return `Dans ${minutes}m`
}

// Lifecycle
onMounted(() => {
  loadOAuthConfig()
})
</script>
