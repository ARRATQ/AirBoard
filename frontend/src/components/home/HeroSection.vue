<template>
  <div class="hero-section">
    <!-- Animated Background -->
    <div class="hero-background">
      <div class="gradient-orb orb-1"></div>
      <div class="gradient-orb orb-2"></div>
      <div class="gradient-orb orb-3"></div>
    </div>

    <!-- Hero Content -->
    <div class="hero-content">
      <!-- User Avatar & Welcome -->
      <div class="hero-header">
        <div class="user-avatar-wrapper">
          <div class="user-avatar" :class="{ 'has-avatar': !!authStore.user?.avatar_url }">
            <template v-if="authStore.user?.avatar_url">
              <img :src="authStore.user.avatar_url" alt="Avatar" class="avatar-image" />
            </template>
            <template v-else>
              <span class="avatar-initials">{{ authStore.userInitials }}</span>
            </template>
            <div class="avatar-status"></div>
          </div>
          <div class="welcome-text">
            <h1 class="hero-title">
              {{ getGreeting() }},
              <span class="user-name">{{ userName }}</span>
              <span class="wave-emoji"></span>
            </h1>
            <transition name="fade-message" mode="out-in">
              <p :key="dynamicMessage" class="hero-subtitle">{{ dynamicMessage }}</p>
            </transition>
          </div>
        </div>

        <!-- Quick Stats Bar (Mini) -->
        <div v-if="showQuickStats" class="quick-stats-bar">
          <div class="stat-badge">
            <Icon icon="mdi:apps" class="stat-icon" />
            <span class="stat-value">{{ stats.apps || 0 }}</span>
            <span class="stat-label">{{ $t('home.hero.apps') }}</span>
          </div>
          <div class="stat-badge">
            <Icon icon="mdi:calendar-star" class="stat-icon" />
            <span class="stat-value">{{ stats.events || 0 }}</span>
            <span class="stat-label">{{ $t('home.hero.events') }}</span>
          </div>
          <div class="stat-badge">
            <Icon icon="mdi:newspaper-variant-outline" class="stat-icon" />
            <span class="stat-value">{{ stats.news || 0 }}</span>
            <span class="stat-label">{{ $t('home.hero.news') }}</span>
          </div>
          <div class="stat-badge">
            <Icon icon="mdi:poll" class="stat-icon" />
            <span class="stat-value">{{ stats.polls || 0 }}</span>
            <span class="stat-label">{{ $t('home.hero.polls') }}</span>
          </div>
        </div>
      </div>

      <!-- Announcements Carousel (Integrated) -->
      <div v-if="announcements.length > 0" class="announcements-compact">
        <transition :name="slideDirection" mode="out-in">
          <div
            :key="currentAnnouncementIndex"
            :class="getAnnouncementClass(currentAnnouncement.type)"
            class="announcement-card"
          >
            <Icon :icon="getAnnouncementIcon(currentAnnouncement.type)" class="announcement-icon" />
            <div class="announcement-content">
              <h3 class="announcement-title">{{ currentAnnouncement.title }}</h3>
              <p v-if="currentAnnouncement.content" class="announcement-text">
                {{ currentAnnouncement.content }}
              </p>
            </div>
            <!-- Navigation Dots -->
            <div v-if="announcements.length > 1" class="announcement-dots">
              <button
                v-for="(_, index) in announcements"
                :key="index"
                @click="goToAnnouncement(index)"
                :class="index === currentAnnouncementIndex ? 'dot-active' : 'dot-inactive'"
                class="announcement-dot"
              />
            </div>
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Icon } from '@iconify/vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  stats: {
    type: Object,
    default: () => ({})
  },
  showQuickStats: {
    type: Boolean,
    default: true
  },
  announcements: {
    type: Array,
    default: () => []
  },
  heroMessages: {
    type: Array,
    default: () => []
  }
})

const authStore = useAuthStore()
const { t } = useI18n()

const currentAnnouncementIndex = ref(0)
const slideDirection = ref('slide-left')
let announcementInterval = null

const userName = computed(() => {
  return authStore.user?.first_name || authStore.user?.username || 'User'
})

// Messages par défaut (Fallback si aucun message n'est configuré en base)
const staticMessages = [
  "Le plus grand risque est de ne prendre aucun risque.",
  "La connaissance est une lumière que Dieu place dans le cœur.",
  "Le succès n'est pas final, l'échec n'est pas fatal : c'est le courage de continuer qui compte.",
  "Un sourire est une aumône.",
  "Votre travail va occuper une grande partie de votre vie, la seule façon d'être vraiment satisfait est de faire ce que vous croyez être un excellent travail.",
  "Cherchez le savoir du berceau jusqu'au tombeau.",
  "Agissez comme s'il était impossible d'échouer.",
  "Celui qui ne remercie pas les gens ne remercie pas Dieu.",
  "La patience est la clé de la réussite.",
  "Le meilleur d'entre vous est celui qui est le plus utile aux autres."
]

