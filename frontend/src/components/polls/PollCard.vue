<template>
  <div class="poll-card" :key="`poll-${poll.id}`" :class="{ 'is-expanded': isExpanded }" :data-poll-id="poll.id">
    <!-- En-tête du sondage -->
    <div class="poll-header" @click="toggleExpanded" :class="{ 'clickable': true }">
      <div class="poll-header-content">
        <div
          class="h-7 w-7 rounded-lg flex items-center justify-center shadow-sm flex-shrink-0"
          :style="{ backgroundColor: poll.is_active ? '#8B5CF6' : '#10b981' }"
        >
          <Icon :icon="poll.is_active ? 'mdi:poll' : 'mdi:poll-check'" class="h-4 w-4 text-white" />
        </div>
        <div class="poll-info">
          <div class="flex items-center gap-2 flex-1 min-w-0">
            <h3 class="poll-title">{{ poll.title }}</h3>
            <!-- Indicateur de vote utilisateur -->
            <div v-if="hasVoted" class="voted-indicator-header">
              <Icon icon="mdi:check-circle" class="voted-icon" />
            </div>
            <div class="poll-badges">
              <span v-if="!poll.is_active" class="badge badge-closed">
                <Icon icon="mdi:lock" class="h-3 w-3" />
                <span class="badge-text">{{ t('polls.status.closed') }}</span>
              </span>
              <span v-else-if="isExpiringSoon" class="badge badge-warning">
                <Icon icon="mdi:clock-alert" class="h-3 w-3" />
                <span class="badge-text">{{ t('polls.status.expiringSoon') }}</span>
              </span>
              <span v-if="poll.is_anonymous" class="badge badge-anonymous">
                <Icon icon="mdi:incognito" class="h-3 w-3" />
                <span class="badge-text">{{ t('polls.status.anonymous') }}</span>
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Bouton de toggle -->
      <button 
        @click.stop="toggleExpanded" 
        class="toggle-btn"
        :class="{ 'expanded': isExpanded }"
      >
        <Icon icon="mdi:chevron-down" class="chevron-icon" />
      </button>
    </div>

    <!-- Contenu collapsible -->
    <Transition name="slide-fade">
      <div v-if="isExpanded" class="collapsible-content">
        <!-- Informations sur le sondage (publicateur et date) -->
        <div class="poll-meta">
          <div class="meta-item">
            <Icon icon="mdi:account" class="h-4 w-4" />
            <span class="text-sm">{{ poll.author?.first_name }} {{ poll.author?.last_name }}</span>
          </div>
          <div class="meta-item">
            <Icon icon="mdi:calendar" class="h-4 w-4" />
            <span class="text-sm">{{ formatDate(poll.created_at) }}</span>
          </div>
          <div v-if="poll.end_date" class="meta-item">
            <Icon icon="mdi:calendar-end" class="h-4 w-4" />
            <span class="text-sm">{{ t('polls.status.until') }} {{ formatDate(poll.end_date) }}</span>
          </div>
        </div>

        <!-- Description du sondage -->
        <div v-if="poll.description" class="poll-description">
          <p class="text-sm">{{ poll.description }}</p>
        </div>

        <!-- Zone de vote ou résultats -->
        <div class="poll-content">
          <!-- Formulaire de vote -->
          <div v-if="canVote" class="poll-voting">
            <div v-if="poll.poll_type === 'single'" class="poll-options">
              <label
                v-for="option in poll.options"
                :key="option.id"
                class="poll-option"
              >
                <input
                  type="radio"
                  :name="`poll-${poll.id}`"
                  :value="option.id"
                  v-model="selectedOptions[0]"
                  :disabled="isVoting"
                />
                <span class="option-text">{{ option.text }}</span>
              </label>
            </div>

            <div v-else class="poll-options">
              <label
                v-for="option in poll.options"
                :key="option.id"
                class="poll-option"
              >
                <input
                  type="checkbox"
                  :value="option.id"
                  v-model="selectedOptions"
                  :disabled="isVoting"
                />
                <span class="option-text">{{ option.text }}</span>
              </label>
            </div>

            <button
              @click="submitVote"
              :disabled="!canSubmitVote || isVoting"
              class="btn btn-primary vote-btn"
            >
              <Icon v-if="isVoting" icon="mdi:loading" class="spin" />
              <Icon v-else icon="mdi:vote" />
              {{ isVoting ? t('polls.voting.sending') : t('polls.voting.vote') }}
            </button>
          </div>

          <!-- Résultats (affichés séparément si show_results === 'always' ou après vote) -->
          <div v-if="showResults" class="poll-results-section">
            <PollResults :poll="poll" :compact="compact" @view-details="$emit('view-details', poll.id)" />
          </div>

          <!-- Message d'attente (seulement si pas de formulaire de vote et pas de résultats) -->
          <div v-else-if="!canVote" class="poll-waiting">
            <Icon icon="mdi:lock-clock" class="waiting-icon" />
            <p class="waiting-message">
              {{ getWaitingMessage() }}
            </p>
            <button
              v-if="hasVoted && poll.show_results === 'after'"
              @click="showResults = true"
              class="btn btn-secondary"
            >
              <Icon icon="mdi:chart-bar" />
              {{ t('polls.voting.viewResults') }}
            </button>
          </div>

          <button
            v-if="canRemoveVote"
            @click="removeVote"
            :disabled="isRemovingVote"
            class="btn btn-danger-outline unvote-btn"
          >
            <Icon v-if="isRemovingVote" icon="mdi:loading" class="spin" />
            <Icon v-else icon="mdi:undo" />
            {{ isRemovingVote ? t('polls.voting.removingVote') : t('polls.voting.removeVote') }}
          </button>
        </div>

        <!-- Actions -->
        <div v-if="canManage" class="poll-actions">
          <button @click="$emit('edit', poll)" class="btn btn-ghost btn-sm">
            <Icon icon="mdi:pencil" />
            {{ t('polls.actions.edit') }}
          </button>
          <button
            v-if="poll.is_active && canClose"
            @click="$emit('close', poll)"
            class="btn btn-ghost btn-sm"
          >
            <Icon icon="mdi:lock" />
            {{ t('polls.actions.close') }}
          </button>
          <button @click="$emit('delete', poll)" class="btn btn-ghost btn-sm text-danger">
            <Icon icon="mdi:delete" />
            {{ t('polls.actions.delete') }}
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { pollsService } from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import PollResults from './PollResults.vue'
import { format, isAfter, isBefore, differenceInDays } from 'date-fns'
import { fr } from 'date-fns/locale'

