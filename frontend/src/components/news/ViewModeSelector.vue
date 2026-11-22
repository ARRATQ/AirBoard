<template>
  <div class="flex items-center gap-1 bg-gray-100 dark:bg-gray-800 rounded-lg p-1">
    <button
      v-for="mode in viewModes"
      :key="mode.value"
      type="button"
      :aria-pressed="modelValue === mode.value"
      @click="$emit('update:modelValue', mode.value)"
      :class="[
        'flex items-center gap-2 px-3 py-1.5 rounded text-sm font-medium transition-all duration-200',
        modelValue === mode.value
          ? 'bg-white dark:bg-gray-700 text-primary-600 dark:text-primary-400 shadow-sm'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
      ]"
      :title="t(mode.labelKey)"
      :aria-label="t(mode.labelKey)"
    >
      <Icon :icon="mode.icon" class="h-4 w-4" aria-hidden="true" />
      <span class="hidden sm:inline">{{ t(mode.labelKey) }}</span>
    </button>
  </div>
</template>

<script setup>
import { Icon } from '@iconify/vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  modelValue: {
    type: String,
    default: 'grid'
  }
})

defineEmits(['update:modelValue'])

// Libellés traduits au rendu (clés i18n)
const viewModes = [
  {
    value: 'grid',
    labelKey: 'dashboard.viewModes.grid',
    icon: 'lucide:layout-grid'
  },
  {
    value: 'list',
    labelKey: 'dashboard.viewModes.list',
    icon: 'lucide:list'
  },
  {
    value: 'compact',
    labelKey: 'news.viewModes.compact',
    icon: 'lucide:rows-4'
  }
]
</script>
