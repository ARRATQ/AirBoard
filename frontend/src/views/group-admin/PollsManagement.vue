<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-6">
    <div class="w-full">
      <!-- Header -->
      <div class="flex justify-between items-center mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <Icon icon="mdi:poll" class="h-8 w-8 text-purple-600" />
            {{ t('groupAdmin.pollsManagement.title') }}
          </h1>
          <p class="mt-1 text-gray-600 dark:text-gray-400">
            {{ t('groupAdmin.pollsManagement.subtitle') }}
          </p>
        </div>
        <button
          @click="openCreateModal"
          class="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium transition-colors flex items-center gap-2"
        >
          <Icon icon="mdi:plus" class="h-5 w-5" />
          {{ t('groupAdmin.pollsManagement.newPoll') }}
        </button>
      </div>

      <!-- Stats -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="p-3 bg-purple-100 dark:bg-purple-900/20 rounded-lg">
              <Icon icon="mdi:poll" class="h-6 w-6 text-purple-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('groupAdmin.pollsManagement.stats.total') }}</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ polls.length }}</p>
            </div>
          </div>
        </div>

        <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="p-3 bg-green-100 dark:bg-green-900/20 rounded-lg">
              <Icon icon="mdi:poll-check" class="h-6 w-6 text-green-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('groupAdmin.pollsManagement.stats.active') }}</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ polls.filter(p => p.is_active).length }}
              </p>
            </div>
          </div>
        </div>

        <div class="card p-4">
          <div class="flex items-center gap-3">
            <div class="p-3 bg-gray-100 dark:bg-gray-700 rounded-lg">
              <Icon icon="mdi:lock" class="h-6 w-6 text-gray-600" />
            </div>
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('groupAdmin.pollsManagement.stats.closed') }}</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ polls.filter(p => !p.is_active).length }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Filters -->
      <div class="card p-4 mb-6">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('groupAdmin.pollsManagement.filters.status') }}
            </label>
            <select
              v-model="filters.status"
              @change="loadPolls"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="">{{ t('groupAdmin.pollsManagement.filters.all') }}</option>
              <option value="active">{{ t('groupAdmin.pollsManagement.filters.active') }}</option>
              <option value="closed">{{ t('groupAdmin.pollsManagement.filters.closed') }}</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('groupAdmin.pollsManagement.filters.search') }}
            </label>
            <input
              v-model="filters.search"
              @input="debounceSearch"
              type="text"
              :placeholder="t('groupAdmin.pollsManagement.filters.searchPlaceholder')"
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            />
          </div>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="isLoading" class="flex justify-center items-center py-12">
        <Icon icon="mdi:loading" class="h-8 w-8 animate-spin text-purple-600" />
      </div>

      <!-- Polls List -->
      <div v-else-if="polls.length > 0" class="space-y-4">
        <div
          v-for="poll in polls"
          :key="poll.id"
          class="card p-6"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-3 mb-2">
                <span
                  :class="poll.is_active ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300'"
                  class="px-2 py-1 rounded-md text-xs font-medium"
                >
                  {{ poll.is_active ? t('common.active') : t('polls.status.closed') }}
                </span>
                <span class="px-2 py-1 rounded-md text-xs font-medium bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200">
                  {{ poll.poll_type === 'single' ? t('polls.filters.type_single') : t('polls.filters.type_multiple') }}
                </span>
                <span v-if="poll.is_anonymous" class="px-2 py-1 rounded-md text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                  <Icon icon="mdi:incognito" class="inline h-3 w-3" />
                  {{ t('polls.status.anonymous') }}
                </span>
              </div>

              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
                {{ poll.title }}
              </h3>

              <p v-if="poll.description" class="text-sm text-gray-600 dark:text-gray-400 mb-3">
                {{ poll.description }}
              </p>

              <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
                <span>
                  <Icon icon="mdi:calendar" class="inline h-4 w-4" />
                  {{ formatDate(poll.created_at) }}
                </span>
                <span v-if="poll.end_date">
                  <Icon icon="mdi:calendar-end" class="inline h-4 w-4" />
                  {{ t('polls.status.until') }} {{ formatDate(poll.end_date) }}
                </span>
                <span>
                  <Icon icon="mdi:format-list-checkbox" class="inline h-4 w-4" />
                  {{ poll.options?.length || 0 }} {{ t('polls.adminPollsManagement.details.options') }}
                </span>
                <span v-if="poll.target_groups && poll.target_groups.length > 0">
                  <Icon icon="mdi:account-group" class="inline h-4 w-4" />
                  {{ poll.target_groups.map(g => g.name).join(', ') }}
                </span>
              </div>
            </div>

            <div class="flex items-center gap-2 ml-4">
              <button
                @click="viewResults(poll)"
                class="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                :title="t('polls.actions.viewResults')"
              >
                <Icon icon="mdi:chart-bar" class="h-5 w-5" />
              </button>
              <button
                @click="openEditModal(poll)"
                class="p-2 text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-900/20 rounded-lg transition-colors"
                :title="t('polls.actions.edit')"
              >
                <Icon icon="mdi:pencil" class="h-5 w-5" />
              </button>
              <button
                v-if="poll.is_active && canClose(poll)"
                @click="closePoll(poll)"
                class="p-2 text-orange-600 hover:bg-orange-50 dark:hover:bg-orange-900/20 rounded-lg transition-colors"
                :title="t('polls.actions.close')"
              >
                <Icon icon="mdi:lock" class="h-5 w-5" />
              </button>
              <button
                @click="confirmDelete(poll)"
                class="p-2 text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                :title="t('polls.actions.delete')"
              >
                <Icon icon="mdi:delete" class="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-12">
        <Icon icon="mdi:poll-box-outline" class="h-16 w-16 mx-auto text-gray-400 mb-4" />
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          {{ t('groupAdmin.pollsManagement.empty.title') }}
        </h3>
        <p class="text-gray-600 dark:text-gray-400 mb-6">
          {{ t('groupAdmin.pollsManagement.empty.description') }}
        </p>
        <button
          @click="openCreateModal"
          class="px-4 py-2 bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium transition-colors"
        >
          {{ t('groupAdmin.pollsManagement.empty.createButton') }}
        </button>
      </div>
    </div>

    <!-- Poll Editor Modal -->
    <PollEditor
      :show="showModal"
      :poll="selectedPoll"
      :groups="managedGroups"
      @close="closeModal"
      @success="handleSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { pollsService } from '@/services/api'