const dynamicMessage = ref('')
let messageInterval = null

const setRandomMessage = () => {
  const sourceMessages = props.heroMessages?.length > 0 
    ? props.heroMessages.map(m => m.content)
    : staticMessages
    
  const randomIndex = Math.floor(Math.random() * sourceMessages.length)
  dynamicMessage.value = sourceMessages[randomIndex]
}


onMounted(() => {
  setRandomMessage()
  // Changer de message toutes le 10 secondes
  messageInterval = setInterval(setRandomMessage, 10000)
})

const currentAnnouncement = computed(() => {
  return props.announcements[currentAnnouncementIndex.value] || {}
})

const getGreeting = () => {
  const hour = new Date().getHours()
  if (hour < 12) return t('home.hero.greeting.morning')
  if (hour < 18) return t('home.hero.greeting.afternoon')
  return t('home.hero.greeting.evening')
}

const goToAnnouncement = (index) => {
  slideDirection.value = index > currentAnnouncementIndex.value ? 'slide-left' : 'slide-right'
  currentAnnouncementIndex.value = index
}

const nextAnnouncement = () => {
  slideDirection.value = 'slide-left'
  currentAnnouncementIndex.value = (currentAnnouncementIndex.value + 1) % props.announcements.length
}

const getAnnouncementClass = (type) => {
  const classes = {
    info: 'announcement-info',
    warning: 'announcement-warning',
    success: 'announcement-success',
    error: 'announcement-error'
  }
  return classes[type] || classes.info
}

const getAnnouncementIcon = (type) => {
  const icons = {
    info: 'mdi:information',
    warning: 'mdi:alert',
    success: 'mdi:check-circle',
    error: 'mdi:alert-circle'
  }
  return icons[type] || icons.info
}

onMounted(() => {
  if (props.announcements.length > 1) {
    announcementInterval = setInterval(nextAnnouncement, 5000)
  }
})

onUnmounted(() => {
  if (announcementInterval) clearInterval(announcementInterval)
  if (messageInterval) clearInterval(messageInterval)
})
</script>

<style scoped>
.hero-section {
  position: relative;
  padding: 2rem 2rem 1.5rem;
  margin: 0 0 1.5rem;
  overflow: hidden;
  border-radius: 0rem;
  min-height: auto;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

/* Animated Background */
.hero-background {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  overflow: hidden;
}

.dark .hero-background {
  background: linear-gradient(135deg, #1a1f3a 0%, #2d1b4e 50%, #3a1f5d 100%);
}

.gradient-orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
  opacity: 0.6;
  animation: float 20s ease-in-out infinite;
}

.orb-1 {
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.8) 0%, transparent 70%);
  top: -150px;
  left: -100px;
  animation-delay: 0s;
}

.orb-2 {
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(236, 72, 153, 0.8) 0%, transparent 70%);
  top: 50%;
  right: -50px;
  animation-delay: -7s;
}

.orb-3 {
  width: 350px;
  height: 350px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.8) 0%, transparent 70%);
  bottom: -100px;
  left: 40%;
  animation-delay: -14s;
}

@keyframes float {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(30px, -30px) scale(1.1);
  }
  66% {
    transform: translate(-20px, 20px) scale(0.9);
  }
}

/* Hero Content */
.hero-content {
  position: relative;
  z-index: 1;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
  overflow-x: hidden;
}

/* Optimisation spécifique pour 1280x720 */
@media (width: 1280px) and (height: 720px) {
  .hero-section {
    padding: 1.25rem 1.25rem 1rem;
    min-height: auto;
  }
  
  .hero-content {
    max-width: 1100px;
  }
  
  .hero-header {
    margin-bottom: 1rem;
    gap: 1.25rem;
  }
  
  .hero-title {
    font-size: 1.65rem;
    line-height: 1.1;
  }
  
  .hero-subtitle {
    margin-top: 0.25rem;
    font-size: 0.9rem;
  }
  
  .quick-stats-bar {
    gap: 0.75rem;
  }
  
  .stat-badge {
    padding: 0.6rem 1rem;
    border-radius: 12px;
  }
  
  .stat-value {
    font-size: 1.1rem;
  }
  
  .stat-label {
    font-size: 0.7rem;
  }
  
  .announcements-compact {
    margin-top: 1rem;
  }
  
  .announcement-card {
    padding: 0.875rem 1rem;
  }
  
  .announcement-title {
    font-size: 0.8rem;
  }
  
  .announcement-text {
    font-size: 0.75rem;
  }
}

