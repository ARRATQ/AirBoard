<template>
  <section class="wdc" aria-labelledby="wdc-title">
    <header class="wdc__head">
      <h2 id="wdc-title" class="wdc__title">{{ $t('agenda.calculator.title') }}</h2>
      <p class="wdc__sub">{{ $t('agenda.calculator.subtitle') }}</p>
    </header>

    <div class="wdc__body">
      <div class="wdc__form">
        <div class="wdc__dates">
          <label class="wdc__field">
            <span class="wdc__label">{{ $t('agenda.calculator.startDate') }}</span>
            <input v-model="startDate" type="date" class="wdc__input" />
          </label>
          <label class="wdc__field">
            <span class="wdc__label">{{ $t('agenda.calculator.endDate') }}</span>
            <input v-model="endDate" type="date" class="wdc__input" :min="startDate || undefined" />
          </label>
        </div>

        <fieldset class="wdc__weekend">
          <legend class="wdc__label">{{ $t('agenda.calculator.weekendDays') }}</legend>
          <div class="wdc__days">
            <button
              v-for="day in weekOrder"
              :key="day.index"
              type="button"
              class="wdc__day"
              :class="{ 'wdc__day--on': weekendDays.includes(day.index) }"
              :aria-pressed="weekendDays.includes(day.index)"
              :title="day.long"
              @click="toggleWeekendDay(day.index)"
            >
              {{ day.short }}
            </button>
          </div>
          <p class="wdc__hint">{{ $t('agenda.calculator.weekendHint') }}</p>
        </fieldset>

        <label class="wdc__check">
          <input v-model="excludeHolidays" type="checkbox" />
          <span>{{ $t('agenda.calculator.excludeHolidays') }}</span>
        </label>
      </div>

      <div class="wdc__result" aria-live="polite">
        <template v-if="results">
          <p class="wdc__label">{{ $t('agenda.calculator.workingDays') }}</p>
          <p class="wdc__big">{{ results.workingDays }}</p>
          <dl class="wdc__breakdown">
            <div><dt>{{ $t('agenda.calculator.totalDays') }}</dt><dd>{{ results.totalDays }}</dd></div>
            <div><dt>{{ $t('agenda.calculator.weekendCount') }}</dt><dd>{{ results.weekendDays }}</dd></div>
            <div v-if="excludeHolidays"><dt>{{ $t('agenda.calculator.holidaysCount') }}</dt><dd>{{ results.holidayDays }}</dd></div>
            <div><dt>{{ $t('agenda.calculator.nonWorkingDays') }}</dt><dd>{{ results.nonWorkingDays }}</dd></div>
          </dl>
          <div v-if="excludeHolidays && results.holidays.length" class="wdc__holidays">
            <p class="wdc__label">{{ $t('agenda.calculator.holidaysInPeriod') }}</p>
            <ul>
              <li v-for="holiday in results.holidays" :key="holiday.date">
                <span>{{ holiday.title }}</span>
                <time :datetime="holiday.date">{{ formatDate(holiday.date) }}</time>
              </li>
            </ul>
          </div>
        </template>
        <p v-else class="wdc__empty">{{ $t('agenda.calculator.invalidRange') }}</p>
      </div>
    </div>
  </section>
</template>

<script setup>
// Calculateur de jours ouvrés : le résultat suit chaque changement (dates, jours de repos,
// fériés), sans bouton. Les fériés viennent de l'agenda (prop holidays : Map clé → nom).
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { DEFAULT_WEEKEND, countWorkingDays, dateKey, parseDateKey } from '@/utils/agenda'
import { dateLocaleFor } from '@/utils/homeFormat'

const props = defineProps({
  holidays: { type: Map, default: () => new Map() },
})

const { locale } = useI18n()

const today = new Date()
const startDate = ref(dateKey(today))
const endDate = ref(dateKey(new Date(today.getFullYear(), today.getMonth() + 1, 0)))
const weekendDays = ref([...DEFAULT_WEEKEND])
const excludeHolidays = ref(true)

