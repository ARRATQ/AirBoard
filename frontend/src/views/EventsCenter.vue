<template>
  <div class="content-area agenda">
    <PageHeader :eyebrow="$t('home.rail.agenda')" :title="$t('agenda.title')" :description="$t('agenda.subtitle')">
      <template #actions>
        <SegmentedControl v-model="currentView" :options="views" :label="$t('agenda.viewLabel')" />
      </template>
    </PageHeader>

    <!-- Filtres : catégories et jours fériés -->
    <div class="agenda__filters" role="group" :aria-label="$t('agenda.filterByCategory')">
      <button type="button" class="chip" :aria-pressed="isAllSelected" @click="selectAll">
        {{ $t('agenda.allCategories') }}<span class="chip__count">{{ sourceEvents.length }}</span>
      </button>
      <button
        v-for="category in categories"
        :key="category.id"
        type="button"
        class="chip"
        :aria-pressed="selectedCategoryId === category.id"
        @click="selectCategory(category.id)"
      >
        <span class="chip__dot" :style="{ background: category.color || 'var(--muted)' }" aria-hidden="true"></span>
        {{ category.name }}<span class="chip__count">{{ countFor((e) => e.category_id === category.id) }}</span>
      </button>
      <button type="button" class="chip" :aria-pressed="showOnlyHolidays" @click="toggleHolidaysFilter">
        <span class="chip__hatch" aria-hidden="true"></span>
        {{ $t('agenda.holidays') }}<span class="chip__count">{{ countFor((e) => e.is_holiday === true) }}</span>
      </button>
      <span v-if="!categories.length && !isLoading" class="agenda__nocat">{{ $t('agenda.noCategories') }}</span>
    </div>

    <EventCalendar
      v-if="currentView === 'calendar'"
      ref="calendarComponent"
      :events="displayEvents"
      :holidays="holidays"
      :loading="isLoading"
    />

    <section v-else class="agenda__list" aria-labelledby="agenda-list-title">
      <h2 id="agenda-list-title" class="agenda__list-title">
        {{ $t('events.center.title') }} <span class="agenda__list-count">{{ displayEvents.length }}</span>
      </h2>

      <div v-if="isLoading" class="agenda__state" role="status">
        <Icon icon="mdi:loading" class="agenda__spin" aria-hidden="true" />
        <span class="sr-only">{{ $t('agenda.loading') }}</span>
      </div>

      <EmptyState v-else-if="!listRows.length" icon="mdi:calendar-outline" :title="$t('agenda.noEventsFound')">
        <BaseButton size="sm" icon="mdi:refresh" @click="loadData">{{ $t('agenda.reload') }}</BaseButton>
      </EmptyState>

      <BoardTable v-else :columns="listColumns" :rows="listRows" :caption="$t('events.center.title')" min-width="760px">
        <template #cell-when="{ row }">
          <time :datetime="row.event.start_date" class="agenda__when">{{ row.when }}</time>
        </template>
        <template #cell-title="{ row }">
          <button type="button" class="agenda__title" :class="{ 'is-cancelled': row.event.status === 'cancelled' }" @click="viewEvent(row.event)">
            {{ row.event.title }}
          </button>
          <span v-if="row.excerpt" class="agenda__excerpt">{{ row.excerpt }}</span>
        </template>
        <template #cell-category="{ row }">
          <span v-if="row.event.is_holiday" class="agenda__cat"><span class="chip__hatch" aria-hidden="true"></span>{{ $t('agenda.holidays') }}</span>
          <span v-else-if="row.event.category" class="agenda__cat">
            <span class="chip__dot" :style="{ background: row.event.category.color || 'var(--muted)' }" aria-hidden="true"></span>{{ row.event.category.name }}
          </span>
        </template>
        <template #cell-status="{ row }">
          <StatusBadge v-if="row.tone" :tone="row.tone">{{ $t(`events.status.${row.event.status}`) }}</StatusBadge>
          <StatusBadge v-else-if="row.event.is_published === false" tone="neutral">{{ $t('agenda.draft') }}</StatusBadge>
        </template>
      </BoardTable>
    </section>

    <WorkingDaysCalculator class="agenda__calc" :holidays="holidays" />

    <EventDetailModal v-if="selectedEvent" :event="selectedEvent" @close="selectedEvent = null" @updated="handleEventUpdated" />
  </div>
