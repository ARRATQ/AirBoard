<template>
  <div class="oauth-callback">
    <div class="callback-container">
      <div v-if="isLoading" class="loading-state">
        <div class="spinner"></div>
        <p class="loading-text">Connexion en cours...</p>
        <p class="loading-subtext">Authentification avec {{ providerName }}</p>
      </div>

      <div v-else-if="error" class="error-state">
        <Icon icon="mdi:alert-circle" class="error-icon" />
        <h2>Erreur de connexion</h2>
        <p class="error-message">{{ error }}</p>
        <button @click="redirectToLogin" class="btn-primary">
          Retour à la connexion
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { oauthService } from '@/services/api'
import { Icon } from '@iconify/vue'

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

    // Récupérer le state et nonce depuis sessionStorage si pas dans les query params
    let state = route.query.state
    let nonce = route.query.nonce

    if (!state || !nonce) {
      console.log('⚠️ State/nonce not in query params, trying sessionStorage...')
      state = state || sessionStorage.getItem('oauth_state')
      nonce = nonce || sessionStorage.getItem('oauth_nonce')
      
      if (!state || !nonce) {
        console.log('⚠️ No state/nonce found anywhere, proceeding without validation')
        // Pour Microsoft, on peut procéder sans state si pas de sessionStorage
        // (cela réduit la sécurité mais permet l'authentification)
      }
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
      localStorage.setItem('airboard_token', response.token)
      localStorage.setItem('airboard_refresh_token', response.refresh_token)
      localStorage.setItem('airboard_user', JSON.stringify(response.user))

      // Mettre à jour le store
      authStore.token = response.token
      authStore.refreshToken = response.refresh_token
      authStore.user = response.user
      authStore.isAuthenticated = true

      console.log('🎉 User authenticated:', response.user.username)

      // Rediriger vers la page d'accueil
      await router.push('/home')
    } else {
      throw new Error('Réponse invalide du serveur')
    }
  } catch (err) {
    console.error('❌ OAuth callback error:', err)
    error.value = err.response?.data?.message || err.message || 'Une erreur est survenue lors de l\'authentification'
    isLoading.value = false
  }
})

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
  border-radius: 12px;
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
  border-radius: 8px;
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
