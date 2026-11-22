<template>
  <span class="tag status-badge" :class="[`tag--${tone}`, { 'status-badge--blink': blink }]">
    <slot>{{ label }}</slot>
  </span>
</template>

<script setup>
// Statut commun à tous les modules (direction A+) : la couleur ne signale que l'exception.
// tone : ok | warn | bad | info | neutral. blink : réservé à l'urgence. Dessin : .tag (main.css).
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
.status-badge { gap: 6px; }

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
