<template>
  <div class="polls-widget">
    <!-- Widget Header -->
    <div class="widget-header">
      <div class="header-left">
        <Icon icon="mdi:poll" class="header-icon" />
        <h3 class="widget-title">{{ $t('home.polls.title') }}</h3>
      </div>
      <router-link to="/polls" class="view-all-link">
        <span>{{ $t('home.viewAll') }}</span>
        <Icon icon="mdi:arrow-right" class="link-icon" />
      </router-link>
    </div>

    <!-- Empty State -->
    <div v-if="!polls || polls.length === 0" class="empty-state">
      <Icon icon="mdi:poll-box-outline" class="empty-icon" />
      <p class="empty-text">{{ $t('home.polls.noPolls') }}</p>
    </div>

    <!-- Polls List -->
    <div v-else class="polls-list">
      <!-- Nouveau sondage -->
      <div v-if="newestPoll" class="poll-item poll-new" @click="goToPoll(newestPoll)">
        <div class="poll-item-icon" style="background-color: #10b981;">
          <Icon icon="mdi:new-box" class="h-4 w-4 text-white" />
        </div>
        <div class="poll-item-content">
          <div class="poll-item-label">{{ $t('home.polls.newest') }}</div>
          <h4 class="poll-item-title">{{ newestPoll.title }}</h4>
          <div class="poll-item-meta">
            <Icon icon="mdi:calendar" class="h-3 w-3" />
            <span>{{ formatDate(newestPoll.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- Sondage bientôt fermé -->
      <div v-if="expiringSoonPoll" class="poll-item poll-expiring" @click="goToPoll(expiringSoonPoll)">
        <div class="poll-item-icon" style="background-color: #f59e0b;">
          <Icon icon="mdi:clock-alert" class="h-4 w-4 text-white" />
        </div>
        <div class="poll-item-content">
          <div class="poll-item-label">{{ $t('home.polls.expiringSoon') }}</div>
          <h4 class="poll-item-title">{{ expiringSoonPoll.title }}</h4>
          <div class="poll-item-meta">
            <Icon icon="mdi:calendar-end" class="h-3 w-3" />
            <span>{{ $t('home.polls.closesIn') }} {{ daysUntilClose(expiringSoonPoll.end_date) }}</span>
          </div>
        </div>
      </div>

      <!-- Dernier résultat de sondage fermé -->
      <div v-if="latestClosedPoll" class="poll-item poll-closed" @click="goToPoll(latestClosedPoll)">
        <div class="poll-item-icon" style="background-color: #6366f1;">
          <Icon icon="mdi:chart-bar" class="h-4 w-4 text-white" />
        </div>
        <div class="poll-item-content">
          <div class="poll-item-label">{{ $t('home.polls.latestResults') }}</div>
          <h4 class="poll-item-title">{{ latestClosedPoll.title }}</h4>
          <div class="poll-item-meta">
            <Icon icon="mdi:account-group" class="h-3 w-3" />
            <span>{{ $t('home.polls.votes', { count: latestClosedPoll.total_votes || 0 }) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { differenceInDays } from 'date-fns'

const router = useRouter()
const { t } = useI18n()

const props = defineProps({
  polls: {
    type: Array,
    required: true
  }
})

// Filtrer et trier les sondages
const newestPoll = computed(() => {
  // Sondage actif le plus récent
  const activePolls = props.polls.filter(p => p.is_active)
  if (activePolls.length === 0) return null
  return activePolls.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))[0]
})

const expiringSoonPoll = computed(() => {
  // Sondage actif qui expire dans les 3 prochains jours
  const now = new Date()
  const activePolls = props.polls.filter(p => {
    if (!p.is_active || !p.end_date) return false
    const daysLeft = differenceInDays(new Date(p.end_date), now)
    return daysLeft >= 0 && daysLeft <= 3
  })

  if (activePolls.length === 0) return null

  // Trier par date de fin (le plus proche en premier)
  return activePolls.sort((a, b) => new Date(a.end_date) - new Date(b.end_date))[0]
})

const latestClosedPoll = computed(() => {
  // Dernier sondage fermé
  const closedPolls = props.polls.filter(p => !p.is_active)
  if (closedPolls.length === 0) return null

  // On suppose que le backend retourne les sondages triés par date de création
  // Sinon, on peut trier ici
  return closedPolls.sort((a, b) => new Date(b.created_at) - new Date(a.created_at))[0]
})

const goToPoll = (poll) => {
  if (!poll) return
  router.push('/polls')
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const now = new Date()
  const diffInMs = now - date
  const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24))

  if (diffInDays === 0) return t('common.today')
  if (diffInDays === 1) return t('common.yesterday')
  if (diffInDays < 7) return t('common.daysAgo', { count: diffInDays })

  return date.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })
}