const props = defineProps({
  poll: {
    type: Object,
    required: true
  },
  compact: {
    type: Boolean,
    default: false
  },
  canManage: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['vote-success', 'edit', 'delete', 'close', 'view-details'])

const authStore = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const selectedOptions = ref([])
const isVoting = ref(false)
const isRemovingVote = ref(false)
const hasVoted = ref(false)
const showResults = ref(false)
const userVotes = ref([]) // Stocker les votes de l'utilisateur

// État d'expansion isolé pour chaque bloc
const isExpanded = ref(false)

// Fonction pour toggle l'expansion
const toggleExpanded = () => {
  isExpanded.value = !isExpanded.value
  
  // Add a small delay to ensure smooth transition
  if (isExpanded.value) {
    // Force a small reflow to ensure the card adapts properly
    requestAnimationFrame(() => {
      const card = document.querySelector(`[data-poll-id="${props.poll.id}"]`)
      if (card) {
        card.classList.add('is-expanded')
      }
    })
  } else {
    const card = document.querySelector(`[data-poll-id="${props.poll.id}"]`)
    if (card) {
      card.classList.remove('is-expanded')
    }
  }
}

// Réinitialiser l'état d'expansion quand le sondage change
watch(() => props.poll.id, () => {
  isExpanded.value = false
})

// Vérifier si l'utilisateur peut voter
const canVote = computed(() => {
  if (!props.poll.is_active) return false
  if (hasVoted.value) return false

  const now = new Date()
  if (props.poll.start_date && isBefore(now, new Date(props.poll.start_date))) return false
  if (props.poll.end_date && isAfter(now, new Date(props.poll.end_date))) return false

  return true
})

// Vérifier si le sondage expire bientôt (dans les 3 prochains jours)
const isExpiringSoon = computed(() => {
  if (!props.poll.end_date || !props.poll.is_active) return false
  const daysLeft = differenceInDays(new Date(props.poll.end_date), new Date())
  return daysLeft >= 0 && daysLeft <= 3
})

// Vérifier si l'utilisateur peut soumettre son vote
const canSubmitVote = computed(() => {
  return selectedOptions.value.length > 0
})

const canRemoveVote = computed(() => {
  if (!hasVoted.value || !props.poll?.is_active || isVoting.value || isRemovingVote.value) return false
  if (props.poll.end_date && isAfter(new Date(), new Date(props.poll.end_date))) return false
  return true
})

const selectedOtherOption = computed(() => {
  if (!props.poll?.options?.length || !selectedOptions.value.length) {
    return null
  }

  return props.poll.options.find(
    (option) => option.is_other && selectedOptions.value.includes(option.id)
  ) || null
})

// Vérifier si l'utilisateur peut fermer le sondage
const canClose = computed(() => {
  const user = authStore.user
  return user.role === 'admin' || props.poll.author_id === user.id
})

// Charger le statut de vote de l'utilisateur
const checkUserVoteStatus = async () => {
  if (!props.poll?.id) return

  try {
    const response = await pollsService.getResults(props.poll.id)
    hasVoted.value = response.user_voted || false
    userVotes.value = response.user_votes || []

    // Mettre à jour les options sélectionnées si l'utilisateur a déjà voté
    if (hasVoted.value && userVotes.value.length > 0) {
      selectedOptions.value = [...userVotes.value]
    }

    // Déterminer si on doit afficher les résultats
    if (props.poll.show_results === 'always') {
      showResults.value = true
    } else if (!props.poll.is_active && props.poll.show_results === 'closed') {
      showResults.value = true
    } else if (hasVoted.value && props.poll.show_results === 'after') {
      showResults.value = true
    }
  } catch (error) {
    // Si erreur (ex: résultats pas encore disponibles), ce n'est pas grave
    console.debug('Vote status check:', error.response?.data?.error || error.message)
  }
}

// Surveiller les changements du sondage
watch(() => props.poll, (newPoll) => {
  if (newPoll) {
    checkUserVoteStatus()
  }
}, { immediate: true })

// Soumettre le vote
const submitVote = async () => {
  if (!canSubmitVote.value || isVoting.value) return

  try {
    isVoting.value = true
    await pollsService.vote(props.poll.id, selectedOptions.value)

    // Marquer comme ayant voté
    hasVoted.value = true
    userVotes.value = [...selectedOptions.value]

    // Afficher les résultats si la configuration le permet
    if (props.poll.show_results === 'after' || props.poll.show_results === 'always') {
      showResults.value = true
    }

    // Émettre l'événement de succès pour recharger les données
    emit('vote-success', props.poll.id)

    // Notification de succès
    if (window.$toast) {
      window.$toast.success(t('polls.voting.success'))
    }

    if (selectedOtherOption.value) {
      router.push({
        path: '/suggestions',
        query: {
          from_poll: '1',
          poll_id: String(props.poll.id),
          poll_option_id: String(selectedOtherOption.value.id),
          poll_title: props.poll.title,
          option_text: selectedOtherOption.value.text
        }
      })

      if (window.$toast) {
        window.$toast.info(t('polls.voting.redirectToSuggestion'))
      }
    }
  } catch (error) {
    console.error(t('polls.voting.error'), error)
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.voting.error'))
    }
  } finally {
    isVoting.value = false
  }
}