/* User Avatar & Welcome */
.hero-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 1.5rem;
  margin-bottom: 1.5rem;
  width: 100%;
  box-sizing: border-box;
  overflow-x: hidden;
}

.user-avatar-wrapper {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex: 1;
  min-width: 0;
  max-width: 100%;
}

.user-avatar {
  position: relative;
  width: 40px;
  height: 40px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  transition: transform 0.3s ease;
}

.user-avatar:hover {
  transform: scale(1.05) rotate(-5deg);
}

.avatar-icon {
  font-size: 2.5rem;
  color: white;
}

.avatar-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

.avatar-initials {
  font-size: 1.125rem;
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
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.avatar-status {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 14px;
  height: 14px;
  background: #10b981;
  border: 3px solid white;
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.8;
    transform: scale(1.1);
  }
}

.welcome-text {
  color: white;
  flex: 1;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

.hero-title {
  font-size: 2rem;
  font-weight: 700;
  margin: 0;
  line-height: 1.2;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  word-break: break-word;
  overflow-wrap: break-word;
}

.user-name {
  background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.wave-emoji {
  display: inline-block;
  animation: wave 2.5s ease-in-out infinite;
  transform-origin: 70% 70%;
}

@keyframes wave {
  0%, 100% { transform: rotate(0deg); }
  10%, 30% { transform: rotate(14deg); }
  20% { transform: rotate(-8deg); }
  40%, 100% { transform: rotate(0deg); }
}

.hero-subtitle {
  margin: 0.5rem 0 0;
  font-size: 1rem;
  opacity: 0.9;
  font-weight: 400;
}

/* Quick Stats Bar */
.quick-stats-bar {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  max-width: 100%;
  overflow-x: hidden;
}

.stat-badge {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: white;
  transition: all 0.3s ease;
}

.stat-badge:hover {
  background: rgba(255, 255, 255, 0.25);
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 1.5rem;
  opacity: 0.9;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  line-height: 1;
}

.stat-label {
  font-size: 0.75rem;
  opacity: 0.8;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* Announcements Compact */
.announcements-compact {
  margin-top: 1.5rem;
  position: relative;
}

.announcement-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: 16px;
  border-left: 4px solid;
  backdrop-filter: blur(10px);
  position: relative;
  transition: all 0.3s ease;
}

.announcement-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.announcement-content {
  flex: 1;
  min-width: 0;
}

.announcement-title {
  font-size: 0.875rem;
  font-weight: 600;
  margin: 0 0 0.25rem;
  line-height: 1.4;
}

.announcement-text {
  font-size: 0.8125rem;
  margin: 0;
  line-height: 1.5;
  opacity: 0.95;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.announcement-dots {
  display: flex;
  gap: 0.375rem;
  position: absolute;
  bottom: 0.75rem;
  right: 1rem;
}

.announcement-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  padding: 0;
}

.dot-active {
  width: 1.25rem;
  background: currentColor;
  opacity: 1;
}

.dot-inactive {
  background: currentColor;
  opacity: 0.3;
}

.dot-inactive:hover {
  opacity: 0.6;
}

/* Announcement Types */
.announcement-info {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.8);
  color: rgba(255, 255, 255, 0.95);
}

.announcement-warning {
  background: rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.8);
  color: rgba(255, 255, 255, 0.95);
}

.announcement-success {
  background: rgba(16, 185, 129, 0.15);
  border-color: rgba(16, 185, 129, 0.8);
  color: rgba(255, 255, 255, 0.95);
}

.announcement-error {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.8);
  color: rgba(255, 255, 255, 0.95);
}

/* Slide Transitions */
.slide-left-enter-active,
.slide-left-leave-active,
.slide-right-enter-active,
.slide-right-leave-active {
  transition: all 0.4s ease;
}

.slide-left-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.slide-left-leave-to {
  transform: translateX(-100%);
  opacity: 0;
}

.slide-right-enter-from {
  transform: translateX(-100%);
  opacity: 0;
}

.slide-right-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

/* Fade Message Transition */
.fade-message-enter-active,
.fade-message-leave-active {
  transition: opacity 0.5s ease, transform 0.5s ease;
}

.fade-message-enter-from,
.fade-message-leave-to {
  opacity: 0;
  transform: translateY(10px);
}

/* Optimisation pour 1280x720 et résolutions similaires */
@media (max-width: 1366px) {
  .hero-section {
    padding: 1.5rem 1.5rem 1.25rem;
  }
  
  .hero-content {
    max-width: 1000px;
  }
  
  .hero-title {
    font-size: 1.75rem;
  }
  
  .hero-subtitle {
    font-size: 0.95rem;
  }
  
  /* Réduire la taille des éléments décoratifs */
  .orb-1 {
    width: 300px;
    height: 300px;
    top: -100px;
    left: -50px;
  }
  
  .orb-2 {
    width: 250px;
    height: 250px;
    right: -30px;
  }
  
  .orb-3 {
    width: 280px;
    height: 280px;
    bottom: -80px;
  }
}

