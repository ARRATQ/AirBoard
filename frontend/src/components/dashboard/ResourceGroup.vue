<template>
  <section class="rgroup" :aria-labelledby="`${domId}-title`">
    <h2 :id="`${domId}-title`" class="rgroup__head">
      <button
        type="button"
        class="rgroup__toggle"
        :aria-expanded="!collapsed"
        :aria-controls="`${domId}-body`"
        @click="$emit('toggle')"
      >
        <Icon :icon="icon || 'mdi:folder-outline'" class="rgroup__icon" :style="{ color: iconColor }" aria-hidden="true" />
        <span class="rgroup__name">{{ title }}</span>
        <span class="rgroup__count">{{ count }}</span>
        <Icon icon="mdi:chevron-down" class="rgroup__chev" :class="{ 'rgroup__chev--up': !collapsed }" aria-hidden="true" />
      </button>
      <span v-if="isPrivate" class="rgroup__private">
        {{ $t('dashboard.privateBadge') }}<template v-if="ownerName"> · {{ ownerName }}</template>
      </span>
    </h2>
    <div v-show="!collapsed" :id="`${domId}-body`" class="rgroup__body" :class="`rgroup__body--${layout}`">
      <slot />
    </div>
  </section>
</template>

<script setup>
// Catégorie de ressources repliable : en-tête bouton (aria-expanded), mention « Privé » pour les
// catégories d'un groupe, tuiles empilées (layout stack) ou en grille (layout grid).
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  groupKey: { type: [String, Number], required: true },
  title: { type: String, required: true },
  icon: { type: String, default: '' },
  color: { type: String, default: '' },
  count: { type: Number, default: 0 },
  collapsed: { type: Boolean, default: false },
  isPrivate: { type: Boolean, default: false },
  ownerName: { type: String, default: '' },
  layout: { type: String, default: 'grid', validator: (v) => ['grid', 'stack'].includes(v) },
})

defineEmits(['toggle'])

const domId = computed(() => `rgroup-${props.groupKey}`)
const iconColor = computed(() => props.color || 'var(--muted)')
</script>

<style scoped>
.rgroup { min-width: 0; }

.rgroup__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin: 0 0 8px;
}

.rgroup__toggle {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  max-width: 100%;
  min-height: 28px;
  padding: 0 4px;
  border: 0;
  border-radius: 3px;
  background: none;
  color: var(--text);
  font: inherit;
  cursor: pointer;
}

.rgroup__toggle:hover { background: var(--sunken); }

.rgroup__toggle:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}

.rgroup__icon { flex: none; width: 17px; height: 17px; }

.rgroup__name {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: var(--f-num);
  font-size: 11.5px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
}

.rgroup__count {
  font-family: var(--f-num);
  font-size: 11.5px;
  color: var(--muted);
}

.rgroup__chev {
  flex: none;
  width: 16px;
  height: 16px;
  color: var(--muted);
  transition: transform 150ms ease;
}

.rgroup__chev--up { transform: rotate(180deg); }

@media (prefers-reduced-motion: reduce) {
  .rgroup__chev { transition: none; }
}

.rgroup__private {
  padding: 1px 6px;
  border: 1px dashed var(--line);
  border-radius: 3px;
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: var(--muted);
}

.rgroup__body--grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 8px;
}

.rgroup__body--stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
</style>
