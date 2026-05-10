<template>
  <div class="gamification-hub">

    <!-- ── Header ─────────────────────────────────────────── -->
    <div class="hub-header" data-aos="fade-down">
      <div class="header-left">
        <div class="header-icon-wrap">
          <Icon icon="mdi:trophy-variant" class="header-icon" />
        </div>
        <div>
          <h1>{{ $t('gamification.title') }}</h1>
          <p>{{ $t('gamification.subtitle') }}</p>
        </div>
      </div>
      <div class="header-pills">
        <div class="header-pill">
          <Icon icon="mdi:account-group" />
          <span>{{ leaderboard.length }} joueurs</span>
        </div>
        <div class="header-pill accent">
          <Icon icon="mdi:medal-outline" />
          <span>{{ achievements.length }} badges</span>
        </div>
        <div v-if="myRank" class="header-pill gold">
          <Icon icon="mdi:podium" />
          <span>Votre rang : #{{ myRank }}</span>
        </div>
      </div>
    </div>

    <!-- ── Activity Strip (always visible) ───────────────── -->
    <div class="activity-strip" data-aos="fade-up" data-aos-delay="100">
      <div class="strip-label">
        <Icon icon="mdi:lightning-bolt" class="strip-label-icon" />
        <span>Activité récente</span>
      </div>
      <div class="strip-scroll">
        <div
          v-for="tx in recentTransactions"
          :key="tx.id"
          class="strip-item"
          :class="tx.reason"
        >
          <div class="strip-icon">
            <Icon :icon="getActivityIcon(tx.reason)" />
          </div>
          <div class="strip-content">
            <span class="strip-action">{{ getActivityLabel(tx.reason) }}</span>
            <span class="strip-time">{{ formatRelative(tx.created_at) }}</span>
          </div>
          <span class="strip-xp">+{{ tx.amount }} XP</span>
        </div>
        <div v-if="recentTransactions.length === 0" class="strip-empty">
          Aucune activité récente
        </div>
      </div>
    </div>

    <!-- ── Main Grid ──────────────────────────────────────── -->
    <div class="hub-grid" data-aos="fade-up" data-aos-delay="150">

      <!-- Left: Profile Card -->
      <div class="column profile-column">
        <div class="bento-card profile-card">
          <!-- Avatar -->
          <div class="avatar-wrap">
            <div class="avatar-ring">
              <template v-if="authStore.user?.avatar_url">
                <img :src="authStore.user.avatar_url" alt="Avatar" class="avatar-img" />
              </template>
              <template v-else>
                <span class="avatar-initials">{{ authStore.userInitials }}</span>
              </template>
            </div>
            <div class="level-chip">{{ profile.level }}</div>
          </div>

          <h2 class="profile-name">{{ authStore.user?.first_name }} {{ authStore.user?.last_name }}</h2>
          <p class="profile-username">@{{ authStore.user?.username }}</p>

          <!-- XP Bar -->
          <div class="xp-section">
            <div class="xp-row">
              <span class="xp-level">Niveau {{ profile.level }}</span>
              <span class="xp-value">{{ profile.xp.toLocaleString() }} XP</span>
            </div>
            <div class="xp-bar-bg">
              <div class="xp-bar-fill" :style="{ width: progressPercentage + '%' }">
                <span class="xp-bar-shine" />
              </div>
            </div>
            <p class="xp-next">{{ xpToNextLevel.toLocaleString() }} XP pour le niveau {{ profile.level + 1 }}</p>
          </div>

          <!-- Stats -->
          <div class="profile-stats">
            <div class="pstat">
              <div class="pstat-icon trophy"><Icon icon="mdi:trophy-outline" /></div>
              <div class="pstat-value">{{ profile.badges_count ?? unlockedAchievements.length }}</div>
              <div class="pstat-label">Badges</div>
            </div>
            <div class="pstat-divider" />
            <div class="pstat">
              <div class="pstat-icon fire"><Icon icon="mdi:fire" /></div>
              <div class="pstat-value">{{ profile.login_streak || 0 }}</div>
              <div class="pstat-label">Série</div>
            </div>
            <div class="pstat-divider" />
            <div class="pstat">
              <div class="pstat-icon rank"><Icon icon="mdi:podium-gold" /></div>
              <div class="pstat-value">#{{ myRank || '—' }}</div>
              <div class="pstat-label">Classement</div>
            </div>
          </div>

          <!-- Unlocked badges preview -->
          <div v-if="unlockedAchievements.length > 0" class="badge-preview">
            <p class="badge-preview-label">Derniers badges</p>
            <div class="badge-preview-list">
              <div
                v-for="ua in unlockedAchievements.slice(0, 5)"
                :key="ua.id"
                class="badge-preview-item"
                :title="ua.achievement?.name"
                :style="{ background: ua.achievement?.color || '#6366f1' }"
              >
                <Icon :icon="ua.achievement?.icon || 'mdi:medal'" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Middle: Achievements -->
      <div class="column achievements-column">
        <div class="bento-card achievements-card">
          <div class="achievements-header">
            <div class="card-title">
              <Icon icon="mdi:shield-star-outline" class="card-icon" />
              <h3>Badges</h3>
              <span class="badge-count-pill">{{ unlockedAchievements.length }}/{{ achievements.length }}</span>
            </div>
            <div class="tabs">
              <button :class="{ active: currentTab === 'all' }" @click="currentTab = 'all'">Tous</button>
              <button :class="{ active: currentTab === 'unlocked' }" @click="currentTab = 'unlocked'">
                Débloqués
              </button>
              <button :class="{ active: currentTab === 'locked' }" @click="currentTab = 'locked'">
                Verrouillés
              </button>
            </div>
          </div>

          <div class="achievements-grid">
            <div
              v-for="achievement in filteredAchievements"
              :key="achievement.id"
              class="ach-item"
              :class="{ unlocked: isUnlocked(achievement.code), locked: !isUnlocked(achievement.code) }"
            >
              <div
                class="ach-icon"
                :style="{ background: isUnlocked(achievement.code) ? achievement.color : undefined }"
              >
                <Icon :icon="achievement.icon" />
                <div v-if="!isUnlocked(achievement.code)" class="ach-lock">
                  <Icon icon="mdi:lock" />
                </div>
              </div>
              <div class="ach-body">
                <h4>{{ achievement.name }}</h4>
                <p>{{ achievement.description }}</p>
                <div v-if="isUnlocked(achievement.code)" class="ach-date">
                  <Icon icon="mdi:check-circle" />
                  {{ getUnlockedDate(achievement.code) }}
                </div>
                <div v-else class="ach-xp">
                  <Icon icon="mdi:star-outline" />
                  {{ achievement.xp_reward }} XP
                </div>
              </div>
            </div>

            <div v-if="filteredAchievements.length === 0" class="ach-empty">
              <Icon icon="mdi:trophy-broken" />
              <p>Aucun badge dans cette catégorie</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Leaderboard -->
      <div class="column leaderboard-column">
        <div class="bento-card leaderboard-card">
          <div class="card-title mb-4">
            <Icon icon="mdi:format-list-numbered" class="card-icon" />
            <h3>Classement</h3>
          </div>

          <div class="leaderboard-list">
            <div
              v-for="(user, index) in leaderboard"
              :key="user.user_id"
              class="lb-item"
              :class="{
                'lb-gold': index === 0,
                'lb-silver': index === 1,
                'lb-bronze': index === 2,
                'lb-me': user.user_id === authStore.user?.id
              }"
            >
              <!-- Rank -->
              <div class="lb-rank">
                <span v-if="index === 0" class="medal">🥇</span>
                <span v-else-if="index === 1" class="medal">🥈</span>
                <span v-else-if="index === 2" class="medal">🥉</span>
                <span v-else class="rank-num">{{ index + 1 }}</span>
              </div>

              <!-- User info -->
              <div class="lb-user">
                <div class="lb-avatar">
                  <Icon icon="mdi:account-circle" />
                </div>
                <div class="lb-names">
                  <span class="lb-fullname">{{ user.first_name }} {{ user.last_name }}</span>
                  <span class="lb-username">@{{ user.username }}</span>
                  <!-- Badges row -->
                  <div v-if="user.badges && user.badges.length > 0" class="lb-badges">
                    <div
                      v-for="badge in user.badges"
                      :key="badge.code"
                      class="lb-badge"
                      :style="{ background: badge.color }"
                      :title="badge.name"
                    >
                      <Icon :icon="badge.icon" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- XP + level -->
              <div class="lb-score">
                <div class="lb-level">Niv. {{ user.level }}</div>
                <div class="lb-xp">{{ user.xp.toLocaleString() }} <span>XP</span></div>
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
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/fr'
import AOS from 'aos'
import 'aos/dist/aos.css'

