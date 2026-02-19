<template>
  <div class="gamification-hub">
    <!-- Header -->
    <div class="hub-header" data-aos="fade-down">
      <div class="header-content">
        <div class="icon-bg">
          <Icon icon="mdi:medal" class="header-icon" />
        </div>
        <div class="header-text">
          <h1>{{ $t('gamification.title') }}</h1>
          <p>{{ $t('gamification.subtitle') }}</p>
        </div>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="hub-grid">
      <!-- Left Column: Profile & Progress -->
      <div class="column profile-column" data-aos="fade-right">
        <div class="bento-card profile-card">
          <div class="user-info">
            <div class="avatar-container" :class="{ 'has-avatar': !!authStore.user?.avatar_url }">
              <div class="avatar-shine"></div>
              <template v-if="authStore.user?.avatar_url">
                <img :src="authStore.user.avatar_url" alt="Avatar" class="avatar-image" />
              </template>
              <template v-else>
                <span class="avatar-initials">{{ authStore.userInitials }}</span>
              </template>
              <div class="level-badge">{{ profile.level }}</div>
            </div>
            <h2>{{ authStore.user?.first_name }} {{ authStore.user?.last_name }}</h2>
            <p class="username">@{{ authStore.user?.username }}</p>
          </div>

          <div class="progress-section">
            <div class="stats-header">
              <span class="level-label">{{ $t('gamification.level', { level: profile.level }) }}</span>
              <span class="xp-label">{{ profile.xp }} XP</span>
            </div>
            <div class="progress-bar-container">
              <div class="progress-bar-fill" :style="{ width: progressPercentage + '%' }">
                <div class="fill-glow"></div>
              </div>
            </div>
            <p class="next-level-info">{{ $t('gamification.nextLevel', { xp: xpToNextLevel }) }}</p>
          </div>

          <div class="mini-stats">
            <div class="mini-stat-item">
              <Icon icon="mdi:trophy-outline" class="stat-icon" />
              <div class="stat-details">
                <span class="stat-value">{{ unlockedAchievements.length }}</span>
                <span class="stat-name">{{ $t('gamification.achievements') }}</span>
              </div>
            </div>
            <div class="mini-stat-item">
              <Icon icon="mdi:fire" class="stat-icon" />
              <div class="stat-details">
                <span class="stat-value">{{ profile.login_streak || 0 }}</span>
                <span class="stat-name">{{ $t('gamification.streak') }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Recent Activity Feed (Simplified) -->
        <div class="bento-card activity-card mt-4">
          <div class="card-header">
            <Icon icon="mdi:history" class="card-icon" />
            <h3>{{ $t('gamification.recentActivity') }}</h3>
          </div>
          <div class="activity-list">
            <div v-for="tx in recentTransactions" :key="tx.id" class="activity-item">
              <div class="activity-icon-container" :class="tx.reason">
                <Icon :icon="getActivityIcon(tx.reason)" />
              </div>
              <div class="activity-info">
                <p class="activity-reason">{{ getActivityLabel(tx.reason) }}</p>
                <span class="activity-date">{{ formatDate(tx.created_at) }}</span>
              </div>
              <div class="activity-xp">+{{ tx.amount }} XP</div>
            </div>
          </div>
        </div>
      </div>

      <!-- Middle Column: Achievements -->
      <div class="column achievements-column" data-aos="fade-up">
        <div class="bento-card achievements-card">
          <div class="card-header flex justify-between items-center">
            <div class="flex items-center gap-2">
              <Icon icon="mdi:trophy-variant" class="card-icon" />
              <h3>{{ $t('gamification.achievements') }}</h3>
            </div>
            <div class="tabs">
              <button 
                :class="{ active: currentTab === 'all' }" 
                @click="currentTab = 'all'"
              >{{ $t('gamification.all') }}</button>
              <button 
                :class="{ active: currentTab === 'unlocked' }" 
                @click="currentTab = 'unlocked'"
              >{{ $t('gamification.unlocked') }}</button>
            </div>
          </div>

          <div class="achievements-grid">
            <div 
              v-for="achievement in filteredAchievements" 
              :key="achievement.id" 
              class="achievement-item"
              :class="{ locked: !isUnlocked(achievement.code) }"
              @mouseenter="hoveredAchievement = achievement"
              @mouseleave="hoveredAchievement = null"
            >
              <div class="achievement-icon-wrapper" :style="{ backgroundColor: isUnlocked(achievement.code) ? achievement.color : '#cbd5e1' }">
                <Icon :icon="achievement.icon" class="achievement-icon" />
                <div v-if="!isUnlocked(achievement.code)" class="lock-overlay">
                  <Icon icon="mdi:lock" />
                </div>
              </div>
              <div class="achievement-info">
                <h4>{{ achievement.name }}</h4>
                <p>{{ achievement.description }}</p>
                <div v-if="isUnlocked(achievement.code)" class="unlocked-date">
                  {{ $t('gamification.unlockedAt', { date: getUnlockedDate(achievement.code) }) }}
                </div>
                <div v-else class="xp-reward">{{ $t('gamification.xpReward', { xp: achievement.xp_reward }) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Leaderboard -->
      <div class="column leaderboard-column" data-aos="fade-left">
        <div class="bento-card leaderboard-card">
          <div class="card-header">
            <Icon icon="mdi:format-list-numbered" class="card-icon" />
            <h3>{{ $t('gamification.leaderboard') }}</h3>
          </div>
          
          <div class="leaderboard-list">
            <div 
              v-for="(user, index) in leaderboard" 
              :key="user.user_id" 
              class="leaderboard-item"
              :class="{ 'is-current-user': user.user_id === authStore.user?.id }"
            >
              <div class="rank">{{ index + 1 }}</div>
              <div class="user-meta">
                <div class="avatar">
                  <span v-if="index === 0">👑</span>
                  <span v-else-if="index === 1">🥈</span>
                  <span v-else-if="index === 2">🥉</span>
                  <Icon v-else icon="mdi:account" />
                </div>
                <div class="names">
                  <span class="fullname">{{ user.first_name }} {{ user.last_name }}</span>
                  <span class="username">@{{ user.username }}</span>
                </div>
              </div>
              <div class="user-stats">
                <div class="level">Lvl. {{ user.level }}</div>
                <div class="xp">{{ user.xp }} XP</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { gamificationService } from '@/services/api'
import dayjs from 'dayjs'
import AOS from 'aos'
import 'aos/dist/aos.css'

const authStore = useAuthStore()

// State
const profile = ref({ level: 1, xp: 0 })
const achievements = ref([])
const unlockedAchievements = ref([])
const leaderboard = ref([])
const recentTransactions = ref([])
const currentTab = ref('all')
const hoveredAchievement = ref(null)
const isLoading = ref(true)

// Computed
const progressPercentage = computed(() => {
  const currentXP = profile.value.xp
  const currentLevel = profile.value.level
  const xpForCurrentLevel = Math.pow(currentLevel - 1, 2) * 100
  const xpForNextLevel = Math.pow(currentLevel, 2) * 100
  
  const xpInRange = currentXP - xpForCurrentLevel
  const rangeNeed = xpForNextLevel - xpForCurrentLevel
  
  if (rangeNeed === 0) return 0
  return Math.min(Math.max((xpInRange / rangeNeed) * 100, 0), 100)
})

const xpToNextLevel = computed(() => {
  const nextLevel = profile.value.level
  const xpForNextLevel = Math.pow(nextLevel, 2) * 100
  return Math.max(xpForNextLevel - profile.value.xp, 0)
})

const filteredAchievements = computed(() => {
  if (currentTab.value === 'unlocked') {
    return achievements.value.filter(a => isUnlocked(a.code))
  }
  return achievements.value
})

// Methods
const fetchData = async () => {
  try {
    isLoading.value = true
    const [profileData, achievementsData, myAchievementsData, leaderboardData, transactionsData] = await Promise.all([
      gamificationService.getProfile(),
      gamificationService.getAllAchievements(),
      gamificationService.getMyAchievements(),
      gamificationService.getLeaderboard(),
      gamificationService.getTransactions()
    ])
    
    profile.value = profileData
    achievements.value = achievementsData
    unlockedAchievements.value = myAchievementsData || []
    leaderboard.value = leaderboardData
    recentTransactions.value = transactionsData || []
  } catch (error) {
    console.error('Failed to fetch gamification data:', error)
  } finally {
    isLoading.value = false
  }
}

const isUnlocked = (code) => {
  if (!unlockedAchievements.value) return false
  return unlockedAchievements.value.some(ua => ua.achievement?.code === code)
}

const getUnlockedDate = (code) => {
  if (!unlockedAchievements.value) return ''
  const ua = unlockedAchievements.value.find(ua => ua.achievement?.code === code)
  return ua ? dayjs(ua.unlocked_at).format('DD/MM/YYYY') : ''
}

const getActivityIcon = (reason) => {
  const icons = {
    'daily_login': 'mdi:calendar-check',
    'app_click': 'mdi:mouse-left-click',
    'news_publish': 'mdi:newspaper-plus',
    'event_publish': 'mdi:calendar-plus',
    'poll_vote': 'mdi:vote-outline',
    'poll_create': 'mdi:poll',
    'comment_create': 'mdi:comment-text-outline',
    'suggestion_create': 'mdi:lightbulb-on-outline',
    'suggestion_vote': 'mdi:thumb-up-outline',
    'suggestion_comment': 'mdi:comment-edit-outline',
    'achievement_unlock': 'mdi:medal'
  }
  return icons[reason] || 'mdi:star'
}

const getActivityLabel = (reason) => {
  const labels = {
    'daily_login': 'Connexion quotidienne',
    'app_click': 'Ouverture d\'application',
    'news_publish': 'Publication d\'article',
    'event_publish': 'Création d\'événement',
    'poll_vote': 'Vote sondage',
    'poll_create': 'Création sondage',
    'comment_create': 'Commentaire',
    'suggestion_create': 'Nouvelle suggestion',
    'suggestion_vote': 'Vote suggestion',
    'suggestion_comment': 'Commentaire suggestion',
    'achievement_unlock': 'Badge débloqué'
  }
  return labels[reason] || reason
}

const formatDate = (date) => {
  return dayjs(date).format('DD MMM, HH:mm')
}

onMounted(() => {
  fetchData()
  AOS.init({
    duration: 800,
    easing: 'ease-out-cubic',
    once: true
  })
})
</script>


<style scoped>
.gamification-hub {
  padding: 1rem;
  min-height: calc(100vh - 64px);
  background: transparent;
}

/* Header */
.hub-header {
  margin-bottom: 1.25rem;
  background: white;
  padding: 1rem 1.5rem;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
  display: flex;
  align-items: center;
}

.dark .hub-header {
  background: #1e293b;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.icon-bg {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 16px rgba(245, 158, 11, 0.3);
}

.header-icon {
  font-size: 1.75rem;
  color: white;
}

.header-text h1 {
  font-size: 1.5rem;
  font-weight: 800;
  color: #1e293b;
  margin: 0;
}

.dark .header-text h1 {
  color: white;
}

.header-text p {
  color: #64748b;
  margin: 0;
  font-size: 1rem;
}

/* Grid Layout */
.hub-grid {
  display: grid;
  grid-template-columns: 320px 1fr 340px;
  gap: 1.5rem;
}

@media (max-width: 1280px) {
  .hub-grid {
    grid-template-columns: 1fr 1fr;
  }
  .leaderboard-column {
    grid-column: span 2;
  }
}

@media (max-width: 768px) {
  .hub-grid {
    grid-template-columns: 1fr;
  }
  .profile-column, .achievements-column, .leaderboard-column {
    grid-column: span 1;
  }
}

/* Bento Card Style */
.bento-card {
  background: white;
  border-radius: 16px;
  padding: 1.125rem;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.05);
  height: 100%;
}

.dark .bento-card {
  background: #1e293b;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  margin-bottom: 1rem;
}

.card-icon {
  font-size: 1.5rem;
  color: #6366f1;
}

.card-header h3 {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  color: #1e293b;
}

.dark .card-header h3 {
  color: white;
}

/* Profile Card */
.profile-card {
  text-align: center;
}

.avatar-container {
  position: relative;
  width: 80px;
  height: 80px;
  margin: 0 auto 1rem;
}

.user-avatar {
  font-size: 80px;
  color: #e2e8f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dark .user-avatar {
  color: #334155;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
  border: 4px solid white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.dark .avatar-image {
  border-color: #1e293b;
}

.avatar-initials {
  font-size: 1.75rem;
  font-weight: 700;
  color: white;
  text-transform: uppercase;
  background: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 4px solid white;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.dark .avatar-initials {
  border-color: #1e293b;
}

.level-badge {
  position: absolute;
  bottom: 0;
  right: 0;
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
  color: white;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 0.875rem;
  border: 3px solid white;
  box-shadow: 0 4px 8px rgba(99, 102, 241, 0.4);
}

.dark .level-badge {
  border-color: #1e293b;
}

.profile-card h2 {
  font-size: 1.25rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.username {
  color: #64748b;
  font-size: 0.8125rem;
  margin-bottom: 1.25rem;
}

/* Progress Section */
.progress-section {
  margin-bottom: 1.5rem;
}

.stats-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.level-label {
  font-weight: 700;
  font-size: 0.875rem;
  color: #4f46e5;
}

.xp-label {
  font-weight: 600;
  font-size: 0.875rem;
  color: #64748b;
}

.progress-bar-container {
  height: 12px;
  background: #f1f5f9;
  border-radius: 6px;
  overflow: hidden;
  margin-bottom: 0.75rem;
}

.dark .progress-bar-container {
  background: #334155;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366f1 0%, #a855f7 100%);
  border-radius: 6px;
  position: relative;
  transition: width 1s ease-out;
}

.fill-glow {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  width: 20px;
  background: rgba(255, 255, 255, 0.4);
  filter: blur(5px);
}

.next-level-info {
  font-size: 0.75rem;
  color: #94a3b8;
}

/* Mini Stats */
.mini-stats {
  display: flex;
  gap: 1rem;
}

.mini-stat-item {
  flex: 1;
  background: #f8fafc;
  padding: 0.75rem;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  border: 1px solid #f1f5f9;
}

.dark .mini-stat-item {
  background: #0f172a;
  border-color: #1e293b;
}

.mini-stat-item .stat-icon {
  font-size: 1.5rem;
  color: #f59e0b;
}

.stat-details {
  text-align: left;
}

.stat-value {
  display: block;
  font-weight: 800;
  font-size: 1.125rem;
  color: #1e293b;
  line-height: 1;
}

.dark .stat-value {
  color: white;
}

.stat-name {
  font-size: 0.75rem;
  color: #94a3b8;
}

/* Achievements Grid */
.achievements-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.achievement-item {
  display: flex;
  gap: 0.875rem;
  padding: 0.875rem;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  transition: all 0.3s ease;
}

.dark .achievement-item {
  background: #0f172a;
  border-color: #1e293b;
}

.achievement-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.05);
}

.achievement-item.locked {
  opacity: 0.6;
  filter: grayscale(1);
}

.achievement-icon-wrapper {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  position: relative;
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.achievement-icon {
  font-size: 1.75rem;
  color: white;
}

.lock-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 1.25rem;
}

.achievement-info h4 {
  margin: 0 0 0.25rem 0;
  font-size: 1rem;
  font-weight: 700;
}

.achievement-info p {
  margin: 0;
  font-size: 0.75rem;
  color: #64748b;
  line-height: 1.3;
}

.unlocked-date {
  font-size: 0.75rem;
  color: #10b981;
  font-weight: 600;
  margin-top: 0.5rem;
}

.xp-reward {
  font-size: 0.75rem;
  color: #6366f1;
  font-weight: 600;
  margin-top: 0.5rem;
}

/* Leaderboard */
.leaderboard-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.leaderboard-item {
  display: flex;
  align-items: center;
  padding: 0.625rem 0.875rem;
  background: #f8fafc;
  border-radius: 10px;
  gap: 0.75rem;
  border: 1px solid #f1f5f9;
}

.dark .leaderboard-item {
  background: #0f172a;
  border-color: #1e293b;
}

.leaderboard-item.is-current-user {
  border: 2px solid #6366f1;
  background: #e0e7ff33;
}

.rank {
  font-weight: 800;
  color: #94a3b8;
  width: 20px;
}

.user-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
}

.user-meta .avatar {
  width: 36px;
  height: 36px;
  background: #e2e8f0;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.dark .user-meta .avatar {
  background: #334155;
}

.names {
  display: flex;
  flex-direction: column;
}

.fullname {
  font-weight: 600;
  font-size: 0.875rem;
}

.names .username {
  font-size: 0.75rem;
  color: #94a3b8;
  margin: 0;
}

.user-stats {
  text-align: right;
}

.user-stats .level {
  font-size: 0.75rem;
  font-weight: 700;
  color: #4f46e5;
}

.user-stats .xp {
  font-size: 0.875rem;
  font-weight: 700;
}

/* Tabs */
.tabs {
  background: #f1f5f9;
  padding: 0.25rem;
  border-radius: 10px;
  display: flex;
}

.dark .tabs {
  background: #334155;
}

.tabs button {
  padding: 0.375rem 1rem;
  border-radius: 8px;
  border: none;
  background: transparent;
  font-size: 0.8125rem;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  transition: all 0.2s;
}

.tabs button.active {
  background: white;
  color: #4f46e5;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.dark .tabs button.active {
  background: #1e293b;
  color: white;
}
</style>