const daysUntilClose = (endDate) => {
  if (!endDate) return ''
  const days = differenceInDays(new Date(endDate), new Date())
  if (days === 0) return t('common.today')
  if (days === 1) return t('common.tomorrow')
  return t('common.daysRemaining', { count: days })
}
</script>

<style scoped>
.polls-widget {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Widget Header */
.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(139, 92, 246, 0.1);
}

.dark .widget-header {
  border-bottom-color: rgba(139, 92, 246, 0.2);
}

.dark .header-icon {
  color: #a78bfa;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-icon {
  font-size: 1.125rem;
  color: #8b5cf6;
}

.widget-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
  line-height: 1.2;
}

.dark .widget-title {
  color: white;
}

.view-all-link {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: #6366f1;
  text-decoration: none;
  transition: all 0.2s ease;
}

.view-all-link:hover {
  color: #4f46e5;
  gap: 0.5rem;
}

.link-icon {
  font-size: 1rem;
  transition: transform 0.2s ease;
}

.view-all-link:hover .link-icon {
  transform: translateX(2px);
}

/* Empty State */
.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem 0.5rem;
  text-align: center;
}

.empty-icon {
  font-size: 2rem;
  color: #9ca3af;
  margin-bottom: 0.375rem;
}

.empty-text {
  font-size: 0.875rem;
  color: #6b7280;
}

.dark .empty-text {
  color: #9ca3af;
}

/* Polls List */
.polls-list {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  flex: 1;
}

/* Poll Item */
.poll-item {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.625rem;
  background-color: rgb(255 255 255);
  border: 1px solid rgb(229 231 235);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.poll-item:hover {
  background-color: rgb(249 250 251);
  border-color: rgb(209 213 219);
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1);
  transform: translateY(-2px);
}

.dark .poll-item {
  background-color: rgb(31 41 55);
  border-color: rgb(55 65 81);
}

.dark .poll-item:hover {
  background-color: rgb(55 65 81);
  border-color: rgb(75 85 99);
}

/* Poll Item Icon */
.poll-item-icon {
  width: 32px;
  height: 32px;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

/* Poll Item Content */
.poll-item-content {
  flex: 1;
  min-width: 0;
}

.poll-item-label {
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: #8B5CF6;
  margin-bottom: 0.25rem;
}

.dark .poll-item-label {
  color: #a78bfa;
}

.poll-item-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 0.375rem 0;
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.dark .poll-item-title {
  color: white;
}

.poll-item-meta {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: #6b7280;
}

.dark .poll-item-meta {
  color: #9ca3af;
}

/* Variants */
.poll-new {
  border-left: 3px solid #10b981;
}

.poll-expiring {
  border-left: 3px solid #f59e0b;
}

.poll-closed {
  border-left: 3px solid #6366f1;
}

/* Responsive */
@media (max-width: 640px) {
  .widget-header {
    margin-bottom: 1rem;
    padding-bottom: 0.75rem;
  }

  .header-icon {
    font-size: 1.25rem;
  }

  .widget-title {
    font-size: 1rem;
  }

  .poll-item {
    padding: 0.75rem;
  }

  .poll-item-icon {
    width: 36px;
    height: 36px;
  }
}
</style>