@media (max-width: 1024px) {
  .hero-content {
    max-width: 800px;
  }
  
  .hero-title {
    font-size: 1.6rem;
  }
  
  .hero-header {
    gap: 1rem;
    margin-bottom: 1.25rem;
  }
  
  .quick-stats-bar {
    gap: 0.75rem;
  }
  
  .stat-badge {
    padding: 0.6rem 1rem;
  }
  
  /* Réduire encore plus les éléments décoratifs */
  .orb-1 {
    width: 250px;
    height: 250px;
    top: -80px;
    left: -30px;
  }
  
  .orb-2 {
    width: 200px;
    height: 200px;
    right: -20px;
  }
  
  .orb-3 {
    width: 230px;
    height: 230px;
    bottom: -60px;
  }
}

/* Tablettes et petites résolutions */
@media (max-width: 1024px) and (min-width: 769px) {
  .hero-section {
    padding: 1.5rem 1.25rem 1.25rem;
    margin: 0 0 1.5rem;
    border-radius: 1.5rem;
  }

  .hero-content {
    max-width: 100%;
    padding: 0;
  }

  .hero-header {
    gap: 1rem;
    margin-bottom: 1.25rem;
  }

  .hero-title {
    font-size: 1.65rem;
  }

  .quick-stats-bar {
    gap: 0.75rem;
  }

  .stat-badge {
    padding: 0.65rem 1rem;
    font-size: 0.9rem;
  }
}

@media (max-width: 768px) {
  .hero-section {
    padding: 1.25rem 1rem 1rem;
    margin: 0 0 1.25rem;
    border-radius: 1.5rem;
  }

  .hero-content {
    max-width: 100%;
    padding: 0;
  }

  .hero-header {
    flex-direction: column;
    align-items: flex-start;
    margin-bottom: 1rem;
    gap: 0.75rem;
  }

  .user-avatar-wrapper {
    max-width: 100%;
  }

  .hero-title {
    font-size: 1.35rem;
  }

  .hero-subtitle {
    font-size: 0.85rem;
  }

  .user-avatar {
    width: 48px;
    height: 48px;
    flex-shrink: 0;
  }

  .avatar-icon {
    font-size: 2rem;
  }

  .quick-stats-bar {
    width: 100%;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .stat-badge {
    flex: 1;
    flex-direction: column;
    text-align: center;
    padding: 0.65rem 0.4rem;
    min-width: 70px;
    max-width: calc(33.333% - 0.35rem);
    gap: 0.25rem;
    box-sizing: border-box;
  }

  .stat-icon {
    font-size: 1.25rem;
  }

  .stat-value {
    font-size: 1.1rem;
  }

  .stat-label {
    font-size: 0.65rem;
  }

  /* Éléments décoratifs encore plus petits sur mobile */
  .orb-1 {
    width: 200px;
    height: 200px;
    top: -60px;
    left: -20px;
  }

  .orb-2 {
    width: 150px;
    height: 150px;
    right: -15px;
  }

  .orb-3 {
    width: 180px;
    height: 180px;
    bottom: -40px;
  }

  .announcements-compact {
    margin-top: 1rem;
  }

  .announcement-card {
    padding: 0.85rem 0.75rem;
    gap: 0.75rem;
  }

  .announcement-title {
    font-size: 0.8rem;
  }

  .announcement-text {
    font-size: 0.75rem;
  }
}

/* Téléphones très petits */
@media (max-width: 480px) {
  .hero-section {
    padding: 1rem 0.75rem 0.875rem;
    margin: 0 0 1rem;
    border-radius: 1.25rem;
  }

  .hero-title {
    font-size: 1.2rem;
    line-height: 1.3;
  }

  .hero-subtitle {
    font-size: 0.8rem;
  }

  .user-avatar {
    width: 40px;
    height: 40px;
    flex-shrink: 0;
  }

  .avatar-icon {
    font-size: 1.75rem;
  }

  .quick-stats-bar {
    gap: 0.4rem;
    width: 100%;
  }

  .stat-badge {
    padding: 0.5rem 0.3rem;
    min-width: 60px;
    max-width: calc(33.333% - 0.3rem);
    border-radius: 12px;
    box-sizing: border-box;
  }

  .stat-icon {
    font-size: 1.1rem;
  }

  .stat-value {
    font-size: 1rem;
  }

  .stat-label {
    font-size: 0.6rem;
  }
}
</style>
