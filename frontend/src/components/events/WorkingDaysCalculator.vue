<template>
  <div class="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
    <!-- Header -->
    <div class="flex items-center gap-3 mb-6">
      <div class="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
        <Icon icon="mdi:calculator" class="h-6 w-6 text-blue-600 dark:text-blue-400" />
      </div>
      <div>
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ $t('agenda.calculator.title') }}
        </h3>
        <p class="text-sm text-gray-500 dark:text-gray-400">
          {{ $t('agenda.calculator.subtitle') }}
        </p>
      </div>
    </div>

    <!-- Weekend Configuration -->
    <div class="mb-6">
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
        {{ $t('agenda.calculator.weekendDays') }}
      </label>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="(day, index) in weekDays"
          :key="index"
          @click="toggleWeekendDay(index)"
          :class="[
            'px-3 py-2 rounded-lg text-sm font-medium transition-all border-2',
            weekendDays.includes(index)
              ? 'bg-orange-100 dark:bg-orange-900/30 border-orange-500 text-orange-700 dark:text-orange-300'
              : 'bg-gray-50 dark:bg-gray-700 border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 hover:border-gray-300'
          ]"
        >
          {{ day }}
        </button>
      </div>
      <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">
        {{ $t('agenda.calculator.weekendHint') }}
      </p>
    </div>

    <!-- Date Range -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-6">
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          {{ $t('agenda.calculator.startDate') }}
        </label>
        <input
          v-model="startDate"
          type="date"
          class="input w-full"
        />
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
          {{ $t('agenda.calculator.endDate') }}
        </label>
        <input
          v-model="endDate"
          type="date"
          class="input w-full"
        />
      </div>
    </div>

    <!-- Include Holidays Toggle -->
    <div class="mb-6">
      <label class="flex items-center gap-3 cursor-pointer">
        <input
          v-model="excludeHolidays"
          type="checkbox"
          class="w-5 h-5 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
        />
        <span class="text-sm text-gray-700 dark:text-gray-300">
          {{ $t('agenda.calculator.excludeHolidays') }}
        </span>
      </label>
    </div>

    <!-- Calculate Button -->
    <button
      @click="calculate"
      :disabled="!startDate || !endDate"
      class="btn btn-primary w-full mb-6"
    >
      <Icon icon="mdi:calculator-variant" class="h-5 w-5 mr-2" />
      {{ $t('agenda.calculator.calculate') }}
    </button>

    <!-- Results -->
    <div v-if="results" class="space-y-4">
      <!-- Total Days -->
      <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
        <div class="text-center">
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">
            {{ $t('agenda.calculator.totalDays') }}
          </p>
          <p class="text-3xl font-bold text-gray-900 dark:text-white">
            {{ results.totalDays }}
          </p>
        </div>
      </div>

      <!-- Working vs Non-Working -->
      <div class="grid grid-cols-2 gap-4">
        <div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg p-4 text-center">
          <Icon icon="mdi:briefcase-check" class="h-8 w-8 mx-auto mb-2 text-green-600 dark:text-green-400" />
          <p class="text-sm text-green-600 dark:text-green-400 mb-1">
            {{ $t('agenda.calculator.workingDays') }}
          </p>
          <p class="text-2xl font-bold text-green-700 dark:text-green-300">
            {{ results.workingDays }}
          </p>
        </div>
        <div class="bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-800 rounded-lg p-4 text-center">
          <Icon icon="mdi:beach" class="h-8 w-8 mx-auto mb-2 text-orange-600 dark:text-orange-400" />
          <p class="text-sm text-orange-600 dark:text-orange-400 mb-1">
            {{ $t('agenda.calculator.nonWorkingDays') }}
          </p>
          <p class="text-2xl font-bold text-orange-700 dark:text-orange-300">
            {{ results.nonWorkingDays }}
          </p>
        </div>
      </div>

      <!-- Breakdown -->
      <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
          {{ $t('agenda.calculator.breakdown') }}
        </h4>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-600 dark:text-gray-400">
              {{ $t('agenda.calculator.weekendCount') }}
            </span>
            <span class="font-medium text-gray-900 dark:text-white">
              {{ results.weekendDays }} {{ $t('agenda.calculator.days') }}
            </span>
          </div>
          <div v-if="excludeHolidays" class="flex justify-between">
            <span class="text-gray-600 dark:text-gray-400">
              {{ $t('agenda.calculator.holidaysCount') }}
            </span>
            <span class="font-medium text-gray-900 dark:text-white">
              {{ results.holidayDays }} {{ $t('agenda.calculator.days') }}
            </span>
          </div>
        </div>
      </div>

      <!-- Holiday List (if any) -->
      <div v-if="excludeHolidays && results.holidays.length > 0" class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
        <h4 class="text-sm font-medium text-red-700 dark:text-red-300 mb-3 flex items-center gap-2">
          <Icon icon="mdi:calendar-star" class="h-4 w-4" />
          {{ $t('agenda.calculator.holidaysInPeriod') }}
        </h4>
        <ul class="space-y-1 text-sm">
          <li
            v-for="holiday in results.holidays"
            :key="holiday.date"
            class="flex justify-between text-red-600 dark:text-red-400"
          >
            <span>{{ holiday.title }}</span>
            <span class="text-red-500 dark:text-red-500">{{ formatDate(holiday.date) }}</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { useEventsStore } from '@/stores/events'

