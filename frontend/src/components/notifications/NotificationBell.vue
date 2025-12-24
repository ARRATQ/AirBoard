<template>
  <div class="notification-bell-container relative">
    <!-- Bell Icon with Badge -->
    <button
      @click="togglePanel"
      class="notification-bell-button group"
      :class="{ 'has-unread': notificationStore.hasUnread }"
      :title="$t('notifications.title')"
    >
      <Icon
        :icon="notificationStore.hasUnread ? 'mdi:bell-badge' : 'mdi:bell-outline'"
        class="h-6 w-6 transition-transform group-hover:scale-110"
        :class="{ 
          'animate-shake text-yellow-500 dark:text-yellow-400': notificationStore.hasUnread,
          'text-gray-600 dark:text-gray-300 group-hover:text-gray-900 dark:group-hover:text-white': !notificationStore.hasUnread
        }"
      />
      <span
        v-if="notificationStore.unreadCount > 0"
        class="notification-badge"
      >
        {{ notificationStore.unreadCount > 99 ? '99+' : notificationStore.unreadCount }}
      </span>
    </button>

    <!-- Notification Panel (Dropdown) -->
    <Transition name="slide-fade">
      <div 
        v-if="showPanel" 
        class="notification-panel fixed z-50 bg-white dark:bg-gray-800 rounded-xl shadow-2xl border border-gray-200 dark:border-gray-700 overflow-hidden flex flex-col"
        ref="panel" 
        :style="panelStyle"
      >
        <!-- Header -->
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50 flex items-center justify-between">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
            {{ $t('notifications.title') }}
          </h3>
          <div class="flex items-center gap-3">
            <button
              v-if="notificationStore.hasUnread"
              @click="markAllAsRead"
              class="text-xs text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300 font-medium transition-colors"
              :title="$t('notifications.markAllRead')"
            >
              {{ $t('notifications.markAllRead') }}
            </button>
            <button
              @click="openNotificationCenter"
              class="text-xs text-gray-600 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 font-medium transition-colors"
              :title="$t('notifications.viewAll')"
            >
              {{ $t('notifications.viewAll') }}
            </button>
          </div>
        </div>

        <!-- Notifications List -->
        <div class="flex-1 overflow-y-auto max-h-[380px] overscroll-contain">
          <div v-if="notificationStore.loading" class="flex flex-col items-center justify-center py-12 gap-2">
            <Icon icon="mdi:loading" class="h-6 w-6 animate-spin text-blue-600 dark:text-blue-400" />
            <span class="text-sm text-gray-600 dark:text-gray-400">{{ $t('common.loading') }}</span>
          </div>

          <div v-else-if="notificationStore.notifications.length === 0" class="flex flex-col items-center justify-center py-12 px-4 text-center">
            <div class="bg-gray-100 dark:bg-gray-700 rounded-full p-3 mb-3">
              <Icon icon="mdi:bell-off-outline" class="h-6 w-6 text-gray-400 dark:text-gray-500" />
            </div>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ $t('notifications.noNotifications') }}</p>
          </div>

          <div v-else class="divide-y divide-gray-100 dark:divide-gray-700/50">
            <div
              v-for="notification in recentNotifications"
              :key="notification.id"
              @click="handleNotificationClick(notification)"
              class="group relative flex items-start gap-3 p-3 cursor-pointer transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/50"
              :class="{ 'bg-blue-50/50 dark:bg-blue-900/10 hover:bg-blue-50 dark:hover:bg-blue-900/20': !notification.is_read }"
            >
              <!-- Icon -->
              <div 
                class="flex-shrink-0 h-9 w-9 rounded-lg flex items-center justify-center"
                :style="{ backgroundColor: notification.color + '20' }"
              >
                <Icon :icon="notification.icon" class="h-5 w-5" :style="{ color: notification.color }" />
              </div>

              <!-- Content -->
              <div class="flex-1 min-w-0">
                <div class="flex items-start justify-between gap-2 mb-0.5">
                  <p 
                    class="text-sm font-semibold truncate pr-6"
                    :class="notification.is_read ? 'text-gray-700 dark:text-gray-200' : 'text-gray-900 dark:text-white'"
                  >
                    {{ notification.title }}
                  </p>
                  <span class="text-[10px] text-gray-400 dark:text-gray-500 whitespace-nowrap flex-shrink-0">
                    {{ formatTime(notification.created_at) }}
                  </span>
                </div>
                
                <div class="flex items-center gap-2 mb-0.5">
                  <span v-if="notification.priority === 2" class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 uppercase tracking-wide">
                    {{ $t('notifications.urgent') }}
                  </span>
                </div>

                <p class="text-xs text-gray-600 dark:text-gray-400 line-clamp-2 leading-relaxed">
                  {{ notification.message }}
                </p>
              </div>

              <!-- Delete Action -->
              <button
                @click.stop="deleteNotification(notification.id)"
                class="absolute top-3 right-3 opacity-0 group-hover:opacity-100 p-1 rounded-md text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-all"
                :title="$t('notifications.delete')"
              >
                <Icon icon="mdi:close" class="h-4 w-4" />
              </button>

              <!-- Unread Indicator Dot -->
              <div 
                v-if="!notification.is_read"
                class="absolute bottom-3 right-3 h-2 w-2 rounded-full bg-blue-500 dark:bg-blue-400"
              ></div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="p-3 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50">
          <button
            @click="openNotificationCenter"
            class="w-full flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-blue-600 dark:text-blue-400 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 hover:bg-blue-50 dark:hover:bg-gray-700 transition-colors"
          >
            {{ $t('notifications.viewAll') }}
            <Icon icon="mdi:arrow-right" class="h-4 w-4" />
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notifications'
import { useAppStore } from '@/stores/app'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const notificationStore = useNotificationStore()
const appStore = useAppStore()

