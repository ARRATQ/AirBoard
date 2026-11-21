<template>
  <header class="home-top">
    <AnnouncementPanel v-if="announcements.length" :announcements="announcements" />
    <div v-else class="home-top__spacer" />

    <form class="home-top__search" role="search" @submit.prevent="goSearch">
      <Icon icon="mdi:magnify" class="home-top__search-icon" aria-hidden="true" />
      <input
        v-model="query"
        type="search"
        class="home-top__search-input"
        :placeholder="$t('home.hero.searchPlaceholder')"
        :aria-label="$t('home.searchLabel')"
        @keydown.esc="query = ''"
      />
      <kbd class="home-top__kbd">{{ shortcut }}</kbd>
    </form>

    <p class="home-top__meta">
      <span v-if="weather.temp !== null" class="home-top__weather" :title="weather.city">
        <Icon :icon="weather.icon" class="home-top__weather-icon" aria-hidden="true" />
        <span class="tabular">{{ weather.temp }}°</span>
        <span v-if="weather.city" class="home-top__city">{{ weather.city }}</span>
      </span>
      <span class="home-top__date">{{ dateLabel }}</span>
    </p>
  </header>
</template>

<script setup>
// Barre du haut de l'accueil : annonces, recherche, météo et date.
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import AnnouncementPanel from '@/components/ui/AnnouncementPanel.vue'
import { dateLocaleFor } from '@/utils/homeFormat'
import { useWeather } from '@/composables/useWeather'

defineProps({
  announcements: { type: Array, default: () => [] },
})

const router = useRouter()
const { locale } = useI18n()

const query = ref('')
const goSearch = () => {
  const q = query.value.trim()
  router.push(q ? `/search?q=${encodeURIComponent(q)}` : '/search')
}

const isMac = typeof navigator !== 'undefined' && /Mac|iPhone|iPad/.test(navigator.platform || '')
const shortcut = isMac ? '⌘K' : 'Ctrl K'

const dateLabel = computed(() =>
  new Date().toLocaleDateString(dateLocaleFor(locale.value), { weekday: 'short', day: 'numeric', month: 'short' }),
)

// ── Météo (mise en cache, voir composables/useWeather.js) ────────────
const { weather, loadWeather } = useWeather()

onMounted(loadWeather)
</script>

<style scoped>
.home-top {
  flex: none;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px 12px;
  padding: 10px 20px;
  border-block-end: 1px solid var(--line);
  background: var(--surface);
}

.home-top__spacer { flex: 1 1 0; }

.home-top__search {
  flex: 0 1 340px;
  min-width: 200px;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 34px;
  padding-inline: 10px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--bg);
  color: var(--muted);
}

.home-top__search:focus-within {
  border-color: var(--acc);
  box-shadow: 0 0 0 2px color-mix(in oklab, var(--acc) 25%, transparent);
}

.home-top__search-icon { flex: none; width: 16px; height: 16px; }

.home-top__search-input {
  flex: 1;
  min-width: 0;
  padding: 0;
  border: 0;
  background: none;
  color: var(--text);
  font: inherit;
  font-size: 13px;
}

.home-top__search-input:focus { outline: none; box-shadow: none; }
.home-top__search-input::placeholder { color: var(--muted); }
.home-top__search-input::-webkit-search-cancel-button { display: none; }

.home-top__kbd {
  flex: none;
  padding: 1px 6px;
  border: 1px solid var(--line);
  border-radius: 3px;
  font-family: var(--f-num);
  font-size: 10.5px;
  color: var(--muted);
}

.home-top__meta {
  flex: none;
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-family: var(--f-num);
  font-size: 12.5px;
  color: var(--muted);
  white-space: nowrap;
  text-transform: uppercase;
}

.home-top__weather { display: inline-flex; align-items: center; gap: 5px; }
.home-top__weather::after { content: '·'; margin-inline-start: 3px; }
.home-top__weather-icon { width: 16px; height: 16px; }

@media (max-width: 1100px) {
  .home-top__city { display: none; }
}

/* Mobile : annonces sur toute la largeur, recherche et météo en dessous. */
@media (max-width: 720px) {
  .home-top { padding: 8px 12px; }
  .home-top__spacer { display: none; }
  .home-top__search { flex: 1 1 0; min-width: 0; }
  .home-top__kbd,
  .home-top__date { display: none; }
  .home-top__weather::after { content: none; }
}
</style>
