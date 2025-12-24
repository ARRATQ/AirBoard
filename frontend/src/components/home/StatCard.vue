<template>
  <div class="stat-card" :class="colorClass">
    <div class="stat-card-inner">
      <!-- Icon with animated background -->
      <div class="stat-icon-wrapper">
        <div class="stat-icon-bg"></div>
        <Icon :icon="icon" class="stat-icon" />
      </div>

      <!-- Content -->
      <div class="stat-content">
        <p class="stat-label">{{ label }}</p>
        <p class="stat-value">
          <span ref="counterEl">{{ displayValue }}</span>
        </p>
      </div>

      <!-- Trend Indicator (optional) -->
      <div v-if="trend" class="stat-trend" :class="trendClass">
        <Icon :icon="trendIcon" class="trend-icon" />
        <span class="trend-value">{{ trend }}</span>
      </div>
    </div>

    <!-- Animated gradient background -->
    <div class="stat-card-gradient"></div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  icon: { type: String, required: true },
  value: { type: Number, required: true },
  label: { type: String, required: true },
  color: { type: String, default: 'blue' },
  trend: { type: String, default: null } // e.g., '+12%' or '-5%'
})

const counterEl = ref(null)
const displayValue = ref(0)

// Color classes
const colorClass = computed(() => {
  const colors = {
    blue: 'stat-card-blue',
    purple: 'stat-card-purple',
    green: 'stat-card-green',
    orange: 'stat-card-orange',
    indigo: 'stat-card-indigo',
    cyan: 'stat-card-cyan',
    teal: 'stat-card-teal',
    pink: 'stat-card-pink',
    red: 'stat-card-red'
  }
  return colors[props.color] || colors.blue
})

// Trend icon and class
const trendIcon = computed(() => {
  if (!props.trend) return null
  return props.trend.startsWith('+') ? 'mdi:trending-up' : 'mdi:trending-down'
})

const trendClass = computed(() => {
  if (!props.trend) return ''
  return props.trend.startsWith('+') ? 'trend-positive' : 'trend-negative'
})

// Animated counter
const animateCounter = (target) => {
  const duration = 1500 // 1.5 seconds
  const start = displayValue.value
  const diff = target - start
  const startTime = Date.now()

  const updateCounter = () => {
    const elapsed = Date.now() - startTime
    const progress = Math.min(elapsed / duration, 1)

    // Easing function (ease-out)
    const easeOut = 1 - Math.pow(1 - progress, 3)

    displayValue.value = Math.round(start + diff * easeOut)

    if (progress < 1) {
      requestAnimationFrame(updateCounter)
    }
  }

  requestAnimationFrame(updateCounter)
}

// Watch for value changes
watch(() => props.value, (newValue) => {
  animateCounter(newValue)
}, { immediate: false })

// Initialize on mount
onMounted(() => {
  animateCounter(props.value)
})
</script>

<style scoped>
.stat-card {
  position: relative;
  border-radius: 16px;
  padding: 1.5rem;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  cursor: pointer;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.stat-card::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.1) 0%, rgba(255, 255, 255, 0.05) 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-card:hover {
  transform: translateY(-4px) scale(1.02);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.15);
}

.dark .stat-card:hover {
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.4);
}

.stat-card-inner {
  position: relative;
  z-index: 1;
}

/* Icon Wrapper */
.stat-icon-wrapper {
  position: relative;
  width: 56px;
  height: 56px;
  margin-bottom: 1rem;
}

.stat-icon-bg {
  position: absolute;
  inset: 0;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px);
  animation: pulse-glow 3s ease-in-out infinite;
}

.stat-icon {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 100%;
  font-size: 2rem;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.2));
}

@keyframes pulse-glow {
  0%, 100% {
    opacity: 0.8;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.05);
  }
}

/* Content */
.stat-content {
  margin-bottom: 0.5rem;
}

.stat-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 0.5rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-value {
  font-size: 2.25rem;
  font-weight: 700;
  color: white;
  line-height: 1;
  font-variant-numeric: tabular-nums;
}

/* Trend Indicator */
.stat-trend {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.25rem 0.625rem;
  border-radius: 8px;
  font-size: 0.75rem;
  font-weight: 600;
  margin-top: 0.75rem;
}

.trend-positive {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.trend-negative {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.trend-icon {
  font-size: 1rem;
}

/* Animated Gradient Background */
.stat-card-gradient {
  position: absolute;
  inset: 0;
  z-index: 0;
  opacity: 1;
  transition: opacity 0.3s ease;
}

.stat-card:hover .stat-card-gradient {
  opacity: 0.9;
}

/* Color Variants */
.stat-card-blue .stat-card-gradient {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.stat-card-purple .stat-card-gradient {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
}

.stat-card-green .stat-card-gradient {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.stat-card-orange .stat-card-gradient {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.stat-card-indigo .stat-card-gradient {
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
}

.stat-card-cyan .stat-card-gradient {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
}

.stat-card-teal .stat-card-gradient {
  background: linear-gradient(135deg, #14b8a6 0%, #0d9488 100%);
}

.stat-card-pink .stat-card-gradient {
  background: linear-gradient(135deg, #ec4899 0%, #db2777 100%);
}

.stat-card-red .stat-card-gradient {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
}

/* Dark mode adjustments */
.dark .stat-card {
  border-color: rgba(255, 255, 255, 0.05);
}

/* Responsive */
@media (max-width: 640px) {
  .stat-card {
    padding: 1.25rem;
  }

  .stat-icon-wrapper {
    width: 48px;
    height: 48px;
    margin-bottom: 0.875rem;
  }

  .stat-icon {
    font-size: 1.75rem;
  }

  .stat-value {
    font-size: 2rem;
  }

  .stat-label {
    font-size: 0.8125rem;
  }
}
</style>
