<template>
  <div :class="['collapsible-menu', { 'nested': nested }]">
    <!-- Menu Header (clickable to toggle) -->
    <button
      type="button"
      @click="toggle"
      :class="[
        'collapsible-menu-header',
        { 'is-open': isOpen, 'has-active-child': hasActiveChild, 'is-compact': compact }
      ]"
      :aria-expanded="isOpen"
      :title="compact ? title : undefined"
      :aria-label="compact ? title : undefined"
    >
      <div class="flex items-center gap-3 flex-1 min-w-0" :class="{ 'justify-center': compact }">
        <Icon :icon="icon" class="h-4 w-4 flex-shrink-0" aria-hidden="true" />
        <span v-if="!compact" class="truncate">{{ title }}</span>
      </div>
      <Icon
        v-if="!compact"
        :icon="isOpen ? 'lucide:chevron-up' : 'lucide:chevron-down'"
        class="h-4 w-4 flex-shrink-0 transition-transform duration-200"
        aria-hidden="true"
      />
    </button>

    <!-- Menu Content (collapsible) -->
    <Transition
      enter-active-class="transition-all duration-200 ease-out"
      enter-from-class="opacity-0 max-h-0"
      enter-to-class="opacity-100 max-h-[500px]"
      leave-active-class="transition-all duration-150 ease-in"
      leave-from-class="opacity-100 max-h-[500px]"
      leave-to-class="opacity-0 max-h-0"
    >
      <div v-show="isOpen" class="collapsible-menu-content">
        <slot></slot>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  title: {
    type: String,
    required: true
  },
  icon: {
    type: String,
    default: 'lucide:folder'
  },
  storageKey: {
    type: String,
    default: null
  },
  defaultOpen: {
    type: Boolean,
    default: false
  },
  nested: {
    type: Boolean,
    default: false
  },
  // Menu latéral replié : seule l'icône reste, le titre passe en infobulle
  compact: {
    type: Boolean,
    default: false
  }
})

const isOpen = ref(false)

// Check if any child route is active
const hasActiveChild = computed(() => {
  return props.defaultOpen
})

// Load state from localStorage on mount
onMounted(() => {
  if (props.storageKey) {
    const saved = localStorage.getItem(props.storageKey)
    if (saved !== null) {
      isOpen.value = saved === 'true'
    } else {
      isOpen.value = props.defaultOpen
    }
  } else {
    isOpen.value = props.defaultOpen
  }

  // If a child is active, force open
  if (props.defaultOpen) {
    isOpen.value = true
  }
})

// Watch for route changes to auto-open if child becomes active
watch(() => props.defaultOpen, (newVal) => {
  if (newVal) {
    isOpen.value = true
  }
})

// Toggle function
const toggle = () => {
  isOpen.value = !isOpen.value

  // Save state to localStorage
  if (props.storageKey) {
    localStorage.setItem(props.storageKey, isOpen.value.toString())
  }
}
</script>

<style scoped>
.collapsible-menu {
  @apply mb-1;
}

.collapsible-menu.nested {
  @apply ms-0;
}

.collapsible-menu-header {
  @apply w-full flex items-center justify-between gap-2 px-3 py-2 font-semibold rounded-lg transition-colors;
  font-size: var(--fs-nav);
  @apply text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700;
  @apply hover:text-gray-900 dark:hover:text-white;
}

.collapsible-menu-header.is-compact {
  @apply justify-center px-2;
}

.collapsible-menu-header.is-open {
  @apply text-gray-900 dark:text-white;
}

.collapsible-menu-header.has-active-child {
  @apply text-primary-600 dark:text-primary-400;
}

.collapsible-menu.nested .collapsible-menu-header {
  @apply py-1.5 ps-2;
  font-size: var(--fs-small);
}

.collapsible-menu-content {
  @apply overflow-hidden;
}

.collapsible-menu.nested .collapsible-menu-content {
  @apply ps-2 border-s-2 border-gray-200 dark:border-gray-700 ms-3 mt-1;
}

.collapsible-menu-content :deep(.nav-link) {
  @apply ps-3 ms-3;
  font-size: var(--fs-small);
}

.collapsible-menu.nested .collapsible-menu-content :deep(.nav-link) {
  @apply ps-4 py-1.5;
  font-size: var(--fs-small);
}
</style>
