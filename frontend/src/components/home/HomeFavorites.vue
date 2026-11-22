<template>
  <section class="home-card" aria-labelledby="home-fav-title">
    <div class="home-card__head">
      <h2 id="home-fav-title" class="home-card__title">
        <Icon icon="lucide:star" class="fav__star" aria-hidden="true" />
        {{ $t('home.favorites.title') }}
      </h2>
      <router-link to="/resources" class="home-card__more">
        {{ $t('home.rail.portal') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </div>
    <ul class="fav__grid">
      <li v-for="app in apps.slice(0, 4)" :key="app.id">
        <a
          :href="app.url"
          class="fav__tile"
          :style="{ '--c': app.color || 'var(--muted)' }"
          :target="app.open_in_new_tab ? '_blank' : undefined"
          rel="noopener noreferrer"
        >
          <span class="fav__icon"><Icon :icon="app.icon || 'lucide:app-window'" aria-hidden="true" /></span>
          <span class="fav__name">{{ app.name }}</span>
        </a>
      </li>
    </ul>
  </section>
</template>

<script setup>
// Ressources favorites (les quatre premières) ; le portail complet est sur /resources.
import { Icon } from '@iconify/vue'

defineProps({
  apps: { type: Array, required: true },
})
</script>

<style scoped>
.fav__star { width: 15px; height: 15px; color: #d99a00; }

.fav__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.fav__tile {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
  padding: 6px 8px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  color: var(--text);
  transition: border-color 120ms ease;
}

.fav__tile:hover { border-color: var(--acc); }

.fav__tile:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.fav__icon {
  flex: none;
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: var(--r-s);
  background: color-mix(in oklab, var(--c) 14%, var(--surface));
  color: color-mix(in oklab, var(--c) 85%, var(--text));
  font-size: 17px;
}

.fav__name {
  min-width: 0;
  font-size: 12.5px;
  font-weight: 500;
  line-height: 1.25;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>
