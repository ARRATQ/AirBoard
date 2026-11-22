<template>
  <WorkingDaysCalculator compact :holidays="holidays">
    <template #action>
      <router-link to="/events" class="home-card__more">
        {{ $t('home.rail.agenda') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </template>
    <template #note>
      <p v-if="loadFailed" class="wd__note" role="status">
        <Icon icon="lucide:circle-alert" aria-hidden="true" />
        {{ $t('home.workingDays.holidaysUnavailable') }}
      </p>
    </template>
  </WorkingDaysCalculator>
</template>

<script setup>
// Calculateur de jours ouvrés sur l'accueil, avec les jours fériés de l'agenda (année en cours
// et suivante). Si l'agenda ne répond pas, le calcul reste possible et le dit.
import { onMounted, ref } from 'vue'
import { Icon } from '@iconify/vue'
import { eventsService } from '@/services/api'
import { holidayRange, holidaysFromCalendar } from '@/utils/calendarHolidays'
import WorkingDaysCalculator from '@/components/events/WorkingDaysCalculator.vue'

const holidays = ref(new Map())
const loadFailed = ref(false)

onMounted(async () => {
  const { start, end } = holidayRange()
  try {
    holidays.value = holidaysFromCalendar(await eventsService.getCalendarView(start, end))
  } catch (error) {
    console.error('Failed to load holidays for the working days calculator:', error)
    loadFailed.value = true
  }
})
</script>

<style scoped>
.wd__note {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin: 10px 0 0;
  font-size: 12px;
  line-height: 1.4;
  color: var(--warn);
}

.wd__note :deep(svg) { flex: none; width: 14px; height: 14px; margin-top: 1px; }
</style>
