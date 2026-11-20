<template>
  <section class="figures" :aria-label="$t('home.keyFigures')">
    <router-link v-for="fig in figures" :key="fig.key" :to="fig.to" class="figures__item">
      <span class="figures__label">{{ $t(fig.labelKey) }}</span>
      <span class="figures__value"><FlapDigits :value="fig.value" /></span>
    </router-link>
  </section>
</template>

<script setup>
// Bandeau des chiffres clés : le second (et dernier) élément « tableau d'affichage » de l'accueil.
// Les chiffres défilent comme des palettes (FlapDigits, figé sous prefers-reduced-motion).
import FlapDigits from '@/components/ui/FlapDigits.vue'

defineProps({
  figures: { type: Array, required: true },
})
</script>

<style scoped>
.figures {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(128px, 1fr));
  border-radius: 6px;
  background: var(--board);
  overflow: hidden;
}

.figures__item {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
  min-width: 0;
  padding: 12px 14px 14px;
  color: var(--board-text);
  box-shadow: inset 1px 0 0 rgb(255 255 255 / 0.06);
  transition: background-color 120ms ease;
}

.figures__item:hover { background: var(--board-2); }

.figures__item:focus-visible {
  outline: 2px solid var(--board-hi);
  outline-offset: -2px;
}

.figures__label {
  font-family: var(--f-num);
  font-size: 10.5px;
  letter-spacing: 0.1em;
  line-height: 1.3;
  text-transform: uppercase;
  color: var(--board-dim);
}

.figures__value {
  font-family: var(--f-num);
  font-size: 24px;
  font-weight: 500;
  line-height: 1;
}

@media (max-width: 640px) {
  .figures { grid-template-columns: repeat(2, 1fr); }
  .figures__value { font-size: 20px; }
}
</style>
