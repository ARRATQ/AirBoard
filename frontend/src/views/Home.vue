<template>
  <div class="home">
    <HomeTopbar :announcements="homeData.announcements || []" />

    <div class="home__body">
      <div class="home__inner">
        <div v-if="isLoading" class="home__state" role="status">
          <Icon icon="mdi:loading" class="home__spin" aria-hidden="true" />
          <p>{{ $t('home.loading') }}</p>
        </div>

        <div v-else-if="error" class="home__state" role="alert">
          <EmptyState icon="mdi:alert-circle-outline" :title="$t('home.error.title')" :description="$t('home.error.message')">
            <BaseButton icon="mdi:refresh" @click="loadHomeData">{{ $t('common.retry') }}</BaseButton>
          </EmptyState>
        </div>

        <template v-else>
          <div class="home__head">
            <HomeGreeting
              :first-name="firstName"
              :pending-text="pendingActionsText"
              :active-poll="activePoll"
              :latest-news-group="latestNewsGroup"
              :next-event="nextEvent"
            />
            <HomeXpCard :gamification="homeData.gamification || null" />
          </div>

          <HomeKeyFigures v-if="figures.length" :figures="figures" />

          <div class="home__split">
            <HomeNewsFeed :groups="homeData.recent_news_by_type || []" />
            <div class="home__rail">
              <HomeFavorites v-if="homeData.favorite_apps?.length" :apps="homeData.favorite_apps" />
              <HomePollCard v-if="activePoll" :poll="activePoll" />
              <HomeUpcomingEvents v-if="homeData.upcoming_events?.length" :events="homeData.upcoming_events" />
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
// Accueil : charge homeService.getHomeData() et répartit les données entre les composants de
// components/home/. Hiérarchie A+ : salutation et progression, bandeau des chiffres clés,
// puis le fil d'actualités avec, à côté, favoris, sondage ouvert et prochains événements.
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { homeService } from '@/services/api'
import { keyFigures } from '@/utils/homeFormat'
import EmptyState from '@/components/ui/EmptyState.vue'
import BaseButton from '@/components/ui/BaseButton.vue'
import HomeTopbar from '@/components/home/HomeTopbar.vue'
import HomeGreeting from '@/components/home/HomeGreeting.vue'
import HomeXpCard from '@/components/home/HomeXpCard.vue'
import HomeKeyFigures from '@/components/home/HomeKeyFigures.vue'
import HomeNewsFeed from '@/components/home/HomeNewsFeed.vue'
import HomeFavorites from '@/components/home/HomeFavorites.vue'
import HomePollCard from '@/components/home/HomePollCard.vue'
import HomeUpcomingEvents from '@/components/home/HomeUpcomingEvents.vue'
import '@/components/home/home.css'

const { t } = useI18n()
const authStore = useAuthStore()

const homeData = ref({})
const isLoading = ref(true)
const error = ref(null)

const firstName = computed(() => (authStore.userDisplayName || authStore.user?.username || '').split(' ')[0])

const pendingActionsText = computed(() => {
  const pending = homeData.value.pending_items
  if (!pending) return ''
  const parts = []
  if (pending.unread_news) parts.push(t('home.hero.pendingNews', { count: pending.unread_news }))
  if (pending.unanswered_polls) parts.push(t('home.hero.pendingPolls', { count: pending.unanswered_polls }))
  if (pending.new_apps_count) parts.push(t('home.hero.pendingApps', { count: pending.new_apps_count }))
  return parts.length ? t('home.hero.pendingText', { items: parts.join(t('common.and')) }) : ''
})

const activePoll = computed(() => homeData.value.polls?.find((p) => p.is_active) || null)
const latestNewsGroup = computed(() => homeData.value.recent_news_by_type?.[0] || null)
const nextEvent = computed(() => homeData.value.upcoming_events?.[0] || null)
const figures = computed(() =>
  keyFigures(homeData.value.stats, homeData.value.user_role, homeData.value.managed_group_ids),
)

async function loadHomeData() {
  isLoading.value = true
  error.value = null
  try {
    homeData.value = await homeService.getHomeData()
  } catch (err) {
    console.error('Failed to load home data:', err)
    error.value = err
  } finally {
    isLoading.value = false
  }
}

onMounted(loadHomeData)
</script>

<style scoped>
.home {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg);
  overflow: hidden;
}

.home__body {
  flex: 1;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--line) transparent;
}

.home__inner {
  display: flex;
  flex-direction: column;
  gap: 18px;
  max-width: 1400px;
  margin-inline: auto;
  padding: 22px 24px 32px;
}

.home__state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  min-height: 320px;
  color: var(--muted);
}

.home__spin {
  width: 32px;
  height: 32px;
  animation: home-spin 1s linear infinite;
}

@keyframes home-spin {
  to { transform: rotate(360deg); }
}

@media (prefers-reduced-motion: reduce) {
  .home__spin { animation: none; }
}

/* Salutation à gauche, progression à droite. */
.home__head {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  align-items: start;
  gap: 16px 24px;
}

/* Fil d'actualités dominant, colonne de droite fixe. */
.home__split {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 320px;
  align-items: start;
  gap: 16px;
}

.home__rail {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

@media (max-width: 1024px) {
  .home__split { grid-template-columns: minmax(0, 1fr); }

  .home__rail {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    align-items: start;
  }
}

@media (max-width: 900px) {
  .home__head { grid-template-columns: minmax(0, 1fr); }
}

@media (max-width: 640px) {
  .home__inner { gap: 14px; padding: 16px 16px 24px; }
}
</style>
