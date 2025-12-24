<template>
  <div class="poll-results">
    <!-- Chargement -->
    <div v-if="loading" class="results-loading">
      <Icon icon="mdi:loading" class="spin loading-icon" />
      <p>{{ t('polls.results.loading') }}</p>
    </div>

    <!-- Résultats -->
    <div v-else-if="results" class="results-content">
      <!-- Options avec barres de progression -->
      <div class="results-options">
        <div
          v-for="optionResult in results.options"
          :key="optionResult.poll_option.id"
          class="result-option"
          :class="{ 
            'user-choice': isUserChoice(optionResult.poll_option.id),
            'leading-option': isLeadingOption(optionResult)
          }"
        >
          <div class="option-header">
            <span class="option-text">{{ optionResult.poll_option.text }}</span>
            <div class="option-stats">
              <span class="vote-count">{{ optionResult.vote_count }}</span>
              <span class="vote-percentage">({{ optionResult.percentage.toFixed(1) }}%)</span>
            </div>
          </div>
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: `${optionResult.percentage}%` }"
            ></div>
          </div>
        </div>
      </div>

      <!-- Détails des votants (si non anonyme et autorisé) -->
      <div v-if="!compact && results.voter_details && results.voter_details.length > 0" class="voter-details">
        <h4 class="voter-details-title">
          <Icon icon="mdi:account-details" />
          {{ t('polls.results.voterDetails') }}
        </h4>
        <div class="voter-list">
          <div
            v-for="voter in results.voter_details"
            :key="`${voter.user.id}-${voter.poll_option_id}`"
            class="voter-item"
          >
            <div class="voter-info">
              <Icon icon="mdi:account-circle" />
              <span class="voter-name">
                {{ voter.user.first_name }} {{ voter.user.last_name }}
              </span>
            </div>
            <span class="voter-choice">
              {{ getOptionText(voter.poll_option_id) }}
            </span>
            <span class="voter-date">
              {{ formatDate(voter.voted_at) }}
            </span>
          </div>
        </div>
      </div>

      <!-- Bouton pour voir les détails complets -->
      <div v-if="compact && !hideViewDetails" class="view-details">
        <button @click="$emit('view-details')" class="btn btn-secondary">
          <Icon icon="mdi:chart-bar" />
          {{ t('polls.results.viewFullDetails') }}
        </button>
      </div>
    </div>

    <!-- Erreur -->
    <div v-else-if="error" class="results-error">
      <Icon icon="mdi:alert-circle" class="error-icon" />
      <p>{{ error }}</p>
      <button @click="loadResults" class="btn btn-primary btn-sm">
        <Icon icon="mdi:refresh" />
        {{ t('polls.results.retry') }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { pollsService } from '@/services/api'
import { format } from 'date-fns'
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
  hideViewDetails: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['view-details'])

const { t } = useI18n()

const loading = ref(false)
const results = ref(null)
const error = ref(null)

// Charger les résultats
const loadResults = async () => {
  try {
    loading.value = true
    error.value = null
    results.value = await pollsService.getResults(props.poll.id)
  } catch (err) {
    console.error(t('polls.results.loadError'), err)
    error.value = err.response?.data?.error || t('polls.results.loadError')
  } finally {
    loading.value = false
  }
}

// Vérifier si une option a été choisie par l'utilisateur
const isUserChoice = (optionId) => {
  if (!results.value || !results.value.user_votes) return false
  return results.value.user_votes.includes(optionId)
}

// Vérifier si une option a le plus de votes
const isLeadingOption = (optionResult) => {
  if (!results.value || !results.value.options) return false
  const maxVotes = Math.max(...results.value.options.map(o => o.vote_count))
  return optionResult.vote_count === maxVotes && optionResult.vote_count > 0
}

// Obtenir le texte d'une option par son ID
const getOptionText = (optionId) => {
  const option = results.value?.options.find(o => o.poll_option.id === optionId)
  return option?.poll_option.text || t('polls.results.unknownOption')
}

// Formater la date
const formatDate = (date) => {
  if (!date) return ''
  return format(new Date(date), "dd MMM yyyy 'à' HH:mm", { locale: fr })
}

// Charger les résultats au montage et quand le sondage change
watch(() => props.poll.id, () => {
  loadResults()
}, { immediate: true })
</script>

<style scoped>
.poll-results {
  width: 100%;
  position: relative;
}

.results-content {
  position: relative;
}

.results-loading,
.results-error {
  text-align: center;
  padding: 2rem;
  background: #f9fafb;
  border-radius: 8px;
}

:is(.dark .results-loading),
:is(.dark .results-error) {
  background: #374151;
}

.loading-icon {
  font-size: 3rem;
  color: #8B5CF6;
  margin-bottom: 1rem;
}

.error-icon {
  font-size: 3rem;
  color: #ef4444;
  margin-bottom: 1rem;
}

