<template>
  <span class="flaps" role="img" :aria-label="String(value)">
    <i
      v-for="(ch, i) in shown"
      :key="i"
      class="flaps__cell"
      :class="{ 'flaps__cell--space': ch === ' ' }"
      aria-hidden="true"
    >{{ ch }}</i>
  </span>
</template>

<script setup>
// Chiffres façon panneau à palettes : les chiffres défilent puis se figent, de gauche à droite.
// Aucune animation sous prefers-reduced-motion ; la valeur réelle est toujours dans aria-label.
import { onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps({
  value: { type: [String, Number], required: true },
  stagger: { type: Number, default: 130 }, // ms entre deux cases
})

const FRAME_MS = 55
const BASE_MS = 260
const shown = ref([...String(props.value)])
let frame = 0

const prefersReducedMotion = () =>
  typeof window !== 'undefined' && window.matchMedia?.('(prefers-reduced-motion: reduce)').matches

const isDigit = (ch) => ch >= '0' && ch <= '9'

function animate(target) {
  cancelAnimationFrame(frame)
  const chars = [...target]
  if (prefersReducedMotion() || typeof requestAnimationFrame === 'undefined') {
    shown.value = chars
    return
  }
  const t0 = performance.now()
  let last = 0
  const step = (now) => {
    if (now - last >= FRAME_MS) {
      last = now
      shown.value = chars.map((ch, i) =>
        isDigit(ch) && now - t0 < BASE_MS + i * props.stagger ? String(Math.floor(Math.random() * 10)) : ch,
      )
    }
    if (now - t0 < BASE_MS + chars.length * props.stagger) {
      frame = requestAnimationFrame(step)
    } else {
      shown.value = chars
    }
  }
  frame = requestAnimationFrame(step)
}

watch(() => String(props.value), animate, { immediate: true })
onBeforeUnmount(() => cancelAnimationFrame(frame))
</script>

<style scoped>
.flaps {
  display: inline-flex;
  gap: 2px;
  direction: ltr;
  font-family: var(--f-num);
  font-variant-numeric: tabular-nums;
}

.flaps__cell {
  position: relative;
  min-width: 0.74em;
  padding: 1px;
  border-radius: 2px;
  background: var(--flap-cell, #0a0b0d);
  box-shadow: var(--flap-ring, none);
  color: var(--board-text);
  font-style: normal;
  text-align: center;
}

.flaps__cell::after {
  content: '';
  position: absolute;
  inset-inline: 0;
  top: 50%;
  height: 1px;
  background: var(--flap-rule, rgb(255 255 255 / 0.14));
}

.flaps__cell--space {
  min-width: 0.35em;
  background: none;
}

.flaps__cell--space::after { display: none; }
</style>
