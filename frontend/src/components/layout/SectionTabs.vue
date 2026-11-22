<template>
  <div>
    <nav
      v-if="tabs.length > 1"
      class="section-tabs"
      :aria-label="$t('nav.tabs.label')"
    >
      <router-link
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="section-tabs__tab"
        :class="{ 'is-active': isActive(tab) }"
        :aria-current="isActive(tab) ? 'page' : undefined"
      >
        {{ $t(tab.labelKey) }}
      </router-link>
    </nav>
    <router-view />
  </div>
</template>

<script setup>
// Page parente d'un module et de ses taxonomies (catégories, types) : une barre
// d'onglets au-dessus de la sous-page courante. Les onglets viennent de
// config/navigation.js (SECTION_TABS) et suivent le rôle de l'utilisateur.
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getRoles } from '@/config/access'
import { visibleTabs } from '@/config/navigation'
import { isPathActive } from '@/utils/navigation'

const props = defineProps({
  section: { type: String, required: true }
})

const route = useRoute()
const authStore = useAuthStore()

const tabs = computed(() => visibleTabs(props.section, getRoles(authStore.user)))
const isActive = (tab) => isPathActive(route.path, tab.to, { exact: tab.exact })
</script>

<style scoped>
.section-tabs {
  display: flex;
  gap: 4px;
  overflow-x: auto;
  padding-inline: 1.5rem;
  padding-block-start: 1rem;
  border-block-end: 1px solid var(--line);
  background-color: var(--surface);
}

.section-tabs__tab {
  flex-shrink: 0;
  padding: 10px 14px;
  margin-block-end: -1px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  border-block-end: 2px solid transparent;
  border-start-start-radius: var(--r-s);
  border-start-end-radius: var(--r-s);
  transition: color 120ms ease, background-color 120ms ease, border-color 120ms ease;
}

.section-tabs__tab:hover {
  color: var(--text);
  background-color: var(--hover);
}

.section-tabs__tab:focus-visible {
  outline: 2px solid var(--fill);
  outline-offset: -2px;
}

.section-tabs__tab.is-active {
  color: var(--text);
  border-block-end-color: var(--fill);
}

@media (max-width: 1024px) {
  .section-tabs {
    padding-inline: 1rem;
  }
}
</style>