.results-options {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.result-option {
  padding: 0.5rem;
  background: #f9fafb;
  border-radius: 8px;
  transition: all 0.3s;
}

:is(.dark .result-option) {
  background: #374151;
}

.result-option.user-choice {
  background: #ede9fe;
  border: 2px solid #8B5CF6;
}

.result-option.leading-option {
  background: rgba(59, 130, 246, 0.08);
  border: none;
  border-left: 4px solid #3B82F6;
  border-radius: 4px;
}

:is(.dark .result-option.leading-option) {
  background: rgba(59, 130, 246, 0.15);
  border: none;
  border-left: 4px solid #3B82F6;
}

/* Text colors for leading option to match event style */
.leading-option .option-text,
.leading-option .vote-count {
  color: #1e40af;
}

:is(.dark .leading-option .option-text),
:is(.dark .leading-option .vote-count) {
  color: #bfdbfe;
}

.leading-option .vote-percentage {
  color: #1e40af;
  opacity: 0.8;
}

:is(.dark .leading-option .vote-percentage) {
  color: #bfdbfe;
  opacity: 0.8;
}

/* Force event style (Blue) even if user selected it */
.result-option.user-choice.leading-option {
  background: rgba(59, 130, 246, 0.08);
  border: none;
  border-left: 4px solid #3B82F6;
  border-radius: 4px;
}

:is(.dark .result-option.user-choice.leading-option) {
  background: rgba(59, 130, 246, 0.15);
  border: none;
  border-left: 4px solid #3B82F6;
}

/* Force blue progress bar for leading option even if user selected it */
.user-choice.leading-option .progress-fill {
  background: linear-gradient(90deg, #3B82F6, #60a5fa);
  box-shadow: 0 0 4px rgba(59, 130, 246, 0.5);
}

:is(.dark .user-choice.leading-option .progress-fill) {
  background: linear-gradient(90deg, #3B82F6, #2563eb);
  box-shadow: 0 0 4px rgba(59, 130, 246, 0.6);
}

.option-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  gap: 1rem;
}

.option-text {
  font-weight: 500;
  font-size: 0.875rem;
  color: #1f2937;
  flex: 1;
}

:is(.dark .option-text) {
  color: #ffffff;
}

.option-stats {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: #6b7280;
  font-weight: 600;
  white-space: nowrap;
}

:is(.dark .option-stats) {
  color: #a1a1aa;
}

.vote-count {
  color: #1f2937;
}

:is(.dark .vote-count) {
  color: #ffffff;
}

.vote-percentage {
  color: #6b7280;
}

:is(.dark .vote-percentage) {
  color: #a1a1aa;
}

.progress-bar {
  height: 8px;
  background: #e5e7eb;
  border-radius: 9999px;
  overflow: hidden;
}

:is(.dark .progress-bar) {
  background: #4b5563;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #8B5CF6, #a78bfa);
  border-radius: 9999px;
  transition: width 0.6s ease-out;
}

.leading-option .progress-fill {
  background: linear-gradient(90deg, #3B82F6, #60a5fa);
  box-shadow: 0 0 4px rgba(59, 130, 246, 0.5);
}

:is(.dark .leading-option .progress-fill) {
  background: linear-gradient(90deg, #3B82F6, #2563eb);
  box-shadow: 0 0 4px rgba(59, 130, 246, 0.6);
}

.user-choice .progress-fill {
  background: linear-gradient(90deg, rgba(16, 185, 129, 0.8), rgba(52, 211, 153, 0.8));
  box-shadow: 0 0 4px rgba(16, 185, 129, 0.5);
}

.voter-details {
  margin-top: 2rem;
  padding: 1.5rem;
  background: #f9fafb;
  border-radius: 8px;
}

:is(.dark .voter-details) {
  background: #374151;
}

.voter-details-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
  font-size: 1rem;
  font-weight: 600;
  color: #1f2937;
}

:is(.dark .voter-details-title) {
  color: #ffffff;
}

.voter-details-title svg {
  font-size: 1.25rem;
  color: #8B5CF6;
}

.voter-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.voter-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: white;
  border-radius: 6px;
}

:is(.dark .voter-item) {
  background: #4b5563;
}

.voter-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
}

.voter-info svg {
  font-size: 1.5rem;
  color: #6b7280;
}

.voter-name {
  font-weight: 500;
  font-size: 0.875rem;
  color: #1f2937;
}

:is(.dark .voter-name) {
  color: #ffffff;
}

.voter-choice {
  color: #6b7280;
  font-size: 0.75rem;
  padding: 0.25rem 0.75rem;
  background: #e5e7eb;
  border-radius: 9999px;
}

:is(.dark .voter-choice) {
  background: #6b7280;
  color: #d1d5db;
}

.voter-date {
  color: #9ca3af;
  font-size: 0.75rem;
  white-space: nowrap;
}

.view-details {
  text-align: center;
  margin-top: 1rem;
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

.btn-primary:hover {
  background: #7c3aed;
}

.btn-secondary {
  background: #e5e7eb;
  color: #374151;
}

.btn-secondary:hover {
  background: #d1d5db;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.875rem;
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
