<template>
  <div class="feedback-widget">
    <div class="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ $t('feedback.helpful') }}
      </span>

      <div class="flex items-center gap-2">
        <!-- Thumbs Up -->
        <button
          :class="[
            'flex items-center gap-1 px-3 py-2 rounded-lg transition-all',
            userFeedback === 'positive'
              ? 'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300'
              : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-600'
          ]"
          :disabled="!isAuthenticated || isSubmitting"
          @click="toggleFeedback('positive')"
        >
          <Icon icon="mdi:thumb-up" class="h-5 w-5" />
          <span class="text-sm font-medium">{{ stats.positive_count }}</span>
        </button>

        <!-- Thumbs Down -->
        <button
          :class="[
            'flex items-center gap-1 px-3 py-2 rounded-lg transition-all',
            userFeedback === 'negative'
              ? 'bg-red-100 dark:bg-red-900 text-red-700 dark:text-red-300'
              : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-600'
          ]"
          :disabled="!isAuthenticated || isSubmitting"
          @click="toggleFeedback('negative')"
        >
          <Icon icon="mdi:thumb-down" class="h-5 w-5" />
          <span class="text-sm font-medium">{{ stats.negative_count }}</span>
        </button>
      </div>

      <!-- Percentage (optionnel) -->
      <div v-if="stats.total_count > 0" class="ml-auto text-sm text-gray-500 dark:text-gray-400">
        {{ positivePercentage }}% {{ $t('feedback.positive') }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { getFeedbackStats, addFeedback } from '@/services/api'

const props = defineProps({
  entityType: {
    type: String,
    required: true,
    validator: (value) => ['news', 'application', 'event'].includes(value)
  },
  entityId: {
    type: Number,
    required: true
  }
})

const authStore = useAuthStore()
const stats = ref({
  positive_count: 0,
  negative_count: 0,
  total_count: 0,
  user_feedback: null
})
const isSubmitting = ref(false)

const isAuthenticated = computed(() => authStore.isAuthenticated)

const userFeedback = computed(() => stats.value.user_feedback)

const positivePercentage = computed(() => {
  if (stats.value.total_count === 0) return 0
  return Math.round((stats.value.positive_count / stats.value.total_count) * 100)
})

const loadStats = async () => {
  try {
    const response = await getFeedbackStats(props.entityType, props.entityId)
    stats.value = response
  } catch (error) {
    console.error('Erreur lors du chargement des statistiques de feedback:', error)
  }
}

const toggleFeedback = async (type) => {
  if (!isAuthenticated.value || isSubmitting.value) return

  isSubmitting.value = true
  try {
    await addFeedback({
      entity_type: props.entityType,
      entity_id: props.entityId,
      feedback_type: type
    })
    await loadStats()
  } catch (error) {
    console.error('Erreur lors de l\'envoi du feedback:', error)
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  loadStats()
})
</script>
