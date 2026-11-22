<template>
  <section class="cal" :aria-busy="busy">
    <header class="cal__head">
      <h2 class="cal__period">{{ periodLabel }}</h2>
      <div class="cal__tools">
        <SegmentedControl v-model="currentViewMode" :options="viewModes" :label="$t('calendar.title')" />
        <div class="cal__nav">
          <button type="button" class="cal__navbtn" :aria-label="$t('calendar.previous')" @click="navigate(-1)">
            <Icon icon="lucide:chevron-left" class="cal__flip" aria-hidden="true" />
          </button>
          <button type="button" class="cal__navbtn cal__navbtn--text" @click="goToToday">{{ $t('calendar.today') }}</button>
          <button type="button" class="cal__navbtn" :aria-label="$t('calendar.next')" @click="navigate(1)">
            <Icon icon="lucide:chevron-right" class="cal__flip" aria-hidden="true" />
          </button>
        </div>
      </div>
    </header>

    <div v-if="busy" class="cal__loading" role="status">
      <Icon icon="lucide:loader-circle" class="cal__spin" aria-hidden="true" />
      <span class="sr-only">{{ $t('calendar.loading') }}</span>
    </div>

    <!-- Mois -->
    <div v-if="currentViewMode === 'month'" class="cal__month">
      <div class="cal__weekdays" aria-hidden="true">
        <span v-for="(name, i) in weekdayNames" :key="i" class="cal__weekday">
          <span class="cal__weekday-long">{{ name.short }}</span>
          <span class="cal__weekday-narrow">{{ name.narrow }}</span>
        </span>
      </div>
      <ol class="cal__grid">
        <li v-for="day in monthDays" :key="day.key" class="cal__cell" :class="dayClasses(day)">
          <button type="button" class="cal__date" :aria-label="dayLabel(day)" @click="openDay(day.date)">
            {{ day.date.getDate() }}
          </button>
          <p v-if="day.holiday" class="cal__holiday">{{ day.holiday }}</p>
          <ul class="cal__events">
            <li v-for="event in day.events.slice(0, MONTH_LIMIT)" :key="event.id">
              <button type="button" class="cal__chip" :class="chipClasses(event)" :style="chipStyle(event)" @click="selectedEvent = event">
                <Icon v-if="statusIcon(event)" :icon="statusIcon(event)" class="cal__chip-status" aria-hidden="true" />
                <span class="cal__chip-title">{{ event.title }}</span>
                <span v-if="statusText(event)" class="sr-only">({{ statusText(event) }})</span>
              </button>
            </li>
          </ul>
          <button v-if="day.events.length > MONTH_LIMIT" type="button" class="cal__more" @click="openDay(day.date)">
            {{ $t('calendar.andMore', { count: day.events.length - MONTH_LIMIT }) }}
          </button>
        </li>
      </ol>
    </div>

    <!-- Semaine -->
    <ol v-else-if="currentViewMode === 'week'" class="cal__week">
      <li v-for="day in weekDays" :key="day.key" class="cal__wday" :class="dayClasses(day)">
        <button type="button" class="cal__wday-head" @click="openDay(day.date)">
          <span class="cal__wday-name">{{ formatDay(day.date, { weekday: 'short' }) }}</span>
          <span class="cal__wday-num">{{ day.date.getDate() }}</span>
        </button>
        <p v-if="day.holiday" class="cal__holiday">{{ day.holiday }}</p>
        <ul class="cal__events">
          <li v-for="event in day.events" :key="event.id">
            <button type="button" class="cal__chip cal__chip--tall" :class="chipClasses(event)" :style="chipStyle(event)" @click="selectedEvent = event">
              <span class="cal__chip-time">{{ timeLabel(event) }}</span>
              <span class="cal__chip-title">
                <Icon v-if="statusIcon(event)" :icon="statusIcon(event)" class="cal__chip-status" aria-hidden="true" />
                {{ event.title }}
              </span>
              <span v-if="statusText(event)" class="sr-only">({{ statusText(event) }})</span>
            </button>
          </li>
        </ul>
      </li>
    </ol>

    <!-- Jour -->
    <div v-else class="cal__day">
      <p v-if="currentHoliday" class="cal__day-holiday">
        <span class="cal__swatch" aria-hidden="true"></span>{{ $t('agenda.holidays') }} · {{ currentHoliday }}
      </p>
      <ul v-if="dayEvents.length" class="cal__day-list">
        <li v-for="event in dayEvents" :key="event.id">
          <button type="button" class="cal__row" :class="chipClasses(event)" :style="chipStyle(event)" @click="selectedEvent = event">
            <span class="cal__row-time">{{ timeLabel(event) }}</span>
            <span class="cal__row-main">
              <span class="cal__row-title">{{ event.title }}</span>
              <span v-if="event.location" class="cal__row-place">
                <Icon icon="lucide:map-pin" aria-hidden="true" />{{ event.location }}
              </span>
            </span>
            <StatusBadge v-if="statusTone(event)" :tone="statusTone(event)">{{ statusText(event) }}</StatusBadge>
          </button>
        </li>
      </ul>
      <EmptyState v-else compact icon="lucide:calendar" :title="$t('calendar.noEventsForDay')" />
    </div>

    <EmptyState
      v-if="!busy && events.length === 0 && currentViewMode !== 'day'"
      compact
      icon="lucide:calendar"
      :title="$t('calendar.noEvents')"
      :description="$t('calendar.noEventsHelp')"
    />

    <EventDetailModal v-if="selectedEvent" :event="selectedEvent" @close="selectedEvent = null" @updated="selectedEvent = null" />
  </section>
