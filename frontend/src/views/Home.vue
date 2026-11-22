<template>
  <div class="home">
    <HomeTopbar :announcements="homeData.announcements || []" />

    <div class="home__body">
      <div class="home__inner">
        <div v-if="isLoading" class="home__state" role="status">
          <Icon icon="lucide:loader-circle" class="home__spin" aria-hidden="true" />
          <p>{{ $t('home.loading') }}</p>
        </div>

        <div v-else-if="error" class="home__state" role="alert">
          <EmptyState icon="lucide:circle-alert" :title="$t('home.error.title')" :description="$t('home.error.message')">
            <BaseButton icon="lucide:refresh-cw" @click="loadHomeData">{{ $t('common.retry') }}</BaseButton>
          </EmptyState>
        </div>

        <template v-else>
          <!-- Deux colonnes : à gauche ce qu'on lit (salutation, chiffres, à la une, nouveautés et
               agenda, actualités), à droite
               les outils et rappels. Sur écran étroit, une seule colonne dans l'ordre d'usage. -->
          <div class="home__layout">
            <div class="home__main">
              <HomeGreeting
                class="home__greeting"
                :first-name="firstName"
                :pending-text="pendingActionsText"
                :active-poll="activePoll"
                :latest-news-group="latestNewsGroup"
                :next-event="nextEvent"
              />
              <HomeKeyFigures v-if="figures.length" class="home__figures" :figures="figures" />
              <HomeFeatured v-if="featuredNews.length" class="home__featured" :articles="featuredNews" :types="newsTypes" />
              <div class="home__duo">
                <HomeNewResources :apps="homeData.new_apps || []" />
                <HomeUpcomingEvents :events="agenda" />
              </div>
              <HomeNewsFeed class="home__news" :groups="homeData.recent_news_by_type || []" />
            </div>

            <aside class="home__rail" :aria-label="$t('home.rail.label')">
              <HomeXpCard class="home__xp" :gamification="homeData.gamification || null" />
              <HomeFavorites v-if="homeData.favorite_apps?.length" class="home__favorites" :apps="homeData.favorite_apps" />
              <HomeWorkingDays class="home__workdays" />
              <HomePollCard v-if="activePoll" class="home__poll" :poll="activePoll" />
            </aside>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
// Accueil : charge homeService.getHomeData() et répartit les données entre les composants de
// components/home/. Colonne principale : salutation, chiffres clés, deux derniers articles,
// ressources récentes et agenda, fil d'actualités ; colonne de droite : progression, favoris,
// calculateur de jours ouvrés, sondage ouvert.
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { homeService } from '@/services/api'
import { keyFigures } from '@/utils/homeFormat'
import { agendaAhead, latestNews } from '@/utils/homeHighlights'
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
import HomeWorkingDays from '@/components/home/HomeWorkingDays.vue'
import HomeFeatured from '@/components/home/HomeFeatured.vue'
import HomeNewResources from '@/components/home/HomeNewResources.vue'
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
  if (pending.unread_news) parts.push(t('home.hero.pendingNews', { count: pending.unread_news }, pending.unread_news))
  if (pending.unanswered_polls) parts.push(t('home.hero.pendingPolls', { count: pending.unanswered_polls }, pending.unanswered_polls))
  if (pending.new_apps_count) parts.push(t('home.hero.pendingApps', { count: pending.new_apps_count }, pending.new_apps_count))
  return parts.length ? t('home.hero.pendingText', { items: parts.join(t('common.and')) }) : ''
})

const activePoll = computed(() => homeData.value.polls?.find((p) => p.is_active) || null)
const latestNewsGroup = computed(() => homeData.value.recent_news_by_type?.[0] || null)
const nextEvent = computed(() => homeData.value.upcoming_events?.[0] || null)

const FEATURED_COUNT = 2
const AGENDA_COUNT = 4
const featuredNews = computed(() => latestNews(homeData.value.recent_news, FEATURED_COUNT))
const newsTypes = computed(() => (homeData.value.recent_news_by_type || []).map((g) => g.type))
const agenda = computed(() =>
  agendaAhead(homeData.value.today_events, homeData.value.upcoming_events, AGENDA_COUNT),
)
const figures = computed(() =>
  keyFigures(homeData.value.stats, homeData.value.user_role),
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
  container-type: inline-size;
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

.home__layout {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 340px;
  align-items: start;
  gap: 20px;
}

.home__main,
.home__rail {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-width: 0;
}

/* Ressources récentes et agenda côte à côte, l'un sous l'autre quand la colonne rétrécit. */
.home__duo {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  align-items: start;
  gap: 16px;
}

/* Les colonnes suivent la largeur de la zone de contenu (la barre latérale peut être
   repliée ou masquée), pas celle de la fenêtre. */
@container (max-width: 1000px) {
  .home__layout { grid-template-columns: minmax(0, 1fr) 300px; }
}

/* Zone étroite : les deux colonnes se fondent en une seule, dans l'ordre d'usage. */
@container (max-width: 760px) {
  .home__layout { display: flex; flex-direction: column; align-items: stretch; gap: 16px; }
  .home__main,
  .home__rail { display: contents; }

  .home__greeting { order: 1; }
  .home__xp { order: 2; }
  .home__figures { order: 3; }
  .home__featured { order: 4; }
  .home__favorites { order: 5; }
  .home__duo { order: 6; }
  .home__news { order: 7; }
  .home__workdays { order: 8; }
  .home__poll { order: 9; }
}

@media (max-width: 640px) {
  .home__inner { gap: 14px; padding: 16px 16px 24px; }
}
</style>