const removeVote = async () => {
  if (!canRemoveVote.value || isRemovingVote.value) return

  try {
    isRemovingVote.value = true
    await pollsService.removeVote(props.poll.id)

    hasVoted.value = false
    userVotes.value = []
    selectedOptions.value = []
    showResults.value = props.poll.show_results === 'always'

    if (window.$toast) {
      window.$toast.success(t('polls.voting.removeVoteSuccess'))
    }
  } catch (error) {
    if (window.$toast) {
      window.$toast.error(error.response?.data?.error || t('polls.voting.removeVoteError'))
    }
  } finally {
    isRemovingVote.value = false
  }
}

// Formater la date
const formatDate = (date) => {
  if (!date) return ''
  return format(new Date(date), 'dd MMM yyyy', { locale: fr })
}

// Message d'attente
const getWaitingMessage = () => {
  if (!props.poll.is_active) {
    if (props.poll.show_results === 'closed') {
      return t('polls.voting.pollClosedResults')
    }
    return t('polls.voting.pollClosed')
  }

  if (hasVoted.value) {
    if (props.poll.show_results === 'after') {
      return t('polls.voting.thanksForVote')
    } else if (props.poll.show_results === 'closed') {
      return t('polls.voting.thanksResultsLater')
    }
  }

  return t('polls.voting.voteToViewResults')
}
</script>

