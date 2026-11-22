<template>
  <component
    :is="tag"
    :type="tag === 'button' ? type : undefined"
    class="base-btn"
    :class="[`base-btn--${variant}`, `base-btn--${size}`, { 'base-btn--block': block }]"
    :disabled="tag === 'button' ? disabled || loading : undefined"
    :aria-busy="loading || undefined"
  >
    <Icon v-if="loading" icon="lucide:loader-circle" class="base-btn__icon base-btn__spin" aria-hidden="true" />
    <Icon v-else-if="icon" :icon="icon" class="base-btn__icon" :class="{ 'rtl-flip': isDirectional }" aria-hidden="true" />
    <slot />
  </component>
</template>

<script setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps({
  variant: {
    type: String,
    default: 'secondary',
    validator: (v) => ['primary', 'secondary', 'ghost', 'danger'].includes(v),
  },
  size: { type: String, default: 'md', validator: (v) => ['sm', 'md'].includes(v) },
  type: { type: String, default: 'button' },
  tag: { type: [String, Object], default: 'button' },
  icon: { type: String, default: '' },
  loading: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  block: { type: Boolean, default: false },
})

// Chevrons et flèches gauche/droite indiquent un sens de lecture : ils s'inversent en arabe.
const isDirectional = computed(() => /(chevron|arrow)-(left|right)$/.test(props.icon))
</script>

<style scoped>
.base-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 32px;
  padding-inline: 12px;
  border: 1px solid transparent;
  border-radius: var(--r-m);
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
  text-decoration: none;
  cursor: pointer;
  transition: background-color 0.12s, border-color 0.12s, filter 0.12s;
}

.base-btn--sm {
  height: 28px;
  padding-inline: 10px;
  font-size: 12.5px;
  border-radius: var(--r-s);
}

.base-btn--block { width: 100%; }

.base-btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: 2px;
}

.base-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.base-btn--primary {
  background: var(--fill);
  color: var(--on-fill);
}

.base-btn--primary:not(:disabled):hover { filter: brightness(1.08); }

.base-btn--secondary {
  background: var(--surface);
  color: var(--text);
  border-color: var(--line);
}

.base-btn--secondary:not(:disabled):hover { background: var(--sunken); }

.base-btn--ghost {
  background: transparent;
  color: var(--muted);
}

.base-btn--ghost:not(:disabled):hover {
  background: var(--sunken);
  color: var(--text);
}

.base-btn--danger {
  background: var(--bad);
  color: #fff;
}

html.dark .base-btn--danger { color: #141210; }

.base-btn--danger:not(:disabled):hover { filter: brightness(1.08); }

.base-btn__icon {
  width: 16px;
  height: 16px;
  flex: none;
}

.base-btn__spin { animation: base-btn-spin 0.8s linear infinite; }

@keyframes base-btn-spin {
  to { transform: rotate(360deg); }
}
</style>
