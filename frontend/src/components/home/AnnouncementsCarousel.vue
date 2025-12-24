<template>
  <div class="card relative overflow-hidden">
    <transition :name="slideDirection" mode="out-in">
      <div
        :key="currentIndex"
        :class="getAnnouncementClass(currentAnnouncement.type)"
        class="p-4 border-l-4"
      >
        <div class="flex items-start gap-3">
          <Icon :icon="getAnnouncementIcon(currentAnnouncement.type)" class="h-5 w-5 flex-shrink-0 mt-0.5" />
          <div class="flex-1 min-w-0">
            <h3 class="font-semibold text-sm mb-1">{{ currentAnnouncement.title }}</h3>
            <p v-if="currentAnnouncement.content" class="text-sm whitespace-pre-wrap line-clamp-3">
              {{ currentAnnouncement.content }}
            </p>
          </div>
        </div>
      </div>
    </transition>

    <!-- Navigation -->
    <div v-if="announcements.length > 1" class="flex items-center justify-between px-4 pb-3">
      <div class="flex gap-2">
        <button
          v-for="(_, index) in announcements"
          :key="index"
          @click="goTo(index)"
          :class="index === currentIndex ? 'w-6 h-2 bg-blue-600' : 'w-2 h-2 bg-gray-300 dark:bg-gray-600'"
          class="rounded-full transition-all"
          :aria-label="`Go to announcement ${index + 1}`"
        />
      </div>
      <span class="text-xs text-gray-600 dark:text-gray-400">
        {{ currentIndex + 1 }}/{{ announcements.length }}
      </span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  announcements: {
    type: Array,
    required: true
  }
})

const currentIndex = ref(0)
const slideDirection = ref('slide-left')
let interval = null

const currentAnnouncement = computed(() => {
  return props.announcements[currentIndex.value] || {}
})

const goTo = (index) => {
  slideDirection.value = index > currentIndex.value ? 'slide-left' : 'slide-right'
  currentIndex.value = index
}

const next = () => {
  slideDirection.value = 'slide-left'
  currentIndex.value = (currentIndex.value + 1) % props.announcements.length
}

const getAnnouncementClass = (type) => {
  const classes = {
    info: 'bg-blue-50 dark:bg-blue-900/20 border-blue-500 text-blue-900 dark:text-blue-100',
    warning: 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-500 text-yellow-900 dark:text-yellow-100',
    success: 'bg-green-50 dark:bg-green-900/20 border-green-500 text-green-900 dark:text-green-100',
    error: 'bg-red-50 dark:bg-red-900/20 border-red-500 text-red-900 dark:text-red-100'
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
    interval = setInterval(next, 5000)
  }
})

onUnmounted(() => {
  if (interval) clearInterval(interval)
})
</script>

<style scoped>
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

.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
