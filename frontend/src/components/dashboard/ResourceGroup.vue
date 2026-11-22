<template>
  <section class="rgroup" :class="{ 'rgroup--collapsed': collapsed }" :style="{ '--gc': groupColor }" :aria-labelledby="`${domId}-title`">
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
      <span v-if="isPrivate" class="tag rgroup__private">
        {{ $t('dashboard.privateBadge') }}<template v-if="ownerName"> · {{ ownerName }}</template>
      </span>
    </h2>
    <div v-show="!collapsed" :id="`${domId}-body`" class="rgroup__body" :class="`rgroup__body--${layout}`">
      <slot />
    </div>
  </section>
</template>

<script setup>
// Catégorie de ressources repliable, encadrée comme un panneau : bandeau à la couleur de la
// catégorie, en-tête teinté avec nom et nombre, tuiles posées sur un fond en retrait. En-tête
// bouton (aria-expanded), mention « Privé » pour les catégories d'un groupe, tuiles empilées
// (layout stack) ou en grille (layout grid).
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
// Couleur de la catégorie : bandeau du haut et teinte de l'en-tête (accent par défaut).
const groupColor = computed(() => props.color || 'var(--acc)')
</script>

<style scoped>
.rgroup {
  min-width: 0;
  border: 1px solid var(--line);
  border-top: 3px solid var(--gc);
  border-radius: var(--r-l);
  background: var(--sunken);
  overflow: hidden;
}

.rgroup__head {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin: 0;
  padding: 10px 12px;
  border-bottom: 1px solid var(--line);
  background: color-mix(in oklab, var(--gc) 7%, var(--surface));
}

.rgroup--collapsed .rgroup__head { border-bottom: 0; }

.rgroup__toggle {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  max-width: 100%;
  min-height: 32px;
  padding: 0 6px;
  margin-inline-start: -6px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--text);
  font: inherit;
  cursor: pointer;
}

.rgroup__toggle:hover { background: color-mix(in oklab, var(--gc) 12%, var(--surface)); }

.rgroup__toggle:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 1px;
}

.rgroup__icon {
  flex: none;
  box-sizing: content-box;
  width: 18px;
  height: 18px;
  padding: 5px;
  border-radius: var(--r-m);
  background: color-mix(in oklab, var(--gc) 16%, var(--surface));
}

.rgroup__name {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  font-family: var(--f-sans);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0;
}

.rgroup__count {
  flex: none;
  min-width: 22px;
  padding: 1px 7px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--surface);
  font-family: var(--f-num);
  font-size: 11.5px;
  line-height: 18px;
  text-align: center;
  color: var(--muted);
}

.rgroup__chev {
  flex: none;
  width: 18px;
  height: 18px;
  color: var(--muted);
  transition: transform 150ms ease;
}

.rgroup__chev--up { transform: rotate(180deg); }

@media (prefers-reduced-motion: reduce) {
  .rgroup__chev { transition: none; }
}


.rgroup__body { padding: 12px; }

.rgroup__body--grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
  gap: 10px;
}

.rgroup__body--stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

@media (max-width: 480px) {
  .rgroup__head { padding: 8px 10px; }
  .rgroup__body { padding: 10px; }
}
</style>
