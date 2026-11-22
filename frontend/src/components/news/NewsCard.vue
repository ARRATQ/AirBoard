<template>
  <!-- Carte cliquable : le bouton du titre s'étend sur toute la carte (pseudo-élément).
       Cadre et liseré supérieur à la couleur du type, pour que chaque actualité se détache. -->
  <article class="ncard" :style="{ '--tc': typeColor }">
    <div class="ncard__head">
      <span class="tag tag--custom nbadge--type">{{ typeLabel }}</span>

      <span v-if="news.priority === 'urgent'" class="tag tag--bad">
        <Icon icon="mdi:alert" aria-hidden="true" />
        {{ $t('news.priority.urgent') }}
      </span>
      <span v-else-if="news.priority === 'important'" class="tag tag--warn">
        <Icon icon="mdi:star" aria-hidden="true" />
        {{ $t('news.priority.important') }}
      </span>

      <span
        class="ncard__audience"
        :class="{ 'ncard__audience--groups': hasGroups }"
        :title="getGroupsTooltip(news.target_groups)"
      >
        <Icon :icon="hasGroups ? 'mdi:lock' : 'mdi:earth'" aria-hidden="true" />
        <span class="ncard__ellipsis">{{ getVisibilityLabel(news.target_groups) }}</span>
      </span>

      <span v-if="news.category" class="ncard__category" :title="news.category.name">
        <Icon :icon="news.category.icon || 'mdi:folder'" :style="{ color: news.category.color }" aria-hidden="true" />
        <span class="ncard__ellipsis" dir="auto">{{ news.category.name }}</span>
      </span>
    </div>

    <h3 class="ncard__title">
      <button
        type="button"
        class="btn-reset ncard__link after:absolute after:inset-0 after:content-['']"
        @click="handleClick"
      >
        {{ news.title }}
      </button>
    </h3>

    <p v-if="news.summary" class="ncard__summary">{{ news.summary }}</p>

    <ul v-if="cardTags.shown.length" class="ncard__tags">
      <li
        v-for="tag in cardTags.shown"
        :key="tag.id"
        class="tag tag--custom ntag"
        :style="{ '--tag-c': tag.color || 'var(--muted)' }"
        :title="tag.name"
      >
        <span class="ncard__ellipsis" dir="auto">{{ tag.name }}</span>
      </li>
      <li
        v-if="cardTags.rest.length"
        class="tag ntag ntag--more"
        dir="ltr"
        :title="cardTags.rest.map((tag) => tag.name).join(', ')"
      >
        +{{ cardTags.rest.length }}
      </li>
    </ul>

    <div class="ncard__foot">
      <span class="ncard__meta ncard__author" :title="authorName">
        <Icon icon="mdi:account" aria-hidden="true" />
        <span class="ncard__ellipsis" dir="auto">{{ authorName }}</span>
      </span>
      <span class="ncard__meta">
        <Icon icon="mdi:calendar" aria-hidden="true" />
        {{ formatDate(news.published_at || news.created_at) }}
      </span>
      <span class="ncard__meta">
        <Icon icon="mdi:eye" aria-hidden="true" />
        {{ news.view_count || 0 }}
      </span>
      <span v-if="articleReadingTime" class="ncard__meta">
        <Icon icon="mdi:clock-outline" aria-hidden="true" />
        {{ articleReadingTime }}
      </span>
      <span
        v-for="reaction in aggregateReactions(news.reactions)"
        :key="reaction.type"
        class="ncard__meta"
      >
        <span aria-hidden="true">{{ getReactionEmoji(reaction.type) }}</span>
        {{ reaction.count }}
      </span>
    </div>
  </article>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import { useNewsTypesStore } from '@/stores/newsTypes'
import { useI18n } from 'vue-i18n'
import { readingMinutes } from '@/utils/readingTime'
import { dateLocaleFor } from '@/utils/homeFormat'
import { splitTags } from '@/utils/tagChips'

// Étiquettes visibles sur une carte avant le « +N »
const MAX_CARD_TAGS = 2

const { t, locale } = useI18n()
const typesStore = useNewsTypesStore()

