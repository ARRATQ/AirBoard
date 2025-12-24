<template>
  <div class="notifications-page p-4 sm:p-8 max-w-7xl mx-auto">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-8">
      <div>
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
          <Icon icon="mdi:bell" class="h-8 w-8 text-blue-600 dark:text-blue-400" />
          {{ $t('notifications.center') }}
        </h1>
        <p class="mt-2 text-gray-600 dark:text-gray-400">
          {{ $t('notifications.manageNotifications') }}
        </p>
      </div>
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <button
          v-if="notificationStore.hasUnread"
          @click="markAllAsRead"
          class="flex-1 sm:flex-none btn-primary flex items-center justify-center gap-2"
        >
          <Icon icon="mdi:check-all" class="h-5 w-5" />
          {{ $t('notifications.markAllRead') }}
        </button>
        <button
          v-if="notificationStore.readNotifications.length > 0"
          @click="deleteAllRead"
          class="flex-1 sm:flex-none btn-secondary flex items-center justify-center gap-2"
        >
          <Icon icon="mdi:delete-sweep" class="h-5 w-5" />
          {{ $t('notifications.deleteAllRead') }}
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      <!-- Total -->
      <div class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700 flex items-center gap-4">
        <div class="h-12 w-12 rounded-xl bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
          <Icon icon="mdi:bell-badge" class="h-6 w-6 text-blue-600 dark:text-blue-400" />
        </div>
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('notifications.total') }}</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ notificationStore.total }}</p>
        </div>
      </div>

      <!-- Unread -->
      <div class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700 flex items-center gap-4">
        <div class="h-12 w-12 rounded-xl bg-yellow-50 dark:bg-yellow-900/20 flex items-center justify-center">
          <Icon icon="mdi:bell-ring" class="h-6 w-6 text-yellow-600 dark:text-yellow-400" />
        </div>
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('notifications.unread') }}</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ notificationStore.unreadCount }}</p>
        </div>
      </div>

      <!-- Read -->
      <div class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700 flex items-center gap-4">
        <div class="h-12 w-12 rounded-xl bg-green-50 dark:bg-green-900/20 flex items-center justify-center">
          <Icon icon="mdi:bell-check" class="h-6 w-6 text-green-600 dark:text-green-400" />
        </div>
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('notifications.read') }}</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ notificationStore.readNotifications.length }}</p>
        </div>
      </div>

      <!-- Urgent -->
      <div class="bg-white dark:bg-gray-800 rounded-xl p-6 shadow-sm border border-gray-200 dark:border-gray-700 flex items-center gap-4">
        <div class="h-12 w-12 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
          <Icon icon="mdi:bell-alert" class="h-6 w-6 text-red-600 dark:text-red-400" />
        </div>
        <div>
          <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('notifications.urgent') }}</p>
          <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ notificationStore.urgentNotifications.length }}</p>
        </div>
      </div>
    </div>

    <!-- Filters -->
    <div class="flex items-center gap-3 mb-6">
      <button
        @click="notificationStore.toggleUnreadOnly"
        class="flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all border"
        :class="notificationStore.showUnreadOnly 
          ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 border-blue-200 dark:border-blue-800' 
          : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'"
      >
        <Icon icon="mdi:filter" class="h-4 w-4" />
        {{ notificationStore.showUnreadOnly ? $t('notifications.showAll') : $t('notifications.showUnreadOnly') }}
      </button>
    </div>

    <!-- Notifications List -->
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden min-h-[400px]">
      <div v-if="notificationStore.loading" class="flex flex-col items-center justify-center py-20">
        <Icon icon="mdi:loading" class="h-12 w-12 animate-spin text-blue-600 dark:text-blue-400" />
        <p class="text-gray-600 dark:text-gray-400 mt-4">{{ $t('common.loading') }}</p>
      </div>

      <div v-else-if="notificationStore.notifications.length === 0" class="flex flex-col items-center justify-center py-20 px-4 text-center">
        <div class="bg-gray-100 dark:bg-gray-700 rounded-full p-6 mb-4">
          <Icon icon="mdi:bell-off-outline" class="h-16 w-16 text-gray-400 dark:text-gray-500" />
        </div>
        <h3 class="text-xl font-semibold text-gray-900 dark:text-white">{{ $t('notifications.noNotifications') }}</h3>
        <p class="text-gray-500 dark:text-gray-400 mt-2">{{ $t('notifications.allCaughtUp') }}</p>
      </div>

      <div v-else class="divide-y divide-gray-100 dark:divide-gray-700">
        <TransitionGroup name="list">
          <div
            v-for="notification in notificationStore.notifications"
            :key="notification.id"
            class="group relative flex items-start gap-4 p-5 transition-colors hover:bg-gray-50 dark:hover:bg-gray-700/30"
            :class="{
              'bg-blue-50/40 dark:bg-blue-900/10': !notification.is_read,
              'border-l-4 border-l-red-500': notification.priority === 2,
              'border-l-4 border-l-yellow-500': notification.priority === 1,
              'border-l-4 border-l-transparent': notification.priority === 0
            }"
          >
            <!-- Priority Badge for Mobile/Compact -->
            <div v-if="notification.priority > 0" 
              class="absolute top-2 right-2 flex items-center justify-center w-6 h-6 rounded-full"
              :class="{
                'bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400': notification.priority === 2,
                'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400': notification.priority === 1
              }"
            >
              <Icon icon="mdi:alert" class="h-3.5 w-3.5" />
            </div>

            <!-- Icon -->
            <div 
              class="flex-shrink-0 h-12 w-12 rounded-xl flex items-center justify-center shadow-sm"
              :style="{ backgroundColor: notification.color + '20' }"
            >
              <Icon :icon="notification.icon" class="h-6 w-6" :style="{ color: notification.color }" />
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0 py-0.5">
              <div class="flex items-start justify-between gap-4 mb-1">
                <h4 class="text-base font-semibold text-gray-900 dark:text-white flex items-center flex-wrap gap-2">
                  {{ notification.title }}
                  <span v-if="notification.priority === 2" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 uppercase tracking-wide">
                    {{ $t('notifications.urgent') }}
                  </span>
                  <span v-else-if="notification.priority === 1" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-bold bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400 uppercase tracking-wide">
                    {{ $t('notifications.important') }}
                  </span>
                </h4>
                <span class="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap flex-shrink-0 mt-1">
                  {{ formatTime(notification.created_at) }}
                </span>
              </div>

              <p class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed mb-3">
                {{ notification.message }}
              </p>

              <div class="flex items-center flex-wrap gap-3">
                <span 
                  class="inline-flex items-center px-2.5 py-0.5 rounded-md text-xs font-medium"
                  :style="{ backgroundColor: notification.color + '20', color: notification.color }"
                >
                  {{ $t(`notifications.types.${notification.type}`) }}
                </span>
                
                <span v-if="notification.is_read" class="inline-flex items-center gap-1 text-xs text-green-600 dark:text-green-400">
                  <Icon icon="mdi:check" class="h-3.5 w-3.5" />
                  {{ $t('notifications.readAt') }}: {{ formatTime(notification.read_at) }}
                </span>
              </div>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2 self-center sm:self-start sm:mt-1 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity">
              <button
                v-if="!notification.is_read"
                @click="markAsRead(notification.id)"
                class="p-2 rounded-lg text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition-colors"
                :title="$t('notifications.markAsRead')"
              >
                <Icon icon="mdi:check" class="h-5 w-5" />
              </button>
              
              <button
                v-if="notification.action_url"
                @click="navigateTo(notification)"
                class="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                :title="$t('notifications.view')"
              >
                <Icon icon="mdi:eye" class="h-5 w-5" />
              </button>
              
              <button
                @click="deleteNotification(notification.id)"
                class="p-2 rounded-lg text-gray-400 hover:text-red-600 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors"
                :title="$t('notifications.delete')"
              >
                <Icon icon="mdi:delete" class="h-5 w-5" />
              </button>
            </div>
          </div>
        </TransitionGroup>
      </div>

      <!-- Pagination -->
      <div v-if="notificationStore.totalPages > 1" class="border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/50 p-4 flex items-center justify-center gap-4">
        <button
          @click="notificationStore.loadPreviousPage"
          :disabled="notificationStore.page === 1"
          class="flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Icon icon="mdi:chevron-left" class="h-5 w-5" />
          {{ $t('common.previous') }}
        </button>

        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ $t('common.page') }} {{ notificationStore.page }} {{ $t('common.of') }} {{ notificationStore.totalPages }}
        </span>

        <button
          @click="notificationStore.loadNextPage"
          :disabled="notificationStore.page === notificationStore.totalPages"
          class="flex items-center gap-2 px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {{ $t('common.next') }}
          <Icon icon="mdi:chevron-right" class="h-5 w-5" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useNotificationStore } from '@/stores/notifications'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const notificationStore = useNotificationStore()