</template>

<script setup>
// Calendrier de l'agenda (direction A+) : vues Mois, Semaine, Jour. Week-ends grisés, jours
// fériés hachurés avec leur nom, événements provisoires (pointillés) et annulés (barrés)
// signalés comme exceptions. Les événements arrivent filtrés par la page ; les fériés sont
// passés à part pour rester visibles quel que soit le filtre.
import { computed, ref } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useEventsStore } from '@/stores/events'
import { DEFAULT_WEEKEND, dateKey, eventCoversDay, eventStatusTone } from '@/utils/agenda'
import { dateLocaleFor } from '@/utils/homeFormat'
import SegmentedControl from '@/components/ui/SegmentedControl.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import EventDetailModal from './EventDetailModal.vue'

const props = defineProps({
  events: { type: Array, default: () => [] },
  holidays: { type: Map, default: () => new Map() },
  loading: { type: Boolean, default: false },
})

const { t, locale } = useI18n()
const eventsStore = useEventsStore()

const MONTH_LIMIT = 3
const selectedEvent = ref(null)
const currentDate = ref(new Date())
const currentViewMode = ref('month')
const busy = computed(() => props.loading || eventsStore.isLoadingCalendar)
const loc = computed(() => dateLocaleFor(locale.value))

const viewModes = computed(() => [
  { value: 'month', label: t('calendar.month') },
  { value: 'week', label: t('calendar.week') },
  { value: 'day', label: t('calendar.day') },
])

// ── Dates ────────────────────────────────────────────────────────────
const startOfWeek = (date) => {
  const d = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  d.setDate(d.getDate() - ((d.getDay() + 6) % 7)) // lundi
  return d
}
const addDays = (date, n) => new Date(date.getFullYear(), date.getMonth(), date.getDate() + n)
const formatDay = (date, options) => date.toLocaleDateString(loc.value, options)

const weekdayNames = computed(() => {
  const short = new Intl.DateTimeFormat(loc.value, { weekday: 'short' })
  const narrow = new Intl.DateTimeFormat(loc.value, { weekday: 'narrow' })
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(2024, 0, 1 + i) // 1er janvier 2024 : un lundi
    return { short: short.format(d), narrow: narrow.format(d) }
  })
})

