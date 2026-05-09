import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { setLocale as setI18nLocale, resolveInitialLocale } from '@/i18n'
import { getPalette, generateCustomPalette, generatePaletteFromColors } from '@/config/palettes'

export const useAppStore = defineStore('app', () => {
  // Déterminer si on est sur mobile pour l'état initial de la sidebar
  const isMobile = () => typeof window !== 'undefined' && window.innerWidth < 1024

  // État
  const isDarkMode = ref(false)
  const currentPalette = ref(localStorage.getItem('airboard_palette') || 'claude')
  const customPalettes = ref(
    (() => { try { return JSON.parse(localStorage.getItem('airboard_custom_palettes') || '[]') } catch { return [] } })()
  )
  const sidebarOpen = ref(!isMobile()) // Fermée par défaut sur mobile, ouverte sur desktop
  const showNavbar = ref(true)
  const isLoading = ref(false)
  const notifications = ref([])
  const appSettings = ref(null)
  const settingsLastUpdated = ref(Date.now())
  const locale = ref(resolveInitialLocale())
  const zoomLevel = ref('100')

  // Getters
  const hasNotifications = computed(() => notifications.value.length > 0)
  const unreadNotifications = computed(() =>
    notifications.value.filter(n => !n.read)
  )
  const unreadCount = computed(() => unreadNotifications.value.length)
  const zoomScale = computed(() => parseInt(zoomLevel.value) / 100)

  const _injectPaletteCSS = (paletteObj) => {
    const toKebab = (str) => str.replace(/([A-Z])/g, m => '-' + m.toLowerCase())
    const makeCSSVars = (vars) =>
      Object.entries(vars).map(([k, v]) => `  --${toKebab(k)}: ${v};`).join('\n')
    const css = `:root {\n${makeCSSVars(paletteObj.light)}\n}\nhtml.dark {\n${makeCSSVars(paletteObj.dark)}\n}`
    let el = document.getElementById('palette-vars')
    if (!el) {
      el = document.createElement('style')
      el.id = 'palette-vars'
      document.head.appendChild(el)
    }
    el.textContent = css
  }

  // Applique les CSS vars d'un objet palette sans changer currentPalette (pour la prévisualisation).
  const previewPaletteColors = (paletteObj) => _injectPaletteCSS(paletteObj)

  // Applique une palette en injectant les CSS custom properties sur <html>
  const applyPalette = (slug, customHex = null) => {
    let palette
    if (slug === 'custom') {
      const hex = customHex || localStorage.getItem('airboard_custom_accent') || '#d97757'
      palette = generateCustomPalette(hex)
      if (customHex) localStorage.setItem('airboard_custom_accent', customHex)
    } else if (slug.startsWith('cp_')) {
      const cp = customPalettes.value.find(p => p.id === slug)
      if (cp) {
        palette = (cp.auto_bg !== false || !cp.bg)
          ? generateCustomPalette(cp.accent)
          : generatePaletteFromColors(cp.accent, cp.bg)
      } else {
        palette = getPalette('claude')
      }
    } else {
      palette = getPalette(slug)
    }
    currentPalette.value = slug
    localStorage.setItem('airboard_palette', slug)
    _injectPaletteCSS(palette)
  }

  const setCustomPalettes = (palettes) => {
    customPalettes.value = Array.isArray(palettes) ? palettes : []
    localStorage.setItem('airboard_custom_palettes', JSON.stringify(customPalettes.value))
  }

  // Actions
  const toggleDarkMode = () => {
    isDarkMode.value = !isDarkMode.value
    updateTheme()
    localStorage.setItem('airboard_dark_mode', isDarkMode.value.toString())
  }

  const setDarkMode = (value) => {
    isDarkMode.value = value
    updateTheme()
    localStorage.setItem('airboard_dark_mode', value.toString())
  }

  const updateTheme = () => {
    const htmlElement = document.documentElement
    if (isDarkMode.value) {
      htmlElement.classList.add('dark')
    } else {
      htmlElement.classList.remove('dark')
    }
  }

  const toggleSidebar = () => {
    sidebarOpen.value = !sidebarOpen.value
    localStorage.setItem('airboard_sidebar_open', sidebarOpen.value.toString())
  }

  const setSidebarOpen = (value) => {
    sidebarOpen.value = value
    localStorage.setItem('airboard_sidebar_open', value.toString())
  }

  const toggleNavbar = () => {
    showNavbar.value = !showNavbar.value
    localStorage.setItem('airboard_show_navbar', showNavbar.value.toString())
  }

  const setShowNavbar = (value) => {
    showNavbar.value = value
    localStorage.setItem('airboard_show_navbar', value.toString())
  }

  const setZoomLevel = (value) => {
    zoomLevel.value = value
    localStorage.setItem('airboard_zoom_level', value)
  }

  const setLoading = (value) => {
    isLoading.value = value
  }

  const addNotification = (notification) => {
    const id = Date.now().toString()
    const newNotification = {
      id,
      type: 'info', // info, success, warning, error
      title: '',
      message: '',
      read: false,
      createdAt: new Date(),
      ...notification
    }

    notifications.value.unshift(newNotification)

    // Auto-remove après un délai (5 secondes pour succès/info, 8 secondes pour erreurs/warnings)
    const autoRemoveDelay = notification.type === 'error' || notification.type === 'warning' ? 8000 : 5000
    setTimeout(() => {
      removeNotification(id)
    }, autoRemoveDelay)

    return id
  }

  const removeNotification = (id) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
    }
  }

  const markNotificationAsRead = (id) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification) {
      notification.read = true
    }
  }

  const clearAllNotifications = () => {
    notifications.value = []
  }

  const loadFromStorage = () => {
    try {
      // Mode sombre
      const storedDarkMode = localStorage.getItem('airboard_dark_mode')
      if (storedDarkMode !== null) {
        isDarkMode.value = storedDarkMode === 'true'
      } else {
        // Détecter la préférence système
        isDarkMode.value = window.matchMedia('(prefers-color-scheme: dark)').matches
      }

      // Sidebar - sur mobile, toujours fermée au démarrage
      const storedSidebarOpen = localStorage.getItem('airboard_sidebar_open')
      if (isMobile()) {
        sidebarOpen.value = false
      } else if (storedSidebarOpen !== null) {
        sidebarOpen.value = storedSidebarOpen === 'true'
      }

      // Navbar visibility
      const storedShowNavbar = localStorage.getItem('airboard_show_navbar')
      if (storedShowNavbar !== null) {
        showNavbar.value = storedShowNavbar === 'true'
      }

      // Zoom level
      const storedZoomLevel = localStorage.getItem('airboard_zoom_level')
      if (storedZoomLevel) {
        zoomLevel.value = storedZoomLevel
      }

      // Appliquer le thème
      updateTheme()

      // Appliquer la palette depuis localStorage (synchrone → zero flash)
      const storedPalette = localStorage.getItem('airboard_palette')
      if (storedPalette) {
        const customHex = storedPalette === 'custom' ? localStorage.getItem('airboard_custom_accent') : null
        applyPalette(storedPalette, customHex)
      } else {
        applyPalette('claude')
      }

      // Langue
      const storedLocale = localStorage.getItem('airboard_locale')
      if (storedLocale) {
        locale.value = storedLocale
        setI18nLocale(storedLocale)
      } else {
        setI18nLocale(locale.value)
      }
    } catch (error) {
      console.error('Erreur lors du chargement des préférences:', error)
    }
  }

  const setAppLocale = (newLocale) => {
    if (!newLocale) return
    locale.value = newLocale
    setI18nLocale(newLocale)
  }

  // Écouter les changements de préférence système
  const initSystemThemeWatcher = () => {
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')

    const handleChange = (e) => {
      // Seulement si l'utilisateur n'a pas de préférence sauvegardée
      if (localStorage.getItem('airboard_dark_mode') === null) {
        setDarkMode(e.matches)
      }
    }

    mediaQuery.addEventListener('change', handleChange)

    // Cleanup function
    return () => mediaQuery.removeEventListener('change', handleChange)
  }

  // Gérer le redimensionnement de la fenêtre pour la sidebar
  const initResizeWatcher = () => {
    let resizeTimeout

    const handleResize = () => {
      clearTimeout(resizeTimeout)
      resizeTimeout = setTimeout(() => {
        // Si on passe en mode mobile et que la sidebar est ouverte, la fermer
        if (isMobile() && sidebarOpen.value) {
          sidebarOpen.value = false
        }
      }, 150) // Debounce de 150ms
    }

    window.addEventListener('resize', handleResize)

    // Cleanup function
    return () => {
      clearTimeout(resizeTimeout)
      window.removeEventListener('resize', handleResize)
    }
  }

  // Helpers pour les notifications toast
  const showSuccess = (message, title = 'Succès') => {
    return addNotification({
      type: 'success',
      title,
      message
    })
  }

  const showError = (message, title = 'Erreur') => {
    return addNotification({
      type: 'error',
      title,
      message
    })
  }

  const showWarning = (message, title = 'Attention') => {
    return addNotification({
      type: 'warning',
      title,
      message
    })
  }

  const showInfo = (message, title = 'Information') => {
    return addNotification({
      type: 'info',
      title,
      message
    })
  }

  // App Settings Management
  const setAppSettings = (settings) => {
    appSettings.value = settings
    settingsLastUpdated.value = Date.now()
  }

  const refreshAppSettings = async () => {
    try {
      // Import dynamique pour éviter les dépendances circulaires
      const { adminService } = await import('@/services/api')
      const settings = await adminService.getAppSettings()
      setAppSettings(settings)
      return settings
    } catch (error) {
      console.error('Failed to refresh app settings:', error)
      throw error
    }
  }

  const invalidateSettings = () => {
    settingsLastUpdated.value = Date.now()
  }

  return {
    // État
    isDarkMode,
    currentPalette,
    customPalettes,
    sidebarOpen,
    showNavbar,
    isLoading,
    notifications,
    appSettings,
    settingsLastUpdated,
    locale,
    zoomLevel,

    // Getters
    hasNotifications,
    unreadNotifications,
    unreadCount,
    zoomScale,

    // Actions
    applyPalette,
    setCustomPalettes,
    previewPaletteColors,
    toggleDarkMode,
    setDarkMode,
    toggleSidebar,
    setSidebarOpen,
    toggleNavbar,
    setShowNavbar,
    setZoomLevel,
    setLoading,
    addNotification,
    removeNotification,
    markNotificationAsRead,
    clearAllNotifications,
    loadFromStorage,
    initSystemThemeWatcher,
    initResizeWatcher,
    setAppLocale,

    // App Settings
    setAppSettings,
    refreshAppSettings,
    invalidateSettings,

    // Helpers
    showSuccess,
    showError,
    showWarning,
    showInfo,
  }
})