dayjs.extend(relativeTime)
dayjs.locale('fr')

const authStore = useAuthStore()

const profile = ref({ level: 1, xp: 0 })
const achievements = ref([])
const unlockedAchievements = ref([])
const leaderboard = ref([])
const recentTransactions = ref([])
const currentTab = ref('all')

const progressPercentage = computed(() => {
  const xp = profile.value.xp
  const lvl = profile.value.level
  const xpCurrent = Math.pow(lvl - 1, 2) * 100
  const xpNext = Math.pow(lvl, 2) * 100
  const range = xpNext - xpCurrent
  if (range === 0) return 100
  return Math.min(Math.max(((xp - xpCurrent) / range) * 100, 0), 100)
})

const xpToNextLevel = computed(() => {
  const xpNext = Math.pow(profile.value.level, 2) * 100
  return Math.max(xpNext - profile.value.xp, 0)
})

const myRank = computed(() => profile.value.rank || null)

const filteredAchievements = computed(() => {
  if (currentTab.value === 'unlocked') return achievements.value.filter(a => isUnlocked(a.code))
  if (currentTab.value === 'locked') return achievements.value.filter(a => !isUnlocked(a.code))
  return achievements.value
})

const fetchData = async () => {
  try {
    const [profileData, allAch, myAch, lb, txs] = await Promise.all([
      gamificationService.getProfile(),
      gamificationService.getAllAchievements(),
      gamificationService.getMyAchievements(),
      gamificationService.getLeaderboard(),
      gamificationService.getTransactions()
    ])
    profile.value = profileData
    achievements.value = allAch
    unlockedAchievements.value = myAch || []
    leaderboard.value = lb
    recentTransactions.value = txs || []
  } catch (e) {
    console.error('Gamification fetch error', e)
  }
}

