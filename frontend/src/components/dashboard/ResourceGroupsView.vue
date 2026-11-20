<template>
  <div class="rgv" :class="`rgv--${layout}`">
    <ResourceGroup
      v-for="group in appGroups"
      :key="group.id"
      :group-key="group.id"
      :title="group.name"
      :icon="group.icon"
      :color="group.color"
      :count="group.applications?.length || 0"
      :collapsed="collapsedGroups.has(group.id)"
      :is-private="group.is_private"
      :owner-name="group.owner_group?.name || ''"
      :layout="layout === 'columns' ? 'stack' : 'grid'"
      @toggle="$emit('toggle-group', group.id)"
    >
      <ResourceTile
        v-for="app in group.applications"
        :key="app.id"
        :resource="app"
        :favorite="favoritesStore.isFavorite(app.id)"
        @toggle-favorite="toggleFavorite"
        @show-details="$emit('show-details', $event)"
      />
    </ResourceGroup>
  </div>
</template>

<script setup>
// Ressources rangées par catégorie. layout « wide » : une catégorie par ligne, tuiles en grille ;
// layout « columns » : catégories en colonnes, tuiles empilées.
import { useFavoritesStore } from '@/stores/favorites'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import ResourceGroup from './ResourceGroup.vue'
import ResourceTile from './ResourceTile.vue'

defineProps({
  appGroups: { type: Array, required: true },
  collapsedGroups: { type: Set, required: true },
  layout: { type: String, default: 'wide', validator: (v) => ['wide', 'columns'].includes(v) },
})

defineEmits(['toggle-group', 'show-details'])

const favoritesStore = useFavoritesStore()
const appStore = useAppStore()
const { t } = useI18n()

async function toggleFavorite(app) {
  try {
    await favoritesStore.toggleFavorite(app.id)
  } catch (error) {
    console.error('Error toggling favorite:', error)
    appStore.showError(t('ui.toast.error'))
  }
}
</script>

<style scoped>
.rgv--wide {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.rgv--columns {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  align-items: start;
  gap: 22px 16px;
}
</style>