const showPanel = ref(false)
const panel = ref(null)
const panelStyle = ref({})

// Afficher seulement les 5 dernières notifications dans le panel
const recentNotifications = computed(() => {
  return notificationStore.notifications.slice(0, 5)
})

// Calculate optimal panel position
const calculatePanelPosition = () => {
  const button = document.querySelector('.notification-bell-button')
  if (!button) return
  
  const buttonRect = button.getBoundingClientRect()
  const windowWidth = window.innerWidth
  
  // Simple positioning: just below the button, within screen bounds
  let left = Math.min(buttonRect.left, windowWidth - 400) // 400px = panel width + margin
  left = Math.max(20, left) // Ensure minimum 20px from left edge
  
  panelStyle.value = {
    position: 'fixed',
    left: `${left}px`,
    top: `${buttonRect.bottom + 8}px`,
    width: '380px',
    maxWidth: `calc(100vw - ${left + 20}px)`
  }
}

// Toggle panel
const togglePanel = async () => {
  showPanel.value = !showPanel.value
  if (showPanel.value) {
    notificationStore.fetchNotifications({ page: 1, pageSize: 5 })
    await nextTick()
    calculatePanelPosition()
  }
}

// Ouvrir le centre de notifications
const openNotificationCenter = () => {
  showPanel.value = false
  router.push('/notifications')
}

// Marquer toutes comme lues
const markAllAsRead = async () => {
  try {
    await notificationStore.markAllAsRead()
  } catch (error) {
    console.error('Erreur lors du marquage:', error)
  }
}

// Gérer le clic sur une notification
const handleNotificationClick = async (notification) => {
  // Marquer comme lue
  if (!notification.is_read) {
    await notificationStore.markAsRead(notification.id)
  }

  // Rediriger si une URL est définie
  if (notification.action_url) {
    showPanel.value = false
    router.push(notification.action_url)
  }
}

// Supprimer une notification
const deleteNotification = async (id) => {
  try {
    await notificationStore.deleteNotification(id)
  } catch (error) {
    console.error('Erreur lors de la suppression:', error)
  }
}

// Formater le temps
const formatTime = (dateString) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now - date
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffMins < 1) return t('notifications.justNow')
  if (diffMins < 60) return t('notifications.minutesAgo', { count: diffMins })
  if (diffHours < 24) return t('notifications.hoursAgo', { count: diffHours })
  if (diffDays < 7) return t('notifications.daysAgo', { count: diffDays })

  return date.toLocaleDateString()
}

// Click outside handler
const handleClickOutside = (event) => {
  if (panel.value && !panel.value.contains(event.target) && !event.target.closest('.notification-bell-button')) {
    showPanel.value = false
  }
}

// Watch for sidebar changes that might affect button position
watch(() => appStore.sidebarOpen, () => {
  if (showPanel.value) {
    nextTick(() => {
      calculatePanelPosition()
    })
  }
})

// Handle window resize
const handleResize = () => {
  if (showPanel.value) {
    calculatePanelPosition()
  }
}

onMounted(() => {
  // Charger le compteur initial
  notificationStore.fetchUnreadCount()

  // Démarrer le polling (toutes les 30 secondes)
  notificationStore.startPolling(30000)

  // Ajouter le listener de clic extérieur
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  // Arrêter le polling
  notificationStore.stopPolling()

  // Retirer le listener
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.notification-bell-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 0.5rem;
  transition: all 0.2s;
  cursor: pointer;
  border: none;
  background-color: transparent;
}

.notification-bell-button:hover {
  background-color: rgba(156, 163, 175, 0.1);
}

.dark .notification-bell-button:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

.notification-badge {
  position: absolute;
  top: 4px;
  right: 4px;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background-color: #EF4444;
  color: white;
  font-size: 0.6rem;
  font-weight: 700;
  line-height: 1;
  box-shadow: 0 0 0 2px white;
}

.dark .notification-badge {
  box-shadow: 0 0 0 2px #1F2937;
}

@keyframes shake {
  0%, 100% { transform: rotate(0deg); }
  10%, 30%, 50%, 70%, 90% { transform: rotate(-10deg); }
  20%, 40%, 60%, 80% { transform: rotate(10deg); }
}

.animate-shake {
  animation: shake 0.5s ease-in-out;
}

/* Transitions */
.slide-fade-enter-active {
  transition: all 0.2s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.15s ease-in;
}

.slide-fade-enter-from {
  transform: translateY(-10px);
  opacity: 0;
}

.slide-fade-leave-to {
  transform: translateY(-10px);
  opacity: 0;
}
</style>
