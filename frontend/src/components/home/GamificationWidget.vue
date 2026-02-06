<template>
  <div class="bento-card gamification-widget h-full flex flex-col" data-aos="fade-up">
    <div class="widget-header flex items-center justify-between mb-4">
      <div class="flex items-center gap-2">
        <Icon icon="mdi:medal" class="text-amber-500 text-xl" />
        <h3 class="font-bold text-gray-800 dark:text-white uppercase tracking-wider text-sm">
          Mon Niveau
        </h3>
      </div>
      <router-link 
        to="/gamification" 
        class="text-xs text-blue-500 hover:text-blue-600 font-semibold flex items-center gap-1 transition-all"
      >
        Détails
        <Icon icon="mdi:chevron-right" />
      </router-link>
    </div>

    <!-- Progress Overview -->
    <div class="flex items-center gap-4 mb-6">
      <div class="relative flex-shrink-0">
        <div class="level-circle">
          <span class="level-value">{{ summary.level }}</span>
        </div>
        <div class="xp-badge">
          {{ summary.xp }} XP
        </div>
      </div>
      
      <div class="flex-1">
        <div class="flex justify-between items-end mb-1">
          <span class="text-xs text-gray-500 dark:text-gray-400">Progression</span>
          <span class="text-xs font-bold text-blue-500">{{ summary.progress_percent }}%</span>
        </div>
        <div class="progress-bar-bg">
          <div 
            class="progress-bar-fill" 
            :style="{ width: summary.progress_percent + '%' }"
          ></div>
        </div>
        <p class="text-[10px] text-gray-400 mt-1 italic">
          Plus que {{ summary.next_level_xp - summary.xp }} XP avant le niveau {{ summary.level + 1 }}
        </p>
      </div>
    </div>

    <!-- Recent Badges -->
    <div v-if="summary.recent_badges?.length" class="mt-auto">
      <p class="text-[10px] text-gray-500 dark:text-gray-400 uppercase font-bold mb-2">
        Derniers badges obtenus
      </p>
      <div class="flex flex-wrap gap-2">
        <div 
          v-for="badge in summary.recent_badges" 
          :key="badge.id"
          class="mini-badge-icon"
          :style="{ backgroundColor: badge.color + '22' }"
          :title="badge.name"
        >
          <Icon :icon="badge.icon" :style="{ color: badge.color }" />
        </div>
        <div 
          v-if="summary.recent_badges.length === 0" 
          class="text-xs text-gray-400 italic py-1"
        >
          Aucun badge pour le moment
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'

defineProps({
  summary: {
    type: Object,
    required: true,
    default: () => ({
      level: 1,
      xp: 0,
      next_level_xp: 100,
      progress_percent: 0,
      recent_badges: []
    })
  }
})
</script>

<style scoped>
.gamification-widget {
  background: white;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.dark .gamification-widget {
  background: #1e293b;
}

.gamification-widget:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.05);
}

.level-circle {
  width: 54px;
  height: 54px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(79, 70, 229, 0.3);
  position: relative;
  border: 3px solid white;
}

.dark .level-circle {
  border-color: #1e293b;
}

.level-value {
  color: white;
  font-weight: 800;
  font-size: 1.5rem;
}

.xp-badge {
  position: absolute;
  bottom: -6px;
  right: -10px;
  background: #F59E0B;
  color: white;
  font-size: 9px;
  font-weight: 800;
  padding: 2px 6px;
  border-radius: 10px;
  border: 2px solid white;
}

.dark .xp-badge {
  border-color: #1e293b;
}

.progress-bar-bg {
  height: 10px;
  background: #f1f5f9;
  border-radius: 5px;
  overflow: hidden;
}

.dark .progress-bar-bg {
  background: #334155;
}

.progress-bar-fill {
  height: 100%;
  background: linear-gradient(90deg, #6366f1 0%, #a855f7 100%);
  border-radius: 5px;
  transition: width 1s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.mini-badge-icon {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  transition: transform 0.2s ease;
  cursor: pointer;
}

.mini-badge-icon:hover {
  transform: scale(1.1);
}
</style>