const isUnlocked = (code) =>
  unlockedAchievements.value.some(ua => ua.achievement?.code === code)

const getUnlockedDate = (code) => {
  const ua = unlockedAchievements.value.find(ua => ua.achievement?.code === code)
  return ua ? dayjs(ua.unlocked_at).format('DD/MM/YYYY') : ''
}

const getActivityIcon = (reason) => ({
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
})[reason] || 'mdi:star-outline'

const getActivityLabel = (reason) => ({
  daily_login: 'Connexion',
  app_click: 'Application',
  news_read: 'Article lu',
  news_publish: 'Publication',
  event_publish: 'Événement',
  poll_vote: 'Sondage voté',
  poll_create: 'Sondage créé',
  comment_create: 'Commentaire',
  suggestion_create: 'Suggestion',
  suggestion_vote: 'Vote idée',
  suggestion_comment: 'Débat',
  achievement_unlock: 'Badge !',
  chat_message: 'Message',
  chatbot_use: 'IA utilisée',
})[reason] || reason

const formatRelative = (date) => dayjs(date).fromNow()

onMounted(() => {
  fetchData()
  AOS.init({ duration: 600, easing: 'ease-out-cubic', once: true })
})
</script>

<style scoped>
/* ── Root ──────────────────────────────────────────────────── */
.gamification-hub {
  padding: 1rem;
  min-height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* ── Header ────────────────────────────────────────────────── */
.hub-header {
  background: white;
  border-radius: 16px;
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 1rem;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
}
.dark .hub-header { background: #1e293b; }

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}
.header-icon-wrap {
  width: 46px; height: 46px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 6px 16px rgba(245,158,11,.35);
  flex-shrink: 0;
}
.header-icon { font-size: 1.625rem; color: white; }
.header-left h1 {
  font-size: 1.375rem; font-weight: 800;
  color: #1e293b; margin: 0;
}
.dark .header-left h1 { color: white; }
.header-left p { color: #64748b; margin: 0; font-size: 0.875rem; }

.header-pills { display: flex; gap: 0.5rem; flex-wrap: wrap; }
.header-pill {
  display: flex; align-items: center; gap: 0.375rem;
  background: var(--bg-surface); border-radius: 999px;
  padding: 0.375rem 0.875rem;
  font-size: 0.8125rem; font-weight: 600; color: var(--text-muted);
}
.header-pill.accent { background: #ede9fe; color: #7c3aed; }
.dark .header-pill.accent { background: #3b0764; color: #c4b5fd; }
.header-pill.gold { background: #fef3c7; color: #92400e; }
.dark .header-pill.gold { background: #451a03; color: #fbbf24; }

/* ── Activity Strip ────────────────────────────────────────── */
.activity-strip {
  background: var(--bg-card);
  border-radius: 14px;
  padding: 0.625rem 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 2px 12px rgba(var(--shadow-rgb),0.05);
  overflow: hidden;
}

.strip-label {
  display: flex; align-items: center; gap: 0.375rem;
  font-size: 0.75rem; font-weight: 700; color: #6366f1;
  white-space: nowrap; flex-shrink: 0;
}
.strip-label-icon { font-size: 1rem; }

.strip-scroll {
  display: flex; gap: 0.5rem;
  overflow-x: auto; flex: 1;
  padding-bottom: 2px;
  scrollbar-width: none;
}
.strip-scroll::-webkit-scrollbar { display: none; }

.strip-item {
  display: flex; align-items: center; gap: 0.5rem;
  background: var(--bg-surface);
  border-radius: 10px;
  padding: 0.375rem 0.75rem;
  white-space: nowrap;
  flex-shrink: 0;
  border: 1px solid var(--border);
  transition: background 0.2s;
}

.strip-icon {
  width: 24px; height: 24px;
  border-radius: 6px;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.875rem;
  background: #e0e7ff; color: #4f46e5;
  flex-shrink: 0;
}
.strip-item.achievement_unlock .strip-icon { background: #fef3c7; color: #d97706; }
.strip-item.daily_login .strip-icon { background: #dcfce7; color: #16a34a; }
.strip-item.news_publish .strip-icon { background: #ede9fe; color: #7c3aed; }
.strip-item.chat_message .strip-icon { background: #cffafe; color: #0891b2; }
.strip-item.chatbot_use .strip-icon { background: #f3e8ff; color: #9333ea; }

.strip-content { display: flex; flex-direction: column; line-height: 1.2; }
.strip-action { font-size: 0.75rem; font-weight: 600; color: var(--text-primary); }
.strip-time { font-size: 0.6875rem; color: #94a3b8; }

.strip-xp {
  font-size: 0.75rem; font-weight: 700;
  color: #10b981;
  margin-left: auto; padding-left: 0.5rem;
}

.strip-empty {
  font-size: 0.8125rem; color: #94a3b8;
  padding: 0.25rem 0;
}

/* ── Main Grid ─────────────────────────────────────────────── */
.hub-grid {
  display: grid;
  grid-template-columns: 270px 1fr 360px;
  gap: 1rem;
  flex: 1;
  align-items: start;
}

@media (max-width: 1280px) {
  .hub-grid { grid-template-columns: 270px 1fr; }
  .leaderboard-column { grid-column: span 2; }
}
@media (max-width: 768px) {
  .hub-grid { grid-template-columns: 1fr; }
  .leaderboard-column { grid-column: 1; }
}

/* ── Bento Card ────────────────────────────────────────────── */
.bento-card {
  background: var(--bg-card);
  border-radius: 16px;
  padding: 1.25rem;
  box-shadow: 0 2px 12px rgba(var(--shadow-rgb),0.06);
  overflow: hidden;
  min-width: 0;
}

.card-title {
  display: flex; align-items: center; gap: 0.625rem;
}
.card-title h3 {
  font-size: 1.125rem; font-weight: 700; margin: 0;
  color: var(--text-primary);
}
.card-icon { font-size: 1.375rem; color: #6366f1; }

/* ── Profile Card ──────────────────────────────────────────── */
.profile-card { text-align: center; }

.avatar-wrap { position: relative; width: 88px; height: 88px; margin: 0 auto 0.875rem; }
.avatar-ring {
  width: 88px; height: 88px; border-radius: 50%;
  background: linear-gradient(135deg, #6366f1, #a855f7);
  padding: 3px;
  box-shadow: 0 0 0 3px #e0e7ff, 0 8px 20px rgba(99,102,241,.3);
}
.dark .avatar-ring { box-shadow: 0 0 0 3px #312e81, 0 8px 20px rgba(99,102,241,.4); }
.avatar-img {
  width: 100%; height: 100%;
  border-radius: 50%; object-fit: cover;
  border: 3px solid white;
}
.dark .avatar-img { border-color: var(--bg-card); }
.avatar-initials {
  width: 100%; height: 100%;
  border-radius: 50%; border: 3px solid white;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.75rem; font-weight: 800; color: white;
  background: transparent;
  text-shadow: 0 2px 4px rgba(0,0,0,.2);
}
.dark .avatar-initials { border-color: var(--bg-card); }
.level-chip {
  position: absolute; bottom: -2px; right: -2px;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  color: white; width: 30px; height: 30px;
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  font-weight: 800; font-size: 0.875rem;
  border: 3px solid white;
  box-shadow: 0 4px 8px rgba(99,102,241,.45);
}
.dark .level-chip { border-color: var(--bg-card); }

.profile-name { font-size: 1.125rem; font-weight: 700; margin: 0 0 0.125rem; color: var(--text-primary); overflow-wrap: break-word; word-break: break-word; max-width: 100%; }
.profile-username { font-size: 0.8125rem; color: #94a3b8; margin-bottom: 1.125rem; overflow-wrap: break-word; word-break: break-word; max-width: 100%; }

/* XP */
.xp-section { margin-bottom: 1.25rem; }
.xp-row { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
.xp-level { font-weight: 700; font-size: 0.8125rem; color: #4f46e5; }
.xp-value { font-weight: 600; font-size: 0.8125rem; color: #64748b; }
.xp-bar-bg {
  height: 10px; background: var(--bg-surface); border-radius: 5px; overflow: hidden; margin-bottom: 0.5rem;
}
.xp-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366f1, #a855f7);
  border-radius: 5px;
  position: relative;
  transition: width 1.2s cubic-bezier(.4,0,.2,1);
}
.xp-bar-shine {
  position: absolute; top: 0; right: 0; bottom: 0; width: 24px;
  background: rgba(255,255,255,.35);
  filter: blur(6px);
}
.xp-next { font-size: 0.6875rem; color: #94a3b8; }

/* Profile Stats */
.profile-stats {
  display: grid;
  grid-template-columns: 1fr 1px 1fr 1px 1fr;
  align-items: center;
  background: var(--bg-surface); border-radius: 12px;
  padding: 0.625rem 0.5rem; margin-bottom: 1.125rem;
  border: 1px solid var(--border);
  overflow: hidden;
}
.pstat { display: flex; flex-direction: column; align-items: center; gap: 0.25rem; min-width: 0; padding: 0 0.25rem; }
.pstat-divider { width: 1px; height: 28px; background: var(--border); justify-self: center; }
.pstat-icon {
  width: 26px; height: 26px; border-radius: 7px;
  display: flex; align-items: center; justify-content: center;
  font-size: 0.875rem; flex-shrink: 0;
}
.pstat-icon.trophy { background: #fef3c7; color: #d97706; }
.pstat-icon.fire   { background: #fee2e2; color: #ef4444; }
.pstat-icon.rank   { background: #dcfce7; color: #16a34a; }
.pstat-value { font-weight: 800; font-size: 0.9375rem; color: var(--text-primary); line-height: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 100%; }
.pstat-label { font-size: 0.625rem; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.04em; }

/* Badge preview */
.badge-preview-label { font-size: 0.75rem; font-weight: 600; color: #64748b; margin-bottom: 0.5rem; text-align: left; }
.badge-preview-list { display: flex; gap: 0.375rem; flex-wrap: wrap; justify-content: flex-start; }
.badge-preview-item {
  width: 32px; height: 32px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 1rem;
  box-shadow: 0 2px 6px rgba(0,0,0,.15);
  transition: transform .2s;
}
.badge-preview-item:hover { transform: scale(1.15); }

/* ── Achievements Card ─────────────────────────────────────── */
.achievements-card { display: flex; flex-direction: column; }
.achievements-header {
  display: flex; align-items: center; justify-content: space-between;
  flex-wrap: wrap; gap: 0.75rem; margin-bottom: 1rem;
}
.badge-count-pill {
  background: #ede9fe; color: #7c3aed;
  font-size: 0.75rem; font-weight: 700;
  padding: 0.125rem 0.625rem; border-radius: 999px; margin-left: 0.25rem;
}
.dark .badge-count-pill { background: #3b0764; color: #c4b5fd; }

.tabs {
  display: flex; background: var(--bg-surface); border-radius: 10px; padding: 0.2rem; gap: 0.1rem;
}
.tabs button {
  padding: 0.3rem 0.75rem; border-radius: 8px; border: none;
  background: transparent; font-size: 0.75rem; font-weight: 600;
  color: var(--text-muted); cursor: pointer; transition: all .2s;
}
.tabs button.active {
  background: var(--bg-card); color: var(--accent);
  box-shadow: 0 1px 4px rgba(var(--shadow-rgb),.08);
}

.achievements-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 0.75rem;
  max-height: 580px; overflow-y: auto;
  padding-right: 2px;
}
.achievements-grid::-webkit-scrollbar { width: 4px; }
.achievements-grid::-webkit-scrollbar-track { background: transparent; }
.achievements-grid::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 2px; }
.dark .achievements-grid::-webkit-scrollbar-thumb { background: #334155; }

.ach-item {
  display: flex; gap: 0.75rem; align-items: flex-start;
  padding: 0.875rem; border-radius: 12px;
  background: var(--bg-surface); border: 1px solid var(--border);
  transition: all 0.25s ease;
  cursor: default;
}
.ach-item.unlocked:hover { transform: translateY(-3px); box-shadow: 0 6px 16px rgba(0,0,0,.08); }
.ach-item.locked { opacity: 0.55; filter: grayscale(0.7); }

.ach-icon {
  width: 48px; height: 48px; border-radius: 12px;
  display: flex; align-items: center; justify-content: center;
  font-size: 1.625rem; color: white; flex-shrink: 0;
  position: relative;
  background: #cbd5e1;
  box-shadow: 0 3px 8px rgba(0,0,0,.12);
}
.ach-lock {
  position: absolute; inset: 0; border-radius: 12px;
  background: rgba(0,0,0,.25);
  display: flex; align-items: center; justify-content: center;
  font-size: 1rem; color: white;
}

.ach-body h4 { margin: 0 0 0.2rem; font-size: 0.875rem; font-weight: 700; color: var(--text-primary); }
.ach-body p { margin: 0; font-size: 0.75rem; color: #64748b; line-height: 1.35; }

.ach-date {
  display: flex; align-items: center; gap: 0.25rem;
  font-size: 0.6875rem; color: #10b981; font-weight: 600; margin-top: 0.375rem;
}
.ach-xp {
  display: flex; align-items: center; gap: 0.25rem;
  font-size: 0.6875rem; color: #6366f1; font-weight: 600; margin-top: 0.375rem;
}

.ach-empty {
  grid-column: 1/-1; text-align: center; padding: 2rem;
  color: #94a3b8; font-size: 0.875rem;
}
.ach-empty svg { font-size: 2.5rem; margin-bottom: 0.5rem; display: block; margin-inline: auto; }

/* ── Leaderboard ───────────────────────────────────────────── */
.leaderboard-card {}
.leaderboard-list { display: flex; flex-direction: column; gap: 0.5rem; }

.lb-item {
  display: flex; align-items: center; gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-surface); border-radius: 12px;
  border: 1.5px solid var(--border);
  transition: all .2s;
}
.lb-item:hover { transform: translateX(2px); }

.lb-item.lb-gold   { border-color: #fbbf24; background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%); }
.lb-item.lb-silver { border-color: #94a3b8; background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%); }
.lb-item.lb-bronze { border-color: #d97706; background: linear-gradient(135deg, #fff7ed 0%, #fef3c7 100%); }
.dark .lb-item.lb-gold   { background: linear-gradient(135deg, #451a03 0%, #78350f 100%); border-color: #d97706; }
.dark .lb-item.lb-silver { background: linear-gradient(135deg, var(--bg-page) 0%, var(--bg-card) 100%); border-color: #64748b; }
.dark .lb-item.lb-bronze { background: linear-gradient(135deg, var(--bg-page) 0%, var(--bg-surface) 100%); border-color: #92400e; }
.lb-item.lb-me { border-color: #6366f1 !important; box-shadow: 0 0 0 2px #e0e7ff; }
.dark .lb-item.lb-me { box-shadow: 0 0 0 2px #312e81; }

.lb-rank { width: 28px; text-align: center; flex-shrink: 0; }
.medal { font-size: 1.375rem; }
.rank-num { font-size: 0.875rem; font-weight: 800; color: #94a3b8; }

.lb-user { display: flex; align-items: center; gap: 0.625rem; flex: 1; min-width: 0; }
.lb-avatar {
  width: 34px; height: 34px; border-radius: 9px;
  background: var(--bg-surface); color: var(--text-muted);
  display: flex; align-items: center; justify-content: center;
  font-size: 1.25rem; flex-shrink: 0;
}

.lb-names { display: flex; flex-direction: column; min-width: 0; gap: 0.125rem; }
.lb-fullname { font-size: 0.8125rem; font-weight: 700; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.lb-username { font-size: 0.6875rem; color: #94a3b8; }

.lb-badges { display: flex; gap: 0.25rem; flex-wrap: wrap; margin-top: 0.25rem; }
.lb-badge {
  width: 20px; height: 20px; border-radius: 5px;
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 0.75rem;
  box-shadow: 0 1px 4px rgba(0,0,0,.2);
  flex-shrink: 0;
  cursor: default;
  transition: transform .15s;
}
.lb-badge:hover { transform: scale(1.25); }

.lb-score { text-align: right; flex-shrink: 0; }
.lb-level { font-size: 0.6875rem; font-weight: 700; color: #4f46e5; }
.lb-xp { font-size: 0.875rem; font-weight: 800; color: var(--text-primary); }
.lb-xp span { font-size: 0.6875rem; font-weight: 600; color: #94a3b8; }

.mb-4 { margin-bottom: 1rem; }
</style>
