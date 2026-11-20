import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import i18n from '@/i18n'
import { setupIcons } from '@/plugins/icons'

// Styles CSS
import './assets/css/main.css'

// Services
import { setupInterceptors } from './services/api'

// Stores
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

// Configuration de l'application
const app = createApp(App)
const pinia = createPinia()

// Icônes embarquées (aucun appel à api.iconify.design)
setupIcons()

// Plugins
app.use(pinia)
app.use(router)
app.use(i18n)

// Initialiser le store après l'installation de Pinia
const authStore = useAuthStore()

// Des vues appellent window.$toast.*(message) : ces appels passent par les toasts du store.
const appStore = useAppStore()
window.$toast = {
  success: (message) => appStore.showSuccess(message),
  error: (message) => appStore.showError(message),
  warning: (message) => appStore.showWarning(message),
  info: (message) => appStore.showInfo(message),
}

// Configuration des intercepteurs API
setupInterceptors(router, () => {
  authStore.logout()
})

// Montage de l'application
app.mount('#app')