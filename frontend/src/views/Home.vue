<template>
  <div class="pulse-home" :class="{ dark: appStore.isDarkMode }">
    <!-- ─── TOPBAR ──────────────────────────────────────────────────── -->
    <header class="pulse-topbar">
      <!-- ── Announcements Ticker ───────────────────────────────────── -->
      <div
        v-if="activeAnnouncements.length"
        class="pulse-ticker"
        @mouseenter="pauseTicker"
        @mouseleave="resumeTicker"
      >
        <Icon :icon="tickerIcon" class="pulse-ticker-icon" :class="`pulse-ticker-icon--${activeAnnouncements[tickerIndex]?.type || 'info'}`" />
        <div class="pulse-ticker-track">
          <Transition name="ticker-slide" mode="out-in">
            <span :key="tickerIndex" class="pulse-ticker-text">
              <strong v-if="activeAnnouncements[tickerIndex]?.title" class="pulse-ticker-title">
                {{ activeAnnouncements[tickerIndex].title }}
              </strong>
              <span v-if="activeAnnouncements[tickerIndex]?.content" class="pulse-ticker-content">
                {{ activeAnnouncements[tickerIndex].content }}
              </span>
            </span>
          </Transition>
        </div>
        <div v-if="activeAnnouncements.length > 1" class="pulse-ticker-nav">
          <button class="pulse-ticker-dot-btn" @click="prevTicker">
            <Icon icon="mdi:chevron-up" />
          </button>
          <span class="pulse-ticker-counter">{{ tickerIndex + 1 }}/{{ activeAnnouncements.length }}</span>
          <button class="pulse-ticker-dot-btn" @click="nextTicker">
            <Icon icon="mdi:chevron-down" />
          </button>
        </div>
      </div>

      <div class="pulse-topbar-spacer" />

      <form class="pulse-search" @submit.prevent="goSearch">
        <Icon icon="mdi:magnify" class="pulse-search-icon" />
        <input
          v-model="searchQuery"
          class="pulse-search-input"
          :placeholder="$t('search.placeholder') || 'Cherche une app, un article…'"
          @keydown.esc="searchQuery = ''"
        />
        <kbd class="pulse-search-kbd">⌘K</kbd>
      </form>

      <div class="pulse-topbar-meta">
        <div v-if="weather.temp !== null" class="pulse-weather" :title="weather.city">
          <Icon :icon="weather.icon" class="pulse-weather-icon" />
          <span class="pulse-weather-temp">{{ weather.temp }}°</span>
          <span class="pulse-weather-city">{{ weather.city }}</span>
          <span class="pulse-topbar-sep">·</span>
        </div>
        <span>{{ currentDateLabel }}</span>
      </div>
    </header>

    <!-- ─── SCROLLABLE BODY ────────────────────────────────────────── -->
    <div class="pulse-body">

      <!-- Loading State -->
      <div v-if="isLoading" class="pulse-loading">
        <Icon icon="mdi:loading" class="pulse-loading-icon" />
        <p>{{ $t('home.loading') }}</p>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="pulse-error">
        <Icon icon="mdi:alert-circle" class="text-red-500 text-4xl mb-2" />
        <h3 class="font-semibold text-gray-800 dark:text-white">{{ $t('home.error.title') }}</h3>
        <p class="text-sm text-gray-500 mb-3">{{ $t('home.error.message') }}</p>
        <button class="pulse-retry-btn" @click="loadHomeData">
          <Icon icon="mdi:refresh" class="h-4 w-4" />
          {{ $t('common.retry') }}
        </button>
      </div>

      <template v-else>

        <!-- ① GREETING HERO + XP ─────────────────────────────────── -->
        <div class="pulse-hero">
          <!-- LEFT: gradient greeting -->
          <div class="pulse-greeting" :style="greetingStyle">
            <div class="pulse-greeting-blobs">
              <div class="blob blob-1" />
              <div class="blob blob-2" />
            </div>
            <div class="pulse-greeting-inner">
              <p class="greeting-eyebrow">{{ currentDateFull }}</p>
              <h1 class="greeting-title">
                <em>{{ greetingWord }},</em> <span class="greeting-name">{{ firstName }}.</span>
              </h1>
              <p class="greeting-sub" v-if="pendingActionsText">{{ pendingActionsText }}</p>
              <div class="greeting-pills">
                <button
                  v-if="activePoll"
                  class="greeting-pill"
                  @click="router.push('/polls')"
                >
                  <Icon icon="mdi:poll" class="h-3.5 w-3.5" />
                  {{ $t('home.vote') }}
                </button>
                <button
                  v-if="latestNewsGroup"
                  class="greeting-pill"
                  @click="router.push('/news')"
                >
                  <Icon icon="mdi:newspaper-variant-outline" class="h-3.5 w-3.5" />
                  {{ latestNewsGroup.type.name }}
                </button>
                <button
                  v-if="nextEvent"
                  class="greeting-pill"
                  @click="router.push('/events')"
                >
                  <Icon icon="mdi:calendar" class="h-3.5 w-3.5" />
                  {{ nextEvent.title }}
                </button>
              </div>
            </div>
          </div>

          <!-- RIGHT: XP / gamification panel -->
          <div class="pulse-xp">
            <div class="xp-header">
              <span class="xp-title">{{ $t('home.xp.title') }}</span>
              <router-link to="/gamification" class="xp-detail-link">{{ $t('common.details') }} →</router-link>
            </div>

            <!-- Level ring + score -->
            <div class="xp-progress-row" v-if="homeData.gamification">
              <div class="xp-ring-wrap">
                <svg class="xp-ring" viewBox="0 0 60 60">
                  <circle cx="30" cy="30" r="26" fill="none" stroke-width="5" class="xp-ring-track" />
                  <circle
                    cx="30" cy="30" r="26" fill="none" stroke-width="5"
                    class="xp-ring-fill"
                    :style="ringStyle"
                    stroke-linecap="round"
                    transform="rotate(-90 30 30)"
                  />
                </svg>
                <span class="xp-level">{{ homeData.gamification.level }}</span>
              </div>
              <div class="xp-score-col">
                <div class="xp-score">{{ homeData.gamification.xp.toLocaleString('fr-FR') }} <span class="xp-unit">XP</span></div>
                <div class="xp-next">{{ homeData.gamification.next_level_xp - homeData.gamification.xp }} {{ $t('home.xp.beforeLevel', { level: homeData.gamification.level + 1 }) }}</div>
                <div class="xp-bar-wrap">
                  <div class="xp-bar-fill" :style="{ width: homeData.gamification.progress_percent + '%' }" />
                </div>
              </div>
            </div>

            <!-- Mini stats -->
            <div class="xp-mini-stats">
              <div class="xp-mini-stat">
                <Icon icon="mdi:fire" class="text-orange-500" />
                <span class="xp-mini-val">{{ homeData.gamification?.streak_days || '–' }}</span>
                <span class="xp-mini-label">{{ $t('home.xp.streak') }}</span>
              </div>
              <div class="xp-mini-stat">
                <Icon icon="mdi:trophy" class="text-amber-500" />
                <span class="xp-mini-val">{{ homeData.gamification?.badges_count || homeData.gamification?.recent_badges?.length || 0 }}</span>
                <span class="xp-mini-label">{{ $t('home.xp.badges') }}</span>
              </div>
              <div class="xp-mini-stat">
                <Icon icon="mdi:star" class="text-yellow-400" />
                <span class="xp-mini-val">{{ homeData.gamification?.rank ? '#' + homeData.gamification.rank : '–' }}</span>
                <span class="xp-mini-label">{{ $t('home.xp.rank') }}</span>
              </div>
            </div>

            <!-- Badge row -->
            <div class="xp-badges" v-if="homeData.gamification?.recent_badges?.length">
              <div
                v-for="badge in homeData.gamification.recent_badges.slice(0, 5)"
                :key="badge.id"
                class="xp-badge-chip"
                :style="{ background: badge.color + '22' }"
                :title="badge.name"
              >
                <Icon :icon="badge.icon" :style="{ color: badge.color }" class="text-base" />
              </div>
            </div>
          </div>
        </div>

        <!-- ② 7 KPI CARDS ─────────────────────────────────────────── -->
        <div class="pulse-kpis" v-if="kpiCards.length">
          <div
            v-for="kpi in kpiCards"
            :key="kpi.label"
            class="pulse-kpi"
            :class="{ 'pulse-kpi--hover': hoveredKpi === kpi.label }"
            @mouseenter="hoveredKpi = kpi.label"
            @mouseleave="hoveredKpi = null"
            @click="router.push(kpi.link)"
          >
            <div class="kpi-top-bar" :class="{ 'kpi-top-bar--visible': hoveredKpi === kpi.label, 'kpi-top-bar--live': kpi.live }" />
            <span class="kpi-icon">{{ kpi.icon }}</span>
            <div class="kpi-text">
              <div class="kpi-value">{{ kpi.value }}</div>
              <div class="kpi-label">{{ kpi.label }}</div>
            </div>
            <Icon
              icon="mdi:chevron-right"
              class="kpi-arrow"
              :class="{ 'kpi-arrow--visible': hoveredKpi === kpi.label }"
            />
          </div>
        </div>

        <!-- ③ MAIN GRID: News + Right Rail ──────────────────────── -->
        <div class="pulse-main-grid">

          <!-- LEFT: TABBED NEWS (dominant) -->
          <div class="pulse-news-card">
            <!-- Tab header -->
            <div class="news-header">
              <div class="news-header-top">
                <span class="news-title">{{ $t('home.newsHub') }}</span>
                <router-link to="/news" class="news-see-all">{{ $t('home.viewAll') }} →</router-link>
              </div>
              <div class="news-tabs" v-if="newsTabs.length > 1">
                <button
                  v-for="tab in newsTabs"
                  :key="tab.key"
                  class="news-tab"
                  :class="{ 'news-tab--active': activeTab === tab.key }"
                  @click="selectTab(tab.key)"
                >
                  <Icon :icon="tab.icon || 'mdi:newspaper'" class="h-3.5 w-3.5" />
                  {{ tab.label }}
                  <span v-if="tab.newCount > 0" class="news-tab-count">{{ tab.newCount }}</span>
                </button>
              </div>
            </div>

            <!-- News items -->
            <div class="news-list-scroll">
              <template v-if="activeNewsItems.length">
                <div
                  v-for="article in activeNewsItems"
                  :key="article.id"
                  class="news-row"
                  @click="router.push('/news/' + article.slug)"
                >
                  <div
                    class="news-row-icon"
                    :style="{ background: (article.news_type?.color || article.category?.color || '#3b82f6') + '22', borderColor: (article.news_type?.color || article.category?.color || '#3b82f6') + '44' }"
                  >
                    <Icon
                      :icon="article.news_type?.icon || article.category?.icon || 'mdi:newspaper'"
                      :style="{ color: article.news_type?.color || article.category?.color || '#3b82f6' }"
                      class="text-xl"
                    />
                  </div>
                  <div class="news-row-body">
                    <div class="news-row-title">{{ article.title }}</div>
                    <div class="news-row-meta">
                      <span
                        class="news-row-tag"
                        :style="{
                          background: (article.news_type?.color || article.category?.color || '#3b82f6') + '22',
                          color: article.news_type?.color || article.category?.color || '#3b82f6'
                        }"
                      >{{ article.news_type?.name || article.category?.name || 'Article' }}</span>
                      <span class="news-row-author">{{ article.author?.username || article.author_name }}</span>
                      <span class="news-row-dot">·</span>
                      <span class="news-row-date">{{ formatRelDate(article.published_at || article.created_at) }}</span>
                    </div>
                  </div>
                  <Icon icon="mdi:chevron-right" class="news-row-arrow" />
                </div>

                <!-- Skeleton filler rows -->
                <div
                  v-for="i in Math.max(0, newsPerTab - activeNewsItems.length)"
                  :key="'sk' + i"
                  class="news-row news-row--skeleton"
                >
                  <div class="news-skeleton-icon" />
                  <div class="news-skeleton-body">
                    <div class="news-skeleton-line" :style="{ width: (75 - i * 10) + '%' }" />
                    <div class="news-skeleton-line" style="width: 40%" />
                  </div>
                </div>
              </template>

              <div v-else class="news-empty">
                <Icon icon="mdi:newspaper-variant-outline" class="text-4xl text-gray-300 dark:text-gray-600 mb-2" />
                <p class="text-sm text-gray-400">{{ $t('home.noNews') || 'Aucun article pour le moment' }}</p>
              </div>
            </div>
          </div>

          <!-- RIGHT RAIL ────────────────────────────────────────── -->
          <div class="pulse-rail">

            <!-- Favorite Apps -->
            <div class="rail-card" v-if="homeData.favorite_apps?.length">
              <div class="rail-card-header">
                <span class="rail-card-title">
                  <Icon icon="mdi:star" class="text-amber-400" />
                  {{ $t('home.favorites.title') || 'Tes applications' }}
                </span>
                <router-link to="/dashboard" class="rail-see-all">Portail →</router-link>
              </div>
              <div class="apps-grid">
                <div
                  v-for="app in homeData.favorite_apps.slice(0, 4)"
                  :key="app.id"
                  class="app-tile"
                  @click="openApp(app)"
                >
                  <Icon
                    :icon="app.icon || 'mdi:application'"
                    class="app-tile-icon"
                    :style="{ color: app.color || '#6366f1' }"
                  />
                  <span class="app-tile-name">{{ app.name }}</span>
                </div>
              </div>
            </div>

            <!-- Active Poll -->
            <div class="rail-card" v-if="activePoll">
              <div class="rail-card-header">
                <span class="rail-card-title">
                  <Icon icon="mdi:poll" class="text-violet-500" />
                  {{ $t('home.polls.title') || 'Sondage actif' }}
                </span>
                <span class="poll-live-badge">
                  <span class="poll-live-dot" />
                  {{ daysLeft(activePoll.end_date) }}
                </span>
              </div>
              <p class="poll-question">{{ activePoll.title }}</p>
              <div class="poll-options">
                <div
                  v-for="opt in (activePoll.options || []).slice(0, 4)"
                  :key="opt.id"
                  class="poll-option"
                >
                  <div class="poll-option-label">
                    <span>{{ opt.text }}</span>
                    <span class="poll-option-pct">{{ pollPercent(opt, activePoll) }}%</span>
                  </div>
                  <div class="poll-bar-track">
                    <div
                      class="poll-bar-fill"
                      :class="{ 'poll-bar-fill--lead': isLeading(opt, activePoll) }"
                      :style="{ width: pollPercent(opt, activePoll) + '%' }"
                    />
                  </div>
                </div>
              </div>
              <button class="poll-vote-btn" @click="router.push('/polls')">
                {{ $t('home.polls.vote') }}
              </button>
            </div>

            <!-- Upcoming Events -->
            <div class="rail-card rail-card--flex" v-if="homeData.upcoming_events?.length">
              <div class="rail-card-header">
                <span class="rail-card-title">
                  <Icon icon="mdi:calendar-clock" class="text-blue-500" />
                  {{ $t('home.upcomingEvents.title') || 'Événements' }}
                </span>
                <router-link to="/events" class="rail-see-all">Agenda →</router-link>
              </div>
              <div class="events-list">
                <div
                  v-for="ev in homeData.upcoming_events.slice(0, 4)"
                  :key="ev.id"
                  class="ev-row"
                  @click="router.push('/events/' + ev.slug)"
                >
                  <div class="ev-date">
                    <div class="ev-month">{{ formatMonth(ev.start_date) }}</div>
                    <div class="ev-day">{{ formatDay(ev.start_date) }}</div>
                  </div>
                  <div class="ev-bar" :style="{ background: ev.color || accentColor }" />
                  <div class="ev-title">{{ ev.title }}</div>
                </div>
              </div>
            </div>

          </div>
        </div>

      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { homeService } from '@/services/api'
