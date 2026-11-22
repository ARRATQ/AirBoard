<template>
  <div class="w-full p-6">
    <!-- Header -->
    <PageHeader :title="t('polls.title')" :description="t('polls.subtitle')" />

    <!-- Filters Bar -->
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <!-- Search -->
        <div class="md:col-span-2">
          <div class="relative">
            <Icon icon="lucide:search" class="absolute start-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-5 w-5" />
            <input
              v-model="filters.search"
              type="text"
              :placeholder="t('polls.filters.searchPlaceholder')"
              class="w-full ps-10 pe-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
              @input="debounceSearch"
            />
          </div>
        </div>

        <!-- Status Filter -->
        <div>
          <select
            v-model="filters.status"
            class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
            @change="loadPolls"
          >
            <option value="">{{ t('polls.filters.status_all') }}</option>
            <option value="active">{{ t('polls.filters.status_active') }}</option>
            <option value="closed">{{ t('polls.filters.status_closed') }}</option>
          </select>
        </div>

        <!-- Type Filter -->
        <div>
          <select
            v-model="filters.type"
            class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
            @change="loadPolls"
          >
            <option value="">{{ t('polls.filters.type_all') }}</option>
            <option value="single">{{ t('polls.filters.type_single') }}</option>
            <option value="multiple">{{ t('polls.filters.type_multiple') }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Chargement -->
    <div v-if="loading" class="loading-state">
      <Icon icon="lucide:loader-circle" class="spin loading-icon" />
      <p>{{ t('polls.loading') }}</p>
    </div>

    <!-- Liste vide -->
    <div v-else-if="polls.length === 0" class="empty-state">
      <Icon icon="lucide:chart-column" class="empty-icon" />
      <h3>{{ t('polls.empty.title') }}</h3>
      <p>{{ t('polls.empty.description') }}</p>
    </div>

    <!-- Polls List -->
    <div v-else class="poll-grid-container">
      <div class="three-column-grid" ref="gridContainer">
        <PollCard
          v-for="poll in polls"
          :key="poll.id"
          :poll="poll"
          :compact="false"
          @vote-success="handleVoteSuccess"
          @view-details="viewPoll"
        />
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="pagination.total_pages > 1" class="mt-8 flex justify-center">
      <nav class="flex items-center gap-2">
        <button
          @click="changePage(pagination.page - 1)"
          :disabled="pagination.page === 1"
          class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <Icon icon="lucide:chevron-left" class="rtl-flip h-5 w-5" />
        </button>

        <button
          v-for="page in visiblePages"
          :key="page"
          @click="changePage(page)"
          :class="[
            'px-4 py-2 rounded-lg border',
            page === pagination.page
              ? 'bg-primary-500 text-white border-primary-500'
              : 'border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'
          ]"
        >
          {{ page }}
        </button>

        <button
          @click="changePage(pagination.page + 1)"
          :disabled="pagination.page === pagination.total_pages"
          class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <Icon icon="lucide:chevron-right" class="rtl-flip h-5 w-5" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { pollsService } from '@/services/api'
import PollCard from '@/components/polls/PollCard.vue'
import PageHeader from '@/components/ui/PageHeader.vue'

const router = useRouter()
const { t } = useI18n()

const polls = ref([])
const loading = ref(false)
const gridContainer = ref(null)
const filters = ref({
  status: '',
  type: '',
  search: ''
})
const pagination = ref({
  page: 1,
  page_size: 12,
  total: 0,
  total_pages: 1
})

// Computed
const visiblePages = computed(() => {
  const pages = []
  const start = Math.max(1, pagination.value.page - 2)
  const end = Math.min(pagination.value.total_pages, pagination.value.page + 2)

  for (let i = start; i <= end; i++) {
    pages.push(i)
  }

  return pages
})

let searchTimeout = null

// Removed complex grid rearrangement logic to prevent overlaps
// The grid now handles expansion automatically with CSS

// Charger les sondages
const loadPolls = async () => {
  try {
    loading.value = true
    const response = await pollsService.getPolls({
      ...filters.value,
      page: pagination.value.page,
      page_size: pagination.value.page_size
    })

    polls.value = response.polls
    pagination.value = {
      page: response.page,
      page_size: response.page_size,
      total: response.total,
      total_pages: response.total_pages
    }

    // Grid now handles layout automatically with CSS
  } catch (error) {
    console.error(t('polls.actions.loadError'), error)
    if (window.$toast) {
      window.$toast.error(t('polls.actions.loadError'))
    }
  } finally {
    loading.value = false
  }
}

// Debounce pour la recherche
const debounceSearch = () => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    pagination.value.page = 1
    loadPolls()
  }, 500)
}

// Changer de page
const changePage = (newPage) => {
  if (newPage >= 1 && newPage <= pagination.value.total_pages) {
    pagination.value.page = newPage
    loadPolls()
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

// Voir un sondage
const viewPoll = (pollId) => {
  router.push(`/polls/${pollId}`)
}

// Gérer le succès du vote
const handleVoteSuccess = (pollId) => {
  // Recharger les sondages pour mettre à jour les stats
  loadPolls()
}

onMounted(() => {
  loadPolls()
})
</script>

<style scoped>
/* Masonry Grid Layout - 3 columns with variable heights */
.poll-grid-container {
  contain: layout;
  min-height: 0;
}

/* Colonnes d'au moins 300 px selon la largeur réelle de la zone de contenu (barre latérale
   ouverte ou repliée) ; minmax(0, …) empêche un long titre ou un mot sans espace d'élargir
   sa colonne au-delà de la page. */
.three-column-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 300px), 1fr));
  gap: 1.25rem;
  align-items: start;
  grid-auto-flow: dense;
}

.three-column-grid > * { min-width: 0; }

/* Poll cards with masonry layout support */
.three-column-grid .poll-card {
  /* Masonry: height adapts to content automatically */
  height: auto;
  min-height: 4rem;
  /* Prevent content overflow */
  overflow: visible;
  /* Smooth transitions */
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  /* Prevent layout conflicts */
  contain: layout;
  /* Ensure grid item spans are reset for masonry */
  grid-column: auto;
  grid-row: auto;
}

/* Ouvert, un sondage prend toute la largeur de la grille au lieu de rester dans sa colonne
   de 300 px ; « dense » laisse les cartes fermées combler la place libérée. */
.three-column-grid .poll-card.is-expanded {
  grid-column: 1 / -1;
  /* Allow natural expansion in masonry layout */
  height: auto;
  min-height: 6rem;
  /* Smooth transition to new size */
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  /* Ensure content stays visible but contained */
  overflow: visible;
}

/* Fade Scale Animation for Grid View */
.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: all 0.3s ease;
}

.fade-scale-enter-from {
  opacity: 0;
  transform: scale(0.95);
}

.fade-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

.fade-scale-move {
  transition: transform 0.3s ease;
}


</style>
