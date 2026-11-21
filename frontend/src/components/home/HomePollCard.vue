<template>
  <section class="home-card" aria-labelledby="home-poll-title">
    <div class="home-card__head">
      <h2 id="home-poll-title" class="home-card__title">{{ $t('home.polls.title') }}</h2>
      <StatusBadge :tone="status.tone">{{ status.label }}</StatusBadge>
    </div>

    <p class="poll__question">{{ poll.title }}</p>

    <ul class="poll__options">
      <li v-for="opt in (poll.options || []).slice(0, 4)" :key="opt.id" class="poll__option">
        <span class="poll__label">
          <span class="poll__text">{{ opt.text }}</span>
          <span class="poll__pct">{{ pollPercent(opt, poll) }} %</span>
        </span>
        <span class="poll__track" aria-hidden="true">
          <span
            class="poll__fill"
            :class="{ 'poll__fill--lead': isLeadingOption(opt, poll) }"
            :style="{ inlineSize: pollPercent(opt, poll) + '%' }"
          />
        </span>
      </li>
    </ul>

    <BaseButton tag="router-link" :to="`/polls/${poll.id}`" variant="primary" size="sm" block>
      {{ $t(isEnded ? 'home.polls.seeResults' : 'home.polls.vote') }}
    </BaseButton>
  </section>
</template>

<script setup>
// Sondage ouvert sur l'accueil : question, tendance des réponses et accès au vote.
// Seule la fin proche (≤ 2 jours) est colorée, comme toute exception en A+.
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import { daysUntil, isLeadingOption, pollPercent } from '@/utils/homeFormat'

const props = defineProps({
  poll: { type: Object, required: true },
})

const ENDING_SOON_DAYS = 2
const { t } = useI18n()

// Sondage arrivé à échéance : on propose les résultats plutôt qu'un vote.
const isEnded = computed(() => {
  const days = daysUntil(props.poll.end_date)
  return days !== null && days <= 0
})

const status = computed(() => {
  const days = daysUntil(props.poll.end_date)
  if (days === null) return { tone: 'neutral', label: t('home.polls.open') }
  if (isEnded.value) return { tone: 'neutral', label: t('home.polls.ended') }
  return { tone: days <= ENDING_SOON_DAYS ? 'warn' : 'neutral', label: t('home.polls.endsIn', days) }
})
</script>

<style scoped>
.poll__question {
  margin: 0 0 10px;
  font-weight: 600;
  line-height: 1.35;
  color: var(--text);
}

.poll__options {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin: 0 0 12px;
  padding: 0;
  list-style: none;
}

.poll__label {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  font-size: 12.5px;
  color: var(--text);
}

.poll__text { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }

.poll__pct {
  flex: none;
  font-family: var(--f-num);
  font-size: 11.5px;
  color: var(--muted);
}

.poll__track {
  display: block;
  height: 5px;
  margin-block-start: 4px;
  border-radius: 3px;
  background: var(--sunken);
  box-shadow: inset 0 0 0 1px var(--line);
  overflow: hidden;
}

.poll__fill {
  display: block;
  height: 100%;
  border-radius: 3px;
  background: color-mix(in oklab, var(--muted) 45%, transparent);
}

.poll__fill--lead { background: var(--fill); }
</style>