const router = useRouter()
const { t, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

// ── State ───────────────────────────────────────────────────────────
const searchQuery = ref('')
const goSearch = () => {
  const q = searchQuery.value.trim()
  router.push(q ? `/search?q=${encodeURIComponent(q)}` : '/search')
}

const homeData = ref({})
const isLoading = ref(true)
const error = ref(null)
const activeTab = ref('')
const hoveredKpi = ref(null)

// Per-tab "last seen" timestamps (ISO strings in localStorage)
const tabLastSeen = ref({})

const loadTabLastSeen = () => {
  const stored = localStorage.getItem('airboard_tab_last_seen')
  tabLastSeen.value = stored ? JSON.parse(stored) : {}
}

const markTabSeen = (slug) => {
  tabLastSeen.value[slug] = new Date().toISOString()
  localStorage.setItem('airboard_tab_last_seen', JSON.stringify(tabLastSeen.value))
}

// ── Colors ──────────────────────────────────────────────────────────
const accentColor = 'var(--accent)'

const greetingStyle = computed(() => ({
  background: 'linear-gradient(120deg, var(--gradient-from), var(--gradient-mid) 58%, var(--gradient-to))'
}))

const ringStyle = computed(() => {
  const pct = homeData.value.gamification?.progress_percent || 0
  const circumference = 2 * Math.PI * 26
  const dash = (pct / 100) * circumference
  return { strokeDasharray: `${dash} ${circumference}` }
})

// ── Greeting ─────────────────────────────────────────────────────────
const dateLocale = computed(() => {
  const map = { fr: 'fr-FR', en: 'en-US', es: 'es-ES', ar: 'ar-MA' }
  return map[locale.value] || locale.value
})

const currentDateFull = computed(() => {
  const now = new Date()
  return now.toLocaleDateString(dateLocale.value, { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }).toUpperCase() +
    ' · ' + now.toLocaleTimeString(dateLocale.value, { hour: '2-digit', minute: '2-digit' })
})

const currentDateLabel = computed(() => {
  const now = new Date()
  return now.toLocaleDateString(dateLocale.value, { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })
})

const greetingWord = computed(() => {
  const h = new Date().getHours()
  if (h < 12) return t('home.hero.greeting.morning')
  if (h < 18) return t('home.hero.greeting.afternoon')
  return t('home.hero.greeting.evening')
})

const firstName = computed(() => {
  const name = authStore.userDisplayName || authStore.user?.username || ''
  return name.split(' ')[0]
})

const pendingActionsText = computed(() => {
  const polls = homeData.value.polls?.filter(p => p.is_active).length || 0
  const news = homeData.value.recent_news_by_type?.reduce((s, g) => s + g.news.length, 0) || 0
  if (!polls && !news) return ''
  const parts = []
  if (polls) parts.push(t('home.hero.pendingPoll', { count: polls }))
  if (news) parts.push(t('home.hero.pendingNews', { count: news }))
  return t('home.hero.pendingText', { items: parts.join(t('common.and')) })
})

const activePoll = computed(() => homeData.value.polls?.find(p => p.is_active) || null)
const latestNewsGroup = computed(() => homeData.value.recent_news_by_type?.[0] || null)
const nextEvent = computed(() => homeData.value.upcoming_events?.[0] || null)

// ── KPI Cards ────────────────────────────────────────────────────────
const kpiCards = computed(() => {
  const stats = homeData.value.stats
  const role = homeData.value.user_role
  if (!stats) return []

  if (role === 'admin') {
    return [
      { value: stats.total_users ?? 0, label: 'Utilisateurs', icon: '👥', link: '/admin/users' },
      { value: stats.total_groups ?? 0, label: 'Groupes', icon: '🏷️', link: '/admin/groups' },
      { value: stats.total_app_groups ?? 0, label: 'Groupes d\'apps', icon: '📦', link: '/admin/app-groups' },
      { value: stats.total_apps ?? 0, label: 'Applications', icon: '🔲', link: '/admin/applications' },
      { value: stats.total_news ?? 0, label: 'Articles', icon: '📰', link: '/admin/news' },
      { value: stats.total_events ?? 0, label: 'Événements', icon: '📅', link: '/admin/events' },
      { value: stats.total_polls ?? 0, label: 'Sondages', icon: '📊', link: '/admin/polls', live: (stats.total_polls ?? 0) > 0 },
    ]
  }

  if (homeData.value.managed_group_ids?.length) {
    return [
      { value: stats.managed_groups_count ?? 0, label: 'Groupes gérés', icon: '🏷️', link: '/group-admin' },
      { value: stats.total_members_count ?? 0, label: 'Membres', icon: '👥', link: '/group-admin' },
      { value: stats.managed_app_groups_count ?? 0, label: 'Groupes d\'apps', icon: '📦', link: '/group-admin/app-groups' },
      { value: stats.managed_apps_count ?? 0, label: 'Applications', icon: '🔲', link: '/group-admin/applications' },
      { value: stats.managed_news_count ?? 0, label: 'Articles', icon: '📰', link: '/group-admin/news' },
      { value: stats.managed_polls_count ?? 0, label: 'Sondages', icon: '📊', link: '/group-admin/polls', live: true },
      { value: stats.total_accessible_apps ?? 0, label: 'Apps dispo.', icon: '✅', link: '/dashboard' },
    ]
  }

  return [
    { value: stats.total_accessible_apps ?? 0, label: 'Applications', icon: '🔲', link: '/dashboard' },
    { value: stats.total_news ?? 0, label: 'Articles', icon: '📰', link: '/news' },
    { value: stats.total_polls ?? 0, label: 'Sondages', icon: '📊', link: '/polls', live: true },
    { value: stats.total_accessible_events ?? 0, label: 'Événements', icon: '📅', link: '/events' },
  ]
})

// ── News Tabs ────────────────────────────────────────────────────────
const newsPerTab = computed(() => homeData.value.app_settings?.news_per_tab || 5)

const newsTabs = computed(() => {
  return (homeData.value.recent_news_by_type || []).map(g => {
    const lastSeen = tabLastSeen.value[g.type.slug]
    const newCount = lastSeen
      ? g.news.filter(n => new Date(n.published_at || n.created_at) > new Date(lastSeen)).length
      : g.news.length
    return {
      key: g.type.slug,
      label: g.type.name,
      icon: g.type.icon,
      newCount,
    }
  })
})

const selectTab = (slug) => {
  activeTab.value = slug
  markTabSeen(slug)
}

const activeNewsItems = computed(() => {
  if (!activeTab.value) return []
  const group = homeData.value.recent_news_by_type?.find(g => g.type.slug === activeTab.value)
  return group?.news || []
})

// ── Poll helpers ─────────────────────────────────────────────────────
const pollPercent = (opt, poll) => {
  const total = poll.total_votes || poll.options?.reduce((s, o) => s + (o.votes || 0), 0) || 1
  return Math.round(((opt.votes || 0) / total) * 100)
}

const isLeading = (opt, poll) => {
  const maxVotes = Math.max(...(poll.options || []).map(o => o.votes || 0))
  return (opt.votes || 0) === maxVotes && maxVotes > 0
}

const daysLeft = (endDate) => {
  if (!endDate) return 'Actif'
  const diff = Math.ceil((new Date(endDate) - new Date()) / (1000 * 60 * 60 * 24))
  if (diff <= 0) return 'Expiré'
  return diff + 'j'
}

// ── Date helpers ─────────────────────────────────────────────────────
const formatRelDate = (dateStr) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const diff = Math.floor((Date.now() - d) / 86400000)
  if (diff === 0) return 'Aujourd\'hui'
  if (diff === 1) return 'Hier'
  if (diff < 7) return `il y a ${diff}j`
  return d.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short' })
}

