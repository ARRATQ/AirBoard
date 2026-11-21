<template>
  <section class="home-card" aria-labelledby="home-events-title">
    <div class="home-card__head">
      <h2 id="home-events-title" class="home-card__title">{{ $t('home.upcomingEvents.title') }}</h2>
      <router-link to="/events" class="home-card__more">
        {{ $t('home.rail.agenda') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </div>
    <ul class="up__list">
      <li v-for="ev in events.slice(0, 4)" :key="ev.id">
        <router-link :to="`/events/${ev.slug}`" class="up__row" :style="{ '--c': ev.color || 'var(--muted)' }">
          <time class="up__date" :datetime="ev.start_date">
            <span class="up__month">{{ month(ev.start_date) }}</span>
            <span class="up__day">{{ day(ev.start_date) }}</span>
          </time>
          <span class="up__body">
            <span class="up__title">{{ ev.title }}</span>
            <span class="up__time">{{ ev.is_all_day ? $t('events.badge.allDayShort') : time(ev.start_date) }}</span>
          </span>
        </router-link>
      </li>
    </ul>
  </section>
</template>

<script setup>
// Quatre prochains événements ; l'agenda complet est sur /events.
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { dateLocaleFor } from '@/utils/homeFormat'

defineProps({
  events: { type: Array, required: true },
})

const { locale } = useI18n()
const loc = computed(() => dateLocaleFor(locale.value))

const month = (v) => new Date(v).toLocaleDateString(loc.value, { month: 'short' })
const day = (v) => new Date(v).getDate()
const time = (v) => new Date(v).toLocaleTimeString(loc.value, { hour: '2-digit', minute: '2-digit' })
</script>

<style scoped>
.up__list { margin: 0; padding: 0; list-style: none; }

.up__row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 48px;
  padding: 6px 4px;
  border-block-end: 1px solid var(--line);
  color: var(--text);
}

.up__list li:last-child .up__row { border-block-end: 0; }
.up__row:hover { background: var(--sunken); }

.up__row:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: -2px;
  border-radius: 4px;
}

.up__date {
  flex: none;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 40px;
  padding-inline-end: 8px;
  border-inline-end: 3px solid var(--c);
  font-family: var(--f-num);
  line-height: 1.1;
}

.up__month {
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.up__day { font-size: 17px; font-weight: 600; }

.up__body { min-width: 0; display: flex; flex-direction: column; }

.up__title {
  font-size: 13px;
  font-weight: 500;
  line-height: 1.3;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.up__time { font-family: var(--f-num); font-size: 11.5px; color: var(--muted); }
</style>
