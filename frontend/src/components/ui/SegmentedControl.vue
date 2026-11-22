<template>
  <div class="seg" role="group" :aria-label="label">
    <button
      v-for="opt in options"
      :key="opt.value"
      type="button"
      class="seg__btn"
      :class="{ 'seg__btn--on': modelValue === opt.value, 'seg__btn--icon-sm': iconOnlySmall && opt.icon }"
      :aria-pressed="modelValue === opt.value"
      :title="iconOnlySmall ? opt.label : undefined"
      @click="$emit('update:modelValue', opt.value)"
    >
      <Icon v-if="opt.icon" :icon="opt.icon" class="seg__icon" aria-hidden="true" />
      <span class="seg__label">{{ opt.label }}</span>
      <span v-if="opt.count !== undefined" class="seg__count">{{ opt.count }}</span>
    </button>
  </div>
</template>

<script setup>
// Boutons segmentés (direction A+) : un seul choix actif, aria-pressed sur chaque bouton.
// options : [{ value, label, icon?, count? }]. iconOnlySmall : libellés masqués sous 640 px.
import { Icon } from '@iconify/vue'

defineProps({
  modelValue: { type: [String, Number], default: '' },
  options: { type: Array, required: true },
  label: { type: String, required: true },
  iconOnlySmall: { type: Boolean, default: false },
})

defineEmits(['update:modelValue'])
</script>

<style scoped>
.seg {
  display: inline-flex;
  max-width: 100%;
  gap: 2px;
  padding: 2px;
  border: 1px solid var(--line);
  border-radius: var(--r-s);
  background: var(--sunken);
  overflow-x: auto;
  scrollbar-width: none;
}

.seg__btn {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding-inline: 10px;
  border: 0;
  border-radius: var(--r-s);
  background: none;
  color: var(--muted);
  font: inherit;
  font-size: 12.5px;
  white-space: nowrap;
  cursor: pointer;
}

.seg__btn:hover { color: var(--text); }

.seg__btn--on {
  background: var(--surface);
  color: var(--text);
  font-weight: 600;
  box-shadow: 0 0 0 1px var(--line);
}

.seg__btn:focus-visible {
  outline: 2px solid var(--acc);
  outline-offset: -2px;
}

.seg__icon { width: 15px; height: 15px; }

.seg__count {
  font-family: var(--f-num);
  font-size: 11px;
  font-weight: 500;
  color: var(--muted);
}

@media (max-width: 640px) {
  .seg__btn--icon-sm .seg__label { position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0); white-space: nowrap; }
}
</style>
