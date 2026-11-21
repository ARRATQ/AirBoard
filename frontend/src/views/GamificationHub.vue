<template>
  <div class="content-area xp-page">
    <PageHeader :title="t('gamification.title')" :description="t('gamification.subtitle')" />

    <div v-if="loading" class="xp-state" role="status">
      <Icon icon="mdi:loading" class="xp-state__spin" aria-hidden="true" />
      <span>{{ t('common.loading') }}</span>
    </div>

    <EmptyState
      v-else-if="loadError"
      :title="t('gamification.loadError')"
      icon="mdi:alert-circle-outline"
    >
      <BaseButton variant="secondary" size="sm" icon="mdi:refresh" @click="fetchData">
        {{ t('common.retry') }}
      </BaseButton>
    </EmptyState>

    <template v-else>
      <dl class="xp-figures">
        <div><dt>{{ t('gamification.figRank') }}</dt><dd>{{ myRank ? `#${myRank}` : '—' }}<span class="xp-figures__of"> / {{ leaderboard.length }}</span></dd></div>
        <div><dt>{{ t('gamification.figBadges') }}</dt><dd>{{ unlockedAchievements.length }}<span class="xp-figures__of"> / {{ achievements.length }}</span></dd></div>
        <div><dt>{{ t('gamification.streak') }}</dt><dd>{{ profile.login_streak || 0 }}</dd></div>
        <div><dt>{{ t('gamification.figPlayers') }}</dt><dd>{{ leaderboard.length }}</dd></div>
      </dl>

      <div class="xp-grid">
        <!-- Profil et progression -->
        <section class="xp-panel xp-profile" aria-labelledby="xp-profile-title">
          <div class="xp-profile__id">
            <span class="xp-avatar" aria-hidden="true">
              <img v-if="authStore.user?.avatar_url" :src="authStore.user.avatar_url" alt="" />
              <template v-else>{{ authStore.userInitials }}</template>
            </span>
            <div class="xp-profile__names">
              <h2 id="xp-profile-title" class="xp-profile__name">{{ displayName(authStore.user) }}</h2>
              <p class="xp-profile__user">@{{ authStore.user?.username }}</p>
            </div>
          </div>

          <div class="xp-progress">
            <div class="xp-progress__row">
              <span class="xp-progress__level">{{ t('gamification.level', { level: profile.level }) }}</span>
              <span class="xp-progress__xp">{{ fmt(profile.xp) }} XP</span>
            </div>
            <div
              class="xp-progress__track"
              role="progressbar"
              :aria-valuenow="Math.round(progressPercentage)"
              aria-valuemin="0"
              aria-valuemax="100"
              :aria-label="t('gamification.xpToNext', { xp: fmt(xpToNextLevel), level: profile.level + 1 })"
            >
              <span class="xp-progress__fill" :style="{ inlineSize: progressPercentage + '%' }" />
            </div>
            <p class="xp-progress__next">{{ t('gamification.xpToNext', { xp: fmt(xpToNextLevel), level: profile.level + 1 }) }}</p>
          </div>

          <div v-if="unlockedAchievements.length" class="xp-latest">
            <h3 class="xp-eyebrow">{{ t('gamification.latestBadges') }}</h3>
            <ul class="xp-latest__list">
              <li
                v-for="ua in unlockedAchievements.slice(0, 6)"
                :key="ua.id"
                class="xp-swatch"
                :style="{ background: ua.achievement?.color || 'var(--fill)' }"
                :title="ua.achievement?.name"
              >
                <Icon :icon="ua.achievement?.icon || 'mdi:medal'" aria-hidden="true" />
                <span class="sr-only">{{ ua.achievement?.name }}</span>
              </li>
            </ul>
          </div>
        </section>

        <!-- Activité récente -->
        <section class="xp-panel xp-activity" aria-labelledby="xp-activity-title">
          <h2 id="xp-activity-title" class="xp-eyebrow">{{ t('gamification.recentActivity') }}</h2>
          <ol v-if="recentTransactions.length" class="xp-activity__list">
            <li v-for="tx in recentTransactions.slice(0, 8)" :key="tx.id" class="xp-activity__item">
              <Icon :icon="activityIcon(tx.reason)" class="xp-activity__icon" aria-hidden="true" />
              <span class="xp-activity__label">{{ activityLabel(tx.reason) }}</span>
              <time class="xp-activity__time" :datetime="tx.created_at">{{ relativeDay(tx.created_at, dateLocale) }}</time>
              <span class="xp-activity__xp">+{{ fmt(tx.amount) }}</span>
            </li>
          </ol>
          <p v-else class="xp-muted">{{ t('gamification.noActivity') }}</p>
        </section>

        <!-- Badges -->
        <section class="xp-panel xp-badges" aria-labelledby="xp-badges-title">
          <div class="xp-panel__head">
            <h2 id="xp-badges-title" class="xp-eyebrow">{{ t('gamification.badges') }}</h2>
            <SegmentedControl v-model="currentTab" :options="tabOptions" :label="t('gamification.badgeFilter')" />
          </div>
          <ul v-if="filteredAchievements.length" class="xp-badges__list">
            <li
              v-for="achievement in filteredAchievements"
              :key="achievement.id"
              class="xp-badge"
              :class="{ 'is-locked': !isUnlocked(achievement.code) }"
            >
              <span
                class="xp-badge__icon"
                :style="isUnlocked(achievement.code) ? { background: achievement.color || 'var(--fill)' } : undefined"
                aria-hidden="true"
              >
                <Icon :icon="isUnlocked(achievement.code) ? achievement.icon : 'mdi:lock-outline'" />
              </span>
              <div class="xp-badge__body">
                <h3 class="xp-badge__name">{{ achievement.name }}</h3>
                <p class="xp-badge__desc">{{ achievement.description }}</p>
                <p class="xp-badge__meta">
                  <template v-if="isUnlocked(achievement.code)">{{ t('gamification.unlockedAt', { date: unlockedDate(achievement.code) }) }}</template>
                  <template v-else><span class="sr-only">{{ t('gamification.locked') }} · </span>{{ fmt(achievement.xp_reward) }} XP</template>
                </p>
              </div>
            </li>
          </ul>
          <EmptyState v-else compact :title="t('gamification.noBadges')" icon="mdi:shield-outline" />
        </section>

        <!-- Classement -->
        <section class="xp-panel xp-board" aria-labelledby="xp-board-title">
          <h2 id="xp-board-title" class="xp-eyebrow">{{ t('gamification.leaderboard') }}</h2>
          <ol class="xp-board__list">
            <li
              v-for="(user, index) in leaderboard"
              :key="user.user_id"
              class="xp-board__row"
              :class="{ 'is-me': user.user_id === authStore.user?.id }"
              :aria-current="user.user_id === authStore.user?.id ? 'true' : undefined"
            >
              <span class="xp-board__rank">{{ index + 1 }}</span>
              <div class="xp-board__who">
                <span class="xp-board__name">
                  {{ displayName(user) }}
                  <span v-if="user.user_id === authStore.user?.id" class="xp-board__you">({{ t('gamification.you') }})</span>
                </span>
                <span class="xp-board__user">@{{ user.username }}</span>
                <ul v-if="user.badges?.length" class="xp-board__badges">
                  <li
                    v-for="badge in user.badges"
                    :key="badge.code"
                    class="xp-swatch xp-swatch--sm"
                    :style="{ background: badge.color || 'var(--fill)' }"
                    :title="badge.name"
                  >
                    <Icon :icon="badge.icon" aria-hidden="true" />
                    <span class="sr-only">{{ badge.name }}</span>
                  </li>
                </ul>
              </div>
              <div class="xp-board__score">
                <span class="xp-board__xp">{{ fmt(user.xp) }} XP</span>
                <span class="xp-board__level">{{ t('gamification.levelShort', { level: user.level }) }}</span>
              </div>
            </li>
          </ol>
        </section>
      </div>
    </template>
  </div>
