<template>
  <div class="tile" :class="{ 'tile--note': note }" :style="{ '--c': color }">
    <span class="tile__icon" aria-hidden="true">
      <span v-if="meta.glyph" class="tile__glyph">{{ meta.glyph }}</span>
      <Icon v-else :icon="iconName" />
    </span>

    <div class="tile__body">
      <a
        :href="resource.url"
        class="tile__link"
        :target="resource.open_in_new_tab ? '_blank' : undefined"
        rel="noopener noreferrer"
        @click="track"
      >
        {{ resource.name }}
        <Icon v-if="resource.open_in_new_tab" icon="lucide:external-link" class="tile__ext" aria-hidden="true" />
        <span v-if="resource.open_in_new_tab" class="sr-only">({{ $t('dashboard.opensInNewTab') }})</span>
      </a>
      <small v-if="resource.description" class="tile__desc">{{ resource.description }}</small>
      <span class="tile__meta">
        <span class="tile__type">{{ $t(`dashboard.types.${type}`) }}</span>
        <StatusBadge v-if="isNew" tone="info">{{ $t('dashboard.newBadge') }}</StatusBadge>
      </span>
      <span v-if="note" class="tile__note">{{ note }}</span>
    </div>

    <div class="tile__actions">
      <button
        type="button"
        class="tile__btn"
        :aria-label="`${$t('common.details')} : ${resource.name}`"
        @click="$emit('show-details', resource)"
      >
        <Icon icon="lucide:info" aria-hidden="true" />
      </button>
      <button
        type="button"
        class="tile__btn tile__btn--star"
        :class="{ 'tile__btn--on': favorite }"
        :aria-pressed="favorite"
        :aria-label="`${$t(favorite ? 'common.removeFromFavorites' : 'common.addToFavorites')} : ${resource.name}`"
        @click="$emit('toggle-favorite', resource)"
      >
        <Icon icon="lucide:star" :class="{ 'icon-fill': favorite }" aria-hidden="true" />
      </button>
    </div>
  </div>
</template>

<script setup>
// Tuile de ressource (direction A+) : repère de type déduit de l'URL, « Nouveau » 14 jours,
// note de l'admin tant que sa date de fin n'est pas passée. Le lien couvre toute la tuile ;
// les boutons Détails et Favori restent au-dessus.
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import StatusBadge from '@/components/ui/StatusBadge.vue'
import { analyticsService } from '@/services/api'
import { TYPE_META, activeAdminNote, isNewResource, resourceTypeOf } from '@/utils/resourceType'

const props = defineProps({
  resource: { type: Object, required: true },
  favorite: { type: Boolean, default: false },
})

defineEmits(['toggle-favorite', 'show-details'])

const type = computed(() => resourceTypeOf(props.resource.url))
const meta = computed(() => TYPE_META[type.value])
const isNew = computed(() => isNewResource(props.resource))
const note = computed(() => activeAdminNote(props.resource))

const DEFAULT_ICON = 'lucide:app-window'
const iconName = computed(() => {
  if (type.value === 'app') return props.resource.icon || DEFAULT_ICON
  // Dossier, site, lien : l'icône choisie par l'admin si elle n'est pas celle par défaut.
  return props.resource.icon && props.resource.icon !== DEFAULT_ICON ? props.resource.icon : meta.value.icon
})
const color = computed(() => (type.value === 'app' ? props.resource.color || 'var(--muted)' : meta.value.color))

// Clic enregistré pour les statistiques, sans bloquer l'ouverture.
const track = () => {
  analyticsService.trackClick(props.resource.id)
}
</script>

<style scoped>
.tile {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  min-width: 0;
  padding: 12px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  transition: border-color 120ms ease;
}

.tile:hover { border-color: var(--acc); }
.tile:focus-within { border-color: var(--acc); box-shadow: 0 0 0 1px var(--acc); }

.tile__icon {
  flex: none;
  display: grid;
  place-items: center;
  width: 34px;
  height: 34px;
  border-radius: var(--r-s);
  background: color-mix(in oklab, var(--c) 14%, var(--surface));
  color: color-mix(in oklab, var(--c) 85%, var(--text));
  font-size: 19px;
}

.tile__glyph {
  font-family: var(--f-num);
  font-size: 12px;
  font-weight: 600;
  direction: ltr;
}

.tile__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding-inline-end: 48px;
}

.tile__link {
  font-weight: 600;
  line-height: 1.3;
  color: var(--text);
  overflow-wrap: anywhere;
}

/* Le lien couvre toute la tuile. */
.tile__link::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: var(--r-s);
}

.tile__link:focus-visible { outline: none; }

.tile__ext {
  display: inline-block;
  width: 12px;
  height: 12px;
  margin-inline-start: 2px;
  vertical-align: -1px;
  color: var(--muted);
}

[dir='rtl'] .tile__ext { transform: scaleX(-1); }

.tile__desc {
  margin-block-start: 2px;
  font-size: 12px;
  line-height: 1.35;
  color: var(--muted);
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.tile__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-block-start: 6px;
}

.tile__type {
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
}

.tile__note {
  margin-block-start: 6px;
  padding-inline-start: 7px;
  border-inline-start: 2px solid var(--fill);
  font-size: 12px;
  line-height: 1.35;
  color: var(--text);
}

.tile__actions {
  position: absolute;
  inset-block-start: 6px;
  inset-inline-end: 6px;
  z-index: 1;
  display: flex;
  gap: 2px;
}

.tile__btn {
  display: grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  font-size: 16px;
  cursor: pointer;
}

.tile__btn:hover { background: var(--sunken); color: var(--text); }

.tile__btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}

.tile__btn--on { color: #d99a00; }
</style>
