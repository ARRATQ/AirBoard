<template>
  <section class="home-card feed" aria-labelledby="home-feed-title">
    <div class="home-card__head">
      <h2 id="home-feed-title" class="home-card__title">{{ $t('home.newsHub') }}</h2>
      <router-link to="/news" class="home-card__more">
        {{ $t('home.viewAll') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </div>

    <div v-if="tabs.length > 1" class="feed__tabs" role="tablist" :aria-label="$t('home.newsHub')">
      <button
        v-for="tab in tabs"
        :id="`feed-tab-${tab.key}`"
        :key="tab.key"
        ref="tabEls"
        type="button"
        role="tab"
        class="feed__tab"
        :class="{ 'feed__tab--on': activeTab === tab.key }"
        :aria-selected="activeTab === tab.key"
        :aria-controls="`feed-panel-${tab.key}`"
        :tabindex="activeTab === tab.key ? 0 : -1"
        @click="selectTab(tab.key)"
        @keydown="onTabKey"
      >
        <Icon :icon="tab.icon || 'lucide:newspaper'" class="feed__tab-icon" aria-hidden="true" />
        {{ tab.label }}
        <span v-if="tab.newCount > 0" class="feed__tab-count">{{ tab.newCount }}</span>
      </button>
    </div>

    <div
      :id="activeTab ? `feed-panel-${activeTab}` : undefined"
      :role="tabs.length > 1 ? 'tabpanel' : undefined"
      :aria-labelledby="tabs.length > 1 ? `feed-tab-${activeTab}` : undefined"
      class="feed__panel"
    >
      <ul v-if="items.length" class="feed__list">
        <li v-for="article in items" :key="article.id">
          <router-link :to="`/news/${article.slug}`" class="feed__row" :style="{ '--c': colorOf(article) }">
            <span class="feed__icon">
              <Icon :icon="article.news_type?.icon || activeGroup?.type.icon || article.category?.icon || 'lucide:newspaper'" aria-hidden="true" />
            </span>
            <span class="feed__body">
              <span class="feed__title">{{ article.title }}</span>
              <span class="feed__meta">
                <span class="feed__type">{{ article.news_type?.name || activeGroup?.type.name || article.category?.name || $t('news.types.article') }}</span>
                <span v-if="authorOf(article)">{{ authorOf(article) }}</span>
                <time class="feed__date" :datetime="article.published_at || article.created_at">
                  {{ relativeDay(article.published_at || article.created_at, dateLocale) }}
                </time>
              </span>
            </span>
            <Icon icon="lucide:chevron-right" class="feed__chev" aria-hidden="true" />
          </router-link>
        </li>
      </ul>
      <EmptyState v-else compact icon="lucide:newspaper" :title="$t('home.noNews')" />
    </div>
  </section>
</template>

<script setup>
// Fil d'actualités de l'accueil, un onglet par type d'article. Le nombre de nouveautés par onglet
// se calcule depuis la dernière visite de l'onglet (localStorage, clé airboard_tab_last_seen).
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import { dateLocaleFor, relativeDay } from '@/utils/homeFormat'

const props = defineProps({
  // homeData.recent_news_by_type : [{ type: { slug, name, icon }, news: [...] }]
  groups: { type: Array, default: () => [] },
})

const STORAGE_KEY = 'airboard_tab_last_seen'
const FALLBACK_COLOR = '#64748b'

const { locale } = useI18n()
const dateLocale = computed(() => dateLocaleFor(locale.value))

function readLastSeen() {
  try {
    return JSON.parse(localStorage.getItem(STORAGE_KEY) || '{}') || {}
  } catch {
    return {}
  }
}

const lastSeen = ref(readLastSeen())
const activeTab = ref(null)
const tabEls = ref([])

// Un type d'article peut avoir un slug vide en base : clé de secours par position.
const keyOf = (group, index) => group.type.slug || `type-${index}`
const keyed = computed(() => props.groups.map((g, i) => ({ ...g, key: keyOf(g, i) })))
const activeGroup = computed(() => keyed.value.find((g) => g.key === activeTab.value) || null)

function markSeen(slug) {
  lastSeen.value = { ...lastSeen.value, [slug]: new Date().toISOString() }
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(lastSeen.value))
  } catch {
    /* stockage indisponible : les compteurs repartent à chaque visite */
  }
}