</template>

<script setup>
// My XP : niveau et progression, activité récente, badges (filtrés), classement.
// États de chargement et d'erreur (avec « Réessayer ») autour des cinq appels de l'API.
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { gamificationService } from '@/services/api'
import { dateLocaleFor, relativeDay } from '@/utils/homeFormat'
import PageHeader from '@/components/ui/PageHeader.vue'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import SegmentedControl from '@/components/ui/SegmentedControl.vue'

const { t, locale } = useI18n()
const authStore = useAuthStore()

const loading = ref(true)
const loadError = ref(false)
const profile = ref({ level: 1, xp: 0 })
const achievements = ref([])
const unlockedAchievements = ref([])
const leaderboard = ref([])
const recentTransactions = ref([])
const currentTab = ref('all')

const ACTIVITY_ICONS = {
  daily_login: 'mdi:calendar-check',
  app_click: 'mdi:cursor-default-click-outline',
  news_read: 'mdi:book-open-variant',
  news_publish: 'mdi:newspaper-plus',
  event_publish: 'mdi:calendar-star',
  poll_vote: 'mdi:vote-outline',
  poll_create: 'mdi:poll',
  comment_create: 'mdi:comment-text-outline',
  suggestion_create: 'mdi:lightbulb-on-outline',
  suggestion_vote: 'mdi:thumb-up-outline',
  suggestion_comment: 'mdi:comment-edit-outline',
  achievement_unlock: 'mdi:medal',
  chat_message: 'mdi:chat-processing-outline',
  chatbot_use: 'mdi:robot-excited-outline',
}