const formatMonth = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).toLocaleDateString('fr-FR', { month: 'short' }).toUpperCase()
}

const formatDay = (dateStr) => {
  if (!dateStr) return ''
  return new Date(dateStr).getDate()
}

// ── App opener ───────────────────────────────────────────────────────
const openApp = (app) => {
  if (app.open_in_new_tab) window.open(app.url, '_blank', 'noopener,noreferrer')
  else window.location.href = app.url
}

// ── Data loading ─────────────────────────────────────────────────────
const loadHomeData = async () => {
  try {
    isLoading.value = true
    error.value = null
    loadTabLastSeen()
    homeData.value = await homeService.getHomeData()
    // Set default tab to first news type and mark it seen
    if (homeData.value.recent_news_by_type?.length) {
      const firstSlug = homeData.value.recent_news_by_type[0].type.slug
      activeTab.value = firstSlug
      markTabSeen(firstSlug)
    }
    tickerIndex.value = 0
    startTicker()
  } catch (err) {
    console.error('Failed to load home data:', err)
    error.value = err
  } finally {
    isLoading.value = false
  }
}

// ── Weather ──────────────────────────────────────────────────────────
const weather = ref({ temp: null, icon: 'mdi:weather-cloudy', city: '' })

// WMO weather code → mdi icon
const wmoIcon = (code) => {
  if (code === 0) return 'mdi:weather-sunny'
  if (code <= 2) return 'mdi:weather-partly-cloudy'
  if (code <= 3) return 'mdi:weather-cloudy'
  if (code <= 49) return 'mdi:weather-fog'
  if (code <= 59) return 'mdi:weather-rainy'
  if (code <= 69) return 'mdi:weather-snowy-rainy'
  if (code <= 79) return 'mdi:weather-snowy'
  if (code <= 82) return 'mdi:weather-pouring'
  if (code <= 84) return 'mdi:weather-hail'
  if (code <= 99) return 'mdi:weather-lightning-rainy'
  return 'mdi:weather-cloudy'
}