<style scoped>
.poll-card {
  background-color: rgb(255 255 255);
  border: 2px solid rgb(229 231 235);
  border-radius: 0.75rem;
  box-shadow: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  position: relative;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
  isolation: isolate;
  /* Prevent content from bleeding outside */
  contain: layout;
  /* Ensure proper stacking context */
  z-index: 1;
}

/* Ensure proper isolation between poll cards */
.poll-card .collapsible-content {
  flex-shrink: 0;
}

/* Simplified layout management */
.poll-card {
  /* Enable adaptive height based on content */
  height: auto;
  min-height: 4rem;
  /* Prevent layout conflicts */
  contain: layout style;
  /* Ensure proper box model */
  box-sizing: border-box;
}

/* Simplified state management */
.poll-card.is-expanded {
  min-height: 6rem;
}

/* Smooth height transitions */
.poll-card {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Controlled content overflow */
.poll-card .collapsible-content {
  will-change: height;
  /* Prevent content overflow while allowing smooth transitions */
  overflow: hidden;
  /* Ensure content stays within bounds */
  max-width: 100%;
}

/* Ajuster la hauteur automatiquement selon le contenu */
.poll-card.has-content {
  min-height: auto;
}

.poll-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
}

:is(.dark .poll-card) {
  background-color: rgb(31 41 55);
  border-color: rgb(55 65 81);
}

:is(.dark .poll-card:hover) {
  box-shadow: 0 10px 15px -3px rgb(0 0 0 / 0.3), 0 4px 6px -4px rgb(0 0 0 / 0.3);
}

.poll-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid rgb(229 231 235);
}

.poll-header:hover {
  background-color: rgb(249 250 251);
}

:is(.dark .poll-header) {
  border-bottom-color: rgb(55 65 81);
}

:is(.dark .poll-header:hover) {
  background-color: rgb(55 65 81 / 0.5);
}

.poll-header-content {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  flex: 1;
  min-width: 0;
}

.poll-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #111827;
  margin: 0;
  line-height: 1.4;
  word-wrap: break-word;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Indicateur de vote dans le titre */
.voted-indicator-header {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #10b981;
  color: white;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  flex-shrink: 0;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.voted-indicator-header .voted-icon {
  font-size: 12px;
}

:is(.dark .poll-title) {
  color: #ffffff;
}

.poll-badges {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  white-space: nowrap;
  height: auto;
}

.badge svg {
  width: 0.875rem;
  height: 0.875rem;
  stroke-width: 2;
  flex-shrink: 0;
}

.badge-text {
  font-size: inherit;
  line-height: 1;
}

.badge-closed {
  background: rgb(239 68 68);
  color: white;
}

:is(.dark .badge-closed) {
  background: rgb(220 38 38);
  color: white;
}

.badge-warning {
  background: rgb(245 158 11);
  color: white;
}

:is(.dark .badge-warning) {
  background: rgb(217 119 6);
  color: white;
}

.badge-anonymous {
  background: rgb(224 231 255);
  color: rgb(99 102 241);
}

:is(.dark .badge-anonymous) {
  background: rgb(49 46 129);
  color: rgb(165 180 252);
}

/* Bouton de toggle */
.toggle-btn {
  background: none;
  border: none;
  padding: 0.5rem;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  flex-shrink: 0;
}

.toggle-btn:hover {
  background: rgb(243 244 246);
}

:is(.dark .toggle-btn:hover) {
  background: rgb(75 85 99);
}

.chevron-icon {
  font-size: 1rem;
  color: rgb(107 114 128);
  transition: transform 0.3s ease;
}

.toggle-btn.expanded .chevron-icon {
  transform: rotate(180deg);
}

:is(.dark .chevron-icon) {
  color: rgb(156 163 175);
}

/* Contenu collapsible avec expansion contrôlée */
.collapsible-content {
  padding: 1rem;
  /* Prevent content bleeding */
  overflow: hidden;
  width: 100%;
  box-sizing: border-box;
  background-color: rgb(255 255 255);
  position: relative;
  z-index: 1;
  /* Smooth expansion with proper bounds */
  max-height: 2000px;
  /* Ensure content doesn't get cut off at bottom */
  margin-bottom: 1rem;
  /* Prevent layout shifts */
  contain: layout;
}

:is(.dark .collapsible-content) {
  background-color: rgb(31 41 55);
}

/* Informations sur le sondage */
/* Description du sondage */
.poll-description {
  margin-bottom: 1rem;
  padding: 0.75rem;
  background-color: rgb(249 250 251);
  border-radius: 0.5rem;
  border: 1px solid rgb(229 231 235);
}

:is(.dark .poll-description) {
  background-color: rgb(55 65 81);
  border-color: rgb(75 85 99);
}

/* Transitions Vue simplifiées pour l'ouverture/fermeture */
.slide-fade-enter-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  /* Prevent layout shift during transition */
  will-change: max-height, opacity;
}

