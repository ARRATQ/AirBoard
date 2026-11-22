<template>
  <header class="greet">
    <p class="greet__eyebrow">{{ dateLabel }}</p>
    <h1 class="greet__title">{{ greetingWord }}, <span class="greet__name">{{ firstName }}</span></h1>
    <p v-if="pendingText" class="greet__sub">{{ pendingText }}</p>

    <nav v-if="activePoll || latestNewsGroup || nextEvent" class="greet__pills" :aria-label="$t('home.quickActions.title')">
      <router-link v-if="activePoll" :to="`/polls/${activePoll.id}`" class="greet__pill">
        <Icon icon="lucide:chart-column" class="greet__pill-icon" aria-hidden="true" />
        {{ $t('home.vote') }}
      </router-link>
      <router-link v-if="latestNewsGroup" to="/news" class="greet__pill">
        <Icon :icon="latestNewsGroup.type.icon || 'lucide:newspaper'" class="greet__pill-icon" aria-hidden="true" />
        {{ latestNewsGroup.type.name }}
      </router-link>
      <router-link v-if="nextEvent" :to="`/events/${nextEvent.slug}`" class="greet__pill">
        <span class="greet__pill-date">{{ shortDate(nextEvent.start_date) }}</span>
        <span class="greet__pill-text">{{ nextEvent.title }}</span>
      </router-link>
    </nav>
  </header>
</template>

<script setup>
// En-tête de l'accueil : date, salutation, ce qui attend l'utilisateur, raccourcis réels
// (sondage ouvert, dernier type d'actualité, prochain événement).
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { dateLocaleFor } from '@/utils/homeFormat'

defineProps({
  firstName: { type: String, default: '' },
  pendingText: { type: String, default: '' },
  activePoll: { type: Object, default: null },
  latestNewsGroup: { type: Object, default: null },
  nextEvent: { type: Object, default: null },
})

const { t, locale } = useI18n()
const dateLocale = computed(() => dateLocaleFor(locale.value))

const dateLabel = computed(() =>
  new Date().toLocaleDateString(dateLocale.value, { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }),
)

const greetingWord = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return t('home.hero.greeting.morning')
  if (h < 18) return t('home.hero.greeting.afternoon')
  return t('home.hero.greeting.evening')
})

const shortDate = (value) =>
  value ? new Date(value).toLocaleDateString(dateLocale.value, { day: '2-digit', month: '2-digit' }) : ''
</script>

<style scoped>
.greet { min-width: 0; }

.greet__eyebrow {
  margin: 0 0 6px;
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.greet__title {
  margin: 0;
  font-size: 28px;
  font-weight: 600;
  line-height: 1.2;
  letter-spacing: -0.015em;
  color: var(--text);
}

.greet__name { color: var(--acc-ink); }

.greet__sub {
  margin: 8px 0 0;
  max-width: 70ch;
  color: var(--muted);
}

.greet__pills {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-block-start: 14px;
}

.greet__pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  height: 30px;
  padding-inline: 10px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  transition: border-color 120ms ease;
}

.greet__pill:hover { border-color: var(--acc); }

.greet__pill:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.greet__pill-icon { flex: none; width: 15px; height: 15px; color: var(--muted); }

.greet__pill-date {
  flex: none;
  font-family: var(--f-num);
  font-size: 11.5px;
  color: var(--muted);
  direction: ltr;
}

.greet__pill-text {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

@media (max-width: 640px) {
  .greet__title { font-size: 22px; }
}
</style>
