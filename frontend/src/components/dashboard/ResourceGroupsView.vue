<template>
  <div ref="root" class="rgv" :class="`rgv--${layout}`">
    <template v-if="layout === 'columns'">
      <div v-for="(column, index) in columns" :key="index" class="rgv__col">
        <div
          v-for="group in column"
          :key="group.id"
          :ref="(el) => trackGroup(group.id, el)"
          class="rgv__item"
        >
          <ResourceGroup
            :group-key="group.id"
            :title="group.name"
            :icon="group.icon"
            :color="group.color"
            :count="group.applications?.length || 0"
            :collapsed="collapsedGroups.has(group.id)"
            :is-private="group.is_private"
            :owner-name="group.owner_group?.name || ''"
            layout="stack"
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
      </div>
    </template>

    <template v-else>
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
        layout="grid"
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
    </template>
  </div>
</template>

<script setup>
// Ressources rangées par catégorie. layout « wide » : une catégorie par ligne, tuiles en grille ;
// layout « columns » : catégories en colonnes, tuiles empilées. En colonnes, chaque catégorie
// va dans la colonne la plus courte (maçonnerie), d'après sa hauteur réelle mesurée : une
// catégorie courte ne laisse plus de grand vide à côté d'une longue.
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useFavoritesStore } from '@/stores/favorites'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { columnCountFor, packColumns } from '@/utils/packColumns'
import ResourceGroup from './ResourceGroup.vue'
import ResourceTile from './ResourceTile.vue'

const props = defineProps({
  appGroups: { type: Array, required: true },
  collapsedGroups: { type: Set, required: true },
  layout: { type: String, default: 'wide', validator: (v) => ['wide', 'columns'].includes(v) },
})

defineEmits(['toggle-group', 'show-details'])

const favoritesStore = useFavoritesStore()
const appStore = useAppStore()
const { t } = useI18n()

// Mêmes valeurs que le CSS ci-dessous.
const COLUMN_MIN_WIDTH = 300
const COLUMN_GAP = 16
const ITEM_GAP = 18
// Hauteur supposée avant la première mesure : en-tête, puis environ 88 px par tuile empilée.
const HEADER_HEIGHT = 58
const TILE_HEIGHT = 88

const root = ref(null)
const width = ref(0)
const heights = ref(new Map())

const estimate = (id) => {
  const group = props.appGroups.find((g) => g.id === id)
  if (!group || props.collapsedGroups.has(id)) return HEADER_HEIGHT
  return HEADER_HEIGHT + 24 + (group.applications?.length || 0) * TILE_HEIGHT
}

const columns = computed(() => {
  const count = columnCountFor(width.value, COLUMN_MIN_WIDTH, COLUMN_GAP)
  const byId = new Map(props.appGroups.map((g) => [g.id, g]))
  return packColumns(props.appGroups.map((g) => g.id), heights.value, count, ITEM_GAP, estimate)
    .map((ids) => ids.map((id) => byId.get(id)))
})

// Un seul observateur pour la largeur du conteneur et la hauteur de chaque catégorie.
const elementIds = new WeakMap()
const observed = new Map()
let observer = null

function onResize(entries) {
  let next = null
  for (const entry of entries) {
    if (entry.target === root.value) {
      width.value = entry.contentRect.width
      continue
    }
    const id = elementIds.get(entry.target)
    if (id === undefined) continue
    const height = Math.round(entry.borderBoxSize?.[0]?.blockSize ?? entry.target.getBoundingClientRect().height)
    if (Math.abs((heights.value.get(id) ?? -1) - height) > 1) {
      next ||= new Map(heights.value)
      next.set(id, height)
    }
  }
  if (next) heights.value = next
}

function trackGroup(id, el) {
  const previous = observed.get(id)
  if (previous === el) return
  if (previous && observer) observer.unobserve(previous)
  if (el) {
    elementIds.set(el, id)
    observed.set(id, el)
    observer?.observe(el)
  } else {
    observed.delete(id)
  }
}

onMounted(() => {
  if (typeof ResizeObserver === 'undefined') return
  observer = new ResizeObserver(onResize)
  if (root.value) {
    width.value = root.value.clientWidth
    observer.observe(root.value)
  }
  observed.forEach((el) => observer.observe(el))
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})

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
  gap: 18px;
}

.rgv--columns {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.rgv__col {
  display: flex;
  flex: 1 1 0;
  flex-direction: column;
  gap: 18px;
  min-width: 0;
}
</style>
