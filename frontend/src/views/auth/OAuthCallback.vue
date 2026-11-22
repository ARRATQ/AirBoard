<template>
  <div class="oauth-callback">
    <div class="callback-container">
      <div v-if="isLoading" class="loading-state">
        <div class="spinner"></div>
        <p class="loading-text" role="status">{{ t('auth.oauthCallback.loading') }}</p>
        <p class="loading-subtext">{{ t('auth.oauthCallback.with', { provider: providerName }) }}</p>
      </div>

      <div v-else-if="error" class="error-state">
        <Icon icon="mdi:alert-circle" class="error-icon" />
        <h2>{{ t('auth.oauthCallback.errorTitle') }}</h2>
        <p class="error-message">{{ error }}</p>
        <button type="button" @click="redirectToLogin" class="btn-primary">
          {{ t('auth.oauthCallback.backToLogin') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { apiErrorMessage } from '@/services/apiError'
import { useAuthStore } from '@/stores/auth'
import { oauthService } from '@/services/api'
import { Icon } from '@iconify/vue'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isLoading = ref(true)
const error = ref(null)
const providerName = ref('')

onMounted(async () => {
  try {
    // Extraire le provider depuis le path
    const pathParts = route.path.split('/')
    const providerIndex = pathParts.indexOf('oauth') + 1
    providerName.value = pathParts[providerIndex] || 'unknown'

    console.log('🔐 OAuth Callback:', {
      provider: providerName.value,
      code: route.query.code ? 'present' : 'missing',
      state: route.query.state,
      nonce: route.query.nonce,
      error: route.query.error,
      allQueryParams: Object.keys(route.query)
    })

    // Vérifier si Microsoft a retourné une erreur
    if (route.query.error) {
      const errorDescription = route.query.error_description || route.query.error
      throw new Error(errorDescription)
    }

    // Vérifier que le code est présent
    if (!route.query.code) {
      throw new Error('Code d\'autorisation manquant')
    }

    // SECURITY (CSRF / login forcé) : le state renvoyé par le provider doit être
    // celui généré dans CE navigateur. Le nonce vient uniquement du sessionStorage,
    // jamais de l'URL (qu'un attaquant pourrait forger).
    const state = route.query.state
    const savedState = sessionStorage.getItem('oauth_state')
    const nonce = sessionStorage.getItem('oauth_nonce')
    clearOAuthSession()

    if (!state || !savedState || state !== savedState || !nonce) {
      throw new Error('La session de connexion a expiré ou est invalide. Veuillez relancer la connexion.')
    }

    // Appeler le backend pour échanger le code contre un token
    console.log('📞 Calling backend callback (POST)...')
    console.log('📦 Sending:', {
      provider: providerName.value,
      codeLength: route.query.code?.length,
      stateLength: state?.length,
      nonceLength: nonce?.length
    })

    const response = await oauthService.handleCallback(
      providerName.value,
      route.query.code,
      state,
      nonce
    )

    console.log('✅ OAuth callback successful:', response)

    // Stocker les tokens et l'utilisateur
    if (response.token && response.user) {
      // Enrichir l'objet user avec managed_group_ids du JWT si absent
      const userData = { ...response.user }
      if (!userData.managed_group_ids || userData.managed_group_ids.length === 0) {
        try {
          const payload = response.token.split('.')[1]
          const decoded = JSON.parse(atob(payload))
          userData.managed_group_ids = decoded.managed_group_ids || []
          console.log('🔑 managed_group_ids extrait du JWT:', userData.managed_group_ids)
        } catch (e) {
          console.warn('Erreur décodage JWT:', e)
          userData.managed_group_ids = []
        }
      }

      localStorage.setItem('airboard_token', response.token)
      localStorage.setItem('airboard_refresh_token', response.refresh_token)
      localStorage.setItem('airboard_user', JSON.stringify(userData))

      // Mettre à jour le store via les setters appropriés
      authStore.setToken(response.token)
      authStore.setRefreshToken(response.refresh_token)
      authStore.setUser(userData)

      console.log('🎉 User authenticated:', userData.username, 'managed_group_ids:', userData.managed_group_ids)

      // Rediriger vers la page d'accueil
      await router.push('/home')
    } else {
      throw new Error('Réponse invalide du serveur')
    }
  } catch (err) {
    console.error('❌ OAuth callback error:', err)
    error.value = apiErrorMessage(err, err.message || t('auth.oauth.failed'))
    isLoading.value = false
  }
})

function clearOAuthSession() {
  sessionStorage.removeItem('oauth_state')
  sessionStorage.removeItem('oauth_nonce')
  sessionStorage.removeItem('oauth_provider')
}

const redirectToLogin = () => {
  router.push('/auth/login')
}
</script>

<style scoped>
.oauth-callback {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.callback-container {
  background: white;
  border-radius: var(--r-s);
  padding: 3rem;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  max-width: 450px;
  width: 90%;
  text-align: center;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.spinner {
  width: 50px;
  height: 50px;
  border: 4px solid #f3f4f6;
  border-top: 4px solid #667eea;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.loading-text {
  font-size: 1.25rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.loading-subtext {
  font-size: 0.95rem;
  color: #6b7280;
  margin: 0;
}

.error-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
}

.error-icon {
  font-size: 4rem;
  color: #ef4444;
}

.error-state h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.error-message {
  color: #6b7280;
  margin: 0;
  line-height: 1.6;
}

.btn-primary {
  background: #667eea;
  color: white;
  border: none;
  padding: 0.75rem 2rem;
  border-radius: var(--r-s);
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 1rem;
}

.btn-primary:hover {
  background: #5568d3;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}
</style>