const publishedAt = (n) => new Date(n.published_at || n.created_at)

// Un onglet ouvert est marqué comme vu : son compteur retombe à zéro.
const tabs = computed(() =>
  keyed.value.map((g) => {
    const seen = lastSeen.value[g.key]
    const newCount = seen ? g.news.filter((n) => publishedAt(n) > new Date(seen)).length : g.news.length
    return { key: g.key, label: g.type.name, icon: g.type.icon, newCount }
  }),
)

const items = computed(() => activeGroup.value?.news || [])

function selectTab(key) {
  activeTab.value = key
  markSeen(key)
}

watch(
  keyed,
  (groups) => {
    if (groups.length && !groups.some((g) => g.key === activeTab.value)) selectTab(groups[0].key)
  },
  { immediate: true },
)

// Flèches gauche/droite entre onglets (inversées en RTL), Début/Fin.
function onTabKey(event) {
  const keys = tabs.value.map((t) => t.key)
  const i = keys.indexOf(activeTab.value)
  const rtl = document.documentElement.dir === 'rtl'
  let next = null
  if (event.key === 'ArrowRight') next = i + (rtl ? -1 : 1)
  else if (event.key === 'ArrowLeft') next = i + (rtl ? 1 : -1)
  else if (event.key === 'Home') next = 0
  else if (event.key === 'End') next = keys.length - 1
  if (next === null) return
  event.preventDefault()
  const target = (next + keys.length) % keys.length
  selectTab(keys[target])
  nextTick(() => tabEls.value[target]?.focus())
}

const colorOf = (a) => a.news_type?.color || activeGroup.value?.type.color || a.category?.color || FALLBACK_COLOR
const authorOf = (a) => a.author?.username || a.author_name || ''
</script>

<style scoped>
.feed__tabs {
  display: flex;
  gap: 2px;
  margin-block-end: 6px;
  padding: 2px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
  overflow-x: auto;
  scrollbar-width: none;
}

.feed__tab {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding-inline: 10px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  font: inherit;
  font-size: 12.5px;
  white-space: nowrap;
  cursor: pointer;
}

.feed__tab:hover { color: var(--text); }

.feed__tab--on {
  background: var(--surface);
  color: var(--text);
  font-weight: 600;
  box-shadow: 0 0 0 1px var(--line);
}

.feed__tab:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: -2px;
}

.feed__tab-icon { width: 15px; height: 15px; }

.feed__tab-count {
  min-width: 18px;
  padding: 0 5px;
  border-radius: var(--r-s);
  background: var(--fill);
  color: var(--on-fill);
  font-family: var(--f-num);
  font-size: 10.5px;
  line-height: 16px;
  text-align: center;
}

.feed__list { margin: 0; padding: 0; list-style: none; }

.feed__row {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 56px;
  padding: 8px 6px;
  border-block-end: 1px solid var(--line);
  color: var(--text);
}

.feed__list li:last-child .feed__row { border-block-end: 0; }
.feed__row:hover { background: var(--sunken); }

.feed__row:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: -2px;
  border-radius: var(--r-s);
}

.feed__icon {
  flex: none;
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border-radius: var(--r-s);
  background: color-mix(in oklab, var(--c) 14%, var(--surface));
  color: color-mix(in oklab, var(--c) 85%, var(--text));
  font-size: 18px;
}

.feed__body { flex: 1; min-width: 0; }

.feed__title {
  display: block;
  font-weight: 600;
  line-height: 1.35;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.feed__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px 10px;
  margin-block-start: 2px;
  font-size: 12px;
  color: var(--muted);
}

.feed__type {
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.feed__date { font-family: var(--f-num); font-size: 11.5px; }

.feed__chev { flex: none; width: 18px; height: 18px; color: var(--muted); }
[dir='rtl'] .feed__chev { transform: scaleX(-1); }
</style>