const dateLocale = computed(() => dateLocaleFor(locale.value))
const fmt = (n) => Number(n || 0).toLocaleString(dateLocale.value)
const displayName = (u) => [u?.first_name, u?.last_name].filter(Boolean).join(' ') || u?.username || ''

// Seuils de niveau : (niveau - 1)² × 100 XP, comme le calcul du serveur.
const progressPercentage = computed(() => {
  const { xp, level } = profile.value
  const from = Math.pow(level - 1, 2) * 100
  const range = Math.pow(level, 2) * 100 - from
  if (range === 0) return 100
  return Math.min(Math.max(((xp - from) / range) * 100, 0), 100)
})

const xpToNextLevel = computed(() => Math.max(Math.pow(profile.value.level, 2) * 100 - profile.value.xp, 0))
const myRank = computed(() => profile.value.rank || null)

const unlockedCodes = computed(() => new Set(unlockedAchievements.value.map((ua) => ua.achievement?.code)))
const isUnlocked = (code) => unlockedCodes.value.has(code)

const tabOptions = computed(() => [
  { value: 'all', label: t('gamification.all'), count: achievements.value.length },
  { value: 'unlocked', label: t('gamification.unlocked'), count: unlockedAchievements.value.length },
  { value: 'locked', label: t('gamification.lockedPlural'), count: Math.max(achievements.value.length - unlockedAchievements.value.length, 0) },
])

const filteredAchievements = computed(() => {
  if (currentTab.value === 'unlocked') return achievements.value.filter((a) => isUnlocked(a.code))
  if (currentTab.value === 'locked') return achievements.value.filter((a) => !isUnlocked(a.code))
  return achievements.value
})

const unlockedDate = (code) => {
  const ua = unlockedAchievements.value.find((u) => u.achievement?.code === code)
  return ua?.unlocked_at
    ? new Date(ua.unlocked_at).toLocaleDateString(dateLocale.value, { day: '2-digit', month: 'short', year: 'numeric' })
    : ''
}

const activityIcon = (reason) => ACTIVITY_ICONS[reason] || 'mdi:star-outline'
const activityLabel = (reason) =>
  t(ACTIVITY_ICONS[reason] ? `gamification.activity.${reason}` : 'gamification.activity.other')