const periodLabel = computed(() => {
  const d = currentDate.value
  if (currentViewMode.value === 'day') return formatDay(d, { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })
  if (currentViewMode.value === 'week') {
    const start = startOfWeek(d)
    return `${formatDay(start, { day: 'numeric', month: 'short' })} – ${formatDay(addDays(start, 6), { day: 'numeric', month: 'short', year: 'numeric' })}`
  }
  return formatDay(d, { month: 'long', year: 'numeric' })
})

// ── Jours et événements ──────────────────────────────────────────────
const todayKey = computed(() => dateKey(new Date()))

const eventsOn = (key) =>
  props.events
    .filter((e) => !e.is_holiday && e.is_published !== false && eventCoversDay(e, key))
    .sort((a, b) => new Date(a.start_date) - new Date(b.start_date))

const describeDay = (date, month = null) => {
  const key = dateKey(date)
  return {
    key,
    date,
    events: eventsOn(key),
    holiday: props.holidays.get(key) || '',
    isWeekend: DEFAULT_WEEKEND.includes(date.getDay()),
    isToday: key === todayKey.value,
    isOutside: month !== null && date.getMonth() !== month,
  }
}

const monthDays = computed(() => {
  const first = new Date(currentDate.value.getFullYear(), currentDate.value.getMonth(), 1)
  const start = startOfWeek(first)
  return Array.from({ length: 42 }, (_, i) => describeDay(addDays(start, i), first.getMonth()))
})

const weekDays = computed(() => {
  const start = startOfWeek(currentDate.value)
  return Array.from({ length: 7 }, (_, i) => describeDay(addDays(start, i)))
})

const dayEvents = computed(() => eventsOn(dateKey(currentDate.value)))
const currentHoliday = computed(() => props.holidays.get(dateKey(currentDate.value)) || '')

// ── Présentation ─────────────────────────────────────────────────────
const dayClasses = (day) => ({
  'is-weekend': day.isWeekend,
  'is-holiday': !!day.holiday,
  'is-today': day.isToday,
  'is-outside': day.isOutside,
})

const dayLabel = (day) =>
  [formatDay(day.date, { weekday: 'long', day: 'numeric', month: 'long' }), day.holiday].filter(Boolean).join(' · ')

const statusTone = (event) => eventStatusTone(event.status)
const statusText = (event) => (statusTone(event) ? t(`events.status.${event.status}`) : '')
const statusIcon = (event) => ({ warn: 'lucide:circle-help', bad: 'lucide:ban' })[statusTone(event)] || ''

const chipClasses = (event) => ({
  'is-tentative': event.status === 'tentative',
  'is-cancelled': event.status === 'cancelled',
})
const chipStyle = (event) => ({ '--ev': event.color || event.category?.color || 'var(--acc)' })

const timeLabel = (event) => {
  if (event.is_all_day) return t('events.badge.allDayShort')
  const d = new Date(event.start_date)
  return Number.isNaN(d.getTime()) ? '' : d.toLocaleTimeString(loc.value, { hour: '2-digit', minute: '2-digit' })
}

// ── Navigation ───────────────────────────────────────────────────────
function navigate(direction) {
  const d = currentDate.value
  if (currentViewMode.value === 'day') currentDate.value = addDays(d, direction)
  else if (currentViewMode.value === 'week') currentDate.value = addDays(d, 7 * direction)
  else currentDate.value = new Date(d.getFullYear(), d.getMonth() + direction, 1)
}

const goToToday = () => {
  currentDate.value = new Date()
}

function openDay(date) {
  currentDate.value = date
  currentViewMode.value = 'day'
}

const setViewMode = (mode) => {
  currentViewMode.value = mode
}

defineExpose({
  refreshCalendar: () => {}, // les données sont gérées par la page
  setViewMode,
})
</script>

<style scoped>
.cal {
  position: relative;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
}

.cal__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 16px;
  padding: 12px 14px;
  border-block-end: 1px solid var(--line);
}

