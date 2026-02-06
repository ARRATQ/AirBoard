<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 p-4 md:p-8">
    <div class="max-w-7xl mx-auto">
      <!-- Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between mb-8 gap-4">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-3">
            <div class="p-2 bg-blue-600 rounded-lg">
              <Icon icon="mdi:chart-box" class="h-6 w-6 text-white" />
            </div>
            {{ $t('analytics.title') }}
          </h1>
          <p class="mt-1 text-gray-500 dark:text-gray-400">
            {{ $t('analytics.subtitle') }}
          </p>
        </div>
        <div class="flex items-center gap-3">
          <button @click="mockExport" class="btn-secondary flex items-center gap-2">
            <Icon icon="mdi:download" class="h-4 w-4" />
            {{ $t('analytics.exportData') }}
          </button>
        </div>
      </div>

      <!-- Loading state -->
      <div v-if="isLoading" class="flex flex-col items-center justify-center py-24">
        <Icon icon="mdi:loading" class="h-12 w-12 animate-spin text-blue-500" />
        <p class="mt-4 text-gray-500 animate-pulse">Chargement des analyses...</p>
      </div>

      <!-- Analytics Dashboard -->
      <div v-else-if="analytics" class="space-y-8 animate-in fade-in duration-700">
        <!-- KPI Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <!-- Total Clicks -->
          <div class="kpi-card group">
            <div class="flex justify-between items-start">
              <div>
                <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('analytics.totalClicks') }}</p>
                <h3 class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ analytics.total_clicks.toLocaleString() }}</h3>
                <div v-if="analytics.clicks_growth !== 0" class="flex items-center mt-2 text-xs font-semibold" :class="analytics.clicks_growth > 0 ? 'text-green-500' : 'text-red-500'">
                  <Icon :icon="analytics.clicks_growth > 0 ? 'mdi:trending-up' : 'mdi:trending-down'" class="mr-1" />
                  {{ Math.abs(analytics.clicks_growth).toFixed(1) }}% 
                  <span class="ml-1 font-normal text-gray-400">{{ $t('analytics.vsPrevious') }}</span>
                </div>
              </div>
              <div class="p-3 bg-blue-50 dark:bg-blue-900/30 rounded-full group-hover:scale-110 transition-transform">
                <Icon icon="mdi:cursor-default-click" class="h-6 w-6 text-blue-600 dark:text-blue-400" />
              </div>
            </div>
          </div>

          <!-- Active Users -->
          <div class="kpi-card group">
            <div class="flex justify-between items-start">
              <div>
                <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('analytics.uniqueUsers') }}</p>
                <h3 class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ analytics.total_unique_users.toLocaleString() }}</h3>
                <div v-if="analytics.users_growth !== 0" class="flex items-center mt-2 text-xs font-semibold" :class="analytics.users_growth > 0 ? 'text-green-500' : 'text-red-500'">
                  <Icon :icon="analytics.users_growth > 0 ? 'mdi:trending-up' : 'mdi:trending-down'" class="mr-1" />
                  {{ Math.abs(analytics.users_growth).toFixed(1) }}%
                  <span class="ml-1 font-normal text-gray-400">{{ $t('analytics.vsPrevious') }}</span>
                </div>
              </div>
              <div class="p-3 bg-green-50 dark:bg-green-900/30 rounded-full group-hover:scale-110 transition-transform">
                <Icon icon="mdi:account-group" class="h-6 w-6 text-green-600 dark:text-green-400" />
              </div>
            </div>
          </div>

          <!-- Engagement Rate -->
          <div class="kpi-card group">
            <div class="flex justify-between items-start">
              <div>
                <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('analytics.engagementRate') }}</p>
                <h3 class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ engagementRate }}%</h3>
                <div class="w-24 bg-gray-200 dark:bg-gray-700 h-1.5 rounded-full mt-3 overflow-hidden">
                  <div class="bg-purple-500 h-full transition-all duration-1000" :style="{ width: engagementRate + '%' }"></div>
                </div>
              </div>
              <div class="p-3 bg-purple-50 dark:bg-purple-900/30 rounded-full group-hover:scale-110 transition-transform">
                <Icon icon="mdi:lightning-bolt" class="h-6 w-6 text-purple-600 dark:text-purple-400" />
              </div>
            </div>
          </div>

          <!-- Peak Usage -->
          <div class="kpi-card group">
            <div class="flex justify-between items-start">
              <div>
                <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ $t('analytics.peakUsage') }}</p>
                <h3 class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ peakUsageLabel }}</h3>
                <p class="text-xs text-gray-400 mt-2 italic">Distribution sur 30j</p>
              </div>
              <div class="p-3 bg-orange-50 dark:bg-orange-900/30 rounded-full group-hover:scale-110 transition-transform">
                <Icon icon="mdi:clock-fast" class="h-6 w-6 text-orange-600 dark:text-orange-400" />
              </div>
            </div>
          </div>
        </div>

        <!-- Main Charts Section -->
        <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <!-- Daily Activity Curve -->
          <div class="card p-6 lg:col-span-2">
            <div class="flex items-center justify-between mb-8">
              <h2 class="text-lg font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <Icon icon="mdi:chart-line" class="text-blue-500" />
                {{ $t('analytics.dailyActivity') }}
              </h2>
              <div class="bg-blue-50 dark:bg-blue-900/20 px-3 py-1 rounded-full text-xs font-medium text-blue-600 dark:text-blue-400">
                Action tracking
              </div>
            </div>

            <div v-if="normalizedDailyActivity.length > 0" class="h-72 relative">
              <svg viewBox="0 0 1000 200" class="w-full h-full overflow-visible" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.2" />
                    <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
                  </linearGradient>
                </defs>
                
                <!-- Grid -->
                <line v-for="i in 5" :key="i" x1="0" :y1="(i-1) * 40" x2="1000" :y2="(i-1) * 40" class="stroke-gray-100 dark:stroke-gray-800" stroke-dasharray="4" />
                
                <!-- Area -->
                <path :d="chartAreaPath" fill="url(#areaGradient)" class="transition-all duration-1000" />
                
                <!-- Line -->
                <path :d="chartPath" class="fill-none stroke-blue-500 dark:stroke-blue-400 stroke-2 transition-all duration-1000" stroke-linejoin="round" stroke-linecap="round" />
                
                <!-- Interactive Dots -->
                <g v-for="(point, i) in chartPoints" :key="'p'+i" class="group">
                  <circle :cx="point.x" :cy="point.y" r="4" class="fill-white stroke-blue-500 stroke-2 opacity-0 group-hover:opacity-100 transition-opacity" />
                  <rect :x="point.x - 15" y="0" width="30" height="200" fill="transparent" class="cursor-pointer">
                    <title>{{ formatDate(point.date) }}: {{ point.val }} clicks</title>
                  </rect>
                </g>
              </svg>
              
              <!-- X-Axis -->
              <div class="flex justify-between mt-6 px-2 border-t border-gray-100 dark:border-gray-800 pt-4">
                <span v-for="(day, i) in xAxisLabels" :key="i" class="text-[10px] font-medium text-gray-400 uppercase tracking-tighter">
                  {{ formatDateShort(day.date) }}
                </span>
              </div>
            </div>
            <div v-else class="flex items-center justify-center h-full text-gray-400 italic">
              {{ $t('analytics.noData') }}
            </div>
          </div>

          <!-- Hourly Distribution Heatmap -->
          <div class="card p-6">
            <h2 class="text-lg font-bold text-gray-900 dark:text-white mb-8 flex items-center gap-2">
              <Icon icon="mdi:camera-timer" class="text-orange-500" />
              {{ $t('analytics.hourlyActivity') }}
            </h2>
            
            <div class="flex items-end justify-between h-64 gap-1">
              <div v-for="h in normalizedHourlyActivity" :key="h.hour" class="flex-1 group relative">
                <div 
                  class="w-full rounded-t-sm transition-all duration-500 bg-orange-100 dark:bg-orange-900/20 group-hover:bg-orange-500 dark:group-hover:bg-orange-500 hover:scale-x-110"
                  :style="{ height: getHourlyHeight(h.click_count) + '%' }"
                >
                  <div class="absolute -top-10 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity bg-gray-900 text-white text-[10px] py-1 px-2 rounded whitespace-nowrap z-10">
                    {{ h.hour }}h: {{ h.click_count }}
                  </div>
                </div>
              </div>
            </div>
            
            <!-- X-Axis (Hours) -->
            <div class="flex justify-between mt-6 px-1 text-[10px] text-gray-400 font-bold">
              <span>0h</span>
              <span>6h</span>
              <span>12h</span>
              <span>18h</span>
              <span>23h</span>
            </div>
          </div>
        </div>

        <!-- Ranking Tables -->
        <div class="grid grid-cols-1 xl:grid-cols-2 gap-8">
          <!-- Top Applications -->
          <div class="card overflow-hidden">
            <div class="p-6 border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50 flex justify-between items-center">
              <h2 class="font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <Icon icon="mdi:trophy" class="text-yellow-500" />
                {{ $t('analytics.topApplications') }}
              </h2>
              <span class="text-xs text-gray-400 font-medium">Top 10</span>
            </div>
            <div class="p-0">
              <div v-if="analytics.top_applications && analytics.top_applications.length > 0">
                <div v-for="(app, index) in analytics.top_applications" :key="app.application_id" class="flex items-center gap-4 p-4 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors border-b last:border-0 border-gray-100 dark:border-gray-800">
                  <span class="w-6 text-sm font-bold text-gray-400">#{{ index + 1 }}</span>
                  <div class="h-10 w-10 flex-shrink-0 rounded-xl flex items-center justify-center shadow-sm" :style="{ backgroundColor: app.color || '#6366f1' }">
                    <Icon :icon="app.icon || 'mdi:application'" class="h-6 w-6 text-white" />
                  </div>
                  <div class="flex-1 min-w-0">
                    <h4 class="text-sm font-bold text-gray-900 dark:text-white truncate">{{ app.application_name }}</h4>
                    <div class="flex items-center gap-2 mt-0.5">
                      <span class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{{ app.click_count }} Clics</span>
                      <span class="text-[10px] text-gray-300">•</span>
                      <span class="text-[10px] font-bold text-gray-400 uppercase tracking-widest">{{ app.unique_users }} Users</span>
                    </div>
                  </div>
                  <div class="w-24 flex-shrink-0">
                    <div class="h-1.5 w-full bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden">
                      <div class="h-full bg-gradient-to-r from-blue-500 to-indigo-600 rounded-full transition-all duration-1000" :style="{ width: (app.click_count / analytics.top_applications[0].click_count * 100) + '%' }"></div>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="py-12 text-center text-gray-400 italic">No data found</div>
            </div>
          </div>

          <!-- Top Users -->
          <div class="card overflow-hidden">
            <div class="p-6 border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50 flex justify-between items-center">
              <h2 class="font-bold text-gray-900 dark:text-white flex items-center gap-2">
                <Icon icon="mdi:account-star" class="text-green-500" />
                {{ $t('analytics.topUsers') }}
              </h2>
              <span class="text-xs text-gray-400 font-medium">Top 10</span>
            </div>
            <div class="p-0">
              <div v-if="analytics.top_users && analytics.top_users.length > 0">
                <div v-for="(user, index) in analytics.top_users" :key="user.user_id" class="flex items-center gap-4 p-4 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors border-b last:border-0 border-gray-100 dark:border-gray-800">
                  <span class="w-6 text-sm font-bold text-gray-400">#{{ index + 1 }}</span>
                  <div class="h-10 w-10 flex-shrink-0 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-700 dark:to-gray-600 flex items-center justify-center border-2 border-white dark:border-gray-800 shadow-sm relative">
                    <span class="text-xs font-black text-gray-600 dark:text-gray-300">{{ getUserInitials(user) }}</span>
                    <div class="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full bg-green-500 border border-white dark:border-gray-800"></div>
                  </div>
                  <div class="flex-1 min-w-0">
                    <h4 class="text-sm font-bold text-gray-900 dark:text-white truncate">{{ user.first_name }} {{ user.last_name }}</h4>
                    <span class="text-[10px] font-bold text-gray-400 uppercase tracking-widest mt-0.5 block">@{{ user.username }}</span>
                  </div>
                  <div class="text-right">
                    <span class="text-sm font-black text-gray-900 dark:text-white">{{ user.click_count }}</span>
                    <span class="text-[10px] font-bold text-gray-400 block uppercase tracking-tighter">Actions</span>
                  </div>
                </div>
              </div>
              <div v-else class="py-12 text-center text-gray-400 italic">No activity detected</div>
            </div>
          </div>
        </div>

        <!-- News Stats Section -->
        <div v-if="newsStats" class="card p-8 border-t-4 border-t-red-500">
          <div class="flex items-center gap-4 mb-8">
            <div class="p-3 bg-red-50 dark:bg-red-900/20 rounded-2xl">
              <Icon icon="mdi:newspaper" class="h-8 w-8 text-red-600" />
            </div>
            <div>
              <h2 class="text-2xl font-black text-gray-900 dark:text-white tracking-tight leading-none uppercase">Statistiques News Hub</h2>
              <p class="text-gray-500 text-sm mt-1 uppercase tracking-widest font-bold">Performance du contenu</p>
            </div>
          </div>

          <div class="grid grid-cols-2 lg:grid-cols-4 gap-8 mb-12">
            <div class="news-kpi">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Articles totaux</p>
              <p class="text-3xl font-black text-gray-900 dark:text-white">{{ newsStats.total_news }}</p>
            </div>
            <div class="news-kpi border-l border-gray-100 dark:border-gray-800 pl-8">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Publiés</p>
              <p class="text-3xl font-black text-green-600">{{ newsStats.published_news }}</p>
            </div>
            <div class="news-kpi border-l border-gray-100 dark:border-gray-800 pl-8">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Vues totales</p>
              <p class="text-3xl font-black text-blue-600">{{ newsStats.total_views }}</p>
            </div>
            <div class="news-kpi border-l border-gray-100 dark:border-gray-800 pl-8">
              <p class="text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Réactions</p>
              <p class="text-3xl font-black text-red-600">{{ newsStats.total_reactions }}</p>
            </div>
          </div>

          <!-- Top News -->
          <div v-if="newsStats.top_news && newsStats.top_news.length > 0">
            <h3 class="text-xs font-black text-gray-400 uppercase tracking-[0.2em] mb-4">Articles les plus populaires</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              <div
                v-for="(article, index) in newsStats.top_news"
                :key="article.id"
                class="group p-4 bg-gray-50 dark:bg-gray-800/40 border border-gray-100 dark:border-gray-800 rounded-xl hover:border-red-200 dark:hover:border-red-900/50 transition-all cursor-pointer"
                @click="viewArticle(article.slug)"
              >
                <div class="flex gap-4 items-start">
                  <span class="text-lg font-black text-gray-200 dark:text-gray-700">0{{ index + 1 }}</span>
                  <div class="flex-1 min-w-0">
                    <p class="text-sm font-bold text-gray-900 dark:text-white leading-tight group-hover:text-red-600 transition-colors line-clamp-2">{{ article.title }}</p>
                    <div class="flex gap-3 mt-3">
                      <div class="flex items-center gap-1 text-[10px] font-bold text-gray-400 uppercase tracking-tighter">
                        <Icon icon="mdi:eye" class="h-3 w-3" /> {{ article.view_count }}
                      </div>
                      <div class="flex items-center gap-1 text-[10px] font-bold text-gray-400 uppercase tracking-tighter">
                        <Icon icon="mdi:heart" class="h-3 w-3 text-red-500" /> {{ article.reaction_count }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="empty-state bg-white dark:bg-gray-800 rounded-3xl p-12 shadow-sm border-2 border-dashed border-gray-200 dark:border-gray-700">
        <div class="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-full mb-6">
          <Icon icon="mdi:chart-scatter-plot" class="h-16 w-16 text-gray-300" />
        </div>
        <h3 class="empty-state-title">{{ $t('analytics.noAnalytics') }}</h3>
        <p class="empty-state-description">{{ $t('analytics.noAnalyticsText') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { analyticsService, newsService } from '@/services/api'

const router = useRouter()
const analytics = ref(null)
const newsStats = ref(null)
const isLoading = ref(false)

const loadAnalytics = async () => {
  try {
    isLoading.value = true
    analytics.value = await analyticsService.getDashboard()
  } catch (error) {
    console.error('Error loading analytics:', error)
  } finally {
    isLoading.value = false
  }
}

const loadNewsStats = async () => {
  try {
    newsStats.value = await newsService.getAnalytics()
  } catch (error) {
    console.error('Error loading news stats:', error)
  }
}

const getUserInitials = (user) => {
  const first = user.first_name?.[0] || ''
  const last = user.last_name?.[0] || user.username?.[0] || '?'
  return (first + last).toUpperCase()
}

// Data Normalization for Daily Curve (30 days)
const normalizedDailyActivity = computed(() => {
  if (!analytics.value?.daily_activity) return []
  
  const days = []
  const now = new Date()
  
  for (let i = 29; i >= 0; i--) {
    const d = new Date(now)
    d.setDate(d.getDate() - i)
    const dateStr = d.toISOString().split('T')[0]
    
    // Check if we have data for this date
    const existing = analytics.value.daily_activity.find(a => a.date.startsWith(dateStr))
    days.push({
      date: dateStr,
      click_count: existing ? existing.click_count : 0,
    })
  }
  return days
})

// Hourly Distribution Normalization (24 hours)
const normalizedHourlyActivity = computed(() => {
  const hours = []
  const data = analytics.value?.hourly_activity || []
  
  for (let i = 0; i < 24; i++) {
    const existing = data.find(h => Number(h.hour) === i)
    hours.push({
      hour: i,
      click_count: existing ? existing.click_count : 0
    })
  }
  return hours
})

// Metrics calculations
const engagementRate = computed(() => {
  if (!analytics.value?.total_active_users || analytics.value.total_active_users === 0) return 0
  return ((analytics.value.total_unique_users / analytics.value.total_active_users) * 100).toFixed(1)
})

const peakUsageLabel = computed(() => {
  if (!analytics.value?.hourly_activity || analytics.value.hourly_activity.length === 0) return '-'
  const sorted = [...analytics.value.hourly_activity].sort((a, b) => b.click_count - a.click_count)
  const peak = Number(sorted[0].hour)
  return `${peak}h - ${peak + 1}h`
})

// SVG Curve generation
const chartPoints = computed(() => {
  const data = normalizedDailyActivity.value
  if (data.length === 0) return []
  
  const maxClicks = Math.max(...data.map(d => d.click_count), 2)
  const width = 1000
  const height = 200
  
  return data.map((d, i) => ({
    x: (i / (data.length - 1)) * width,
    y: height - (d.click_count / maxClicks) * height,
    val: d.click_count,
    date: d.date
  }))
})

const chartPath = computed(() => {
  const points = chartPoints.value
  if (points.length === 0) return ''
  
  let path = `M ${points[0].x} ${points[0].y}`
  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i]
    const p1 = points[i + 1]
    const cp1x = p0.x + (p1.x - p0.x) / 2
    path += ` C ${cp1x} ${p0.y}, ${cp1x} ${p1.y}, ${p1.x} ${p1.y}`
  }
  return path
})

const chartAreaPath = computed(() => {
  const path = chartPath.value
  if (!path) return ''
  const points = chartPoints.value
  return `${path} V 200 H ${points[0].x} Z`
})

// UI Helpers
const xAxisLabels = computed(() => {
  const data = normalizedDailyActivity.value
  return data.filter((_, i) => i % 6 === 0)
})

const formatDateShort = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', { day: '2-digit', month: 'short' })
}