const loadWeather = () => {
  if (!navigator.geolocation) return
  navigator.geolocation.getCurrentPosition(async ({ coords }) => {
    try {
      const { latitude: lat, longitude: lon } = coords
      const [meteo, geo] = await Promise.all([
        fetch(`https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current_weather=true&temperature_unit=celsius`).then(r => r.json()),
        fetch(`https://nominatim.openstreetmap.org/reverse?lat=${lat}&lon=${lon}&format=json&accept-language=fr`).then(r => r.json()),
      ])
      weather.value = {
        temp: Math.round(meteo.current_weather.temperature),
        icon: wmoIcon(meteo.current_weather.weathercode),
        city: geo.address?.city || geo.address?.town || geo.address?.village || geo.address?.county || '',
      }
    } catch { /* silently ignore */ }
  }, () => { /* permission denied — keep no weather */ })
}

// ── Announcements Ticker ─────────────────────────────────────────────
const tickerIndex = ref(0)
let tickerTimer = null

const activeAnnouncements = computed(() => homeData.value.announcements || [])

const tickerIconMap = { info: 'mdi:information-outline', warning: 'mdi:alert-outline', success: 'mdi:check-circle-outline', error: 'mdi:alert-circle-outline' }
const tickerIcon = computed(() => tickerIconMap[activeAnnouncements.value[tickerIndex.value]?.type] || 'mdi:information-outline')