.cal__period {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  color: var(--text);
  text-transform: capitalize;
}

.cal__tools { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.cal__nav { display: inline-flex; gap: 2px; }

.cal__navbtn {
  display: grid;
  place-items: center;
  min-width: 32px;
  height: 32px;
  padding-inline: 6px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
  font-size: 18px;
  cursor: pointer;
}

.cal__navbtn--text { font-size: 12.5px; padding-inline: 10px; }
.cal__navbtn:hover { background: var(--sunken); }
.cal__navbtn:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }
[dir='rtl'] .cal__flip { transform: scaleX(-1); }

.cal__loading {
  position: absolute;
  inset: 0;
  z-index: 2;
  display: grid;
  place-items: center;
  background: color-mix(in oklab, var(--surface) 70%, transparent);
}

.cal__spin { width: 28px; height: 28px; color: var(--muted); animation: cal-spin 1s linear infinite; }
@keyframes cal-spin { to { transform: rotate(360deg); } }

/* ── Mois ── */
.cal__weekdays,
.cal__grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
}

.cal__weekday {
  padding: 8px 6px 6px;
  border-block-end: 1px solid var(--line);
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
  text-align: start;
}

.cal__weekday-narrow { display: none; }

.cal__grid { margin: 0; padding: 0; list-style: none; }

.cal__cell {
  min-width: 0;
  min-height: 104px;
  padding: 4px;
  border-block-end: 1px solid var(--line);
  border-inline-end: 1px solid var(--line);
  background: var(--surface);
}

.cal__cell:nth-child(7n) { border-inline-end: 0; }
.cal__cell:nth-last-child(-n + 7) { border-block-end: 0; }

/* Week-end grisé, férié hachuré (les deux peuvent se cumuler). */
.is-weekend { background: var(--sunken); }

.is-holiday {
  background-image: repeating-linear-gradient(
    135deg,
    transparent 0 6px,
    color-mix(in oklab, var(--muted) 22%, transparent) 6px 7px
  );
}

.is-outside .cal__date,
.is-outside .cal__events { opacity: 0.45; }

.cal__date {
  display: grid;
  place-items: center;
  min-width: 24px;
  height: 24px;
  padding-inline: 4px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--text);
  font-family: var(--f-num);
  font-size: 12px;
  cursor: pointer;
}

.cal__date:hover { background: var(--line); }
.is-today .cal__date { background: var(--fill); color: var(--on-fill); font-weight: 600; }

.cal__date:focus-visible,
.cal__chip:focus-visible,
.cal__more:focus-visible,
.cal__row:focus-visible,
.cal__wday-head:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}

.cal__holiday {
  margin: 2px 0 0;
  padding-inline: 2px;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: var(--f-num);
  font-size: 10px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--muted);
}

.cal__events { margin: 3px 0 0; padding: 0; list-style: none; display: grid; gap: 2px; }

.cal__chip {
  display: flex;
  align-items: center;
  gap: 3px;
  width: 100%;
  min-width: 0;
  padding: 2px 5px;
  border: 0;
  border-inline-start: 3px solid var(--ev);
  border-radius: 2px;
  background: color-mix(in oklab, var(--ev) 13%, var(--surface));
  color: var(--text);
  font: inherit;
  font-size: 11.5px;
  line-height: 1.3;
  text-align: start;
  cursor: pointer;
}

.cal__chip:hover { background: color-mix(in oklab, var(--ev) 22%, var(--surface)); }
.cal__chip-title { min-width: 0; overflow: hidden; white-space: nowrap; text-overflow: ellipsis; }
.cal__chip-status { flex: none; width: 12px; height: 12px; }

/* Exceptions : provisoire en pointillés, annulé barré et atténué. */
.is-tentative {
  outline: 1px dashed var(--warn);
  outline-offset: -1px;
  background: var(--surface);
}

.is-tentative .cal__chip-status { color: var(--warn); }

.is-cancelled {
  border-inline-start-color: var(--bad);
  background: var(--surface);
  color: var(--muted);
}