// Marquer toutes comme lues
const markAllAsRead = async () => {
  try {
    await notificationStore.markAllAsRead()
  } catch (error) {
    console.error('Erreur:', error)
  }
}

// Supprimer toutes les lues
const deleteAllRead = async () => {
  if (confirm(t('notifications.confirmDeleteAllRead'))) {
    try {
      await notificationStore.deleteAllRead()
    } catch (error) {
      console.error('Erreur:', error)
    }
  }
}

// Marquer comme lue
const markAsRead = async (id) => {
  try {
    await notificationStore.markAsRead(id)
  } catch (error) {
    console.error('Erreur:', error)
  }
}

// Supprimer une notification
const deleteNotification = async (id) => {
  try {
    await notificationStore.deleteNotification(id)
  } catch (error) {
    console.error('Erreur:', error)
  }
}

// Naviguer vers l'URL d'action
const navigateTo = async (notification) => {
  if (!notification.is_read) {
    await markAsRead(notification.id)
  }
  if (notification.action_url) {
    router.push(notification.action_url)
  }
}

// Formater le temps
const formatTime = (dateString) => {
  if (!dateString) return ''
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

onMounted(() => {
  notificationStore.fetchNotifications()
  notificationStore.fetchUnreadCount()
})
</script>

<style scoped>
/* Transitions */
.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.list-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