async function fetchData() {
  loading.value = true
  loadError.value = false
  try {
    const [profileData, allAch, myAch, lb, txs] = await Promise.all([
      gamificationService.getProfile(),
      gamificationService.getAllAchievements(),
      gamificationService.getMyAchievements(),
      gamificationService.getLeaderboard(),
      gamificationService.getTransactions(),
    ])
    profile.value = profileData || { level: 1, xp: 0 }
    achievements.value = allAch || []
    unlockedAchievements.value = myAch || []
    leaderboard.value = lb || []
    recentTransactions.value = txs || []
  } catch (e) {
    console.error('Gamification fetch error', e)
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<style scoped>
.xp-page { max-width: 1400px; }

.xp-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  min-height: 240px;
  color: var(--muted);
}

.xp-state__spin { width: 22px; height: 22px; animation: xp-spin 0.9s linear infinite; }
@keyframes xp-spin { to { transform: rotate(360deg); } }

/* Chiffres clés */
.xp-figures {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin: 0 0 16px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
}

.xp-figures > div { min-width: 0; padding: 10px 14px; border-inline-start: 1px solid var(--line); }
.xp-figures > div:first-child { border-inline-start: 0; }

.xp-figures dt,
.xp-eyebrow {
  font-family: var(--f-num);
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--muted);
}

.xp-figures dd {
  margin: 2px 0 0;
  font-family: var(--f-num);
  font-size: 24px;
  font-weight: 600;
  line-height: 1.1;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

.xp-figures__of { font-size: 13px; font-weight: 400; color: var(--muted); }

/* Grille : profil + activité | badges | classement */
.xp-grid {
  display: grid;
  grid-template-columns: minmax(250px, 280px) minmax(0, 1fr) minmax(280px, 340px);
  grid-template-areas:
    'profile badges board'
    'activity badges board';
  grid-template-rows: auto 1fr;
  align-items: start;
  gap: 16px;
}

.xp-profile { grid-area: profile; }
.xp-activity { grid-area: activity; }
.xp-badges { grid-area: badges; }
.xp-board { grid-area: board; }

.xp-panel {
  min-width: 0;
  padding: 16px;
  border: 1px solid var(--line);
  border-radius: 4px;
  background: var(--surface);
}

.xp-panel__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 8px 12px;
  margin-block-end: 12px;
}

.xp-eyebrow { margin: 0 0 10px; }
.xp-panel__head .xp-eyebrow { margin: 0; }
.xp-muted { margin: 0; font-size: 13px; color: var(--muted); }

/* Profil */
.xp-profile__id { display: flex; align-items: center; gap: 12px; margin-block-end: 16px; }

.xp-avatar {
  display: grid;
  place-items: center;
  flex: none;
  width: 48px;
  height: 48px;
  overflow: hidden;
  border-radius: 50%;
  background: var(--sunken);
  border: 1px solid var(--line);
  font-family: var(--f-num);
  font-size: 16px;
  font-weight: 600;
  color: var(--text);
}

.xp-avatar img { width: 100%; height: 100%; object-fit: cover; }

.xp-profile__names { min-width: 0; }
.xp-profile__name { margin: 0; font-size: 16px; font-weight: 600; color: var(--text); overflow-wrap: anywhere; }
.xp-profile__user { margin: 0; font-size: 12.5px; color: var(--muted); overflow-wrap: anywhere; }

.xp-progress__row { display: flex; justify-content: space-between; gap: 8px; margin-block-end: 6px; }
.xp-progress__level { font-size: 13px; font-weight: 600; color: var(--text); }
.xp-progress__xp { font-family: var(--f-num); font-size: 13px; color: var(--text); font-variant-numeric: tabular-nums; }

.xp-progress__track {
  display: block;
  height: 6px;
  overflow: hidden;
  border-radius: 3px;
  background: var(--sunken);
}

.xp-progress__fill {
  display: block;
  height: 100%;
  border-radius: 3px;
  background: var(--acc);
  transition: inline-size 600ms ease;
}

.xp-progress__next { margin: 6px 0 0; font-size: 12px; color: var(--muted); }

.xp-latest { margin-block-start: 16px; padding-block-start: 14px; border-block-start: 1px solid var(--line); }
.xp-latest__list,
.xp-board__badges { display: flex; flex-wrap: wrap; gap: 6px; margin: 0; padding: 0; list-style: none; }

.xp-swatch {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  color: #fff;
  font-size: 15px;
}

.xp-swatch--sm { width: 18px; height: 18px; border-radius: 3px; font-size: 11px; }

/* Activité */
.xp-activity__list { margin: 0; padding: 0; list-style: none; }

