<template>
  <div class="card hover:shadow-lg transition-shadow">
    <div class="flex items-start gap-4">
      <!-- Indicateur de statut -->
      <div class="flex-shrink-0 pt-1">
        <div
          :class="event.is_published ? 'bg-green-500' : 'bg-gray-400'"
          class="h-3 w-3 rounded-full"
          :title="event.is_published ? t('events.admin.published') : t('events.admin.draft')"
        ></div>
      </div>

      <!-- Contenu -->
      <div class="flex-1 min-w-0">
        <div class="flex items-start justify-between gap-4">
          <div class="flex-1">
            <div class="flex items-center gap-2 mb-1">
              <Icon
                v-if="event.is_recurring"
                icon="mdi:repeat"
                class="h-4 w-4 text-green-500"
                :title="t('groupAdmin.eventsManagement.indicators.recurring')"
              />
              <Icon
                v-if="event.is_all_day"
                icon="mdi:weather-sunny"
                class="h-4 w-4 text-blue-500"
                :title="t('groupAdmin.eventsManagement.indicators.allDay')"
              />

              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                <router-link
                  v-if="event.slug"
                  :to="{ name: 'EventDetail', params: { slug: event.slug } }"
                  class="hover:text-primary-600 dark:hover:text-primary-400 transition-colors"
                >
                  {{ event.title }}
                </router-link>
                <template v-else>{{ event.title }}</template>
              </h3>

              <!-- Priorité -->
              <span
                v-if="event.priority && event.priority !== 'normal'"
                :class="{
                  'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200': event.priority === 'urgent',
                  'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200': event.priority === 'important'
                }"
                class="px-2 py-0.5 rounded text-xs font-medium"
              >
                {{ t(`events.priority.${event.priority}`) }}
              </span>
            </div>

            <p v-if="event.description" class="text-sm text-gray-600 dark:text-gray-400 mb-2 line-clamp-2">
              {{ extractTextFromTiptap(event.description) }}
            </p>

            <div class="flex flex-wrap items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
              <div v-if="event.category" class="flex items-center gap-1">
                <Icon :icon="event.category.icon || 'mdi:folder'" class="h-4 w-4" />
                <span>{{ event.category.name }}</span>
              </div>

              <div class="flex items-center gap-1">
                <Icon icon="mdi:account" class="h-4 w-4" />
                <span>{{ authorName }}</span>
              </div>

              <div class="flex items-center gap-1">
                <Icon icon="mdi:calendar" class="h-4 w-4" />
                <span>{{ formatDateRange(event.start_date, event.end_date) }}</span>
              </div>

              <div v-if="!event.is_all_day" class="flex items-center gap-1">
                <Icon icon="mdi:clock" class="h-4 w-4" />
                <span>{{ formatTimeRange(event.start_date, event.end_date) }}</span>
              </div>

              <div v-if="event.location" class="flex items-center gap-1">
                <Icon icon="mdi:map-marker" class="h-4 w-4" />
                <span class="line-clamp-1">{{ event.location }}</span>
              </div>

              <div v-if="event.views" class="flex items-center gap-1">
                <Icon icon="mdi:eye" class="h-4 w-4" />
                <span>{{ event.views }}</span>
              </div>

              <!-- Visibilité : groupes ciblés ou public -->
              <div v-if="targetGroups.length > 0" class="flex items-center gap-1">
                <Icon icon="mdi:account-group" class="h-4 w-4" />
                <div class="flex gap-1">
                  <span
                    v-for="group in targetGroups.slice(0, MAX_VISIBLE_GROUPS)"
                    :key="group.id"
                    class="px-1.5 py-0.5 bg-primary-100 dark:bg-primary-900 rounded text-xs"
                  >
                    {{ group.name }}
                  </span>
                  <span v-if="targetGroups.length > MAX_VISIBLE_GROUPS" class="text-xs">
                    +{{ targetGroups.length - MAX_VISIBLE_GROUPS }}
                  </span>
                </div>
              </div>
              <div v-else class="flex items-center gap-1">
                <Icon icon="mdi:earth" class="h-4 w-4" />
                <span>{{ t('events.visibility.public') }}</span>
              </div>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center gap-2">
            <router-link
              :to="`/admin/events/${event.slug}/edit`"
              class="btn btn-secondary btn-sm"
              :title="t('events.admin.edit')"
            >
              <Icon icon="mdi:pencil" class="h-4 w-4" />
            </router-link>

            <button
              type="button"
              class="btn btn-secondary btn-sm"
              :title="event.is_published ? t('events.admin.unpublish') : t('events.admin.publish')"
              @click="emit('toggle-publish', event)"
            >
              <Icon :icon="event.is_published ? 'mdi:eye-off' : 'mdi:eye'" class="h-4 w-4" />
            </button>

            <button
              type="button"
              class="btn btn-danger btn-sm"
              :title="t('events.admin.delete')"
              @click="emit('delete', event)"
            >
              <Icon icon="mdi:delete" class="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'

// Nombre de groupes ciblés affichés avant le compteur « +N »
const MAX_VISIBLE_GROUPS = 2

const props = defineProps({
  event: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['toggle-publish', 'delete'])

const { t } = useI18n()

const targetGroups = computed(() => props.event.target_groups || [])

const authorName = computed(() => {
  const author = props.event.author
  const first = author?.first_name || author?.username || t('groupAdmin.eventsManagement.author.unknown')
  return `${first} ${author?.last_name || ''}`.trim()
})

const DATE_OPTIONS = { year: 'numeric', month: 'short', day: 'numeric' }
const TIME_OPTIONS = { hour: '2-digit', minute: '2-digit' }

const formatDateRange = (startDate, endDate) => {
  if (!startDate) return ''
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  if (end && start.toDateString() !== end.toDateString()) {
    return `${start.toLocaleDateString(undefined, DATE_OPTIONS)} - ${end.toLocaleDateString(undefined, DATE_OPTIONS)}`
  }
  return start.toLocaleDateString(undefined, DATE_OPTIONS)
}

const formatTimeRange = (startDate, endDate) => {
  if (!startDate) return ''
  const start = new Date(startDate)
  const end = endDate ? new Date(endDate) : null
  if (end) {
    return `${start.toLocaleTimeString(undefined, TIME_OPTIONS)} - ${end.toLocaleTimeString(undefined, TIME_OPTIONS)}`
  }
  return start.toLocaleTimeString(undefined, TIME_OPTIONS)
}

// Texte brut d'un contenu Tiptap (JSON) ou d'une chaîne simple
const extractTextFromTiptap = (content) => {
  if (!content) return ''

  let contentObj = content
  if (typeof content === 'string') {
    try {
      contentObj = JSON.parse(content)
    } catch {
      // Texte brut : on l'affiche tel quel
      return content
    }
  }

  const extractText = (node) => {
    if (!node) return ''
    const own = node.type === 'text' && node.text ? node.text : ''
    const children = Array.isArray(node.content)
      ? node.content.map((child) => extractText(child) + (child.type === 'paragraph' ? ' ' : '')).join('')
      : ''
    return own + children
  }

  return extractText(contentObj).trim()
}
</script>

<style scoped>
.line-clamp-2 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
}

.line-clamp-1 {
  overflow: hidden;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 1;
  line-clamp: 1;
}
</style>