const nextTicker = () => { tickerIndex.value = (tickerIndex.value + 1) % activeAnnouncements.value.length }
const prevTicker = () => { tickerIndex.value = (tickerIndex.value - 1 + activeAnnouncements.value.length) % activeAnnouncements.value.length }

const startTicker = () => {
  if (tickerTimer) clearInterval(tickerTimer)
  if (activeAnnouncements.value.length > 1) tickerTimer = setInterval(nextTicker, 5000)
}
const pauseTicker = () => clearInterval(tickerTimer)
const resumeTicker = () => startTicker()

onMounted(() => { loadHomeData(); loadWeather() })
onUnmounted(() => clearInterval(tickerTimer))
</script>

<style>
/* Google Fonts — unscoped so they load globally */
@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=Newsreader:ital,wght@1,400;1,600&display=swap');
</style>

<style scoped>
/* ── Root ────────────────────────────────────────────────────────── */
.pulse-home {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--bg-page);
  font-family: 'Plus Jakarta Sans', system-ui, sans-serif;
  overflow: hidden;
}

/* ── Topbar ──────────────────────────────────────────────────────── */
.pulse-topbar {
  height: 68px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 0 24px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border);
}

.pulse-search {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 0 0 auto;
  width: 440px;
  padding: 6px 12px;
  background: var(--bg-page);
  border: 1.5px solid var(--border);
  border-radius: 9px;
  cursor: text;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.pulse-search:focus-within {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent) 15%, transparent);
}

