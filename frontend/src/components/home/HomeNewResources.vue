<template>
  <section class="home-card" aria-labelledby="home-newres-title">
    <div class="home-card__head">
      <h2 id="home-newres-title" class="home-card__title">{{ $t('home.newApps.title') }}</h2>
      <router-link to="/resources" class="home-card__more">
        {{ $t('home.rail.portal') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </div>

    <ul v-if="apps.length" class="nr__list">
      <li v-for="app in apps.slice(0, MAX_SHOWN)" :key="app.id">
        <a
          :href="app.url"
          class="nr__row"
          :style="{ '--c': app.color || 'var(--muted)' }"
          :target="app.open_in_new_tab ? '_blank' : undefined"
          rel="noopener noreferrer"
        >
          <span class="nr__icon"><Icon :icon="app.icon || 'lucide:app-window'" aria-hidden="true" /></span>
          <span class="nr__body">
            <span class="nr__name" dir="auto">{{ app.name }}</span>
            <span class="nr__meta">
              <span v-if="app.app_group?.name" class="nr__group" dir="auto">{{ app.app_group.name }}</span>
              <time :datetime="app.created_at">{{ relativeDay(app.created_at, dateLocale) }}</time>
            </span>
          </span>
          <span v-if="isNewResource(app)" class="tag tag--acc">{{ $t('home.newApps.newBadge') }}</span>
        </a>
      </li>
    </ul>
    <p v-else class="nr__empty">{{ $t('home.newApps.empty') }}</p>
  </section>
</template>

<script setup>
// Dernières ressources ajoutées (liens vers des outils, pas des services surveillés).
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { dateLocaleFor, relativeDay } from '@/utils/homeFormat'
import { isNewResource } from '@/utils/resourceType'

defineProps({
  apps: { type: Array, required: true },
})

const MAX_SHOWN = 4
const { locale } = useI18n()
const dateLocale = computed(() => dateLocaleFor(locale.value))
</script>

<style scoped>
.nr__list { margin: 0; padding: 0; list-style: none; }

.nr__row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-height: 48px;
  padding: 6px 4px;
  border-block-end: 1px solid var(--line);
  color: var(--text);
}

.nr__list li:last-child .nr__row { border-block-end: 0; }
.nr__row:hover { background: var(--sunken); }

.nr__row:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: -2px;
  border-radius: var(--r-s);
}

.nr__icon {
  flex: none;
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: var(--r-s);
  background: color-mix(in oklab, var(--c) 14%, var(--surface));
  color: var(--c);
}

.nr__icon :deep(svg) { width: 18px; height: 18px; }

.nr__body { flex: 1; min-width: 0; display: flex; flex-direction: column; }

.nr__name,
.nr__group {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.nr__name { font-size: 13px; font-weight: 500; line-height: 1.3; }

.nr__meta {
  display: flex;
  gap: 8px;
  min-width: 0;
  font-size: 11.5px;
  color: var(--muted);
}

.nr__meta time { flex: none; }


.nr__empty { margin: 0; padding: 8px 4px; font-size: 13px; color: var(--muted); }
</style>