.xp-activity__item {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr) auto;
  grid-template-areas: 'icon label xp' 'icon time xp';
  column-gap: 10px;
  align-items: center;
  padding-block: 8px;
  border-block-start: 1px solid var(--line);
}

.xp-activity__item:first-child { border-block-start: 0; padding-block-start: 0; }
.xp-activity__icon { grid-area: icon; width: 16px; height: 16px; color: var(--muted); }
.xp-activity__label { grid-area: label; font-size: 13px; color: var(--text); }
.xp-activity__time { grid-area: time; font-size: 11.5px; color: var(--muted); }

.xp-activity__xp {
  grid-area: xp;
  font-family: var(--f-num);
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text);
  font-variant-numeric: tabular-nums;
}

/* Badges */
.xp-badges__list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.xp-badge {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 10px;
  border: 1px solid var(--line);
  border-radius: 4px;
}

.xp-badge__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 36px;
  height: 36px;
  border-radius: 4px;
  color: #fff;
  font-size: 19px;
}

.xp-badge.is-locked { border-style: dashed; }
.xp-badge.is-locked .xp-badge__icon { background: var(--sunken); color: var(--muted); font-size: 16px; }
.xp-badge.is-locked .xp-badge__name { color: var(--muted); }

.xp-badge__body { min-width: 0; }
.xp-badge__name { margin: 0; font-size: 13.5px; font-weight: 600; color: var(--text); }
.xp-badge__desc { margin: 2px 0 0; font-size: 12px; line-height: 1.4; color: var(--muted); }

.xp-badge__meta {
  margin: 6px 0 0;
  font-family: var(--f-num);
  font-size: 11px;
  letter-spacing: 0.02em;
  color: var(--muted);
}

.xp-badge:not(.is-locked) .xp-badge__meta { color: var(--ok); }

/* Classement */
.xp-board__list { margin: 0; padding: 0; list-style: none; }

.xp-board__row {
  display: grid;
  grid-template-columns: 26px minmax(0, 1fr) auto;
  gap: 10px;
  align-items: start;
  padding: 10px 8px;
  border-block-start: 1px solid var(--line);
}

.xp-board__row:first-child { border-block-start: 0; }

.xp-board__row.is-me {
  border-radius: 3px;
  background: var(--sunken);
  box-shadow: inset 3px 0 0 var(--acc);
}

[dir='rtl'] .xp-board__row.is-me { box-shadow: inset -3px 0 0 var(--acc); }

.xp-board__rank {
  font-family: var(--f-num);
  font-size: 13px;
  font-weight: 600;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}

.xp-board__row:nth-child(-n + 3) .xp-board__rank { color: var(--text); }

.xp-board__who { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.xp-board__name { font-size: 13px; font-weight: 600; color: var(--text); overflow-wrap: anywhere; }
.xp-board__you { font-weight: 400; color: var(--muted); }
.xp-board__user { font-size: 11.5px; color: var(--muted); overflow-wrap: anywhere; }
.xp-board__badges { margin-block-start: 4px; gap: 4px; }

.xp-board__score { display: flex; flex-direction: column; align-items: flex-end; gap: 2px; }

.xp-board__xp {
  font-family: var(--f-num);
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.xp-board__level { font-family: var(--f-num); font-size: 11px; color: var(--muted); white-space: nowrap; }

@media (max-width: 1280px) {
  .xp-grid {
    grid-template-columns: minmax(240px, 280px) minmax(0, 1fr);
    grid-template-areas:
      'profile badges'
      'activity badges'
      'board board';
  }
}

@media (max-width: 768px) {
  .xp-grid {
    grid-template-columns: minmax(0, 1fr);
    grid-template-areas: 'profile' 'activity' 'badges' 'board';
  }

  .xp-figures { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .xp-figures > div:nth-child(3) { border-inline-start: 0; }
  .xp-figures > div:nth-child(n + 3) { border-block-start: 1px solid var(--line); }
  .xp-figures > div { padding: 8px 10px; }
  .xp-figures dd { font-size: 20px; }
  .xp-panel { padding: 14px; }
}

@media (prefers-reduced-motion: reduce) {
  .xp-progress__fill { transition: none; }
  .xp-state__spin { animation: none; }
}
</style>
