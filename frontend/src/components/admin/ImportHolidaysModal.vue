<template>
  <div
    class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
    @click.self="$emit('close')"
  >
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-red-100 dark:bg-red-900 rounded-lg">
            <Icon icon="mdi:calendar-star" class="h-6 w-6 text-red-600 dark:text-red-400" />
          </div>
          <div>
            <h2 class="text-xl font-bold text-gray-900 dark:text-white">
              {{ $t('events.holidays.importTitle') }}
            </h2>
            <p class="text-sm text-gray-500 dark:text-gray-400">
              {{ $t('events.holidays.importSubtitle') }}
            </p>
          </div>
        </div>
        <button
          @click="$emit('close')"
          class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
        >
          <Icon icon="mdi:close" class="h-5 w-5 text-gray-500" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 overflow-y-auto flex-1">
        <!-- Step 1: Country Selection -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('events.holidays.selectCountry') }}
          </label>
          <div class="relative">
            <select
              v-model="selectedCountry"
              class="input w-full"
              :disabled="isLoadingCountries"
            >
              <option value="">{{ $t('events.holidays.chooseCountry') }}</option>
              <option v-for="country in countries" :key="country.countryCode" :value="country.countryCode">
                {{ country.name }} ({{ country.countryCode }})
              </option>
            </select>
            <div v-if="isLoadingCountries" class="absolute right-3 top-1/2 -translate-y-1/2">
              <Icon icon="mdi:loading" class="h-5 w-5 animate-spin text-gray-400" />
            </div>
          </div>
        </div>

        <!-- Step 2: Year Selection -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('events.holidays.selectYear') }}
          </label>
          <select v-model="selectedYear" class="input w-full">
            <option v-for="year in availableYears" :key="year" :value="year">
              {{ year }}
            </option>
          </select>
        </div>

        <!-- Step 3: Category Selection (Optional) -->
        <div class="mb-6">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            {{ $t('events.holidays.selectCategory') }}
            <span class="text-gray-400 font-normal">({{ $t('common.optional') }})</span>
          </label>
          <select v-model="selectedCategoryId" class="input w-full">
            <option :value="null">{{ $t('events.holidays.noCategory') }}</option>
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.name }}
            </option>
          </select>
        </div>

        <!-- Preview Button -->
        <div class="mb-6">
          <button
            @click="previewHolidays"
            :disabled="!selectedCountry || isLoadingPreview"
            class="btn btn-secondary w-full"
          >
            <Icon v-if="isLoadingPreview" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
            <Icon v-else icon="mdi:eye" class="h-4 w-4 mr-2" />
            {{ $t('events.holidays.preview') }}
          </button>
        </div>

        <!-- Preview Results -->
        <div v-if="previewData" class="mb-6">
          <div class="bg-gray-50 dark:bg-gray-700 rounded-lg p-4">
            <div class="flex items-center justify-between mb-4">
              <h3 class="font-medium text-gray-900 dark:text-white">
                {{ $t('events.holidays.previewResults', { count: previewData.count }) }}
              </h3>
              <span class="text-sm text-gray-500 dark:text-gray-400">
                {{ selectedYear }}
              </span>
            </div>
            <div class="max-h-48 overflow-y-auto space-y-2">
              <div
                v-for="holiday in previewData.holidays"
                :key="holiday.date"
                class="flex items-center justify-between py-2 px-3 bg-white dark:bg-gray-600 rounded-lg"
              >
                <div class="flex items-center gap-3">
                  <div class="w-2 h-2 rounded-full bg-red-500"></div>
                  <span class="font-medium text-gray-900 dark:text-white">{{ holiday.localName }}</span>
                </div>
                <span class="text-sm text-gray-500 dark:text-gray-400">
                  {{ formatDate(holiday.date) }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Error Message -->
        <div v-if="error" class="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
          <div class="flex items-center gap-2 text-red-600 dark:text-red-400">
            <Icon icon="mdi:alert-circle" class="h-5 w-5" />
            <span>{{ error }}</span>
          </div>
        </div>

        <!-- Success Message -->
        <div v-if="importResult" class="mb-6 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
          <div class="flex items-center gap-2 text-green-600 dark:text-green-400 mb-2">
            <Icon icon="mdi:check-circle" class="h-5 w-5" />
            <span class="font-medium">{{ $t('events.holidays.importSuccess') }}</span>
          </div>
          <ul class="text-sm text-green-600 dark:text-green-400 ml-7 space-y-1">
            <li>{{ $t('events.holidays.imported', { count: importResult.imported }) }}</li>
            <li v-if="importResult.skipped > 0">{{ $t('events.holidays.skipped', { count: importResult.skipped }) }}</li>
          </ul>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 p-6 border-t border-gray-200 dark:border-gray-700">
        <button @click="$emit('close')" class="btn btn-secondary">
          {{ $t('common.cancel') }}
        </button>
        <button
          @click="importHolidays"
          :disabled="!selectedCountry || isImporting"
          class="btn btn-primary"
        >
          <Icon v-if="isImporting" icon="mdi:loading" class="h-4 w-4 mr-2 animate-spin" />
          <Icon v-else icon="mdi:download" class="h-4 w-4 mr-2" />
          {{ $t('events.holidays.import') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'
import { adminEventsService } from '@/services/api'

const props = defineProps({
  categories: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['close', 'imported'])

const { t, locale } = useI18n()

// State
const countries = ref([])
const selectedCountry = ref('')
const selectedYear = ref(new Date().getFullYear())
const selectedCategoryId = ref(null)
const previewData = ref(null)
const importResult = ref(null)
const error = ref('')

const isLoadingCountries = ref(false)
const isLoadingPreview = ref(false)
const isImporting = ref(false)

// Computed
const availableYears = computed(() => {
  const currentYear = new Date().getFullYear()
  const years = []
  for (let i = currentYear - 2; i <= currentYear + 3; i++) {
    years.push(i)
  }
  return years
})

// Methods
const loadCountries = async () => {
  isLoadingCountries.value = true
  error.value = ''
  try {
    const response = await adminEventsService.getAvailableCountries()
    countries.value = response.countries || []
  } catch (err) {
    console.error('Error loading countries:', err)
    error.value = t('events.holidays.loadCountriesError')
  } finally {
    isLoadingCountries.value = false
  }
}

const previewHolidays = async () => {
  if (!selectedCountry.value) return

  isLoadingPreview.value = true
  error.value = ''
  previewData.value = null

  try {
    const response = await adminEventsService.previewHolidays(selectedCountry.value, selectedYear.value)
    previewData.value = response
  } catch (err) {
    console.error('Error previewing holidays:', err)
    error.value = t('events.holidays.previewError')
  } finally {
    isLoadingPreview.value = false
  }
}

const importHolidays = async () => {
  if (!selectedCountry.value) return

  isImporting.value = true
  error.value = ''
  importResult.value = null

  try {
    const response = await adminEventsService.importHolidays(
      selectedCountry.value,
      selectedYear.value,
      selectedCategoryId.value
    )
    importResult.value = response
    emit('imported', response)
  } catch (err) {
    console.error('Error importing holidays:', err)
    error.value = err.response?.data?.error || t('events.holidays.importError')
  } finally {
    isImporting.value = false
  }
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
    month: 'long'
  })
}

// Lifecycle
onMounted(() => {
  loadCountries()
})
</script>