.pulse-search-icon {
  color: var(--text-muted);
  font-size: 17px;
  flex-shrink: 0;
}

.pulse-search-input {
  flex: 1;
  font-size: 13.5px;
  color: var(--text-primary);
  background: transparent;
  border: none;
  outline: none;
  box-shadow: none;
  font-family: inherit;
  min-width: 0;
  line-height: 1.4;
}

.pulse-search-input:focus,
.pulse-search-input:focus-visible {
  outline: none;
  box-shadow: none;
}

.pulse-search-input::placeholder {
  color: var(--text-muted);
}

.pulse-search-kbd {
  font-size: 11px;
  padding: 3px 7px;
  border: 1px solid var(--border);
  border-radius: 6px;
  color: var(--text-muted);
  background: var(--bg-card);
  font-family: inherit;
  flex-shrink: 0;
  letter-spacing: 0.02em;
}

.pulse-topbar-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-muted);
  flex-shrink: 0;
}

.pulse-weather {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pulse-weather-icon {
  font-size: 15px;
}

.pulse-weather-temp {
  font-weight: 600;
  color: var(--text-primary);
}

.pulse-weather-city {
  color: var(--text-muted);
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pulse-topbar-sep {
  color: var(--border);
}

.pulse-topbar-spacer {
  flex: 1;
}

/* ── Ticker ──────────────────────────────────────────────────────── */
.pulse-ticker {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}

.pulse-ticker-icon {
  font-size: 17px;
  flex-shrink: 0;
}

.pulse-ticker-icon--info    { color: #3b82f6; }
.pulse-ticker-icon--warning { color: #f59e0b; }
.pulse-ticker-icon--success { color: #10b981; }
.pulse-ticker-icon--error   { color: #ef4444; }

.pulse-ticker-track {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  height: 20px;
  position: relative;
}

.pulse-ticker-text {
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 12.5px;
  line-height: 20px;
  position: absolute;
  inset: 0;
}

.pulse-ticker-title {
  font-weight: 700;
  color: var(--text-primary);
  flex-shrink: 0;
}

.pulse-ticker-content {
  color: var(--text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
}

.pulse-ticker-nav {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 2px;
}

.pulse-ticker-dot-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
  border-radius: 4px;
  padding: 0;
  font-size: 14px;
  transition: color 0.15s, background 0.15s;
}

.pulse-ticker-dot-btn:hover {
  color: var(--text-primary);
  background: var(--border);
}

.pulse-ticker-counter {
  font-size: 10px;
  color: var(--text-muted);
  font-weight: 600;
  min-width: 24px;
  text-align: center;
}

/* Transition verticale du ticker */
.ticker-slide-enter-active,
.ticker-slide-leave-active {
  transition: transform 0.3s ease, opacity 0.3s ease;
}

.ticker-slide-enter-from {
  transform: translateY(100%);
  opacity: 0;
}

.ticker-slide-leave-to {
  transform: translateY(-100%);
  opacity: 0;
}

/* ── Body ────────────────────────────────────────────────────────── */
.pulse-body {
  flex: 1;
  overflow-y: auto;
  padding: 18px 24px 24px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  scrollbar-width: thin;
  scrollbar-color: var(--border) transparent;
}

/* ── Loading / Error ─────────────────────────────────────────────── */
.pulse-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  gap: 12px;
  color: var(--text-muted);
}

.pulse-loading-icon {
  font-size: 2.5rem;
  color: var(--accent);
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.pulse-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  text-align: center;
}

.pulse-retry-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  background: var(--accent);
  color: #fff;
  border: none;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
}

/* ── Hero ────────────────────────────────────────────────────────── */
.pulse-hero {
  display: grid;
  grid-template-columns: 1fr 320px;
  border-radius: 18px;
  overflow: hidden;
  box-shadow: 0 4px 24px rgba(var(--shadow-rgb), 0.12);
  flex-shrink: 0;
}

/* Greeting */
.pulse-greeting {
  position: relative;
  padding: 26px 32px;
  overflow: hidden;
}

.pulse-greeting-blobs {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.blob {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.07);
}

.blob-1 {
  width: 280px;
  height: 280px;
  top: -80px;
  right: -60px;
}

.blob-2 {
  width: 160px;
  height: 160px;
  bottom: -50px;
  left: -30px;
}

.pulse-greeting-inner {
  position: relative;
}

.greeting-eyebrow {
  font-size: 10px;
  color: rgba(255,255,255,0.55);
  letter-spacing: 0.1em;
  margin-bottom: 8px;
  font-weight: 500;
}

.greeting-title {
  font-family: 'Newsreader', Georgia, serif;
  font-size: 34px;
  color: #fff;
  line-height: 1.1;
  margin: 0 0 12px;
  font-style: italic;
  font-weight: 600;
  letter-spacing: -0.02em;
}

.greeting-name {
  font-style: normal;
}

.greeting-sub {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.75);
  margin: 0 0 18px;
  line-height: 1.6;
}

.greeting-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.greeting-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 13px;
  border-radius: 99px;
  border: 1px solid rgba(255, 255, 255, 0.28);
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  font-family: inherit;
  backdrop-filter: blur(8px);
  transition: background 0.15s;
}

.greeting-pill:hover {
  background: rgba(255, 255, 255, 0.28);
}

/* XP Panel */
.pulse-xp {
  padding: 20px 22px;
  background: var(--bg-card);
  border-left: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 12px;
}

.xp-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.xp-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
}

.xp-detail-link {
  font-size: 11px;
  font-weight: 600;
  color: var(--accent);
  text-decoration: none;
}

.xp-progress-row {
  display: flex;
  gap: 14px;
  align-items: center;
}

.xp-ring-wrap {
  position: relative;
  width: 56px;
  height: 56px;
  flex-shrink: 0;
}

.xp-ring {
  width: 56px;
  height: 56px;
}

.xp-ring-track {
  stroke: var(--border);
}

.xp-ring-fill {
  stroke: var(--accent);
  transition: stroke-dasharray 0.8s ease;
}

.xp-level {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 800;
  color: var(--accent);
}

.xp-score-col {
  flex: 1;
  min-width: 0;
}

.xp-score {
  font-size: 22px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.03em;
  line-height: 1;
}

.xp-unit {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-muted);
}