// Lundi → dimanche ; index = Date.getDay(). 2024-01-01 est un lundi.
const weekOrder = computed(() => {
  const loc = dateLocaleFor(locale.value)
  const short = new Intl.DateTimeFormat(loc, { weekday: 'short' })
  const long = new Intl.DateTimeFormat(loc, { weekday: 'long' })
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(2024, 0, 1 + i)
    return { index: d.getDay(), short: short.format(d), long: long.format(d) }
  })
})

function toggleWeekendDay(index) {
  weekendDays.value = weekendDays.value.includes(index)
    ? weekendDays.value.filter((d) => d !== index)
    : [...weekendDays.value, index]
}

const results = computed(() =>
  countWorkingDays({
    start: startDate.value,
    end: endDate.value,
    weekendDays: weekendDays.value,
    holidays: props.holidays,
    excludeHolidays: excludeHolidays.value,
  }),
)

const formatDate = (key) =>
  parseDateKey(key)?.toLocaleDateString(dateLocaleFor(locale.value), { weekday: 'short', day: 'numeric', month: 'short' }) || key
</script>

<style scoped>
.wdc {
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
}

.wdc__head {
  padding: 14px 16px 12px;
  border-block-end: 1px solid var(--line);
}

.wdc__title { margin: 0; font-size: 15px; font-weight: 600; color: var(--text); }
.wdc__sub { margin: 2px 0 0; font-size: 12.5px; color: var(--muted); }

.wdc__body {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr);
}

.wdc__form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.wdc__dates {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.wdc__field { display: flex; flex-direction: column; gap: 4px; }

.wdc__label {
  margin: 0;
  padding: 0;
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.wdc__input {
  height: 36px;
  padding-inline: 10px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
  color: var(--text);
  font-family: var(--f-num);
  font-size: 13px;
}

.wdc__input:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.wdc__weekend { margin: 0; padding: 0; border: 0; min-width: 0; }

.wdc__days {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-block-start: 6px;
}

.wdc__day {
  min-width: 44px;
  height: 32px;
  padding-inline: 8px;
  border: 1px solid var(--line);
  border-radius: 3px;
  background: var(--surface);
  color: var(--text);
  font-size: 12.5px;
  cursor: pointer;
}

.wdc__day:hover { border-color: var(--muted); }
.wdc__day--on { background: var(--sunken); color: var(--muted); text-decoration: line-through; }
.wdc__day:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.wdc__hint { margin: 6px 0 0; font-size: 12px; color: var(--muted); }

.wdc__check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text);
  cursor: pointer;
}

.wdc__check input { width: 16px; height: 16px; accent-color: var(--fill); }

.wdc__result {
  padding: 16px;
  border-inline-start: 1px solid var(--line);
  background: var(--sunken);
}

.wdc__big {
  margin: 2px 0 10px;
  font-family: var(--f-num);
  font-size: 44px;
  font-weight: 600;
  line-height: 1;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

.wdc__breakdown { margin: 0; display: grid; gap: 4px; }

.wdc__breakdown div {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  padding-block: 4px;
  border-block-end: 1px dashed var(--line);
  font-size: 12.5px;
}

.wdc__breakdown dt { color: var(--muted); }
.wdc__breakdown dd { margin: 0; font-family: var(--f-num); color: var(--text); }

.wdc__holidays { margin-block-start: 14px; }

.wdc__holidays ul { margin: 6px 0 0; padding: 0; list-style: none; display: grid; gap: 4px; }

.wdc__holidays li {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 12.5px;
  color: var(--text);
}

.wdc__holidays time { font-family: var(--f-num); color: var(--muted); white-space: nowrap; }

.wdc__empty { margin: 0; font-size: 13px; color: var(--muted); }

@media (max-width: 720px) {
  .wdc__body { grid-template-columns: minmax(0, 1fr); }
  .wdc__result { border-inline-start: 0; border-block-start: 1px solid var(--line); }
}

@media (max-width: 400px) {
  .wdc__dates { grid-template-columns: minmax(0, 1fr); }
}
</style>
