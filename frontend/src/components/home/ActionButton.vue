<template>
  <component
    :is="to ? 'router-link' : 'button'"
    :to="to"
    @click="handleClick"
    :class="['action-button', buttonColorClass]"
  >
    <Icon :icon="icon" :class="['h-6 w-6', iconColorClass]" />
    <span :class="['text-sm font-medium', textColorClass]">{{ label }}</span>
  </component>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  icon: { type: String, required: true },
  label: { type: String, required: true },
  to: { type: String, default: null },
  color: { type: String, default: 'blue' }
})

const emit = defineEmits(['click'])

const handleClick = () => {
  if (!props.to) {
    emit('click')
  }
}

const buttonColorClass = computed(() => {
  const colors = {
    blue: 'bg-blue-50 dark:bg-blue-900/20 hover:bg-blue-100 dark:hover:bg-blue-900/30 border-blue-200 dark:border-blue-800',
    purple: 'bg-purple-50 dark:bg-purple-900/20 hover:bg-purple-100 dark:hover:bg-purple-900/30 border-purple-200 dark:border-purple-800',
    green: 'bg-green-50 dark:bg-green-900/20 hover:bg-green-100 dark:hover:bg-green-900/30 border-green-200 dark:border-green-800',
    orange: 'bg-orange-50 dark:bg-orange-900/20 hover:bg-orange-100 dark:hover:bg-orange-900/30 border-orange-200 dark:border-orange-800',
    indigo: 'bg-indigo-50 dark:bg-indigo-900/20 hover:bg-indigo-100 dark:hover:bg-indigo-900/30 border-indigo-200 dark:border-indigo-800',
    teal: 'bg-teal-50 dark:bg-teal-900/20 hover:bg-teal-100 dark:hover:bg-teal-900/30 border-teal-200 dark:border-teal-800',
    yellow: 'bg-yellow-50 dark:bg-yellow-900/20 hover:bg-yellow-100 dark:hover:bg-yellow-900/30 border-yellow-200 dark:border-yellow-800'
  }
  return colors[props.color] || colors.blue
})

const iconColorClass = computed(() => {
  const colors = {
    blue: 'text-blue-600 dark:text-blue-400',
    purple: 'text-purple-600 dark:text-purple-400',
    green: 'text-green-600 dark:text-green-400',
    orange: 'text-orange-600 dark:text-orange-400',
    indigo: 'text-indigo-600 dark:text-indigo-400',
    teal: 'text-teal-600 dark:text-teal-400',
    yellow: 'text-yellow-600 dark:text-yellow-400'
  }
  return colors[props.color] || colors.blue
})

const textColorClass = computed(() => {
  const colors = {
    blue: 'text-blue-900 dark:text-blue-100',
    purple: 'text-purple-900 dark:text-purple-100',
    green: 'text-green-900 dark:text-green-100',
    orange: 'text-orange-900 dark:text-orange-100',
    indigo: 'text-indigo-900 dark:text-indigo-100',
    teal: 'text-teal-900 dark:text-teal-100',
    yellow: 'text-yellow-900 dark:text-yellow-100'
  }
  return colors[props.color] || colors.blue
})
</script>

<style scoped>
.action-button {
  @apply flex flex-col items-center justify-center gap-2 p-4 rounded-lg transition-colors cursor-pointer;
  @apply border hover:border-current;
}
</style>
