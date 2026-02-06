import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'
import i18n from '@/i18n'

// Styles CSS
import './assets/css/main.css'

// Services
import { setupInterceptors } from './services/api'

// Stores
import { useAuthStore } from '@/stores/auth'

// Configuration de l'application
const app = createApp(App)
const pinia = createPinia()

// Plugins
app.use(pinia)
app.use(router)
app.use(i18n)

// Initialiser le store après l'installation de Pinia
const authStore = useAuthStore()

// Configuration des intercepteurs API
setupInterceptors(router, () => {
  authStore.logout()
})

// Montage de l'application
app.mount('#app')