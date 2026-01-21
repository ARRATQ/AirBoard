import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService } from '@/services/api'

export const useAuthStore = defineStore('auth', () => {
  // État
  const user = ref(null)
  const token = ref(null)
  const refreshToken = ref(null)
  const isLoading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const userInitials = computed(() => {
    if (!user.value) return 'U'
    const firstName = user.value.first_name || user.value.username
    const lastName = user.value.last_name || ''
    
    if (firstName && lastName) {
      return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase()
    }
    return firstName.charAt(0).toUpperCase()
  })

  const userDisplayName = computed(() => {
    if (!user.value) return 'Utilisateur'

    if (user.value.first_name && user.value.last_name) {
      return `${user.value.first_name} ${user.value.last_name}`
    }

    return user.value.username
  })

  const isGroupAdmin = computed(() => {
    // Un utilisateur est group admin s'il administre au moins un groupe (indépendamment de son rôle)
    // Vérifier admin_of_groups (tableau de groupes) OU managed_group_ids (tableau d'IDs)
    const hasAdminOfGroups = user.value?.admin_of_groups && user.value.admin_of_groups.length > 0
    const hasManagedGroupIds = user.value?.managed_group_ids && user.value.managed_group_ids.length > 0
    return hasAdminOfGroups || hasManagedGroupIds
  })
  const isEditor = computed(() => user.value?.role === 'editor')
  const canManageContent = computed(() => {
    const hasAdminOfGroups = user.value?.admin_of_groups && user.value.admin_of_groups.length > 0
    const hasManagedGroupIds = user.value?.managed_group_ids && user.value.managed_group_ids.length > 0
    return user.value?.role === 'admin' ||
      user.value?.role === 'editor' ||
      hasAdminOfGroups ||
      hasManagedGroupIds
  })
  const managedGroupIds = computed(() => {
    // Récupérer les IDs des groupes administrés depuis admin_of_groups ou managed_group_ids
    if (user.value?.admin_of_groups && user.value.admin_of_groups.length > 0) {
      return user.value.admin_of_groups.map(g => g.id)
    }
    if (user.value?.managed_group_ids && user.value.managed_group_ids.length > 0) {
      return user.value.managed_group_ids
    }
    return []
  })

  // Helper pour décoder le JWT et extraire managed_group_ids
  const extractManagedGroupIdsFromToken = (jwtToken) => {
    try {
      const payload = jwtToken.split('.')[1]
      const decoded = JSON.parse(atob(payload))
      return decoded.managed_group_ids || []
    } catch (e) {
      console.warn('Erreur décodage JWT:', e)
      return []
    }
  }

  // Actions
  const login = async (credentials) => {
    try {
      isLoading.value = true
      const response = await authService.login(credentials)

      // Stocker les données dans l'ordre correct
      token.value = response.token
      refreshToken.value = response.refresh_token

      // Enrichir l'objet user avec managed_group_ids du JWT si absent
      const userData = { ...response.user }
      console.log('🔍 user avant enrichissement:', userData.managed_group_ids)
      if (!userData.managed_group_ids || userData.managed_group_ids.length === 0) {
        userData.managed_group_ids = extractManagedGroupIdsFromToken(response.token)
        console.log('🔑 managed_group_ids extrait du JWT:', userData.managed_group_ids)
      }
      user.value = userData

      // Persistance locale
      localStorage.setItem('airboard_token', response.token)
      localStorage.setItem('airboard_refresh_token', response.refresh_token)
      localStorage.setItem('airboard_user', JSON.stringify(userData))

      // Forcer la mise à jour des intercepteurs
      console.log('🔐 Token stocké:', response.token.substring(0, 20) + '...')
      console.log('👤 User final avec managed_group_ids:', userData.managed_group_ids)

      return response
    } catch (error) {
      console.error('Erreur de connexion:', error)
      // Nettoyer en cas d'erreur
      logout()
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const register = async (userData) => {
    try {
      isLoading.value = true
      const response = await authService.register(userData)
      
      // Stocker les données
      user.value = response.user
      token.value = response.token
      refreshToken.value = response.refresh_token

      // Persistance locale
      localStorage.setItem('airboard_token', response.token)
      localStorage.setItem('airboard_refresh_token', response.refresh_token)
      localStorage.setItem('airboard_user', JSON.stringify(response.user))

      return response
    } catch (error) {
      console.error('Erreur d\'inscription:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const logout = () => {
    // Nettoyer le store
    user.value = null
    token.value = null
    refreshToken.value = null

    // Nettoyer le localStorage
    authService.logout()
  }

  const loadFromStorage = () => {
    try {
      const storedToken = localStorage.getItem('airboard_token')
      const storedRefreshToken = localStorage.getItem('airboard_refresh_token')
      const storedUser = localStorage.getItem('airboard_user')

      if (storedToken && storedUser) {
        token.value = storedToken
        refreshToken.value = storedRefreshToken

        // Enrichir l'objet user avec managed_group_ids du JWT si absent
        const userData = JSON.parse(storedUser)
        if (!userData.managed_group_ids || userData.managed_group_ids.length === 0) {
          userData.managed_group_ids = extractManagedGroupIdsFromToken(storedToken)
        }
        user.value = userData
        return true
      }
    } catch (error) {
      console.error('Erreur lors du chargement depuis le storage:', error)
      logout()
    }
    return false
  }

  const updateProfile = async () => {
    try {
      const profile = await authService.getProfile()

      // Enrichir avec managed_group_ids du JWT si absent
      if (!profile.managed_group_ids || profile.managed_group_ids.length === 0) {
        const storedToken = localStorage.getItem('airboard_token')
        if (storedToken) {
          profile.managed_group_ids = extractManagedGroupIdsFromToken(storedToken)
        }
      }

      user.value = profile
      localStorage.setItem('airboard_user', JSON.stringify(profile))
      return profile
    } catch (error) {
      console.error('Erreur lors de la mise à jour du profil:', error)
      throw error
    }
  }

  const saveProfile = async (profileData) => {
    try {
      isLoading.value = true
      const updatedUser = await authService.updateProfile(profileData)
      user.value = updatedUser
      localStorage.setItem('airboard_user', JSON.stringify(updatedUser))
      return updatedUser
    } catch (error) {
      console.error('Erreur lors de la sauvegarde du profil:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const uploadAvatar = async (file) => {
    try {
      isLoading.value = true
      const updatedUser = await authService.uploadAvatar(file)
      user.value = updatedUser
      localStorage.setItem('airboard_user', JSON.stringify(updatedUser))
      return updatedUser
    } catch (error) {
      console.error('Erreur lors de l\'upload de l\'avatar:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const deleteAvatar = async () => {
    try {
      isLoading.value = true
      const updatedUser = await authService.deleteAvatar()
      user.value = updatedUser
      localStorage.setItem('airboard_user', JSON.stringify(updatedUser))
      return updatedUser
    } catch (error) {
      console.error('Erreur lors de la suppression de l\'avatar:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const setUser = (userData) => {
    user.value = userData
    localStorage.setItem('airboard_user', JSON.stringify(userData))
  }

  const setToken = (tokenData) => {
    token.value = tokenData
    localStorage.setItem('airboard_token', tokenData)
  }

  const setRefreshToken = (refreshTokenData) => {
    refreshToken.value = refreshTokenData
    localStorage.setItem('airboard_refresh_token', refreshTokenData)
  }

  const updateTokens = (newToken, newRefreshToken) => {
    token.value = newToken
    refreshToken.value = newRefreshToken
    localStorage.setItem('airboard_token', newToken)
    localStorage.setItem('airboard_refresh_token', newRefreshToken)
  }

  const autoLoginSSO = async () => {
    try {
      isLoading.value = true

      // Appeler la route SSO auto-login
      const response = await authService.ssoAutoLogin()

      // Stocker les données
      token.value = response.token
      refreshToken.value = response.refresh_token

      // Enrichir l'objet user avec managed_group_ids du JWT si absent
      const userData = { ...response.user }
      if (!userData.managed_group_ids || userData.managed_group_ids.length === 0) {
        userData.managed_group_ids = extractManagedGroupIdsFromToken(response.token)
      }
      user.value = userData

      // Persistance locale
      localStorage.setItem('airboard_token', response.token)
      localStorage.setItem('airboard_refresh_token', response.refresh_token)
      localStorage.setItem('airboard_user', JSON.stringify(userData))

      return response
    } catch (error) {
      // Si SSO échoue, ne pas afficher d'erreur (mode classique)
      return null
    } finally {
      isLoading.value = false
    }
  }

  return {
    // État
    user,
    token,
    refreshToken,
    isLoading,

    // Getters
    isAuthenticated,
    isAdmin,
    isGroupAdmin,
    isEditor,
    canManageContent,
    managedGroupIds,
    userInitials,
    userDisplayName,

    // Actions
    login,
    register,
    logout,
    loadFromStorage,
    updateProfile,
    saveProfile,
    uploadAvatar,
    deleteAvatar,
    setUser,
    setToken,
    setRefreshToken,
    updateTokens,
    autoLoginSSO,
  }
})