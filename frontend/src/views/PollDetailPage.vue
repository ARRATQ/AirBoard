<template>
  <div class="poll-detail-page">
    <!-- Chargement -->
    <div v-if="loading" class="loading-state">
      <Icon icon="mdi:loading" class="spin loading-icon" />
      <p>{{ t('polls.loading') }}</p>
    </div>

    <!-- Erreur -->
    <div v-else-if="error" class="error-state">
      <Icon icon="mdi:alert-circle" class="error-icon" />
      <h3>{{ t('polls.detail.notFound') }}</h3>
      <p>{{ error }}</p>
      <button @click="$router.push('/polls')" class="btn-primary">
        <Icon icon="mdi:arrow-left" />
        {{ t('polls.detail.backToPolls') }}
      </button>
    </div>

    <!-- Détail du sondage -->
    <div v-else-if="poll" class="poll-detail-container">
      <!-- Breadcrumb -->
      <nav class="breadcrumb">
        <router-link to="/polls" class="breadcrumb-link">
          <Icon icon="mdi:poll" />
          {{ t('polls.detail.breadcrumbText') }}
        </router-link>
        <Icon icon="mdi:chevron-right" class="breadcrumb-separator" />
        <span class="breadcrumb-current">{{ poll.title }}</span>
      </nav>

      <!-- Carte du sondage -->
      <PollCard
        :poll="poll"
        :compact="false"
        :can-manage="canManage"
        @vote-success="handleVoteSuccess"
        @edit="openEditor"
        @delete="confirmDelete"
        @close="confirmClose"
      />

      <!-- Informations supplémentaires -->
      <div class="additional-info">
        <!-- Groupes cibles -->
        <div v-if="poll.target_groups && poll.target_groups.length > 0" class="info-section">
          <h3 class="info-title">
            <Icon icon="mdi:account-group" />
            {{ t('polls.detail.targetGroupsTitle') }}
          </h3>
          <div class="groups-list">
            <span
              v-for="group in poll.target_groups"
              :key="group.id"
              class="group-badge"
              :style="{ backgroundColor: group.color + '20', color: group.color }"
            >
              {{ group.name }}
            </span>
          </div>
        </div>

        <!-- Lié à un article -->
        <div v-if="poll.news" class="info-section">
          <h3 class="info-title">
            <Icon icon="mdi:newspaper" />
            {{ t('polls.detail.linkedArticleTitle') }}
          </h3>
          <router-link :to="`/news/${poll.news.slug}`" class="linked-item">
            <Icon icon="mdi:newspaper" class="linked-icon" />
            <div class="linked-content">
              <h4>{{ poll.news.title }}</h4>
              <p>{{ poll.news.summary }}</p>
            </div>
            <Icon icon="mdi:chevron-right" class="linked-arrow" />
          </router-link>
        </div>

        <!-- Lié à une annonce -->
        <div v-if="poll.announcement" class="info-section">
          <h3 class="info-title">
            <Icon icon="mdi:bullhorn" />
            {{ t('polls.detail.linkedAnnouncementTitle') }}
          </h3>
          <div class="linked-item">
            <Icon icon="mdi:bullhorn" class="linked-icon" />
            <div class="linked-content">
              <h4>{{ poll.announcement.title }}</h4>
              <p>{{ poll.announcement.content }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal d'édition -->
    <PollEditor
      :show="showEditor"
      :poll="poll"
      :groups="groups"
      @close="showEditor = false"
      @success="handleEditSuccess"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { pollsService, adminService } from '@/services/api'
import PollCard from '@/components/polls/PollCard.vue'
import PollEditor from '@/components/polls/PollEditor.vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { t } = useI18n()

const poll = ref(null)
const groups = ref([])
const loading = ref(false)
const error = ref(null)
const showEditor = ref(false)

// Vérifier si l'utilisateur peut gérer ce sondage
const canManage = computed(() => {
  if (!poll.value) return false
  const user = authStore.user

  if (user.role === 'admin') return true
  if (user.role === 'editor') return true

  if (user.role === 'group_admin') {
    // Peut gérer si auteur ou si le sondage cible un de ses groupes
    if (poll.value.author_id === user.id) return true

    const managedGroupIds = user.managed_group_ids || []
    const targetGroupIds = poll.value.target_groups?.map(g => g.id) || []

    return targetGroupIds.some(id => managedGroupIds.includes(id))
  }

  return false
})

// Charger le sondage
const loadPoll = async () => {
  try {
    loading.value = true
    error.value = null
    poll.value = await pollsService.getPollById(route.params.id)
  } catch (err) {
    console.error('Error loading poll:', err)
    error.value = err.response?.data?.error || t('polls.detail.loadError')
  } finally {
    loading.value = false
  }
}

// Charger les groupes (pour l'éditeur)
const loadGroups = async () => {
  try {
    const response = await adminService.getGroups()
    groups.value = response.data || response
  } catch (err) {
    console.error('Error loading groups:', err)
  }
}

// Gérer le succès du vote
const handleVoteSuccess = () => {
  loadPoll()
}

// Ouvrir l'éditeur
const openEditor = () => {
  if (groups.value.length === 0) {
    loadGroups()
  }
  showEditor.value = true
}

// Gérer le succès de l'édition
const handleEditSuccess = () => {
  loadPoll()
}

// Confirmer la suppression
const confirmDelete = async () => {
  if (!confirm(t('polls.actions.confirmDelete'))) {
    return
  }

  try {
    const user = authStore.user

    if (user.role === 'admin') {
      await pollsService.deletePoll(poll.value.id)
    } else if (user.role === 'editor') {
      await pollsService.deletePollAsEditor(poll.value.id)
    } else if (user.role === 'group_admin') {
      await pollsService.deletePollAsGroupAdmin(poll.value.id)
    }

    if (window.$toast) {
      window.$toast.success(t('polls.actions.deleteSuccess'))
    }

    router.push('/polls')
  } catch (err) {
    console.error('Erreur lors de la suppression:', err)
    if (window.$toast) {
      window.$toast.error(err.response?.data?.error || t('polls.actions.deleteError'))
    }
  }
}

// Confirmer la fermeture
const confirmClose = async () => {
  if (!confirm(t('polls.actions.confirmClose'))) {
    return
  }

  try {
    const user = authStore.user

    if (user.role === 'admin') {
      await pollsService.closePoll(poll.value.id)
    } else if (user.role === 'group_admin') {
      await pollsService.closePollAsGroupAdmin(poll.value.id)
    }

    if (window.$toast) {
      window.$toast.success(t('polls.actions.closeSuccess'))
    }

    await loadPoll()
  } catch (err) {
    console.error('Erreur lors de la fermeture:', err)
    if (window.$toast) {
      window.$toast.error(err.response?.data?.error || t('polls.actions.closeError'))
    }
  }
}

onMounted(() => {
  loadPoll()
})
</script>

<style scoped>
.poll-detail-page {
  max-width: 900px;
  margin: 0 auto;
  padding: 2rem;
}

.loading-state,
.error-state {
  text-align: center;
  padding: 4rem 2rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.loading-icon {
  font-size: 4rem;
  color: #8B5CF6;
  margin-bottom: 1rem;
}

.error-icon {
  font-size: 4rem;
  color: #ef4444;
  margin-bottom: 1rem;
}

.loading-state p,
.error-state p {
  color: #6b7280;
  margin: 0.5rem 0 1rem 0;
}

.error-state h3 {
  color: #1f2937;
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
}

.breadcrumb-link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  color: #6b7280;
  text-decoration: none;
  transition: color 0.2s;
}

.breadcrumb-link:hover {
  color: #8B5CF6;
}

.breadcrumb-separator {
  color: #d1d5db;
  font-size: 1rem;
}

.breadcrumb-current {
  color: #1f2937;
  font-weight: 500;
}

.poll-detail-container {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.additional-info {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.info-section {
  background: white;
  padding: 1.5rem;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.info-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 1rem 0;
}

.info-title svg {
  font-size: 1.5rem;
  color: #8B5CF6;
}

.groups-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.group-badge {
  padding: 0.5rem 1rem;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.linked-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #f9fafb;
  border-radius: 8px;
  text-decoration: none;
  transition: all 0.2s;
  cursor: pointer;
}

.linked-item:hover {
  background: #f3f4f6;
}

.linked-icon {
  font-size: 2rem;
  color: #8B5CF6;
  flex-shrink: 0;
}

.linked-content {
  flex: 1;
}

.linked-content h4 {
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.25rem 0;
}

.linked-content p {
  font-size: 0.875rem;
  color: #6b7280;
  margin: 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.linked-arrow {
  font-size: 1.5rem;
  color: #9ca3af;
  flex-shrink: 0;
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  background: #8B5CF6;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: #7c3aed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
