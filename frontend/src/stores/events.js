import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { eventsService, adminEventsService, editorEventsService, groupAdminEventsService } from '@/services/api'

export const useEventsStore = defineStore('events', () => {
  // État
  const events = ref([])
  const calendarEvents = ref([])
  const recurringInstances = ref([])
  const categories = ref([])
  const currentEvent = ref(null)
  const analytics = ref(null)
  const isLoading = ref(false)
  const isLoadingCalendar = ref(false)
  const isLoadingCategories = ref(false)
  const isLoadingAnalytics = ref(false)
  const pagination = ref({
    page: 1,
    limit: 10,
    total: 0,
    pages: 0
  })

  // Getters
  const publishedEvents = computed(() => 
    events.value.filter(event => event.is_published)
  )

  const upcomingEvents = computed(() => {
    const now = new Date()
    return events.value.filter(event => {
      if (!event.is_published) return false
      const eventDate = new Date(event.start_date)
      return eventDate >= now
    }).sort((a, b) => new Date(a.start_date) - new Date(b.start_date))
  })

  const pastEvents = computed(() => {
    const now = new Date()
    return events.value.filter(event => {
      if (!event.is_published) return false
      const eventDate = new Date(event.start_date)
      return eventDate < now
    }).sort((a, b) => new Date(b.start_date) - new Date(a.start_date))
  })

  const eventsByCategory = computed(() => {
    const grouped = {}
    events.value.forEach(event => {
      if (event.category) {
        const categoryName = event.category.name
        if (!grouped[categoryName]) {
          grouped[categoryName] = []
        }
        grouped[categoryName].push(event)
      }
    })
    return grouped
  })

  const recurringEvents = computed(() => 
    events.value.filter(event => event.is_recurring)
  )

  // Actions
  const fetchEvents = async (params = {}) => {
    try {
      isLoading.value = true
      const response = await eventsService.getEvents(params)
      
      // Handle different response structures - API returns events in response.events
      events.value = response.events || response.data || []
      
      // Update pagination if available
      if (response.pagination) {
        pagination.value = { ...pagination.value, ...response.pagination }
      } else if (response.page && response.total) {
        // Handle direct pagination properties
        pagination.value = {
          page: response.page || 1,
          limit: response.page_size || 10,
          total: response.total || 0,
          pages: response.total_pages || 1
        }
      }
      
      return response
    } catch (error) {
      console.error('Erreur lors du chargement des événements:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const fetchCalendarEvents = async (startDate, endDate, params = {}) => {
    try {
      isLoadingCalendar.value = true
      const response = await eventsService.getCalendarView(startDate, endDate, params)
      
      // L'API retourne directement les données, pas dans response.data
      const events = response.Events || response.events || []
      const instances = response.RecurringInstances || response.recurring_instances || []
      
      // Validate and clean event data before storing
      calendarEvents.value = events.filter(event => {
        return event && event.start_date && !isNaN(new Date(event.start_date).getTime())
      })
      
      recurringInstances.value = instances.filter(instance => {
        return instance && instance.start_date && !isNaN(new Date(instance.start_date).getTime())
      })

      return response
    } catch (error) {
      console.error('Erreur lors du chargement du calendrier:', error)
      // Clear problematic data
      calendarEvents.value = []
      recurringInstances.value = []
      throw error
    } finally {
      isLoadingCalendar.value = false
    }
  }

  const fetchEventBySlug = async (slug) => {
    try {
      isLoading.value = true
      const response = await eventsService.getEventBySlug(slug)

      // L'API retourne directement l'événement, pas dans response.data
      currentEvent.value = response.data || response

      return response
    } catch (error) {
      console.error('Erreur lors du chargement de l\'événement:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const fetchCategories = async () => {
    try {
      isLoadingCategories.value = true
      const response = await eventsService.getCategories()
      
      // API returns categories directly or in response.data
      categories.value = response || response.data || []
      
      return response
    } catch (error) {
      console.error('Erreur lors du chargement des catégories:', error)
      throw error
    } finally {
      isLoadingCategories.value = false
    }
  }

  // Admin Actions
  const fetchAdminEvents = async (params = {}) => {
    try {
      isLoading.value = true
      const response = await adminEventsService.getEvents(params)
      
      // API returns EventListResponse structure directly: { events: [], total: 20, page: 1, page_size: 10, total_pages: 2 }
      // No need to extract response.data again
      const responseData = response
      
      // Extract events array from the response
      events.value = responseData.events || []
      
      // Update pagination from response data
      pagination.value = {
        page: responseData.page || 1,
        limit: responseData.page_size || 10,
        total: responseData.total || 0,
        pages: responseData.total_pages || 1
      }

      return response
    } catch (error) {
      console.error('Erreur lors du chargement des événements admin:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const createEvent = async (eventData) => {
    try {
      isLoading.value = true
      const response = await adminEventsService.createEvent(eventData)
      
      // Add to events list
      events.value.unshift(response.data)
      
      return response
    } catch (error) {
      console.error('Erreur lors de la création de l\'événement:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const updateEvent = async (slug, eventData) => {
    try {
      isLoading.value = true
      const response = await adminEventsService.updateEvent(slug, eventData)
      
      // Update in events list
      const index = events.value.findIndex(event => event.slug === slug)
      if (index !== -1) {
        events.value[index] = response.data
      }
      
      // Update current event if it matches
      if (currentEvent.value?.slug === slug) {
        currentEvent.value = response.data
      }
      
      return response
    } catch (error) {
      console.error('Erreur lors de la mise à jour de l\'événement:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const deleteEvent = async (id) => {
    try {
      isLoading.value = true
      await adminEventsService.deleteEvent(id)
      
      // Remove from events list
      events.value = events.value.filter(event => event.id !== id)
      
      // Clear current event if it matches
      if (currentEvent.value?.id === id) {
        currentEvent.value = null
      }
      
      return true
    } catch (error) {
      console.error('Erreur lors de la suppression de l\'événement:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const fetchAnalytics = async () => {
    try {
      isLoadingAnalytics.value = true
      const response = await adminEventsService.getAnalytics()
      
      analytics.value = response.data
      
      return response
    } catch (error) {
      console.error('Erreur lors du chargement des analytics:', error)
      throw error
    } finally {
      isLoadingAnalytics.value = false
    }
  }

  const createCategory = async (categoryData) => {
    try {
      isLoading.value = true
      const response = await adminEventsService.createCategory(categoryData)
      
      categories.value.unshift(response.data)
      
      return response
    } catch (error) {
      console.error('Erreur lors de la création de la catégorie:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const updateCategory = async (id, categoryData) => {
    try {
      isLoading.value = true
      const response = await adminEventsService.updateCategory(id, categoryData)
      
      // Update in categories list
      const index = categories.value.findIndex(category => category.id === id)
      if (index !== -1) {
        categories.value[index] = response.data
      }
      
      return response
    } catch (error) {
      console.error('Erreur lors de la mise à jour de la catégorie:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const deleteCategory = async (id) => {
    try {
      isLoading.value = true
      await adminEventsService.deleteCategory(id)
      
      categories.value = categories.value.filter(category => category.id !== id)
      
      return true
    } catch (error) {
      console.error('Erreur lors de la suppression de la catégorie:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // Group Admin Actions
  const fetchGroupAdminEvents = async (params = {}) => {
    try {
      isLoading.value = true
      const response = await groupAdminEventsService.getEvents(params)
      
      events.value = response.data || []
      
      return response
    } catch (error) {
      console.error('Erreur lors du chargement des événements de groupe:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const createGroupAdminEvent = async (eventData) => {
    try {
      isLoading.value = true
      const response = await groupAdminEventsService.createEvent(eventData)
      
      events.value.unshift(response.data)
      
      return response
    } catch (error) {
      console.error('Erreur lors de la création de l\'événement de groupe:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const updateGroupAdminEvent = async (slug, eventData) => {
    try {
      isLoading.value = true
      const response = await groupAdminEventsService.updateEvent(slug, eventData)
      
      // Update in events list
      const index = events.value.findIndex(event => event.slug === slug)
      if (index !== -1) {
        events.value[index] = response.data
      }
      
      return response
    } catch (error) {
      console.error('Erreur lors de la mise à jour de l\'événement de groupe:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const deleteGroupAdminEvent = async (id) => {
    try {
      isLoading.value = true
      await groupAdminEventsService.deleteEvent(id)
      
      events.value = events.value.filter(event => event.id !== id)
      
      return true
    } catch (error) {
      console.error('Erreur lors de la suppression de l\'événement de groupe:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // Editor Actions
  const createEditorEvent = async (eventData) => {
    try {
      isLoading.value = true
      const response = await editorEventsService.createEvent(eventData)
      
      events.value.unshift(response.data)
      
      return response
    } catch (error) {
      console.error('Erreur lors de la création de l\'événement (editor):', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const updateEditorEvent = async (slug, eventData) => {
    try {
      isLoading.value = true
      const response = await editorEventsService.updateEvent(slug, eventData)
      
      // Update in events list
      const index = events.value.findIndex(event => event.slug === slug)
      if (index !== -1) {
        events.value[index] = response.data
      }
      
      return response
    } catch (error) {
      console.error('Erreur lors de la mise à jour de l\'événement (editor):', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const deleteEditorEvent = async (id) => {
    try {
      isLoading.value = true
      await editorEventsService.deleteEvent(id)
      
      events.value = events.value.filter(event => event.id !== id)
      
      return true
    } catch (error) {
      console.error('Erreur lors de la suppression de l\'événement (editor):', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // Utility actions
  const clearCurrentEvent = () => {
    currentEvent.value = null
  }

  const clearEvents = () => {
    events.value = []
    calendarEvents.value = []
    currentEvent.value = null
  }

  const setCurrentEvent = (event) => {
    currentEvent.value = event
  }

  return {
    // État
    events,
    calendarEvents,
    recurringInstances,
    categories,
    currentEvent,
    analytics,
    isLoading,
    isLoadingCalendar,
    isLoadingCategories,
    isLoadingAnalytics,
    pagination,

    // Getters
    publishedEvents,
    upcomingEvents,
    pastEvents,
    eventsByCategory,
    recurringEvents,

    // Actions
    fetchEvents,
    fetchCalendarEvents,
    fetchEventBySlug,
    fetchCategories,
    
    // Admin Actions
    fetchAdminEvents,
    createEvent,
    updateEvent,
    deleteEvent,
    fetchAnalytics,
    createCategory,
    updateCategory,
    deleteCategory,
    
    // Group Admin Actions
    fetchGroupAdminEvents,
    createGroupAdminEvent,
    updateGroupAdminEvent,
    deleteGroupAdminEvent,
    
    // Editor Actions
    createEditorEvent,
    updateEditorEvent,
    deleteEditorEvent,
    
    // Utility actions
    clearCurrentEvent,
    clearEvents,
    setCurrentEvent,
  }
})