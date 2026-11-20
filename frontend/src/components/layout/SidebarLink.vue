<template>
  <router-link
    :to="to"
    :class="['nav-link', isActive ? 'nav-link-active' : 'nav-link-inactive']"
    :title="compact ? label : undefined"
    :aria-label="compact ? label : undefined"
    :aria-current="isActive ? 'page' : undefined"
  >
    <Icon :icon="icon" class="h-4 w-4 flex-shrink-0" aria-hidden="true" />
    <span>{{ label }}</span>
  </router-link>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useAppStore } from '@/stores/app'
import { isPathActive } from '@/utils/navigation'

const props = defineProps({
  to: { type: String, required: true },
  icon: { type: String, required: true },
  label: { type: String, required: true },
  // Actif seulement sur son propre chemin, pas sur ses sous-pages
  exact: { type: Boolean, default: false }
})

const route = useRoute()
const appStore = useAppStore()

const isActive = computed(() => isPathActive(route.path, props.to, { exact: props.exact }))
// Menu replié : le libellé est masqué, il passe en infobulle et en nom accessible
const compact = computed(() => !appStore.sidebarOpen)
</script>