.is-cancelled .cal__chip-title,
.is-cancelled .cal__row-title { text-decoration: line-through; }
.is-cancelled .cal__chip-status { color: var(--bad); }

.cal__more {
  margin-block-start: 2px;
  padding: 0 4px;
  border: 0;
  background: none;
  color: var(--muted);
  font-size: 11px;
  cursor: pointer;
}

.cal__more:hover { color: var(--text); text-decoration: underline; }

/* ── Semaine ── */
.cal__week {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  margin: 0;
  padding: 0;
  list-style: none;
}

.cal__wday {
  min-width: 0;
  min-height: 220px;
  padding: 6px;
  border-inline-end: 1px solid var(--line);
}

.cal__wday:last-child { border-inline-end: 0; }

.cal__wday-head {
  display: flex;
  align-items: baseline;
  gap: 6px;
  width: 100%;
  margin-block-end: 4px;
  padding: 2px 4px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--text);
  cursor: pointer;
}

.cal__wday-head:hover { background: var(--line); }

.cal__wday-name {
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.cal__wday-num { font-family: var(--f-num); font-size: 15px; font-weight: 600; }
.is-today .cal__wday-num { color: var(--acc-ink); }

.cal__chip--tall { flex-direction: column; align-items: stretch; gap: 0; padding-block: 4px; }
.cal__chip--tall .cal__chip-title { white-space: normal; }
.cal__chip-time { font-family: var(--f-num); font-size: 10.5px; color: var(--muted); }

/* ── Jour ── */
.cal__day { padding: 14px; }

.cal__day-holiday {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 12px;
  font-size: 13px;
  color: var(--text);
}

.cal__swatch {
  width: 18px;
  height: 18px;
  border: 1px solid var(--line);
  border-radius: 2px;
  background-image: repeating-linear-gradient(135deg, transparent 0 4px, var(--muted) 4px 5px);
}

.cal__day-list { margin: 0; padding: 0; list-style: none; display: grid; gap: 6px; }

.cal__row {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-inline-start: 3px solid var(--ev);
  border-radius: var(--r-s);
  background: var(--surface);
  color: var(--text);
  font: inherit;
  text-align: start;
  cursor: pointer;
}

.cal__row:hover { background: var(--sunken); }
.cal__row.is-cancelled { border-inline-start-color: var(--bad); color: var(--muted); }
.cal__row.is-tentative { border-style: dashed; border-color: var(--warn); outline: 0; }

.cal__row-time {
  flex: none;
  min-width: 64px;
  font-family: var(--f-num);
  font-size: 12.5px;
  color: var(--muted);
}

.cal__row-main { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
.cal__row-title { font-weight: 600; }
.cal__row-place { display: inline-flex; align-items: center; gap: 4px; font-size: 12.5px; color: var(--muted); }

@media (max-width: 767px) {
  .cal__weekday-long { display: none; }
  .cal__weekday-narrow { display: inline; }
  .cal__cell { min-height: 64px; padding: 2px; }
  .cal__holiday { display: none; }
  /* Mois sur mobile : une barre de couleur par événement, le titre reste lu par les lecteurs d'écran. */
  .cal__month .cal__chip { height: 6px; padding: 0; border-inline-start-width: 0; background: var(--ev); }
  .cal__month .cal__chip.is-cancelled { background: var(--line); }
  .cal__month .cal__chip.is-tentative { background: transparent; outline: 1px dashed var(--warn); }
  .cal__month .cal__chip-title { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
  .cal__month .cal__chip-status { display: none; }
  .cal__more { font-size: 10px; padding: 0 2px; }

  /* Semaine en liste de jours sur mobile. */
  .cal__week { grid-template-columns: minmax(0, 1fr); }
  .cal__wday { min-height: 0; border-inline-end: 0; border-block-end: 1px solid var(--line); }
  .cal__wday:last-child { border-block-end: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .cal__spin { animation: none; }
}
</style>
