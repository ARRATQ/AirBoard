<template>
  <span class="status-badge" :class="[`status-badge--${tone}`, { 'status-badge--blink': blink }]">
    <slot>{{ label }}</slot>
  </span>
</template>

<script setup>
// Statut commun à tous les modules (direction A+) : la couleur ne signale que l'exception.
// tone : ok | warn | bad | info | neutral. blink : réservé à l'urgence.
defineProps({
  label: { type: String, default: '' },
  tone: {
    type: String,
    default: 'neutral',
    validator: (v) => ['ok', 'warn', 'bad', 'info', 'neutral'].includes(v),
  },
  blink: { type: Boolean, default: false },
})
</script>

<style scoped>
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 7px;
  border: 1px solid currentColor;
  border-radius: var(--r-s);
  font-family: var(--f-num);
  font-size: 10.5px;
  font-weight: 600;
  line-height: 1.4;
  letter-spacing: 0.07em;
  text-transform: uppercase;
  white-space: nowrap;
}

.status-badge--ok { color: var(--ok); }
.status-badge--warn { color: var(--warn); }
.status-badge--bad { color: var(--bad); }
.status-badge--info { color: var(--info); }
.status-badge--neutral { color: var(--muted); }

.status-badge--blink::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
  animation: status-blink 1.2s steps(2) infinite;
}

@keyframes status-blink {
  50% { opacity: 0.15; }
}

@media (prefers-reduced-motion: reduce) {
  .status-badge--blink::before { animation: none; }
}
</style>