</template>

<script setup>
// Agenda : calendrier (mois, semaine, jour) ou liste, filtres par catégorie et jours fériés,
// calculateur de jours ouvrés. Les fériés sont calculés sur tous les événements, filtre ou non.
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useEventsStore } from '@/stores/events'
import { useAppStore } from '@/stores/app'
import { eventStatusTone, holidayMap } from '@/utils/agenda'
import { dateLocaleFor } from '@/utils/homeFormat'
import PageHeader from '@/components/ui/PageHeader.vue'
import SegmentedControl from '@/components/ui/SegmentedControl.vue'
import BoardTable from '@/components/ui/BoardTable.vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import EventCalendar from '@/components/events/EventCalendar.vue'
import EventDetailModal from '@/components/events/EventDetailModal.vue'
import WorkingDaysCalculator from '@/components/events/WorkingDaysCalculator.vue'

const router = useRouter()
const eventsStore = useEventsStore()
const appStore = useAppStore()
const { t, locale } = useI18n()

const isLoading = ref(false)
const currentView = ref('calendar')
const calendarComponent = ref(null)
const selectedEvent = ref(null)
const selectedCategoryId = ref(null)
const showOnlyHolidays = ref(false)

const views = computed(() => [
  { value: 'calendar', label: t('agenda.calendar'), icon: 'mdi:calendar-month-outline' },
  { value: 'list', label: t('agenda.list'), icon: 'mdi:format-list-bulleted' },
])

const categories = computed(() => eventsStore.categories || [])

// Événements de la liste ; à défaut, ceux du calendrier de l'année.
const sourceEvents = computed(() => {
  const regular = eventsStore.events || []
  const list = regular.length ? regular : eventsStore.calendarEvents || []
  return list.filter((e) => e?.start_date && !Number.isNaN(new Date(e.start_date).getTime()))
})

const holidays = computed(() => holidayMap(sourceEvents.value))

const displayEvents = computed(() =>
  sourceEvents.value.filter((e) => {
    if (selectedCategoryId.value !== null && e.category_id !== selectedCategoryId.value) return false
    if (showOnlyHolidays.value && e.is_holiday !== true) return false
    return true
  }),
)

const countFor = (predicate) => sourceEvents.value.filter(predicate).length
const isAllSelected = computed(() => selectedCategoryId.value === null && !showOnlyHolidays.value)

function selectAll() {
  selectedCategoryId.value = null
  showOnlyHolidays.value = false
}

function selectCategory(id) {
  selectedCategoryId.value = id
  showOnlyHolidays.value = false
}

function toggleHolidaysFilter() {
  showOnlyHolidays.value = !showOnlyHolidays.value
  if (showOnlyHolidays.value) selectedCategoryId.value = null
}

// ── Liste ────────────────────────────────────────────────────────────
const listColumns = computed(() => [
  { key: 'when', label: t('agenda.columns.when') },
  { key: 'title', label: t('agenda.columns.event') },
  { key: 'category', label: t('agenda.categories') },
  { key: 'location', label: t('agenda.columns.place'), tone: 'dim' },
  { key: 'status', label: t('agenda.columns.status'), align: 'end' },
])

const formatWhen = (event) => {
  const date = new Date(event.start_date)
  const loc = dateLocaleFor(locale.value)
  const day = date.toLocaleDateString(loc, { weekday: 'short', day: '2-digit', month: 'short', year: 'numeric' })
  if (event.is_all_day || event.is_holiday) return day
  return `${day} · ${date.toLocaleTimeString(loc, { hour: '2-digit', minute: '2-digit' })}`
}

const listRows = computed(() =>
  [...displayEvents.value]
    .sort((a, b) => new Date(a.start_date) - new Date(b.start_date))
    .map((event) => ({
      id: event.id,
      event,
      when: formatWhen(event),
      location: event.location || '—',
      excerpt: extractTextFromTiptap(event.description).slice(0, 140),
      tone: eventStatusTone(event.status),
    })),
)