const { t, locale } = useI18n()
const eventsStore = useEventsStore()

// Week days (0 = Sunday, 1 = Monday, etc.)
const weekDays = computed(() => {
  const days = t('agenda.calculator.daysOfWeek')
  // Parse if it's a string (fallback)
  if (typeof days === 'string') {
    return ['Dim', 'Lun', 'Mar', 'Mer', 'Jeu', 'Ven', 'Sam']
  }
  return days
})

// State
const startDate = ref('')
const endDate = ref('')
const weekendDays = ref([0, 6]) // Default: Sunday (0) and Saturday (6)
const excludeHolidays = ref(true)
const results = ref(null)

// Methods
const toggleWeekendDay = (dayIndex) => {
  const index = weekendDays.value.indexOf(dayIndex)
  if (index > -1) {
    weekendDays.value.splice(index, 1)
  } else {
    weekendDays.value.push(dayIndex)
  }
}

const calculate = () => {
  if (!startDate.value || !endDate.value) return

  const start = new Date(startDate.value)
  const end = new Date(endDate.value)

  // Ensure start <= end
  if (start > end) {
    results.value = null
    return
  }

  // Get holidays in the period
  const holidaysInPeriod = getHolidaysInPeriod(start, end)

  let totalDays = 0
  let workingDays = 0
  let weekendDaysCount = 0
  let holidayDays = 0

  const currentDate = new Date(start)
  while (currentDate <= end) {
    totalDays++

    const dayOfWeek = currentDate.getDay()
    const isWeekend = weekendDays.value.includes(dayOfWeek)
    const dateStr = currentDate.toISOString().split('T')[0]
    const isHoliday = excludeHolidays.value && holidaysInPeriod.some(h => h.date === dateStr)

    if (isWeekend) {
      weekendDaysCount++
    } else if (isHoliday) {
      holidayDays++
    } else {
      workingDays++
    }

    // Move to next day
    currentDate.setDate(currentDate.getDate() + 1)
  }

  results.value = {
    totalDays,
    workingDays,
    nonWorkingDays: totalDays - workingDays,
    weekendDays: weekendDaysCount,
    holidayDays,
    holidays: holidaysInPeriod.filter(h => {
      // Only include holidays that fall on working days
      const date = new Date(h.date)
      return !weekendDays.value.includes(date.getDay())
    })
  }
}

const getHolidaysInPeriod = (start, end) => {
  const events = eventsStore.events || []
  const calendarEvents = eventsStore.calendarEvents || []
  const allEvents = events.length > 0 ? events : calendarEvents

  return allEvents
    .filter(event => {
      if (!event.is_holiday) return false
      const eventDate = new Date(event.start_date)
      return eventDate >= start && eventDate <= end
    })
    .map(event => ({
      date: event.start_date.split('T')[0],
      title: event.title
    }))
}

const formatDate = (dateString) => {
  const date = new Date(dateString)
  const localeMap = {
    'fr': 'fr-FR',
    'en': 'en-US',
    'es': 'es-ES',
    'ar': 'ar-SA'
  }
  return date.toLocaleDateString(localeMap[locale.value] || 'fr-FR', {
    weekday: 'short',
    day: 'numeric',
    month: 'short'
  })
}
</script>