const formatDate = (dateString) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('fr-FR', { day: '2-digit', month: 'long' })
}

const getHourlyHeight = (count) => {
  if (!analytics.value?.hourly_activity || analytics.value.hourly_activity.length === 0) return 0
  const max = Math.max(...analytics.value.hourly_activity.map(h => h.click_count))
  if (max === 0) return 0
  return Math.max((count / max) * 100, 2)
}

const viewArticle = (slug) => {
  router.push({ name: 'NewsDetail', params: { slug } })
}

const mockExport = () => {
  alert('Export des données au format CSV initié...')
}

onMounted(() => {
  loadAnalytics()
  loadNewsStats()
})
</script>

<style scoped>
.btn-secondary {
  @apply px-4 py-2 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl text-sm font-bold text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-all shadow-sm;
}

.kpi-card {
  @apply bg-white dark:bg-gray-800 p-6 rounded-3xl shadow-sm border border-gray-100 dark:border-gray-800 hover:shadow-xl hover:-translate-y-1 transition-all duration-300;
}

.card {
  @apply bg-white dark:bg-gray-800 rounded-3xl shadow-sm border border-gray-100 dark:border-gray-800;
}

.news-kpi {
  @apply flex flex-col;
}

.empty-state {
  @apply flex flex-col items-center justify-center text-center;
}

.empty-state-title {
  @apply text-2xl font-black text-gray-900 dark:text-white mb-2 uppercase tracking-tighter;
}

.empty-state-description {
  @apply text-gray-400 font-medium max-w-sm;
}

/* Animations */
.animate-in {
  animation: fadeIn 0.5s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