.xp-next {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 3px;
}

.xp-bar-wrap {
  margin-top: 8px;
  height: 5px;
  background: var(--border);
  border-radius: 99px;
  overflow: hidden;
}

.xp-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--accent), var(--accent-light));
  border-radius: 99px;
  transition: width 0.8s ease;
}

.xp-mini-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.xp-mini-stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 8px 4px;
  border-radius: 10px;
  background: var(--bg-page);
  border: 1px solid var(--border);
}

.dark .xp-mini-stat {
  background: var(--bg-surface);
  border-color: var(--border);
}

.xp-mini-val {
  font-size: 16px;
  font-weight: 800;
  color: var(--text-primary);
  line-height: 1;
}

.xp-mini-label {
  font-size: 9px;
  color: var(--text-muted);
  font-weight: 500;
}

.xp-badges {
  display: flex;
  gap: 6px;
}

.xp-badge-chip {
  flex: 1;
  height: 28px;
  border-radius: 8px;
  border: 1px solid rgba(var(--shadow-rgb), 0.08);
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: transform 0.15s;
}

.xp-badge-chip:hover {
  transform: scale(1.1);
}

/* ── KPI Strip ───────────────────────────────────────────────────── */
.pulse-kpis {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
  gap: 10px;
  flex-shrink: 0;
}

.pulse-kpi {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 12px;
  border: 1px solid var(--border);
  background: var(--bg-card);
  cursor: pointer;
  overflow: hidden;
  transition: transform 0.15s, box-shadow 0.15s, border-color 0.15s;
  box-shadow: 0 2px 8px rgba(var(--shadow-rgb), 0.05);
}

.pulse-kpi--hover {
  transform: translateY(-2px);
  border-color: rgba(var(--accent-rgb), 0.45);
  box-shadow: 0 6px 18px rgba(var(--shadow-rgb), 0.10);
}

.kpi-top-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  border-radius: 12px 12px 0 0;
  background: var(--accent);
  opacity: 0;
  transition: opacity 0.18s;
}

.kpi-top-bar--live {
  background: #22c55e;
}

.kpi-top-bar--visible {
  opacity: 1;
}

.kpi-icon {
  font-size: 16px;
  line-height: 1;
  flex-shrink: 0;
}

.kpi-text {
  flex: 1;
  min-width: 0;
}

.kpi-value {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-primary);
  letter-spacing: -0.04em;
  line-height: 1;
}

.kpi-label {
  font-size: 10px;
  color: var(--text-muted);
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.kpi-arrow {
  color: var(--accent);
  font-size: 11px;
  opacity: 0;
  transition: opacity 0.18s;
  flex-shrink: 0;
}

.kpi-arrow--visible {
  opacity: 1;
}

/* ── Main Grid ───────────────────────────────────────────────────── */
.pulse-main-grid {
  display: grid;
  grid-template-columns: 1fr 290px;
  gap: 14px;
  flex: 1;
  min-height: 0;
}

/* ── News Card ───────────────────────────────────────────────────── */
.pulse-news-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(var(--shadow-rgb), 0.06);
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.news-header {
  padding: 18px 20px 0;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.news-header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.news-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.news-see-all {
  font-size: 12px;
  font-weight: 600;
  color: var(--accent);
  text-decoration: none;
}

.news-tabs {
  display: flex;
  gap: 3px;
  margin-bottom: -1px;
}

.news-tab {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 7px 14px;
  border: none;
  background: transparent;
  border-radius: 8px 8px 0 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-muted);
  cursor: pointer;
  font-family: inherit;
  transition: background 0.15s, color 0.15s;
  position: relative;
}

.news-tab--active {
  background: var(--accent);
  color: #fff;
}

.news-tab:not(.news-tab--active):hover {
  background: var(--hover);
  color: var(--text-primary);
}

.news-tab-count {
  display: inline-flex;
  padding: 1px 6px;
  border-radius: 99px;
  font-size: 10px;
  background: rgba(255, 255, 255, 0.28);
}

.news-tab:not(.news-tab--active) .news-tab-count {
  background: var(--border);
  color: var(--text-muted);
}

.news-list-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 8px 12px 14px;
  scrollbar-width: thin;
  scrollbar-color: var(--border) transparent;
}

