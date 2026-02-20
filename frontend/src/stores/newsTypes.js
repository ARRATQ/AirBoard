import { defineStore } from 'pinia'
import { ref } from 'vue'
import { newsService } from '@/services/api'

export const useNewsTypesStore = defineStore('newsTypes', () => {
  const types = ref([])
  const loaded = ref(false)

  const fetchTypes = async () => {
    if (loaded.value) return
    try {
      types.value = await newsService.getTypes()
      loaded.value = true
    } catch (error) {
      console.error('Error fetching news types:', error)
    }
  }

  const getTypeBySlug = (slug) => {
    return types.value.find(t => t.slug === slug)
  }

  const refreshTypes = async () => {
    loaded.value = false
    await fetchTypes()
  }

  return {
    types,
    loaded,
    fetchTypes,
    getTypeBySlug,
    refreshTypes
  }
})
