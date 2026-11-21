import { defineStore } from 'pinia'
import { apiErrorMessage } from '@/services/apiError'
import { notificationService } from '@/services/api'

export const NOTIFICATION_POLL_MS = 60000

// Hors du state : un handler DOM n'a rien à faire dans l'état réactif.
let visibilityHandler = null

export const useNotificationStore = defineStore('notifications', {
  state: () => ({
    notifications: [],
    unreadCount: 0,
    loading: false,
    error: null,
    page: 1,
    pageSize: 20,
    total: 0,
    totalPages: 0,
    showUnreadOnly: false,
    pollInterval: null
  }),

  getters: {
    unreadNotifications: (state) => state.notifications.filter(n => !n.is_read),
    readNotifications: (state) => state.notifications.filter(n => n.is_read),
    hasUnread: (state) => state.unreadCount > 0,

    // Notifications groupées par priorité
    urgentNotifications: (state) => state.notifications.filter(n => n.priority === 2),
    importantNotifications: (state) => state.notifications.filter(n => n.priority === 1),
    normalNotifications: (state) => state.notifications.filter(n => n.priority === 0)
  },

  actions: {
    // Récupérer les notifications
    async fetchNotifications(params = {}) {
      this.loading = true
      this.error = null

      try {
        const queryParams = {
          page: params.page || this.page,
          page_size: params.pageSize || this.pageSize,
          unread: this.showUnreadOnly
        }

        const response = await notificationService.getNotifications(queryParams)

        this.notifications = response.data || []
        this.total = response.total || 0
        this.page = response.page || 1
        this.pageSize = response.page_size || 20
        this.totalPages = response.total_pages || 0
      } catch (error) {
        console.error('Erreur lors de la récupération des notifications:', error)
        this.error = apiErrorMessage(error, 'Erreur lors de la récupération des notifications')
      } finally {
        this.loading = false
      }
    },

    // Récupérer le nombre de notifications non lues
    async fetchUnreadCount() {
      try {
        const response = await notificationService.getUnreadCount()
        this.unreadCount = response.count || 0
      } catch (error) {
        console.error('Erreur lors de la récupération du compteur:', error)
      }
    },

    // Marquer une notification comme lue
    async markAsRead(id) {
      try {
        await notificationService.markAsRead(id)

        // Mettre à jour localement
        const notification = this.notifications.find(n => n.id === id)
        if (notification && !notification.is_read) {
          notification.is_read = true
          notification.read_at = new Date().toISOString()
          this.unreadCount = Math.max(0, this.unreadCount - 1)
        }
      } catch (error) {
        console.error('Erreur lors du marquage comme lu:', error)
        throw error
      }
    },

    // Marquer toutes les notifications comme lues
    async markAllAsRead() {
      try {
        await notificationService.markAllAsRead()

        // Mettre à jour localement
        this.notifications.forEach(n => {
          if (!n.is_read) {
            n.is_read = true
            n.read_at = new Date().toISOString()
          }
        })
        this.unreadCount = 0
      } catch (error) {
        console.error('Erreur lors du marquage de toutes comme lues:', error)
        throw error
      }
    },

    // Supprimer une notification
    async deleteNotification(id) {
      try {
        await notificationService.deleteNotification(id)

        // Retirer localement
        const index = this.notifications.findIndex(n => n.id === id)
        if (index !== -1) {
          const notification = this.notifications[index]
          if (!notification.is_read) {
            this.unreadCount = Math.max(0, this.unreadCount - 1)
          }
          this.notifications.splice(index, 1)
          this.total = Math.max(0, this.total - 1)
        }
      } catch (error) {
        console.error('Erreur lors de la suppression:', error)
        throw error
      }
    },

    // Supprimer toutes les notifications lues
    async deleteAllRead() {
      try {
        await notificationService.deleteAllRead()

        // Retirer localement
        this.notifications = this.notifications.filter(n => !n.is_read)
        this.total = this.notifications.length
      } catch (error) {
        console.error('Erreur lors de la suppression des notifications lues:', error)
        throw error
      }
    },

    // Basculer le filtre non lues uniquement
    toggleUnreadOnly() {
      this.showUnreadOnly = !this.showUnreadOnly
      this.fetchNotifications({ page: 1 })
    },

    // Charger la page suivante
    async loadNextPage() {
      if (this.page < this.totalPages) {
        await this.fetchNotifications({ page: this.page + 1 })
      }
    },

    // Charger la page précédente
    async loadPreviousPage() {
      if (this.page > 1) {
        await this.fetchNotifications({ page: this.page - 1 })
      }
    },

    // Démarrer le polling automatique du compteur.
    // En pause quand l'onglet est caché ; au retour, le compteur est relu tout de suite.
    startPolling(interval = NOTIFICATION_POLL_MS) {
      this.stopPolling()

      const startTimer = () => {
        if (this.pollInterval) return
        this.pollInterval = setInterval(() => {
          this.fetchUnreadCount()
        }, interval)
      }
      const stopTimer = () => {
        if (this.pollInterval) {
          clearInterval(this.pollInterval)
          this.pollInterval = null
        }
      }

      visibilityHandler = () => {
        if (document.hidden) {
          stopTimer()
        } else {
          this.fetchUnreadCount()
          startTimer()
        }
      }
      document.addEventListener('visibilitychange', visibilityHandler)

      if (!document.hidden) startTimer()
    },

    // Arrêter le polling
    stopPolling() {
      if (this.pollInterval) {
        clearInterval(this.pollInterval)
        this.pollInterval = null
      }
      if (visibilityHandler) {
        document.removeEventListener('visibilitychange', visibilityHandler)
        visibilityHandler = null
      }
    },

    // Réinitialiser le store
    reset() {
      this.stopPolling()
      this.notifications = []
      this.unreadCount = 0
      this.loading = false
      this.error = null
      this.page = 1
      this.total = 0
      this.totalPages = 0
      this.showUnreadOnly = false
    }
  }
})