// Texte brut d'un contenu Tiptap (JSON ou chaîne)
function extractTextFromTiptap(content) {
  if (!content) return ''
  let node = content
  if (typeof content === 'string') {
    try {
      node = JSON.parse(content)
    } catch {
      return content
    }
  }
  const walk = (n) => {
    if (!n) return ''
    const own = n.type === 'text' && n.text ? n.text : ''
    return own + (Array.isArray(n.content) ? n.content.map(walk).join(' ') : '')
  }
  return walk(node).replace(/\s+/g, ' ').trim()
}

// ── Navigation et détail ─────────────────────────────────────────────
function viewEvent(event) {
  if (event.slug && event.slug.trim()) {
    router.push(`/events/${event.slug}`)
    return
  }
  // Sans slug : détail dans la modale
  selectedEvent.value = {
    ...event,
    title: event.title || '',
    description: event.description || '',
    location: event.location || '',
    category: event.category || null,
    color: event.color || '#3B82F6',
    priority: event.priority || 'normal',
    status: event.status || 'confirmed',
    is_published: event.is_published !== false,
  }
}

function handleEventUpdated() {
  selectedEvent.value = null
  loadData()
}

// ── Chargement ───────────────────────────────────────────────────────
async function loadData() {
  isLoading.value = true
  try {
    const year = new Date().getFullYear()
    await Promise.all([
      eventsStore.fetchEvents(),
      eventsStore.fetchCategories(),
      eventsStore.fetchCalendarEvents(`${year}-01-01`, `${year}-12-31`),
    ])
  } catch (error) {
    console.error('Error loading data:', error)
    appStore.showError(t('ui.toast.error'))
  } finally {
    isLoading.value = false
  }
}

loadData()
</script>

<style scoped>
.agenda { display: flex; flex-direction: column; gap: 16px; }

.agenda__filters { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; }

.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding-inline: 10px;
  border: 1px solid var(--line);
  border-radius: 3px;
  background: var(--surface);
  color: var(--text);
  font: inherit;
  font-size: 12.5px;
  cursor: pointer;
}

.chip:hover { border-color: var(--muted); }
.chip[aria-pressed='true'] { border-color: var(--text); box-shadow: inset 0 0 0 1px var(--text); font-weight: 600; }
.chip:focus-visible { outline: 2px solid var(--acc); outline-offset: 1px; }

.chip__count { font-family: var(--f-num); font-size: 11px; font-weight: 400; color: var(--muted); }
.chip__dot { width: 8px; height: 8px; border-radius: 50%; }

.chip__hatch {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 1px solid var(--line);
  border-radius: 2px;
  background-image: repeating-linear-gradient(135deg, transparent 0 3px, var(--muted) 3px 4px);
}

.agenda__nocat { font-size: 12.5px; color: var(--muted); }

.agenda__list-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
  margin: 0 0 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text);
}

.agenda__list-count { font-family: var(--f-num); font-size: 12px; font-weight: 400; color: var(--muted); }

.agenda__when { white-space: nowrap; color: var(--muted); }

.agenda__title {
  padding: 0;
  border: 0;
  background: none;
  color: var(--text);
  font-family: var(--f-body);
  font-size: 13.5px;
  font-weight: 600;
  text-align: start;
  cursor: pointer;
}

.agenda__title:hover { color: var(--acc-ink); text-decoration: underline; }
.agenda__title:focus-visible { outline: 2px solid var(--acc); outline-offset: 2px; border-radius: 2px; }
.agenda__title.is-cancelled { color: var(--muted); text-decoration: line-through; }

.agenda__excerpt {
  display: block;
  max-width: 44ch;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: var(--f-body);
  font-size: 12px;
  color: var(--muted);
}

.agenda__cat { display: inline-flex; align-items: center; gap: 6px; font-family: var(--f-body); }

.agenda__state { display: flex; justify-content: center; padding-block: 40px; color: var(--muted); }
.agenda__spin { width: 28px; height: 28px; animation: agenda-spin 1s linear infinite; }
@keyframes agenda-spin { to { transform: rotate(360deg); } }

.agenda__calc { margin-block-start: 8px; }

@media (prefers-reduced-motion: reduce) {
  .agenda__spin { animation: none; }
}
</style>