.slide-fade-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  will-change: max-height, opacity;
}

.slide-fade-enter-from {
  opacity: 0;
  max-height: 0;
  transform: translateY(-10px);
}

.slide-fade-leave-to {
  opacity: 0;
  max-height: 0;
  transform: translateY(-10px);
}

.slide-fade-enter-to,
.slide-fade-leave-from {
  opacity: 1;
  max-height: 1500px;
  transform: translateY(0);
}

/* Informations sur le sondage */
.poll-meta {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgb(229 231 235);
}

:is(.dark .poll-meta) {
  border-bottom-color: rgb(55 65 81);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: rgb(107 114 128);
  font-size: 0.75rem;
  white-space: nowrap;
}

:is(.dark .meta-item) {
  color: rgb(156 163 175);
}

.meta-item svg {
  font-size: 1rem;
}

.poll-content {
  margin-bottom: 1rem;
}

.poll-voting {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.poll-options {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.poll-option {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  background: white;
}

:is(.dark .poll-option) {
  border-color: #4b5563;
  background: #374151;
}

.poll-option:hover {
  border-color: #8B5CF6;
  background: #f9fafb;
}

:is(.dark .poll-option:hover) {
  background: #4b5563;
}

.poll-option input[type="radio"],
.poll-option input[type="checkbox"] {
  width: 1.25rem;
  height: 1.25rem;
  cursor: pointer;
}

.option-text {
  flex: 1;
  font-size: 1rem;
  color: #374151;
}

:is(.dark .option-text) {
  color: #f3f4f6;
}

.vote-btn {
  align-self: flex-start;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.poll-results-section {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 2px solid #e5e7eb;
}

:is(.dark .poll-results-section) {
  border-top-color: #4b5563;
}

.poll-waiting {
  text-align: center;
  padding: 2rem;
  background: #f9fafb;
  border-radius: 8px;
}

:is(.dark .poll-waiting) {
  background: #374151;
}

.waiting-icon {
  font-size: 3rem;
  color: #9ca3af;
  margin-bottom: 1rem;
}

.waiting-message {
  color: #6b7280;
  font-size: 1rem;
  margin-bottom: 1rem;
}

:is(.dark .waiting-message) {
  color: #a1a1aa;
}

.poll-actions {
  display: flex;
  gap: 0.5rem;
  padding-top: 1rem;
  border-top: 1px solid #e5e7eb;
}

:is(.dark .poll-actions) {
  border-top-color: #4b5563;
}

.btn {
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-primary {
  background: #8B5CF6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #7c3aed;
}

.btn-primary:disabled {
  background: #d1d5db;
  cursor: not-allowed;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover {
  background: #d1d5db;
}

.btn-danger-outline {
  background: #fff1f2;
  color: #be123c;
  border: 1px solid #fecdd3;
}

.btn-danger-outline:hover:not(:disabled) {
  background: #ffe4e6;
  border-color: #fda4af;
}

.btn-danger-outline:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-ghost {
  background: transparent;
  color: #6b7280;
}

.btn-ghost:hover {
  background: #f3f4f6;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
}

.unvote-btn {
  margin: 0.5rem auto 0;
}

.text-danger {
  color: #ef4444;
}

.text-danger:hover {
  color: #dc2626;
  background: #fee2e2;
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
