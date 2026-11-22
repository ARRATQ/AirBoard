<template>
  <section class="home-card xp" aria-labelledby="home-xp-title">
    <div class="home-card__head">
      <h2 id="home-xp-title" class="home-card__title">{{ $t('home.xp.title') }}</h2>
      <router-link to="/gamification" class="home-card__more">
        {{ $t('common.details') }} <span class="home-card__arrow" aria-hidden="true">→</span>
      </router-link>
    </div>

    <div v-if="gamification" class="xp__level">
      <span class="xp__level-box">
        <span class="xp__level-label">{{ $t('home.xp.levelShort') }}</span>
        <span class="xp__level-num">{{ gamification.level }}</span>
      </span>
      <div class="xp__progress">
        <p class="xp__score">
          <span class="tabular">{{ formatNumber(gamification.xp) }}</span>
          <span class="xp__unit">XP</span>
        </p>
        <div
          class="xp__bar"
          role="progressbar"
          :aria-valuenow="progress"
          aria-valuemin="0"
          aria-valuemax="100"
          :aria-label="$t('gamification.level', { level: gamification.level + 1 })"
        >
          <span class="xp__bar-fill" :style="{ inlineSize: progress + '%' }" />
        </div>
        <p class="xp__next">
          {{ formatNumber(Math.max(0, gamification.next_level_xp - gamification.xp)) }} XP
          {{ $t('home.xp.beforeLevel', { level: gamification.level + 1 }) }}
        </p>
      </div>
    </div>

    <dl class="xp__stats">
      <div class="xp__stat">
        <dt>{{ $t('home.xp.streak') }}</dt>
        <dd class="tabular">{{ gamification?.streak_days || '–' }}</dd>
      </div>
      <div class="xp__stat">
        <dt>{{ $t('home.xp.badges') }}</dt>
        <dd class="tabular">{{ gamification?.badges_count || gamification?.recent_badges?.length || 0 }}</dd>
      </div>
      <div class="xp__stat">
        <dt>{{ $t('home.xp.rank') }}</dt>
        <dd class="tabular">{{ gamification?.rank ? '#' + gamification.rank : '–' }}</dd>
      </div>
    </dl>

    <ul v-if="gamification?.recent_badges?.length" class="xp__badges">
      <li
        v-for="badge in gamification.recent_badges.slice(0, 5)"
        :key="badge.id"
        class="xp__badge"
        :style="{ '--c': badge.color || 'var(--muted)' }"
        :title="badge.name"
      >
        <Icon :icon="badge.icon" class="xp__badge-icon" aria-hidden="true" />
        <span class="sr-only">{{ badge.name }}</span>
      </li>
    </ul>
  </section>
</template>

<script setup>
// Progression de l'utilisateur sur l'accueil (gamification) : niveau, XP, série, badges, rang.
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { dateLocaleFor } from '@/utils/homeFormat'

const props = defineProps({
  gamification: { type: Object, default: null },
})

const { locale } = useI18n()
const progress = computed(() => Math.min(100, Math.max(0, Math.round(props.gamification?.progress_percent || 0))))
const formatNumber = (n) => Number(n || 0).toLocaleString(dateLocaleFor(locale.value))
</script>

<style scoped>
.xp__level {
  display: flex;
  align-items: center;
  gap: 14px;
}

.xp__level-box {
  flex: none;
  display: grid;
  place-items: center;
  width: 54px;
  height: 54px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
}

.xp__level-label {
  font-family: var(--f-num);
  font-size: 9.5px;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: var(--muted);
}

.xp__level-num {
  margin-block-start: -6px;
  font-family: var(--f-num);
  font-size: 22px;
  font-weight: 600;
  color: var(--acc-ink);
}

.xp__progress { flex: 1; min-width: 0; }

.xp__score {
  margin: 0;
  font-family: var(--f-num);
  font-size: 20px;
  font-weight: 600;
  color: var(--text);
}

.xp__unit {
  margin-inline-start: 4px;
  font-size: 11px;
  font-weight: 500;
  color: var(--muted);
}

.xp__bar {
  height: 6px;
  margin-block: 6px 5px;
  border-radius: var(--r-s);
  background: var(--sunken);
  box-shadow: inset 0 0 0 1px var(--line);
  overflow: hidden;
}

.xp__bar-fill {
  display: block;
  height: 100%;
  border-radius: var(--r-s);
  background: var(--fill);
}

.xp__next {
  margin: 0;
  font-size: 12px;
  color: var(--muted);
}

.xp__stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  margin: 14px 0 0;
  border-block-start: 1px solid var(--line);
}

.xp__stat {
  padding: 10px 0 0;
  text-align: center;
}

.xp__stat + .xp__stat { border-inline-start: 1px solid var(--line); }

.xp__stat dt {
  font-family: var(--f-num);
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.xp__stat dd {
  margin: 2px 0 0;
  font-family: var(--f-num);
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.xp__badges {
  display: flex;
  gap: 6px;
  margin-block-start: 12px;
}

.xp__badge {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  border: 1px solid currentColor;
  border-radius: var(--r-s);
  color: color-mix(in oklab, var(--c) 78%, var(--text));
}

.xp__badge-icon { width: 17px; height: 17px; }
</style>