import PollEditor from '@/components/polls/PollEditor.vue'
import { format } from 'date-fns'
import { fr } from 'date-fns/locale'

const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const polls = ref([])
const managedGroups = ref([])
const isLoading = ref(false)
const showModal = ref(false)
const selectedPoll = ref(null)

const filters = ref({
  status: '',
  search: ''
})

let searchTimeout = null

// Charger les sondages du group admin
const loadPolls = async () => {
  try {
    isLoading.value = true
    const response = await pollsService.getPollsAsGroupAdmin({
      ...filters.value,
      page: 1,
      page_size: 100
    })

    polls.value = response.polls
  } catch (error) {
    console.error('Erreur lors du chargement des sondages:', error)
  } finally {
    isLoading.value = false
  }
}

// Charger les groupes administrés
const loadManagedGroups = async () => {
  try {
    const response = await fetch('/api/v1/group-admin/managed-groups', {
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('airboard_token')}`
      }
    })
    const data = await response.json()
    managedGroups.value = data || []
  } catch (error) {
    console.error('Erreur lors du chargement des groupes:', error)
  }
}

// Vérifier si le user peut fermer le sondage
const canClose = (poll) => {
  const user = authStore.user
  return poll.author_id === user.id
}

// Debounce search
const debounceSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    loadPolls()
  }, 500)
}

// Ouvrir le modal de création
const openCreateModal = () => {
  selectedPoll.value = null
  showModal.value = true
}

// Ouvrir le modal d'édition
const openEditModal = (poll) => {
  selectedPoll.value = poll
  showModal.value = true
}

// Fermer le modal
const closeModal = () => {
  showModal.value = false
  selectedPoll.value = null
}

// Gérer le succès
const handleSuccess = () => {
  loadPolls()
}

// Voir les résultats
const viewResults = (poll) => {
  router.push(`/polls/${poll.id}`)
}

// Fermer un sondage
const closePoll = async (poll) => {
  if (!confirm(t('polls.actions.confirmClose'))) {
    return
  }

  try {
    await pollsService.closePollAsGroupAdmin(poll.id)
    if (window.$toast) {
      window.$toast.success(t('polls.actions.closeSuccess'))
    }
    await loadPolls()
  } catch (error) {
    console.error('Erreur lors de la fermeture:', error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.actions.closeError'))
    }
  }
}

// Confirmer la suppression
const confirmDelete = async (poll) => {
  if (!confirm(t('polls.adminPollsManagement.confirmation.deleteMessage', { title: poll.title }))) {
    return
  }

  try {
    await pollsService.deletePollAsGroupAdmin(poll.id)
    if (window.$toast) {
      window.$toast.success(t('polls.actions.deleteSuccess'))
    }
    await loadPolls()
  } catch (error) {
    console.error('Erreur lors de la suppression:', error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.actions.deleteError'))
    }
  }
}

// Formater la date
const formatDate = (date) => {
  if (!date) return ''
  return format(new Date(date), 'dd MMM yyyy', { locale: fr })
}

onMounted(() => {
  loadPolls()
  loadManagedGroups()
})
</script>

<style scoped>
.card {
  @apply bg-white dark:bg-gray-800 rounded-lg shadow-sm;
}
</style>