/* News Row */
.news-row {
  display: flex;
  gap: 12px;
  padding: 13px 10px;
  border-radius: 12px;
  align-items: flex-start;
  cursor: pointer;
  transition: background 0.15s;
}

.news-row:hover {
  background: var(--hover);
}

.news-row-icon {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  border: 1px solid;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}

.news-row-body {
  flex: 1;
  min-width: 0;
}

.news-row-title {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.4;
  margin-bottom: 7px;
}

.news-row-meta {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
}

.news-row-tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.news-row-author,
.news-row-date,
.news-row-dot {
  font-size: 11px;
  color: var(--text-muted);
}

.news-row-arrow {
  color: var(--text-muted);
  font-size: 16px;
  flex-shrink: 0;
  margin-top: 3px;
  opacity: 0.5;
}

/* Skeleton rows */
.news-row--skeleton {
  opacity: 0.45;
  cursor: default;
}

.news-skeleton-icon {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  background: var(--border);
  flex-shrink: 0;
}

.news-skeleton-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  justify-content: center;
}

.news-skeleton-line {
  height: 12px;
  border-radius: 6px;
  background: var(--border);
}

.news-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
}

/* ── Right Rail ──────────────────────────────────────────────────── */
.pulse-rail {
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: var(--border) transparent;
}

.rail-card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 14px;
  padding: 15px 16px;
  box-shadow: 0 2px 10px rgba(var(--shadow-rgb), 0.05);
  flex-shrink: 0;
}

.rail-card--flex {
  flex: 1;
}

.rail-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.rail-card-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13.5px;
  font-weight: 700;
  color: var(--text-primary);
}

.rail-see-all {
  font-size: 11px;
  font-weight: 600;
  color: var(--accent);
  text-decoration: none;
}

/* Apps Grid */
.apps-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.app-tile {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 9px 10px;
  border-radius: 10px;
  border: 1px solid var(--border);
  background: var(--bg-page);
  cursor: pointer;
  transition: background 0.15s;
}

.app-tile:hover {
  background: var(--hover);
}

.app-tile-icon {
  width: 22px;
  height: 22px;
  flex-shrink: 0;
}

.app-tile-name {
  font-size: 11.5px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

/* Poll */
.poll-live-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 99px;
  background: #22c55e22;
  color: #22c55e;
  font-weight: 700;
}

.poll-live-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #22c55e;
  display: inline-block;
}

.poll-question {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.4;
  margin-bottom: 12px;
}

.poll-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.poll-option {}

.poll-option-label {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.poll-option-pct {
  color: var(--text-muted);
  font-weight: 600;
}

.poll-bar-track {
  height: 6px;
  background: var(--border);
  border-radius: 99px;
  overflow: hidden;
}

.poll-bar-fill {
  height: 100%;
  border-radius: 99px;
  background: color-mix(in srgb, var(--text-muted) 33%, transparent);
  transition: width 0.6s ease;
}

.poll-bar-fill--lead {
  background: linear-gradient(90deg, var(--accent), var(--accent-light));
}

.poll-vote-btn {
  width: 100%;
  padding: 9px;
  border-radius: 10px;
  border: none;
  background: var(--accent);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.15s;
}

.poll-vote-btn:hover {
  background: var(--accent-dark);
}

/* Events */
.events-list {
  display: flex;
  flex-direction: column;
}

.ev-row {
  display: flex;
  gap: 10px;
  padding: 11px 0;
  align-items: center;
  border-top: 1px solid var(--border);
  cursor: pointer;
  transition: opacity 0.15s;
}

.ev-row:first-child {
  border-top: none;
}

.ev-row:hover {
  opacity: 0.72;
}

.ev-date {
  width: 38px;
  flex-shrink: 0;
  text-align: center;
}

.ev-month {
  font-size: 8px;
  letter-spacing: 0.12em;
  font-weight: 700;
  color: var(--accent);
  line-height: 1;
}

.ev-day {
  font-size: 20px;
  font-weight: 800;
  color: var(--text-primary);
  line-height: 1.1;
}

.ev-bar {
  width: 3px;
  height: 32px;
  border-radius: 99px;
  flex-shrink: 0;
}

.ev-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
}

/* ── Responsive tweaks ───────────────────────────────────────────── */
@media (max-width: 1200px) {
  .pulse-main-grid {
    grid-template-columns: 1fr 260px;
  }

  .pulse-hero {
    grid-template-columns: 1fr 280px;
  }
}

@media (max-width: 900px) {
  .pulse-hero {
    grid-template-columns: 1fr;
  }

  .pulse-xp {
    display: none;
  }

  .pulse-main-grid {
    grid-template-columns: 1fr;
  }

  .pulse-rail {
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .pulse-kpis {
    grid-template-columns: repeat(4, 1fr);
  }
}
</style>
