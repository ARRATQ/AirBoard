<template>
  <button
    @click="cycleZoom"
    class="flex items-center gap-2 px-3 py-2 rounded-lg bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors text-sm font-medium text-gray-700 dark:text-gray-300"
    :title="currentLevel.label"
    :aria-label="currentLevel.label"
  >
    <Icon :icon="currentLevel.icon" class="h-4 w-4" aria-hidden="true" />
    <span class="text-xs">{{ isAuto ? $t('dashboard.zoom.autoShort', { percent: autoPercent }) : `${modelValue}%` }}</span>
  </button>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps({
  modelValue: {
    type: String,
    default: 'auto'
  }
})

const emit = defineEmits(['update:modelValue'])

// « auto » d'abord : taille adaptée à l'écran ; puis les tailles fixes.
const zoomLevels = computed(() => [
  { value: 'auto', label: t('dashboard.zoom.auto'), icon: 'lucide:scan' },
  { value: '75', label: t('dashboard.zoom.extraSmall'), icon: 'lucide:zoom-out' },
  { value: '85', label: t('dashboard.zoom.small'), icon: 'lucide:zoom-out' },
  { value: '100', label: t('dashboard.zoom.normal'), icon: 'lucide:search' },
  { value: '115', label: t('dashboard.zoom.large'), icon: 'lucide:zoom-in' }
])

const currentLevel = computed(() => {
  return zoomLevels.value.find(level => level.value === props.modelValue) || zoomLevels.value[0]
})

const isAuto = computed(() => currentLevel.value.value === 'auto')
const autoPercent = computed(() => Math.round(appStore.zoomScale * 100))

const cycleZoom = () => {
  const currentIndex = zoomLevels.value.findIndex(level => level.value === props.modelValue)
  const nextIndex = (currentIndex + 1) % zoomLevels.value.length
  emit('update:modelValue', zoomLevels.value[nextIndex].value)
}
</script>
