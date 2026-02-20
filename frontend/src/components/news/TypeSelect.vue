<template>
  <div class="relative">
    <!-- Button/Input Display -->
    <button
      type="button"
      @click="isOpen = !isOpen"
      class="w-full px-3 py-2 border-2 border-gray-200 dark:border-gray-700 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-800 dark:text-white text-sm text-left flex items-center justify-between hover:border-gray-300 dark:hover:border-gray-600 transition-colors"
    >
      <div v-if="selectedType" class="flex items-center gap-2">
        <Icon :icon="selectedType.icon" class="h-4 w-4" :style="{ color: selectedType.color }" />
        <span>{{ selectedType.name }}</span>
      </div>
      <div v-else class="text-gray-500">Select a type...</div>
      <Icon icon="mdi:chevron-down" class="h-4 w-4" :class="{ 'rotate-180': isOpen }" />
    </button>

    <!-- Dropdown -->
    <Teleport v-if="isOpen" to="body">
      <div
        class="fixed inset-0 z-40"
        @click="isOpen = false"
      />
    </Teleport>
    <transition name="dropdown">
      <div
        v-if="isOpen"
        class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-50 max-h-64 overflow-y-auto"
      >
        <div
          v-for="type in types"
          :key="type.id"
          @click="selectType(type)"
          class="px-3 py-2 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2 transition-colors border-l-3"
          :style="{ borderColor: modelValue === type.slug ? type.color : 'transparent' }"
        >
          <Icon :icon="type.icon" class="h-4 w-4" :style="{ color: type.color }" />
          <span class="dark:text-white">{{ type.name }}</span>
          <Icon v-if="modelValue === type.slug" icon="mdi:check" class="h-4 w-4 ml-auto" :style="{ color: type.color }" />
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  modelValue: {
    type: String,
    required: true
  },
  types: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)

const selectedType = computed(() => {
  return props.types.find(t => t.slug === props.modelValue)
})

const selectType = (type) => {
  emit('update:modelValue', type.slug)
  isOpen.value = false
}
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