const props = defineProps({
  news: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['click'])

const articleReadingTime = computed(() => {
  const n = readingMinutes(props.news.content)
  return n ? t('news.minutes', { n }) : null
})

// Get type data from store
const typeData = computed(() => {
  if (!typesStore.types.length && !typesStore.loaded) {
    // If types not loaded yet, use defaults
    const defaults = {
      article: { color: '#3B82F6', icon: 'mdi:text-box', name: t('news.types.article') },
      tutorial: { color: '#10B981', icon: 'mdi:book-open-variant', name: t('news.types.tutorial') },
      announcement: { color: '#8B5CF6', icon: 'mdi:bullhorn', name: t('news.types.announcement') },
      faq: { color: '#F59E0B', icon: 'mdi:help-circle', name: t('news.types.faq') }
    }
    return defaults[props.news.type] || defaults.article
  }
  return typesStore.getTypeBySlug(props.news.type) || {}
})

const typeColor = computed(() => typeData.value.color || '#3B82F6')
const hasGroups = computed(() => (props.news.target_groups?.length || 0) > 0)
const authorName = computed(() => props.news.author?.full_name || props.news.author?.email || '')
const cardTags = computed(() => splitTags(props.news.tags, MAX_CARD_TAGS))

// Libellé du type : nom venant du store (types gérés en admin)
const typeLabel = computed(() => {
  const type = typeData.value
  return type.name || props.news.type
})

const handleClick = () => {
  // Emit event for parent components to handle navigation
  emit('click')
}

// Date relative dans la langue de l'interface (jours puis semaines, date courte au-delà d'un mois)
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  const now = new Date()
  const diffInDays = Math.floor((now - date) / (1000 * 60 * 60 * 24))
  const dateLocale = dateLocaleFor(locale.value)
  const relative = new Intl.RelativeTimeFormat(dateLocale, { numeric: 'auto' })

  if (diffInDays < 7) return relative.format(-diffInDays, 'day')
  if (diffInDays < 30) return relative.format(-Math.floor(diffInDays / 7), 'week')

  return date.toLocaleDateString(dateLocale, { day: 'numeric', month: 'short', year: 'numeric' })
}

const aggregateReactions = (reactions) => {
  if (!reactions) return []

  const counts = {}
  reactions.forEach(r => {
    counts[r.reaction_type] = (counts[r.reaction_type] || 0) + 1
  })

  return Object.entries(counts).map(([type, count]) => ({ type, count }))
}

const getReactionEmoji = (type) => {
  const emojis = {
    thumbs_up: '👍',
    heart: '❤️',
    celebrate: '🎉'
  }
  return emojis[type] || '👍'
}

const getVisibilityLabel = (targetGroups) => {
  if (!targetGroups || targetGroups.length === 0) {
    return t('events.visibility.public')
  }
  if (targetGroups.length === 1) {
    return targetGroups[0].name
  }
  return t('events.visibility.groupsCount', { count: targetGroups.length })
}

const getGroupsTooltip = (targetGroups) => {
  if (!targetGroups || targetGroups.length === 0) {
    return t('events.visibility.tooltip.public')
  }
  const groupNames = targetGroups.map(g => g.name).join(', ')
  return t('events.visibility.tooltip.groups', { groups: groupNames })
}
</script>

<style scoped>
.ncard {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  height: 100%;
  min-width: 0;
  padding: 14px 16px 12px;
  border: 1px solid var(--line);
  border-top: 3px solid var(--tc);
  border-radius: var(--r-l);
  background: var(--surface);
  box-shadow: 0 1px 2px rgb(var(--shadow-rgb) / 0.06);
  cursor: pointer;
  transition: border-color 150ms ease, box-shadow 150ms ease;
}

.ncard:hover {
  border-color: color-mix(in oklab, var(--tc) 45%, var(--line));
  box-shadow: 0 4px 14px -6px rgb(var(--shadow-rgb) / 0.25);
}

.ncard:focus-within {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.ncard__link:focus-visible { outline: none; }

.ncard__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.nbadge--type { --tag-c: var(--tc); }

.ncard__audience,
.ncard__category,
.ncard__meta {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
  font-size: 12px;
  color: var(--muted);
}

.ncard__audience :deep(svg),
.ncard__category :deep(svg),
.ncard__meta :deep(svg) { flex: none; width: 14px; height: 14px; }

.ncard__audience { max-width: 12em; }

/* La liste des groupes reste lisible au survol, au-dessus du lien étendu */
.ncard__audience--groups { position: relative; z-index: 1; cursor: help; }

.ncard__category { max-width: 14em; margin-inline-start: auto; }

.ncard__ellipsis {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.ncard__title {
  margin: 0;
  font-size: 17px;
  font-weight: 600;
  line-height: 1.3;
  color: var(--text);
}

.ncard__link {
  display: -webkit-box;
  width: 100%;
  overflow: hidden;
  text-align: start;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.ncard__summary {
  display: -webkit-box;
  margin: 0;
  overflow: hidden;
  font-size: 14px;
  line-height: 1.5;
  color: var(--muted);
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

/* Étiquettes : une seule ligne, texte discret, pastille de couleur ; noms longs tronqués */
.ncard__tags {
  display: flex;
  gap: 6px;
  min-width: 0;
  margin: 0;
  padding: 0;
  overflow: hidden;
  list-style: none;
}

.ntag { min-width: 0; max-width: 11em; }
.ntag--more { flex: none; }

.ncard__foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 12px;
  margin-top: auto;
  padding-top: 10px;
  border-top: 1px solid var(--line);
}

.ncard__author { flex: 1 1 8em; }

@media (prefers-reduced-motion: reduce) {
  .ncard { transition: none; }
}
</style>
